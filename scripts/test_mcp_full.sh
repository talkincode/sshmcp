#!/bin/bash
# SSHX MCP 完整功能测试
# 测试所有 MCP 工具的实际使用场景

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

log_error() {
    echo -e "${RED}[✗]${NC} $1"
}

log_test() {
    echo -e "${CYAN}[TEST]${NC} $1"
}

# 发送 MCP 请求并获取响应
send_mcp() {
    local request="$1"
    echo "$request" | ./bin/sshx mcp-stdio 2>/dev/null
}

# 提取结果文本
extract_result() {
    echo "$1" | jq -r '.result.content[0].text' 2>/dev/null || echo "$1"
}

# 检查是否成功
check_success() {
    local response="$1"
    if echo "$response" | jq -e '.result' > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

echo "=========================================="
echo "  SSHX MCP 完整功能测试"
echo "=========================================="
echo ""

# 检查 jq
if ! command -v jq &> /dev/null; then
    log_error "需要安装 jq: brew install jq"
    exit 1
fi

# 检查可执行文件
if [ ! -f "./bin/sshx" ]; then
    log_error "找不到 ./bin/sshx，请先运行: make build"
    exit 1
fi

failed=0

# ==================== 基础测试 ====================
echo ""
echo "=========================================="
echo "  1. 基础协议测试"
echo "=========================================="

log_test "初始化 MCP 服务器"
resp=$(send_mcp '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}')
if check_success "$resp"; then
    log_success "MCP 服务器初始化成功"
else
    log_error "初始化失败"
    ((failed++))
fi

log_test "获取工具列表"
resp=$(send_mcp '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}')
tool_count=$(echo "$resp" | jq '.result.tools | length' 2>/dev/null || echo 0)
if [ "$tool_count" -gt 0 ]; then
    log_success "获取到 $tool_count 个工具"
else
    log_error "未获取到工具列表"
    ((failed++))
fi

# ==================== SSH 命令执行测试 ====================
echo ""
echo "=========================================="
echo "  2. SSH 命令执行测试"
echo "=========================================="

log_test "执行基本命令: whoami"
resp=$(send_mcp '{"jsonrpc":"2.0","id":10,"method":"tools/call","params":{"name":"ssh_execute","arguments":{"host":"10.201.15.192","command":"whoami","user":"master","port":"22"}}}')
if check_success "$resp"; then
    result=$(extract_result "$resp")
    log_success "命令执行成功: $(echo $result | tr -d '\r\n')"
else
    log_error "命令执行失败"
    echo "$resp" | jq .
    ((failed++))
fi

log_test "执行系统信息命令: uname -a"
resp=$(send_mcp '{"jsonrpc":"2.0","id":11,"method":"tools/call","params":{"name":"ssh_execute","arguments":{"host":"10.201.15.192","command":"uname -a","user":"master","port":"22"}}}')
if check_success "$resp"; then
    result=$(extract_result "$resp")
    log_success "系统信息: $(echo $result | head -c 60)..."
else
    log_error "命令执行失败"
    ((failed++))
fi

log_test "执行复杂命令: df -h | head -3"
resp=$(send_mcp '{"jsonrpc":"2.0","id":12,"method":"tools/call","params":{"name":"ssh_execute","arguments":{"host":"10.201.15.192","command":"df -h | head -3","user":"master","port":"22"}}}')
if check_success "$resp"; then
    log_success "管道命令执行成功"
else
    log_error "管道命令执行失败"
    ((failed++))
fi

# ==================== SFTP 测试 ====================
echo ""
echo "=========================================="
echo "  3. SFTP 文件操作测试"
echo "=========================================="

log_test "列出远程目录: /tmp"
resp=$(send_mcp '{"jsonrpc":"2.0","id":20,"method":"tools/call","params":{"name":"sftp_list","arguments":{"host":"10.201.15.192","remote_path":"/tmp","user":"master","port":"22"}}}')
if check_success "$resp"; then
    item_count=$(extract_result "$resp" | grep -c "^[drwx-]" || echo 0)
    log_success "列出 /tmp 目录，包含 $item_count 个项目"
else
    log_error "列出目录失败"
    ((failed++))
fi

log_test "创建测试目录: /tmp/sshx-mcp-test"
resp=$(send_mcp '{"jsonrpc":"2.0","id":21,"method":"tools/call","params":{"name":"sftp_mkdir","arguments":{"host":"10.201.15.192","remote_path":"/tmp/sshx-mcp-test","user":"master","port":"22"}}}')
if check_success "$resp"; then
    log_success "目录创建成功"
else
    log_error "目录创建失败"
    ((failed++))
fi

log_test "创建本地测试文件"
echo "Hello from SSHX MCP $(date)" > /tmp/test-upload.txt
if [ -f /tmp/test-upload.txt ]; then
    log_success "本地测试文件创建成功"
else
    log_error "本地测试文件创建失败"
    ((failed++))
fi

log_test "上传文件到远程服务器"
resp=$(send_mcp '{"jsonrpc":"2.0","id":22,"method":"tools/call","params":{"name":"sftp_upload","arguments":{"host":"10.201.15.192","local_path":"/tmp/test-upload.txt","remote_path":"/tmp/sshx-mcp-test/test-file.txt","user":"master","port":"22"}}}')
if check_success "$resp"; then
    log_success "文件上传成功"
else
    log_error "文件上传失败"
    echo "$resp" | jq .
    ((failed++))
fi

log_test "下载文件从远程服务器"
resp=$(send_mcp '{"jsonrpc":"2.0","id":23,"method":"tools/call","params":{"name":"sftp_download","arguments":{"host":"10.201.15.192","remote_path":"/tmp/sshx-mcp-test/test-file.txt","local_path":"/tmp/test-download.txt","user":"master","port":"22"}}}')
if check_success "$resp"; then
    log_success "文件下载成功"
    if diff /tmp/test-upload.txt /tmp/test-download.txt > /dev/null 2>&1; then
        log_success "文件内容验证通过"
    else
        log_error "文件内容不匹配"
        ((failed++))
    fi
else
    log_error "文件下载失败"
    ((failed++))
fi

log_test "删除远程测试文件"
resp=$(send_mcp '{"jsonrpc":"2.0","id":24,"method":"tools/call","params":{"name":"sftp_remove","arguments":{"host":"10.201.15.192","remote_path":"/tmp/sshx-mcp-test/test-file.txt","user":"master","port":"22"}}}')
if check_success "$resp"; then
    log_success "远程文件删除成功"
else
    log_error "远程文件删除失败"
    ((failed++))
fi

log_test "删除远程测试目录"
resp=$(send_mcp '{"jsonrpc":"2.0","id":25,"method":"tools/call","params":{"name":"sftp_remove","arguments":{"host":"10.201.15.192","remote_path":"/tmp/sshx-mcp-test","user":"master","port":"22"}}}')
if check_success "$resp"; then
    log_success "远程目录删除成功"
else
    log_error "远程目录删除失败"
    ((failed++))
fi

# 清理本地文件
rm -f /tmp/test-upload.txt /tmp/test-download.txt

# ==================== 主机管理测试 ====================
echo ""
echo "=========================================="
echo "  4. 主机管理测试"
echo "=========================================="

log_test "列出配置的主机"
resp=$(send_mcp '{"jsonrpc":"2.0","id":30,"method":"tools/call","params":{"name":"host_list","arguments":{}}}')
if check_success "$resp"; then
    result=$(extract_result "$resp")
    host_count=$(echo "$result" | grep -c "^\[" || echo 0)
    log_success "当前配置了 $host_count 个主机"
else
    log_error "列出主机失败"
    ((failed++))
fi

log_test "添加测试主机"
resp=$(send_mcp '{"jsonrpc":"2.0","id":31,"method":"tools/call","params":{"name":"host_add","arguments":{"name":"mcp-test-host","host":"192.168.1.200","description":"MCP测试主机","port":"22","user":"testuser","type":"linux"}}}')
if check_success "$resp"; then
    log_success "主机添加成功"
else
    log_error "主机添加失败"
    ((failed++))
fi

log_test "测试真实主机连接: appserver"
resp=$(send_mcp '{"jsonrpc":"2.0","id":32,"method":"tools/call","params":{"name":"host_test","arguments":{"name":"appserver"}}}')
if check_success "$resp"; then
    log_success "主机连接测试成功"
else
    log_error "主机连接测试失败"
    echo "$resp" | jq .
    ((failed++))
fi

log_test "删除测试主机"
resp=$(send_mcp '{"jsonrpc":"2.0","id":33,"method":"tools/call","params":{"name":"host_remove","arguments":{"name":"mcp-test-host"}}}')
if check_success "$resp"; then
    log_success "主机删除成功"
else
    log_error "主机删除失败"
    ((failed++))
fi

# ==================== 连接池测试 ====================
echo ""
echo "=========================================="
echo "  5. 连接池状态测试"
echo "=========================================="

log_test "获取连接池统计信息"
resp=$(send_mcp '{"jsonrpc":"2.0","id":40,"method":"tools/call","params":{"name":"pool_stats","arguments":{}}}')
if check_success "$resp"; then
    result=$(extract_result "$resp")
    log_success "连接池状态:"
    echo "$result" | grep -E "(Total|Active|Idle|Health)" | sed 's/^/    /'
else
    log_error "获取连接池统计失败"
    ((failed++))
fi

# ==================== 脚本执行测试 ====================
echo ""
echo "=========================================="
echo "  6. 脚本执行测试"
echo "=========================================="

log_test "创建测试脚本"
cat > /tmp/test-script.sh << 'EOF'
#!/bin/bash
echo "=== Test Script Output ==="
echo "Hostname: $(hostname)"
echo "Date: $(date)"
echo "User: $(whoami)"
echo "Current Directory: $(pwd)"
echo "=== Script Completed ==="
EOF
chmod +x /tmp/test-script.sh

if [ -f /tmp/test-script.sh ]; then
    log_success "测试脚本创建成功"
else
    log_error "测试脚本创建失败"
    ((failed++))
fi

log_test "执行远程脚本"
resp=$(send_mcp '{"jsonrpc":"2.0","id":50,"method":"tools/call","params":{"name":"script_execute","arguments":{"host":"10.201.15.192","script_path":"/tmp/test-script.sh","user":"master","port":"22"}}}')
if check_success "$resp"; then
    result=$(extract_result "$resp")
    if echo "$result" | grep -q "Script Completed"; then
        log_success "脚本执行成功"
        echo "$result" | grep -E "(Hostname|Date|User)" | sed 's/^/    /'
    else
        log_error "脚本执行结果异常"
        ((failed++))
    fi
else
    log_error "脚本执行失败"
    echo "$resp" | jq .
    ((failed++))
fi

# 清理测试脚本
rm -f /tmp/test-script.sh

# ==================== 错误处理测试 ====================
echo ""
echo "=========================================="
echo "  7. 错误处理测试"
echo "=========================================="

log_test "测试无效的主机"
resp=$(send_mcp '{"jsonrpc":"2.0","id":60,"method":"tools/call","params":{"name":"ssh_execute","arguments":{"host":"192.168.255.255","command":"whoami","user":"master","port":"22"}}}')
if echo "$resp" | jq -e '.error' > /dev/null 2>&1; then
    log_success "正确返回错误（无效主机）"
else
    log_error "未正确处理无效主机错误"
    ((failed++))
fi

log_test "测试缺少必需参数"
resp=$(send_mcp '{"jsonrpc":"2.0","id":61,"method":"tools/call","params":{"name":"ssh_execute","arguments":{"host":"10.201.15.192"}}}')
if echo "$resp" | jq -e '.error' > /dev/null 2>&1; then
    error_msg=$(echo "$resp" | jq -r '.error.message')
    if echo "$error_msg" | grep -q "command is required"; then
        log_success "正确返回错误（缺少参数）"
    else
        log_error "错误消息不正确"
        ((failed++))
    fi
else
    log_error "未正确处理缺少参数错误"
    ((failed++))
fi

log_test "测试未知工具"
resp=$(send_mcp '{"jsonrpc":"2.0","id":62,"method":"tools/call","params":{"name":"unknown_tool","arguments":{}}}')
if echo "$resp" | jq -e '.error' > /dev/null 2>&1; then
    log_success "正确返回错误（未知工具）"
else
    log_error "未正确处理未知工具错误"
    ((failed++))
fi

# ==================== 总结 ====================
echo ""
echo "=========================================="
echo "  测试总结"
echo "=========================================="

if [ $failed -eq 0 ]; then
    echo -e "${GREEN}"
    echo "  ██████╗  █████╗ ███████╗███████╗"
    echo "  ██╔══██╗██╔══██╗██╔════╝██╔════╝"
    echo "  ██████╔╝███████║███████╗███████╗"
    echo "  ██╔═══╝ ██╔══██║╚════██║╚════██║"
    echo "  ██║     ██║  ██║███████║███████║"
    echo "  ╚═╝     ╚═╝  ╚═╝╚══════╝╚══════╝"
    echo -e "${NC}"
    log_success "所有测试通过！SSHX MCP 服务器运行正常。"
    exit 0
else
    echo -e "${RED}"
    echo "  ███████╗ █████╗ ██╗██╗     "
    echo "  ██╔════╝██╔══██╗██║██║     "
    echo "  █████╗  ███████║██║██║     "
    echo "  ██╔══╝  ██╔══██║██║██║     "
    echo "  ██║     ██║  ██║██║███████╗"
    echo "  ╚═╝     ╚═╝  ╚═╝╚═╝╚══════╝"
    echo -e "${NC}"
    log_error "$failed 个测试失败"
    exit 1
fi
