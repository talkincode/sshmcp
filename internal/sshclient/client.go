package sshclient

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

const (
	DefaultSSHPort    = "22"
	DefaultSSHUser    = "master"
	DefaultSudoKey    = "master"
	DefaultTimeout    = 30 * time.Second
	SudoPrompt        = "[sudo] password"
	PasswordPromptEnd = ": "
)

// Config represents SSH configuration properties for connecting to remote hosts.
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	KeyPath  string
	SudoKey  string
	Command  string
	Mode     string

	SafetyCheck bool
	Force       bool

	SftpAction string
	LocalPath  string
	RemotePath string

	PasswordAction string
	PasswordKey    string
	PasswordValue  string

	// Host management fields
	HostAction      string
	HostName        string
	HostDescription string
	HostType        string
}

// SSHClient wraps an ssh.Client with optional pooled and sftp helpers.
type SSHClient struct {
	config     *Config
	client     *ssh.Client
	sftpClient *sftp.Client
}

// getHostKeyCallback returns a secure host key callback function
// It tries to use known_hosts file, falls back to InsecureIgnoreHostKey with warning
func getHostKeyCallback() ssh.HostKeyCallback {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Printf("⚠️  Warning: Unable to get home directory, using insecure host key verification")
		// #nosec G106 -- This is a fallback when known_hosts is unavailable
		return ssh.InsecureIgnoreHostKey()
	}

	knownHostsPath := filepath.Join(home, ".ssh", "known_hosts")

	// Try to use known_hosts file
	hostKeyCallback, err := knownhosts.New(knownHostsPath)
	if err != nil {
		// If known_hosts doesn't exist or can't be read, create it or use insecure mode
		if os.IsNotExist(err) {
			log.Printf("⚠️  Warning: known_hosts file not found at %s", knownHostsPath)
			log.Printf("⚠️  Using insecure host key verification (vulnerable to MITM attacks)")
			log.Printf("💡 Tip: Run 'ssh-keyscan %s >> %s' to add host keys", "HOST", knownHostsPath)
		} else {
			log.Printf("⚠️  Warning: Failed to load known_hosts: %v", err)
			log.Printf("⚠️  Using insecure host key verification")
		}
		// #nosec G106 -- Documented fallback with user warning
		return ssh.InsecureIgnoreHostKey()
	}

	// Wrap the callback to handle key verification errors gracefully
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		err := hostKeyCallback(hostname, remote, key)
		if err != nil {
			// Check if it's a knownhosts.KeyError (host key mismatch or unknown host)
			var keyErr *knownhosts.KeyError
			if keyError, ok := err.(*knownhosts.KeyError); ok {
				keyErr = keyError
				// If there are known keys but they don't match, it's a key change
				if len(keyErr.Want) > 0 {
					return fmt.Errorf("⚠️  HOST KEY VERIFICATION FAILED!\n"+
						"The host key for %s has changed.\n"+
						"This could indicate a man-in-the-middle attack.\n"+
						"Remove the old key from %s and verify the new key before connecting.\n"+
						"Original error: %w", hostname, knownHostsPath, err)
				}
				// If no known keys exist, it's an unknown host
				return fmt.Errorf("⚠️  Host %s is not in known_hosts file.\n"+
					"To add this host, run:\n"+
					"  ssh-keyscan -H %s >> %s\n"+
					"Or connect manually first:\n"+
					"  ssh %s@%s\n"+
					"Original error: %w",
					hostname, hostname, knownHostsPath, "USER", hostname, err)
			}
			return err
		}
		return nil
	}
}

// NewSSHClient 创建SSH客户端
func NewSSHClient(config *Config) (*SSHClient, error) {
	if config.Host == "" {
		return nil, fmt.Errorf("host is required")
	}
	if config.Port == "" {
		config.Port = DefaultSSHPort
	}
	if config.User == "" {
		config.User = DefaultSSHUser
	}
	if config.KeyPath == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			config.KeyPath = filepath.Join(home, ".ssh", "id_rsa")
		}
	}

	return &SSHClient{config: config}, nil
}

// Connect establishes an SSH connection (prefers using connection pool)
func (c *SSHClient) Connect() error {
	pool := GetConnectionPool()
	client, err := pool.GetConnection(c.config)
	if err == nil {
		c.client = client
		return nil
	}

	log.Printf("Connection pool failed, falling back to direct connection: %v", err)
	return c.connectDirect()
}

