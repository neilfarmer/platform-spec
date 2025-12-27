package system

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// executeFilesystemTest executes a filesystem/mount point test
func executeFilesystemTest(ctx context.Context, provider core.Provider, test core.FilesystemTest) core.Result {
	start := time.Now()
	result := core.Result{
		Name:    test.Name,
		Status:  core.StatusPass,
		Details: make(map[string]interface{}),
	}

	// Check if path is mounted using findmnt
	stdout, _, exitCode, err := provider.ExecuteCommand(ctx, fmt.Sprintf("findmnt --noheadings --output TARGET,FSTYPE,OPTIONS,SIZE,USED,USE%% --target %s 2>/dev/null", test.Path))
	if err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Error checking filesystem %s: %v", test.Path, err)
		result.Duration = time.Since(start)
		return result
	}

	isMounted := (exitCode == 0 && stdout != "")

	// Check mount state
	if test.State == "mounted" && !isMounted {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("Filesystem %s is not mounted", test.Path)
		result.Duration = time.Since(start)
		return result
	} else if test.State == "unmounted" && isMounted {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("Filesystem %s is mounted but should be unmounted", test.Path)
		result.Duration = time.Since(start)
		return result
	}

	// If checking for unmounted and it is unmounted, we're done
	if test.State == "unmounted" && !isMounted {
		result.Message = fmt.Sprintf("Filesystem %s is not mounted as expected", test.Path)
		result.Duration = time.Since(start)
		return result
	}

	// Parse mount information
	stdout = strings.TrimSpace(stdout)
	fields := strings.Fields(stdout)
	if len(fields) < 6 {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Unexpected findmnt output for %s", test.Path)
		result.Duration = time.Since(start)
		return result
	}

	fstype := fields[1]
	options := fields[2]
	size := fields[3]
	used := fields[4]
	usagePercent := strings.TrimSuffix(fields[5], "%")

	result.Details["fstype"] = fstype
	result.Details["options"] = options
	result.Details["size"] = size
	result.Details["used"] = used
	result.Details["usage_percent"] = usagePercent

	// Check filesystem type
	if test.Fstype != "" && fstype != test.Fstype {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("Filesystem %s type is %s, expected %s", test.Path, fstype, test.Fstype)
		result.Duration = time.Since(start)
		return result
	}

	// Check mount options
	if len(test.Options) > 0 {
		mountOpts := strings.Split(options, ",")
		for _, requiredOpt := range test.Options {
			found := false
			for _, mountOpt := range mountOpts {
				if mountOpt == requiredOpt {
					found = true
					break
				}
			}
			if !found {
				result.Status = core.StatusFail
				result.Message = fmt.Sprintf("Filesystem %s missing required mount option '%s'", test.Path, requiredOpt)
				result.Duration = time.Since(start)
				return result
			}
		}
	}

	// Check minimum size
	if test.MinSizeGB > 0 {
		// Get size in GB - use df for more reliable size info
		stdout, _, _, err := provider.ExecuteCommand(ctx, fmt.Sprintf("df -BG --output=size %s | tail -1 | tr -d 'G '", test.Path))
		if err == nil {
			var actualSizeGB int
			_, scanErr := fmt.Sscanf(strings.TrimSpace(stdout), "%d", &actualSizeGB)
			if scanErr != nil {
				result.Status = core.StatusError
				result.Message = fmt.Sprintf("Error parsing filesystem size for %s: %v", test.Path, scanErr)
				result.Duration = time.Since(start)
				return result
			}
			if actualSizeGB < test.MinSizeGB {
				result.Status = core.StatusFail
				result.Message = fmt.Sprintf("Filesystem %s size is %dGB, minimum required is %dGB", test.Path, actualSizeGB, test.MinSizeGB)
				result.Duration = time.Since(start)
				return result
			}
		}
	}

	// Check maximum usage percentage
	if test.MaxUsagePercent > 0 {
		var actualUsagePercent int
		_, scanErr := fmt.Sscanf(usagePercent, "%d", &actualUsagePercent)
		if scanErr != nil {
			result.Status = core.StatusError
			result.Message = fmt.Sprintf("Error parsing filesystem usage percent for %s: %v", test.Path, scanErr)
			result.Duration = time.Since(start)
			return result
		}
		if actualUsagePercent > test.MaxUsagePercent {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Filesystem %s usage is %d%%, maximum allowed is %d%%", test.Path, actualUsagePercent, test.MaxUsagePercent)
			result.Duration = time.Since(start)
			return result
		}
	}

	// Build success message
	result.Message = fmt.Sprintf("Filesystem %s is mounted", test.Path)
	if test.Fstype != "" {
		result.Message += fmt.Sprintf(" as %s", fstype)
	}
	if len(test.Options) > 0 {
		result.Message += " with correct options"
	}

	result.Duration = time.Since(start)
	return result
}
