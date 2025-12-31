package inventory

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseInventoryFile_Valid(t *testing.T) {
	content := `# Web servers
web-server-01.example.com
web-server-02.example.com

# Database servers
db-server-01.example.com

# IP addresses
192.168.1.10
192.168.1.11
`

	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)

	inv, err := ParseInventoryFile(tmpfile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedHosts := []string{
		"web-server-01.example.com",
		"web-server-02.example.com",
		"db-server-01.example.com",
		"192.168.1.10",
		"192.168.1.11",
	}

	if len(inv.Hosts) != len(expectedHosts) {
		t.Fatalf("expected %d hosts, got %d", len(expectedHosts), len(inv.Hosts))
	}

	for i, expected := range expectedHosts {
		if inv.Hosts[i] != expected {
			t.Errorf("host %d: expected %s, got %s", i, expected, inv.Hosts[i])
		}
	}
}

func TestParseInventoryFile_CommentsOnly(t *testing.T) {
	content := `# Only comments
# No actual hosts
`

	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)

	_, err := ParseInventoryFile(tmpfile)
	if err == nil {
		t.Fatal("expected error for inventory with only comments, got nil")
	}

	if err.Error() != "inventory file is empty (no valid hosts found)" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestParseInventoryFile_EmptyFile(t *testing.T) {
	content := ``

	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)

	_, err := ParseInventoryFile(tmpfile)
	if err == nil {
		t.Fatal("expected error for empty inventory, got nil")
	}

	if err.Error() != "inventory file is empty (no valid hosts found)" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestParseInventoryFile_EmptyLinesOnly(t *testing.T) {
	content := `


`

	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)

	_, err := ParseInventoryFile(tmpfile)
	if err == nil {
		t.Fatal("expected error for inventory with only empty lines, got nil")
	}
}

func TestParseInventoryFile_FileNotFound(t *testing.T) {
	_, err := ParseInventoryFile("/nonexistent/inventory.txt")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestParseInventoryFile_Whitespace(t *testing.T) {
	content := `  web-server-01.example.com
	db-server-01.example.com
192.168.1.10
`

	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)

	inv, err := ParseInventoryFile(tmpfile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedHosts := []string{
		"web-server-01.example.com",
		"db-server-01.example.com",
		"192.168.1.10",
	}

	if len(inv.Hosts) != len(expectedHosts) {
		t.Fatalf("expected %d hosts, got %d", len(expectedHosts), len(inv.Hosts))
	}

	for i, expected := range expectedHosts {
		if inv.Hosts[i] != expected {
			t.Errorf("host %d: expected %s, got %s", i, expected, inv.Hosts[i])
		}
	}
}

func TestParseInventoryFile_MixedContent(t *testing.T) {
	content := `# Header comment
web-server-01.example.com

# Another comment
web-server-02.example.com

192.168.1.10
# Inline comment after hosts
`

	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)

	inv, err := ParseInventoryFile(tmpfile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(inv.Hosts) != 3 {
		t.Fatalf("expected 3 hosts, got %d", len(inv.Hosts))
	}
}

func TestParseInventoryFile_HostnameFormats(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "simple hostnames",
			content:  "host1\nhost2\nhost3",
			expected: []string{"host1", "host2", "host3"},
		},
		{
			name:     "FQDNs",
			content:  "web-01.example.com\ndb-01.example.org",
			expected: []string{"web-01.example.com", "db-01.example.org"},
		},
		{
			name:     "IP addresses",
			content:  "192.168.1.1\n10.0.0.5\n172.16.0.10",
			expected: []string{"192.168.1.1", "10.0.0.5", "172.16.0.10"},
		},
		{
			name:     "mixed formats",
			content:  "host1\nweb.example.com\n192.168.1.1",
			expected: []string{"host1", "web.example.com", "192.168.1.1"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpfile := createTempFile(t, tc.content)
			defer os.Remove(tmpfile)

			inv, err := ParseInventoryFile(tmpfile)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(inv.Hosts) != len(tc.expected) {
				t.Fatalf("expected %d hosts, got %d", len(tc.expected), len(inv.Hosts))
			}

			for i, expected := range tc.expected {
				if inv.Hosts[i] != expected {
					t.Errorf("host %d: expected %s, got %s", i, expected, inv.Hosts[i])
				}
			}
		})
	}
}

func TestValidateHost(t *testing.T) {
	testCases := []struct {
		name      string
		host      string
		expectErr bool
	}{
		{"valid hostname", "web-server-01.example.com", false},
		{"valid IP", "192.168.1.10", false},
		{"valid simple hostname", "localhost", false},
		{"empty string", "", true},
		{"whitespace in middle", "web server 01", true},
		{"leading space", " web-server-01", true},
		{"trailing space", "web-server-01 ", true},
		{"tab character", "web-server\t01", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateHost(tc.host)
			if tc.expectErr && err == nil {
				t.Errorf("expected error for host '%s', got nil", tc.host)
			}
			if !tc.expectErr && err != nil {
				t.Errorf("unexpected error for host '%s': %v", tc.host, err)
			}
		})
	}
}

func TestParseInventoryFile_MalformedEntry(t *testing.T) {
	content := `web-server-01.example.com
web server 02
db-server-01.example.com
`

	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)

	_, err := ParseInventoryFile(tmpfile)
	if err == nil {
		t.Fatal("expected error for malformed entry, got nil")
	}

	// Error should mention line number
	expectedSubstring := "line 2"
	if !contains(err.Error(), expectedSubstring) {
		t.Errorf("error message should contain '%s', got: %s", expectedSubstring, err.Error())
	}
}

// Helper function to create a temporary file with content
func createTempFile(t *testing.T, content string) string {
	t.Helper()

	tmpdir := t.TempDir()
	tmpfile := filepath.Join(tmpdir, "inventory.txt")

	err := os.WriteFile(tmpfile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	return tmpfile
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		len(s) > 0 && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
