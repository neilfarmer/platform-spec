package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// executeKubernetesDeploymentTest executes a Kubernetes deployment test
func executeKubernetesDeploymentTest(ctx context.Context, provider core.Provider, test core.KubernetesDeploymentTest) core.Result {
	start := time.Now()
	result := core.Result{
		Name:    test.Name,
		Status:  core.StatusPass,
		Details: make(map[string]interface{}),
	}

	// Build kubectl command
	cmd := fmt.Sprintf("kubectl get deployment %s -n %s -o json 2>&1", test.Deployment, test.Namespace)
	stdout, stderr, exitCode, err := provider.ExecuteCommand(ctx, cmd)

	if err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Error checking deployment %s: %v", test.Deployment, err)
		result.Duration = time.Since(start)
		return result
	}

	// Handle not found
	if exitCode == 1 && strings.Contains(stderr, "not found") {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("Deployment %s not found in namespace %s", test.Deployment, test.Namespace)
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
	var deployment map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &deployment); err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Failed to parse deployment JSON: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	result.Details["namespace"] = test.Namespace

	// Check state if not "exists"
	if test.State != "exists" {
		conditions, _ := getNestedSlice(deployment, "status", "conditions")
		isAvailable := false
		isProgressing := false

		for _, cond := range conditions {
			if condMap, ok := cond.(map[string]interface{}); ok {
				condType, _ := condMap["type"].(string)
				condStatus, _ := condMap["status"].(string)

				if condType == "Available" && condStatus == "True" {
					isAvailable = true
				}
				if condType == "Progressing" && condStatus == "True" {
					isProgressing = true
				}
			}
		}

		if test.State == "available" && !isAvailable {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Deployment %s is not available", test.Deployment)
			result.Duration = time.Since(start)
			return result
		}

		if test.State == "progressing" && !isProgressing {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Deployment %s is not progressing", test.Deployment)
			result.Duration = time.Since(start)
			return result
		}

		result.Details["available"] = isAvailable
		result.Details["progressing"] = isProgressing
	}

	// Check replicas if specified
	if test.Replicas > 0 {
		replicas, _ := getNestedFloat64(deployment, "spec", "replicas")
		if int(replicas) != test.Replicas {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Deployment %s has %d replicas, expected %d", test.Deployment, int(replicas), test.Replicas)
			result.Duration = time.Since(start)
			return result
		}
		result.Details["replicas"] = test.Replicas
	}

	// Check ready replicas if specified
	if test.ReadyReplicas > 0 {
		readyReplicas, _ := getNestedFloat64(deployment, "status", "readyReplicas")
		if int(readyReplicas) != test.ReadyReplicas {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Deployment %s has %d ready replicas, expected %d", test.Deployment, int(readyReplicas), test.ReadyReplicas)
			result.Duration = time.Since(start)
			return result
		}
		result.Details["readyReplicas"] = test.ReadyReplicas
	}

	// Check image if specified
	if test.Image != "" {
		containers, _ := getNestedSlice(deployment, "spec", "template", "spec", "containers")
		imageFound := false
		for _, c := range containers {
			if cMap, ok := c.(map[string]interface{}); ok {
				if image, ok := cMap["image"].(string); ok && strings.Contains(image, test.Image) {
					imageFound = true
					result.Details["image"] = image
					break
				}
			}
		}
		if !imageFound {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Deployment %s does not contain image %s", test.Deployment, test.Image)
			result.Duration = time.Since(start)
			return result
		}
	}

	result.Message = fmt.Sprintf("Deployment %s is %s", test.Deployment, test.State)
	result.Duration = time.Since(start)
	return result
}
