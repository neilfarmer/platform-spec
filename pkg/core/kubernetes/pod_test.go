package kubernetes

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestExecutor_KubernetesPodTest(t *testing.T) {
	tests := []struct {
		name         string
		podTest      core.KubernetesPodTest
		setupMock    func(*core.MockProvider)
		wantStatus   core.Status
		wantContains string
	}{
		{
			name: "pod running",
			podTest: core.KubernetesPodTest{
				Name:      "Nginx pod running",
				Pod:       "nginx-abc123",
				Namespace: "default",
				State:     "running",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get pod nginx-abc123 -n default -o json 2>&1", `{"metadata":{"name":"nginx-abc123"},"status":{"phase":"Running"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "Pod nginx-abc123 is running",
		},
		{
			name: "pod not found",
			podTest: core.KubernetesPodTest{
				Name:      "Pod exists",
				Pod:       "missing-pod",
				Namespace: "default",
				State:     "running",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get pod missing-pod -n default -o json 2>&1", "", "Error from server (NotFound): pods \"missing-pod\" not found", 1, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "Pod missing-pod not found",
		},
		{
			name: "pod pending when expecting running",
			podTest: core.KubernetesPodTest{
				Name:      "Pod running",
				Pod:       "nginx-abc123",
				Namespace: "default",
				State:     "running",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get pod nginx-abc123 -n default -o json 2>&1", `{"metadata":{"name":"nginx-abc123"},"status":{"phase":"Pending"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "phase is Pending, expected Running",
		},
		{
			name: "pod with correct image",
			podTest: core.KubernetesPodTest{
				Name:      "Pod with nginx image",
				Pod:       "nginx-abc123",
				Namespace: "default",
				State:     "exists",
				Image:     "nginx:1.21",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get pod nginx-abc123 -n default -o json 2>&1", `{"metadata":{"name":"nginx-abc123"},"status":{"phase":"Running"},"spec":{"containers":[{"name":"nginx","image":"nginx:1.21"}]}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "Pod nginx-abc123 is running",
		},
		{
			name: "pod with wrong image",
			podTest: core.KubernetesPodTest{
				Name:      "Pod image check",
				Pod:       "nginx-abc123",
				Namespace: "default",
				State:     "exists",
				Image:     "apache",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get pod nginx-abc123 -n default -o json 2>&1", `{"metadata":{"name":"nginx-abc123"},"status":{"phase":"Running"},"spec":{"containers":[{"name":"nginx","image":"nginx:1.21"}]}}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "does not contain image apache",
		},
		{
			name: "pod ready",
			podTest: core.KubernetesPodTest{
				Name:      "Pod ready check",
				Pod:       "nginx-abc123",
				Namespace: "default",
				State:     "running",
				Ready:     true,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get pod nginx-abc123 -n default -o json 2>&1", `{"metadata":{"name":"nginx-abc123"},"status":{"phase":"Running","containerStatuses":[{"ready":true}]}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "Pod nginx-abc123 is running",
		},
		{
			name: "pod not ready",
			podTest: core.KubernetesPodTest{
				Name:      "Pod ready check",
				Pod:       "nginx-abc123",
				Namespace: "default",
				State:     "running",
				Ready:     true,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get pod nginx-abc123 -n default -o json 2>&1", `{"metadata":{"name":"nginx-abc123"},"status":{"phase":"Running","containerStatuses":[{"ready":false}]}}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "containers not all ready",
		},
		{
			name: "pod with correct labels",
			podTest: core.KubernetesPodTest{
				Name:      "Pod with labels",
				Pod:       "nginx-abc123",
				Namespace: "default",
				State:     "exists",
				Labels:    map[string]string{"app": "nginx", "version": "1.21"},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get pod nginx-abc123 -n default -o json 2>&1", `{"metadata":{"name":"nginx-abc123","labels":{"app":"nginx","version":"1.21"}},"status":{"phase":"Running"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "Pod nginx-abc123 is running",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := core.NewMockProvider()
			tt.setupMock(mock)

			ctx := context.Background()

			result := executeKubernetesPodTest(ctx, mock, tt.podTest)

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
