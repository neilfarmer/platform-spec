package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// executeKubernetesIngressTest executes a Kubernetes ingress test
func executeKubernetesIngressTest(ctx context.Context, provider core.Provider, test core.KubernetesIngressTest) core.Result {
	start := time.Now()
	result := core.Result{
		Name:    test.Name,
		Status:  core.StatusPass,
		Details: make(map[string]interface{}),
	}

	// Build kubectl command
	cmd := fmt.Sprintf("kubectl get ingress %s -n %s -o json 2>&1", test.Ingress, test.Namespace)
	stdout, stderr, exitCode, err := provider.ExecuteCommand(ctx, cmd)

	if err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Error checking ingress %s: %v", test.Ingress, err)
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
		result.Message = fmt.Sprintf("Ingress %s not found in namespace %s", test.Ingress, test.Namespace)
		result.Duration = time.Since(start)
		return result
	}

	if test.State == "absent" && exists {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("Ingress %s exists but should be absent", test.Ingress)
		result.Duration = time.Since(start)
		return result
	}

	// If expecting absent and it is absent, we're done
	if test.State == "absent" {
		result.Message = fmt.Sprintf("Ingress %s is absent", test.Ingress)
		result.Duration = time.Since(start)
		return result
	}

	result.Details["namespace"] = test.Namespace

	// Parse ingress JSON for further validation
	var ingress map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &ingress); err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Failed to parse ingress JSON: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// Check ingress class if specified
	if test.IngressClass != "" {
		// Try spec.ingressClassName first (newer API)
		actualClass, _ := getNestedString(ingress, "spec", "ingressClassName")
		if actualClass == "" {
			// Fall back to annotation (older API)
			annotations, _ := getNestedMap(ingress, "metadata", "annotations")
			if classVal, ok := annotations["kubernetes.io/ingress.class"]; ok {
				actualClass = classVal
			}
		}
		if actualClass != test.IngressClass {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Ingress %s has class %s, expected %s", test.Ingress, actualClass, test.IngressClass)
			result.Duration = time.Since(start)
			return result
		}
		result.Details["ingress_class"] = actualClass
	}

	// Check hosts if specified
	if len(test.Hosts) > 0 {
		rules, _ := getNestedSlice(ingress, "spec", "rules")
		var actualHosts []string
		for _, rule := range rules {
			if ruleMap, ok := rule.(map[string]interface{}); ok {
				if host, ok := ruleMap["host"].(string); ok && host != "" {
					actualHosts = append(actualHosts, host)
				}
			}
		}

		// Check all expected hosts are present
		var missingHosts []string
		for _, expectedHost := range test.Hosts {
			found := false
			for _, actualHost := range actualHosts {
				if actualHost == expectedHost {
					found = true
					break
				}
			}
			if !found {
				missingHosts = append(missingHosts, expectedHost)
			}
		}

		if len(missingHosts) > 0 {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Ingress %s missing hosts: %s", test.Ingress, strings.Join(missingHosts, ", "))
			result.Duration = time.Since(start)
			return result
		}
		result.Details["hosts"] = actualHosts
	}

	// Check TLS if specified
	if test.TLS {
		tlsSlice, _ := getNestedSlice(ingress, "spec", "tls")
		if len(tlsSlice) == 0 {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Ingress %s does not have TLS configured", test.Ingress)
			result.Duration = time.Since(start)
			return result
		}
		result.Details["tls_configured"] = true
	}

	// Build success message
	result.Message = fmt.Sprintf("Ingress %s exists", test.Ingress)
	if test.IngressClass != "" || len(test.Hosts) > 0 || test.TLS {
		parts := []string{}
		if test.IngressClass != "" {
			parts = append(parts, "correct class")
		}
		if len(test.Hosts) > 0 {
			parts = append(parts, "all hosts")
		}
		if test.TLS {
			parts = append(parts, "TLS configured")
		}
		result.Message = fmt.Sprintf("Ingress %s exists with %s", test.Ingress, strings.Join(parts, ", "))
	}

	result.Duration = time.Since(start)
	return result
}
