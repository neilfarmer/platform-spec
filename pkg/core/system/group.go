package system

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// executeGroupTest executes a group test
func executeGroupTest(ctx context.Context, provider core.Provider, test core.GroupTest) core.Result {
	start := time.Now()
	result := core.Result{
		Name:    test.Name,
		Status:  core.StatusPass,
		Details: make(map[string]interface{}),
	}

	for _, group := range test.Groups {
		exists, gid, err := groupExists(ctx, provider, group)
		if err != nil {
			result.Status = core.StatusError
			result.Message = fmt.Sprintf("Error checking group %s: %v", group, err)
			result.Duration = time.Since(start)
			return result
		}

		if test.State == "present" && !exists {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Group %s does not exist", group)
			result.Details[group] = "not found"
			break
		} else if test.State == "absent" && exists {
			result.Status = core.StatusFail
			result.Message = fmt.Sprintf("Group %s exists but should be absent", group)
			result.Details[group] = fmt.Sprintf("gid: %s", gid)
			break
		} else if exists {
			result.Details[group] = fmt.Sprintf("gid: %s", gid)
		} else {
			result.Details[group] = "absent"
		}
	}

	if result.Status == core.StatusPass {
		if test.State == "present" {
			result.Message = fmt.Sprintf("All %d groups exist", len(test.Groups))
		} else {
			result.Message = fmt.Sprintf("All %d groups are absent as expected", len(test.Groups))
		}
	}

	result.Duration = time.Since(start)
	return result
}

// groupExists checks if a group exists
func groupExists(ctx context.Context, provider core.Provider, groupname string) (bool, string, error) {
	stdout, _, exitCode, err := provider.ExecuteCommand(ctx, fmt.Sprintf("getent group %s 2>/dev/null", groupname))
	if err != nil {
		return false, "", err
	}

	if exitCode != 0 || stdout == "" {
		return false, "", nil
	}

	// Parse group line: groupname:x:gid:members
	parts := strings.Split(strings.TrimSpace(stdout), ":")
	if len(parts) < 3 {
		return false, "", fmt.Errorf("unexpected group format")
	}

	return true, parts[2], nil
}
