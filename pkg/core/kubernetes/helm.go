package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// executeKubernetesHelmTest executes a Kubernetes Helm release test
func executeKubernetesHelmTest(ctx context.Context, provider core.Provider, test core.KubernetesHelmTest) core.Result {
	start := time.Now()
	result := core.Result{
		Name:    test.Name,
		Status:  core.StatusPass,
		Details: make(map[string]interface{}),
	}

	// Check Helm release status
	cmd := fmt.Sprintf("helm list -n %s -o json 2>&1", test.Namespace)
	stdout, stderr, exitCode, err := provider.ExecuteCommand(ctx, cmd)

	if err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Error checking Helm releases: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	if exitCode != 0 {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("helm error: %s", strings.TrimSpace(stderr))
		result.Duration = time.Since(start)
		return result
	}

	// Parse Helm list output
	var releases []map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &releases); err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Failed to parse Helm output: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// Find the release
	var releaseFound bool
	var releaseStatus string
	for _, release := range releases {
		if relName, ok := release["name"].(string); ok && relName == test.Release {
			releaseFound = true
			releaseStatus, _ = release["status"].(string)
			result.Details["release_status"] = releaseStatus
			if chart, ok := release["chart"].(string); ok {
				result.Details["chart"] = chart
			}
			if revision, ok := release["revision"].(float64); ok {
				result.Details["revision"] = int(revision)
			}
			break
		}
	}

	// Check if release exists
	if !releaseFound {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("Helm release %s not found in namespace %s", test.Release, test.Namespace)
		result.Duration = time.Since(start)
		return result
	}

	// Check release state
	if releaseStatus != test.State {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("Helm release %s is %s, expected %s", test.Release, releaseStatus, test.State)
		result.Duration = time.Since(start)
		return result
	}

	// If all_pods_ready is set, check all pods from this release
	if test.AllPodsReady {
		podsReady, podsMessage := checkHelmReleasePods(ctx, provider, test.Release, test.Namespace)
		if !podsReady {
			result.Status = core.StatusFail
			result.Message = podsMessage
			result.Duration = time.Since(start)
			return result
		}
		result.Details["all_pods_ready"] = "true"
	}

	// Build success message
	result.Message = fmt.Sprintf("Helm release %s is %s", test.Release, test.State)
	if test.AllPodsReady {
		result.Message += " with all pods ready"
	}
	result.Duration = time.Since(start)
	return result
}

// checkHelmReleasePods checks if all pods from a Helm release are ready
func checkHelmReleasePods(ctx context.Context, provider core.Provider, release, namespace string) (bool, string) {
	// Query pods with Helm's standard label
	cmd := fmt.Sprintf("kubectl get pods -n %s -l app.kubernetes.io/instance=%s -o json 2>&1", namespace, release)
	stdout, stderr, exitCode, err := provider.ExecuteCommand(ctx, cmd)

	if err != nil {
		return false, fmt.Sprintf("Error checking pods for release %s: %v", release, err)
	}

	if exitCode != 0 {
		return false, fmt.Sprintf("kubectl error checking pods: %s", strings.TrimSpace(stderr))
	}

	// Parse pod list
	var podList map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &podList); err != nil {
		return false, fmt.Sprintf("Failed to parse pod JSON: %v", err)
	}

	items, ok := podList["items"].([]interface{})
	if !ok {
		return false, "Invalid pod JSON: items not found"
	}

	// Check if there are any pods
	if len(items) == 0 {
		return false, fmt.Sprintf("No pods found for Helm release %s", release)
	}

	// Check each pod
	var notReadyPods []string
	var crashLoopPods []string
	totalPods := len(items)
	readyPods := 0

	for _, item := range items {
		pod, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		podName, _ := getNestedString(pod, "metadata", "name")
		phase, _ := getNestedString(pod, "status", "phase")

		// Check for CrashLoopBackOff or other bad states
		containerStatuses, _ := getNestedSlice(pod, "status", "containerStatuses")
		for _, cs := range containerStatuses {
			if csMap, ok := cs.(map[string]interface{}); ok {
				if state, ok := csMap["state"].(map[string]interface{}); ok {
					if waiting, ok := state["waiting"].(map[string]interface{}); ok {
						if reason, ok := waiting["reason"].(string); ok {
							if strings.Contains(reason, "CrashLoopBackOff") || strings.Contains(reason, "ImagePullBackOff") || strings.Contains(reason, "ErrImagePull") {
								crashLoopPods = append(crashLoopPods, fmt.Sprintf("%s (%s)", podName, reason))
							}
						}
					}
				}
			}
		}

		// Check if pod is ready
		if phase == "Running" {
			allReady := true
			for _, cs := range containerStatuses {
				if csMap, ok := cs.(map[string]interface{}); ok {
					if ready, ok := csMap["ready"].(bool); ok && !ready {
						allReady = false
						break
					}
				}
			}
			if allReady && len(containerStatuses) > 0 {
				readyPods++
			} else {
				notReadyPods = append(notReadyPods, podName)
			}
		} else if phase != "Succeeded" {
			// Pods that aren't Running or Succeeded are not ready
			notReadyPods = append(notReadyPods, fmt.Sprintf("%s (%s)", podName, phase))
		}
	}

	// Report issues
	if len(crashLoopPods) > 0 {
		return false, fmt.Sprintf("Helm release %s has pods in bad state: %s", release, strings.Join(crashLoopPods, ", "))
	}

	if len(notReadyPods) > 0 {
		return false, fmt.Sprintf("Helm release %s has %d/%d pods ready, not ready: %s", release, readyPods, totalPods, strings.Join(notReadyPods, ", "))
	}

	return true, ""
}
