package app

import "fmt"

// PrintUsage prints the usage information for the sshx command.
func PrintUsage() {
	fmt.Println(`
SSH & SFTP Remote Tool with Password Manager (Cross-Platform)

Usage:
  sshx mcp-stdio                                  # MCP stdio mode (for AI assistants)
  sshx -h=<host> [options] <command>              # SSH mode
  sshx -h=<host> [options] --upload=<file>        # SFTP upload
  sshx -h=<host> [options] --download=<file>      # SFTP download
  sshx --password-set=<key>[:<value>]             # Password management
  sshx --password-get=<key>                       # Get password
  sshx --password-list                            # List passwords

MCP Mode:
  sshx mcp-stdio            Start MCP server in stdio mode
  sshx --mcp-stdio          Alternative MCP mode flag

  MCP Tools Available:
    - ssh_execute           Execute SSH commands with sudo support
    - sftp_upload           Upload files via SFTP
    - sftp_download         Download files via SFTP
    - sftp_list             List directory contents
    - sftp_mkdir            Create remote directory
    - sftp_remove           Remove files/directories
    - password_set          Store password in system keyring
    - password_get          Retrieve password from keyring
    - password_delete       Delete password from keyring
    - password_list         List common password keys

SSH Options:
  -h, --host=HOST       Remote host address (required)
  -p, --port=PORT       SSH port (default: 22)
  -u, --user=USER       SSH username (default: master)
  -i, --key=PATH        SSH private key path (default: ~/.ssh/id_rsa)
  --help                Show this help message

Safety Options:
  -f, --force           Force execution, bypass safety checks (use with caution!)
  --no-safety-check     Disable safety checks completely (not recommended)

  Safety checks protect against:
    - Destructive operations (rm -rf /, mkfs, dd)
    - System shutdown/reboot commands
    - Critical file modifications (/etc/passwd, /etc/shadow)
    - Dangerous pipe operations (curl | sh)
    - Fork bombs and other malicious patterns

SFTP Options:
  --upload=<local>      Upload file (use with --to=<remote>)
  --download=<remote>   Download file (use with --to=<local>)
  --to=<path>           Target path for upload/download
  --list=<path>         List directory contents (alias: --ls)
  --mkdir=<path>        Create remote directory
  --rm=<path>           Remove remote file or directory

Password Management (Cross-Platform):
  --password-set=<key>[:<password>]   Set password in system keyring
                                      If password omitted, will prompt
  --password-get=<key>                Get password from keyring
  --password-check=<key>              Check if password exists (alias: --password-exists)
  --password-delete=<key>             Delete password from keyring (alias: --password-del)
  --password-list                     List common password keys (alias: --password-ls)

  Platform Support:
    macOS:   Uses Keychain
    Linux:   Uses Secret Service (gnome-keyring/kwallet)
    Windows: Uses Credential Manager

Environment Variables (.env):
  SSH_PASSWORD          SSH password (not recommended, use SSH keys or keyring)
  SSH_KEY_PATH          SSH private key path
  SSH_SUDO_KEY          Sudo password keyring key name (default: master)
  SSH_NO_SAFETY_CHECK   Disable safety checks (true/false)
  SSH_FORCE             Force execution mode (true/false)

SSH Examples:
  # Execute simple command (default user: master)
  sshx -h=192.168.1.100 "uptime"

  # Execute sudo command (auto password from keyring: master)
  sshx -h=192.168.1.100 "sudo systemctl status docker"

  # Custom SSH port
  sshx -h=192.168.1.100 -p=2222 "ps aux | grep nginx"

  # Dangerous command will be blocked
  sshx -h=192.168.1.100 "sudo rm -rf /tmp/*"  # Safe
  sshx -h=192.168.1.100 "sudo rm -rf /"       # ⚠️ BLOCKED!

  # Force execute (bypass safety check - use with caution!)
  sshx -h=192.168.1.100 --force "sudo reboot"
  sshx -h=192.168.1.100 -f "sudo systemctl reboot"

SFTP Examples:
  # Upload file
  sshx -h=192.168.1.100 --upload=local.txt --to=/tmp/remote.txt

  # Download file
  sshx -h=192.168.1.100 --download=/var/log/app.log --to=./app.log

  # List directory
  sshx -h=192.168.1.100 --list=/var/log

  # Create directory
  sshx -h=192.168.1.100 --mkdir=/tmp/newdir

  # Remove file
  sshx -h=192.168.1.100 --rm=/tmp/oldfile.txt

  # Batch upload
  for file in *.txt; do
    sshx -h=192.168.1.100 --upload=$file --to=/backup/$file
  done

Password Management Examples:
  # Set sudo password (interactive prompt)
  sshx --password-set=master

  # Set sudo password (inline, not recommended for security)
  sshx --password-set=master:mypassword

  # Set custom password
  sshx --password-set=myserver

  # Get password
  sshx --password-get=master

  # Check if password exists
  sshx --password-check=test-password

  # List common password keys
  sshx --password-list

  # Delete password
  sshx --password-delete=master

  # Set password for specific server
  sshx --password-set=prod-server:secretpass
  sshx -h=prod-server "sudo reboot"  # Will use ENV SSH_SUDO_KEY or master

Note:
  - SSH key authentication is tried first, then password authentication
  - Sudo password is automatically retrieved from system keyring
  - SFTP operations use the same SSH connection
  - Password manager works across macOS/Linux/Windows
  - Default user: master, Default sudo key: master`)
}
