package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// executeKubernetesNamespaceTest executes a Kubernetes namespace test
func executeKubernetesNamespaceTest(ctx context.Context, provider core.Provider, test core.KubernetesNamespaceTest) core.Result {
	start := time.Now()
	result := core.Result{
		Name:    test.Name,
		Status:  core.StatusPass,
		Details: make(map[string]interface{}),
	}

	// Build kubectl command
	cmd := fmt.Sprintf("kubectl get namespace %s -o json 2>&1", test.Namespace)
	stdout, stderr, exitCode, err := provider.ExecuteCommand(ctx, cmd)

	if err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Error checking namespace %s: %v", test.Namespace, err)
		result.Duration = time.Since(start)
		return result
	}

	// Handle "not found" case
	exists := (exitCode == 0)
	if exitCode == 1 && strings.Contains(stderr, "not found") {
		exists = false
	} else if exitCode != 0 && exitCode != 1 {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("kubectl error: %s", strings.TrimSpace(stderr))
		result.Duration = time.Since(start)
		return result
	}

	// Check state
	if test.State == "present" && !exists {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("Namespace %s not found", test.Namespace)
		result.Duration = time.Since(start)
		return result
	}

	if test.State == "absent" && exists {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("Namespace %s exists but should be absent", test.Namespace)
		result.Duration = time.Since(start)
		return result
	}

	// If expecting absent and it is absent, we're done
	if test.State == "absent" {
		result.Message = fmt.Sprintf("Namespace %s is absent", test.Namespace)
		result.Duration = time.Since(start)
		return result
	}

	// Parse JSON for label validation
	if len(test.Labels) > 0 {
		var namespace map[string]interface{}
		if err := json.Unmarshal([]byte(stdout), &namespace); err != nil {
			result.Status = core.StatusError
			result.Message = fmt.Sprintf("Failed to parse namespace JSON: %v", err)
			result.Duration = time.Since(start)
			return result
		}

		labels, _ := getNestedMap(namespace, "metadata", "labels")
		for key, expectedVal := range test.Labels {
			if actualVal, ok := labels[key]; !ok || actualVal != expectedVal {
				result.Status = core.StatusFail
				result.Message = fmt.Sprintf("Namespace %s label %s=%s not found (actual: %v)", test.Namespace, key, expectedVal, actualVal)
				result.Duration = time.Since(start)
				return result
			}
		}
	}

	result.Message = fmt.Sprintf("Namespace %s exists", test.Namespace)
	result.Duration = time.Since(start)
	return result
}