// connectDirect establishes a direct SSH connection (without using connection pool)
func (c *SSHClient) connectDirect() error {
	var authMethods []ssh.AuthMethod

	if c.config.KeyPath != "" {
		// Expand ~ in path to user home directory
		keyPath := c.config.KeyPath
		if strings.HasPrefix(keyPath, "~/") {
			if home, err := os.UserHomeDir(); err == nil {
				keyPath = filepath.Join(home, keyPath[2:])
			}
		}

		if key, err := os.ReadFile(keyPath); err == nil { //nolint:gosec // G304: key path is provided by user
			signer, signerErr := ssh.ParsePrivateKey(key)
			if signerErr == nil {
				authMethods = append(authMethods, ssh.PublicKeys(signer))
				log.Printf("Using SSH key: %s", keyPath)
			} else {
				log.Printf("Warning: failed to parse SSH key: %v", signerErr)
			}
		} else {
			log.Printf("Warning: failed to read SSH key file %s: %v", keyPath, err)
		}
	}

	if c.config.Password != "" {
		authMethods = append(authMethods, ssh.Password(c.config.Password))
		log.Println("Using password authentication")
	}

	if len(authMethods) == 0 {
		return fmt.Errorf("no authentication method available")
	}

	sshConfig := &ssh.ClientConfig{
		User:            c.config.User,
		Auth:            authMethods,
		HostKeyCallback: getHostKeyCallback(),
		Timeout:         DefaultTimeout,
	}

	addr := fmt.Sprintf("%s:%s", c.config.Host, c.config.Port)
	log.Printf("Connecting to %s@%s...", c.config.User, addr)

	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}

	c.client = client
	log.Println("✓ Connected successfully")
	return nil
}

// ExecuteCommand executes a command
func (c *SSHClient) ExecuteCommand() (err error) {
	if c.config.SafetyCheck && !c.config.Force {
		if validateErr := ValidateCommand(c.config.Command); validateErr != nil {
			return validateErr
		}
	} else if c.config.Force {
		log.Println("⚠️  Safety check skipped (--force mode)")
	}

	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer CloseIgnore(&err, session)

	if c.config.Password != "" && strings.Contains(c.config.Command, "sudo") {
		return c.executeInteractive(session)
	}

	return c.executeWithPTY(session)
}

// ExecuteCommandWithOutput executes a command and returns the output
func (c *SSHClient) ExecuteCommandWithOutput() (output string, err error) {
	if c.config.SafetyCheck && !c.config.Force {
		if validateErr := ValidateCommand(c.config.Command); validateErr != nil {
			return "", validateErr
		}
	}

	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer CloseIgnore(&err, session)

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	var execErr error
	if c.config.Password != "" && strings.Contains(c.config.Command, "sudo") {
		actualCmd := strings.TrimPrefix(c.config.Command, "sudo ")
		actualCmd = strings.TrimSpace(actualCmd)
		finalCmd := fmt.Sprintf(`printf '%%s\n' '%s' | sudo -S %s`, c.config.Password, actualCmd)

		execErr = session.Run(finalCmd)
	} else {
		execErr = session.Run(c.config.Command)
	}

	// Build detailed output regardless of success/failure
	output = stdout.String()
	stderrStr := stderr.String()

	if execErr != nil {
		// Build comprehensive error message
		errMsg := fmt.Sprintf("command failed: %v", execErr)

		if stderrStr != "" {
			errMsg += fmt.Sprintf("\nStderr: %s", stderrStr)
		}

		if output != "" {
			errMsg += fmt.Sprintf("\nStdout: %s", output)
		}

		// Include exit code if available
		if exitErr, ok := execErr.(*ssh.ExitError); ok {
			errMsg += fmt.Sprintf("\nExit Code: %d", exitErr.ExitStatus())
		}

		return "", fmt.Errorf("%s", errMsg)
	}

	// For successful execution, include stderr in output if present
	if stderrStr != "" {
		output += "\n--- STDERR ---\n" + stderrStr
	}

	return output, nil
}

// executeWithPTY executes a command using PTY
func (c *SSHClient) executeWithPTY(session *ssh.Session) error {
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		log.Printf("Warning: failed to request PTY, falling back to normal execution: %v", err)
		return c.executeNormal(session)
	}

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	log.Printf("Executing (with PTY): %s", c.config.Command)

	if err := session.Run(c.config.Command); err != nil {
		if stderr.Len() > 0 {
			fmt.Fprintf(os.Stderr, "STDERR:\n%s", stderr.String())
		}
		return fmt.Errorf("command failed: %w", err)
	}

	if stdout.Len() > 0 {
		fmt.Print(stdout.String())
	}
	if stderr.Len() > 0 {
		fmt.Fprintf(os.Stderr, "%s", stderr.String())
	}

	return nil
}

// executeNormal executes a normal command (without PTY)
func (c *SSHClient) executeNormal(session *ssh.Session) error {
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	log.Printf("Executing: %s", c.config.Command)

	if err := session.Run(c.config.Command); err != nil {
		if stderr.Len() > 0 {
			fmt.Fprintf(os.Stderr, "STDERR:\n%s", stderr.String())
		}
		return fmt.Errorf("command failed: %w", err)
	}

	if stdout.Len() > 0 {
		fmt.Print(stdout.String())
	}
	if stderr.Len() > 0 {
		fmt.Fprintf(os.Stderr, "%s", stderr.String())
	}

	return nil
}

