package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// executeKubernetesNodeTest executes a Kubernetes node test
func executeKubernetesNodeTest(ctx context.Context, provider core.Provider, test core.KubernetesNodeTest) core.Result {
	start := time.Now()
	result := core.Result{
		Name:    test.Name,
		Status:  core.StatusPass,
		Details: make(map[string]interface{}),
	}

	// Build kubectl command
	cmd := "kubectl get nodes -o json 2>&1"
	stdout, stderr, exitCode, err := provider.ExecuteCommand(ctx, cmd)

	if err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Error getting nodes: %v", err)
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
	var nodeList map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &nodeList); err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Failed to parse nodes JSON: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// Extract items
	items, ok := nodeList["items"].([]interface{})
	if !ok {
		result.Status = core.StatusError
		result.Message = "Invalid nodes JSON: items not found"
		result.Duration = time.Since(start)
		return result
	}

	// Filter nodes by labels if specified
	var filteredNodes []interface{}
	for _, item := range items {
		node, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// Check labels if specified
		if len(test.Labels) > 0 {
			nodeLabels, _ := getNestedMap(node, "metadata", "labels")
			matchesLabels := true
			for key, expectedVal := range test.Labels {
				if actualVal, ok := nodeLabels[key]; !ok || actualVal != expectedVal {
					matchesLabels = false
					break
				}
			}
			if !matchesLabels {
				continue
			}
		}

		filteredNodes = append(filteredNodes, node)
	}

	totalCount := len(filteredNodes)
	result.Details["total_nodes"] = totalCount

	// Count ready nodes
	readyCount := 0
	var versions []string
	for _, item := range filteredNodes {
		node, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// Check if node is ready
		if isNodeReady(node) {
			readyCount++
		}

		// Extract version
		version, _ := getNestedString(node, "status", "nodeInfo", "kubeletVersion")
		if version != "" {
			versions = append(versions, version)
		}
	}
	result.Details["ready_nodes"] = readyCount

	// Check count if specified
	if test.Count > 0 {
		if totalCount != test.Count {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Found %d nodes, expected exactly %d", totalCount, test.Count)
			result.Duration = time.Since(start)
			return result
		}
	}

	// Check min_count if specified
	if test.MinCount > 0 {
		if totalCount < test.MinCount {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Found %d nodes, expected at least %d", totalCount, test.MinCount)
			result.Duration = time.Since(start)
			return result
		}
	}

	// Check min_ready if specified
	if test.MinReady > 0 {
		if readyCount < test.MinReady {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Found %d ready nodes, expected at least %d", readyCount, test.MinReady)
			result.Duration = time.Since(start)
			return result
		}
	}

	// Check min_version if specified
	if test.MinVersion != "" {
		failedVersions := []string{}
		for i, version := range versions {
			if !isVersionAtLeast(version, test.MinVersion) {
				failedVersions = append(failedVersions, version)
			}
			// Store individual node versions in details
			if i < 10 { // Limit to first 10 versions to avoid cluttering details
				result.Details[fmt.Sprintf("node_%d_version", i)] = version
			}
		}
		if len(failedVersions) > 0 {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Found %d nodes with version < %s: %v", len(failedVersions), test.MinVersion, failedVersions)
			result.Duration = time.Since(start)
			return result
		}
	}

	// Build success message
	var conditions []string
	if test.Count > 0 {
		conditions = append(conditions, fmt.Sprintf("%d nodes", test.Count))
	}
	if test.MinCount > 0 {
		conditions = append(conditions, fmt.Sprintf("≥%d nodes", test.MinCount))
	}
	if test.MinReady > 0 {
		conditions = append(conditions, fmt.Sprintf("≥%d ready", test.MinReady))
	}
	if test.MinVersion != "" {
		conditions = append(conditions, fmt.Sprintf("version ≥%s", test.MinVersion))
	}

	result.Message = fmt.Sprintf("Cluster has %s", strings.Join(conditions, ", "))
	result.Duration = time.Since(start)
	return result
}

// isNodeReady checks if a node has Ready=True condition
func isNodeReady(node map[string]interface{}) bool {
	conditions, ok := getNestedSlice(node, "status", "conditions")
	if !ok {
		return false
	}

	for _, cond := range conditions {
		condMap, ok := cond.(map[string]interface{})
		if !ok {
			continue
		}
		condType, _ := condMap["type"].(string)
		condStatus, _ := condMap["status"].(string)
		if condType == "Ready" && condStatus == "True" {
			return true
		}
	}
	return false
}

// isVersionAtLeast checks if actualVersion >= minVersion
// Supports basic semantic versioning (e.g., "v1.28.0")
func isVersionAtLeast(actualVersion, minVersion string) bool {
	// Strip 'v' prefix if present
	actual := strings.TrimPrefix(actualVersion, "v")
	min := strings.TrimPrefix(minVersion, "v")

	// Simple string comparison for now
	// For full semver support, would use a proper semver library
	return actual >= min
}
