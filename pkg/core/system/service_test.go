package system

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestExecutor_ServiceTest(t *testing.T) {
	tests := []struct {
		name         string
		serviceTest  core.ServiceTest
		setupMock    func(*core.MockProvider)
		wantStatus   core.Status
		wantContains string
	}{
		{
			name: "service running and enabled",
			serviceTest: core.ServiceTest{
				Name:    "Docker running",
				Service: "docker",
				State:   "running",
				Enabled: true,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("systemctl is-active docker 2>/dev/null", "active", "", 0, nil)
				m.SetCommandResult("systemctl is-enabled docker 2>/dev/null", "enabled", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "running",
		},
		{
			name: "service not running",
			serviceTest: core.ServiceTest{
				Name:    "Docker running",
				Service: "docker",
				State:   "running",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("systemctl is-active docker 2>/dev/null", "inactive", "", 3, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "not running",
		},
		{
			name: "service stopped check - success",
			serviceTest: core.ServiceTest{
				Name:    "Telnet stopped",
				Service: "telnet",
				State:   "stopped",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("systemctl is-active telnet 2>/dev/null", "inactive", "", 3, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "stopped",
		},
		{
			name: "service stopped check - failure",
			serviceTest: core.ServiceTest{
				Name:    "Telnet stopped",
				Service: "telnet",
				State:   "stopped",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("systemctl is-active telnet 2>/dev/null", "active", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "should be stopped",
		},
		{
			name: "service not enabled",
			serviceTest: core.ServiceTest{
				Name:    "Docker enabled",
				Service: "docker",
				State:   "running",
				Enabled: true,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("systemctl is-active docker 2>/dev/null", "active", "", 0, nil)
				m.SetCommandResult("systemctl is-enabled docker 2>/dev/null", "disabled", "", 1, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "not enabled",
		},
		{
			name: "multiple services",
			serviceTest: core.ServiceTest{
				Name:     "Critical services running",
				Services: []string{"docker", "nginx"},
				State:    "running",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("systemctl is-active docker 2>/dev/null", "active", "", 0, nil)
				m.SetCommandResult("systemctl is-enabled docker 2>/dev/null", "enabled", "", 0, nil)
				m.SetCommandResult("systemctl is-active nginx 2>/dev/null", "active", "", 0, nil)
				m.SetCommandResult("systemctl is-enabled nginx 2>/dev/null", "enabled", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "2 services",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := core.NewMockProvider()
			tt.setupMock(mock)

			ctx := context.Background()

			result := executeServiceTest(ctx, mock, tt.serviceTest)

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
