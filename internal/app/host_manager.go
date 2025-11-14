package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/talkincode/sshmcp/internal/sshclient"
	"github.com/talkincode/sshmcp/pkg/errutil"
	"github.com/talkincode/sshmcp/pkg/logger"
)

// HandleHostManagement handles host management commands
func HandleHostManagement(config *sshclient.Config) error {
	switch config.HostAction {
	case "add":
		return handleHostAdd(config)
	case "update":
		return handleHostUpdate(config)
	case "list":
		return handleHostList(config)
	case "test":
		return handleHostTest(config)
	case "remove":
		return handleHostRemove(config)
	default:
		return fmt.Errorf("unknown host action: %s", config.HostAction)
	}
}

// handleHostAdd adds a new host to settings
func handleHostAdd(config *sshclient.Config) error {
	// Load settings
	settings, err := LoadSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	var host HostConfig

	// If host configuration is provided via command line
	if config.HostName != "" {
		host = HostConfig{
			Name:        config.HostName,
			Description: config.HostDescription,
			Host:        config.Host,
			Port:        config.Port,
			User:        config.User,
			PasswordKey: config.SudoKey,
			Type:        config.HostType,
		}
	} else {
		// Interactive mode
		reader := bufio.NewReader(os.Stdin)

		fmt.Println("=== Add New Host ===")

		// Host name (required)
		fmt.Print("Host name (unique identifier): ")
		name, readErr := reader.ReadString('\n')
		if readErr != nil {
			return fmt.Errorf("failed to read host name: %w", readErr)
		}
		host.Name = strings.TrimSpace(name)

		// Host address (required)
		fmt.Print("Host address (IP or hostname): ")
		addr, readErr := reader.ReadString('\n')
		if readErr != nil {
			return fmt.Errorf("failed to read host address: %w", readErr)
		}
		host.Host = strings.TrimSpace(addr)

		// Description (optional)
		fmt.Print("Description (optional): ")
		if desc, err := reader.ReadString('\n'); err == nil {
			host.Description = strings.TrimSpace(desc)
		}

		// Port (optional, default: 22)
		fmt.Print("Port (default: 22): ")
		if port, err := reader.ReadString('\n'); err == nil {
			host.Port = strings.TrimSpace(port)
		}

		// User (optional, default: master)
		fmt.Print("User (default: master): ")
		if user, err := reader.ReadString('\n'); err == nil {
			host.User = strings.TrimSpace(user)
		}

		// Password key (optional)
		fmt.Print("Password key (optional): ")
		if pwdKey, err := reader.ReadString('\n'); err == nil {
			host.PasswordKey = strings.TrimSpace(pwdKey)
		}

		// Type (optional, default: linux)
		fmt.Print("System type [linux/windows/macos] (default: linux): ")
		if sysType, err := reader.ReadString('\n'); err == nil {
			host.Type = strings.TrimSpace(sysType)
		}
		if host.Type == "" {
			host.Type = "linux"
		}
	}

	// Add host to settings
	if err := AddHost(settings, host); err != nil {
		return fmt.Errorf("failed to add host: %w", err)
	}

	// Save settings
	if err := SaveSettings(settings); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	logger.GetLogger().Success("Host '%s' added successfully", host.Name)
	return nil
}

// handleHostUpdate updates an existing host in settings
func handleHostUpdate(config *sshclient.Config) error {
	// Load settings
	settings, err := LoadSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	// Host name is required for update
	if config.HostName == "" {
		return fmt.Errorf("host name is required for update (use --host-name=<name>)")
	}

	// Check if host exists
	_, err = GetHost(settings, config.HostName)
	if err != nil {
		return fmt.Errorf("host '%s' not found, use --host-add to create it", config.HostName)
	}

	// Build updated host config
	host := HostConfig{
		Name: config.HostName,
	}

	// Update fields if provided, otherwise keep existing values
	existingHost, err := GetHost(settings, config.HostName)
	if err != nil {
		// If host doesn't exist, we'll create it with only the provided fields
		existingHost = &HostConfig{}
	}

	if config.Host != "" {
		host.Host = config.Host
	} else {
		host.Host = existingHost.Host
	}

	if config.HostDescription != "" {
		host.Description = config.HostDescription
	} else {
		host.Description = existingHost.Description
	}

	if config.Port != "" && config.Port != "22" {
		host.Port = config.Port
	} else if existingHost.Port != "" {
		host.Port = existingHost.Port
	} else {
		host.Port = "22"
	}

	if config.User != "" && config.User != "master" {
		host.User = config.User
	} else if existingHost.User != "" {
		host.User = existingHost.User
	} else {
		host.User = "master"
	}

	if config.SudoKey != "" && config.SudoKey != sshclient.DefaultSudoKey {
		host.PasswordKey = config.SudoKey
	} else if existingHost.PasswordKey != "" {
		host.PasswordKey = existingHost.PasswordKey
	}

	if config.HostType != "" {
		host.Type = config.HostType
	} else if existingHost.Type != "" {
		host.Type = existingHost.Type
	} else {
		host.Type = "linux"
	}

	// Update host
	if err := UpdateHost(settings, host); err != nil {
		return fmt.Errorf("failed to update host: %w", err)
	}

	// Save settings
	if err := SaveSettings(settings); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	logger.GetLogger().Success("Host '%s' updated successfully", host.Name)
	return nil
}

