#!/bin/bash
# MCP 调试测试 - 临时启用日志查看连接详情

# 临时修改 main.go 启用日志
cd /Volumes/ExtDISK/github/sshx

# 备份原文件
cp cmd/sshx/main.go cmd/sshx/main.go.backup
cp internal/app/app.go internal/app/app.go.backup

# 注释掉 log.SetOutput(io.Discard)
sed -i '' 's/log.SetOutput(io.Discard)/\/\/ log.SetOutput(io.Discard)/' internal/app/app.go

# 重新构建
make build

# 运行测试
echo '{"jsonrpc":"2.0","id":99,"method":"tools/call","params":{"name":"ssh_execute","arguments":{"host":"10.201.15.192","command":"whoami","user":"master","port":"22"}}}' | ./bin/sshx mcp-stdio

# 恢复原文件
mv cmd/sshx/main.go.backup cmd/sshx/main.go
mv internal/app/app.go.backup internal/app/app.go

# 重新构建
make build
