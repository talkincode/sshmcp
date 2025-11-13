package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/talkincode/sshmcp/internal/sshclient"
)

// MCP Protocol types
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MCP Tool definitions
type MCPTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

type ToolSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default,omitempty"`
}

// MCPServer represents an MCP server instance
type MCPServer struct {
	stdin  *bufio.Reader
	stdout io.Writer
	tools  []MCPTool
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer() *MCPServer {
	return &MCPServer{
		stdin:  bufio.NewReader(os.Stdin),
		stdout: os.Stdout,
		tools:  defineMCPTools(),
	}
}

// defineMCPTools defines all available MCP tools
func defineMCPTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "ssh_execute",
			Description: "Execute a command on remote server via SSH. Supports sudo with automatic password handling.",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"host": {
						Type:        "string",
						Description: "Remote host address (IP or hostname)",
					},
					"command": {
						Type:        "string",
						Description: "Command to execute on remote server",
					},
					"port": {
						Type:        "string",
						Description: "SSH port",
						Default:     "22",
					},
					"user": {
						Type:        "string",
						Description: "SSH username",
						Default:     "master",
					},
					"key_path": {
						Type:        "string",
						Description: "Path to SSH private key",
					},
					"sudo_key": {
						Type:        "string",
						Description: "Key name for sudo password",
						Default:     "master",
					},
					"force": {
						Type:        "string",
						Description: "Force execution, bypass safety checks (use with caution!)",
						Enum:        []string{"true", "false"},
						Default:     "false",
					},
				},
				Required: []string{"host", "command"},
			},
		},
		{
			Name:        "sftp_upload",
			Description: "Upload a file to remote server via SFTP",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"host": {
						Type:        "string",
						Description: "Remote host address",
					},
					"local_path": {
						Type:        "string",
						Description: "Local file path to upload",
					},
					"remote_path": {
						Type:        "string",
						Description: "Remote destination path",
					},
					"port": {
						Type:        "string",
						Description: "SSH port",
						Default:     "22",
					},
					"user": {
						Type:        "string",
						Description: "SSH username",
						Default:     "master",
					},
				},
				Required: []string{"host", "local_path", "remote_path"},
			},
		},
		{
			Name:        "sftp_download",
			Description: "Download a file from remote server via SFTP",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"host": {
						Type:        "string",
						Description: "Remote host address",
					},
					"remote_path": {
						Type:        "string",
						Description: "Remote file path to download",
					},
					"local_path": {
						Type:        "string",
						Description: "Local destination path",
					},
					"port": {
						Type:        "string",
						Description: "SSH port",
						Default:     "22",
					},
					"user": {
						Type:        "string",
						Description: "SSH username",
						Default:     "master",
					},
				},
				Required: []string{"host", "remote_path", "local_path"},
			},
		},
		{
			Name:        "sftp_list",
			Description: "List directory contents on remote server via SFTP",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"host": {
						Type:        "string",
						Description: "Remote host address",
					},
					"remote_path": {
						Type:        "string",
						Description: "Remote directory path to list",
						Default:     ".",
					},
					"port": {
						Type:        "string",
						Description: "SSH port",
						Default:     "22",
					},
					"user": {
						Type:        "string",
						Description: "SSH username",
						Default:     "master",
					},
				},
				Required: []string{"host"},
			},
		},
		{
			Name:        "sftp_mkdir",
			Description: "Create a directory on remote server via SFTP",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"host": {
						Type:        "string",
						Description: "Remote host address",
					},
					"remote_path": {
						Type:        "string",
						Description: "Remote directory path to create",
					},
					"port": {
						Type:        "string",
						Description: "SSH port",
						Default:     "22",
					},
					"user": {
						Type:        "string",
						Description: "SSH username",
						Default:     "master",
					},
				},
				Required: []string{"host", "remote_path"},
			},
		},
		{
			Name:        "sftp_remove",
			Description: "Remove a file or directory on remote server via SFTP",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"host": {
						Type:        "string",
						Description: "Remote host address",
					},
					"remote_path": {
						Type:        "string",
						Description: "Remote file or directory path to remove",
					},
					"port": {
						Type:        "string",
						Description: "SSH port",
						Default:     "22",
					},
					"user": {
						Type:        "string",
						Description: "SSH username",
						Default:     "master",
					},
				},
				Required: []string{"host", "remote_path"},
			},
		},
		{
			Name:        "script_execute",
			Description: "Upload and execute a local script file on remote server. Automatically detects script type (bash/python/perl/ruby) and cleans up after execution.",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"host": {
						Type:        "string",
						Description: "Remote host address",
					},
					"script_path": {
						Type:        "string",
						Description: "Local script file path to upload and execute",
					},
					"args": {
						Type:        "string",
						Description: "Optional arguments to pass to the script (space-separated)",
					},
					"port": {
						Type:        "string",
						Description: "SSH port",
						Default:     "22",
					},
					"user": {
						Type:        "string",
						Description: "SSH username",
						Default:     "master",
					},
				},
				Required: []string{"host", "script_path"},
			},
		},
		{
			Name:        "pool_stats",
			Description: "Get SSH connection pool statistics (active/idle connections, health check interval, etc.)",
			InputSchema: ToolSchema{
				Type:       "object",
				Properties: map[string]Property{},
				Required:   []string{},
			},
		},
	}
}

