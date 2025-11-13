package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSSHConfig(t *testing.T) {
	// Create temporary SSH config file
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	configContent := `# SSH Config File
Host prod-web
    HostName 192.168.1.100
    User root
    Port 22
    IdentityFile ~/.ssh/id_rsa

Host dev-server
    HostName dev.example.com
    User developer
    Port 2222

Host simple-host
    HostName 10.0.0.1

# Wildcard entry (should be skipped)
Host *.example.com
    User admin

Host localhost-test
    HostName 127.0.0.1
    Port 22022
    IdentityFile ~/.ssh/special_key
`

	configPath := filepath.Join(sshDir, "config")
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write SSH config: %v", err)
	}

	// Set HOME to tmpDir
	oldHome := os.Getenv("HOME")
	t.Cleanup(func() {
		if err := os.Setenv("HOME", oldHome); err != nil {
			t.Logf("Warning: failed to restore HOME: %v", err)
		}
	})
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("Failed to set HOME: %v", err)
	}

	// Parse SSH config
	entries, err := ParseSSHConfig()
	if err != nil {
		t.Fatalf("ParseSSHConfig() error = %v", err)
	}

	// Verify entries
	if len(entries) != 5 {
		t.Errorf("Expected 5 entries, got %d", len(entries))
	}

	// Check first entry
	if entries[0].Host != "prod-web" {
		t.Errorf("Expected host 'prod-web', got '%s'", entries[0].Host)
	}
	if entries[0].HostName != "192.168.1.100" {
		t.Errorf("Expected hostname '192.168.1.100', got '%s'", entries[0].HostName)
	}
	if entries[0].User != "root" {
		t.Errorf("Expected user 'root', got '%s'", entries[0].User)
	}
	if entries[0].Port != "22" {
		t.Errorf("Expected port '22', got '%s'", entries[0].Port)
	}

	// Check second entry
	if entries[1].Host != "dev-server" {
		t.Errorf("Expected host 'dev-server', got '%s'", entries[1].Host)
	}
	if entries[1].Port != "2222" {
		t.Errorf("Expected port '2222', got '%s'", entries[1].Port)
	}

	// Check simple entry
	if entries[2].Host != "simple-host" {
		t.Errorf("Expected host 'simple-host', got '%s'", entries[2].Host)
	}
}

func TestParseSSHConfig_NotExist(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	t.Cleanup(func() {
		if err := os.Setenv("HOME", oldHome); err != nil {
			t.Logf("Warning: failed to restore HOME: %v", err)
		}
	})
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("Failed to set HOME: %v", err)
	}

	_, err := ParseSSHConfig()
	if err == nil {
		t.Error("ParseSSHConfig() should return error when config file doesn't exist")
	}
}

func TestConvertSSHConfigToHostConfig(t *testing.T) {
	tests := []struct {
		name     string
		entry    SSHConfigEntry
		expected HostConfig
	}{
		{
			name: "complete entry",
			entry: SSHConfigEntry{
				Host:         "test-host",
				HostName:     "192.168.1.100",
				User:         "root",
				Port:         "2222",
				IdentityFile: "/home/user/.ssh/id_rsa",
			},
			expected: HostConfig{
				Name: "test-host",
				Host: "192.168.1.100",
				User: "root",
				Port: "2222",
				Type: "linux",
			},
		},
		{
			name: "minimal entry",
			entry: SSHConfigEntry{
				Host: "minimal",
			},
			expected: HostConfig{
				Name: "minimal",
				Host: "minimal",
				User: "master",
				Port: "22",
				Type: "linux",
			},
		},
		{
			name: "entry with hostname",
			entry: SSHConfigEntry{
				Host:     "alias",
				HostName: "actual.example.com",
			},
			expected: HostConfig{
				Name: "alias",
				Host: "actual.example.com",
				User: "master",
				Port: "22",
				Type: "linux",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertSSHConfigToHostConfig(tt.entry)

			if result.Name != tt.expected.Name {
				t.Errorf("Name = %s, want %s", result.Name, tt.expected.Name)
			}
			if result.Host != tt.expected.Host {
				t.Errorf("Host = %s, want %s", result.Host, tt.expected.Host)
			}
			if result.User != tt.expected.User {
				t.Errorf("User = %s, want %s", result.User, tt.expected.User)
			}
			if result.Port != tt.expected.Port {
				t.Errorf("Port = %s, want %s", result.Port, tt.expected.Port)
			}
			if result.Type != tt.expected.Type {
				t.Errorf("Type = %s, want %s", result.Type, tt.expected.Type)
			}
		})
	}
}

