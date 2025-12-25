package system

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestExecutor_FileTest(t *testing.T) {
	tests := []struct {
		name         string
		fileTest     core.FileTest
		setupMock    func(*core.MockProvider)
		wantStatus   core.Status
		wantContains string
	}{
		{
			name: "directory exists with correct permissions",
			fileTest: core.FileTest{
				Name:  "App directory",
				Path:  "/opt/app",
				Type:  "directory",
				Owner: "appuser",
				Group: "appuser",
				Mode:  "0755",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("stat -c '%F:%U:%G:%a' /opt/app 2>/dev/null || echo 'notfound'", "directory:appuser:appuser:755", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "correct properties",
		},
		{
			name: "file does not exist",
			fileTest: core.FileTest{
				Name: "Config file",
				Path: "/etc/myapp/config.yml",
				Type: "file",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("stat -c '%F:%U:%G:%a' /etc/myapp/config.yml 2>/dev/null || echo 'notfound'", "notfound", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "does not exist",
		},
		{
			name: "wrong file type",
			fileTest: core.FileTest{
				Name: "Should be directory",
				Path: "/opt/app",
				Type: "directory",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("stat -c '%F:%U:%G:%a' /opt/app 2>/dev/null || echo 'notfound'", "regular file:root:root:644", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "expected directory",
		},
		{
			name: "wrong owner",
			fileTest: core.FileTest{
				Name:  "Config file",
				Path:  "/etc/app/config",
				Type:  "file",
				Owner: "appuser",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("stat -c '%F:%U:%G:%a' /etc/app/config 2>/dev/null || echo 'notfound'", "regular file:root:root:644", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "owner is root, expected appuser",
		},
		{
			name: "wrong permissions",
			fileTest: core.FileTest{
				Name: "Private directory",
				Path: "/opt/secrets",
				Type: "directory",
				Mode: "0700",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("stat -c '%F:%U:%G:%a' /opt/secrets 2>/dev/null || echo 'notfound'", "directory:user:user:755", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "mode is 755, expected 700",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := core.NewMockProvider()
			tt.setupMock(mock)

			ctx := context.Background()

			result := executeFileTest(ctx, mock, tt.fileTest)

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
