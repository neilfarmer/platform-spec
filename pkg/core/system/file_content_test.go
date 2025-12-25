package system

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestExecutor_FileContentTest(t *testing.T) {
	tests := []struct {
		name            string
		fileContentTest core.FileContentTest
		setupMock       func(*core.MockProvider)
		wantStatus      core.Status
		wantContains    string
	}{
		{
			name: "file contains string",
			fileContentTest: core.FileContentTest{
				Name:     "Config contains setting",
				Path:     "/etc/app/config.yml",
				Contains: []string{"log-driver"},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("test -f /etc/app/config.yml && test -r /etc/app/config.yml", "", "", 0, nil)
				m.SetCommandResult("grep -F 'log-driver' /etc/app/config.yml >/dev/null 2>&1", "", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "contains",
		},
		{
			name: "file missing string",
			fileContentTest: core.FileContentTest{
				Name:     "Config contains setting",
				Path:     "/etc/app/config.yml",
				Contains: []string{"missing-option"},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("test -f /etc/app/config.yml && test -r /etc/app/config.yml", "", "", 0, nil)
				m.SetCommandResult("grep -F 'missing-option' /etc/app/config.yml >/dev/null 2>&1", "", "", 1, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "does not contain",
		},
		{
			name: "file does not exist",
			fileContentTest: core.FileContentTest{
				Name:     "Config contains setting",
				Path:     "/etc/app/missing.yml",
				Contains: []string{"something"},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("test -f /etc/app/missing.yml && test -r /etc/app/missing.yml", "", "", 1, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "does not exist",
		},
		{
			name: "file matches regex pattern",
			fileContentTest: core.FileContentTest{
				Name:    "SSH config secure",
				Path:    "/etc/ssh/sshd_config",
				Matches: "^PermitRootLogin no$",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("test -f /etc/ssh/sshd_config && test -r /etc/ssh/sshd_config", "", "", 0, nil)
				m.SetCommandResult("grep -E '^PermitRootLogin no$' /etc/ssh/sshd_config >/dev/null 2>&1", "", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "matches pattern",
		},
		{
			name: "file does not match pattern",
			fileContentTest: core.FileContentTest{
				Name:    "SSH config check",
				Path:    "/etc/ssh/sshd_config",
				Matches: "^PermitRootLogin yes$",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("test -f /etc/ssh/sshd_config && test -r /etc/ssh/sshd_config", "", "", 0, nil)
				m.SetCommandResult("grep -E '^PermitRootLogin yes$' /etc/ssh/sshd_config >/dev/null 2>&1", "", "", 1, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "does not match",
		},
		{
			name: "file contains multiple strings",
			fileContentTest: core.FileContentTest{
				Name:     "Config has required settings",
				Path:     "/etc/docker/daemon.json",
				Contains: []string{"log-driver", "metrics-addr", "storage-driver"},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("test -f /etc/docker/daemon.json && test -r /etc/docker/daemon.json", "", "", 0, nil)
				m.SetCommandResult("grep -F 'log-driver' /etc/docker/daemon.json >/dev/null 2>&1", "", "", 0, nil)
				m.SetCommandResult("grep -F 'metrics-addr' /etc/docker/daemon.json >/dev/null 2>&1", "", "", 0, nil)
				m.SetCommandResult("grep -F 'storage-driver' /etc/docker/daemon.json >/dev/null 2>&1", "", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "contains all 3",
		},
		{
			name: "file contains strings and matches pattern",
			fileContentTest: core.FileContentTest{
				Name:     "Config complete",
				Path:     "/etc/app/config.yml",
				Contains: []string{"setting1", "setting2"},
				Matches:  "^version:",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("test -f /etc/app/config.yml && test -r /etc/app/config.yml", "", "", 0, nil)
				m.SetCommandResult("grep -F 'setting1' /etc/app/config.yml >/dev/null 2>&1", "", "", 0, nil)
				m.SetCommandResult("grep -F 'setting2' /etc/app/config.yml >/dev/null 2>&1", "", "", 0, nil)
				m.SetCommandResult("grep -E '^version:' /etc/app/config.yml >/dev/null 2>&1", "", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "contains all 2 strings and matches",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := core.NewMockProvider()
			tt.setupMock(mock)

			ctx := context.Background()

			result := executeFileContentTest(ctx, mock, tt.fileContentTest)

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