// Start starts the MCP server and handles JSON-RPC communication
func (s *MCPServer) Start() error {
	// In MCP stdio mode, log output is disabled to avoid interfering with JSON-RPC communication
	// log is set to io.Discard in main.go

	for {
		line, err := s.stdin.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("failed to read from stdin: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var req MCPRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			s.sendError(nil, -32700, "Parse error", err.Error())
			continue
		}

		s.handleRequest(&req)
	}
}

// handleRequest 处理MCP请求
func (s *MCPServer) handleRequest(req *MCPRequest) {
	switch req.Method {
	case "initialize":
		s.handleInitialize(req)
	case "tools/list":
		s.handleToolsList(req)
	case "tools/call":
		s.handleToolsCall(req)
	case "shutdown":
		s.sendResponse(req.ID, map[string]interface{}{})
		os.Exit(0)
	default:
		s.sendError(req.ID, -32601, "Method not found", req.Method)
	}
}

// handleInitialize 处理初始化请求
func (s *MCPServer) handleInitialize(req *MCPRequest) {
	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "sshx-mcp-server",
			"version": "1.0.0",
		},
	}
	s.sendResponse(req.ID, result)
}

// handleToolsList 处理工具列表请求
func (s *MCPServer) handleToolsList(req *MCPRequest) {
	result := map[string]interface{}{
		"tools": s.tools,
	}
	s.sendResponse(req.ID, result)
}

// handleToolsCall 处理工具调用请求
func (s *MCPServer) handleToolsCall(req *MCPRequest) {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(req.ID, -32602, "Invalid params", err.Error())
		return
	}

	result, err := s.executeTool(params.Name, params.Arguments)
	if err != nil {
		// 构建更详细的错误消息
		errorMsg := fmt.Sprintf("Tool '%s' execution failed: %s", params.Name, err.Error())
		s.sendError(req.ID, -32000, errorMsg, map[string]interface{}{
			"tool":      params.Name,
			"arguments": params.Arguments,
			"error":     err.Error(),
		})
		return
	}

	s.sendResponse(req.ID, map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": result,
			},
		},
	})
}

