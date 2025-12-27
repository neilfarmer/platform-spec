package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// executeKubernetesServiceTest executes a Kubernetes service test
func executeKubernetesServiceTest(ctx context.Context, provider core.Provider, test core.KubernetesServiceTest) core.Result {
	start := time.Now()
	result := core.Result{
		Name:    test.Name,
		Status:  core.StatusPass,
		Details: make(map[string]interface{}),
	}

	// Build kubectl command
	cmd := fmt.Sprintf("kubectl get service %s -n %s -o json 2>&1", test.Service, test.Namespace)
	stdout, stderr, exitCode, err := provider.ExecuteCommand(ctx, cmd)

	if err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Error checking service %s: %v", test.Service, err)
		result.Duration = time.Since(start)
		return result
	}

	// Handle not found
	if exitCode == 1 && strings.Contains(stderr, "not found") {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("Service %s not found in namespace %s", test.Service, test.Namespace)
		result.Duration = time.Since(start)
		return result
	}

	if exitCode != 0 {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("kubectl error: %s", strings.TrimSpace(stderr))
		result.Duration = time.Since(start)
		return result
	}

	// Parse JSON
	var service map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &service); err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Failed to parse service JSON: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	result.Details["namespace"] = test.Namespace

	// Check service type if specified
	if test.Type != "" {
		svcType, _ := getNestedString(service, "spec", "type")
		if svcType != test.Type {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Service %s type is %s, expected %s", test.Service, svcType, test.Type)
			result.Duration = time.Since(start)
			return result
		}
		result.Details["type"] = svcType
	}

	// Check ports if specified
	if len(test.Ports) > 0 {
		servicePorts, _ := getNestedSlice(service, "spec", "ports")

		for _, expectedPort := range test.Ports {
			portFound := false
			for _, sp := range servicePorts {
				if spMap, ok := sp.(map[string]interface{}); ok {
					port, _ := spMap["port"].(float64)
					protocol, _ := spMap["protocol"].(string)

					if int(port) == expectedPort.Port && protocol == expectedPort.Protocol {
						portFound = true
						break
					}
				}
			}
			if !portFound {
				result.Status = core.StatusFail
				result.Message = fmt.Sprintf("Service %s does not have port %d/%s", test.Service, expectedPort.Port, expectedPort.Protocol)
				result.Duration = time.Since(start)
				return result
			}
		}
	}

	// Check selector if specified
	if len(test.Selector) > 0 {
		selector, _ := getNestedMap(service, "spec", "selector")
		for key, expectedVal := range test.Selector {
			if actualVal, ok := selector[key]; !ok || actualVal != expectedVal {
				result.Status = core.StatusFail
				result.Message = fmt.Sprintf("Service %s selector %s=%s not found (actual: %v)", test.Service, key, expectedVal, actualVal)
				result.Duration = time.Since(start)
				return result
			}
		}
	}

	result.Message = fmt.Sprintf("Service %s exists", test.Service)
	if test.Type != "" {
		result.Message = fmt.Sprintf("Service %s is type %s", test.Service, test.Type)
	}
	result.Duration = time.Since(start)
	return result
}
