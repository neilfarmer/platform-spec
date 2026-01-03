package remote

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestParseTarget(t *testing.T) {
	tests := []struct {
		name        string
		target      string
		defaultUser string
		wantUser    string
		wantHost    string
		wantErr     bool
	}{
		{
			name:     "user@host format",
			target:   "ubuntu@192.168.1.100",
			wantUser: "ubuntu",
			wantHost: "192.168.1.100",
			wantErr:  false,
		},
		{
			name:        "host only with default user",
			target:      "192.168.1.100",
			defaultUser: "admin",
			wantUser:    "admin",
			wantHost:    "192.168.1.100",
			wantErr:     false,
		},
		{
			name:     "host only defaults to root",
			target:   "server.example.com",
			wantUser: "root",
			wantHost: "server.example.com",
			wantErr:  false,
		},
		{
			name:     "hostname with domain",
			target:   "user@server.example.com",
			wantUser: "user",
			wantHost: "server.example.com",
			wantErr:  false,
		},
		{
			name:    "invalid format",
			target:  "user@host@extra",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, host, err := ParseTarget(tt.target, tt.defaultUser)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTarget() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if user != tt.wantUser {
					t.Errorf("ParseTarget() user = %v, want %v", user, tt.wantUser)
				}
				if host != tt.wantHost {
					t.Errorf("ParseTarget() host = %v, want %v", host, tt.wantHost)
				}
			}
		})
	}
}

func TestNewProvider(t *testing.T) {
	config := &Config{
		Host: "localhost",
		Port: 22,
		User: "testuser",
	}

	provider := NewProvider(config)

	if provider == nil {
		t.Fatal("NewProvider() returned nil")
	}

	if provider.config != config {
		t.Error("NewProvider() did not set config correctly")
	}
}

func TestNewProviderWithJumpHost(t *testing.T) {
	tests := []struct {
		name               string
		config             *Config
		wantJumpHost       string
		wantJumpPort       int
		wantJumpUser       string
		wantJumpIdentity   string
		wantTargetHost     string
		wantTargetPort     int
		wantTargetUser     string
		wantTargetIdentity string
	}{
		{
			name: "with jump host configuration",
			config: &Config{
				Host:     "target-host",
				Port:     22,
				User:     "targetuser",
				JumpHost: "jump-host",
				JumpPort: 22,
				JumpUser: "jumpuser",
			},
			wantJumpHost:       "jump-host",
			wantJumpPort:       22,
			wantJumpUser:       "jumpuser",
			wantJumpIdentity:   "",
			wantTargetHost:     "target-host",
			wantTargetPort:     22,
			wantTargetUser:     "targetuser",
			wantTargetIdentity: "",
		},
		{
			name: "with jump host and custom port",
			config: &Config{
				Host:     "target-host",
				Port:     2222,
				User:     "targetuser",
				JumpHost: "jump-host",
				JumpPort: 2223,
				JumpUser: "jumpuser",
			},
			wantJumpHost:       "jump-host",
			wantJumpPort:       2223,
			wantJumpUser:       "jumpuser",
			wantJumpIdentity:   "",
			wantTargetHost:     "target-host",
			wantTargetPort:     2222,
			wantTargetUser:     "targetuser",
			wantTargetIdentity: "",
		},
		{
			name: "with separate jump and target identity files",
			config: &Config{
				Host:             "target-host",
				Port:             22,
				User:             "targetuser",
				IdentityFile:     "/path/to/target/key",
				JumpHost:         "jump-host",
				JumpPort:         22,
				JumpUser:         "jumpuser",
				JumpIdentityFile: "/path/to/jump/key",
			},
			wantJumpHost:       "jump-host",
			wantJumpPort:       22,
			wantJumpUser:       "jumpuser",
			wantJumpIdentity:   "/path/to/jump/key",
			wantTargetHost:     "target-host",
			wantTargetPort:     22,
			wantTargetUser:     "targetuser",
			wantTargetIdentity: "/path/to/target/key",
		},
		{
			name: "without jump host",
			config: &Config{
				Host: "direct-host",
				Port: 22,
				User: "directuser",
			},
			wantJumpHost:       "",
			wantJumpPort:       0,
			wantJumpUser:       "",
			wantJumpIdentity:   "",
			wantTargetHost:     "direct-host",
			wantTargetPort:     22,
			wantTargetUser:     "directuser",
			wantTargetIdentity: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewProvider(tt.config)

			if provider == nil {
				t.Fatal("NewProvider() returned nil")
			}

			if provider.config.JumpHost != tt.wantJumpHost {
				t.Errorf("config.JumpHost = %v, want %v", provider.config.JumpHost, tt.wantJumpHost)
			}
			if provider.config.JumpPort != tt.wantJumpPort {
				t.Errorf("config.JumpPort = %v, want %v", provider.config.JumpPort, tt.wantJumpPort)
			}
			if provider.config.JumpUser != tt.wantJumpUser {
				t.Errorf("config.JumpUser = %v, want %v", provider.config.JumpUser, tt.wantJumpUser)
			}
			if provider.config.JumpIdentityFile != tt.wantJumpIdentity {
				t.Errorf("config.JumpIdentityFile = %v, want %v", provider.config.JumpIdentityFile, tt.wantJumpIdentity)
			}
			if provider.config.Host != tt.wantTargetHost {
				t.Errorf("config.Host = %v, want %v", provider.config.Host, tt.wantTargetHost)
			}
			if provider.config.Port != tt.wantTargetPort {
				t.Errorf("config.Port = %v, want %v", provider.config.Port, tt.wantTargetPort)
			}
			if provider.config.User != tt.wantTargetUser {
				t.Errorf("config.User = %v, want %v", provider.config.User, tt.wantTargetUser)
			}
			if provider.config.IdentityFile != tt.wantTargetIdentity {
				t.Errorf("config.IdentityFile = %v, want %v", provider.config.IdentityFile, tt.wantTargetIdentity)
			}
		})
	}
}

