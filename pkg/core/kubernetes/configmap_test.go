package kubernetes

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestExecutor_KubernetesConfigMapTest(t *testing.T) {
	tests := []struct {
		name         string
		configMapTest core.KubernetesConfigMapTest
		setupMock    func(*core.MockProvider)
		wantStatus   core.Status
		wantContains string
	}{
		{
			name: "configmap exists",
			configMapTest: core.KubernetesConfigMapTest{
				Name:      "App config exists",
				ConfigMap: "app-config",
				Namespace: "default",
				State:     "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get configmap app-config -n default -o json 2>&1", `{"metadata":{"name":"app-config"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "ConfigMap app-config exists",
		},
		{
			name: "configmap not found",
			configMapTest: core.KubernetesConfigMapTest{
				Name:      "ConfigMap exists",
				ConfigMap: "missing",
				Namespace: "default",
				State:     "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get configmap missing -n default -o json 2>&1", "", "Error from server (NotFound): configmaps \"missing\" not found", 1, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "ConfigMap missing not found",
		},
		{
			name: "configmap absent as expected",
			configMapTest: core.KubernetesConfigMapTest{
				Name:      "ConfigMap should be absent",
				ConfigMap: "deleted",
				Namespace: "default",
				State:     "absent",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get configmap deleted -n default -o json 2>&1", "", "Error from server (NotFound): configmaps \"deleted\" not found", 1, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "ConfigMap deleted is absent",
		},
		{
			name: "configmap exists but should be absent",
			configMapTest: core.KubernetesConfigMapTest{
				Name:      "ConfigMap should not exist",
				ConfigMap: "app-config",
				Namespace: "default",
				State:     "absent",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get configmap app-config -n default -o json 2>&1", `{"metadata":{"name":"app-config"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "exists but should be absent",
		},
		{
			name: "configmap with all required keys",
			configMapTest: core.KubernetesConfigMapTest{
				Name:      "ConfigMap with keys",
				ConfigMap: "app-config",
				Namespace: "default",
				State:     "present",
				HasKeys:   []string{"config.yaml", "settings.json"},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get configmap app-config -n default -o json 2>&1", `{"metadata":{"name":"app-config"},"data":{"config.yaml":"data","settings.json":"data"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "exists with all required keys",
		},
		{
			name: "configmap missing keys",
			configMapTest: core.KubernetesConfigMapTest{
				Name:      "ConfigMap keys",
				ConfigMap: "app-config",
				Namespace: "default",
				State:     "present",
				HasKeys:   []string{"config.yaml", "missing-key"},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get configmap app-config -n default -o json 2>&1", `{"metadata":{"name":"app-config"},"data":{"config.yaml":"data"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "missing keys: missing-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := core.NewMockProvider()
			tt.setupMock(mock)

			ctx := context.Background()

			result := executeKubernetesConfigMapTest(ctx, mock, tt.configMapTest)

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
