package sshclient

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"
)

// ExecuteScript 执行本地脚本文件
// 1. 上传脚本到远程临时目录
// 2. 添加执行权限
// 3. 执行脚本
// 4. 清理临时文件
func (c *SSHClient) ExecuteScript(localScriptPath string) (string, error) {
	// 1. 检查本地脚本是否存在
	if _, err := os.Stat(localScriptPath); err != nil {
		return "", fmt.Errorf("local script not found: %w", err)
	}

	// 2. 读取脚本内容
	scriptContent, err := os.ReadFile(localScriptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read script: %w", err)
	}

	// 3. 生成远程临时文件路径
	scriptName := filepath.Base(localScriptPath)
	timestamp := time.Now().Unix()
	remotePath := fmt.Sprintf("/tmp/sshx-script-%d-%s", timestamp, scriptName)

	// 4. 确保有 SFTP 客户端
	if c.sftpClient == nil {
		sftpClient, err := sftp.NewClient(c.client)
		if err != nil {
			return "", fmt.Errorf("failed to create SFTP client: %w", err)
		}
		c.sftpClient = sftpClient
		defer c.sftpClient.Close()
	}

	// 5. 上传脚本到远程
	remoteFile, err := c.sftpClient.Create(remotePath)
	if err != nil {
		return "", fmt.Errorf("failed to create remote file: %w", err)
	}

	if _, err := remoteFile.Write(scriptContent); err != nil {
		remoteFile.Close()
		return "", fmt.Errorf("failed to write script: %w", err)
	}
	remoteFile.Close()

	// 6. 添加执行权限
	if err := c.sftpClient.Chmod(remotePath, 0755); err != nil {
		return "", fmt.Errorf("failed to chmod script: %w", err)
	}

	// 7. 执行脚本
	output, execErr := c.executeRemoteScript(remotePath)

	// 8. 清理临时文件（无论执行成功与否）
	cleanupCmd := fmt.Sprintf("rm -f %s", remotePath)
	c.executeSimpleCommand(cleanupCmd)

	// 9. 返回执行结果
	if execErr != nil {
		return output, fmt.Errorf("script execution failed: %w", execErr)
	}

	return output, nil
}

// executeRemoteScript 执行远程脚本
func (c *SSHClient) executeRemoteScript(remotePath string) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// 检测脚本类型并执行
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
		// 默认当作 shell 脚本
		command = fmt.Sprintf("bash %s", remotePath)
	}

	output, err := session.CombinedOutput(command)
	return string(output), err
}

// executeSimpleCommand 执行简单命令（用于清理等操作）
func (c *SSHClient) executeSimpleCommand(command string) error {
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	return session.Run(command)
}

// ExecuteScriptWithArgs 执行脚本并传递参数
func (c *SSHClient) ExecuteScriptWithArgs(localScriptPath string, args []string) (string, error) {
	// 1. 检查本地脚本是否存在
	if _, err := os.Stat(localScriptPath); err != nil {
		return "", fmt.Errorf("local script not found: %w", err)
	}

	// 2. 读取脚本内容
	scriptContent, err := os.ReadFile(localScriptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read script: %w", err)
	}

	// 3. 生成远程临时文件路径
	scriptName := filepath.Base(localScriptPath)
	timestamp := time.Now().Unix()
	remotePath := fmt.Sprintf("/tmp/sshx-script-%d-%s", timestamp, scriptName)

	// 4. 确保有 SFTP 客户端
	if c.sftpClient == nil {
		sftpClient, err := sftp.NewClient(c.client)
		if err != nil {
			return "", fmt.Errorf("failed to create SFTP client: %w", err)
		}
		c.sftpClient = sftpClient
		defer c.sftpClient.Close()
	}

	// 5. 上传脚本
	remoteFile, err := c.sftpClient.Create(remotePath)
	if err != nil {
		return "", fmt.Errorf("failed to create remote file: %w", err)
	}

	if _, err := remoteFile.Write(scriptContent); err != nil {
		remoteFile.Close()
		return "", fmt.Errorf("failed to write script: %w", err)
	}
	remoteFile.Close()

	// 6. 添加执行权限
	if err := c.sftpClient.Chmod(remotePath, 0755); err != nil {
		return "", fmt.Errorf("failed to chmod script: %w", err)
	}

	// 7. 构建带参数的命令
	var command string
	interpreter := c.detectInterpreter(remotePath)
	escapedArgs := make([]string, len(args))
	for i, arg := range args {
		// 简单的参数转义
		escapedArgs[i] = fmt.Sprintf("'%s'", strings.ReplaceAll(arg, "'", "'\\''"))
	}

	command = fmt.Sprintf("%s %s %s", interpreter, remotePath, strings.Join(escapedArgs, " "))

	// 8. 执行脚本
	session, err := c.client.NewSession()
	if err != nil {
		c.executeSimpleCommand(fmt.Sprintf("rm -f %s", remotePath))
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, execErr := session.CombinedOutput(command)

	// 9. 清理临时文件
	c.executeSimpleCommand(fmt.Sprintf("rm -f %s", remotePath))

	// 10. 返回执行结果
	if execErr != nil {
		return string(output), fmt.Errorf("script execution failed: %w", execErr)
	}

	return string(output), nil
}

// detectInterpreter 检测脚本解释器
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
	return "bash" // 默认
}
