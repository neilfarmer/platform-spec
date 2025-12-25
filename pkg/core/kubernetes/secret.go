package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// executeKubernetesSecretTest executes a Kubernetes secret test
func executeKubernetesSecretTest(ctx context.Context, provider core.Provider, test core.KubernetesSecretTest) core.Result {
	start := time.Now()
	result := core.Result{
		Name:    test.Name,
		Status:  core.StatusPass,
		Details: make(map[string]interface{}),
	}

	// Build kubectl command
	cmd := fmt.Sprintf("kubectl get secret %s -n %s -o json 2>&1", test.Secret, test.Namespace)
	stdout, stderr, exitCode, err := provider.ExecuteCommand(ctx, cmd)

	if err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Error checking secret %s: %v", test.Secret, err)
		result.Duration = time.Since(start)
		return result
	}

	// Handle "not found" case vs actual errors
	exists := (exitCode == 0)
	if exitCode == 1 {
		if strings.Contains(stderr, "not found") {
			exists = false
		} else {
			// exitCode 1 but not "not found" means a real error
			result.Status = core.StatusError
			result.Message = fmt.Sprintf("kubectl error: %s", strings.TrimSpace(stderr))
			result.Duration = time.Since(start)
			return result
		}
	} else if exitCode != 0 {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("kubectl error: %s", strings.TrimSpace(stderr))
		result.Duration = time.Since(start)
		return result
	}

	// Check state
	if test.State == "present" && !exists {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("Secret %s not found in namespace %s", test.Secret, test.Namespace)
		result.Duration = time.Since(start)
		return result
	}

	if test.State == "absent" && exists {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("Secret %s exists but should be absent", test.Secret)
		result.Duration = time.Since(start)
		return result
	}

	// If expecting absent and it is absent, we're done
	if test.State == "absent" {
		result.Message = fmt.Sprintf("Secret %s is absent", test.Secret)
		result.Duration = time.Since(start)
		return result
	}

	result.Details["namespace"] = test.Namespace

	// Parse secret JSON for type and key checks
	var secret map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &secret); err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Failed to parse secret JSON: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// Check type if specified
	if test.Type != "" {
		actualType, _ := getNestedString(secret, "type")
		if actualType != test.Type {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Secret %s has type %s, expected %s", test.Secret, actualType, test.Type)
			result.Duration = time.Since(start)
			return result
		}
		result.Details["type"] = actualType
	}

	// Check keys if specified
	if len(test.HasKeys) > 0 {
		data, _ := getNestedMap(secret, "data")
		var missingKeys []string
		for _, key := range test.HasKeys {
			if _, ok := data[key]; !ok {
				missingKeys = append(missingKeys, key)
			}
		}

		if len(missingKeys) > 0 {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Secret %s missing keys: %s", test.Secret, strings.Join(missingKeys, ", "))
			result.Duration = time.Since(start)
			return result
		}
	}

	// Build success message
	result.Message = fmt.Sprintf("Secret %s exists", test.Secret)
	if test.Type != "" && len(test.HasKeys) > 0 {
		result.Message = fmt.Sprintf("Secret %s exists with correct type and all required keys", test.Secret)
	} else if test.Type != "" {
		result.Message = fmt.Sprintf("Secret %s exists with correct type", test.Secret)
	} else if len(test.HasKeys) > 0 {
		result.Message = fmt.Sprintf("Secret %s exists with all required keys", test.Secret)
	}

	result.Duration = time.Since(start)
	return result
}
