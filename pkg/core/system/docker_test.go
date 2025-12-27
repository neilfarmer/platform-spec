package system

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestExecutor_DockerTest(t *testing.T) {
	tests := []struct {
		name         string
		dockerTest   core.DockerTest
		setupMock    func(*core.MockProvider)
		wantStatus   core.Status
		wantContains string
	}{
		{
			name: "container running",
			dockerTest: core.DockerTest{
				Name:      "Check nginx container",
				Container: "nginx",
				State:     "running",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("docker inspect --format '{{.State.Status}}|{{.Config.Image}}|{{.HostConfig.RestartPolicy.Name}}|{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' nginx 2>/dev/null", "running|nginx:latest|always|none", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "containers are running",
		},
		{
			name: "container stopped",
			dockerTest: core.DockerTest{
				Name:      "Check stopped container",
				Container: "old-app",
				State:     "stopped",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("docker inspect --format '{{.State.Status}}|{{.Config.Image}}|{{.HostConfig.RestartPolicy.Name}}|{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' old-app 2>/dev/null", "exited|app:1.0|no|none", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "containers are stopped",
		},
		{
			name: "container does not exist",
			dockerTest: core.DockerTest{
				Name:      "Check missing container",
				Container: "missing",
				State:     "running",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("docker inspect --format '{{.State.Status}}|{{.Config.Image}}|{{.HostConfig.RestartPolicy.Name}}|{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' missing 2>/dev/null", "", "", 1, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "does not exist",
		},
		{
			name: "container running but should be stopped",
			dockerTest: core.DockerTest{
				Name:      "Should be stopped",
				Container: "app",
				State:     "stopped",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("docker inspect --format '{{.State.Status}}|{{.Config.Image}}|{{.HostConfig.RestartPolicy.Name}}|{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' app 2>/dev/null", "running|app:latest|always|none", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "is running, expected stopped",
		},
		{
			name: "container with correct image",
			dockerTest: core.DockerTest{
				Name:      "Check container image",
				Container: "web",
				State:     "running",
				Image:     "nginx:1.21",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("docker inspect --format '{{.State.Status}}|{{.Config.Image}}|{{.HostConfig.RestartPolicy.Name}}|{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' web 2>/dev/null", "running|nginx:1.21|always|none", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "with correct image",
		},
		{
			name: "container with wrong image",
			dockerTest: core.DockerTest{
				Name:      "Check wrong image",
				Container: "web",
				State:     "running",
				Image:     "nginx:1.22",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("docker inspect --format '{{.State.Status}}|{{.Config.Image}}|{{.HostConfig.RestartPolicy.Name}}|{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' web 2>/dev/null", "running|nginx:1.21|always|none", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "image is",
		},
		{
			name: "container with correct restart policy",
			dockerTest: core.DockerTest{
				Name:          "Check restart policy",
				Container:     "db",
				State:         "running",
				RestartPolicy: "always",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("docker inspect --format '{{.State.Status}}|{{.Config.Image}}|{{.HostConfig.RestartPolicy.Name}}|{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' db 2>/dev/null", "running|postgres:14|always|none", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "restart policy",
		},
		{
			name: "container with wrong restart policy",
			dockerTest: core.DockerTest{
				Name:          "Check wrong restart policy",
				Container:     "db",
				State:         "running",
				RestartPolicy: "on-failure",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("docker inspect --format '{{.State.Status}}|{{.Config.Image}}|{{.HostConfig.RestartPolicy.Name}}|{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' db 2>/dev/null", "running|postgres:14|always|none", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "restart policy is",
		},
		{
			name: "container with healthy status",
			dockerTest: core.DockerTest{
				Name:      "Check health",
				Container: "api",
				State:     "running",
				Health:    "healthy",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("docker inspect --format '{{.State.Status}}|{{.Config.Image}}|{{.HostConfig.RestartPolicy.Name}}|{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' api 2>/dev/null", "running|api:v1|always|healthy", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "health status",
		},
		{
			name: "container with unhealthy status",
			dockerTest: core.DockerTest{
				Name:      "Check unhealthy",
				Container: "api",
				State:     "running",
				Health:    "healthy",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("docker inspect --format '{{.State.Status}}|{{.Config.Image}}|{{.HostConfig.RestartPolicy.Name}}|{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' api 2>/dev/null", "running|api:v1|always|unhealthy", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "health is",
		},
		{
			name: "container with all properties",
			dockerTest: core.DockerTest{
				Name:          "Check all properties",
				Container:     "webapp",
				State:         "running",
				Image:         "nginx:alpine",
				RestartPolicy: "unless-stopped",
				Health:        "healthy",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("docker inspect --format '{{.State.Status}}|{{.Config.Image}}|{{.HostConfig.RestartPolicy.Name}}|{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' webapp 2>/dev/null", "running|nginx:alpine|unless-stopped|healthy", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "are running",
		},
		{
			name: "multiple containers",
			dockerTest: core.DockerTest{
				Name:       "Check multiple containers",
				Containers: []string{"web1", "web2"},
				State:      "running",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("docker inspect --format '{{.State.Status}}|{{.Config.Image}}|{{.HostConfig.RestartPolicy.Name}}|{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' web1 2>/dev/null", "running|nginx:latest|always|none", "", 0, nil)
				m.SetCommandResult("docker inspect --format '{{.State.Status}}|{{.Config.Image}}|{{.HostConfig.RestartPolicy.Name}}|{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' web2 2>/dev/null", "running|nginx:latest|always|none", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "2 containers are running",
		},
		{
			name: "multiple containers - one fails",
			dockerTest: core.DockerTest{
				Name:       "Check multiple - one missing",
				Containers: []string{"web1", "missing"},
				State:      "running",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("docker inspect --format '{{.State.Status}}|{{.Config.Image}}|{{.HostConfig.RestartPolicy.Name}}|{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' web1 2>/dev/null", "running|nginx:latest|always|none", "", 0, nil)
				m.SetCommandResult("docker inspect --format '{{.State.Status}}|{{.Config.Image}}|{{.HostConfig.RestartPolicy.Name}}|{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' missing 2>/dev/null", "", "", 1, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "does not exist",
		},
		{
			name: "container exists check",
			dockerTest: core.DockerTest{
				Name:      "Just check existence",
				Container: "test",
				State:     "exists",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("docker inspect --format '{{.State.Status}}|{{.Config.Image}}|{{.HostConfig.RestartPolicy.Name}}|{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' test 2>/dev/null", "created|test:v1|no|none", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "exist",
		},
		{
			name: "unexpected docker inspect output",
			dockerTest: core.DockerTest{
				Name:      "Malformed output",
				Container: "bad",
				State:     "running",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("docker inspect --format '{{.State.Status}}|{{.Config.Image}}|{{.HostConfig.RestartPolicy.Name}}|{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' bad 2>/dev/null", "running|incomplete", "", 0, nil)
			},
			wantStatus:   core.StatusError,
			wantContains: "Unexpected docker inspect output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := core.NewMockProvider()
			tt.setupMock(mock)

			ctx := context.Background()

			result := executeDockerTest(ctx, mock, tt.dockerTest)

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
