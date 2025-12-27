package kubernetes

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestExecutor_KubernetesIngressTest(t *testing.T) {
	tests := []struct {
		name         string
		ingressTest  core.KubernetesIngressTest
		setupMock    func(*core.MockProvider)
		wantStatus   core.Status
		wantContains string
	}{
		{
			name: "ingress exists",
			ingressTest: core.KubernetesIngressTest{
				Name:      "App ingress exists",
				Ingress:   "myapp-ingress",
				Namespace: "default",
				State:     "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get ingress myapp-ingress -n default -o json 2>&1", `{"metadata":{"name":"myapp-ingress"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "Ingress myapp-ingress exists",
		},
		{
			name: "ingress not found",
			ingressTest: core.KubernetesIngressTest{
				Name:      "Ingress exists",
				Ingress:   "missing-ingress",
				Namespace: "default",
				State:     "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get ingress missing-ingress -n default -o json 2>&1", "", "Error from server (NotFound): ingresses.networking.k8s.io \"missing-ingress\" not found", 1, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "Ingress missing-ingress not found",
		},
		{
			name: "ingress absent as expected",
			ingressTest: core.KubernetesIngressTest{
				Name:      "Old ingress removed",
				Ingress:   "old-ingress",
				Namespace: "default",
				State:     "absent",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get ingress old-ingress -n default -o json 2>&1", "", "Error from server (NotFound): ingresses.networking.k8s.io \"old-ingress\" not found", 1, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "Ingress old-ingress is absent",
		},
		{
			name: "ingress with correct hosts",
			ingressTest: core.KubernetesIngressTest{
				Name:      "Ingress hosts configured",
				Ingress:   "myapp-ingress",
				Namespace: "default",
				State:     "present",
				Hosts:     []string{"app.example.com", "www.example.com"},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get ingress myapp-ingress -n default -o json 2>&1", `{"metadata":{"name":"myapp-ingress"},"spec":{"rules":[{"host":"app.example.com"},{"host":"www.example.com"}]}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "all hosts",
		},
		{
			name: "ingress missing hosts",
			ingressTest: core.KubernetesIngressTest{
				Name:      "Ingress hosts validation",
				Ingress:   "myapp-ingress",
				Namespace: "default",
				State:     "present",
				Hosts:     []string{"app.example.com", "api.example.com"},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get ingress myapp-ingress -n default -o json 2>&1", `{"metadata":{"name":"myapp-ingress"},"spec":{"rules":[{"host":"app.example.com"}]}}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "missing hosts: api.example.com",
		},
		{
			name: "ingress with TLS",
			ingressTest: core.KubernetesIngressTest{
				Name:      "Ingress TLS configured",
				Ingress:   "myapp-ingress",
				Namespace: "default",
				State:     "present",
				TLS:       true,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get ingress myapp-ingress -n default -o json 2>&1", `{"metadata":{"name":"myapp-ingress"},"spec":{"tls":[{"hosts":["app.example.com"],"secretName":"tls-secret"}]}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "TLS configured",
		},
		{
			name: "ingress without TLS",
			ingressTest: core.KubernetesIngressTest{
				Name:      "Ingress TLS check",
				Ingress:   "myapp-ingress",
				Namespace: "default",
				State:     "present",
				TLS:       true,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get ingress myapp-ingress -n default -o json 2>&1", `{"metadata":{"name":"myapp-ingress"},"spec":{"rules":[{"host":"app.example.com"}]}}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "does not have TLS configured",
		},
		{
			name: "ingress with correct class",
			ingressTest: core.KubernetesIngressTest{
				Name:         "Ingress class validation",
				Ingress:      "myapp-ingress",
				Namespace:    "default",
				State:        "present",
				IngressClass: "nginx",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get ingress myapp-ingress -n default -o json 2>&1", `{"metadata":{"name":"myapp-ingress"},"spec":{"ingressClassName":"nginx"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "correct class",
		},
		{
			name: "ingress with wrong class",
			ingressTest: core.KubernetesIngressTest{
				Name:         "Ingress class mismatch",
				Ingress:      "myapp-ingress",
				Namespace:    "default",
				State:        "present",
				IngressClass: "nginx",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get ingress myapp-ingress -n default -o json 2>&1", `{"metadata":{"name":"myapp-ingress"},"spec":{"ingressClassName":"traefik"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "has class traefik, expected nginx",
		},
		{
			name: "kubectl command fails",
			ingressTest: core.KubernetesIngressTest{
				Name:      "Ingress check",
				Ingress:   "test-ingress",
				Namespace: "default",
				State:     "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get ingress test-ingress -n default -o json 2>&1", "", "error: connection refused", 1, nil)
			},
			wantStatus:   core.StatusError,
			wantContains: "kubectl error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := core.NewMockProvider()
			tt.setupMock(mock)

			ctx := context.Background()

			result := executeKubernetesIngressTest(ctx, mock, tt.ingressTest)

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
