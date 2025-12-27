package kubernetes

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestExecutor_KubernetesHelmTest(t *testing.T) {
	tests := []struct {
		name         string
		helmTest     core.KubernetesHelmTest
		setupMock    func(*core.MockProvider)
		wantStatus   core.Status
		wantContains string
	}{
		{
			name: "helm release deployed",
			helmTest: core.KubernetesHelmTest{
				Name:      "Prometheus deployed",
				Release:   "prometheus",
				Namespace: "monitoring",
				State:     "deployed",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("helm list -n monitoring -o json 2>&1", `[{"name":"prometheus","namespace":"monitoring","revision":"1","updated":"2024-01-01","status":"deployed","chart":"prometheus-25.0.0","app_version":"v2.45.0"}]`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "Helm release prometheus is deployed",
		},
		{
			name: "helm release not found",
			helmTest: core.KubernetesHelmTest{
				Name:      "Release exists",
				Release:   "missing-release",
				Namespace: "default",
				State:     "deployed",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("helm list -n default -o json 2>&1", `[]`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "Helm release missing-release not found",
		},
		{
			name: "helm release in wrong state",
			helmTest: core.KubernetesHelmTest{
				Name:      "Release should be deployed",
				Release:   "prometheus",
				Namespace: "monitoring",
				State:     "deployed",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("helm list -n monitoring -o json 2>&1", `[{"name":"prometheus","namespace":"monitoring","status":"failed"}]`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "is failed, expected deployed",
		},
		{
			name: "helm release with all pods ready",
			helmTest: core.KubernetesHelmTest{
				Name:         "Prometheus fully operational",
				Release:      "prometheus",
				Namespace:    "monitoring",
				State:        "deployed",
				AllPodsReady: true,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("helm list -n monitoring -o json 2>&1", `[{"name":"prometheus","namespace":"monitoring","status":"deployed"}]`, "", 0, nil)
				m.SetCommandResult("kubectl get pods -n monitoring -l app.kubernetes.io/instance=prometheus -o json 2>&1", `{"items":[
					{"metadata":{"name":"prometheus-server"},"status":{"phase":"Running","containerStatuses":[{"ready":true}]}},
					{"metadata":{"name":"prometheus-alertmanager"},"status":{"phase":"Running","containerStatuses":[{"ready":true}]}}
				]}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "with all pods ready",
		},
		{
			name: "helm release with pods not ready",
			helmTest: core.KubernetesHelmTest{
				Name:         "Release with pods check",
				Release:      "prometheus",
				Namespace:    "monitoring",
				State:        "deployed",
				AllPodsReady: true,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("helm list -n monitoring -o json 2>&1", `[{"name":"prometheus","namespace":"monitoring","status":"deployed"}]`, "", 0, nil)
				m.SetCommandResult("kubectl get pods -n monitoring -l app.kubernetes.io/instance=prometheus -o json 2>&1", `{"items":[
					{"metadata":{"name":"prometheus-server"},"status":{"phase":"Running","containerStatuses":[{"ready":true}]}},
					{"metadata":{"name":"prometheus-alertmanager"},"status":{"phase":"Pending","containerStatuses":[{"ready":false}]}}
				]}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "has 1/2 pods ready",
		},
		{
			name: "helm release with crashloop pod",
			helmTest: core.KubernetesHelmTest{
				Name:         "Release health check",
				Release:      "prometheus",
				Namespace:    "monitoring",
				State:        "deployed",
				AllPodsReady: true,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("helm list -n monitoring -o json 2>&1", `[{"name":"prometheus","namespace":"monitoring","status":"deployed"}]`, "", 0, nil)
				m.SetCommandResult("kubectl get pods -n monitoring -l app.kubernetes.io/instance=prometheus -o json 2>&1", `{"items":[
					{"metadata":{"name":"prometheus-server"},"status":{"phase":"Running","containerStatuses":[{"ready":false,"state":{"waiting":{"reason":"CrashLoopBackOff"}}}]}}
				]}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "pods in bad state",
		},
		{
			name: "helm release with no pods",
			helmTest: core.KubernetesHelmTest{
				Name:         "Release with no pods",
				Release:      "prometheus",
				Namespace:    "monitoring",
				State:        "deployed",
				AllPodsReady: true,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("helm list -n monitoring -o json 2>&1", `[{"name":"prometheus","namespace":"monitoring","status":"deployed"}]`, "", 0, nil)
				m.SetCommandResult("kubectl get pods -n monitoring -l app.kubernetes.io/instance=prometheus -o json 2>&1", `{"items":[]}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "No pods found",
		},
		{
			name: "helm command fails",
			helmTest: core.KubernetesHelmTest{
				Name:      "Release check",
				Release:   "prometheus",
				Namespace: "monitoring",
				State:     "deployed",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("helm list -n monitoring -o json 2>&1", "", "Error: Kubernetes cluster unreachable", 1, nil)
			},
			wantStatus:   core.StatusError,
			wantContains: "helm error",
		},
		{
			name: "invalid helm json",
			helmTest: core.KubernetesHelmTest{
				Name:      "Release check",
				Release:   "prometheus",
				Namespace: "monitoring",
				State:     "deployed",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("helm list -n monitoring -o json 2>&1", "invalid json", "", 0, nil)
			},
			wantStatus:   core.StatusError,
			wantContains: "Failed to parse Helm output",
		},
		{
			name: "multiple releases in namespace",
			helmTest: core.KubernetesHelmTest{
				Name:      "Grafana deployed",
				Release:   "grafana",
				Namespace: "monitoring",
				State:     "deployed",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("helm list -n monitoring -o json 2>&1", `[
					{"name":"prometheus","namespace":"monitoring","status":"deployed"},
					{"name":"grafana","namespace":"monitoring","status":"deployed"},
					{"name":"alertmanager","namespace":"monitoring","status":"deployed"}
				]`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "Helm release grafana is deployed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := core.NewMockProvider()
			tt.setupMock(mock)

			ctx := context.Background()

			result := executeKubernetesHelmTest(ctx, mock, tt.helmTest)

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
