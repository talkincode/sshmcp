#!/bin/bash
# MCP 模式测试脚本
# 用于测试 sshx MCP 服务器的各种工具

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# 发送 MCP 请求并获取响应
send_mcp_request() {
    local request="$1"
    echo "$request" | ./bin/sshx mcp-stdio
}

# 测试初始化
test_initialize() {
    log_info "测试 MCP 初始化..."
    
    local request='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}'
    
    local response=$(send_mcp_request "$request")
    
    if echo "$response" | grep -q "protocolVersion"; then
        log_success "初始化成功"
        echo "$response" | jq '.'
        return 0
    else
        log_error "初始化失败"
        echo "$response"
        return 1
    fi
}

# 测试工具列表
test_tools_list() {
    log_info "测试获取工具列表..."
    
    local request='{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'
    
    local response=$(send_mcp_request "$request")
    
    if echo "$response" | grep -q "ssh_execute"; then
        log_success "获取工具列表成功"
        echo "$response" | jq '.result.tools[] | {name: .name, description: .description}'
        return 0
    else
        log_error "获取工具列表失败"
        echo "$response"
        return 1
    fi
}

# 测试 SSH 执行（测试模式）
test_ssh_execute_test_mode() {
    log_info "测试 ssh_execute 工具（测试模式）..."
    
    local request='{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"ssh_execute","arguments":{}}}'
    
    local response=$(send_mcp_request "$request")
    
    if echo "$response" | grep -q "Status: Ready"; then
        log_success "ssh_execute 测试模式成功"
        echo "$response" | jq '.result.content[0].text'
        return 0
    else
        log_error "ssh_execute 测试模式失败"
        echo "$response"
        return 1
    fi
}

# 测试 SFTP 上传（测试模式）
test_sftp_upload_test_mode() {
    log_info "测试 sftp_upload 工具（测试模式）..."
    
    local request='{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"sftp_upload","arguments":{}}}'
    
    local response=$(send_mcp_request "$request")
    
    if echo "$response" | grep -q "Status: Ready"; then
        log_success "sftp_upload 测试模式成功"
        echo "$response" | jq '.result.content[0].text'
        return 0
    else
        log_error "sftp_upload 测试模式失败"
        echo "$response"
        return 1
    fi
}

# 测试连接池状态
test_pool_stats() {
    log_info "测试 pool_stats 工具..."
    
    local request='{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"pool_stats","arguments":{}}}'
    
    local response=$(send_mcp_request "$request")
    
    if echo "$response" | grep -q "Connection Pool Statistics"; then
        log_success "pool_stats 成功"
        echo "$response" | jq -r '.result.content[0].text'
        return 0
    else
        log_error "pool_stats 失败"
        echo "$response"
        return 1
    fi
}

# 测试主机管理 - 列出主机
test_host_list() {
    log_info "测试 host_list 工具..."
    
    local request='{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"host_list","arguments":{}}}'
    
    local response=$(send_mcp_request "$request")
    
    if echo "$response" | grep -q "result"; then
        log_success "host_list 成功"
        echo "$response" | jq -r '.result.content[0].text'
        return 0
    else
        log_error "host_list 失败"
        echo "$response"
        return 1
    fi
}

# 测试主机管理 - 添加主机
test_host_add() {
    log_info "测试 host_add 工具..."
    
    local request='{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"host_add","arguments":{"name":"test-host","host":"192.168.1.100","description":"Test host","port":"22","user":"testuser","type":"linux"}}}'
    
    local response=$(send_mcp_request "$request")
    
    if echo "$response" | grep -q "added successfully"; then
        log_success "host_add 成功"
        echo "$response" | jq -r '.result.content[0].text'
        return 0
    else
        log_error "host_add 失败"
        echo "$response"
        return 1
    fi
}

# 测试主机管理 - 删除主机
test_host_remove() {
    log_info "测试 host_remove 工具..."
    
    local request='{"jsonrpc":"2.0","id":8,"method":"tools/call","params":{"name":"host_remove","arguments":{"name":"test-host"}}}'
    
    local response=$(send_mcp_request "$request")
    
    if echo "$response" | grep -q "removed successfully"; then
        log_success "host_remove 成功"
        echo "$response" | jq -r '.result.content[0].text'
        return 0
    else
        log_error "host_remove 失败"
        echo "$response"
        return 1
    fi
}

# 测试错误处理
test_error_handling() {
    log_info "测试错误处理..."
    
    # 测试未知方法
    local request='{"jsonrpc":"2.0","id":9,"method":"unknown_method","params":{}}'
    
    local response=$(send_mcp_request "$request")
    
    if echo "$response" | grep -q "Method not found"; then
        log_success "错误处理正常（未知方法）"
        echo "$response" | jq '.'
        return 0
    else
        log_error "错误处理失败"
        echo "$response"
        return 1
    fi
}

# 测试无效的 JSON
test_invalid_json() {
    log_info "测试无效 JSON 处理..."
    
    local request='{"invalid json'
    
    local response=$(send_mcp_request "$request")
    
    # 检查是否返回了 Parse error，且没有 id 字段（因为请求无法解析）
    if echo "$response" | jq -e '.error.code == -32700 and .error.message == "Parse error"' > /dev/null 2>&1; then
        log_success "无效 JSON 处理正常 - 正确返回解析错误"
        echo "$response" | jq '.'
        return 0
    else
        log_error "无效 JSON 处理失败 - 期望返回 Parse error"
        echo "$response"
        return 1
    fi
}

