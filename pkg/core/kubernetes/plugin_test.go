package kubernetes

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestNewKubernetesPlugin(t *testing.T) {
	plugin := NewKubernetesPlugin()
	if plugin == nil {
		t.Fatal("NewKubernetesPlugin returned nil")
	}
}

func TestKubernetesPlugin_Execute(t *testing.T) {
	mock := core.NewMockProvider()
	// Namespace lookup
	mock.SetCommandResult("kubectl get namespace test-ns -o json 2>&1", `{"metadata":{"name":"test-ns","labels":{"env":"test"}},"status":{"phase":"Active"}}`, "", 0, nil)
	// Pod lookup
	mock.SetCommandResult("kubectl get pod test-pod -n test-ns -o json 2>&1", `{"metadata":{"name":"test-pod","labels":{"app":"test"}},"status":{"phase":"Running","conditions":[{"type":"Ready","status":"True"}]},"spec":{"containers":[{"name":"main","image":"nginx:1.21"}]}}`, "", 0, nil)
	// Deployment lookup
	mock.SetCommandResult("kubectl get deployment test-deploy -n test-ns -o json 2>&1", `{"metadata":{"name":"test-deploy","labels":{"app":"test"}},"status":{"replicas":3,"readyReplicas":3,"availableReplicas":3},"spec":{"replicas":3}}`, "", 0, nil)
	// Service lookup
	mock.SetCommandResult("kubectl get service test-svc -n test-ns -o json 2>&1", `{"metadata":{"name":"test-svc","labels":{"app":"test"}},"spec":{"type":"ClusterIP","clusterIP":"10.0.0.1","ports":[{"port":80}]}}`, "", 0, nil)
	// ConfigMap lookup
	mock.SetCommandResult("kubectl get configmap test-cm -n test-ns -o json 2>&1", `{"metadata":{"name":"test-cm","labels":{"app":"test"}},"data":{"key1":"value1"}}`, "", 0, nil)
	// Node lookup
	mock.SetCommandResult("kubectl get nodes -o json 2>&1", `{"items":[{"metadata":{"name":"test-node","labels":{"node-role.kubernetes.io/control-plane":""}},"status":{"conditions":[{"type":"Ready","status":"True"}],"nodeInfo":{"kubeletVersion":"v1.28.0"}}}]}`, "", 0, nil)
	// CRD lookup
	mock.SetCommandResult("kubectl get crd test.example.com -o json 2>&1", `{"metadata":{"name":"test.example.com"}}`, "", 0, nil)
	// Helm lookup
	mock.SetCommandResult("helm list -n test-ns -o json 2>&1", `[{"name":"test-release","namespace":"test-ns","status":"deployed","chart":"nginx-1.0.0","app_version":"1.21"}]`, "", 0, nil)
	// StorageClass lookup
	mock.SetCommandResult("kubectl get storageclass fast-ssd -o json 2>&1", `{"metadata":{"name":"fast-ssd"},"provisioner":"kubernetes.io/aws-ebs"}`, "", 0, nil)
	// Secret lookup
	mock.SetCommandResult("kubectl get secret test-secret -n test-ns -o json 2>&1", `{"metadata":{"name":"test-secret"},"type":"Opaque","data":{"key":"dmFsdWU="}}`, "", 0, nil)
	// Ingress lookup
	mock.SetCommandResult("kubectl get ingress test-ingress -n test-ns -o json 2>&1", `{"metadata":{"name":"test-ingress"},"spec":{"rules":[{"host":"example.com"}]}}`, "", 0, nil)
	// PVC lookup
	mock.SetCommandResult("kubectl get pvc test-pvc -n test-ns -o json 2>&1", `{"metadata":{"name":"test-pvc"},"status":{"phase":"Bound","capacity":{"storage":"100Gi"}}}`, "", 0, nil)
	// StatefulSet lookup
	mock.SetCommandResult("kubectl get statefulset test-sts -n test-ns -o json 2>&1", `{"metadata":{"name":"test-sts"},"spec":{"replicas":3},"status":{"readyReplicas":3,"currentReplicas":3}}`, "", 0, nil)

	spec := &core.Spec{
		Tests: core.Tests{
			Kubernetes: core.KubernetesTests{
				Namespaces: []core.KubernetesNamespaceTest{
					{Name: "Namespace", Namespace: "test-ns", State: "present"},
				},
				Pods: []core.KubernetesPodTest{
					{Name: "Pod", Pod: "test-pod", Namespace: "test-ns", State: "running"},
				},
				Deployments: []core.KubernetesDeploymentTest{
					{Name: "Deployment", Deployment: "test-deploy", Namespace: "test-ns", State: "ready"},
				},
				Services: []core.KubernetesServiceTest{
					{Name: "Service", Service: "test-svc", Namespace: "test-ns"},
				},
				ConfigMaps: []core.KubernetesConfigMapTest{
					{Name: "ConfigMap", ConfigMap: "test-cm", Namespace: "test-ns", State: "present"},
				},
				Nodes: []core.KubernetesNodeTest{
					{Name: "Node", MinReady: 1},
				},
				CRDs: []core.KubernetesCRDTest{
					{Name: "CRD", CRD: "test.example.com", State: "present"},
				},
				Helm: []core.KubernetesHelmTest{
					{Name: "Helm", Release: "test-release", Namespace: "test-ns", State: "deployed"},
				},
				StorageClasses: []core.KubernetesStorageClassTest{
					{Name: "StorageClass", StorageClass: "fast-ssd", State: "present"},
				},
				Secrets: []core.KubernetesSecretTest{
					{Name: "Secret", Secret: "test-secret", Namespace: "test-ns", State: "present"},
				},
				Ingress: []core.KubernetesIngressTest{
					{Name: "Ingress", Ingress: "test-ingress", Namespace: "test-ns", State: "present"},
				},
				PVCs: []core.KubernetesPVCTest{
					{Name: "PVC", PVC: "test-pvc", Namespace: "test-ns", State: "present"},
				},
				StatefulSets: []core.KubernetesStatefulSetTest{
					{Name: "StatefulSet", StatefulSet: "test-sts", Namespace: "test-ns", State: "available"},
				},
			},
		},
	}

	plugin := NewKubernetesPlugin()
	ctx := context.Background()

	results, shouldStop := plugin.Execute(ctx, spec, mock, false, nil)

	if len(results) != 13 {
		t.Errorf("Expected 13 results, got %d", len(results))
	}

	if shouldStop {
		t.Error("Should not stop when failFast is false")
	}

	for _, r := range results {
		if r.Status != core.StatusPass {
			t.Errorf("Test %q failed: %v", r.Name, r.Message)
		}
	}
}

