package kubernetes

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestExecutor_KubernetesPVCTest(t *testing.T) {
	tests := []struct {
		name         string
		pvcTest      core.KubernetesPVCTest
		setupMock    func(*core.MockProvider)
		wantStatus   core.Status
		wantContains string
	}{
		{
			name: "pvc exists",
			pvcTest: core.KubernetesPVCTest{
				Name:      "Database volume exists",
				PVC:       "postgres-data",
				Namespace: "default",
				State:     "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get pvc postgres-data -n default -o json 2>&1", `{"metadata":{"name":"postgres-data"},"status":{"phase":"Bound"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "PVC postgres-data exists",
		},
		{
			name: "pvc not found",
			pvcTest: core.KubernetesPVCTest{
				Name:      "PVC exists",
				PVC:       "missing-pvc",
				Namespace: "default",
				State:     "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get pvc missing-pvc -n default -o json 2>&1", "", "Error from server (NotFound): persistentvolumeclaims \"missing-pvc\" not found", 1, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "PVC missing-pvc not found",
		},
		{
			name: "pvc absent as expected",
			pvcTest: core.KubernetesPVCTest{
				Name:      "Old PVC removed",
				PVC:       "old-pvc",
				Namespace: "default",
				State:     "absent",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get pvc old-pvc -n default -o json 2>&1", "", "Error from server (NotFound): persistentvolumeclaims \"old-pvc\" not found", 1, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "PVC old-pvc is absent",
		},
		{
			name: "pvc with correct status",
			pvcTest: core.KubernetesPVCTest{
				Name:      "PVC bound",
				PVC:       "data-pvc",
				Namespace: "default",
				State:     "present",
				Status:    "Bound",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get pvc data-pvc -n default -o json 2>&1", `{"metadata":{"name":"data-pvc"},"status":{"phase":"Bound"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "correct status",
		},
		{
			name: "pvc with wrong status",
			pvcTest: core.KubernetesPVCTest{
				Name:      "PVC status check",
				PVC:       "data-pvc",
				Namespace: "default",
				State:     "present",
				Status:    "Bound",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get pvc data-pvc -n default -o json 2>&1", `{"metadata":{"name":"data-pvc"},"status":{"phase":"Pending"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "has status Pending, expected Bound",
		},
		{
			name: "pvc with correct storage class",
			pvcTest: core.KubernetesPVCTest{
				Name:         "PVC storage class",
				PVC:          "data-pvc",
				Namespace:    "default",
				State:        "present",
				StorageClass: "fast-ssd",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get pvc data-pvc -n default -o json 2>&1", `{"metadata":{"name":"data-pvc"},"spec":{"storageClassName":"fast-ssd"},"status":{"phase":"Bound"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "correct storage class",
		},
		{
			name: "pvc with wrong storage class",
			pvcTest: core.KubernetesPVCTest{
				Name:         "PVC storage class mismatch",
				PVC:          "data-pvc",
				Namespace:    "default",
				State:        "present",
				StorageClass: "fast-ssd",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get pvc data-pvc -n default -o json 2>&1", `{"metadata":{"name":"data-pvc"},"spec":{"storageClassName":"standard"},"status":{"phase":"Bound"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "has storage class standard, expected fast-ssd",
		},
		{
			name: "pvc with sufficient capacity",
			pvcTest: core.KubernetesPVCTest{
				Name:        "PVC capacity check",
				PVC:         "data-pvc",
				Namespace:   "default",
				State:       "present",
				MinCapacity: "100Gi",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get pvc data-pvc -n default -o json 2>&1", `{"metadata":{"name":"data-pvc"},"status":{"phase":"Bound","capacity":{"storage":"200Gi"}}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "sufficient capacity",
		},
		{
			name: "pvc with insufficient capacity",
			pvcTest: core.KubernetesPVCTest{
				Name:        "PVC capacity insufficient",
				PVC:         "data-pvc",
				Namespace:   "default",
				State:       "present",
				MinCapacity: "100Gi",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get pvc data-pvc -n default -o json 2>&1", `{"metadata":{"name":"data-pvc"},"status":{"phase":"Bound","capacity":{"storage":"50Gi"}}}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "has capacity 50Gi, minimum required 100Gi",
		},
		{
			name: "pvc pending without capacity",
			pvcTest: core.KubernetesPVCTest{
				Name:        "PVC not bound yet",
				PVC:         "data-pvc",
				Namespace:   "default",
				State:       "present",
				MinCapacity: "100Gi",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get pvc data-pvc -n default -o json 2>&1", `{"metadata":{"name":"data-pvc"},"status":{"phase":"Pending"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "does not have capacity information",
		},
		{
			name: "kubectl command fails",
			pvcTest: core.KubernetesPVCTest{
				Name:      "PVC check",
				PVC:       "test-pvc",
				Namespace: "default",
				State:     "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get pvc test-pvc -n default -o json 2>&1", "", "error: connection refused", 1, nil)
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

			result := executeKubernetesPVCTest(ctx, mock, tt.pvcTest)

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

func TestParseStorageSize(t *testing.T) {
	tests := []struct {
		name      string
		size      string
		wantBytes int64
		wantErr   bool
	}{
		// Binary units (IEC)
		{"1Ki", "1Ki", 1024, false},
		{"500Mi", "500Mi", 500 * 1024 * 1024, false},
		{"100Gi", "100Gi", 100 * 1024 * 1024 * 1024, false},
		{"1Ti", "1Ti", 1024 * 1024 * 1024 * 1024, false},
		{"1Pi", "1Pi", 1024 * 1024 * 1024 * 1024 * 1024, false},
		{"1Ei", "1Ei", 1024 * 1024 * 1024 * 1024 * 1024 * 1024, false},

		// Decimal units (SI)
		{"1K", "1K", 1000, false},
		{"1M", "1M", 1000 * 1000, false},
		{"1G", "1G", 1000 * 1000 * 1000, false},
		{"1T", "1T", 1000 * 1000 * 1000 * 1000, false},
		{"1P", "1P", 1000 * 1000 * 1000 * 1000 * 1000, false},
		{"1E", "1E", 1000 * 1000 * 1000 * 1000 * 1000 * 1000, false},

		// No unit (bytes)
		{"100", "100", 100, false},
		{"1024", "1024", 1024, false},

		// Decimal values
		{"1.5Gi", "1.5Gi", int64(1.5 * 1024 * 1024 * 1024), false},
		{"0.5Ti", "0.5Ti", int64(0.5 * 1024 * 1024 * 1024 * 1024), false},

		// Error cases
		{"invalid", "invalid", 0, true},
		{"", "", 0, true},
		{"100XYZ", "100XYZ", 0, true},
		{"-100Gi", "-100Gi", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBytes, err := parseStorageSize(tt.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseStorageSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotBytes != tt.wantBytes {
				t.Errorf("parseStorageSize() = %v, want %v", gotBytes, tt.wantBytes)
			}
		})
	}
}
