package system

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// executeServiceTest executes a service test
func executeServiceTest(ctx context.Context, provider core.Provider, test core.ServiceTest) core.Result {
	start := time.Now()
	result := core.Result{
		Name:    test.Name,
		Status:  core.StatusPass,
		Details: make(map[string]interface{}),
	}

	// Get list of services to check
	services := test.Services
	if test.Service != "" {
		services = []string{test.Service}
	}

	for _, service := range services {
		// Check service status using systemctl (systemd)
		running, enabled, err := checkServiceStatus(ctx, provider, service)
		if err != nil {
			result.Status = core.StatusError
			result.Message = fmt.Sprintf("Error checking service %s: %v", service, err)
			result.Duration = time.Since(start)
			return result
		}

		// Check running state
		if test.State == "running" && !running {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Service %s is not running", service)
			result.Details[service] = "not running"
			break
		} else if test.State == "stopped" && running {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Service %s is running but should be stopped", service)
			result.Details[service] = "running"
			break
		}

		// Check enabled state if specified
		if test.Enabled && !enabled {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Service %s is not enabled", service)
			result.Details[service] = "not enabled"
			break
		} else if !test.Enabled && test.State == "stopped" && enabled {
			// Only fail on enabled if state is stopped and we explicitly don't want it enabled
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Service %s is enabled but should be disabled", service)
			result.Details[service] = "enabled"
			break
		}

		// Record status
		status := "stopped"
		if running {
			status = "running"
		}
		if enabled {
			status += " (enabled)"
		}
		result.Details[service] = status
	}

	if result.Status == core.StatusPass {
		if test.State == "running" {
			result.Message = fmt.Sprintf("All %d services are running", len(services))
		} else {
			result.Message = fmt.Sprintf("All %d services are stopped", len(services))
		}
		if test.Enabled {
			result.Message += " and enabled"
		}
	}

	result.Duration = time.Since(start)
	return result
}

// checkServiceStatus checks if a service is running and enabled
func checkServiceStatus(ctx context.Context, provider core.Provider, service string) (running, enabled bool, err error) {
	// Try systemctl (systemd)
	stdout, _, exitCode, err := provider.ExecuteCommand(ctx, fmt.Sprintf("systemctl is-active %s 2>/dev/null", service))
	if err != nil {
		return false, false, err
	}

	stdout = strings.TrimSpace(stdout)
	running = (exitCode == 0 && stdout == "active")

	// Check if enabled
	stdout, _, exitCode, err = provider.ExecuteCommand(ctx, fmt.Sprintf("systemctl is-enabled %s 2>/dev/null", service))
	if err != nil {
		return running, false, nil // Don't fail if we can't check enabled status
	}

	stdout = strings.TrimSpace(stdout)
	enabled = (exitCode == 0 && stdout == "enabled")

	return running, enabled, nil
}
