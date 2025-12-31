package inventory

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Inventory represents a parsed inventory file containing a list of hosts
type Inventory struct {
	Hosts []string // List of hostnames or IP addresses
}

// ParseInventoryFile reads and parses an inventory file
// Returns an error if the file doesn't exist, is empty, or contains malformed entries
func ParseInventoryFile(path string) (*Inventory, error) {
	// #nosec G304 -- Reading user-specified inventory file is intentional and required functionality
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open inventory file: %w", err)
	}
	defer file.Close()

	var hosts []string
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip comments
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Validate the host entry
		if err := validateHost(line); err != nil {
			return nil, fmt.Errorf("invalid entry at line %d: %w", lineNum, err)
		}

		hosts = append(hosts, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading inventory file: %w", err)
	}

	if len(hosts) == 0 {
		return nil, fmt.Errorf("inventory file is empty (no valid hosts found)")
	}

	return &Inventory{Hosts: hosts}, nil
}

// validateHost checks if a host entry is valid
// A valid host is a non-empty string without whitespace
func validateHost(host string) error {
	if host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	if strings.ContainsAny(host, " \t\n\r") {
		return fmt.Errorf("host '%s' contains whitespace", host)
	}

	return nil
}
