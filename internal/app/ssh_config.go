package app

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SSHConfigEntry represents a parsed SSH config entry
type SSHConfigEntry struct {
	Host         string
	HostName     string
	User         string
	Port         string
	IdentityFile string
}

// GetSSHConfigPath returns the path to the SSH config file
func GetSSHConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, ".ssh", "config"), nil
}

// ParseSSHConfig parses the SSH config file and returns a list of host entries
func ParseSSHConfig() ([]SSHConfigEntry, error) {
	configPath, err := GetSSHConfigPath()
	if err != nil {
		return nil, err
	}

	// Check if config file exists
	if _, statErr := os.Stat(configPath); os.IsNotExist(statErr) {
		return nil, fmt.Errorf("SSH config file not found at %s", configPath)
	}

	file, err := os.Open(configPath) // #nosec G304 -- SSH config path is from user's home directory
	if err != nil {
		return nil, fmt.Errorf("failed to open SSH config file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Best-effort close, ignore error
			_ = closeErr
		}
	}()

	var entries []SSHConfigEntry
	var currentEntry *SSHConfigEntry

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split line into key and value
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		key := strings.ToLower(parts[0])
		value := strings.Join(parts[1:], " ")

		switch key {
		case "host":
			// Save previous entry if exists
			if currentEntry != nil && currentEntry.Host != "" {
				entries = append(entries, *currentEntry)
			}
			// Start new entry
			currentEntry = &SSHConfigEntry{
				Host: value,
			}

		case "hostname":
			if currentEntry != nil {
				currentEntry.HostName = value
			}

		case "user":
			if currentEntry != nil {
				currentEntry.User = value
			}

		case "port":
			if currentEntry != nil {
				currentEntry.Port = value
			}

		case "identityfile":
			if currentEntry != nil {
				// Expand ~ in identity file path
				if strings.HasPrefix(value, "~/") {
					home, err := os.UserHomeDir()
					if err == nil {
						value = filepath.Join(home, value[2:])
					}
				}
				currentEntry.IdentityFile = value
			}
		}
	}

	// Add the last entry
	if currentEntry != nil && currentEntry.Host != "" {
		entries = append(entries, *currentEntry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading SSH config file: %w", err)
	}

	return entries, nil
}

// ConvertSSHConfigToHostConfig converts SSH config entries to HostConfig
func ConvertSSHConfigToHostConfig(entry SSHConfigEntry) HostConfig {
	host := HostConfig{
		Name: entry.Host,
		Host: entry.HostName,
		User: entry.User,
		Port: entry.Port,
	}

	// Use Host as hostname if HostName is not specified
	if host.Host == "" {
		host.Host = entry.Host
	}

	// Set default port if not specified
	if host.Port == "" {
		host.Port = "22"
	}

	// Set default user if not specified
	if host.User == "" {
		host.User = "master"
	}

	// Try to detect system type from hostname or use default
	host.Type = detectSystemType(host.Host)

	return host
}

// detectSystemType tries to detect the system type from hostname
func detectSystemType(hostname string) string {
	hostname = strings.ToLower(hostname)

	// Simple heuristic based on hostname
	// Check for mac/darwin first before windows to avoid "darwin" matching "win"
	if strings.Contains(hostname, "mac") || strings.Contains(hostname, "darwin") {
		return "macos"
	}
	if strings.Contains(hostname, "windows") || strings.Contains(hostname, "win") {
		return "windows"
	}

	// Default to linux
	return "linux"
}

// ImportFromSSHConfig imports hosts from SSH config file
func ImportFromSSHConfig(settings *Settings, overwrite bool) (int, error) {
	entries, err := ParseSSHConfig()
	if err != nil {
		return 0, err
	}

	imported := 0
	skipped := 0

	for _, entry := range entries {
		// Skip wildcard entries and special entries
		if strings.Contains(entry.Host, "*") || strings.Contains(entry.Host, "?") {
			continue
		}

		hostConfig := ConvertSSHConfigToHostConfig(entry)

		// Check if host already exists
		exists := false
		for _, h := range settings.Hosts {
			if h.Name == hostConfig.Name {
				exists = true
				break
			}
		}

		if exists && !overwrite {
			skipped++
			continue
		}

		if exists && overwrite {
			// Update existing host
			if err := UpdateHost(settings, hostConfig); err != nil {
				return imported, fmt.Errorf("failed to update host '%s': %w", hostConfig.Name, err)
			}
		} else {
			// Add new host
			if err := AddHost(settings, hostConfig); err != nil {
				return imported, fmt.Errorf("failed to add host '%s': %w", hostConfig.Name, err)
			}
		}

		imported++
	}

	return imported, nil
}
