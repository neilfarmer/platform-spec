package system

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestExecutor_DNSTest(t *testing.T) {
	tests := []struct {
		name         string
		dnsTest      core.DNSTest
		setupMock    func(*core.MockProvider)
		wantStatus   core.Status
		wantContains string
	}{
		{
			name: "successful DNS resolution with single IP",
			dnsTest: core.DNSTest{
				Name: "Resolve google.com",
				Host: "google.com",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("dig +short 'google.com' 2>/dev/null || getent hosts 'google.com' 2>/dev/null | awk '{print $1}'", "142.250.80.46", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "resolved",
		},
		{
			name: "successful DNS resolution with multiple IPs",
			dnsTest: core.DNSTest{
				Name: "Resolve example.com",
				Host: "example.com",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("dig +short 'example.com' 2>/dev/null || getent hosts 'example.com' 2>/dev/null | awk '{print $1}'", "93.184.216.34\n2606:2800:220:1:248:1893:25c8:1946", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "2 address(es)",
		},
		{
			name: "DNS resolution fails - host not found",
			dnsTest: core.DNSTest{
				Name: "Resolve nonexistent.invalid",
				Host: "nonexistent.invalid",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("dig +short 'nonexistent.invalid' 2>/dev/null || getent hosts 'nonexistent.invalid' 2>/dev/null | awk '{print $1}'", "", "", 1, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "resolution failed",
		},
		{
			name: "DNS resolution returns empty output",
			dnsTest: core.DNSTest{
				Name: "Empty resolution",
				Host: "empty.test",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("dig +short 'empty.test' 2>/dev/null || getent hosts 'empty.test' 2>/dev/null | awk '{print $1}'", "", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "resolution failed",
		},
		{
			name: "DNS resolution with localhost",
			dnsTest: core.DNSTest{
				Name: "Resolve localhost",
				Host: "localhost",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("dig +short 'localhost' 2>/dev/null || getent hosts 'localhost' 2>/dev/null | awk '{print $1}'", "127.0.0.1\n::1", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "2 address(es)",
		},
		{
			name: "DNS resolution with whitespace in output",
			dnsTest: core.DNSTest{
				Name: "Resolve with whitespace",
				Host: "test.com",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("dig +short 'test.com' 2>/dev/null || getent hosts 'test.com' 2>/dev/null | awk '{print $1}'", "  192.168.1.1  \n\n  10.0.0.1  \n", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "2 address(es)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := core.NewMockProvider()
			tt.setupMock(mock)

			ctx := context.Background()

			result := executeDNSTest(ctx, mock, tt.dnsTest)

			if result.Status != tt.wantStatus {
				t.Errorf("Status = %v, want %v", result.Status, tt.wantStatus)
			}

			if tt.wantContains != "" && result.Message != "" {
				if !contains(result.Message, tt.wantContains) {
					t.Errorf("Message %q does not contain %q", result.Message, tt.wantContains)
				}
			}
		})
	}
}
