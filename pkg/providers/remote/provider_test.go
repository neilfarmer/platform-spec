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
