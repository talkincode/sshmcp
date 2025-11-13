```
 $$$$$$\   $$$$$$\  $$\   $$\ $$\      $$\  $$$$$$\  $$$$$$$\
$$  __$$\ $$  __$$\ $$ |  $$ |$$$\    $$$ |$$  __$$\ $$  __$$\
$$ /  \__|$$ /  \__|$$ |  $$ |$$$$\  $$$$ |$$ /  \__|$$ |  $$ |
\$$$$$$\  \$$$$$$\  $$$$$$$$ |$$\$$\$$ $$ |$$ |      $$$$$$$  |
 \____$$\  \____$$\ $$  __$$ |$$ \$$$  $$ |$$ |      $$  ____/
$$\   $$ |$$\   $$ |$$ |  $$ |$$ |\$  /$$ |$$ |  $$\ $$ |
\$$$$$$  |\$$$$$$  |$$ |  $$ |$$ | \_/ $$ |\$$$$$$  |$$ |
 \______/  \______/ \__|  \__|\__|     \__| \______/ \__|


Secure SSH & SFTP Client with MCP Protocol Support
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

English | [简体中文](./README_CN.md)

</div>

---

# SSHMCP

`sshx` provides a barrier-free SSH command-line client while implementing the MCP (Model Context Protocol) interface, enabling AI assistants to easily invoke remote SSH/SFTP functionality.

## Why You Need It?

Managing multiple servers means juggling different passwords, repeatedly entering sudo passwords, and manually executing SSH commands in AI assistants. `sshx` securely stores passwords in your system keyring, auto-fills sudo passwords, and enables AI assistants to directly operate remote servers through MCP protocol. One command, multiple servers, zero password hassle.

**New!** Host Configuration Management - Store your frequently used host configurations in `~/.sshmcp/settings.json` and connect with just a name instead of typing full connection details every time. Import from your existing `~/.ssh/config` or add hosts interactively!

## Project Structure

- `cmd/sshx`: Main binary entry point, responsible for command-line argument parsing, MCP mode startup, and password management features.
- `internal/sshclient`: Core SSH/SFTP/script execution logic, command security validation, and connection pool wrapper.
- `internal/mcp`: MCP stdio server implementation, exposing SSH/SFTP/script tools to external tools (e.g., AI assistants).

## Key Features

1. Cross-platform SSH/SFTP operations (supports sudo auto-fill).
2. Password management (Keychain / Secret Service / Credential Manager).
3. MCP stdio mode for AI assistant integration.
4. Connection pooling, script execution, and command security validation.

## Installation

### Quick Install with Go (Recommended for Go Users)

If you have Go 1.21+ installed, you can use Go's built-in tools:

#### Run directly without installation (like npx)

```bash
# Run the latest version
go run github.com/talkincode/sshmcp/cmd/sshx@latest --help

# Run specific version
go run github.com/talkincode/sshmcp/cmd/sshx@v0.0.6 -h=192.168.1.100 "uptime"
```

#### Install globally

```bash
# Install latest version to $GOPATH/bin
go install github.com/talkincode/sshmcp/cmd/sshx@latest

# Then use it anywhere
sshx --help
sshx -h=192.168.1.100 "uptime"
```

**Note:** Make sure `$GOPATH/bin` (typically `~/go/bin`) is in your PATH.

### One-Line Installation Script

#### Linux / macOS

```bash
curl -fsSL https://raw.githubusercontent.com/talkincode/sshmcp/main/install.sh | bash
```

Or download and run:

```bash
wget https://raw.githubusercontent.com/talkincode/sshmcp/main/install.sh
chmod +x install.sh
./install.sh
```

Install specific version:

```bash
./install.sh v0.0.2
```

#### Windows

Open PowerShell as Administrator and run:

```powershell
irm https://raw.githubusercontent.com/talkincode/sshmcp/main/install.ps1 | iex
```

Or download and run:

```powershell
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/talkincode/sshmcp/main/install.ps1" -OutFile "install.ps1"
.\install.ps1
```

Install specific version:

```powershell
.\install.ps1 -Version v0.0.2
```

### Manual Installation

Download pre-built binaries from [Releases](https://github.com/talkincode/sshmcp/releases):

**Linux / macOS:**

```bash
# Download and extract (replace <platform>-<arch> with your system)
tar -xzf sshx-<platform>-<arch>.tar.gz

# Move to system path
sudo mv sshx /usr/local/bin/

# Make executable
sudo chmod +x /usr/local/bin/sshx

# Verify installation
sshx --help
```

**Windows:**

1. Download `sshx-windows-amd64.zip`
2. Extract the archive
3. Move `sshx.exe` to a directory in your PATH (e.g., `C:\Program Files\sshx`)
4. Or add the extracted directory to your system PATH

### Build from Source

```bash
# Clone repository
git clone https://github.com/talkincode/sshmcp.git
cd sshmcp

