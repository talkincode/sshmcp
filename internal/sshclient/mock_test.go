package sshclient

import (
	"fmt"
	"io"

	"golang.org/x/crypto/ssh"
)

// MockSSHClient is a mock implementation of ssh.Client for testing
type MockSSHClient struct {
	ShouldFail      bool
	SessionFailures int
	sessionCount    int
}

// MockSSHSession is a mock implementation of ssh.Session for testing
type MockSSHSession struct {
	ShouldFail bool
	Stdout     io.Writer
	Stderr     io.Writer
}

// NewMockSSHClient creates a new mock SSH client
func NewMockSSHClient(shouldFail bool) *MockSSHClient {
	return &MockSSHClient{
		ShouldFail:   shouldFail,
		sessionCount: 0,
	}
}

// NewSession creates a mock SSH session
func (m *MockSSHClient) NewSession() (*ssh.Session, error) {
	m.sessionCount++

	// Simulate session creation failure after N attempts
	if m.SessionFailures > 0 && m.sessionCount > m.SessionFailures {
		return nil, fmt.Errorf("mock: session creation failed after %d attempts", m.SessionFailures)
	}

	if m.ShouldFail {
		return nil, fmt.Errorf("mock: failed to create session")
	}

	// Note: We can't actually return a valid *ssh.Session from a mock
	// This is a limitation - for real testing, you'd need to use an interface
	// or a test SSH server
	return nil, fmt.Errorf("mock: cannot create real ssh.Session in mock")
}

// Close simulates closing the SSH client
func (m *MockSSHClient) Close() error {
	if m.ShouldFail {
		return fmt.Errorf("mock: failed to close client")
	}
	return nil
}

// Run simulates running a command
func (m *MockSSHSession) Run(cmd string) error {
	if m.ShouldFail {
		return fmt.Errorf("mock: command execution failed")
	}

	// Simulate successful execution
	if m.Stdout != nil && cmd == "echo ping" {
		if _, err := m.Stdout.Write([]byte("ping\n")); err != nil {
			return fmt.Errorf("mock: failed to write output: %w", err)
		}
	}

	return nil
}

// Close simulates closing the session
func (m *MockSSHSession) Close() error {
	if m.ShouldFail {
		return fmt.Errorf("mock: failed to close session")
	}
	return nil
}