func TestBuildAuthMethods(t *testing.T) {
	// Create a temporary SSH key for testing
	tmpDir := t.TempDir()
	keyPath := tmpDir + "/test_key"

	// Generate a test key
	if err := generateTestSSHKey(keyPath); err != nil {
		t.Fatalf("Failed to generate test SSH key: %v", err)
	}

	tests := []struct {
		name         string
		identityFile string
		hostType     string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "valid key file",
			identityFile: keyPath,
			hostType:     "test host",
			wantErr:      false,
		},
		{
			name:         "missing key file",
			identityFile: tmpDir + "/nonexistent",
			hostType:     "test host",
			wantErr:      true,
			errContains:  "failed to read private key for test host",
		},
		{
			name:         "empty identity file uses SSH agent",
			identityFile: "",
			hostType:     "test host",
			wantErr:      false, // Will use SSH agent if available, or fail
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewProvider(&Config{})
			authMethods, err := provider.buildAuthMethods(tt.identityFile, tt.hostType)

			if tt.wantErr {
				if err == nil {
					t.Errorf("buildAuthMethods() expected error but got nil")
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("buildAuthMethods() error = %v, want to contain %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					// If no key file and no SSH agent, this is expected to fail
					if tt.identityFile == "" {
						t.Logf("buildAuthMethods() with empty identity file failed (expected if no SSH agent): %v", err)
						return
					}
					t.Errorf("buildAuthMethods() unexpected error = %v", err)
					return
				}
				if len(authMethods) == 0 {
					t.Errorf("buildAuthMethods() returned empty auth methods")
				}
			}
		})
	}
}

// generateTestSSHKey generates a test RSA private key
func generateTestSSHKey(path string) error {
	// Use ssh-keygen to generate a test key
	cmd := fmt.Sprintf("ssh-keygen -t rsa -b 2048 -f %s -N '' -q", path)
	if err := exec.Command("sh", "-c", cmd).Run(); err != nil {
		return fmt.Errorf("failed to generate SSH key: %w", err)
	}
	return nil
}

// containsString checks if a string contains a substring
func containsString(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestResolveHostFromSSHConfig(t *testing.T) {
	// Create a temporary SSH config file for testing
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config"

	// Write a test SSH config
	configContent := `# Test SSH config
Host web-server
    HostName 192.168.1.100
    User admin

Host db-server
    HostName db.internal.example.com
    Port 2222

Host no-hostname
    User root
    Port 22

Host wildcard-*
    HostName 10.0.0.1
`

	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to create test SSH config: %v", err)
	}

	// Set the SSH config file path for testing
	// ssh_config library reads from ~/.ssh/config by default
	// We'll need to temporarily override the home directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Create ~/.ssh directory in temp location
	sshDir := tmpDir + "/.ssh"
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	// Move config to ~/.ssh/config in temp location
	newConfigPath := sshDir + "/config"
	if err := os.Rename(configPath, newConfigPath); err != nil {
		t.Fatalf("Failed to move config: %v", err)
	}

	// Set HOME to temp directory
	os.Setenv("HOME", tmpDir)

	tests := []struct {
		name         string
		host         string
		wantHostname string
	}{
		{
			name:         "resolve hostname from config",
			host:         "web-server",
			wantHostname: "192.168.1.100",
		},
		{
			name:         "resolve FQDN from config",
			host:         "db-server",
			wantHostname: "db.internal.example.com",
		},
		{
			name:         "host without HostName directive returns original",
			host:         "no-hostname",
			wantHostname: "no-hostname",
		},
		{
			name:         "host not in config returns original",
			host:         "unknown-host",
			wantHostname: "unknown-host",
		},
		{
			name:         "wildcard pattern match",
			host:         "wildcard-prod",
			wantHostname: "10.0.0.1",
		},
		{
			name:         "IP address returns as-is",
			host:         "192.168.1.50",
			wantHostname: "192.168.1.50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved := resolveHostFromSSHConfig(tt.host)
			if resolved != tt.wantHostname {
				t.Errorf("resolveHostFromSSHConfig(%q) = %q, want %q", tt.host, resolved, tt.wantHostname)
			}
		})
	}
}

