package system

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestExecutor_PackageTest(t *testing.T) {
	tests := []struct {
		name         string
		packageTest  core.PackageTest
		setupMock    func(*core.MockProvider)
		wantStatus   core.Status
		wantContains string
	}{
		{
			name: "package installed - dpkg",
			packageTest: core.PackageTest{
				Name:     "Docker installed",
				Packages: []string{"docker-ce"},
				State:    "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("dpkg -l docker-ce 2>/dev/null | grep '^ii'", "ii  docker-ce  5:24.0.7  amd64", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "1 packages are installed",
		},
		{
			name: "package not installed",
			packageTest: core.PackageTest{
				Name:     "Docker installed",
				Packages: []string{"docker-ce"},
				State:    "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("dpkg -l docker-ce 2>/dev/null | grep '^ii'", "", "", 1, nil)
				m.SetCommandResult("rpm -q docker-ce 2>/dev/null", "", "", 1, nil)
				m.SetCommandResult("apk info -e docker-ce 2>/dev/null", "", "", 1, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "not installed",
		},
		{
			name: "package absent check - success",
			packageTest: core.PackageTest{
				Name:     "Telnet removed",
				Packages: []string{"telnet"},
				State:    "absent",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("dpkg -l telnet 2>/dev/null | grep '^ii'", "", "", 1, nil)
				m.SetCommandResult("rpm -q telnet 2>/dev/null", "", "", 1, nil)
				m.SetCommandResult("apk info -e telnet 2>/dev/null", "", "", 1, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "absent as expected",
		},
		{
			name: "package absent check - failure",
			packageTest: core.PackageTest{
				Name:     "Telnet removed",
				Packages: []string{"telnet"},
				State:    "absent",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("dpkg -l telnet 2>/dev/null | grep '^ii'", "ii  telnet  0.17  amd64", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "should be absent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := core.NewMockProvider()
			tt.setupMock(mock)

			ctx := context.Background()

			result := executePackageTest(ctx, mock, tt.packageTest)

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
