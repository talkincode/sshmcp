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

</div>

---

# SSHMCP

`sshx` provides a barrier-free SSH command-line client while implementing the MCP (Model Context Protocol) interface, enabling AI assistants to easily invoke remote SSH/SFTP functionality.

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

### One-Line Installation (Recommended)

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

# Save password for easier access
sshx --set-password host=192.168.1.100 user=root

# Execute command without password flag (uses saved password)
sshx -h=192.168.1.100 -u=root "df -h"

# Start MCP stdio mode
sshx mcp-stdio
```

## Password Management

`sshx` provides secure password storage using the operating system's native credential manager, eliminating the need to enter passwords repeatedly or store them in plaintext.

### Supported Platforms

- **macOS**: Uses Keychain Access
- **Linux**: Uses Secret Service (GNOME Keyring / KDE Wallet)
- **Windows**: Uses Credential Manager

### Password Commands

#### Save Password

```bash
# Save password for a specific host
./bin/sshx --set-password host=192.168.1.100 user=root

# Save password with environment variables (recommended for scripts)
export SSH_HOST=192.168.1.100
export SSH_USER=root
./bin/sshx --set-password
```

You will be prompted to enter the password securely (input is hidden).

#### Check Saved Password

```bash
# Check if password exists for a host
./bin/sshx --check-password host=192.168.1.100 user=root

# Output example:
# ✓ Password exists for root@192.168.1.100
```

#### List All Saved Passwords

```bash
# List all stored SSH credentials
./bin/sshx --list-passwords

# Output example:
# Stored SSH passwords:
# - root@192.168.1.100
# - admin@192.168.1.101
# - ubuntu@192.168.1.102
```

#### Delete Password

```bash
# Delete password for a specific host
./bin/sshx --delete-password host=192.168.1.100 user=root

# Confirmation message:
# ✓ Password deleted for root@192.168.1.100
```

### Using Stored Passwords

Once a password is saved, you can connect without the `-p` flag:

```bash
# Without password management (requires -p flag every time)
./bin/sshx -h=192.168.1.100 -u=root -p=yourpassword "uptime"

# With password management (no -p flag needed)
./bin/sshx --set-password host=192.168.1.100 user=root  # Save once
./bin/sshx -h=192.168.1.100 -u=root "uptime"            # Use forever
```

### Password Priority

When executing SSH commands, `sshx` follows this priority order:

1. **Command-line password** (`-p` flag) - highest priority
2. **Environment variable** (`SSH_PASSWORD`)
3. **Stored password** (from system credential manager)
4. **Interactive prompt** - if none of the above are available

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
# 1. Save password once
./bin/sshx --set-password host=192.168.1.100 user=root
# Enter password: ******

# 2. Verify it's saved
./bin/sshx --check-password host=192.168.1.100 user=root
# ✓ Password exists for root@192.168.1.100

# 3. Use it for SSH commands (no password needed)
./bin/sshx -h=192.168.1.100 -u=root "ls -la /var/log"
./bin/sshx -h=192.168.1.100 -u=root "df -h"

# 4. Use it for SFTP operations (no password needed)
./bin/sshx -h=192.168.1.100 -u=root --sftp-get /etc/hosts ./hosts.txt
./bin/sshx -h=192.168.1.100 -u=root --sftp-put ./local.txt /tmp/remote.txt

# 5. When done, optionally delete the password
./bin/sshx --delete-password host=192.168.1.100 user=root
# ✓ Password deleted for root@192.168.1.100
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