// handleHostList lists all configured hosts
func handleHostList(config *sshclient.Config) error {
	// Load settings
	settings, err := LoadSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	hosts := ListHosts(settings)

	if len(hosts) == 0 {
		fmt.Println("No hosts configured.")
		fmt.Println("\nTo add hosts:")
		fmt.Println("  - Interactive: sshx --host-add")
		return nil
	}

	// Detailed mode
	fmt.Printf("\n=== Configured Hosts (%d) ===\n\n", len(hosts))

	for i, host := range hosts {
		fmt.Printf("[%d] %s\n", i+1, host.Name)
		fmt.Printf("    Host:        %s\n", host.Host)
		if host.Description != "" {
			fmt.Printf("    Description: %s\n", host.Description)
		}
		if host.Port != "" && host.Port != "22" {
			fmt.Printf("    Port:        %s\n", host.Port)
		}
		if host.User != "" {
			fmt.Printf("    User:        %s\n", host.User)
		}
		if host.PasswordKey != "" {
			fmt.Printf("    Password Key: %s\n", host.PasswordKey)
		}
		if host.Type != "" {
			fmt.Printf("    Type:        %s\n", host.Type)
		}
		fmt.Println()
	}

	fmt.Println("Usage:")
	fmt.Printf("  sshx -h=%s \"command\"\n", hosts[0].Name)
	fmt.Printf("  sshx --host-test %s\n", hosts[0].Name)

	return nil
}

// handleHostTest tests host connection
func handleHostTest(config *sshclient.Config) error {
	// Load settings
	settings, err := LoadSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	if config.HostName == "" {
		return fmt.Errorf("host name is required for test")
	}

	// Get host configuration
	hostConfig, err := GetHost(settings, config.HostName)
	if err != nil {
		return fmt.Errorf("host not found: %w", err)
	}

	logger.GetLogger().Info("Testing connection to '%s' (%s)...", hostConfig.Name, hostConfig.Host)

	// Create SSH config for testing
	testConfig := &sshclient.Config{
		Host: hostConfig.Host,
		Port: hostConfig.Port,
		User: hostConfig.User,
	}

	// Get default SSH key if not specified
	if settings.Key != "" {
		testConfig.KeyPath = settings.Key
	}

	// Try to get password if password key is configured
	if hostConfig.PasswordKey != "" {
		password, pwdErr := sshclient.GetSudoPassword(hostConfig.PasswordKey)
		if pwdErr != nil {
			logger.GetLogger().Warning("failed to get password from keyring: %v", pwdErr)
		} else {
			testConfig.Password = password
		}
	}

	// Create SSH client
	client, err := sshclient.NewSSHClient(testConfig)
	if err != nil {
		return fmt.Errorf("failed to create SSH client: %w", err)
	}
	defer errutil.HandleCloseError(&err, client)

	// Test connection
	if connectErr := client.Connect(); connectErr != nil {
		logger.GetLogger().Error("Connection failed: %v", connectErr)
		return fmt.Errorf("connection test failed")
	}

	// Test command execution
	testConfig.Command = "echo 'Connection test successful'"
	client2, err := sshclient.NewSSHClient(testConfig)
	if err != nil {
		return fmt.Errorf("failed to create test client: %w", err)
	}
	defer errutil.HandleCloseError(&err, client2)

	if connectErr := client2.Connect(); connectErr != nil {
		return fmt.Errorf("failed to connect: %w", connectErr)
	}

	output, err := client2.ExecuteCommandWithOutput()
	if err != nil {
		logger.GetLogger().Error("Command execution failed: %v", err)
		return fmt.Errorf("command execution test failed")
	}

	logger.GetLogger().Success("Connection successful!")
	logger.GetLogger().Success("Command execution successful!")
	fmt.Printf("\nTest output: %s\n", strings.TrimSpace(output))

	return nil
}

// handleHostRemove removes a host from settings
func handleHostRemove(config *sshclient.Config) error {
	// Load settings
	settings, err := LoadSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	if config.HostName == "" {
		return fmt.Errorf("host name is required for removal")
	}

	// Remove host
	if err := RemoveHost(settings, config.HostName); err != nil {
		return fmt.Errorf("failed to remove host: %w", err)
	}

	// Save settings
	if err := SaveSettings(settings); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	logger.GetLogger().Success("Host '%s' removed successfully", config.HostName)
	return nil
}
