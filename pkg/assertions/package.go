package assertions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// CommandExecutor defines the interface for executing commands
type CommandExecutor interface {
	ExecuteCommand(ctx context.Context, command string) (stdout, stderr string, exitCode int, err error)
}

// CheckPackages checks if packages are installed/absent
func CheckPackages(ctx context.Context, executor CommandExecutor, test core.PackageTest) core.Result {
	start := time.Now()
	result := core.Result{
		Name:    test.Name,
		Status:  core.StatusPass,
		Details: make(map[string]interface{}),
	}

	for _, pkg := range test.Packages {
		// Detect package manager and check package
		installed, version, err := isPackageInstalled(ctx, executor, pkg)
		if err != nil {
			result.Status = core.StatusError
			result.Message = fmt.Sprintf("Error checking package %s: %v", pkg, err)
			result.Duration = time.Since(start)
			return result
		}

		if test.State == "present" && !installed {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Package %s is not installed", pkg)
			result.Details[pkg] = "not installed"
		} else if test.State == "absent" && installed {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Package %s is installed but should be absent", pkg)
			result.Details[pkg] = fmt.Sprintf("installed (version: %s)", version)
		} else {
			result.Details[pkg] = version
		}

		// If any package fails and we haven't failed yet, update message
		if result.Status == core.StatusFail {
			break
		}
	}

	if result.Status == core.StatusPass {
		if test.State == "present" {
			result.Message = fmt.Sprintf("All %d packages are installed", len(test.Packages))
		} else {
			result.Message = fmt.Sprintf("All %d packages are absent as expected", len(test.Packages))
		}
	}

	result.Duration = time.Since(start)
	return result
}

// isPackageInstalled checks if a package is installed and returns its version
func isPackageInstalled(ctx context.Context, executor CommandExecutor, pkg string) (bool, string, error) {
	// Try dpkg (Debian/Ubuntu)
	stdout, _, exitCode, err := executor.ExecuteCommand(ctx, fmt.Sprintf("dpkg -l %s 2>/dev/null | grep '^ii'", pkg))
	if err != nil {
		return false, "", err
	}
	if exitCode == 0 && stdout != "" {
		version := extractDpkgVersion(stdout)
		return true, version, nil
	}

	// Try rpm (RHEL/CentOS/Fedora)
	stdout, _, exitCode, err = executor.ExecuteCommand(ctx, fmt.Sprintf("rpm -q %s 2>/dev/null", pkg))
	if err != nil {
		return false, "", err
	}
	if exitCode == 0 && stdout != "" {
		return true, strings.TrimSpace(stdout), nil
	}

	// Try apk (Alpine)
	stdout, _, exitCode, err = executor.ExecuteCommand(ctx, fmt.Sprintf("apk info -e %s 2>/dev/null", pkg))
	if err != nil {
		return false, "", err
	}
	if exitCode == 0 && stdout != "" {
		return true, strings.TrimSpace(stdout), nil
	}

	return false, "", nil
}

// extractDpkgVersion extracts version from dpkg output
func extractDpkgVersion(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return "unknown"
	}
	fields := strings.Fields(lines[0])
	if len(fields) >= 3 {
		return fields[2]
	}
	return "unknown"
}
