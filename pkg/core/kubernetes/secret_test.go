package kubernetes

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestExecutor_KubernetesSecretTest(t *testing.T) {
	tests := []struct {
		name         string
		secretTest   core.KubernetesSecretTest
		setupMock    func(*core.MockProvider)
		wantStatus   core.Status
		wantContains string
	}{
		{
			name: "secret exists",
			secretTest: core.KubernetesSecretTest{
				Name:      "Database password exists",
				Secret:    "db-password",
				Namespace: "default",
				State:     "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get secret db-password -n default -o json 2>&1", `{"metadata":{"name":"db-password"},"type":"Opaque"}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "Secret db-password exists",
		},
		{
			name: "secret not found",
			secretTest: core.KubernetesSecretTest{
				Name:      "Secret exists",
				Secret:    "missing-secret",
				Namespace: "default",
				State:     "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get secret missing-secret -n default -o json 2>&1", "", "Error from server (NotFound): secrets \"missing-secret\" not found", 1, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "Secret missing-secret not found",
		},
		{
			name: "secret absent as expected",
			secretTest: core.KubernetesSecretTest{
				Name:      "Old secret removed",
				Secret:    "deprecated-secret",
				Namespace: "default",
				State:     "absent",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get secret deprecated-secret -n default -o json 2>&1", "", "Error from server (NotFound): secrets \"deprecated-secret\" not found", 1, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "Secret deprecated-secret is absent",
		},
		{
			name: "secret exists but should be absent",
			secretTest: core.KubernetesSecretTest{
				Name:      "Secret should not exist",
				Secret:    "db-password",
				Namespace: "default",
				State:     "absent",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get secret db-password -n default -o json 2>&1", `{"metadata":{"name":"db-password"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "exists but should be absent",
		},
		{
			name: "secret with correct type",
			secretTest: core.KubernetesSecretTest{
				Name:      "TLS secret with correct type",
				Secret:    "tls-cert",
				Namespace: "default",
				State:     "present",
				Type:      "kubernetes.io/tls",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get secret tls-cert -n default -o json 2>&1", `{"metadata":{"name":"tls-cert"},"type":"kubernetes.io/tls"}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "exists with correct type",
		},
		{
			name: "secret with wrong type",
			secretTest: core.KubernetesSecretTest{
				Name:      "Secret type validation",
				Secret:    "my-secret",
				Namespace: "default",
				State:     "present",
				Type:      "kubernetes.io/tls",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get secret my-secret -n default -o json 2>&1", `{"metadata":{"name":"my-secret"},"type":"Opaque"}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "has type Opaque, expected kubernetes.io/tls",
		},
		{
			name: "secret with all required keys",
			secretTest: core.KubernetesSecretTest{
				Name:      "Secret with keys",
				Secret:    "app-creds",
				Namespace: "default",
				State:     "present",
				HasKeys:   []string{"username", "password"},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get secret app-creds -n default -o json 2>&1", `{"metadata":{"name":"app-creds"},"type":"Opaque","data":{"username":"dXNlcg==","password":"cGFzcw=="}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "exists with all required keys",
		},
		{
			name: "secret missing keys",
			secretTest: core.KubernetesSecretTest{
				Name:      "Secret keys validation",
				Secret:    "app-creds",
				Namespace: "default",
				State:     "present",
				HasKeys:   []string{"username", "password", "api-key"},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get secret app-creds -n default -o json 2>&1", `{"metadata":{"name":"app-creds"},"type":"Opaque","data":{"username":"dXNlcg==","password":"cGFzcw=="}}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "missing keys: api-key",
		},
		{
			name: "secret with type and keys",
			secretTest: core.KubernetesSecretTest{
				Name:      "Basic auth secret validation",
				Secret:    "basic-auth",
				Namespace: "default",
				State:     "present",
				Type:      "kubernetes.io/basic-auth",
				HasKeys:   []string{"username", "password"},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get secret basic-auth -n default -o json 2>&1", `{"metadata":{"name":"basic-auth"},"type":"kubernetes.io/basic-auth","data":{"username":"dXNlcg==","password":"cGFzcw=="}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "exists with correct type and all required keys",
		},
		{
			name: "TLS secret with required keys",
			secretTest: core.KubernetesSecretTest{
				Name:      "TLS certificate complete",
				Secret:    "tls-secret",
				Namespace: "ingress",
				State:     "present",
				Type:      "kubernetes.io/tls",
				HasKeys:   []string{"tls.crt", "tls.key"},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get secret tls-secret -n ingress -o json 2>&1", `{"metadata":{"name":"tls-secret"},"type":"kubernetes.io/tls","data":{"tls.crt":"Y2VydA==","tls.key":"a2V5"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "exists with correct type and all required keys",
		},
		{
			name: "docker config secret",
			secretTest: core.KubernetesSecretTest{
				Name:      "Docker registry credentials",
				Secret:    "regcred",
				Namespace: "default",
				State:     "present",
				Type:      "kubernetes.io/dockerconfigjson",
				HasKeys:   []string{".dockerconfigjson"},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get secret regcred -n default -o json 2>&1", `{"metadata":{"name":"regcred"},"type":"kubernetes.io/dockerconfigjson","data":{".dockerconfigjson":"eyJhdXRocyI6e319"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "exists with correct type and all required keys",
		},
		{
			name: "kubectl command fails",
			secretTest: core.KubernetesSecretTest{
				Name:      "Secret check",
				Secret:    "test-secret",
				Namespace: "default",
				State:     "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get secret test-secret -n default -o json 2>&1", "", "error: connection refused", 1, nil)
			},
			wantStatus:   core.StatusError,
			wantContains: "kubectl error",
		},
		{
			name: "invalid secret json",
			secretTest: core.KubernetesSecretTest{
				Name:      "Secret with type check",
				Secret:    "test-secret",
				Namespace: "default",
				State:     "present",
				Type:      "Opaque",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get secret test-secret -n default -o json 2>&1", "invalid json", "", 0, nil)
			},
			wantStatus:   core.StatusError,
			wantContains: "Failed to parse secret JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := core.NewMockProvider()
			tt.setupMock(mock)

			ctx := context.Background()

			result := executeKubernetesSecretTest(ctx, mock, tt.secretTest)

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
