package sshclient

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"
)

// ExecuteScript executes a local script file
// 1. Upload script to remote temp directory
// 2. Add execute permission
// 3. Execute script
// 4. Clean up temp file
func (c *SSHClient) ExecuteScript(localScriptPath string) (output string, err error) {
	// 1. Check if local script exists
	if _, statErr := os.Stat(localScriptPath); statErr != nil {
		return "", fmt.Errorf("local script not found: %w", statErr)
	}

	// 2. Read script content
	scriptContent, err := os.ReadFile(localScriptPath) //nolint:gosec // G304: script path is provided by user
	if err != nil {
		return "", fmt.Errorf("failed to read script: %w", err)
	}

	// 3. Generate remote temp file path
	scriptName := filepath.Base(localScriptPath)
	timestamp := time.Now().Unix()
	remotePath := fmt.Sprintf("/tmp/sshx-script-%d-%s", timestamp, scriptName)

	// 4. Ensure SFTP client is available
	if c.sftpClient == nil {
		sftpClient, sftpErr := sftp.NewClient(c.client)
		if sftpErr != nil {
			return "", fmt.Errorf("failed to create SFTP client: %w", sftpErr)
		}
		c.sftpClient = sftpClient
		defer CloseIgnore(&err, c.sftpClient, io.EOF)
	}

	// 5. Upload script to remote
	remoteFile, err := c.sftpClient.Create(remotePath)
	if err != nil {
		return "", fmt.Errorf("failed to create remote file: %w", err)
	}

	if _, err = remoteFile.Write(scriptContent); err != nil {
		if closeErr := remoteFile.Close(); closeErr != nil {
			// Ignore close error when write already failed
			_ = closeErr
		}
		return "", fmt.Errorf("failed to write script: %w", err)
	}
	if err = remoteFile.Close(); err != nil {
		return "", fmt.Errorf("failed to close remote file: %w", err)
	}

	// 6. Add execute permission
	if err := c.sftpClient.Chmod(remotePath, 0755); err != nil {
		return "", fmt.Errorf("failed to chmod script: %w", err)
	}

	// 7. Execute script
	output, execErr := c.executeRemoteScript(remotePath)

	// 8. Clean up temp file (regardless of execution result)
	cleanupCmd := fmt.Sprintf("rm -f %s", remotePath)
	if cleanupErr := c.executeSimpleCommand(cleanupCmd); cleanupErr != nil {
		// Log cleanup error but don't fail the operation
		_ = cleanupErr // Cleanup is best-effort
	}

	// 9. Return execution result
	if execErr != nil {
		return output, fmt.Errorf("script execution failed: %w", execErr)
	}

	return output, nil
}

// executeRemoteScript executes a remote script
func (c *SSHClient) executeRemoteScript(remotePath string) (output string, err error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer CloseIgnore(&err, session, io.EOF)

	// Detect script type and execute
	var command string
	if strings.HasSuffix(remotePath, ".sh") || strings.HasSuffix(remotePath, ".bash") {
		command = fmt.Sprintf("bash %s", remotePath)
	} else if strings.HasSuffix(remotePath, ".py") || strings.HasSuffix(remotePath, ".python") {
		command = fmt.Sprintf("python3 %s", remotePath)
	} else if strings.HasSuffix(remotePath, ".pl") {
		command = fmt.Sprintf("perl %s", remotePath)
	} else if strings.HasSuffix(remotePath, ".rb") {
		command = fmt.Sprintf("ruby %s", remotePath)
	} else {
		// Default to shell script
		command = fmt.Sprintf("bash %s", remotePath)
	}

	outputBytes, err := session.CombinedOutput(command)
	output = string(outputBytes)
	return output, err
}

// executeSimpleCommand executes a simple command (used for cleanup, etc.)
func (c *SSHClient) executeSimpleCommand(command string) (err error) {
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	defer CloseIgnore(&err, session, io.EOF)

	return session.Run(command)
}

// ExecuteScriptWithArgs executes a script with arguments
func (c *SSHClient) ExecuteScriptWithArgs(localScriptPath string, args []string) (output string, err error) {
	// 1. Check if local script exists
	if _, statErr := os.Stat(localScriptPath); statErr != nil {
		return "", fmt.Errorf("local script not found: %w", statErr)
	}

	// 2. Read script content
	scriptContent, err := os.ReadFile(localScriptPath) //nolint:gosec // G304: script path is provided by user
	if err != nil {
		return "", fmt.Errorf("failed to read script: %w", err)
	}

	// 3. Generate remote temp file path
	scriptName := filepath.Base(localScriptPath)
	timestamp := time.Now().Unix()
	remotePath := fmt.Sprintf("/tmp/sshx-script-%d-%s", timestamp, scriptName)

	// 4. Ensure SFTP client is available
	if c.sftpClient == nil {
		sftpClient, sftpErr := sftp.NewClient(c.client)
		if sftpErr != nil {
			return "", fmt.Errorf("failed to create SFTP client: %w", sftpErr)
		}
		c.sftpClient = sftpClient
		defer CloseIgnore(&err, c.sftpClient, io.EOF)
	}

	// 5. Upload script
	remoteFile, err := c.sftpClient.Create(remotePath)
	if err != nil {
		return "", fmt.Errorf("failed to create remote file: %w", err)
	}
	defer CloseIgnore(&err, remoteFile, io.EOF)

	if _, err = remoteFile.Write(scriptContent); err != nil {
		return "", fmt.Errorf("failed to write script: %w", err)
	}

	// 6. Add execute permission
	if err = c.sftpClient.Chmod(remotePath, 0755); err != nil {
		return "", fmt.Errorf("failed to chmod script: %w", err)
	}

	// 7. Build command with arguments
	var command string
	interpreter := c.detectInterpreter(remotePath)
	escapedArgs := make([]string, len(args))
	for i, arg := range args {
		// Simple argument escaping
		escapedArgs[i] = fmt.Sprintf("'%s'", strings.ReplaceAll(arg, "'", "'\\''"))
	}

	command = fmt.Sprintf("%s %s %s", interpreter, remotePath, strings.Join(escapedArgs, " "))

	// 8. Execute script
	session, err := c.client.NewSession()
	if err != nil {
		// Try to clean up on error
		if cleanupErr := c.executeSimpleCommand(fmt.Sprintf("rm -f %s", remotePath)); cleanupErr != nil {
			_ = cleanupErr // Cleanup is best-effort
		}
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer CloseIgnore(&err, session, io.EOF)

	outputBytes, execErr := session.CombinedOutput(command)
	output = string(outputBytes)

	// 9. Clean up temp file
	if cleanupErr := c.executeSimpleCommand(fmt.Sprintf("rm -f %s", remotePath)); cleanupErr != nil {
		_ = cleanupErr // Cleanup is best-effort
	}

	// 10. Return execution result
	if execErr != nil {
		return output, fmt.Errorf("script execution failed: %w", execErr)
	}

	return output, nil
}

// detectInterpreter detects the script interpreter
func (c *SSHClient) detectInterpreter(remotePath string) string {
	if strings.HasSuffix(remotePath, ".sh") || strings.HasSuffix(remotePath, ".bash") {
		return "bash"
	} else if strings.HasSuffix(remotePath, ".py") || strings.HasSuffix(remotePath, ".python") {
		return "python3"
	} else if strings.HasSuffix(remotePath, ".pl") {
		return "perl"
	} else if strings.HasSuffix(remotePath, ".rb") {
		return "ruby"
	}
	return "bash" // Default
}
