# Installation Guide

Complete installation guide for sshx - SSH & SFTP tool with MCP support.

## Table of Contents

- [Automatic Installation](#automatic-installation)
  - [Linux / macOS](#linux--macos)
  - [Windows](#windows)
- [Manual Installation](#manual-installation)
- [Build from Source](#build-from-source)
- [Verification](#verification)
- [Uninstallation](#uninstallation)
- [Troubleshooting](#troubleshooting)

## Automatic Installation

### Linux / macOS

**One-line installation (recommended):**

```bash
curl -fsSL https://raw.githubusercontent.com/talkincode/sshmcp/main/install.sh | bash
```

**Or download and run the script:**

```bash
# Download the installation script
wget https://raw.githubusercontent.com/talkincode/sshmcp/main/install.sh

# Make it executable
chmod +x install.sh

# Run the installer (installs latest version)
./install.sh

# Or install a specific version
./install.sh v0.0.2
```

**What the script does:**

1. Detects your OS (Linux/macOS) and architecture (amd64/arm64)
2. Downloads the appropriate binary from GitHub Releases
3. Extracts and installs to `/usr/local/bin/sshx`
4. Sets executable permissions
5. Verifies the installation

**Supported platforms:**

- Linux x86_64 (amd64)
- Linux ARM64 (aarch64)
- macOS Intel (x86_64)
- macOS Apple Silicon (arm64)

### Windows

**One-line installation (PowerShell):**

Open PowerShell as Administrator and run:

```powershell
irm https://raw.githubusercontent.com/talkincode/sshmcp/main/install.ps1 | iex
```

**Or download and run the script:**

```powershell
# Download the installation script
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/talkincode/sshmcp/main/install.ps1" -OutFile "install.ps1"

# Run the installer (installs latest version)
.\install.ps1

# Or install a specific version
.\install.ps1 -Version v0.0.2

# Custom installation directory
.\install.ps1 -InstallDir "C:\Program Files\sshx"
```

**What the script does:**

1. Detects your architecture (amd64/arm64)
2. Downloads the appropriate binary from GitHub Releases
3. Extracts and installs to `%LOCALAPPDATA%\Programs\sshx`
4. Adds the installation directory to your PATH
5. Verifies the installation

**Note:** You may need to restart your terminal for PATH changes to take effect.

## Manual Installation

### Step 1: Download Binary

Visit the [Releases page](https://github.com/talkincode/sshmcp/releases) and download the appropriate file for your platform:

| Platform            | File                       |
| ------------------- | -------------------------- |
| Linux x86_64        | `sshx-linux-amd64.tar.gz`  |
| Linux ARM64         | `sshx-linux-arm64.tar.gz`  |
| macOS Intel         | `sshx-darwin-amd64.tar.gz` |
| macOS Apple Silicon | `sshx-darwin-arm64.tar.gz` |
| Windows x86_64      | `sshx-windows-amd64.zip`   |

### Step 2: Extract Archive

**Linux / macOS:**

```bash
tar -xzf sshx-<platform>-<arch>.tar.gz
```

**Windows:**

Right-click the ZIP file and select "Extract All..." or use PowerShell:

```powershell
Expand-Archive -Path sshx-windows-amd64.zip -DestinationPath .
```

### Step 3: Install Binary

**Linux / macOS:**

```bash
# Move to system binary directory
sudo mv sshx /usr/local/bin/

# Make executable
sudo chmod +x /usr/local/bin/sshx
```

**Windows:**

1. Create a directory: `C:\Program Files\sshx`
2. Move `sshx.exe` to this directory
3. Add the directory to your system PATH:
   - Right-click "This PC" → Properties → Advanced system settings
   - Click "Environment Variables"
   - Under "System variables", find "Path" and click "Edit"
   - Click "New" and add `C:\Program Files\sshx`
   - Click OK on all dialogs

### Step 4: Verify Installation

```bash
# Check if sshx is accessible
sshx --help

# Check version (if implemented)
sshx --version
```

## Build from Source

### Prerequisites

- Go 1.21 or later
- Git
- Make (optional, for using Makefile)

### Build Steps

```bash
# Clone the repository
git clone https://github.com/talkincode/sshmcp.git
cd sshmcp

# Download dependencies
go mod download

# Build the binary
go build -o bin/sshx ./cmd/sshx

# Or use Make
make build

# Test the binary
./bin/sshx --help
```

### Install System-wide

**Using Make:**

```bash
make install
```

This installs to:

- `$GOPATH/bin/sshx` (if GOPATH is set)
- `~/bin/sshx` (if ~/bin exists)

**Manual installation:**

```bash
# Copy to system directory
sudo cp bin/sshx /usr/local/bin/

# Make executable
sudo chmod +x /usr/local/bin/sshx
```

### Cross-compilation

Build for all platforms:

```bash
make build-all
```

This creates binaries in `bin/` directory:

- `sshx-linux-amd64`
- `sshx-linux-arm64`
- `sshx-darwin-amd64`
- `sshx-darwin-arm64`
- `sshx-windows-amd64.exe`

## Verification

After installation, verify that sshx is working correctly:

```bash
# Check if command is accessible
which sshx

# Display help
sshx --help

# Test with a simple command (replace with your server details)
sshx -h=192.168.1.100 -u=root "echo 'Hello from sshx'"
```

## Uninstallation

### Linux / macOS

**If installed via script or manually:**

```bash
sudo rm /usr/local/bin/sshx
```

**If installed via Make:**

```bash
make uninstall
```

### Windows

**If installed via script:**

1. Remove from PATH:
   - Open Environment Variables
   - Find and remove the sshx directory from PATH
2. Delete the installation directory:

```powershell
Remove-Item "$env:LOCALAPPDATA\Programs\sshx" -Recurse -Force
```

**If installed manually:**

1. Remove from PATH (see above)
2. Delete the directory where you placed `sshx.exe`

## Troubleshooting

### "sshx: command not found"

**Possible causes:**

1. The binary is not in your PATH
2. The installation directory is not in your PATH
3. You haven't restarted your terminal

**Solutions:**

- Check if the binary exists: `ls -l /usr/local/bin/sshx`
- Check your PATH: `echo $PATH`
- Add to PATH manually (Linux/macOS):
  ```bash
  echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.bashrc
  source ~/.bashrc
  ```
- Restart your terminal or shell

### Permission Denied

**Linux / macOS:**

```bash
# Make the binary executable
sudo chmod +x /usr/local/bin/sshx
```

**Windows:**

Run PowerShell as Administrator when installing.

### Download Failed

**Check your internet connection:**

```bash
# Test connection to GitHub
curl -I https://github.com
```

**Use alternative download method:**

If curl fails, try wget, or download manually from the browser.

### Installation Script Fails

**Linux / macOS:**

```bash
# Run with debug output
bash -x ./install.sh
```

**Windows:**

```powershell
# Run with verbose output
.\install.ps1 -Verbose
```

### SSL/TLS Certificate Errors

If you encounter certificate errors during download:

```bash
# Linux/macOS - install ca-certificates
sudo apt-get install ca-certificates  # Debian/Ubuntu
sudo yum install ca-certificates      # RHEL/CentOS

# macOS
brew install curl
```

## Platform-Specific Notes

### macOS

**First run security warning:**

macOS may block the binary because it's not signed. To allow it:

1. Try to run `sshx`
2. Go to System Preferences → Security & Privacy
3. Click "Allow Anyway" for sshx
4. Run `sshx` again and click "Open"

Or use this command:

```bash
sudo xattr -rd com.apple.quarantine /usr/local/bin/sshx
```

### Linux

**SELinux issues:**

If you're using SELinux (RHEL, CentOS, Fedora), you may need to adjust contexts:

```bash
sudo chcon -t bin_t /usr/local/bin/sshx
```

### Windows

**Windows Defender SmartScreen:**

Windows may show a SmartScreen warning. Click "More info" and then "Run anyway".

**PowerShell Execution Policy:**

If you can't run the PowerShell script:

```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

## Next Steps

After successful installation:

1. **Read the documentation:** Check [README.md](README.md) for usage examples
2. **Save passwords:** Use `sshx --set-password` to store credentials securely
3. **Try MCP mode:** Run `sshx mcp-stdio` to start the MCP server
4. **Configure your AI assistant:** Add sshx to your AI tool configuration

## Support

If you encounter issues not covered here:

- Check [GitHub Issues](https://github.com/talkincode/sshmcp/issues)
- Read [CHANGELOG.md](CHANGELOG.md) for version-specific notes
- Submit a new issue with details about your system and error messages
