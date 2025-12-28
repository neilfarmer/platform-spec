package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// executeKubernetesPodTest executes a Kubernetes pod test
func executeKubernetesPodTest(ctx context.Context, provider core.Provider, test core.KubernetesPodTest) core.Result {
	start := time.Now()
	result := core.Result{
		Name:    test.Name,
		Status:  core.StatusPass,
		Details: make(map[string]interface{}),
	}

	// Build kubectl command
	cmd := fmt.Sprintf("kubectl get pod %s -n %s -o json 2>&1", test.Pod, test.Namespace)
	stdout, stderr, exitCode, err := provider.ExecuteCommand(ctx, cmd)

	if err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Error checking pod %s: %v", test.Pod, err)
		result.Duration = time.Since(start)
		return result
	}

	// Handle not found (check both stdout and stderr since we use 2>&1)
	combinedOutput := stdout + stderr
	if exitCode == 1 && (strings.Contains(combinedOutput, "not found") || strings.Contains(combinedOutput, "NotFound")) {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("Pod %s not found in namespace %s", test.Pod, test.Namespace)
		result.Duration = time.Since(start)
		return result
	}

	if exitCode != 0 {
		result.Status = core.StatusError
		errorMsg := strings.TrimSpace(combinedOutput)
		if errorMsg == "" {
			errorMsg = "unknown error"
		}
		result.Message = fmt.Sprintf("kubectl error: %s", errorMsg)
		result.Details["command"] = cmd
		result.Duration = time.Since(start)
		return result
	}

	// Parse JSON
	var pod map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &pod); err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Failed to parse pod JSON: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// Extract status phase
	phase, _ := getNestedString(pod, "status", "phase")
	result.Details["phase"] = phase
	result.Details["namespace"] = test.Namespace
	result.Details["command"] = cmd

	// Check state
	if test.State != "exists" {
		expectedPhase := strings.Title(strings.ToLower(test.State)) // "running" -> "Running"
		if phase != expectedPhase {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Pod %s phase is %s, expected %s", test.Pod, phase, expectedPhase)
			result.Duration = time.Since(start)
			return result
		}
	}

	// Check ready if specified
	if test.Ready {
		containerStatuses, _ := getNestedSlice(pod, "status", "containerStatuses")
		allReady := true
		if len(containerStatuses) == 0 {
			allReady = false
		}
		for _, cs := range containerStatuses {
			if csMap, ok := cs.(map[string]interface{}); ok {
				if ready, ok := csMap["ready"].(bool); ok && !ready {
					allReady = false
					break
				}
			}
		}
		if !allReady {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Pod %s containers not all ready", test.Pod)
			result.Duration = time.Since(start)
			return result
		}
		result.Details["ready"] = "true"
	}

	// Check image if specified
	if test.Image != "" {
		containers, _ := getNestedSlice(pod, "spec", "containers")
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
			result.Message = fmt.Sprintf("Pod %s does not contain image %s", test.Pod, test.Image)
			result.Duration = time.Since(start)
			return result
		}
	}

	// Check labels if specified
	if len(test.Labels) > 0 {
		labels, _ := getNestedMap(pod, "metadata", "labels")
		for key, expectedVal := range test.Labels {
			if actualVal, ok := labels[key]; !ok || actualVal != expectedVal {
				result.Status = core.StatusFail
				result.Message = fmt.Sprintf("Pod %s label %s=%s not found (actual: %v)", test.Pod, key, expectedVal, actualVal)
				result.Duration = time.Since(start)
				return result
			}
		}
	}

	result.Message = fmt.Sprintf("Pod %s is %s", test.Pod, strings.ToLower(phase))
	result.Duration = time.Since(start)
	return result
}
