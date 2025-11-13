//go:build integration

package sshclient

import (
	"testing"
)

// TestGetConnection_ThreadSafety_Integration tests concurrent GetConnection calls with real SSH
// Run with: go test -tags=integration -v ./internal/sshclient -run TestGetConnection_ThreadSafety_Integration
func TestGetConnection_ThreadSafety_Integration(t *testing.T) {
	pool := NewConnectionPool()
	config := &Config{
		Host: "localhost", // Change to your test SSH server
		Port: "22",
		User: "testuser",
	}

	// Launch multiple goroutines trying to get connection
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("GetConnection panicked: %v", r)
				}
				done <- true
			}()

			// This will fail without proper SSH server configuration
			_, _ = pool.GetConnection(config)
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Test passed if no panics occurred
}

// TestIsConnectionAlive_RealSSH tests the health check with a real SSH connection
// Run with: go test -tags=integration -v ./internal/sshclient -run TestIsConnectionAlive_RealSSH
func TestIsConnectionAlive_RealSSH(t *testing.T) {
	t.Skip("Requires a configured SSH server - set up test environment first")

	// This test requires a real SSH server to be configured
	// Uncomment and configure when you have a test SSH server available
	/*
		pool := NewConnectionPool()
		config := &Config{
			Host:    "localhost",
			Port:    "22",
			User:    "testuser",
			KeyPath: "~/.ssh/id_rsa",
		}

		client, err := pool.GetConnection(config)
		require.NoError(t, err)

		// Test that the connection is alive
		assert.True(t, pool.isConnectionAlive(client))

		// Close and verify it's dead
		client.Close()
		assert.False(t, pool.isConnectionAlive(client))
	*/
}
