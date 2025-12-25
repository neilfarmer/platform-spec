package system

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestExecutor_GroupTest(t *testing.T) {
	tests := []struct {
		name         string
		groupTest    core.GroupTest
		setupMock    func(*core.MockProvider)
		wantStatus   core.Status
		wantContains string
	}{
		{
			name: "group exists",
			groupTest: core.GroupTest{
				Name:   "Docker group exists",
				Groups: []string{"docker"},
				State:  "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("getent group docker 2>/dev/null", "docker:x:999:ubuntu", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "groups exist",
		},
		{
			name: "group does not exist",
			groupTest: core.GroupTest{
				Name:   "Test group exists",
				Groups: []string{"testgroup"},
				State:  "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("getent group testgroup 2>/dev/null", "", "", 2, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "does not exist",
		},
		{
			name: "group absent check - success",
			groupTest: core.GroupTest{
				Name:   "Unwanted group absent",
				Groups: []string{"badgroup"},
				State:  "absent",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("getent group badgroup 2>/dev/null", "", "", 2, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "absent as expected",
		},
		{
			name: "group absent check - failure",
			groupTest: core.GroupTest{
				Name:   "Unwanted group absent",
				Groups: []string{"docker"},
				State:  "absent",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("getent group docker 2>/dev/null", "docker:x:999:ubuntu", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "should be absent",
		},
		{
			name: "multiple groups",
			groupTest: core.GroupTest{
				Name:   "Required groups exist",
				Groups: []string{"docker", "sudo"},
				State:  "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("getent group docker 2>/dev/null", "docker:x:999:ubuntu", "", 0, nil)
				m.SetCommandResult("getent group sudo 2>/dev/null", "sudo:x:27:ubuntu", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "2 groups",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := core.NewMockProvider()
			tt.setupMock(mock)

			ctx := context.Background()

			result := executeGroupTest(ctx, mock, tt.groupTest)

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
