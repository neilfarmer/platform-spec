package system

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// executeUserTest executes a user test
func executeUserTest(ctx context.Context, provider core.Provider, test core.UserTest) core.Result {
	start := time.Now()
	result := core.Result{
		Name:    test.Name,
		Status:  core.StatusPass,
		Details: make(map[string]interface{}),
	}

	// Check if user exists
	userInfo, err := getUserInfo(ctx, provider, test.User)
	if err != nil {
		result.Status = core.StatusError
		result.Message = fmt.Sprintf("Error checking user %s: %v", test.User, err)
		result.Duration = time.Since(start)
		return result
	}

	if userInfo == nil {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("User %s does not exist", test.User)
		result.Duration = time.Since(start)
		return result
	}

	result.Details["uid"] = userInfo["uid"]
	result.Details["gid"] = userInfo["gid"]
	result.Details["home"] = userInfo["home"]
	result.Details["shell"] = userInfo["shell"]

	// Check shell if specified
	if test.Shell != "" && userInfo["shell"] != test.Shell {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("User %s shell is %s, expected %s", test.User, userInfo["shell"], test.Shell)
		result.Duration = time.Since(start)
		return result
	}

	// Check home if specified
	if test.Home != "" && userInfo["home"] != test.Home {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("User %s home is %s, expected %s", test.User, userInfo["home"], test.Home)
		result.Duration = time.Since(start)
		return result
	}

	// Check groups if specified
	if len(test.Groups) > 0 {
		userGroups, err := getUserGroups(ctx, provider, test.User)
		if err != nil {
			result.Status = core.StatusError
			result.Message = fmt.Sprintf("Error checking groups for user %s: %v", test.User, err)
			result.Duration = time.Since(start)
			return result
		}

		for _, expectedGroup := range test.Groups {
			found := false
			for _, userGroup := range userGroups {
				if userGroup == expectedGroup {
					found = true
					break
				}
			}
			if !found {
				result.Status = core.StatusFail
				result.Message = fmt.Sprintf("User %s is not in group %s", test.User, expectedGroup)
				result.Duration = time.Since(start)
				return result
			}
		}
		result.Details["groups"] = strings.Join(userGroups, ",")
	}

	result.Message = fmt.Sprintf("User %s exists with correct properties", test.User)
	result.Duration = time.Since(start)
	return result
}

// getUserInfo gets information about a user
func getUserInfo(ctx context.Context, provider core.Provider, username string) (map[string]string, error) {
	stdout, _, exitCode, err := provider.ExecuteCommand(ctx, fmt.Sprintf("id -u %s 2>/dev/null && id -g %s 2>/dev/null && getent passwd %s 2>/dev/null", username, username, username))
	if err != nil {
		return nil, err
	}

	if exitCode != 0 {
		return nil, nil // User does not exist
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) < 3 {
		return nil, nil
	}

	uid := strings.TrimSpace(lines[0])
	gid := strings.TrimSpace(lines[1])
	passwdLine := strings.TrimSpace(lines[2])

	// Parse passwd line: username:x:uid:gid:gecos:home:shell
	parts := strings.Split(passwdLine, ":")
	if len(parts) < 7 {
		return nil, fmt.Errorf("unexpected passwd format")
	}

	return map[string]string{
		"uid":   uid,
		"gid":   gid,
		"home":  parts[5],
		"shell": parts[6],
	}, nil
}

// getUserGroups gets all groups a user belongs to
func getUserGroups(ctx context.Context, provider core.Provider, username string) ([]string, error) {
	stdout, _, exitCode, err := provider.ExecuteCommand(ctx, fmt.Sprintf("id -Gn %s 2>/dev/null", username))
	if err != nil {
		return nil, err
	}

	if exitCode != 0 {
		return nil, fmt.Errorf("failed to get groups for user %s", username)
	}

	groups := strings.Fields(strings.TrimSpace(stdout))
	return groups, nil
}
