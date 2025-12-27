package system

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// executePingTest executes a network reachability test using ping
func executePingTest(ctx context.Context, provider core.Provider, test core.PingTest) core.Result {
	start := time.Now()
	result := core.Result{
		Name:    test.Name,
		Status:  core.StatusPass,
		Details: make(map[string]interface{}),
	}

	// Use ping with 1 packet and 5 second timeout
	// -c 1: send 1 packet
	// -W 5: wait 5 seconds for response
	stdout, stderr, exitCode, err := provider.ExecuteCommand(ctx, fmt.Sprintf("ping -c 1 -W 5 %s 2>&1", core.ShellQuote(test.Host)))
	if err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Error pinging %s: %v", test.Host, err)
		result.Duration = time.Since(start)
		return result
	}

	result.Details["host"] = test.Host
	result.Details["exit_code"] = exitCode

	if exitCode != 0 {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("Host %s is not reachable", test.Host)
		if stderr != "" {
			result.Details["error"] = stderr
		}
		result.Duration = time.Since(start)
		return result
	}

	// Parse ping output for additional info
	if stdout != "" {
		result.Details["output"] = strings.TrimSpace(stdout)
	}

	result.Message = fmt.Sprintf("Host %s is reachable", test.Host)
	result.Duration = time.Since(start)
	return result
}