func TestKubernetesPlugin_Execute_FailFast(t *testing.T) {
	mock := core.NewMockProvider()
	// First namespace test - should fail (exit code 1 means namespace doesn't exist)
	mock.SetCommandResult("kubectl get namespace missing-ns -o json 2>&1", "Error from server (NotFound): namespaces \"missing-ns\" not found", "", 1, nil)
	// Second namespace test - should not be executed due to fail-fast
	mock.SetCommandResult("kubectl get namespace other-ns -o json 2>&1", `{"metadata":{"name":"other-ns"},"status":{"phase":"Active"}}`, "", 0, nil)

	spec := &core.Spec{
		Tests: core.Tests{
			Kubernetes: core.KubernetesTests{
				Namespaces: []core.KubernetesNamespaceTest{
					{Name: "Missing namespace should fail", Namespace: "missing-ns", State: "present"},
					{Name: "Other namespace should not run", Namespace: "other-ns", State: "present"},
				},
			},
		},
	}

	plugin := NewKubernetesPlugin()
	ctx := context.Background()

	results, shouldStop := plugin.Execute(ctx, spec, mock, true, nil)

	if len(results) != 1 {
		t.Errorf("Expected 1 result (fail fast), got %d", len(results))
		for i, r := range results {
			t.Logf("Result %d: %s - %s (status: %v)", i, r.Name, r.Message, r.Status)
		}
	}

	if !shouldStop {
		t.Error("Should stop when failFast is true and test fails")
	}

	if len(results) > 0 && results[0].Status != core.StatusFail {
		t.Errorf("Expected failure, got %v for test %q: %s", results[0].Status, results[0].Name, results[0].Message)
	}
}