# Build command-line tool
go build -o bin/sshx ./cmd/sshx

# Install to system (optional)
make install
```

## Quick Start

```bash
# Execute remote command
sshx -h=192.168.1.100 -u=root "uptime"

# Save password for easier access (interactive input)
sshx --password-set=root

# Or set password for specific host
sshx --password-set=192.168.1.100-root

# Execute command without password flag (uses saved password)
sshx -h=192.168.1.100 -u=root "df -h"

# Start MCP stdio mode
sshx mcp-stdio
```

## Host Configuration Management

**NEW!** Manage your frequently used hosts in `~/.sshmcp/settings.json` for quick access.

### Quick Setup

```bash
# Import hosts from your existing ~/.ssh/config
sshx --host-import

# Or add hosts interactively
sshx --host-add

# Add host with command line options
sshx --host-add --host-name=prod-web -h=192.168.1.100 -u=root --host-desc="Production Web Server"

# List all configured hosts
sshx --host-list

# Test connection to a configured host
sshx --host-test=prod-web

# Use configured host (auto-resolves from settings)
sshx -h=prod-web "systemctl status nginx"
```

### Configuration File Format

Location: `~/.sshmcp/settings.json`

```json
{
  "key": "/Users/username/.ssh/id_rsa",
  "hosts": [
    {
      "name": "prod-web",
      "description": "Production Web Server",
      "host": "192.168.1.100",
      "port": "22",
      "user": "root",
      "password_key": "prod-web-password",
      "type": "linux"
    }
  ]
}
```

### Host Management Commands

- `--host-add` - Add new host (interactive or with options)
- `--host-import` - Import hosts from `~/.ssh/config`
- `--host-list` - List all configured hosts
- `--host-test=<name>` - Test connection to a host
- `--host-remove=<name>` - Remove a host from configuration

**Benefits:**

- 📝 Store connection details once, use everywhere
- 🚀 Connect with just a name: `sshx -h=prod-web "command"`
- 🔄 Import from existing `~/.ssh/config`
- 🔐 Integrate with password manager for each host
- ✅ Test connections before use

## Password Management

`sshx` provides secure password storage using the operating system's native credential manager, eliminating the need to enter passwords repeatedly or store them in plaintext.

### Supported Platforms

- **macOS**: Uses Keychain Access
- **Linux**: Uses Secret Service (GNOME Keyring / KDE Wallet)
- **Windows**: Uses Credential Manager

### Password Commands

#### Save Password

```bash
# Save default sudo password (interactive input, recommended)
sshx --password-set=master

# Save password for specific user
sshx --password-set=root

# Save password for specific host+user combination
sshx --password-set=192.168.1.100-root

# Set password inline (not recommended, insecure)
sshx --password-set=master:yourpassword
```

You will be prompted to enter the password securely (input is hidden).

#### Check Saved Password

```bash
# Check if password exists
sshx --password-check=master
sshx --password-check=root

# Output example:
# ✓ Password exists for key: master
```

#### List Saved Passwords

```bash
# List common password keys
sshx --password-list

# Output example:
# Checking password keys in system keyring...
# Service: sshx
#
# Common keys:
#   ✓ master (exists)
#   ✓ root (exists)
#     sudo (not set)
```

#### Get Password

```bash
# Get stored password (for debugging)
sshx --password-get=master

# Output example:
# ✓ Password retrieved from system keyring
#   Service: sshx
#   Key: master
#
# Password: yourpassword
```

#### Delete Password

```bash
# Delete password
sshx --password-delete=master
sshx --password-delete=root

# Confirmation message:
# ✓ Password deleted from system keyring
#   Service: sshx
#   Key: master
```

### Using Stored Passwords

Once a password is saved, sudo commands will automatically retrieve the password from system keyring:

```bash
# 1. First save sudo password
sshx --password-set=master

# 2. Execute sudo commands (automatically uses stored password)
sshx -h=192.168.1.100 -u=root "sudo systemctl status nginx"
sshx -h=192.168.1.100 -u=root "sudo reboot"

# 3. Multi-server scenario: save different passwords for different servers
sshx --password-set=server-A
sshx --password-set=server-B
sshx --password-set=server-C

# 4. Use -pk parameter to specify sudo password key temporarily
sshx -h=192.168.1.100 -pk=server-A "sudo systemctl restart nginx"
sshx -h=192.168.1.101 -pk=server-B "sudo systemctl restart nginx"
sshx -h=192.168.1.102 -pk=server-C "sudo systemctl restart nginx"
```

### Password Key Names

- **master**: Default sudo password key name, used for sudo commands
- **root**: Password for root user
- **Custom keys**: You can use any key name, e.g., `server-A`, `server-B`, `prod-db`, etc.

### Best Practices for Multi-Server Password Management

If you manage multiple servers with the same username but different passwords, use this strategy:

```bash
# Scenario: Manage 3 servers, all with root user but different passwords