func TestGetHostnameFromConfig(t *testing.T) {
	// Create a temporary SSH config file
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config"

	configContent := `Host testhost
    HostName 192.168.1.1
    User admin
`

	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	tests := []struct {
		name         string
		configPath   string
		host         string
		wantHostname string
	}{
		{
			name:         "valid config with matching host",
			configPath:   configPath,
			host:         "testhost",
			wantHostname: "192.168.1.1",
		},
		{
			name:         "valid config with non-matching host",
			configPath:   configPath,
			host:         "otherhost",
			wantHostname: "",
		},
		{
			name:         "non-existent config file",
			configPath:   tmpDir + "/nonexistent",
			host:         "testhost",
			wantHostname: "",
		},
		{
			name:         "invalid config file",
			configPath:   tmpDir,
			host:         "testhost",
			wantHostname: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hostname := getHostnameFromConfig(tt.configPath, tt.host)
			if hostname != tt.wantHostname {
				t.Errorf("getHostnameFromConfig(%q, %q) = %q, want %q", tt.configPath, tt.host, hostname, tt.wantHostname)
			}
		})
	}
}

func TestNewProviderWithVerbose(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		wantVerbose bool
	}{
		{
			name: "verbose enabled",
			config: &Config{
				Host:    "localhost",
				Port:    22,
				User:    "testuser",
				Verbose: true,
			},
			wantVerbose: true,
		},
		{
			name: "verbose disabled",
			config: &Config{
				Host:    "localhost",
				Port:    22,
				User:    "testuser",
				Verbose: false,
			},
			wantVerbose: false,
		},
		{
			name: "verbose not set defaults to false",
			config: &Config{
				Host: "localhost",
				Port: 22,
				User: "testuser",
			},
			wantVerbose: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewProvider(tt.config)

			if provider == nil {
				t.Fatal("NewProvider() returned nil")
			}

			if provider.config.Verbose != tt.wantVerbose {
				t.Errorf("config.Verbose = %v, want %v", provider.config.Verbose, tt.wantVerbose)
			}
		})
	}
}

func TestFormatElapsed(t *testing.T) {
	tests := []struct {
		name     string
		duration string
		want     string
	}{
		{
			name:     "zero duration",
			duration: "0s",
			want:     "00:00.00",
		},
		{
			name:     "less than one second",
			duration: "500ms",
			want:     "00:00.50",
		},
		{
			name:     "exactly one second",
			duration: "1s",
			want:     "00:01.00",
		},
		{
			name:     "multiple seconds",
			duration: "45s",
			want:     "00:45.00",
		},
		{
			name:     "one minute",
			duration: "1m",
			want:     "01:00.00",
		},
		{
			name:     "one minute and seconds",
			duration: "1m30s",
			want:     "01:30.00",
		},
		{
			name:     "multiple minutes and seconds with milliseconds",
			duration: "2m45s500ms",
			want:     "02:45.50",
		},
		{
			name:     "large duration",
			duration: "10m15s",
			want:     "10:15.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration, err := parseTestDuration(tt.duration)
			if err != nil {
				t.Fatalf("Failed to parse duration %q: %v", tt.duration, err)
			}

			got := formatElapsed(duration)
			if got != tt.want {
				t.Errorf("formatElapsed(%v) = %q, want %q", duration, got, tt.want)
			}
		})
	}
}

// parseTestDuration parses a simple duration string for testing
func parseTestDuration(s string) (time.Duration, error) {
	var total time.Duration
	var num int
	var unit string

	// Parse simple format like "1m30s" or "500ms"
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			num = num*10 + int(s[i]-'0')
		} else {
			// Found a unit character
			j := i
			for j < len(s) && (s[j] < '0' || s[j] > '9') {
				j++
			}
			unit = s[i:j]

			var d time.Duration
			switch unit {
			case "m":
				d = time.Duration(num) * time.Minute
			case "s":
				d = time.Duration(num) * time.Second
			case "ms":
				d = time.Duration(num) * time.Millisecond
			default:
				return 0, fmt.Errorf("unknown unit: %s", unit)
			}

			total += d
			num = 0
			unit = ""
			i = j - 1
		}
	}

	return total, nil
}
