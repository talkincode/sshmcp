package sshclient

import (
	"fmt"
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
func (c *SSHClient) ExecuteScript(localScriptPath string) (string, error) {
	// 1. Check if local script exists
	if _, err := os.Stat(localScriptPath); err != nil {
		return "", fmt.Errorf("local script not found: %w", err)
	}

	// 2. Read script content
	scriptContent, err := os.ReadFile(localScriptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read script: %w", err)
	}

	// 3. Generate remote temp file path
	scriptName := filepath.Base(localScriptPath)
	timestamp := time.Now().Unix()
	remotePath := fmt.Sprintf("/tmp/sshx-script-%d-%s", timestamp, scriptName)

	// 4. Ensure SFTP client is available
	if c.sftpClient == nil {
		sftpClient, err := sftp.NewClient(c.client)
		if err != nil {
			return "", fmt.Errorf("failed to create SFTP client: %w", err)
		}
		c.sftpClient = sftpClient
		defer c.sftpClient.Close()
	}

	// 5. Upload script to remote
	remoteFile, err := c.sftpClient.Create(remotePath)
	if err != nil {
		return "", fmt.Errorf("failed to create remote file: %w", err)
	}

	if _, err := remoteFile.Write(scriptContent); err != nil {
		remoteFile.Close()
		return "", fmt.Errorf("failed to write script: %w", err)
	}
	remoteFile.Close()

	// 6. Add execute permission
	if err := c.sftpClient.Chmod(remotePath, 0755); err != nil {
		return "", fmt.Errorf("failed to chmod script: %w", err)
	}

	// 7. Execute script
	output, execErr := c.executeRemoteScript(remotePath)

	// 8. Clean up temp file (regardless of execution result)
	cleanupCmd := fmt.Sprintf("rm -f %s", remotePath)
	c.executeSimpleCommand(cleanupCmd)

	// 9. Return execution result
	if execErr != nil {
		return output, fmt.Errorf("script execution failed: %w", execErr)
	}

	return output, nil
}

// executeRemoteScript executes a remote script
func (c *SSHClient) executeRemoteScript(remotePath string) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

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

	output, err := session.CombinedOutput(command)
	return string(output), err
}

// executeSimpleCommand executes a simple command (used for cleanup, etc.)
func (c *SSHClient) executeSimpleCommand(command string) error {
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	return session.Run(command)
}

// ExecuteScriptWithArgs executes a script with arguments
func (c *SSHClient) ExecuteScriptWithArgs(localScriptPath string, args []string) (string, error) {
	// 1. Check if local script exists
	if _, err := os.Stat(localScriptPath); err != nil {
		return "", fmt.Errorf("local script not found: %w", err)
	}

	// 2. Read script content
	scriptContent, err := os.ReadFile(localScriptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read script: %w", err)
	}

	// 3. Generate remote temp file path
	scriptName := filepath.Base(localScriptPath)
	timestamp := time.Now().Unix()
	remotePath := fmt.Sprintf("/tmp/sshx-script-%d-%s", timestamp, scriptName)

	// 4. Ensure SFTP client is available
	if c.sftpClient == nil {
		sftpClient, err := sftp.NewClient(c.client)
		if err != nil {
			return "", fmt.Errorf("failed to create SFTP client: %w", err)
		}
		c.sftpClient = sftpClient
		defer c.sftpClient.Close()
	}

	// 5. Upload script
	remoteFile, err := c.sftpClient.Create(remotePath)
	if err != nil {
		return "", fmt.Errorf("failed to create remote file: %w", err)
	}

	if _, err := remoteFile.Write(scriptContent); err != nil {
		remoteFile.Close()
		return "", fmt.Errorf("failed to write script: %w", err)
	}
	remoteFile.Close()

	// 6. Add execute permission
	if err := c.sftpClient.Chmod(remotePath, 0755); err != nil {
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
		c.executeSimpleCommand(fmt.Sprintf("rm -f %s", remotePath))
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, execErr := session.CombinedOutput(command)

	// 9. Clean up temp file
	c.executeSimpleCommand(fmt.Sprintf("rm -f %s", remotePath))

	// 10. Return execution result
	if execErr != nil {
		return string(output), fmt.Errorf("script execution failed: %w", execErr)
	}

	return string(output), nil
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
