package kubernetes

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestExecutor_KubernetesServiceTest(t *testing.T) {
	tests := []struct {
		name         string
		serviceTest  core.KubernetesServiceTest
		setupMock    func(*core.MockProvider)
		wantStatus   core.Status
		wantContains string
	}{
		{
			name: "service exists",
			serviceTest: core.KubernetesServiceTest{
				Name:      "Nginx service exists",
				Service:   "nginx",
				Namespace: "default",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get service nginx -n default -o json 2>&1", `{"metadata":{"name":"nginx"},"spec":{"type":"ClusterIP"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "Service nginx exists",
		},
		{
			name: "service not found",
			serviceTest: core.KubernetesServiceTest{
				Name:      "Service exists",
				Service:   "missing",
				Namespace: "default",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get service missing -n default -o json 2>&1", "", "Error from server (NotFound): services \"missing\" not found", 1, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "Service missing not found",
		},
		{
			name: "service with correct type",
			serviceTest: core.KubernetesServiceTest{
				Name:      "Service type",
				Service:   "nginx",
				Namespace: "default",
				Type:      "LoadBalancer",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get service nginx -n default -o json 2>&1", `{"metadata":{"name":"nginx"},"spec":{"type":"LoadBalancer"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "Service nginx is type LoadBalancer",
		},
		{
			name: "service with wrong type",
			serviceTest: core.KubernetesServiceTest{
				Name:      "Service type",
				Service:   "nginx",
				Namespace: "default",
				Type:      "LoadBalancer",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get service nginx -n default -o json 2>&1", `{"metadata":{"name":"nginx"},"spec":{"type":"ClusterIP"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "type is ClusterIP, expected LoadBalancer",
		},
		{
			name: "service with correct ports",
			serviceTest: core.KubernetesServiceTest{
				Name:      "Service ports",
				Service:   "nginx",
				Namespace: "default",
				Ports: []core.KubernetesServicePort{
					{Port: 80, Protocol: "TCP"},
					{Port: 443, Protocol: "TCP"},
				},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get service nginx -n default -o json 2>&1", `{"metadata":{"name":"nginx"},"spec":{"type":"ClusterIP","ports":[{"port":80,"protocol":"TCP"},{"port":443,"protocol":"TCP"}]}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "Service nginx exists",
		},
		{
			name: "service missing port",
			serviceTest: core.KubernetesServiceTest{
				Name:      "Service ports",
				Service:   "nginx",
				Namespace: "default",
				Ports: []core.KubernetesServicePort{
					{Port: 8080, Protocol: "TCP"},
				},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get service nginx -n default -o json 2>&1", `{"metadata":{"name":"nginx"},"spec":{"type":"ClusterIP","ports":[{"port":80,"protocol":"TCP"}]}}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "does not have port 8080/TCP",
		},
		{
			name: "service with correct selector",
			serviceTest: core.KubernetesServiceTest{
				Name:      "Service selector",
				Service:   "nginx",
				Namespace: "default",
				Selector:  map[string]string{"app": "nginx"},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get service nginx -n default -o json 2>&1", `{"metadata":{"name":"nginx"},"spec":{"type":"ClusterIP","selector":{"app":"nginx"}}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "Service nginx exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := core.NewMockProvider()
			tt.setupMock(mock)

			ctx := context.Background()

			result := executeKubernetesServiceTest(ctx, mock, tt.serviceTest)

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
