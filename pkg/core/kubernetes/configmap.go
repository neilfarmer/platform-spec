package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// executeKubernetesConfigMapTest executes a Kubernetes configmap test
func executeKubernetesConfigMapTest(ctx context.Context, provider core.Provider, test core.KubernetesConfigMapTest) core.Result {
	start := time.Now()
	result := core.Result{
		Name:    test.Name,
		Status:  core.StatusPass,
		Details: make(map[string]interface{}),
	}

	// Build kubectl command
	cmd := fmt.Sprintf("kubectl get configmap %s -n %s -o json 2>&1", test.ConfigMap, test.Namespace)
	stdout, stderr, exitCode, err := provider.ExecuteCommand(ctx, cmd)

	if err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Error checking configmap %s: %v", test.ConfigMap, err)
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
		result.Message = fmt.Sprintf("ConfigMap %s not found in namespace %s", test.ConfigMap, test.Namespace)
		result.Duration = time.Since(start)
		return result
	}

	if test.State == "absent" && exists {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("ConfigMap %s exists but should be absent", test.ConfigMap)
		result.Duration = time.Since(start)
		return result
	}

	// If expecting absent and it is absent, we're done
	if test.State == "absent" {
		result.Message = fmt.Sprintf("ConfigMap %s is absent", test.ConfigMap)
		result.Duration = time.Since(start)
		return result
	}

	result.Details["namespace"] = test.Namespace

	// Check keys if specified
	if len(test.HasKeys) > 0 {
		var configmap map[string]interface{}
		if err := json.Unmarshal([]byte(stdout), &configmap); err != nil {
			result.Status = core.StatusError
			result.Message = fmt.Sprintf("Failed to parse configmap JSON: %v", err)
			result.Duration = time.Since(start)
			return result
		}

		data, _ := getNestedMap(configmap, "data")
		var missingKeys []string
		for _, key := range test.HasKeys {
			if _, ok := data[key]; !ok {
				missingKeys = append(missingKeys, key)
			}
		}

		if len(missingKeys) > 0 {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("ConfigMap %s missing keys: %s", test.ConfigMap, strings.Join(missingKeys, ", "))
			result.Duration = time.Since(start)
			return result
		}
	}

	result.Message = fmt.Sprintf("ConfigMap %s exists", test.ConfigMap)
	if len(test.HasKeys) > 0 {
		result.Message = fmt.Sprintf("ConfigMap %s exists with all required keys", test.ConfigMap)
	}
	result.Duration = time.Since(start)
	return result
}
