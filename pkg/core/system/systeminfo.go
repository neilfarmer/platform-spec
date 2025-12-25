package system

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// executeSystemInfoTest executes a system information validation test
func executeSystemInfoTest(ctx context.Context, provider core.Provider, test core.SystemInfoTest) core.Result {
	start := time.Now()
	result := core.Result{
		Name:    test.Name,
		Status:  core.StatusPass,
		Details: make(map[string]interface{}),
	}

	// Gather system information
	sysInfo := make(map[string]string)

	// Get OS name from /etc/os-release
	stdout, _, _, _ := provider.ExecuteCommand(ctx, "grep '^ID=' /etc/os-release 2>/dev/null | cut -d= -f2 | tr -d '\"'")
	if stdout != "" {
		sysInfo["os"] = strings.TrimSpace(stdout)
	}

	// Get OS version from /etc/os-release
	stdout, _, _, _ = provider.ExecuteCommand(ctx, "grep '^VERSION_ID=' /etc/os-release 2>/dev/null | cut -d= -f2 | tr -d '\"'")
	if stdout != "" {
		sysInfo["os_version"] = strings.TrimSpace(stdout)
	}

	// Get architecture
	stdout, _, _, _ = provider.ExecuteCommand(ctx, "uname -m 2>/dev/null")
	if stdout != "" {
		sysInfo["arch"] = strings.TrimSpace(stdout)
	}

	// Get kernel version
	stdout, _, _, _ = provider.ExecuteCommand(ctx, "uname -r 2>/dev/null")
	if stdout != "" {
		sysInfo["kernel_version"] = strings.TrimSpace(stdout)
	}

	// Get hostname (short)
	stdout, _, _, _ = provider.ExecuteCommand(ctx, "hostname -s 2>/dev/null")
	if stdout != "" {
		sysInfo["hostname"] = strings.TrimSpace(stdout)
	}

	// Get FQDN
	stdout, _, _, _ = provider.ExecuteCommand(ctx, "hostname -f 2>/dev/null")
	if stdout != "" {
		sysInfo["fqdn"] = strings.TrimSpace(stdout)
	}

	// Store all gathered info in details
	for k, v := range sysInfo {
		result.Details[k] = v
	}

	// Validate each specified field
	var failures []string

	// Check OS
	if test.OS != "" {
		if sysInfo["os"] != test.OS {
			failures = append(failures, fmt.Sprintf("OS is '%s', expected '%s'", sysInfo["os"], test.OS))
		}
	}

	// Check OS version
	if test.OSVersion != "" {
		if !versionMatches(sysInfo["os_version"], test.OSVersion, test.VersionMatch) {
			failures = append(failures, fmt.Sprintf("OS version is '%s', expected '%s'", sysInfo["os_version"], test.OSVersion))
		}
	}

	// Check architecture
	if test.Arch != "" {
		if sysInfo["arch"] != test.Arch {
			failures = append(failures, fmt.Sprintf("Architecture is '%s', expected '%s'", sysInfo["arch"], test.Arch))
		}
	}

	// Check kernel version
	if test.KernelVersion != "" {
		if !versionMatches(sysInfo["kernel_version"], test.KernelVersion, test.VersionMatch) {
			failures = append(failures, fmt.Sprintf("Kernel version is '%s', expected '%s'", sysInfo["kernel_version"], test.KernelVersion))
		}
	}

	// Check hostname
	if test.Hostname != "" {
		if sysInfo["hostname"] != test.Hostname {
			failures = append(failures, fmt.Sprintf("Hostname is '%s', expected '%s'", sysInfo["hostname"], test.Hostname))
		}
	}

	// Check FQDN
	if test.FQDN != "" {
		if sysInfo["fqdn"] != test.FQDN {
			failures = append(failures, fmt.Sprintf("FQDN is '%s', expected '%s'", sysInfo["fqdn"], test.FQDN))
		}
	}

	// Set result based on failures
	if len(failures) > 0 {
		result.Status = core.StatusFail
		result.Message = strings.Join(failures, "; ")
	} else {
		result.Message = "System information matches all specified criteria"
	}

	result.Duration = time.Since(start)
	return result
}

// versionMatches checks if a version matches based on the match mode
func versionMatches(actual, expected, matchMode string) bool {
	if matchMode == "exact" {
		return actual == expected
	}
	// prefix mode
	return strings.HasPrefix(actual, expected)
}
