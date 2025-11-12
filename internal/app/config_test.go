package app

import (
	"os"
	"testing"

	"sshx/internal/sshclient"
)

func TestParseArgs_BasicSSH(t *testing.T) {
	args := []string{"sshx", "-h=192.168.1.100", "uptime"}
	config := ParseArgs(args)

	if config.Host != "192.168.1.100" {
		t.Errorf("Expected host 192.168.1.100, got %s", config.Host)
	}
	if config.Command != "uptime" {
		t.Errorf("Expected command 'uptime', got %s", config.Command)
	}
	if config.Mode != "ssh" {
		t.Errorf("Expected mode 'ssh', got %s", config.Mode)
	}
}

func TestParseArgs_SSHWithPort(t *testing.T) {
	args := []string{"sshx", "-h=192.168.1.100", "-p=2222", "ls -la"}
	config := ParseArgs(args)

	if config.Host != "192.168.1.100" {
		t.Errorf("Expected host 192.168.1.100, got %s", config.Host)
	}
	if config.Port != "2222" {
		t.Errorf("Expected port 2222, got %s", config.Port)
	}
	if config.Command != "ls -la" {
		t.Errorf("Expected command 'ls -la', got %s", config.Command)
	}
}

func TestParseArgs_SSHWithUser(t *testing.T) {
	args := []string{"sshx", "-h=example.com", "-u=admin", "whoami"}
	config := ParseArgs(args)

	if config.User != "admin" {
		t.Errorf("Expected user 'admin', got %s", config.User)
	}
	if config.Host != "example.com" {
		t.Errorf("Expected host example.com, got %s", config.Host)
	}
}

func TestParseArgs_SSHWithKeyPath(t *testing.T) {
	args := []string{"sshx", "-h=192.168.1.100", "-i=/path/to/key", "uptime"}
	config := ParseArgs(args)

	if config.KeyPath != "/path/to/key" {
		t.Errorf("Expected key path '/path/to/key', got %s", config.KeyPath)
	}
}

func TestParseArgs_ForceFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"short form", []string{"sshx", "-h=host", "-f", "uptime"}},
		{"long form", []string{"sshx", "-h=host", "--force", "uptime"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ParseArgs(tt.args)
			if !config.Force {
				t.Errorf("Expected Force to be true")
			}
		})
	}
}

func TestParseArgs_NoSafetyCheck(t *testing.T) {
	args := []string{"sshx", "-h=host", "--no-safety-check", "uptime"}
	config := ParseArgs(args)

	if config.SafetyCheck {
		t.Errorf("Expected SafetyCheck to be false")
	}
}

func TestParseArgs_SFTPUpload(t *testing.T) {
	args := []string{"sshx", "-h=host", "--upload=local.txt", "--to=/remote/path.txt"}
	config := ParseArgs(args)

	if config.Mode != "sftp" {
		t.Errorf("Expected mode 'sftp', got %s", config.Mode)
	}
	if config.SftpAction != "upload" {
		t.Errorf("Expected sftp action 'upload', got %s", config.SftpAction)
	}
	if config.LocalPath != "local.txt" {
		t.Errorf("Expected local path 'local.txt', got %s", config.LocalPath)
	}
	if config.RemotePath != "/remote/path.txt" {
		t.Errorf("Expected remote path '/remote/path.txt', got %s", config.RemotePath)
	}
}

func TestParseArgs_SFTPDownload(t *testing.T) {
	args := []string{"sshx", "-h=host", "--download=/remote/file.log", "--to=./local.log"}
	config := ParseArgs(args)

	if config.Mode != "sftp" {
		t.Errorf("Expected mode 'sftp', got %s", config.Mode)
	}
	if config.SftpAction != "download" {
		t.Errorf("Expected sftp action 'download', got %s", config.SftpAction)
	}
	if config.RemotePath != "/remote/file.log" {
		t.Errorf("Expected remote path '/remote/file.log', got %s", config.RemotePath)
	}
	if config.LocalPath != "./local.log" {
		t.Errorf("Expected local path './local.log', got %s", config.LocalPath)
	}
}

func TestParseArgs_SFTPList(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"--list", []string{"sshx", "-h=host", "--list=/var/log"}},
		{"--ls", []string{"sshx", "-h=host", "--ls=/var/log"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ParseArgs(tt.args)
			if config.Mode != "sftp" {
				t.Errorf("Expected mode 'sftp', got %s", config.Mode)
			}
			if config.SftpAction != "list" {
				t.Errorf("Expected sftp action 'list', got %s", config.SftpAction)
			}
			if config.RemotePath != "/var/log" {
				t.Errorf("Expected remote path '/var/log', got %s", config.RemotePath)
			}
		})
	}
}

