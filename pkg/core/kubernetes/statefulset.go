package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// executeKubernetesStatefulSetTest executes a Kubernetes StatefulSet test
func executeKubernetesStatefulSetTest(ctx context.Context, provider core.Provider, test core.KubernetesStatefulSetTest) core.Result {
	start := time.Now()
	result := core.Result{
		Name:    test.Name,
		Status:  core.StatusPass,
		Details: make(map[string]interface{}),
	}

	// Build kubectl command
	cmd := fmt.Sprintf("kubectl get statefulset %s -n %s -o json 2>&1", test.StatefulSet, test.Namespace)
	stdout, stderr, exitCode, err := provider.ExecuteCommand(ctx, cmd)

	if err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Error checking StatefulSet %s: %v", test.StatefulSet, err)
		result.Duration = time.Since(start)
		return result
	}

	// Handle "not found" case
	if exitCode != 0 {
		if strings.Contains(stderr, "not found") {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("StatefulSet %s not found in namespace %s", test.StatefulSet, test.Namespace)
		} else {
			result.Status = core.StatusError
			result.Message = fmt.Sprintf("kubectl error: %s", strings.TrimSpace(stderr))
		}
		result.Duration = time.Since(start)
		return result
	}

	result.Details["namespace"] = test.Namespace

	// Parse StatefulSet JSON
	var statefulSet map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &statefulSet); err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Failed to parse StatefulSet JSON: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// Get replica counts
	desiredReplicas, _ := getNestedFloat64(statefulSet, "spec", "replicas")
	readyReplicas, _ := getNestedFloat64(statefulSet, "status", "readyReplicas")
	currentReplicas, _ := getNestedFloat64(statefulSet, "status", "currentReplicas")

	result.Details["desired_replicas"] = int(desiredReplicas)
	result.Details["ready_replicas"] = int(readyReplicas)
	result.Details["current_replicas"] = int(currentReplicas)

	// Check state
	if test.State == "available" {
		// All replicas must be ready
		if int(readyReplicas) != int(desiredReplicas) || desiredReplicas == 0 {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("StatefulSet %s is not available (ready: %d, desired: %d)",
				test.StatefulSet, int(readyReplicas), int(desiredReplicas))
			result.Duration = time.Since(start)
			return result
		}
	}

	// Check exact replica count if specified
	if test.Replicas > 0 {
		if int(desiredReplicas) != test.Replicas {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("StatefulSet %s has %d replicas, expected %d",
				test.StatefulSet, int(desiredReplicas), test.Replicas)
			result.Duration = time.Since(start)
			return result
		}
	}

	// Check exact ready replica count if specified
	if test.ReadyReplicas > 0 {
		if int(readyReplicas) != test.ReadyReplicas {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("StatefulSet %s has %d ready replicas, expected %d",
				test.StatefulSet, int(readyReplicas), test.ReadyReplicas)
			result.Duration = time.Since(start)
			return result
		}
	}

	// Build success message
	if test.State == "available" {
		result.Message = fmt.Sprintf("StatefulSet %s is available with %d/%d ready replicas",
			test.StatefulSet, int(readyReplicas), int(desiredReplicas))
	} else {
		result.Message = fmt.Sprintf("StatefulSet %s exists with %d/%d ready replicas",
			test.StatefulSet, int(readyReplicas), int(desiredReplicas))
	}

	result.Duration = time.Since(start)
	return result
}
