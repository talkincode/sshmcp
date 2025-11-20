package sshclient

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
)

func TestNewSSHClient(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		checkFunc   func(*testing.T, *SSHClient, *Config)
	}{
		{
			name: "Complete configuration",
			config: &Config{
				Host:       "192.168.1.100",
				Port:       "2222",
				User:       "admin",
				KeyPath:    "/path/to/key",
				UseKeyAuth: true,
			},
			expectError: false,
			checkFunc: func(t *testing.T, client *SSHClient, config *Config) {
				assert.Equal(t, "192.168.1.100", client.config.Host)
				assert.Equal(t, "2222", client.config.Port)
				assert.Equal(t, "admin", client.config.User)
				assert.Equal(t, "/path/to/key", client.config.KeyPath)
			},
		},
		{
			name: "Using default values",
			config: &Config{
				Host:       "example.com",
				UseKeyAuth: true,
			},
			expectError: false,
			checkFunc: func(t *testing.T, client *SSHClient, config *Config) {
				assert.Equal(t, "example.com", client.config.Host)
				assert.Equal(t, DefaultSSHPort, client.config.Port)
				assert.Equal(t, DefaultSSHUser, client.config.User)
				// KeyPath should be set to default ~/.ssh/id_rsa
				home, err := os.UserHomeDir()
				if err != nil {
					t.Fatalf("Failed to get user home dir: %v", err)
				}
				expectedKeyPath := filepath.Join(home, ".ssh", "id_rsa")
				assert.Equal(t, expectedKeyPath, client.config.KeyPath)
			},
		},
		{
			name: "Missing Host",
			config: &Config{
				UseKeyAuth: true,
			},
			expectError: true,
			checkFunc:   nil,
		},
		{
			name: "Custom port and user",
			config: &Config{
				Host:       "test.server.com",
				Port:       "8022",
				User:       "testuser",
				UseKeyAuth: true,
			},
			expectError: false,
			checkFunc: func(t *testing.T, client *SSHClient, config *Config) {
				assert.Equal(t, "test.server.com", client.config.Host)
				assert.Equal(t, "8022", client.config.Port)
				assert.Equal(t, "testuser", client.config.User)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewSSHClient(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				if tt.checkFunc != nil {
					tt.checkFunc(t, client, tt.config)
				}
			}
		})
	}
}

func TestNewSSHClient_DefaultKeyPath(t *testing.T) {
	config := &Config{
		Host:       "test.com",
		UseKeyAuth: true,
	}

	client, err := NewSSHClient(config)

	assert.NoError(t, err)
	assert.NotNil(t, client)

	// Verify default KeyPath setting
	home, homeErr := os.UserHomeDir()
	if homeErr == nil {
		expectedKeyPath := filepath.Join(home, ".ssh", "id_rsa")
		assert.Equal(t, expectedKeyPath, client.config.KeyPath)
	}
}

func TestNewSSHClient_KeyAuthDisabled(t *testing.T) {
	config := &Config{
		Host:       "nokey.test",
		Password:   "secret",
		UseKeyAuth: false,
	}

	client, err := NewSSHClient(config)

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "", client.config.KeyPath)
}

func TestConfig_Defaults(t *testing.T) {
	config := &Config{
		Host:       "testhost",
		UseKeyAuth: true,
	}

	client, err := NewSSHClient(config)
	assert.NoError(t, err)

	// Verify default values
	assert.Equal(t, DefaultSSHPort, client.config.Port)
	assert.Equal(t, DefaultSSHUser, client.config.User)
}

func TestConfig_CustomValues(t *testing.T) {
	config := &Config{
		Host:        "custom.host",
		Port:        "9999",
		User:        "customuser",
		Password:    "custompass",
		KeyPath:     "/custom/path/key",
		SudoKey:     "customsudo",
		Command:     "ls -la",
		Mode:        "ssh",
		SafetyCheck: true,
		Force:       false,
		UseKeyAuth:  true,
	}

	client, err := NewSSHClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	// Verify all custom values
	assert.Equal(t, "custom.host", client.config.Host)
	assert.Equal(t, "9999", client.config.Port)
	assert.Equal(t, "customuser", client.config.User)
	assert.Equal(t, "custompass", client.config.Password)
	assert.Equal(t, "/custom/path/key", client.config.KeyPath)
	assert.Equal(t, "customsudo", client.config.SudoKey)
	assert.Equal(t, "ls -la", client.config.Command)
	assert.Equal(t, "ssh", client.config.Mode)
	assert.True(t, client.config.SafetyCheck)
	assert.False(t, client.config.Force)
}

func TestSSHClient_NilConfig(t *testing.T) {
	config := &Config{
		Host:       "",
		UseKeyAuth: true,
	}

	client, err := NewSSHClient(config)

	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "host is required")
}

func TestConfig_SFTPFields(t *testing.T) {
	config := &Config{
		Host:       "sftp.server",
		Port:       "22",
		User:       "sftpuser",
		SftpAction: "upload",
		LocalPath:  "/local/file.txt",
		RemotePath: "/remote/file.txt",
		UseKeyAuth: true,
	}

	client, err := NewSSHClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	assert.Equal(t, "upload", client.config.SftpAction)
	assert.Equal(t, "/local/file.txt", client.config.LocalPath)
	assert.Equal(t, "/remote/file.txt", client.config.RemotePath)
}

func TestConfig_PasswordFields(t *testing.T) {
	config := &Config{
		Host:           "password.server",
		PasswordAction: "set",
		PasswordKey:    "mykey",
		PasswordValue:  "myvalue",
		UseKeyAuth:     true,
	}

	client, err := NewSSHClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	assert.Equal(t, "set", client.config.PasswordAction)
	assert.Equal(t, "mykey", client.config.PasswordKey)
	assert.Equal(t, "myvalue", client.config.PasswordValue)
}

func TestConstants(t *testing.T) {
	assert.Equal(t, "22", DefaultSSHPort)
	assert.Equal(t, "master", DefaultSSHUser)
	assert.Equal(t, "master", DefaultSudoKey)
	assert.Equal(t, "[sudo] password", SudoPrompt)
	assert.Equal(t, ": ", PasswordPromptEnd)
}

func TestSSHClient_InitialState(t *testing.T) {
	config := &Config{
		Host:       "test.com",
		UseKeyAuth: true,
	}

	client, err := NewSSHClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	// 验证初始状态
	assert.NotNil(t, client.config)
	assert.Nil(t, client.client) // 未连接
	assert.Nil(t, client.sftpClient)
}

func TestConfig_MultipleHosts(t *testing.T) {
	hosts := []string{"host1.com", "host2.com", "192.168.1.1"}

	for _, host := range hosts {
		config := &Config{Host: host, UseKeyAuth: true}
		client, err := NewSSHClient(config)

		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, host, client.config.Host)
	}
}

func TestShouldFallbackToPassword(t *testing.T) {
	authErr := &ssh.ServerAuthError{Errors: []error{fmt.Errorf("publickey denied")}}

	t.Run("requires key auth present", func(t *testing.T) {
		assert.False(t, shouldFallbackToPassword(authErr, false, true))
	})

	t.Run("requires password available", func(t *testing.T) {
		assert.False(t, shouldFallbackToPassword(authErr, true, false))
	})

	t.Run("auth error triggers fallback", func(t *testing.T) {
		assert.True(t, shouldFallbackToPassword(authErr, true, true))
	})

	t.Run("nil error", func(t *testing.T) {
		assert.False(t, shouldFallbackToPassword(nil, true, true))
	})
}