// executeTool 执行工具
func (s *MCPServer) executeTool(name string, args map[string]interface{}) (string, error) {
	// 构建配置
	config := &sshclient.Config{}

	// 通用参数
	if host, ok := args["host"].(string); ok && host != "" {
		config.Host = host
	} else {
		// 如果没有提供 host，使用默认值（用于测试/验证）
		config.Host = "0.0.0.0"
	}

	if port, ok := args["port"].(string); ok {
		config.Port = port
	} else {
		config.Port = sshclient.DefaultSSHPort
	}
	if user, ok := args["user"].(string); ok {
		config.User = user
	} else {
		config.User = sshclient.DefaultSSHUser
	}
	if keyPath, ok := args["key_path"].(string); ok {
		config.KeyPath = keyPath
	}

	switch name {
	case "ssh_execute":
		return s.executeSSH(config, args)
	case "sftp_upload":
		return s.executeSftpUpload(config, args)
	case "sftp_download":
		return s.executeSftpDownload(config, args)
	case "sftp_list":
		return s.executeSftpList(config, args)
	case "sftp_mkdir":
		return s.executeSftpMkdir(config, args)
	case "sftp_remove":
		return s.executeSftpRemove(config, args)
	case "script_execute":
		return s.executeScript(config, args)
	case "pool_stats":
		return s.getPoolStats()
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

// executeSSH 执行SSH命令
func (s *MCPServer) executeSSH(config *sshclient.Config, args map[string]interface{}) (output string, err error) {
	// 检查是否为测试调用(使用默认 host)
	if config.Host == "0.0.0.0" {
		return "MCP Tool: ssh_execute\nStatus: Ready\nNote: Please provide a valid 'host' parameter to execute SSH commands.\nExample: {\"host\": \"192.168.1.100\", \"command\": \"uptime\"}", nil
	}

	command, ok := args["command"].(string)
	if !ok {
		return "", fmt.Errorf("command is required")
	}
	config.Command = command

	// 默认启用安全检查
	config.SafetyCheck = true

	// 处理 force 参数
	if force, ok := args["force"].(string); ok {
		config.Force = force == "true"
	} else {
		config.Force = false
	}

	// 处理 sudo
	if sudoKey, ok := args["sudo_key"].(string); ok {
		config.SudoKey = sudoKey
	} else {
		config.SudoKey = sshclient.DefaultSudoKey
	}

	// 如果命令包含 sudo，尝试获取密码
	if strings.Contains(command, "sudo") && config.SudoKey != "" {
		password, pwdErr := sshclient.GetSudoPassword(config.SudoKey)
		if pwdErr != nil {
			// 静默忽略，MCP 模式下不输出警告
		} else {
			config.Password = password
		}
	}

	client, err := sshclient.NewSSHClient(config)
	if err != nil {
		return "", fmt.Errorf("failed to create SSH client: %w", err)
	}
	defer sshclient.CloseIgnore(&err, client)

	if err = client.Connect(); err != nil {
		return "", fmt.Errorf("failed to connect: %w", err)
	}

	// 使用新的 ExecuteCommandWithOutput 方法直接获取输出
	output, err = client.ExecuteCommandWithOutput()
	if err != nil {
		// 返回详细的错误信息,包含命令和完整的错误详情
		return "", fmt.Errorf("failed to execute command '%s' on %s@%s:%s - %w",
			command, config.User, config.Host, config.Port, err)
	}

	return output, nil
}

// executeSftpUpload 执行SFTP上传
func (s *MCPServer) executeSftpUpload(config *sshclient.Config, args map[string]interface{}) (result string, err error) {
	// 检查是否为测试调用
	if config.Host == "0.0.0.0" {
		return "MCP Tool: sftp_upload\nStatus: Ready\nNote: Please provide valid parameters to upload files.\nExample: {\"host\": \"192.168.1.100\", \"local_path\": \"/local/file.txt\", \"remote_path\": \"/remote/file.txt\"}", nil
	}

	localPath, ok := args["local_path"].(string)
	if !ok {
		return "", fmt.Errorf("local_path is required")
	}
	remotePath, ok := args["remote_path"].(string)
	if !ok {
		return "", fmt.Errorf("remote_path is required")
	}

	config.Mode = "sftp"
	config.SftpAction = "upload"
	config.LocalPath = localPath
	config.RemotePath = remotePath

	client, err := sshclient.NewSSHClient(config)
	if err != nil {
		return "", err
	}
	defer sshclient.CloseIgnore(&err, client)

	if err := client.Connect(); err != nil {
		return "", err
	}

	if err := client.ExecuteSftp(); err != nil {
		return "", err
	}

	return fmt.Sprintf("File uploaded successfully: %s -> %s", localPath, remotePath), nil
}

// executeSftpDownload 执行SFTP下载
func (s *MCPServer) executeSftpDownload(config *sshclient.Config, args map[string]interface{}) (result string, err error) {
	// 检查是否为测试调用
	if config.Host == "0.0.0.0" {
		return "MCP Tool: sftp_download\nStatus: Ready\nNote: Please provide valid parameters to download files.\nExample: {\"host\": \"192.168.1.100\", \"remote_path\": \"/remote/file.txt\", \"local_path\": \"/local/file.txt\"}", nil
	}

	remotePath, ok := args["remote_path"].(string)
	if !ok {
		return "", fmt.Errorf("remote_path is required")
	}
	localPath, ok := args["local_path"].(string)
	if !ok {
		return "", fmt.Errorf("local_path is required")
	}

	config.Mode = "sftp"
	config.SftpAction = "download"
	config.LocalPath = localPath
	config.RemotePath = remotePath

	client, err := sshclient.NewSSHClient(config)
	if err != nil {
		return "", err
	}
	defer sshclient.CloseIgnore(&err, client)

	if err := client.Connect(); err != nil {
		return "", err
	}

	if err := client.ExecuteSftp(); err != nil {
		return "", err
	}

	return fmt.Sprintf("File downloaded successfully: %s -> %s", remotePath, localPath), nil
}

// executeSftpList 执行SFTP列表
func (s *MCPServer) executeSftpList(config *sshclient.Config, args map[string]interface{}) (result string, err error) {
	// 检查是否为测试调用
	if config.Host == "0.0.0.0" {
		return "MCP Tool: sftp_list\nStatus: Ready\nNote: Please provide a valid 'host' parameter to list files.\nExample: {\"host\": \"192.168.1.100\", \"remote_path\": \"/var/log\"}", nil
	}

	remotePath := "."
	if path, ok := args["remote_path"].(string); ok {
		remotePath = path
	}

	config.Mode = "sftp"
	config.SftpAction = "list"
	config.RemotePath = remotePath

	client, err := sshclient.NewSSHClient(config)
	if err != nil {
		return "", err
	}
	defer sshclient.CloseIgnore(&err, client)

	if err = client.Connect(); err != nil {
		return "", err
	}

	// 捕获输出
	var output strings.Builder
	oldStdout := os.Stdout
	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		return "", fmt.Errorf("failed to create pipe: %w", pipeErr)
	}
	os.Stdout = w

	errChan := make(chan error, 1)
	go func() {
		errChan <- client.ExecuteSftp()
	}()

	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			output.WriteString(scanner.Text() + "\n")
		}
	}()

	err = <-errChan
	if closeErr := w.Close(); closeErr != nil {
		// Log best-effort close error
		_ = closeErr
	}
	os.Stdout = oldStdout

	if err != nil {
		return "", err
	}

	return output.String(), nil
}

