package assertions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// CheckFile checks file/directory properties
func CheckFile(ctx context.Context, executor CommandExecutor, test core.FileTest) core.Result {
	start := time.Now()
	result := core.Result{
		Name:    test.Name,
		Status:  core.StatusPass,
		Details: make(map[string]interface{}),
	}

	// Check if path exists and get its type
	stdout, _, exitCode, err := executor.ExecuteCommand(ctx, fmt.Sprintf("stat -c '%%F:%%U:%%G:%%a' %s 2>/dev/null || echo 'notfound'", test.Path))
	if err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Error checking path %s: %v", test.Path, err)
		result.Duration = time.Since(start)
		return result
	}

	stdout = strings.TrimSpace(stdout)
	if stdout == "notfound" || exitCode != 0 {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("Path %s does not exist", test.Path)
		result.Duration = time.Since(start)
		return result
	}

	// Parse stat output: type:owner:group:mode
	parts := strings.Split(stdout, ":")
	if len(parts) != 4 {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Unexpected stat output for %s", test.Path)
		result.Duration = time.Since(start)
		return result
	}

	fileType := normalizeFileType(parts[0])
	owner := parts[1]
	group := parts[2]
	mode := parts[3]

	result.Details["type"] = fileType
	result.Details["owner"] = owner
	result.Details["group"] = group
	result.Details["mode"] = mode

	// Check type
	if test.Type != "" {
		if !matchesFileType(fileType, test.Type) {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Path %s is a %s, expected %s", test.Path, fileType, test.Type)
			result.Duration = time.Since(start)
			return result
		}
	}

	// Check owner
	if test.Owner != "" && owner != test.Owner {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("Path %s owner is %s, expected %s", test.Path, owner, test.Owner)
		result.Duration = time.Since(start)
		return result
	}

	// Check group
	if test.Group != "" && group != test.Group {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("Path %s group is %s, expected %s", test.Path, group, test.Group)
		result.Duration = time.Since(start)
		return result
	}

	// Check mode
	if test.Mode != "" {
		expectedMode := normalizeMode(test.Mode)
		if mode != expectedMode {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Path %s mode is %s, expected %s", test.Path, mode, expectedMode)
			result.Duration = time.Since(start)
			return result
		}
	}

	result.Message = fmt.Sprintf("Path %s exists with correct properties", test.Path)
	result.Duration = time.Since(start)
	return result
}

// normalizeFileType converts stat output to simple type
func normalizeFileType(statType string) string {
	statType = strings.ToLower(statType)
	if strings.Contains(statType, "directory") {
		return "directory"
	}
	if strings.Contains(statType, "regular") {
		return "file"
	}
	if strings.Contains(statType, "symbolic link") {
		return "symlink"
	}
	return statType
}

// matchesFileType checks if actual type matches expected type
func matchesFileType(actual, expected string) bool {
	actual = strings.ToLower(actual)
	expected = strings.ToLower(expected)

	if expected == "directory" && actual == "directory" {
		return true
	}
	if expected == "file" && actual == "file" {
		return true
	}
	if expected == "symlink" && actual == "symlink" {
		return true
	}
	return actual == expected
}

// normalizeMode ensures mode is in octal format without leading 0
func normalizeMode(mode string) string {
	// Remove any leading 0 or 0o prefix
	mode = strings.TrimPrefix(mode, "0o")
	mode = strings.TrimPrefix(mode, "0")

	// Ensure it's 3 or 4 digits
	if len(mode) == 3 || len(mode) == 4 {
		return mode
	}
	return mode
}
