package remote

import "testing"

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
		name           string
		config         *Config
		wantJumpHost   string
		wantJumpPort   int
		wantJumpUser   string
		wantTargetHost string
		wantTargetPort int
		wantTargetUser string
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
			wantJumpHost:   "jump-host",
			wantJumpPort:   22,
			wantJumpUser:   "jumpuser",
			wantTargetHost: "target-host",
			wantTargetPort: 22,
			wantTargetUser: "targetuser",
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
			wantJumpHost:   "jump-host",
			wantJumpPort:   2223,
			wantJumpUser:   "jumpuser",
			wantTargetHost: "target-host",
			wantTargetPort: 2222,
			wantTargetUser: "targetuser",
		},
		{
			name: "without jump host",
			config: &Config{
				Host: "direct-host",
				Port: 22,
				User: "directuser",
			},
			wantJumpHost:   "",
			wantJumpPort:   0,
			wantJumpUser:   "",
			wantTargetHost: "direct-host",
			wantTargetPort: 22,
			wantTargetUser: "directuser",
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
			if provider.config.Host != tt.wantTargetHost {
				t.Errorf("config.Host = %v, want %v", provider.config.Host, tt.wantTargetHost)
			}
			if provider.config.Port != tt.wantTargetPort {
				t.Errorf("config.Port = %v, want %v", provider.config.Port, tt.wantTargetPort)
			}
			if provider.config.User != tt.wantTargetUser {
				t.Errorf("config.User = %v, want %v", provider.config.User, tt.wantTargetUser)
			}
		})
	}
}
