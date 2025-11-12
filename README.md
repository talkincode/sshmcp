# sshx

`sshx` 提供一个支持无障碍 SSH 命令行客户端，同时实现 MCP（Model Context Protocol）接口，方便 AI 助手调用远程 SSH/SFTP 功能。

## 项目结构

- `cmd/sshx`：主二进制入口，负责命令行参数解析、MCP 模式启动及密码管理功能。
- `internal/sshclient`：核心 SSH/SFTP/脚本执行逻辑、命令安全检测及连接池封装。
- `internal/mcp`：MCP stdio 服务器实现，暴露 SSH/SFTP/脚本等工具给外部工具（例如 AI 助手）。

## 主要能力

1. 跨平台 SSH/SFTP 操作（支持 sudo 自动填充）。
2. 密码管理（Keychain / Secret Service / Credential Manager）。
3. MCP stdio 模式，可被 AI 助手调用。
4. 连接池、脚本执行与命令安全校验。

## 快速开始

```bash
# 构建命令行工具
go build -o bin/sshx ./cmd/sshx

# 执行命令
./bin/sshx -h=192.168.1.100 "uptime"

# 启动 MCP stdio 模式
./bin/sshx mcp-stdio
```

## 开发

- `go test ./...` 运行单元测试（目前包括命令安全校验）。
- 代码调整后请运行 `gofmt`，保持 Go 代码风格一致。

## 依赖管理

使用 Go Modules (`go.mod`) 管理第三方依赖。`go test ./...` 会自动下载所需模块。
