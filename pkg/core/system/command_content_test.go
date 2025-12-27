package system

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestExecutor_CommandContentTest(t *testing.T) {
	tests := []struct {
		name               string
		commandContentTest core.CommandContentTest
		setupMock          func(*core.MockProvider)
		wantStatus         core.Status
		wantContains       string
	}{
		{
			name: "command output contains string",
			commandContentTest: core.CommandContentTest{
				Name:     "Disk space check",
				Command:  "df -h",
				Contains: []string{"/var/lib/docker"},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("df -h", "Filesystem      Size  Used Avail Use% Mounted on\n/dev/sda1        50G   20G   30G  40% /\n/dev/sdb1       100G   50G   50G  50% /var/lib/docker", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "contains",
		},
		{
			name: "command output missing string",
			commandContentTest: core.CommandContentTest{
				Name:     "Check for nginx",
				Command:  "ps aux",
				Contains: []string{"nginx"},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("ps aux", "USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND\nroot         1  0.0  0.1  12345  6789 ?        Ss   10:00   0:01 /sbin/init", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "does not contain",
		},
		{
			name: "command exit code check",
			commandContentTest: core.CommandContentTest{
				Name:     "Service status",
				Command:  "systemctl is-active docker",
				ExitCode: 0,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("systemctl is-active docker", "active", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "executed successfully",
		},
		{
			name: "command wrong exit code",
			commandContentTest: core.CommandContentTest{
				Name:     "Service should fail",
				Command:  "systemctl is-active nonexistent",
				ExitCode: 3,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("systemctl is-active nonexistent", "inactive", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "exit code is 0, expected 3",
		},
		{
			name: "command multiple contains",
			commandContentTest: core.CommandContentTest{
				Name:     "Docker info check",
				Command:  "docker info",
				Contains: []string{"Server Version", "Storage Driver", "Logging Driver"},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("docker info", "Server Version: 24.0.7\nStorage Driver: overlay2\nLogging Driver: json-file\nCgroup Driver: systemd", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "contains all 3",
		},
		{
			name: "command with exit code and contains",
			commandContentTest: core.CommandContentTest{
				Name:     "Success with content",
				Command:  "echo hello",
				ExitCode: 0,
				Contains: []string{"hello"},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("echo hello", "hello", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "contains all 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := core.NewMockProvider()
			tt.setupMock(mock)

			ctx := context.Background()

			result := executeCommandContentTest(ctx, mock, tt.commandContentTest)

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
