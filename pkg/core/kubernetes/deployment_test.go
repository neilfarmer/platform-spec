package kubernetes

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestExecutor_KubernetesDeploymentTest(t *testing.T) {
	tests := []struct {
		name         string
		deploymentTest core.KubernetesDeploymentTest
		setupMock    func(*core.MockProvider)
		wantStatus   core.Status
		wantContains string
	}{
		{
			name: "deployment available",
			deploymentTest: core.KubernetesDeploymentTest{
				Name:       "Nginx deployment available",
				Deployment: "nginx",
				Namespace:  "default",
				State:      "available",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get deployment nginx -n default -o json 2>&1", `{"metadata":{"name":"nginx"},"status":{"conditions":[{"type":"Available","status":"True"}]}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "Deployment nginx is available",
		},
		{
			name: "deployment not found",
			deploymentTest: core.KubernetesDeploymentTest{
				Name:       "Deployment exists",
				Deployment: "missing",
				Namespace:  "default",
				State:      "available",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get deployment missing -n default -o json 2>&1", "", "Error from server (NotFound): deployments.apps \"missing\" not found", 1, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "Deployment missing not found",
		},
		{
			name: "deployment not available",
			deploymentTest: core.KubernetesDeploymentTest{
				Name:       "Deployment available",
				Deployment: "nginx",
				Namespace:  "default",
				State:      "available",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get deployment nginx -n default -o json 2>&1", `{"metadata":{"name":"nginx"},"status":{"conditions":[{"type":"Available","status":"False"}]}}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "Deployment nginx is not available",
		},
		{
			name: "deployment with correct replicas",
			deploymentTest: core.KubernetesDeploymentTest{
				Name:       "Deployment replicas",
				Deployment: "nginx",
				Namespace:  "default",
				State:      "exists",
				Replicas:   3,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get deployment nginx -n default -o json 2>&1", `{"metadata":{"name":"nginx"},"spec":{"replicas":3},"status":{}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "Deployment nginx is exists",
		},
		{
			name: "deployment with wrong replicas",
			deploymentTest: core.KubernetesDeploymentTest{
				Name:       "Deployment replicas",
				Deployment: "nginx",
				Namespace:  "default",
				State:      "exists",
				Replicas:   5,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get deployment nginx -n default -o json 2>&1", `{"metadata":{"name":"nginx"},"spec":{"replicas":3},"status":{}}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "has 3 replicas, expected 5",
		},
		{
			name: "deployment with correct ready replicas",
			deploymentTest: core.KubernetesDeploymentTest{
				Name:          "Deployment ready replicas",
				Deployment:    "nginx",
				Namespace:     "default",
				State:         "exists",
				ReadyReplicas: 3,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get deployment nginx -n default -o json 2>&1", `{"metadata":{"name":"nginx"},"status":{"readyReplicas":3}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "Deployment nginx is exists",
		},
		{
			name: "deployment with correct image",
			deploymentTest: core.KubernetesDeploymentTest{
				Name:       "Deployment image",
				Deployment: "nginx",
				Namespace:  "default",
				State:      "exists",
				Image:      "nginx:1.21",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get deployment nginx -n default -o json 2>&1", `{"metadata":{"name":"nginx"},"spec":{"template":{"spec":{"containers":[{"name":"nginx","image":"nginx:1.21"}]}}},"status":{}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "Deployment nginx is exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := core.NewMockProvider()
			tt.setupMock(mock)

			ctx := context.Background()

			result := executeKubernetesDeploymentTest(ctx, mock, tt.deploymentTest)

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
