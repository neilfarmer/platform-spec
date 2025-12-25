package kubernetes

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestExecutor_KubernetesCRDTest(t *testing.T) {
	tests := []struct {
		name         string
		crdTest      core.KubernetesCRDTest
		setupMock    func(*core.MockProvider)
		wantStatus   core.Status
		wantContains string
	}{
		{
			name: "crd exists",
			crdTest: core.KubernetesCRDTest{
				Name:  "Cert-manager CRD exists",
				CRD:   "certificates.cert-manager.io",
				State: "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get crd certificates.cert-manager.io -o json 2>&1", `{"metadata":{"name":"certificates.cert-manager.io"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "CRD certificates.cert-manager.io exists",
		},
		{
			name: "crd does not exist",
			crdTest: core.KubernetesCRDTest{
				Name:  "CRD should exist",
				CRD:   "missing.example.com",
				State: "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get crd missing.example.com -o json 2>&1", "", "Error from server (NotFound): customresourcedefinitions.apiextensions.k8s.io \"missing.example.com\" not found", 1, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "CRD missing.example.com not found",
		},
		{
			name: "crd absent as expected",
			crdTest: core.KubernetesCRDTest{
				Name:  "CRD should be absent",
				CRD:   "deleted.example.com",
				State: "absent",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get crd deleted.example.com -o json 2>&1", "", "Error from server (NotFound): customresourcedefinitions.apiextensions.k8s.io \"deleted.example.com\" not found", 1, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "CRD deleted.example.com is absent",
		},
		{
			name: "crd exists but should be absent",
			crdTest: core.KubernetesCRDTest{
				Name:  "CRD should not exist",
				CRD:   "certificates.cert-manager.io",
				State: "absent",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get crd certificates.cert-manager.io -o json 2>&1", `{"metadata":{"name":"certificates.cert-manager.io"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "exists but should be absent",
		},
		{
			name: "kubectl command fails",
			crdTest: core.KubernetesCRDTest{
				Name:  "CRD check",
				CRD:   "test.example.com",
				State: "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get crd test.example.com -o json 2>&1", "", "error: connection refused", 1, nil)
			},
			wantStatus:   core.StatusError,
			wantContains: "kubectl error",
		},
		{
			name: "common cert-manager CRDs",
			crdTest: core.KubernetesCRDTest{
				Name:  "Certificate CRD exists",
				CRD:   "certificates.cert-manager.io",
				State: "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get crd certificates.cert-manager.io -o json 2>&1", `{"metadata":{"name":"certificates.cert-manager.io"},"spec":{"group":"cert-manager.io"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "CRD certificates.cert-manager.io exists",
		},
		{
			name: "istio CRD exists",
			crdTest: core.KubernetesCRDTest{
				Name:  "VirtualService CRD exists",
				CRD:   "virtualservices.networking.istio.io",
				State: "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get crd virtualservices.networking.istio.io -o json 2>&1", `{"metadata":{"name":"virtualservices.networking.istio.io"},"spec":{"group":"networking.istio.io"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "CRD virtualservices.networking.istio.io exists",
		},
		{
			name: "prometheus operator CRD",
			crdTest: core.KubernetesCRDTest{
				Name:  "ServiceMonitor CRD exists",
				CRD:   "servicemonitors.monitoring.coreos.com",
				State: "present",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("kubectl get crd servicemonitors.monitoring.coreos.com -o json 2>&1", `{"metadata":{"name":"servicemonitors.monitoring.coreos.com"},"spec":{"group":"monitoring.coreos.com"}}`, "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "CRD servicemonitors.monitoring.coreos.com exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := core.NewMockProvider()
			tt.setupMock(mock)

			ctx := context.Background()

			result := executeKubernetesCRDTest(ctx, mock, tt.crdTest)

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