func TestParseArgs_SFTPMkdir(t *testing.T) {
	args := []string{"sshx", "-h=host", "--mkdir=/tmp/newdir"}
	config := ParseArgs(args)

	if config.Mode != "sftp" {
		t.Errorf("Expected mode 'sftp', got %s", config.Mode)
	}
	if config.SftpAction != "mkdir" {
		t.Errorf("Expected sftp action 'mkdir', got %s", config.SftpAction)
	}
	if config.RemotePath != "/tmp/newdir" {
		t.Errorf("Expected remote path '/tmp/newdir', got %s", config.RemotePath)
	}
}

func TestParseArgs_SFTPRemove(t *testing.T) {
	args := []string{"sshx", "-h=host", "--rm=/tmp/oldfile"}
	config := ParseArgs(args)

	if config.Mode != "sftp" {
		t.Errorf("Expected mode 'sftp', got %s", config.Mode)
	}
	if config.SftpAction != "remove" {
		t.Errorf("Expected sftp action 'remove', got %s", config.SftpAction)
	}
	if config.RemotePath != "/tmp/oldfile" {
		t.Errorf("Expected remote path '/tmp/oldfile', got %s", config.RemotePath)
	}
}

func TestParseArgs_PasswordSet(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedKey    string
		expectedValue  string
		expectedAction string
	}{
		{
			name:           "with value",
			args:           []string{"sshx", "--password-set=ma8:mypass"},
			expectedKey:    "ma8",
			expectedValue:  "mypass",
			expectedAction: "set",
		},
		{
			name:           "without value",
			args:           []string{"sshx", "--password-set=ma8"},
			expectedKey:    "ma8",
			expectedValue:  "",
			expectedAction: "set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ParseArgs(tt.args)
			if config.Mode != "password" {
				t.Errorf("Expected mode 'password', got %s", config.Mode)
			}
			if config.PasswordAction != tt.expectedAction {
				t.Errorf("Expected password action '%s', got %s", tt.expectedAction, config.PasswordAction)
			}
			if config.PasswordKey != tt.expectedKey {
				t.Errorf("Expected password key '%s', got %s", tt.expectedKey, config.PasswordKey)
			}
			if config.PasswordValue != tt.expectedValue {
				t.Errorf("Expected password value '%s', got %s", tt.expectedValue, config.PasswordValue)
			}
		})
	}
}

func TestParseArgs_PasswordGet(t *testing.T) {
	args := []string{"sshx", "--password-get=ma8"}
	config := ParseArgs(args)

	if config.Mode != "password" {
		t.Errorf("Expected mode 'password', got %s", config.Mode)
	}
	if config.PasswordAction != "get" {
		t.Errorf("Expected password action 'get', got %s", config.PasswordAction)
	}
	if config.PasswordKey != "ma8" {
		t.Errorf("Expected password key 'ma8', got %s", config.PasswordKey)
	}
}

func TestParseArgs_PasswordDelete(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"--password-delete", []string{"sshx", "--password-delete=testkey"}},
		{"--password-del", []string{"sshx", "--password-del=testkey"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ParseArgs(tt.args)
			if config.Mode != "password" {
				t.Errorf("Expected mode 'password', got %s", config.Mode)
			}
			if config.PasswordAction != "delete" {
				t.Errorf("Expected password action 'delete', got %s", config.PasswordAction)
			}
			if config.PasswordKey != "testkey" {
				t.Errorf("Expected password key 'testkey', got %s", config.PasswordKey)
			}
		})
	}
}

func TestParseArgs_PasswordCheck(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"--password-check", []string{"sshx", "--password-check=testkey"}},
		{"--password-exists", []string{"sshx", "--password-exists=testkey"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ParseArgs(tt.args)
			if config.Mode != "password" {
				t.Errorf("Expected mode 'password', got %s", config.Mode)
			}
			if config.PasswordAction != "check" {
				t.Errorf("Expected password action 'check', got %s", config.PasswordAction)
			}
			if config.PasswordKey != "testkey" {
				t.Errorf("Expected password key 'testkey', got %s", config.PasswordKey)
			}
		})
	}
}