// executeSftpMkdir 执行SFTP创建目录
func (s *MCPServer) executeSftpMkdir(config *sshclient.Config, args map[string]interface{}) (result string, err error) {
	// 检查是否为测试调用
	if config.Host == "0.0.0.0" {
		return "MCP Tool: sftp_mkdir\nStatus: Ready\nNote: Please provide valid parameters to create directories.\nExample: {\"host\": \"192.168.1.100\", \"remote_path\": \"/tmp/newdir\"}", nil
	}

	remotePath, ok := args["remote_path"].(string)
	if !ok {
		return "", fmt.Errorf("remote_path is required")
	}

	config.Mode = "sftp"
	config.SftpAction = "mkdir"
	config.RemotePath = remotePath

	client, err := sshclient.NewSSHClient(config)
	if err != nil {
		return "", err
	}
	defer sshclient.CloseIgnore(&err, client)

	if err := client.Connect(); err != nil {
		return "", err
	}

	if err := client.ExecuteSftp(); err != nil {
		return "", err
	}

	return fmt.Sprintf("Directory created: %s", remotePath), nil
}

// executeSftpRemove 执行SFTP删除
func (s *MCPServer) executeSftpRemove(config *sshclient.Config, args map[string]interface{}) (result string, err error) {
	// 检查是否为测试调用
	if config.Host == "0.0.0.0" {
		return "MCP Tool: sftp_remove\nStatus: Ready\nNote: Please provide valid parameters to remove files/directories.\nExample: {\"host\": \"192.168.1.100\", \"remote_path\": \"/tmp/oldfile.txt\"}", nil
	}

	remotePath, ok := args["remote_path"].(string)
	if !ok {
		return "", fmt.Errorf("remote_path is required")
	}

	config.Mode = "sftp"
	config.SftpAction = "remove"
	config.RemotePath = remotePath

	client, err := sshclient.NewSSHClient(config)
	if err != nil {
		return "", err
	}
	defer sshclient.CloseIgnore(&err, client)

	if err := client.Connect(); err != nil {
		return "", err
	}

	if err := client.ExecuteSftp(); err != nil {
		return "", err
	}

	return fmt.Sprintf("Removed: %s", remotePath), nil
}

