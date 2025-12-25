package kubernetes

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestExecutor_KubernetesNodeTest(t *testing.T) {
	tests := []struct {
		name         string
		nodeTest     core.KubernetesNodeTest
		setupMock    func(*core.MockProvider)
		wantStatus   core.Status
		wantContains string
	}{
		{
			name: "exact count match",
			nodeTest: core.KubernetesNodeTest{
				Name:  "Cluster has 3 nodes",
				Count: 3,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get nodes -o json 2>&1", `{"items":[{"metadata":{"name":"node1"}},{"metadata":{"name":"node2"}},{"metadata":{"name":"node3"}}]}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "3 nodes",
		},
		{
			name: "exact count mismatch",
			nodeTest: core.KubernetesNodeTest{
				Name:  "Cluster has 5 nodes",
				Count: 5,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get nodes -o json 2>&1", `{"items":[{"metadata":{"name":"node1"}},{"metadata":{"name":"node2"}},{"metadata":{"name":"node3"}}]}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "Found 3 nodes, expected exactly 5",
		},
		{
			name: "min count met",
			nodeTest: core.KubernetesNodeTest{
				Name:     "At least 2 nodes",
				MinCount: 2,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get nodes -o json 2>&1", `{"items":[{"metadata":{"name":"node1"}},{"metadata":{"name":"node2"}},{"metadata":{"name":"node3"}}]}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "≥2 nodes",
		},
		{
			name: "min count not met",
			nodeTest: core.KubernetesNodeTest{
				Name:     "At least 5 nodes",
				MinCount: 5,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get nodes -o json 2>&1", `{"items":[{"metadata":{"name":"node1"}},{"metadata":{"name":"node2"}},{"metadata":{"name":"node3"}}]}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "Found 3 nodes, expected at least 5",
		},
		{
			name: "min ready nodes met",
			nodeTest: core.KubernetesNodeTest{
				Name:     "At least 2 ready nodes",
				MinReady: 2,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get nodes -o json 2>&1", `{"items":[
					{"metadata":{"name":"node1"},"status":{"conditions":[{"type":"Ready","status":"True"}]}},
					{"metadata":{"name":"node2"},"status":{"conditions":[{"type":"Ready","status":"True"}]}},
					{"metadata":{"name":"node3"},"status":{"conditions":[{"type":"Ready","status":"False"}]}}
				]}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "≥2 ready",
		},
		{
			name: "min ready nodes not met",
			nodeTest: core.KubernetesNodeTest{
				Name:     "At least 3 ready nodes",
				MinReady: 3,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get nodes -o json 2>&1", `{"items":[
					{"metadata":{"name":"node1"},"status":{"conditions":[{"type":"Ready","status":"True"}]}},
					{"metadata":{"name":"node2"},"status":{"conditions":[{"type":"Ready","status":"False"}]}}
				]}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "Found 1 ready nodes, expected at least 3",
		},
		{
			name: "min version met",
			nodeTest: core.KubernetesNodeTest{
				Name:       "Nodes running v1.28+",
				MinVersion: "v1.28.0",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get nodes -o json 2>&1", `{"items":[
					{"metadata":{"name":"node1"},"status":{"nodeInfo":{"kubeletVersion":"v1.28.2"}}},
					{"metadata":{"name":"node2"},"status":{"nodeInfo":{"kubeletVersion":"v1.29.0"}}}
				]}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "version ≥v1.28.0",
		},
		{
			name: "min version not met",
			nodeTest: core.KubernetesNodeTest{
				Name:       "Nodes running v1.30+",
				MinVersion: "v1.30.0",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get nodes -o json 2>&1", `{"items":[
					{"metadata":{"name":"node1"},"status":{"nodeInfo":{"kubeletVersion":"v1.28.2"}}},
					{"metadata":{"name":"node2"},"status":{"nodeInfo":{"kubeletVersion":"v1.29.0"}}}
				]}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "nodes with version < v1.30.0",
		},
		{
			name: "label filtering",
			nodeTest: core.KubernetesNodeTest{
				Name:  "Worker nodes count",
				Count: 2,
				Labels: map[string]string{
					"node-role.kubernetes.io/worker": "",
				},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get nodes -o json 2>&1", `{"items":[
					{"metadata":{"name":"control","labels":{"node-role.kubernetes.io/control-plane":""}}},
					{"metadata":{"name":"worker1","labels":{"node-role.kubernetes.io/worker":""}}},
					{"metadata":{"name":"worker2","labels":{"node-role.kubernetes.io/worker":""}}}
				]}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "2 nodes",
		},
		{
			name: "combined checks",
			nodeTest: core.KubernetesNodeTest{
				Name:       "Production cluster requirements",
				MinCount:   3,
				MinReady:   3,
				MinVersion: "v1.28.0",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get nodes -o json 2>&1", `{"items":[
					{"metadata":{"name":"node1"},"status":{"conditions":[{"type":"Ready","status":"True"}],"nodeInfo":{"kubeletVersion":"v1.28.2"}}},
					{"metadata":{"name":"node2"},"status":{"conditions":[{"type":"Ready","status":"True"}],"nodeInfo":{"kubeletVersion":"v1.28.2"}}},
					{"metadata":{"name":"node3"},"status":{"conditions":[{"type":"Ready","status":"True"}],"nodeInfo":{"kubeletVersion":"v1.29.0"}}}
				]}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "≥3 nodes, ≥3 ready, version ≥v1.28.0",
		},
		{
			name: "kubectl command fails",
			nodeTest: core.KubernetesNodeTest{
				Name:     "Nodes check",
				MinCount: 1,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get nodes -o json 2>&1", "", "error: connection refused", 1, nil)
			},
			wantStatus:   core.StatusError,
			wantContains: "kubectl error",
		},
		{
			name: "invalid JSON response",
			nodeTest: core.KubernetesNodeTest{
				Name:     "Nodes check",
				MinCount: 1,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get nodes -o json 2>&1", "invalid json", "", 0, nil)
			},
			wantStatus:   core.StatusError,
			wantContains: "Failed to parse nodes JSON",
		},
		{
			name: "no nodes in cluster",
			nodeTest: core.KubernetesNodeTest{
				Name:     "At least 1 node",
				MinCount: 1,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get nodes -o json 2>&1", `{"items":[]}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "Found 0 nodes, expected at least 1",
		},
		{
			name: "node without ready condition",
			nodeTest: core.KubernetesNodeTest{
				Name:     "Ready nodes",
				MinReady: 1,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get nodes -o json 2>&1", `{"items":[
					{"metadata":{"name":"node1"},"status":{"conditions":[{"type":"MemoryPressure","status":"False"}]}}
				]}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "Found 0 ready nodes, expected at least 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := core.NewMockProvider()
			tt.setupMock(mock)

			ctx := context.Background()

			result := executeKubernetesNodeTest(ctx, mock, tt.nodeTest)

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
