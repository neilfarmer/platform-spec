package system

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// executePortTest executes a port listening state test
func executePortTest(ctx context.Context, provider core.Provider, test core.PortTest) core.Result {
	start := time.Now()
	result := core.Result{
		Name:    test.Name,
		Status:  core.StatusPass,
		Details: make(map[string]interface{}),
	}

	// Build ss command based on protocol
	var cmd string
	if test.Protocol == "tcp" {
		cmd = fmt.Sprintf("ss -tln | grep -E ':%d\\s' || true", test.Port)
	} else { // udp
		cmd = fmt.Sprintf("ss -uln | grep -E ':%d\\s' || true", test.Port)
	}

	stdout, stderr, exitCode, err := provider.ExecuteCommand(ctx, cmd)

	if err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Error checking port: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// exitCode will be 0 if grep found something (port is listening)
	// or 0 from '|| true' if grep found nothing (port not listening)
	_ = exitCode

	// Store details
	result.Details["port"] = test.Port
	result.Details["protocol"] = test.Protocol
	result.Details["expected_state"] = test.State

	// Check if port is listening by parsing ss output
	isListening := strings.TrimSpace(stdout) != ""

	if test.State == "listening" {
		if !isListening {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Port %d/%s is not listening", test.Port, test.Protocol)
		} else {
			result.Message = fmt.Sprintf("Port %d/%s is listening", test.Port, test.Protocol)
			result.Details["actual_state"] = "listening"
		}
	} else { // state == "closed"
		if isListening {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Port %d/%s is listening, expected closed", test.Port, test.Protocol)
			result.Details["actual_state"] = "listening"
		} else {
			result.Message = fmt.Sprintf("Port %d/%s is closed", test.Port, test.Protocol)
			result.Details["actual_state"] = "closed"
		}
	}

	// Log stderr if present (for debugging)
	if stderr != "" && result.Status != core.StatusFail {
		result.Details["stderr"] = strings.TrimSpace(stderr)
	}

	result.Duration = time.Since(start)
	return result
}
