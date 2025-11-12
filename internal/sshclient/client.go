package sshclient

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const (
	DefaultSSHPort    = "22"
	DefaultSSHUser    = "master"
	DefaultSudoKey    = "master"
	DefaultTimeout    = 30 * time.Second
	SudoPrompt        = "[sudo] password"
	PasswordPromptEnd = ": "
)

// Config SSH配置
// Config properties for connecting to remote hosts.
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
}

// SSHClient SSH客户端
// SSHClient wraps an ssh.Client with optional pooled and sftp helpers.
type SSHClient struct {
	config     *Config
	client     *ssh.Client
	sftpClient *sftp.Client
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

// Connect 建立SSH连接（优先使用连接池）
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

// connectDirect 直接建立SSH连接（不使用连接池）
func (c *SSHClient) connectDirect() error {
	var authMethods []ssh.AuthMethod

	if c.config.KeyPath != "" {
		if key, err := os.ReadFile(c.config.KeyPath); err == nil {
			signer, err := ssh.ParsePrivateKey(key)
			if err == nil {
				authMethods = append(authMethods, ssh.PublicKeys(signer))
				log.Printf("Using SSH key: %s", c.config.KeyPath)
			} else {
				log.Printf("Warning: failed to parse SSH key: %v", err)
			}
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
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
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

// ExecuteCommand 执行命令
func (c *SSHClient) ExecuteCommand() error {
	if c.config.SafetyCheck && !c.config.Force {
		if err := ValidateCommand(c.config.Command); err != nil {
			return err
		}
	} else if c.config.Force {
		log.Println("⚠️  安全检查已跳过 (--force 模式)")
	}

	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	if c.config.Password != "" && strings.Contains(c.config.Command, "sudo") {
		return c.executeInteractive(session)
	}

	return c.executeWithPTY(session)
}

// ExecuteCommandWithOutput 执行命令并返回输出
func (c *SSHClient) ExecuteCommandWithOutput() (string, error) {
	if c.config.SafetyCheck && !c.config.Force {
		if err := ValidateCommand(c.config.Command); err != nil {
			return "", err
		}
	}

	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if c.config.Password != "" && strings.Contains(c.config.Command, "sudo") {
		actualCmd := strings.TrimPrefix(c.config.Command, "sudo ")
		actualCmd = strings.TrimSpace(actualCmd)
		finalCmd := fmt.Sprintf(`printf '%%s\n' '%s' | sudo -S %s`, c.config.Password, actualCmd)

		if err := session.Run(finalCmd); err != nil {
			return "", fmt.Errorf("command failed: %w\nStderr: %s", err, stderr.String())
		}
	} else {
		if err := session.Run(c.config.Command); err != nil {
			return "", fmt.Errorf("command failed: %w\nStderr: %s", err, stderr.String())
		}
	}

	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\n--- STDERR ---\n" + stderr.String()
	}

	return output, nil
}

// executeWithPTY 使用PTY执行命令
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

// executeNormal 执行普通命令（不使用PTY）
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

// executeInteractive 执行交互式命令（支持自动输入sudo密码）
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

// ExecuteSftp 执行SFTP操作
func (c *SSHClient) ExecuteSftp() error {
	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftpClient.Close()
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

func (c *SSHClient) uploadFile() error {
	localFile, err := os.Open(c.config.LocalPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer localFile.Close()

	remoteFile, err := c.sftpClient.Create(c.config.RemotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file: %w", err)
	}
	defer remoteFile.Close()

	log.Printf("Uploading: %s → %s", c.config.LocalPath, c.config.RemotePath)

	written, err := io.Copy(remoteFile, localFile)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	log.Printf("✓ Uploaded %d bytes successfully", written)
	return nil
}

func (c *SSHClient) downloadFile() error {
	remoteFile, err := c.sftpClient.Open(c.config.RemotePath)
	if err != nil {
		return fmt.Errorf("failed to open remote file: %w", err)
	}
	defer remoteFile.Close()

	localFile, err := os.Create(c.config.LocalPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer localFile.Close()

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

// Close 关闭连接（释放回连接池）
func (c *SSHClient) Close() error {
	if c.config != nil {
		pool := GetConnectionPool()
		pool.ReleaseConnection(c.config)
	}
	return nil
}

// ForceClose 强制关闭连接（不释放回连接池）
func (c *SSHClient) ForceClose() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}
