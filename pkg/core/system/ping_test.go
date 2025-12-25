package system

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestExecutor_PingTest(t *testing.T) {
	tests := []struct {
		name         string
		pingTest     core.PingTest
		setupMock    func(*core.MockProvider)
		wantStatus   core.Status
		wantContains string
	}{
		{
			name: "host is reachable",
			pingTest: core.PingTest{
				Name: "Google DNS reachable",
				Host: "8.8.8.8",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("ping -c 1 -W 5 '8.8.8.8' 2>&1", "PING 8.8.8.8 (8.8.8.8) 56(84) bytes of data.\n64 bytes from 8.8.8.8: icmp_seq=1 ttl=117 time=10.2 ms", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "is reachable",
		},
		{
			name: "host is not reachable",
			pingTest: core.PingTest{
				Name: "Unreachable host",
				Host: "192.0.2.1",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("ping -c 1 -W 5 '192.0.2.1' 2>&1", "", "connect: Network is unreachable", 1, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "not reachable",
		},
		{
			name: "hostname reachable",
			pingTest: core.PingTest{
				Name: "Google reachable",
				Host: "google.com",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("ping -c 1 -W 5 'google.com' 2>&1", "PING google.com (142.250.185.46) 56(84) bytes of data.\n64 bytes from google.com: icmp_seq=1 ttl=117 time=10.5 ms", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "is reachable",
		},
		{
			name: "localhost reachable",
			pingTest: core.PingTest{
				Name: "Localhost check",
				Host: "127.0.0.1",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("ping -c 1 -W 5 '127.0.0.1' 2>&1", "PING 127.0.0.1 (127.0.0.1) 56(84) bytes of data.\n64 bytes from 127.0.0.1: icmp_seq=1 ttl=64 time=0.025 ms", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "is reachable",
		},
		{
			name: "host timeout",
			pingTest: core.PingTest{
				Name: "Timeout check",
				Host: "10.255.255.1",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("ping -c 1 -W 5 '10.255.255.1' 2>&1", "", "", 1, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "not reachable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := core.NewMockProvider()
			tt.setupMock(mock)

			ctx := context.Background()

			result := executePingTest(ctx, mock, tt.pingTest)

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