# 测试缺少必需参数
test_missing_required_params() {
    log_info "测试缺少必需参数..."
    
    # ssh_execute 需要 command 参数
    local request='{"jsonrpc":"2.0","id":10,"method":"tools/call","params":{"name":"ssh_execute","arguments":{"host":"192.168.1.1"}}}'
    
    local response=$(send_mcp_request "$request")
    
    # 应该返回错误
    if echo "$response" | jq -e '.error.code == -32000 and (.error.message | contains("command is required"))' > /dev/null 2>&1; then
        log_success "缺少必需参数处理正常 - 正确返回错误"
        echo "$response" | jq '.'
        return 0
    else
        log_error "缺少必需参数处理失败"
        echo "$response"
        return 1
    fi
}

# 测试未知工具
test_unknown_tool() {
    log_info "测试调用未知工具..."
    
    local request='{"jsonrpc":"2.0","id":11,"method":"tools/call","params":{"name":"unknown_tool","arguments":{}}}'
    
    local response=$(send_mcp_request "$request")
    
    # 应该返回错误
    if echo "$response" | jq -e '.error.code == -32000 and (.error.message | contains("unknown tool"))' > /dev/null 2>&1; then
        log_success "未知工具处理正常 - 正确返回错误"
        echo "$response" | jq '.'
        return 0
    else
        log_error "未知工具处理失败"
        echo "$response"
        return 1
    fi
}

# 测试真实主机连接 - appserver
test_real_host_connection() {
    log_info "测试真实主机连接 (appserver)..."
    
    local request='{"jsonrpc":"2.0","id":12,"method":"tools/call","params":{"name":"host_test","arguments":{"name":"appserver"}}}'
    
    local response=$(send_mcp_request "$request")
    
    # 检查是否连接成功
    if echo "$response" | jq -e '.result.content[0].text | contains("Connection to")' > /dev/null 2>&1; then
        log_success "真实主机连接测试成功"
        echo "$response" | jq -r '.result.content[0].text'
        return 0
    else
        log_error "真实主机连接测试失败"
        echo "$response" | jq '.'
        return 1
    fi
}

# 测试真实 SSH 命令执行
test_real_ssh_execute() {
    log_info "测试真实 SSH 命令执行 (appserver: uptime)..."
    
    local request='{"jsonrpc":"2.0","id":13,"method":"tools/call","params":{"name":"ssh_execute","arguments":{"host":"10.201.15.192","command":"uptime","user":"master","port":"22"}}}'
    
    local response=$(send_mcp_request "$request")
    
    # 检查是否执行成功
    if echo "$response" | jq -e '.result.content[0].text' > /dev/null 2>&1; then
        log_success "SSH 命令执行成功"
        echo "$response" | jq -r '.result.content[0].text'
        return 0
    else
        log_error "SSH 命令执行失败"
        echo "$response" | jq '.'
        return 1
    fi
}

# 测试 SFTP 列出目录
test_real_sftp_list() {
    log_info "测试真实 SFTP 列出目录 (appserver: /tmp)..."
    
    local request='{"jsonrpc":"2.0","id":14,"method":"tools/call","params":{"name":"sftp_list","arguments":{"host":"10.201.15.192","remote_path":"/tmp","user":"master","port":"22"}}}'
    
    local response=$(send_mcp_request "$request")
    
    # 检查是否列出成功
    if echo "$response" | jq -e '.result.content[0].text' > /dev/null 2>&1; then
        log_success "SFTP 列出目录成功"
        echo "$response" | jq -r '.result.content[0].text' | head -20
        return 0
    else
        log_error "SFTP 列出目录失败"
        echo "$response" | jq '.'
        return 1
    fi
}

# 主测试函数
main() {
    echo "=========================================="
    echo "  SSHX MCP 模式测试"
    echo "=========================================="
    echo ""
    
    # 检查 jq 是否安装
    if ! command -v jq &> /dev/null; then
        log_warn "未安装 jq，建议安装以获得更好的输出格式"
        log_warn "macOS: brew install jq"
    fi
    
    # 检查 sshx 可执行文件
    if [ ! -f "./bin/sshx" ]; then
        log_error "找不到 ./bin/sshx 可执行文件"
        log_info "请先运行: make build"
        exit 1
    fi
    
    # 检查是否运行真实主机测试
    local run_real_tests=false
    if [ "$1" = "--real" ] || [ "$1" = "-r" ]; then
        run_real_tests=true
        log_info "启用真实主机测试"
    fi
    
    # 运行测试
    local failed=0
    
    test_initialize || ((failed++))
    echo ""
    
    test_tools_list || ((failed++))
    echo ""
    
    test_ssh_execute_test_mode || ((failed++))
    echo ""
    
    test_sftp_upload_test_mode || ((failed++))
    echo ""
    
    test_pool_stats || ((failed++))
    echo ""
    
    test_host_list || ((failed++))
    echo ""
    
    test_host_add || ((failed++))
    echo ""
    
    test_host_list || ((failed++))
    echo ""
    
    test_host_remove || ((failed++))
    echo ""
    
    test_error_handling || ((failed++))
    echo ""
    
    test_invalid_json || ((failed++))
    echo ""
    
    test_missing_required_params || ((failed++))
    echo ""
    
    test_unknown_tool || ((failed++))
    echo ""
    
    # 真实主机测试
    echo "=========================================="
    echo "  真实主机连接测试 (appserver)"
    echo "=========================================="
    echo ""
    
    test_real_host_connection || ((failed++))
    echo ""
    
    test_real_ssh_execute || ((failed++))
    echo ""
    
    test_real_sftp_list || ((failed++))
    echo ""
    
    # 总结
    echo "=========================================="
    if [ $failed -eq 0 ]; then
        log_success "所有测试通过！"
    else
        log_error "$failed 个测试失败"
        exit 1
    fi
    echo "=========================================="
}

# 运行主函数
main "$@"