// sendResponse 发送响应
func (s *MCPServer) sendResponse(id interface{}, result interface{}) {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	s.writeJSON(resp)
}

// sendError 发送错误
func (s *MCPServer) sendError(id interface{}, code int, message string, data interface{}) {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	s.writeJSON(resp)
}

// writeJSON 写入JSON到stdout
func (s *MCPServer) writeJSON(v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		// 静默忽略，MCP 模式下不输出日志
		return
	}
	if _, writeErr := fmt.Fprintf(s.stdout, "%s\n", data); writeErr != nil {
		// Best-effort write, ignore error
		_ = writeErr
	}
}

// executeScript 执行脚本
func (s *MCPServer) executeScript(config *sshclient.Config, args map[string]interface{}) (output string, err error) {
	// 检查是否为测试调用
	if config.Host == "0.0.0.0" {
		return "MCP Tool: script_execute\nStatus: Ready\nNote: Please provide valid parameters to execute scripts.\nExample: {\"host\": \"192.168.1.100\", \"script_path\": \"/path/to/script.sh\"}", nil
	}

	scriptPath, ok := args["script_path"].(string)
	if !ok {
		return "", fmt.Errorf("script_path is required")
	}

	client, err := sshclient.NewSSHClient(config)
	if err != nil {
		return "", fmt.Errorf("failed to create SSH client: %w", err)
	}
	defer sshclient.CloseIgnore(&err, client)

	if err = client.Connect(); err != nil {
		return "", fmt.Errorf("failed to connect: %w", err)
	}

	// 检查是否有参数
	if argsStr, ok := args["args"].(string); ok && argsStr != "" {
		// 分割参数
		scriptArgs := strings.Fields(argsStr)
		output, err = client.ExecuteScriptWithArgs(scriptPath, scriptArgs)
	} else {
		output, err = client.ExecuteScript(scriptPath)
	}

	if err != nil {
		return "", fmt.Errorf("script execution failed: %w\nOutput: %s", err, output)
	}

	return output, nil
}

// getPoolStats 获取连接池统计
func (s *MCPServer) getPoolStats() (string, error) {
	pool := sshclient.GetConnectionPool()
	stats := pool.Stats()

	// 格式化输出
	var output strings.Builder
	output.WriteString("SSH Connection Pool Statistics:\n")
	output.WriteString("================================\n")
	output.WriteString(fmt.Sprintf("Total Connections:  %v\n", stats["total_connections"]))
	output.WriteString(fmt.Sprintf("Active Connections: %v\n", stats["active_connections"]))
	output.WriteString(fmt.Sprintf("Idle Connections:   %v\n", stats["idle_connections"]))
	output.WriteString(fmt.Sprintf("Max Idle Duration:  %v\n", stats["max_idle_duration"]))
	output.WriteString(fmt.Sprintf("Health Check Interval: %v\n", stats["health_check_interval"]))

	return output.String(), nil
}
