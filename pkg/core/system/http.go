package system

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// executeHTTPTest executes an HTTP endpoint test
func executeHTTPTest(ctx context.Context, provider core.Provider, test core.HTTPTest) core.Result {
	start := time.Now()
	result := core.Result{
		Name:    test.Name,
		Status:  core.StatusPass,
		Details: make(map[string]interface{}),
	}

	// Build curl command with proper shell quoting for write-out format
	var cmdParts []string
	cmdParts = append(cmdParts, "curl -s")

	// Add method
	if test.Method != "GET" {
		cmdParts = append(cmdParts, fmt.Sprintf("-X %s", test.Method))
	}

	// Add follow redirects flag if needed
	if test.FollowRedirects {
		cmdParts = append(cmdParts, "-L")
	}

	// Add insecure flag if needed
	if test.Insecure {
		cmdParts = append(cmdParts, "-k")
	}

	// Add write-out format with proper escaping - use $'...' for newline interpretation
	cmdParts = append(cmdParts, "-w $'\\n%{http_code}'")

	// Add URL
	cmdParts = append(cmdParts, core.ShellQuote(test.URL))

	// Execute command
	cmd := strings.Join(cmdParts, " ")
	stdout, stderr, exitCode, err := provider.ExecuteCommand(ctx, cmd)

	if err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Error making HTTP request: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	if exitCode != 0 {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("HTTP request failed: %s", strings.TrimSpace(stderr))
		result.Duration = time.Since(start)
		return result
	}

	// Parse response: last line is status code, everything else is body
	lines := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
	if len(lines) < 1 {
		result.Status = core.StatusError
		result.Message = "No response from HTTP request"
		result.Duration = time.Since(start)
		return result
	}

	statusCodeStr := lines[len(lines)-1]
	statusCode := 0
	if _, err := fmt.Sscanf(statusCodeStr, "%d", &statusCode); err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Failed to parse status code: %s", statusCodeStr)
		result.Duration = time.Since(start)
		return result
	}

	// Get body (everything except last line)
	body := ""
	if len(lines) > 1 {
		body = strings.Join(lines[:len(lines)-1], "\n")
	}

	// Store details
	result.Details["status_code"] = statusCode
	result.Details["url"] = test.URL
	result.Details["method"] = test.Method

	// Check status code
	if statusCode != test.StatusCode {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("Status code is %d, expected %d", statusCode, test.StatusCode)
		result.Duration = time.Since(start)
		return result
	}

	// Check body contains strings
	if len(test.Contains) > 0 {
		var missingStrings []string
		for _, str := range test.Contains {
			if !strings.Contains(body, str) {
				missingStrings = append(missingStrings, fmt.Sprintf("'%s'", str))
			}
		}

		if len(missingStrings) > 0 {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Response body missing expected strings: %s", strings.Join(missingStrings, ", "))
			result.Duration = time.Since(start)
			return result
		}
	}

	// All checks passed
	result.Message = fmt.Sprintf("HTTP %s returned status %d", test.URL, statusCode)
	if len(test.Contains) > 0 {
		result.Message += fmt.Sprintf(" with all expected content (%d strings)", len(test.Contains))
	}

	result.Duration = time.Since(start)
	return result
}
