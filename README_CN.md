```
 $$$$$$\   $$$$$$\  $$\   $$\ $$\      $$\  $$$$$$\  $$$$$$$\
$$  __$$\ $$  __$$\ $$ |  $$ |$$$\    $$$ |$$  __$$\ $$  __$$\
$$ /  \__|$$ /  \__|$$ |  $$ |$$$$\  $$$$ |$$ /  \__|$$ |  $$ |
\$$$$$$\  \$$$$$$\  $$$$$$$$ |$$\$$\$$ $$ |$$ |      $$$$$$$  |
 \____$$\  \____$$\ $$  __$$ |$$ \$$$  $$ |$$ |      $$  ____/
$$\   $$ |$$\   $$ |$$ |  $$ |$$ |\$  /$$ |$$ |  $$\ $$ |
\$$$$$$  |\$$$$$$  |$$ |  $$ |$$ | \_/ $$ |\$$$$$$  |$$ |
 \______/  \______/ \__|  \__|\__|     \__| \______/ \__|


支持 MCP 协议的安全 SSH 和 SFTP 客户端
```

<div align="center">

[![Go Version](https://img.shields.io/github/go-mod/go-version/talkincode/sshmcp?style=flat-square&logo=go&logoColor=white)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/talkincode/sshmcp?style=flat-square&logo=github)](https://github.com/talkincode/sshmcp/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square)](https://github.com/talkincode/sshmcp/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/talkincode/sshmcp?style=flat-square)](https://goreportcard.com/report/github.com/talkincode/sshmcp)
[![Coverage](https://img.shields.io/badge/coverage-20.0%25-yellow?style=flat-square&logo=go)](https://github.com/talkincode/sshmcp)

[![GitHub Stars](https://img.shields.io/github/stars/talkincode/sshmcp?style=flat-square&logo=github)](https://github.com/talkincode/sshmcp/stargazers)
[![GitHub Forks](https://img.shields.io/github/forks/talkincode/sshmcp?style=flat-square&logo=github)](https://github.com/talkincode/sshmcp/network/members)
[![GitHub Issues](https://img.shields.io/github/issues/talkincode/sshmcp?style=flat-square&logo=github)](https://github.com/talkincode/sshmcp/issues)
[![GitHub Pull Requests](https://img.shields.io/github/issues-pr/talkincode/sshmcp?style=flat-square&logo=github)](https://github.com/talkincode/sshmcp/pulls)

[![GitHub Downloads](https://img.shields.io/github/downloads/talkincode/sshmcp/total?style=flat-square&logo=github)](https://github.com/talkincode/sshmcp/releases)
[![GitHub Contributors](https://img.shields.io/github/contributors/talkincode/sshmcp?style=flat-square&logo=github)](https://github.com/talkincode/sshmcp/graphs/contributors)
[![Last Commit](https://img.shields.io/github/last-commit/talkincode/sshmcp?style=flat-square&logo=github)](https://github.com/talkincode/sshmcp/commits/main)
[![Repo Size](https://img.shields.io/github/repo-size/talkincode/sshmcp?style=flat-square&logo=github)](https://github.com/talkincode/sshmcp)

[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-blue?style=flat-square&logo=linux&logoColor=white)](https://github.com/talkincode/sshmcp/releases)
[![MCP Protocol](https://img.shields.io/badge/MCP-2024--11--05-orange?style=flat-square)](https://modelcontextprotocol.io)
[![Made with Go](https://img.shields.io/badge/Made%20with-Go-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](https://github.com/talkincode/sshmcp/pulls)

[English](./README.md) | 简体中文

</div>

---

# SSHMCP

`sshx` 提供了一个无障碍的 SSH 命令行客户端，同时实现了 MCP（Model Context Protocol，模型上下文协议）接口，使 AI 助手能够轻松调用远程 SSH/SFTP 功能。

## 项目结构

- `cmd/sshx`: 主二进制入口点，负责命令行参数解析、MCP 模式启动和密码管理功能。
- `internal/sshclient`: 核心 SSH/SFTP/脚本执行逻辑、命令安全验证和连接池封装。
- `internal/mcp`: MCP stdio 服务器实现，向外部工具（如 AI 助手）暴露 SSH/SFTP/脚本工具。

## 核心特性

1. 跨平台 SSH/SFTP 操作（支持 sudo 自动填充）。
2. 密码管理（Keychain / Secret Service / Credential Manager）。
3. MCP stdio 模式用于 AI 助手集成。
4. 连接池、脚本执行和命令安全验证。

## 安装

### 使用 Go 快速安装（推荐 Go 用户）

如果您已安装 Go 1.21+，可以使用 Go 的内置工具：

#### 直接运行无需安装（类似 npx）

```bash
# 运行最新版本
go run github.com/talkincode/sshmcp/cmd/sshx@latest --help

# 运行指定版本
go run github.com/talkincode/sshmcp/cmd/sshx@v0.0.6 -h=192.168.1.100 "uptime"
```

#### 全局安装

```bash
# 安装最新版本到 $GOPATH/bin
go install github.com/talkincode/sshmcp/cmd/sshx@latest

# 然后可以在任何地方使用
sshx --help
sshx -h=192.168.1.100 "uptime"
```

**注意：** 确保 `$GOPATH/bin`（通常是 `~/go/bin`）在您的 PATH 中。

### 一键安装脚本

#### Linux / macOS

```bash
curl -fsSL https://raw.githubusercontent.com/talkincode/sshmcp/main/install.sh | bash
```

或下载后运行：

```bash
wget https://raw.githubusercontent.com/talkincode/sshmcp/main/install.sh
chmod +x install.sh
./install.sh
```

安装特定版本：

```bash
./install.sh v0.0.2
```

#### Windows

以管理员身份打开 PowerShell 并运行：

```powershell
irm https://raw.githubusercontent.com/talkincode/sshmcp/main/install.ps1 | iex
```

或下载后运行：

```powershell
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/talkincode/sshmcp/main/install.ps1" -OutFile "install.ps1"
.\install.ps1
```

安装特定版本：

```powershell
.\install.ps1 -Version v0.0.2
```

### 手动安装

从 [Releases](https://github.com/talkincode/sshmcp/releases) 下载预编译二进制文件：

**Linux / macOS:**

```bash
# 下载并解压（将 <platform>-<arch> 替换为您的系统）
tar -xzf sshx-<platform>-<arch>.tar.gz

# 移动到系统路径
sudo mv sshx /usr/local/bin/

# 添加执行权限
sudo chmod +x /usr/local/bin/sshx

# 验证安装
sshx --help
```

**Windows:**

1. 下载 `sshx-windows-amd64.zip`
2. 解压文件
3. 将 `sshx.exe` 移动到 PATH 中的目录（例如 `C:\Program Files\sshx`）
4. 或将解压目录添加到系统 PATH

### 从源代码构建

```bash
# 克隆仓库
git clone https://github.com/talkincode/sshmcp.git
cd sshmcp

# 构建命令行工具
go build -o bin/sshx ./cmd/sshx

# 安装到系统（可选）
make install
```

## 快速开始

```bash
# 执行远程命令
sshx -h=192.168.1.100 -u=root "uptime"

# 保存密码以便更轻松访问
sshx --set-password host=192.168.1.100 user=root

# 执行命令时无需密码标志（使用已保存的密码）
sshx -h=192.168.1.100 -u=root "df -h"

# 启动 MCP stdio 模式
sshx mcp-stdio
```

## 密码管理

`sshx` 使用操作系统的原生凭据管理器提供安全的密码存储，无需重复输入密码或以明文形式存储密码。

### 支持的平台

- **macOS**: 使用 Keychain Access（钥匙串访问）
- **Linux**: 使用 Secret Service（GNOME Keyring / KDE Wallet）
- **Windows**: 使用 Credential Manager（凭据管理器）

### 密码命令

#### 保存密码

```bash
# 为特定主机保存密码
./bin/sshx --set-password host=192.168.1.100 user=root

# 使用环境变量保存密码（推荐用于脚本）
export SSH_HOST=192.168.1.100
export SSH_USER=root
./bin/sshx --set-password
```

系统会提示您安全地输入密码（输入时隐藏）。

#### 检查已保存的密码

```bash
# 检查主机是否存在密码
./bin/sshx --check-password host=192.168.1.100 user=root

# 输出示例：
# ✓ Password exists for root@192.168.1.100
```

#### 列出所有已保存的密码

```bash
# 列出所有存储的 SSH 凭据
./bin/sshx --list-passwords

# 输出示例：
# Stored SSH passwords:
# - root@192.168.1.100
# - admin@192.168.1.101
# - ubuntu@192.168.1.102
```

#### 删除密码

```bash
# 删除特定主机的密码
./bin/sshx --delete-password host=192.168.1.100 user=root

# 确认消息：
# ✓ Password deleted for root@192.168.1.100
```

### 使用已存储的密码

保存密码后，您可以在不使用 `-p` 标志的情况下连接：

```bash
# 不使用密码管理（每次都需要 -p 标志）
./bin/sshx -h=192.168.1.100 -u=root -p=yourpassword "uptime"

# 使用密码管理（不需要 -p 标志）
./bin/sshx --set-password host=192.168.1.100 user=root  # 保存一次
./bin/sshx -h=192.168.1.100 -u=root "uptime"            # 永久使用
```

### 密码优先级

执行 SSH 命令时，`sshx` 遵循以下优先级顺序：

1. **命令行密码**（`-p` 标志）- 最高优先级
2. **环境变量**（`SSH_PASSWORD`）
3. **已存储的密码**（来自系统凭据管理器）
4. **交互式提示** - 如果以上都不可用

### 安全说明

- ✅ 密码使用操作系统原生加密存储
- ✅ 密码永远不会以明文形式存储
- ✅ 每个主机+用户组合都有单独的密码条目
- ✅ 输入时密码被隐藏
- ⚠️ 需要操作系统凭据管理器可用
- ⚠️ 在 Linux 上，需要 Secret Service 守护进程运行（桌面环境通常自动运行）

### 环境变量

您可以使用环境变量来避免重复输入凭据：

```bash
# 在 .env 文件中设置或在 shell 中导出
export SSH_HOST=192.168.1.100
export SSH_USER=root
export SSH_PORT=22
export SUDO_PASSWORD=your_sudo_password

# 然后运行命令时无需标志
./bin/sshx "uptime"
```

### 示例工作流

```bash
# 1. 保存密码一次
./bin/sshx --set-password host=192.168.1.100 user=root
# Enter password: ******

# 2. 验证已保存
./bin/sshx --check-password host=192.168.1.100 user=root
# ✓ Password exists for root@192.168.1.100

# 3. 用于 SSH 命令（不需要密码）
./bin/sshx -h=192.168.1.100 -u=root "ls -la /var/log"
./bin/sshx -h=192.168.1.100 -u=root "df -h"

# 4. 用于 SFTP 操作（不需要密码）
./bin/sshx -h=192.168.1.100 -u=root --sftp-get /etc/hosts ./hosts.txt
./bin/sshx -h=192.168.1.100 -u=root --sftp-put ./local.txt /tmp/remote.txt

# 5. 完成后，可选择删除密码
./bin/sshx --delete-password host=192.168.1.100 user=root
# ✓ Password deleted for root@192.168.1.100
```

## 故障排除

### "sshx: command not found"（命令未找到）

**解决方案：**

- 确保 `/usr/local/bin`（或您的安装目录）在您的 PATH 中
- 安装后重启终端
- 或使用完整路径运行：`/usr/local/bin/sshx`

### macOS 安全警告

macOS 可能在首次运行时阻止二进制文件：

```bash
sudo xattr -rd com.apple.quarantine /usr/local/bin/sshx
```

或前往系统偏好设置 → 安全性与隐私 → 点击"仍要打开"

### Windows SmartScreen 警告

如果 Windows Defender SmartScreen 显示警告，请点击"更多信息"，然后点击"仍要运行"。

### 权限被拒绝

```bash
# 确保二进制文件具有执行权限
sudo chmod +x /usr/local/bin/sshx
```

## 开发

```bash
# 运行测试
go test ./...

# 格式化代码
gofmt -w .

# 为所有平台构建
make build-all

# 运行代码检查
make lint
```

> lint 目标需要 `golangci-lint` v2.6.1 或更高版本。使用 `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.6.1` 安装。

## 许可证

本项目采用 MIT 许可证 - 有关详细信息，请参阅 [LICENSE](LICENSE) 文件。

---

<div align="center">

**[文档](https://github.com/talkincode/sshmcp/wiki)** •
**[问题](https://github.com/talkincode/sshmcp/issues)** •
**[讨论](https://github.com/talkincode/sshmcp/discussions)** •
**[发布版本](https://github.com/talkincode/sshmcp/releases)**

用 ❤️ 制作，作者 [talkincode](https://github.com/talkincode)

</div>
