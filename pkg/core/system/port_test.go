package system

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestExecutor_PortTest(t *testing.T) {
	tests := []struct {
		name         string
		portTest     core.PortTest
		setupMock    func(*core.MockProvider)
		wantStatus   core.Status
		wantContains string
	}{
		{
			name: "TCP port listening",
			portTest: core.PortTest{
				Name:     "SSH listening",
				Port:     22,
				Protocol: "tcp",
				State:    "listening",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("ss -tln | grep -E ':22\\s' || true", "LISTEN    0    128    0.0.0.0:22    0.0.0.0:*", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "Port 22/tcp is listening",
		},
		{
			name: "UDP port listening",
			portTest: core.PortTest{
				Name:     "DNS listening",
				Port:     53,
				Protocol: "udp",
				State:    "listening",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("ss -uln | grep -E ':53\\s' || true", "UNCONN    0    0    0.0.0.0:53    0.0.0.0:*", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "Port 53/udp is listening",
		},
		{
			name: "TCP port not listening (fail)",
			portTest: core.PortTest{
				Name:     "HTTP not listening",
				Port:     80,
				Protocol: "tcp",
				State:    "listening",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("ss -tln | grep -E ':80\\s' || true", "", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "Port 80/tcp is not listening",
		},
		{
			name: "TCP port closed (pass)",
			portTest: core.PortTest{
				Name:     "Port should be closed",
				Port:     9999,
				Protocol: "tcp",
				State:    "closed",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("ss -tln | grep -E ':9999\\s' || true", "", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "Port 9999/tcp is closed",
		},
		{
			name: "TCP port closed but listening (fail)",
			portTest: core.PortTest{
				Name:     "Port should be closed but is listening",
				Port:     8080,
				Protocol: "tcp",
				State:    "closed",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("ss -tln | grep -E ':8080\\s' || true", "LISTEN    0    128    0.0.0.0:8080    0.0.0.0:*", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "Port 8080/tcp is listening, expected closed",
		},
		{
			name: "port check with defaults",
			portTest: core.PortTest{
				Name:     "SSH default",
				Port:     22,
				Protocol: "tcp",
				State:    "listening",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("ss -tln | grep -E ':22\\s' || true", "LISTEN    0    128    0.0.0.0:22    0.0.0.0:*", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "Port 22/tcp is listening",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := core.NewMockProvider()
			tt.setupMock(mock)

			ctx := context.Background()

			result := executePortTest(ctx, mock, tt.portTest)

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