// executeInteractive executes an interactive command (supports auto sudo password input)
func (c *SSHClient) executeInteractive(session *ssh.Session) error {
	var finalCmd string
	if c.config.Password != "" {
		log.Println("Auto-filling sudo password...")
		actualCmd := strings.TrimPrefix(c.config.Command, "sudo ")
		actualCmd = strings.TrimSpace(actualCmd)
		finalCmd = fmt.Sprintf(`printf '%%s\n' '%s' | sudo -S %s`, c.config.Password, actualCmd)
	} else {
		finalCmd = c.config.Command
	}

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	log.Printf("Executing (no PTY): %s", "sudo command")

	if err := session.Run(finalCmd); err != nil {
		if stderr.Len() > 0 {
			fmt.Fprintf(os.Stderr, "STDERR:\n%s", stderr.String())
		}
		return fmt.Errorf("command failed: %w", err)
	}

	if stdout.Len() > 0 {
		fmt.Print(stdout.String())
	}
	if stderr.Len() > 0 {
		fmt.Fprintf(os.Stderr, "%s", stderr.String())
	}

	return nil
}

// ExecuteSftp executes SFTP operations
func (c *SSHClient) ExecuteSftp() (err error) {
	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer CloseIgnore(&err, sftpClient)
	c.sftpClient = sftpClient

	switch c.config.SftpAction {
	case "upload":
		return c.uploadFile()
	case "download":
		return c.downloadFile()
	case "list", "ls":
		return c.listFiles()
	case "mkdir":
		return c.makeDirectory()
	case "remove", "rm":
		return c.removeFile()
	default:
		return fmt.Errorf("unknown SFTP action: %s", c.config.SftpAction)
	}
}

func (c *SSHClient) uploadFile() (err error) {
	localFile, err := os.Open(c.config.LocalPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer CloseIgnore(&err, localFile)

	remoteFile, err := c.sftpClient.Create(c.config.RemotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file: %w", err)
	}
	defer CloseIgnore(&err, remoteFile)

	log.Printf("Uploading: %s → %s", c.config.LocalPath, c.config.RemotePath)

	written, err := io.Copy(remoteFile, localFile)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	log.Printf("✓ Uploaded %d bytes successfully", written)
	return nil
}

func (c *SSHClient) downloadFile() (err error) {
	remoteFile, err := c.sftpClient.Open(c.config.RemotePath)
	if err != nil {
		return fmt.Errorf("failed to open remote file: %w", err)
	}
	defer CloseIgnore(&err, remoteFile)

	localFile, err := os.Create(c.config.LocalPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer CloseIgnore(&err, localFile)

	log.Printf("Downloading: %s → %s", c.config.RemotePath, c.config.LocalPath)

	written, err := io.Copy(localFile, remoteFile)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	log.Printf("✓ Downloaded %d bytes successfully", written)
	return nil
}

func (c *SSHClient) listFiles() error {
	remotePath := c.config.RemotePath
	if remotePath == "" {
		remotePath = "."
	}

	files, err := c.sftpClient.ReadDir(remotePath)
	if err != nil {
		return fmt.Errorf("failed to list directory: %w", err)
	}

	log.Printf("Directory listing: %s", remotePath)
	fmt.Println("\nPermissions  Size      Modified              Name")
	fmt.Println("-------------------------------------------------------")

	for _, file := range files {
		modeStr := file.Mode().String()
		sizeStr := fmt.Sprintf("%10d", file.Size())
		timeStr := file.ModTime().Format("2006-01-02 15:04:05")

		fmt.Printf("%-12s %s  %s  %s\n", modeStr, sizeStr, timeStr, file.Name())
	}

	fmt.Printf("\nTotal: %d items\n", len(files))
	return nil
}

func (c *SSHClient) makeDirectory() error {
	if err := c.sftpClient.MkdirAll(c.config.RemotePath); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	log.Printf("✓ Directory created: %s", c.config.RemotePath)
	return nil
}

func (c *SSHClient) removeFile() error {
	stat, err := c.sftpClient.Stat(c.config.RemotePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	if stat.IsDir() {
		if err := c.removeDirectory(c.config.RemotePath); err != nil {
			return err
		}
		log.Printf("✓ Directory removed: %s", c.config.RemotePath)
	} else {
		if err := c.sftpClient.Remove(c.config.RemotePath); err != nil {
			return fmt.Errorf("failed to remove file: %w", err)
		}
		log.Printf("✓ File removed: %s", c.config.RemotePath)
	}

	return nil
}

func (c *SSHClient) removeDirectory(path string) error {
	files, err := c.sftpClient.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, file := range files {
		fullPath := filepath.Join(path, file.Name())
		if file.IsDir() {
			if err := c.removeDirectory(fullPath); err != nil {
				return err
			}
		} else {
			if err := c.sftpClient.Remove(fullPath); err != nil {
				return fmt.Errorf("failed to remove file %s: %w", fullPath, err)
			}
		}
	}

	return c.sftpClient.RemoveDirectory(path)
}

// Close closes the connection (releases back to connection pool)
func (c *SSHClient) Close() error {
	if c.config != nil {
		pool := GetConnectionPool()
		pool.ReleaseConnection(c.config)
	}
	return nil
}

// ForceClose forcefully closes the connection (does not release back to pool)
func (c *SSHClient) ForceClose() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}
