package kubernetes

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestExecutor_KubernetesStorageClassTest(t *testing.T) {
	tests := []struct {
		name         string
		scTest       core.KubernetesStorageClassTest
		setupMock    func(*core.MockProvider)
		wantStatus   core.Status
		wantContains string
	}{
		{
			name: "storageclass exists",
			scTest: core.KubernetesStorageClassTest{
				Name:         "Fast SSD storage exists",
				StorageClass: "fast-ssd",
				State:        "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get storageclass fast-ssd -o json 2>&1", `{"metadata":{"name":"fast-ssd"},"provisioner":"kubernetes.io/gce-pd"}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "StorageClass fast-ssd exists",
		},
		{
			name: "storageclass does not exist",
			scTest: core.KubernetesStorageClassTest{
				Name:         "Storage should exist",
				StorageClass: "missing-storage",
				State:        "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get storageclass missing-storage -o json 2>&1", "", "Error from server (NotFound): storageclasses.storage.k8s.io \"missing-storage\" not found", 1, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "StorageClass missing-storage not found",
		},
		{
			name: "storageclass absent as expected",
			scTest: core.KubernetesStorageClassTest{
				Name:         "Old storage removed",
				StorageClass: "deprecated-storage",
				State:        "absent",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get storageclass deprecated-storage -o json 2>&1", "", "Error from server (NotFound): storageclasses.storage.k8s.io \"deprecated-storage\" not found", 1, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "StorageClass deprecated-storage is absent",
		},
		{
			name: "storageclass exists but should be absent",
			scTest: core.KubernetesStorageClassTest{
				Name:         "Storage should not exist",
				StorageClass: "fast-ssd",
				State:        "absent",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get storageclass fast-ssd -o json 2>&1", `{"metadata":{"name":"fast-ssd"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "exists but should be absent",
		},
		{
			name: "kubectl command fails",
			scTest: core.KubernetesStorageClassTest{
				Name:         "Storage check",
				StorageClass: "test-storage",
				State:        "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get storageclass test-storage -o json 2>&1", "", "error: connection refused", 1, nil)
			},
			wantStatus:   core.StatusError,
			wantContains: "kubectl error",
		},
		{
			name: "standard storageclass",
			scTest: core.KubernetesStorageClassTest{
				Name:         "Standard storage exists",
				StorageClass: "standard",
				State:        "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get storageclass standard -o json 2>&1", `{"metadata":{"name":"standard"},"provisioner":"kubernetes.io/aws-ebs"}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "StorageClass standard exists",
		},
		{
			name: "gp2 storageclass (AWS)",
			scTest: core.KubernetesStorageClassTest{
				Name:         "GP2 storage exists",
				StorageClass: "gp2",
				State:        "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get storageclass gp2 -o json 2>&1", `{"metadata":{"name":"gp2"},"provisioner":"kubernetes.io/aws-ebs","parameters":{"type":"gp2"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "StorageClass gp2 exists",
		},
		{
			name: "local-path storageclass (k3s/kind)",
			scTest: core.KubernetesStorageClassTest{
				Name:         "Local path storage exists",
				StorageClass: "local-path",
				State:        "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get storageclass local-path -o json 2>&1", `{"metadata":{"name":"local-path"},"provisioner":"rancher.io/local-path"}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "StorageClass local-path exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := core.NewMockProvider()
			tt.setupMock(mock)

			ctx := context.Background()

			result := executeKubernetesStorageClassTest(ctx, mock, tt.scTest)

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
