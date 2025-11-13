package sshclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectInterpreter(t *testing.T) {
	client := &SSHClient{}

	tests := []struct {
		name           string
		remotePath     string
		expectedInterp string
	}{
		{
			name:           "Shell script .sh",
			remotePath:     "/tmp/script.sh",
			expectedInterp: "bash",
		},
		{
			name:           "Bash script .bash",
			remotePath:     "/tmp/script.bash",
			expectedInterp: "bash",
		},
		{
			name:           "Python script .py",
			remotePath:     "/tmp/script.py",
			expectedInterp: "python3",
		},
		{
			name:           "Python script .python",
			remotePath:     "/tmp/script.python",
			expectedInterp: "python3",
		},
		{
			name:           "Perl script .pl",
			remotePath:     "/tmp/script.pl",
			expectedInterp: "perl",
		},
		{
			name:           "Ruby script .rb",
			remotePath:     "/tmp/script.rb",
			expectedInterp: "ruby",
		},
		{
			name:           "No extension defaults to bash",
			remotePath:     "/tmp/script",
			expectedInterp: "bash",
		},
		{
			name:           "Unknown extension defaults to bash",
			remotePath:     "/tmp/script.txt",
			expectedInterp: "bash",
		},
		{
			name:           "Complex path",
			remotePath:     "/var/tmp/scripts/test.py",
			expectedInterp: "python3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := client.detectInterpreter(tt.remotePath)
			assert.Equal(t, tt.expectedInterp, interp)
		})
	}
}

func TestDetectInterpreter_EdgeCases(t *testing.T) {
	client := &SSHClient{}

	tests := []struct {
		name       string
		remotePath string
		expected   string
	}{
		{"Empty path", "", "bash"},
		{"Extension only", ".sh", "bash"},
		{"Extension only py", ".py", "python3"},
		{"Multiple dots", "script.test.sh", "bash"},
		{"Mixed case", "/tmp/Script.PY", "bash"}, // Case sensitive, doesn't match
		{"Dot prefix", "/tmp/.hidden.sh", "bash"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.detectInterpreter(tt.remotePath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSSHClient_NilClientHandling(t *testing.T) {
	client := &SSHClient{}

	// Test behavior without a connection
	assert.Nil(t, client.client)
	assert.Nil(t, client.sftpClient)
}
