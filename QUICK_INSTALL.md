# Quick Install

## One-Line Installation

### Linux / macOS

```bash
curl -fsSL https://raw.githubusercontent.com/talkincode/sshmcp/main/install.sh | bash
```

### Windows (PowerShell as Administrator)

```powershell
irm https://raw.githubusercontent.com/talkincode/sshmcp/main/install.ps1 | iex
```

## Quick Start

```bash
# Execute remote command
sshx -h=192.168.1.100 -u=root "uptime"

# Save password for easier access
sshx --set-password host=192.168.1.100 user=root

# Use saved password
sshx -h=192.168.1.100 -u=root "df -h"
```

## More Information

- [Full Installation Guide](INSTALL.md)
- [README](README.md)
- [Releases](https://github.com/talkincode/sshmcp/releases)
