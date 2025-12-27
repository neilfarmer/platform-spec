package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// executeKubernetesPVCTest executes a Kubernetes PVC test
func executeKubernetesPVCTest(ctx context.Context, provider core.Provider, test core.KubernetesPVCTest) core.Result {
	start := time.Now()
	result := core.Result{
		Name:    test.Name,
		Status:  core.StatusPass,
		Details: make(map[string]interface{}),
	}

	// Build kubectl command
	cmd := fmt.Sprintf("kubectl get pvc %s -n %s -o json 2>&1", test.PVC, test.Namespace)
	stdout, stderr, exitCode, err := provider.ExecuteCommand(ctx, cmd)

	if err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Error checking PVC %s: %v", test.PVC, err)
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
		result.Message = fmt.Sprintf("PVC %s not found in namespace %s", test.PVC, test.Namespace)
		result.Duration = time.Since(start)
		return result
	}

	if test.State == "absent" && exists {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("PVC %s exists but should be absent", test.PVC)
		result.Duration = time.Since(start)
		return result
	}

	// If expecting absent and it is absent, we're done
	if test.State == "absent" {
		result.Message = fmt.Sprintf("PVC %s is absent", test.PVC)
		result.Duration = time.Since(start)
		return result
	}

	result.Details["namespace"] = test.Namespace

	// Parse PVC JSON for further validation
	var pvc map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &pvc); err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Failed to parse PVC JSON: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// Check status if specified
	if test.Status != "" {
		actualStatus, _ := getNestedString(pvc, "status", "phase")
		if actualStatus != test.Status {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("PVC %s has status %s, expected %s", test.PVC, actualStatus, test.Status)
			result.Duration = time.Since(start)
			return result
		}
		result.Details["status"] = actualStatus
	}

	// Check storage class if specified
	if test.StorageClass != "" {
		actualClass, _ := getNestedString(pvc, "spec", "storageClassName")
		if actualClass != test.StorageClass {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("PVC %s has storage class %s, expected %s", test.PVC, actualClass, test.StorageClass)
			result.Duration = time.Since(start)
			return result
		}
		result.Details["storage_class"] = actualClass
	}

	// Check minimum capacity if specified
	if test.MinCapacity != "" {
		actualCapacity, _ := getNestedString(pvc, "status", "capacity", "storage")
		if actualCapacity == "" {
			// PVC might not be bound yet
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("PVC %s does not have capacity information (not bound yet?)", test.PVC)
			result.Duration = time.Since(start)
			return result
		}

		// Parse capacities (e.g., "100Gi")
		minBytes, err := parseStorageSize(test.MinCapacity)
		if err != nil {
			result.Status = core.StatusError
			result.Message = fmt.Sprintf("Invalid min_capacity format: %v", err)
			result.Duration = time.Since(start)
			return result
		}

		actualBytes, err := parseStorageSize(actualCapacity)
		if err != nil {
			result.Status = core.StatusError
			result.Message = fmt.Sprintf("Failed to parse actual capacity: %v", err)
			result.Duration = time.Since(start)
			return result
		}

		if actualBytes < minBytes {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("PVC %s has capacity %s, minimum required %s", test.PVC, actualCapacity, test.MinCapacity)
			result.Duration = time.Since(start)
			return result
		}
		result.Details["capacity"] = actualCapacity
	}

	// Build success message
	result.Message = fmt.Sprintf("PVC %s exists", test.PVC)
	if test.Status != "" || test.StorageClass != "" || test.MinCapacity != "" {
		parts := []string{}
		if test.Status != "" {
			parts = append(parts, "correct status")
		}
		if test.StorageClass != "" {
			parts = append(parts, "correct storage class")
		}
		if test.MinCapacity != "" {
			parts = append(parts, "sufficient capacity")
		}
		result.Message = fmt.Sprintf("PVC %s exists with %s", test.PVC, strings.Join(parts, ", "))
	}

	result.Duration = time.Since(start)
	return result
}

// parseStorageSize parses Kubernetes storage sizes (e.g., "100Gi", "1Ti", "500Mi") to bytes
func parseStorageSize(size string) (int64, error) {
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)(Ki|Mi|Gi|Ti|Pi|Ei|K|M|G|T|P|E)?$`)
	matches := re.FindStringSubmatch(size)
	if len(matches) == 0 {
		return 0, fmt.Errorf("invalid storage size format: %s", size)
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, err
	}

	unit := matches[2]
	multiplier := int64(1)

	switch unit {
	case "Ki":
		multiplier = 1024
	case "Mi":
		multiplier = 1024 * 1024
	case "Gi":
		multiplier = 1024 * 1024 * 1024
	case "Ti":
		multiplier = 1024 * 1024 * 1024 * 1024
	case "Pi":
		multiplier = 1024 * 1024 * 1024 * 1024 * 1024
	case "Ei":
		multiplier = 1024 * 1024 * 1024 * 1024 * 1024 * 1024
	case "K":
		multiplier = 1000
	case "M":
		multiplier = 1000 * 1000
	case "G":
		multiplier = 1000 * 1000 * 1000
	case "T":
		multiplier = 1000 * 1000 * 1000 * 1000
	case "P":
		multiplier = 1000 * 1000 * 1000 * 1000 * 1000
	case "E":
		multiplier = 1000 * 1000 * 1000 * 1000 * 1000 * 1000
	case "":
		multiplier = 1
	}

	return int64(value * float64(multiplier)), nil
}
