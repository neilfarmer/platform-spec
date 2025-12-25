package kubernetes

import (
	"context"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// KubernetesPlugin handles all Kubernetes-specific tests
type KubernetesPlugin struct{}

// NewKubernetesPlugin creates a new Kubernetes plugin
func NewKubernetesPlugin() *KubernetesPlugin {
	return &KubernetesPlugin{}
}

// Execute runs all Kubernetes tests
func (p *KubernetesPlugin) Execute(ctx context.Context, spec *core.Spec, provider core.Provider, failFast bool) ([]core.Result, bool) {
	var results []core.Result

	// Execute namespace tests
	for _, test := range spec.Tests.Kubernetes.Namespaces {
		result := executeKubernetesNamespaceTest(ctx, provider, test)
		results = append(results, result)
		if failFast && result.Status == core.StatusFail {
			return results, true
		}
	}

	// Execute pod tests
	for _, test := range spec.Tests.Kubernetes.Pods {
		result := executeKubernetesPodTest(ctx, provider, test)
		results = append(results, result)
		if failFast && result.Status == core.StatusFail {
			return results, true
		}
	}

	// Execute deployment tests
	for _, test := range spec.Tests.Kubernetes.Deployments {
		result := executeKubernetesDeploymentTest(ctx, provider, test)
		results = append(results, result)
		if failFast && result.Status == core.StatusFail {
			return results, true
		}
	}

	// Execute service tests
	for _, test := range spec.Tests.Kubernetes.Services {
		result := executeKubernetesServiceTest(ctx, provider, test)
		results = append(results, result)
		if failFast && result.Status == core.StatusFail {
			return results, true
		}
	}

	// Execute configmap tests
	for _, test := range spec.Tests.Kubernetes.ConfigMaps {
		result := executeKubernetesConfigMapTest(ctx, provider, test)
		results = append(results, result)
		if failFast && result.Status == core.StatusFail {
			return results, true
		}
	}

	// Execute node tests
	for _, test := range spec.Tests.Kubernetes.Nodes {
		result := executeKubernetesNodeTest(ctx, provider, test)
		results = append(results, result)
		if failFast && result.Status == core.StatusFail {
			return results, true
		}
	}

	// Execute CRD tests
	for _, test := range spec.Tests.Kubernetes.CRDs {
		result := executeKubernetesCRDTest(ctx, provider, test)
		results = append(results, result)
		if failFast && result.Status == core.StatusFail {
			return results, true
		}
	}

	// Execute Helm tests
	for _, test := range spec.Tests.Kubernetes.Helm {
		result := executeKubernetesHelmTest(ctx, provider, test)
		results = append(results, result)
		if failFast && result.Status == core.StatusFail {
			return results, true
		}
	}

	// Execute StorageClass tests
	for _, test := range spec.Tests.Kubernetes.StorageClasses {
		result := executeKubernetesStorageClassTest(ctx, provider, test)
		results = append(results, result)
		if failFast && result.Status == core.StatusFail {
			return results, true
		}
	}

	// Execute Secret tests
	for _, test := range spec.Tests.Kubernetes.Secrets {
		result := executeKubernetesSecretTest(ctx, provider, test)
		results = append(results, result)
		if failFast && result.Status == core.StatusFail {
			return results, true
		}
	}

	// Execute Ingress tests
	for _, test := range spec.Tests.Kubernetes.Ingress {
		result := executeKubernetesIngressTest(ctx, provider, test)
		results = append(results, result)
		if failFast && result.Status == core.StatusFail {
			return results, true
		}
	}

	// Execute PVC tests
	for _, test := range spec.Tests.Kubernetes.PVCs {
		result := executeKubernetesPVCTest(ctx, provider, test)
		results = append(results, result)
		if failFast && result.Status == core.StatusFail {
			return results, true
		}
	}

	// Execute StatefulSet tests
	for _, test := range spec.Tests.Kubernetes.StatefulSets {
		result := executeKubernetesStatefulSetTest(ctx, provider, test)
		results = append(results, result)
		if failFast && result.Status == core.StatusFail {
			return results, true
		}
	}

	return results, false
}
