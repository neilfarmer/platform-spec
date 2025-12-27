package system

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// executeCommandContentTest executes a command content test
func executeCommandContentTest(ctx context.Context, provider core.Provider, test core.CommandContentTest) core.Result {
	start := time.Now()
	result := core.Result{
		Name:    test.Name,
		Status:  core.StatusPass,
		Details: make(map[string]interface{}),
	}

	// Execute the command
	stdout, stderr, exitCode, err := provider.ExecuteCommand(ctx, test.Command)
	if err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Error executing command '%s': %v", test.Command, err)
		result.Duration = time.Since(start)
		return result
	}

	result.Details["exit_code"] = exitCode
	result.Details["stdout_length"] = len(stdout)
	result.Details["stderr_length"] = len(stderr)

	// Check exit code if specified
	if test.ExitCode != 0 && exitCode != test.ExitCode {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("Command exit code is %d, expected %d", exitCode, test.ExitCode)
		result.Duration = time.Since(start)
		return result
	}

	// Check contains strings in stdout
	if len(test.Contains) > 0 {
		for _, searchStr := range test.Contains {
			if !strings.Contains(stdout, searchStr) {
				result.Status = core.StatusFail
				result.Message = fmt.Sprintf("Command output does not contain '%s'", searchStr)
				result.Details["missing"] = searchStr
				result.Duration = time.Since(start)
				return result
			}
			result.Details[searchStr] = "found"
		}
	}

	// Build success message
	if len(test.Contains) > 0 && test.ExitCode != 0 {
		result.Message = fmt.Sprintf("Command exited with code %d and output contains all %d strings", test.ExitCode, len(test.Contains))
	} else if len(test.Contains) > 0 {
		result.Message = fmt.Sprintf("Command output contains all %d strings", len(test.Contains))
	} else if test.ExitCode != 0 {
		result.Message = fmt.Sprintf("Command exited with expected code %d", test.ExitCode)
	} else {
		result.Message = "Command executed successfully"
	}

	result.Duration = time.Since(start)
	return result
}
