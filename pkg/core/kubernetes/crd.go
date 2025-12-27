package kubernetes

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// executeKubernetesCRDTest executes a Kubernetes CustomResourceDefinition test
func executeKubernetesCRDTest(ctx context.Context, provider core.Provider, test core.KubernetesCRDTest) core.Result {
	start := time.Now()
	result := core.Result{
		Name:    test.Name,
		Status:  core.StatusPass,
		Details: make(map[string]interface{}),
	}

	// Build kubectl command
	cmd := fmt.Sprintf("kubectl get crd %s -o json 2>&1", test.CRD)
	_, stderr, exitCode, err := provider.ExecuteCommand(ctx, cmd)

	if err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Error checking CRD %s: %v", test.CRD, err)
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
		result.Message = fmt.Sprintf("CRD %s not found", test.CRD)
		result.Duration = time.Since(start)
		return result
	}

	if test.State == "absent" && exists {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("CRD %s exists but should be absent", test.CRD)
		result.Duration = time.Since(start)
		return result
	}

	// Build success message
	if test.State == "absent" {
		result.Message = fmt.Sprintf("CRD %s is absent", test.CRD)
	} else {
		result.Message = fmt.Sprintf("CRD %s exists", test.CRD)
	}

	result.Duration = time.Since(start)
	return result
}
