package kubernetes

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestExecutor_KubernetesStatefulSetTest(t *testing.T) {
	tests := []struct {
		name             string
		statefulSetTest  core.KubernetesStatefulSetTest
		setupMock        func(*core.MockProvider)
		wantStatus       core.Status
		wantContains     string
	}{
		{
			name: "statefulset available",
			statefulSetTest: core.KubernetesStatefulSetTest{
				Name:        "Database available",
				StatefulSet: "postgres",
				Namespace:   "default",
				State:       "available",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get statefulset postgres -n default -o json 2>&1", `{"metadata":{"name":"postgres"},"spec":{"replicas":3},"status":{"readyReplicas":3,"currentReplicas":3}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "is available",
		},
		{
			name: "statefulset not available",
			statefulSetTest: core.KubernetesStatefulSetTest{
				Name:        "Database should be available",
				StatefulSet: "postgres",
				Namespace:   "default",
				State:       "available",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get statefulset postgres -n default -o json 2>&1", `{"metadata":{"name":"postgres"},"spec":{"replicas":3},"status":{"readyReplicas":1,"currentReplicas":3}}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "is not available",
		},
		{
			name: "statefulset exists",
			statefulSetTest: core.KubernetesStatefulSetTest{
				Name:        "Database exists",
				StatefulSet: "postgres",
				Namespace:   "default",
				State:       "exists",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get statefulset postgres -n default -o json 2>&1", `{"metadata":{"name":"postgres"},"spec":{"replicas":3},"status":{"readyReplicas":1,"currentReplicas":3}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "exists with 1/3 ready replicas",
		},
		{
			name: "statefulset not found",
			statefulSetTest: core.KubernetesStatefulSetTest{
				Name:        "StatefulSet exists",
				StatefulSet: "missing",
				Namespace:   "default",
				State:       "available",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get statefulset missing -n default -o json 2>&1", "", "Error from server (NotFound): statefulsets.apps \"missing\" not found", 1, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "StatefulSet missing not found",
		},
		{
			name: "statefulset with correct replicas",
			statefulSetTest: core.KubernetesStatefulSetTest{
				Name:        "Database has 3 replicas",
				StatefulSet: "postgres",
				Namespace:   "default",
				State:       "available",
				Replicas:    3,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get statefulset postgres -n default -o json 2>&1", `{"metadata":{"name":"postgres"},"spec":{"replicas":3},"status":{"readyReplicas":3,"currentReplicas":3}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "is available",
		},
		{
			name: "statefulset with wrong replicas",
			statefulSetTest: core.KubernetesStatefulSetTest{
				Name:        "Database replica count",
				StatefulSet: "postgres",
				Namespace:   "default",
				State:       "available",
				Replicas:    5,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get statefulset postgres -n default -o json 2>&1", `{"metadata":{"name":"postgres"},"spec":{"replicas":3},"status":{"readyReplicas":3,"currentReplicas":3}}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "has 3 replicas, expected 5",
		},
		{
			name: "statefulset with correct ready replicas",
			statefulSetTest: core.KubernetesStatefulSetTest{
				Name:          "Database ready replicas",
				StatefulSet:   "postgres",
				Namespace:     "default",
				State:         "exists",
				ReadyReplicas: 2,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get statefulset postgres -n default -o json 2>&1", `{"metadata":{"name":"postgres"},"spec":{"replicas":3},"status":{"readyReplicas":2,"currentReplicas":3}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "exists with 2/3 ready replicas",
		},
		{
			name: "statefulset with wrong ready replicas",
			statefulSetTest: core.KubernetesStatefulSetTest{
				Name:          "Database ready count",
				StatefulSet:   "postgres",
				Namespace:     "default",
				State:         "exists",
				ReadyReplicas: 3,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get statefulset postgres -n default -o json 2>&1", `{"metadata":{"name":"postgres"},"spec":{"replicas":3},"status":{"readyReplicas":1,"currentReplicas":3}}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "has 1 ready replicas, expected 3",
		},
		{
			name: "kubectl command fails",
			statefulSetTest: core.KubernetesStatefulSetTest{
				Name:        "StatefulSet check",
				StatefulSet: "test-sts",
				Namespace:   "default",
				State:       "available",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get statefulset test-sts -n default -o json 2>&1", "", "error: connection refused", 1, nil)
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

			result := executeKubernetesStatefulSetTest(ctx, mock, tt.statefulSetTest)

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