# 1. Save password for each server (use meaningful key names)
sshx --password-set=prod-web      # Production web server
sshx --password-set=prod-db       # Production database server
sshx --password-set=dev-server    # Development server

# 2. Execute commands using -pk parameter to specify password key
sshx -h=192.168.1.10 -u=root -pk=prod-web "sudo systemctl status nginx"
sshx -h=192.168.1.20 -u=root -pk=prod-db "sudo systemctl status mysql"
sshx -h=192.168.1.30 -u=root -pk=dev-server "sudo docker ps"

# 3. You can also use aliases to simplify commands (add to ~/.zshrc or ~/.bashrc)
alias ssh-prod-web='sshx -h=192.168.1.10 -u=root -pk=prod-web'
alias ssh-prod-db='sshx -h=192.168.1.20 -u=root -pk=prod-db'
alias ssh-dev='sshx -h=192.168.1.30 -u=root -pk=dev-server'

# Then use simply:
ssh-prod-web "sudo systemctl restart nginx"
ssh-prod-db "sudo systemctl restart mysql"
ssh-dev "sudo docker-compose up -d"
```

### Environment Variables

You can customize the sudo password key name via environment variable (but using `-pk` parameter is more flexible):

```bash
# Use environment variable (can only specify one at a time, needs constant modification)
export SSH_SUDO_KEY=my-sudo-password
sshx --password-set=my-sudo-password
sshx -h=192.168.1.100 "sudo ls -la /root"

# Recommended: Use -pk parameter, more flexible, no need to modify environment variables
sshx -h=192.168.1.100 -pk=server-A "sudo ls -la /root"
sshx -h=192.168.1.101 -pk=server-B "sudo ls -la /root"
```

### Security Notes

- ✅ Passwords are stored using OS-native encryption
- ✅ Passwords are never stored in plaintext
- ✅ Each host+user combination has a separate password entry
- ✅ Password input is hidden during entry
- ⚠️ Requires OS credential manager to be available
- ⚠️ On Linux, requires Secret Service daemon running (usually automatic with desktop environments)

### Environment Variables

You can use environment variables to avoid typing credentials repeatedly:

```bash
# Set in .env file or export in shell
export SSH_HOST=192.168.1.100
export SSH_USER=root
export SSH_PORT=22
export SUDO_PASSWORD=your_sudo_password

# Then run commands without flags
./bin/sshx "uptime"
```

### Example Workflow

```bash
# 1. Save sudo password (interactive input)
sshx --password-set=master
# Enter password for key 'master': ******

# 2. Verify it's saved
sshx --password-check=master
# ✓ Password exists for key: master

# 3. Use for SSH commands (sudo automatically uses stored password)
sshx -h=192.168.1.100 -u=root "sudo systemctl status docker"
sshx -h=192.168.1.100 -u=root "sudo df -h"

# 4. Use for SFTP operations
sshx -h=192.168.1.100 -u=root --upload=local.txt --to=/tmp/remote.txt
sshx -h=192.168.1.100 -u=root --download=/etc/hosts --to=./hosts.txt

# 5. List all saved password keys
sshx --password-list
# Common keys:
#   ✓ master (exists)
#     root (not set)

# 6. When done, optionally delete the password
sshx --password-delete=master
# ✓ Password deleted from system keyring
```

## Troubleshooting

### "sshx: command not found"

**Solution:**

- Ensure `/usr/local/bin` (or your installation directory) is in your PATH
- Restart your terminal after installation
- Or run with full path: `/usr/local/bin/sshx`

### macOS Security Warning

macOS may block the binary on first run:

```bash
sudo xattr -rd com.apple.quarantine /usr/local/bin/sshx
```

Or go to System Preferences → Security & Privacy → Click "Allow Anyway"

### Windows SmartScreen Warning

Click "More info" and then "Run anyway" if Windows Defender SmartScreen shows a warning.

### Permission Denied

```bash
# Make sure the binary has execute permissions
sudo chmod +x /usr/local/bin/sshx
```

## Development

```bash
# Run tests
go test ./...

# Format code
gofmt -w .

# Build for all platforms
make build-all

# Run linter
make lint
```

> The lint target requires `golangci-lint` v2.6.1 or newer. Install it with `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.6.1`.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<div align="center">

**[Documentation](https://github.com/talkincode/sshmcp/wiki)** •
**[Issues](https://github.com/talkincode/sshmcp/issues)** •
**[Discussions](https://github.com/talkincode/sshmcp/discussions)** •
**[Releases](https://github.com/talkincode/sshmcp/releases)**

Made with ❤️ by [talkincode](https://github.com/talkincode)

</div>
