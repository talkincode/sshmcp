package app

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/joho/godotenv"

	"sshx/internal/mcp"
	"sshx/internal/sshclient"
)

// ErrUsage is returned when only the usage information was printed.
var ErrUsage = errors.New("usage displayed")

// Run executes the CLI using the provided arguments (typically os.Args).
func Run(args []string) (err error) {
	// Handle MCP stdio mode
	if len(args) >= 2 && (args[1] == "mcp-stdio" || args[1] == "--mcp-stdio") {
		log.SetOutput(io.Discard)

		server := mcp.NewMCPServer()
		if err := server.Start(); err != nil {
			return err
		}
		return nil
	}

	// Handle usage
	if len(args) < 2 {
		PrintUsage()
		return ErrUsage
	}

	// Load environment variables
	//nolint:errcheck // Loading .env is optional
	_ = godotenv.Load()

	// Parse command-line arguments
	config := ParseArgs(args)

	// Handle password management mode
	if config.Mode == "password" {
		if err := HandlePasswordManagement(config); err != nil {
			return fmt.Errorf("password management failed: %w", err)
		}
		return nil
	}

	// Auto-fill sudo password if needed
	if strings.Contains(config.Command, "sudo") && config.SudoKey != "" {
		password, err := sshclient.GetSudoPassword(config.SudoKey)
		if err != nil {
			log.Printf("Warning: failed to get sudo password from keyring: %v", err)
			log.Println("Continuing without sudo password auto-fill...")
		} else {
			config.Password = password
			log.Printf("✓ Sudo password will be auto-filled when prompted")
		}
	}

	// Create SSH client
	client, err := sshclient.NewSSHClient(config)
	if err != nil {
		return fmt.Errorf("failed to create SSH client: %w", err)
	}
	defer sshclient.CloseIgnore(&err, client)

	// Connect to remote host
	if err = client.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Handle SFTP mode
	if config.Mode == "sftp" {
		if err = client.ExecuteSftp(); err != nil {
			return fmt.Errorf("SFTP operation failed: %w", err)
		}
		return nil
	}

	// Handle SSH command execution
	if err = client.ExecuteCommand(); err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}

	return nil
}
