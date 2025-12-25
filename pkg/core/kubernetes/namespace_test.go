package kubernetes

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestExecutor_KubernetesNamespaceTest(t *testing.T) {
	tests := []struct {
		name         string
		namespaceTest core.KubernetesNamespaceTest
		setupMock    func(*core.MockProvider)
		wantStatus   core.Status
		wantContains string
	}{
		{
			name: "namespace exists",
			namespaceTest: core.KubernetesNamespaceTest{
				Name:      "Production namespace exists",
				Namespace: "production",
				State:     "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get namespace production -o json 2>&1", `{"metadata":{"name":"production"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "Namespace production exists",
		},
		{
			name: "namespace does not exist",
			namespaceTest: core.KubernetesNamespaceTest{
				Name:      "Test namespace exists",
				Namespace: "test",
				State:     "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get namespace test -o json 2>&1", "", "Error from server (NotFound): namespaces \"test\" not found", 1, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "Namespace test not found",
		},
		{
			name: "namespace absent as expected",
			namespaceTest: core.KubernetesNamespaceTest{
				Name:      "Namespace should be absent",
				Namespace: "deleted",
				State:     "absent",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get namespace deleted -o json 2>&1", "", "Error from server (NotFound): namespaces \"deleted\" not found", 1, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "Namespace deleted is absent",
		},
		{
			name: "namespace exists but should be absent",
			namespaceTest: core.KubernetesNamespaceTest{
				Name:      "Namespace should not exist",
				Namespace: "production",
				State:     "absent",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get namespace production -o json 2>&1", `{"metadata":{"name":"production"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "exists but should be absent",
		},
		{
			name: "namespace with correct labels",
			namespaceTest: core.KubernetesNamespaceTest{
				Name:      "Production namespace with labels",
				Namespace: "production",
				State:     "present",
				Labels:    map[string]string{"environment": "prod", "team": "platform"},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get namespace production -o json 2>&1", `{"metadata":{"name":"production","labels":{"environment":"prod","team":"platform"}}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "Namespace production exists",
		},
		{
			name: "namespace with wrong label",
			namespaceTest: core.KubernetesNamespaceTest{
				Name:      "Namespace label check",
				Namespace: "production",
				State:     "present",
				Labels:    map[string]string{"environment": "dev"},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get namespace production -o json 2>&1", `{"metadata":{"name":"production","labels":{"environment":"prod"}}}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "label environment=dev not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := core.NewMockProvider()
			tt.setupMock(mock)

			ctx := context.Background()

			result := executeKubernetesNamespaceTest(ctx, mock, tt.namespaceTest)

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