func TestDetectSystemType(t *testing.T) {
	tests := []struct {
		hostname string
		expected string
	}{
		{"server.example.com", "linux"},
		{"windows-server", "windows"},
		{"win10-desktop", "windows"},
		{"mac-mini", "macos"},
		{"darwin-host", "macos"},
		{"ubuntu-server", "linux"},
		{"192.168.1.100", "linux"},
	}

	for _, tt := range tests {
		t.Run(tt.hostname, func(t *testing.T) {
			result := detectSystemType(tt.hostname)
			if result != tt.expected {
				t.Errorf("detectSystemType(%s) = %s, want %s", tt.hostname, result, tt.expected)
			}
		})
	}
}

func TestImportFromSSHConfig(t *testing.T) {
	// Create temporary SSH config file
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	configContent := `Host server1
    HostName 192.168.1.10
    User admin

Host server2
    HostName 192.168.1.20
    User root
    Port 2222

Host *.wildcard
    User test
`

	configPath := filepath.Join(sshDir, "config")
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write SSH config: %v", err)
	}

	oldHome := os.Getenv("HOME")
	t.Cleanup(func() {
		if err := os.Setenv("HOME", oldHome); err != nil {
			t.Logf("Warning: failed to restore HOME: %v", err)
		}
	})
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("Failed to set HOME: %v", err)
	}

	// Test import
	settings := &Settings{
		Hosts: make([]HostConfig, 0),
	}

	imported, err := ImportFromSSHConfig(settings, false)
	if err != nil {
		t.Fatalf("ImportFromSSHConfig() error = %v", err)
	}

	if imported != 2 {
		t.Errorf("Expected 2 hosts imported (wildcard should be skipped), got %d", imported)
	}

	if len(settings.Hosts) != 2 {
		t.Errorf("Expected 2 hosts in settings, got %d", len(settings.Hosts))
	}

	// Verify imported hosts
	server1, err := GetHost(settings, "server1")
	if err != nil {
		t.Errorf("Failed to get server1: %v", err)
	} else {
		if server1.Host != "192.168.1.10" {
			t.Errorf("server1 host = %s, want 192.168.1.10", server1.Host)
		}
		if server1.User != "admin" {
			t.Errorf("server1 user = %s, want admin", server1.User)
		}
	}

	server2, err := GetHost(settings, "server2")
	if err != nil {
		t.Errorf("Failed to get server2: %v", err)
	} else {
		if server2.Port != "2222" {
			t.Errorf("server2 port = %s, want 2222", server2.Port)
		}
	}
}

func TestImportFromSSHConfig_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	configContent := `Host existing-host
    HostName 192.168.1.100
    User newuser
`

	configPath := filepath.Join(sshDir, "config")
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write SSH config: %v", err)
	}

	oldHome := os.Getenv("HOME")
	t.Cleanup(func() {
		if err := os.Setenv("HOME", oldHome); err != nil {
			t.Logf("Warning: failed to restore HOME: %v", err)
		}
	})
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("Failed to set HOME: %v", err)
	}

	// Create settings with existing host
	settings := &Settings{
		Hosts: []HostConfig{
			{
				Name: "existing-host",
				Host: "old.example.com",
				User: "olduser",
			},
		},
	}

	// Import without overwrite (should skip)
	imported, err := ImportFromSSHConfig(settings, false)
	if err != nil {
		t.Fatalf("ImportFromSSHConfig() error = %v", err)
	}

	if imported != 0 {
		t.Errorf("Expected 0 hosts imported (should skip existing), got %d", imported)
	}

	host, getErr := GetHost(settings, "existing-host")
	if getErr != nil {
		t.Fatalf("GetHost() error = %v", getErr)
	}
	if host.User != "olduser" {
		t.Errorf("Host should not be updated, user = %s, want olduser", host.User)
	}

	// Import with overwrite
	imported, err = ImportFromSSHConfig(settings, true)
	if err != nil {
		t.Fatalf("ImportFromSSHConfig() error = %v", err)
	}

	if imported != 1 {
		t.Errorf("Expected 1 host imported (should overwrite existing), got %d", imported)
	}

	host, getErr = GetHost(settings, "existing-host")
	if getErr != nil {
		t.Fatalf("GetHost() error = %v", getErr)
	}
	if host.User != "newuser" {
		t.Errorf("Host should be updated, user = %s, want newuser", host.User)
	}
	if host.Host != "192.168.1.100" {
		t.Errorf("Host should be updated, host = %s, want 192.168.1.100", host.Host)
	}
}
