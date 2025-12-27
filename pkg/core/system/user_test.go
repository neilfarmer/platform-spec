package system

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestExecutor_UserTest(t *testing.T) {
	tests := []struct {
		name         string
		userTest     core.UserTest
		setupMock    func(*core.MockProvider)
		wantStatus   core.Status
		wantContains string
	}{
		{
			name: "user exists",
			userTest: core.UserTest{
				Name: "Ubuntu user exists",
				User: "ubuntu",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("id -u ubuntu 2>/dev/null && id -g ubuntu 2>/dev/null && getent passwd ubuntu 2>/dev/null", "1000\n1000\nubuntu:x:1000:1000:Ubuntu:/home/ubuntu:/bin/bash", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "exists",
		},
		{
			name: "user does not exist",
			userTest: core.UserTest{
				Name: "Test user exists",
				User: "testuser",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("id -u testuser 2>/dev/null && id -g testuser 2>/dev/null && getent passwd testuser 2>/dev/null", "", "", 1, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "does not exist",
		},
		{
			name: "user with correct shell",
			userTest: core.UserTest{
				Name:  "Ubuntu shell",
				User:  "ubuntu",
				Shell: "/bin/bash",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("id -u ubuntu 2>/dev/null && id -g ubuntu 2>/dev/null && getent passwd ubuntu 2>/dev/null", "1000\n1000\nubuntu:x:1000:1000:Ubuntu:/home/ubuntu:/bin/bash", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "exists",
		},
		{
			name: "user with wrong shell",
			userTest: core.UserTest{
				Name:  "Ubuntu shell",
				User:  "ubuntu",
				Shell: "/bin/zsh",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("id -u ubuntu 2>/dev/null && id -g ubuntu 2>/dev/null && getent passwd ubuntu 2>/dev/null", "1000\n1000\nubuntu:x:1000:1000:Ubuntu:/home/ubuntu:/bin/bash", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "shell",
		},
		{
			name: "user with correct home",
			userTest: core.UserTest{
				Name: "Ubuntu home",
				User: "ubuntu",
				Home: "/home/ubuntu",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("id -u ubuntu 2>/dev/null && id -g ubuntu 2>/dev/null && getent passwd ubuntu 2>/dev/null", "1000\n1000\nubuntu:x:1000:1000:Ubuntu:/home/ubuntu:/bin/bash", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "exists",
		},
		{
			name: "user with wrong home",
			userTest: core.UserTest{
				Name: "Ubuntu home",
				User: "ubuntu",
				Home: "/root",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("id -u ubuntu 2>/dev/null && id -g ubuntu 2>/dev/null && getent passwd ubuntu 2>/dev/null", "1000\n1000\nubuntu:x:1000:1000:Ubuntu:/home/ubuntu:/bin/bash", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "home",
		},
		{
			name: "user with groups",
			userTest: core.UserTest{
				Name:   "Ubuntu groups",
				User:   "ubuntu",
				Groups: []string{"sudo", "docker"},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("id -u ubuntu 2>/dev/null && id -g ubuntu 2>/dev/null && getent passwd ubuntu 2>/dev/null", "1000\n1000\nubuntu:x:1000:1000:Ubuntu:/home/ubuntu:/bin/bash", "", 0, nil)
				m.SetCommandResult("id -Gn ubuntu 2>/dev/null", "ubuntu sudo docker", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "exists",
		},
		{
			name: "user missing group",
			userTest: core.UserTest{
				Name:   "Ubuntu groups",
				User:   "ubuntu",
				Groups: []string{"sudo", "wheel"},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("id -u ubuntu 2>/dev/null && id -g ubuntu 2>/dev/null && getent passwd ubuntu 2>/dev/null", "1000\n1000\nubuntu:x:1000:1000:Ubuntu:/home/ubuntu:/bin/bash", "", 0, nil)
				m.SetCommandResult("id -Gn ubuntu 2>/dev/null", "ubuntu sudo docker", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "not in group",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := core.NewMockProvider()
			tt.setupMock(mock)

			ctx := context.Background()

			result := executeUserTest(ctx, mock, tt.userTest)

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