func TestParseArgs_PasswordList(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"--password-list", []string{"sshx", "--password-list"}},
		{"--password-ls", []string{"sshx", "--password-ls"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ParseArgs(tt.args)
			if config.Mode != "password" {
				t.Errorf("Expected mode 'password', got %s", config.Mode)
			}
			if config.PasswordAction != "list" {
				t.Errorf("Expected password action 'list', got %s", config.PasswordAction)
			}
		})
	}
}

func TestParseArgs_EnvVariables(t *testing.T) {
	// Save original env
	origPassword := os.Getenv("SSH_PASSWORD")
	origKeyPath := os.Getenv("SSH_KEY_PATH")
	origNoSafety := os.Getenv("SSH_NO_SAFETY_CHECK")
	origForce := os.Getenv("SSH_FORCE")
	origSudoKey := os.Getenv("SSH_SUDO_KEY")

	// Cleanup
	defer func() {
		os.Setenv("SSH_PASSWORD", origPassword)
		os.Setenv("SSH_KEY_PATH", origKeyPath)
		os.Setenv("SSH_NO_SAFETY_CHECK", origNoSafety)
		os.Setenv("SSH_FORCE", origForce)
		os.Setenv("SSH_SUDO_KEY", origSudoKey)
	}()

	// Test password from env
	os.Setenv("SSH_PASSWORD", "envpass")
	os.Setenv("SSH_KEY_PATH", "/env/key/path")
	os.Setenv("SSH_NO_SAFETY_CHECK", "true")
	os.Setenv("SSH_FORCE", "true")
	os.Setenv("SSH_SUDO_KEY", "custom-sudo")

	args := []string{"sshx", "-h=host", "uptime"}
	config := ParseArgs(args)

	if config.Password != "envpass" {
		t.Errorf("Expected password from env 'envpass', got %s", config.Password)
	}
	if config.KeyPath != "/env/key/path" {
		t.Errorf("Expected key path from env '/env/key/path', got %s", config.KeyPath)
	}
	if config.SafetyCheck {
		t.Errorf("Expected SafetyCheck to be false from env")
	}
	if !config.Force {
		t.Errorf("Expected Force to be true from env")
	}
	if config.SudoKey != "custom-sudo" {
		t.Errorf("Expected sudo key 'custom-sudo', got %s", config.SudoKey)
	}
}

func TestParseArgs_DefaultSudoKey(t *testing.T) {
	// Clear SSH_SUDO_KEY
	origSudoKey := os.Getenv("SSH_SUDO_KEY")
	os.Unsetenv("SSH_SUDO_KEY")
	defer os.Setenv("SSH_SUDO_KEY", origSudoKey)

	args := []string{"sshx", "-h=host", "uptime"}
	config := ParseArgs(args)

	if config.SudoKey != sshclient.DefaultSudoKey {
		t.Errorf("Expected default sudo key '%s', got %s", sshclient.DefaultSudoKey, config.SudoKey)
	}
}

func TestParseArgs_DefaultValues(t *testing.T) {
	args := []string{"sshx", "-h=host", "uptime"}
	config := ParseArgs(args)

	if config.Mode != "ssh" {
		t.Errorf("Expected default mode 'ssh', got %s", config.Mode)
	}
	if !config.SafetyCheck {
		t.Errorf("Expected default SafetyCheck to be true")
	}
	if config.Force {
		t.Errorf("Expected default Force to be false")
	}
}

func TestParseArgs_LongFormOptions(t *testing.T) {
	args := []string{
		"sshx",
		"--host=example.com",
		"--port=2222",
		"--user=admin",
		"--key=/path/to/key",
		"uptime",
	}
	config := ParseArgs(args)

	if config.Host != "example.com" {
		t.Errorf("Expected host 'example.com', got %s", config.Host)
	}
	if config.Port != "2222" {
		t.Errorf("Expected port '2222', got %s", config.Port)
	}
	if config.User != "admin" {
		t.Errorf("Expected user 'admin', got %s", config.User)
	}
	if config.KeyPath != "/path/to/key" {
		t.Errorf("Expected key path '/path/to/key', got %s", config.KeyPath)
	}
}

func TestParseArgs_ComplexCommand(t *testing.T) {
	args := []string{
		"sshx",
		"-h=host",
		"ps aux | grep nginx | awk '{print $2}'",
	}
	config := ParseArgs(args)

	expected := "ps aux | grep nginx | awk '{print $2}'"
	if config.Command != expected {
		t.Errorf("Expected command '%s', got '%s'", expected, config.Command)
	}
}
