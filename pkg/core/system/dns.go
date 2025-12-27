package system

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// executeDNSTest executes a DNS resolution test
func executeDNSTest(ctx context.Context, provider core.Provider, test core.DNSTest) core.Result {
	start := time.Now()
	result := core.Result{
		Name:    test.Name,
		Status:  core.StatusPass,
		Details: make(map[string]interface{}),
	}

	// Try dig first, fall back to getent hosts
	// dig +short returns just the IP addresses
	stdout, _, exitCode, err := provider.ExecuteCommand(ctx, fmt.Sprintf("dig +short %s 2>/dev/null || getent hosts %s 2>/dev/null | awk '{print $1}'", core.ShellQuote(test.Host), core.ShellQuote(test.Host)))
	if err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Error resolving DNS for %s: %v", test.Host, err)
		result.Duration = time.Since(start)
		return result
	}

	result.Details["host"] = test.Host

	stdout = strings.TrimSpace(stdout)
	if exitCode != 0 || stdout == "" {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("DNS resolution failed for %s", test.Host)
		result.Duration = time.Since(start)
		return result
	}

	// Parse resolved IPs
	ips := strings.Split(stdout, "\n")
	var validIPs []string
	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		if ip != "" {
			validIPs = append(validIPs, ip)
		}
	}

	if len(validIPs) == 0 {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("DNS resolution failed for %s", test.Host)
		result.Duration = time.Since(start)
		return result
	}

	result.Details["resolved_ips"] = validIPs
	result.Message = fmt.Sprintf("DNS resolved %s to %d address(es)", test.Host, len(validIPs))

	result.Duration = time.Since(start)
	return result
}
