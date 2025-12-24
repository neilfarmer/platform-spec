package core

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Executor executes tests against a provider
type Executor struct {
	spec     *Spec
	provider Provider
}

// Provider interface that all providers must implement
type Provider interface {
	ExecuteCommand(ctx context.Context, command string) (stdout, stderr string, exitCode int, err error)
}

// NewExecutor creates a new executor
func NewExecutor(spec *Spec, provider Provider) *Executor {
	return &Executor{
		spec:     spec,
		provider: provider,
	}
}

// Execute runs all tests in the spec
func (e *Executor) Execute(ctx context.Context) (*TestResults, error) {
	startTime := time.Now()

	results := &TestResults{
		SpecName:  e.spec.Metadata.Name,
		StartTime: startTime,
		Results:   []Result{},
	}

	// Import assertions package functions
	// We'll call these directly from the executor
	for _, test := range e.spec.Tests.Packages {
		result := e.executePackageTest(ctx, test)
		results.Results = append(results.Results, result)

		// Check fail-fast
		if e.spec.Config.FailFast && result.Status == StatusFail {
			break
		}
	}

	for _, test := range e.spec.Tests.Files {
		result := e.executeFileTest(ctx, test)
		results.Results = append(results.Results, result)

		// Check fail-fast
		if e.spec.Config.FailFast && result.Status == StatusFail {
			break
		}
	}

	for _, test := range e.spec.Tests.Services {
		result := e.executeServiceTest(ctx, test)
		results.Results = append(results.Results, result)

		// Check fail-fast
		if e.spec.Config.FailFast && result.Status == StatusFail {
			break
		}
	}

	for _, test := range e.spec.Tests.Users {
		result := e.executeUserTest(ctx, test)
		results.Results = append(results.Results, result)

		// Check fail-fast
		if e.spec.Config.FailFast && result.Status == StatusFail {
			break
		}
	}

	for _, test := range e.spec.Tests.Groups {
		result := e.executeGroupTest(ctx, test)
		results.Results = append(results.Results, result)

		// Check fail-fast
		if e.spec.Config.FailFast && result.Status == StatusFail {
			break
		}
	}

	for _, test := range e.spec.Tests.FileContent {
		result := e.executeFileContentTest(ctx, test)
		results.Results = append(results.Results, result)

		// Check fail-fast
		if e.spec.Config.FailFast && result.Status == StatusFail {
			break
		}
	}

	for _, test := range e.spec.Tests.CommandContent {
		result := e.executeCommandContentTest(ctx, test)
		results.Results = append(results.Results, result)

		// Check fail-fast
		if e.spec.Config.FailFast && result.Status == StatusFail {
			break
		}
	}

	for _, test := range e.spec.Tests.Docker {
		result := e.executeDockerTest(ctx, test)
		results.Results = append(results.Results, result)

		// Check fail-fast
		if e.spec.Config.FailFast && result.Status == StatusFail {
			break
		}
	}

	results.Duration = time.Since(startTime)
	return results, nil
}

// executePackageTest executes a package test
func (e *Executor) executePackageTest(ctx context.Context, test PackageTest) Result {
	start := time.Now()
	result := Result{
		Name:    test.Name,
		Status:  StatusPass,
		Details: make(map[string]interface{}),
	}

	for _, pkg := range test.Packages {
		installed, version, err := e.isPackageInstalled(ctx, pkg)
		if err != nil {
			result.Status = StatusError
			result.Message = fmt.Sprintf("Error checking package %s: %v", pkg, err)
			result.Duration = time.Since(start)
			return result
		}

		if test.State == "present" && !installed {
			result.Status = StatusFail
			result.Message = fmt.Sprintf("Package %s is not installed", pkg)
			result.Details[pkg] = "not installed"
		} else if test.State == "absent" && installed {
			result.Status = StatusFail
			result.Message = fmt.Sprintf("Package %s is installed but should be absent", pkg)
			result.Details[pkg] = fmt.Sprintf("installed (version: %s)", version)
		} else {
			result.Details[pkg] = version
		}

		if result.Status == StatusFail {
			break
		}
	}

	if result.Status == StatusPass {
		if test.State == "present" {
			result.Message = fmt.Sprintf("All %d packages are installed", len(test.Packages))
		} else {
			result.Message = fmt.Sprintf("All %d packages are absent as expected", len(test.Packages))
		}
	}

	result.Duration = time.Since(start)
	return result
}

// executeFileTest executes a file test
func (e *Executor) executeFileTest(ctx context.Context, test FileTest) Result {
	start := time.Now()
	result := Result{
		Name:    test.Name,
		Status:  StatusPass,
		Details: make(map[string]interface{}),
	}

	// Implementation moved from assertions/file.go
	stdout, _, exitCode, err := e.provider.ExecuteCommand(ctx, fmt.Sprintf("stat -c '%%F:%%U:%%G:%%a' %s 2>/dev/null || echo 'notfound'", test.Path))
	if err != nil {
		result.Status = StatusError
		result.Message = fmt.Sprintf("Error checking path %s: %v", test.Path, err)
		result.Duration = time.Since(start)
		return result
	}

	stdout = strings.TrimSpace(stdout)
	if stdout == "notfound" || exitCode != 0 {
		result.Status = StatusFail
		result.Message = fmt.Sprintf("Path %s does not exist", test.Path)
		result.Duration = time.Since(start)
		return result
	}

	parts := strings.Split(stdout, ":")
	if len(parts) != 4 {
		result.Status = StatusError
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

	// Validation logic...
	if test.Type != "" && !matchesFileType(fileType, test.Type) {
		result.Status = StatusFail
		result.Message = fmt.Sprintf("Path %s is a %s, expected %s", test.Path, fileType, test.Type)
		result.Duration = time.Since(start)
		return result
	}

	if test.Owner != "" && owner != test.Owner {
		result.Status = StatusFail
		result.Message = fmt.Sprintf("Path %s owner is %s, expected %s", test.Path, owner, test.Owner)
		result.Duration = time.Since(start)
		return result
	}

	if test.Group != "" && group != test.Group {
		result.Status = StatusFail
		result.Message = fmt.Sprintf("Path %s group is %s, expected %s", test.Path, group, test.Group)
		result.Duration = time.Since(start)
		return result
	}

	if test.Mode != "" {
		expectedMode := normalizeMode(test.Mode)
		if mode != expectedMode {
			result.Status = StatusFail
			result.Message = fmt.Sprintf("Path %s mode is %s, expected %s", test.Path, mode, expectedMode)
			result.Duration = time.Since(start)
			return result
		}
	}

	result.Message = fmt.Sprintf("Path %s exists with correct properties", test.Path)
	result.Duration = time.Since(start)
	return result
}

// executeServiceTest executes a service test
func (e *Executor) executeServiceTest(ctx context.Context, test ServiceTest) Result {
	start := time.Now()
	result := Result{
		Name:    test.Name,
		Status:  StatusPass,
		Details: make(map[string]interface{}),
	}

	// Get list of services to check
	services := test.Services
	if test.Service != "" {
		services = []string{test.Service}
	}

	for _, service := range services {
		// Check service status using systemctl (systemd)
		running, enabled, err := e.checkServiceStatus(ctx, service)
		if err != nil {
			result.Status = StatusError
			result.Message = fmt.Sprintf("Error checking service %s: %v", service, err)
			result.Duration = time.Since(start)
			return result
		}

		// Check running state
		if test.State == "running" && !running {
			result.Status = StatusFail
			result.Message = fmt.Sprintf("Service %s is not running", service)
			result.Details[service] = "not running"
			break
		} else if test.State == "stopped" && running {
			result.Status = StatusFail
			result.Message = fmt.Sprintf("Service %s is running but should be stopped", service)
			result.Details[service] = "running"
			break
		}

		// Check enabled state if specified
		if test.Enabled && !enabled {
			result.Status = StatusFail
			result.Message = fmt.Sprintf("Service %s is not enabled", service)
			result.Details[service] = "not enabled"
			break
		} else if !test.Enabled && test.State == "stopped" && enabled {
			// Only fail on enabled if state is stopped and we explicitly don't want it enabled
			result.Status = StatusFail
			result.Message = fmt.Sprintf("Service %s is enabled but should be disabled", service)
			result.Details[service] = "enabled"
			break
		}

		// Record status
		status := "stopped"
		if running {
			status = "running"
		}
		if enabled {
			status += " (enabled)"
		}
		result.Details[service] = status
	}

	if result.Status == StatusPass {
		if test.State == "running" {
			result.Message = fmt.Sprintf("All %d services are running", len(services))
		} else {
			result.Message = fmt.Sprintf("All %d services are stopped", len(services))
		}
		if test.Enabled {
			result.Message += " and enabled"
		}
	}

	result.Duration = time.Since(start)
	return result
}

// Helper functions

// checkServiceStatus checks if a service is running and enabled
func (e *Executor) checkServiceStatus(ctx context.Context, service string) (running, enabled bool, err error) {
	// Try systemctl (systemd)
	stdout, _, exitCode, err := e.provider.ExecuteCommand(ctx, fmt.Sprintf("systemctl is-active %s 2>/dev/null", service))
	if err != nil {
		return false, false, err
	}

	stdout = strings.TrimSpace(stdout)
	running = (exitCode == 0 && stdout == "active")

	// Check if enabled
	stdout, _, exitCode, err = e.provider.ExecuteCommand(ctx, fmt.Sprintf("systemctl is-enabled %s 2>/dev/null", service))
	if err != nil {
		return running, false, nil // Don't fail if we can't check enabled status
	}

	stdout = strings.TrimSpace(stdout)
	enabled = (exitCode == 0 && stdout == "enabled")

	return running, enabled, nil
}
func (e *Executor) isPackageInstalled(ctx context.Context, pkg string) (bool, string, error) {
	// Try dpkg
	stdout, _, exitCode, err := e.provider.ExecuteCommand(ctx, fmt.Sprintf("dpkg -l %s 2>/dev/null | grep '^ii'", pkg))
	if err != nil {
		return false, "", err
	}
	if exitCode == 0 && stdout != "" {
		version := extractDpkgVersion(stdout)
		return true, version, nil
	}

	// Try rpm
	stdout, _, exitCode, err = e.provider.ExecuteCommand(ctx, fmt.Sprintf("rpm -q %s 2>/dev/null", pkg))
	if err != nil {
		return false, "", err
	}
	if exitCode == 0 && stdout != "" {
		return true, strings.TrimSpace(stdout), nil
	}

	// Try apk
	stdout, _, exitCode, err = e.provider.ExecuteCommand(ctx, fmt.Sprintf("apk info -e %s 2>/dev/null", pkg))
	if err != nil {
		return false, "", err
	}
	if exitCode == 0 && stdout != "" {
		return true, strings.TrimSpace(stdout), nil
	}

	return false, "", nil
}

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

func matchesFileType(actual, expected string) bool {
	return strings.ToLower(actual) == strings.ToLower(expected)
}

func normalizeMode(mode string) string {
	mode = strings.TrimPrefix(mode, "0o")
	mode = strings.TrimPrefix(mode, "0")
	return mode
}

// executeUserTest executes a user test
func (e *Executor) executeUserTest(ctx context.Context, test UserTest) Result {
	start := time.Now()
	result := Result{
		Name:    test.Name,
		Status:  StatusPass,
		Details: make(map[string]interface{}),
	}

	// Check if user exists
	userInfo, err := e.getUserInfo(ctx, test.User)
	if err != nil {
		result.Status = StatusError
		result.Message = fmt.Sprintf("Error checking user %s: %v", test.User, err)
		result.Duration = time.Since(start)
		return result
	}

	if userInfo == nil {
		result.Status = StatusFail
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
		result.Status = StatusFail
		result.Message = fmt.Sprintf("User %s shell is %s, expected %s", test.User, userInfo["shell"], test.Shell)
		result.Duration = time.Since(start)
		return result
	}

	// Check home if specified
	if test.Home != "" && userInfo["home"] != test.Home {
		result.Status = StatusFail
		result.Message = fmt.Sprintf("User %s home is %s, expected %s", test.User, userInfo["home"], test.Home)
		result.Duration = time.Since(start)
		return result
	}

	// Check groups if specified
	if len(test.Groups) > 0 {
		userGroups, err := e.getUserGroups(ctx, test.User)
		if err != nil {
			result.Status = StatusError
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
				result.Status = StatusFail
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

// executeGroupTest executes a group test
func (e *Executor) executeGroupTest(ctx context.Context, test GroupTest) Result {
	start := time.Now()
	result := Result{
		Name:    test.Name,
		Status:  StatusPass,
		Details: make(map[string]interface{}),
	}

	for _, group := range test.Groups {
		exists, gid, err := e.groupExists(ctx, group)
		if err != nil {
			result.Status = StatusError
			result.Message = fmt.Sprintf("Error checking group %s: %v", group, err)
			result.Duration = time.Since(start)
			return result
		}

		if test.State == "present" && !exists {
			result.Status = StatusFail
			result.Message = fmt.Sprintf("Group %s does not exist", group)
			result.Details[group] = "not found"
			break
		} else if test.State == "absent" && exists {
			result.Status = StatusFail
			result.Message = fmt.Sprintf("Group %s exists but should be absent", group)
			result.Details[group] = fmt.Sprintf("gid: %s", gid)
			break
		} else if exists {
			result.Details[group] = fmt.Sprintf("gid: %s", gid)
		} else {
			result.Details[group] = "absent"
		}
	}

	if result.Status == StatusPass {
		if test.State == "present" {
			result.Message = fmt.Sprintf("All %d groups exist", len(test.Groups))
		} else {
			result.Message = fmt.Sprintf("All %d groups are absent as expected", len(test.Groups))
		}
	}

	result.Duration = time.Since(start)
	return result
}

// getUserInfo gets information about a user
func (e *Executor) getUserInfo(ctx context.Context, username string) (map[string]string, error) {
	stdout, _, exitCode, err := e.provider.ExecuteCommand(ctx, fmt.Sprintf("id -u %s 2>/dev/null && id -g %s 2>/dev/null && getent passwd %s 2>/dev/null", username, username, username))
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
func (e *Executor) getUserGroups(ctx context.Context, username string) ([]string, error) {
	stdout, _, exitCode, err := e.provider.ExecuteCommand(ctx, fmt.Sprintf("id -Gn %s 2>/dev/null", username))
	if err != nil {
		return nil, err
	}

	if exitCode != 0 {
		return nil, fmt.Errorf("failed to get groups for user %s", username)
	}

	groups := strings.Fields(strings.TrimSpace(stdout))
	return groups, nil
}

// groupExists checks if a group exists
func (e *Executor) groupExists(ctx context.Context, groupname string) (bool, string, error) {
	stdout, _, exitCode, err := e.provider.ExecuteCommand(ctx, fmt.Sprintf("getent group %s 2>/dev/null", groupname))
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

// executeFileContentTest executes a file content test
func (e *Executor) executeFileContentTest(ctx context.Context, test FileContentTest) Result {
	start := time.Now()
	result := Result{
		Name:    test.Name,
		Status:  StatusPass,
		Details: make(map[string]interface{}),
	}

	// First check if file exists and is readable
	_, _, exitCode, err := e.provider.ExecuteCommand(ctx, fmt.Sprintf("test -f %s && test -r %s", test.Path, test.Path))
	if err != nil {
		result.Status = StatusError
		result.Message = fmt.Sprintf("Error checking file %s: %v", test.Path, err)
		result.Duration = time.Since(start)
		return result
	}

	if exitCode != 0 {
		result.Status = StatusFail
		result.Message = fmt.Sprintf("File %s does not exist or is not readable", test.Path)
		result.Duration = time.Since(start)
		return result
	}

	// Check contains strings
	if len(test.Contains) > 0 {
		for _, searchStr := range test.Contains {
			// Use grep -F for fixed string matching (no regex interpretation)
			_, _, exitCode, err := e.provider.ExecuteCommand(ctx, fmt.Sprintf("grep -F %s %s >/dev/null 2>&1", shellQuote(searchStr), test.Path))
			if err != nil {
				result.Status = StatusError
				result.Message = fmt.Sprintf("Error searching for '%s' in %s: %v", searchStr, test.Path, err)
				result.Duration = time.Since(start)
				return result
			}

			if exitCode != 0 {
				result.Status = StatusFail
				result.Message = fmt.Sprintf("File %s does not contain '%s'", test.Path, searchStr)
				result.Details["missing"] = searchStr
				result.Duration = time.Since(start)
				return result
			}

			result.Details[searchStr] = "found"
		}
	}

	// Check matches regex pattern
	if test.Matches != "" {
		// Use grep -E for extended regex
		_, _, exitCode, err := e.provider.ExecuteCommand(ctx, fmt.Sprintf("grep -E %s %s >/dev/null 2>&1", shellQuote(test.Matches), test.Path))
		if err != nil {
			result.Status = StatusError
			result.Message = fmt.Sprintf("Error matching pattern in %s: %v", test.Path, err)
			result.Duration = time.Since(start)
			return result
		}

		if exitCode != 0 {
			result.Status = StatusFail
			result.Message = fmt.Sprintf("File %s does not match pattern '%s'", test.Path, test.Matches)
			result.Details["pattern"] = test.Matches
			result.Duration = time.Since(start)
			return result
		}

		result.Details["pattern"] = "matched"
	}

	if len(test.Contains) > 0 && test.Matches != "" {
		result.Message = fmt.Sprintf("File %s contains all %d strings and matches pattern", test.Path, len(test.Contains))
	} else if len(test.Contains) > 0 {
		result.Message = fmt.Sprintf("File %s contains all %d strings", test.Path, len(test.Contains))
	} else if test.Matches != "" {
		result.Message = fmt.Sprintf("File %s matches pattern", test.Path)
	}

	result.Duration = time.Since(start)
	return result
}

// shellQuote quotes a string for safe use in shell commands
func shellQuote(s string) string {
	// Simple shell quoting - escape single quotes and wrap in single quotes
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// executeCommandContentTest executes a command content test
func (e *Executor) executeCommandContentTest(ctx context.Context, test CommandContentTest) Result {
	start := time.Now()
	result := Result{
		Name:    test.Name,
		Status:  StatusPass,
		Details: make(map[string]interface{}),
	}

	// Execute the command
	stdout, stderr, exitCode, err := e.provider.ExecuteCommand(ctx, test.Command)
	if err != nil {
		result.Status = StatusError
		result.Message = fmt.Sprintf("Error executing command '%s': %v", test.Command, err)
		result.Duration = time.Since(start)
		return result
	}

	result.Details["exit_code"] = exitCode
	result.Details["stdout_length"] = len(stdout)
	result.Details["stderr_length"] = len(stderr)

	// Check exit code if specified
	if test.ExitCode != 0 && exitCode != test.ExitCode {
		result.Status = StatusFail
		result.Message = fmt.Sprintf("Command exit code is %d, expected %d", exitCode, test.ExitCode)
		result.Duration = time.Since(start)
		return result
	}

	// Check contains strings in stdout
	if len(test.Contains) > 0 {
		for _, searchStr := range test.Contains {
			if !strings.Contains(stdout, searchStr) {
				result.Status = StatusFail
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

// executeDockerTest executes a Docker container test
func (e *Executor) executeDockerTest(ctx context.Context, test DockerTest) Result {
	start := time.Now()
	result := Result{
		Name:    test.Name,
		Status:  StatusPass,
		Details: make(map[string]interface{}),
	}

	// Get list of containers to check
	containers := test.Containers
	if test.Container != "" {
		containers = []string{test.Container}
	}

	for _, container := range containers {
		// Use docker inspect to get container details
		// Format: {{.State.Status}}|{{.Config.Image}}|{{.HostConfig.RestartPolicy.Name}}|{{.State.Health.Status}}
		inspectCmd := fmt.Sprintf("docker inspect --format '{{.State.Status}}|{{.Config.Image}}|{{.HostConfig.RestartPolicy.Name}}|{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' %s 2>/dev/null", container)
		stdout, _, exitCode, err := e.provider.ExecuteCommand(ctx, inspectCmd)
		if err != nil {
			result.Status = StatusError
			result.Message = fmt.Sprintf("Error inspecting container %s: %v", container, err)
			result.Duration = time.Since(start)
			return result
		}

		if exitCode != 0 {
			// Container doesn't exist
			if test.State == "exists" || test.State == "running" || test.State == "stopped" {
				result.Status = StatusFail
				result.Message = fmt.Sprintf("Container %s does not exist", container)
				result.Details[container] = "not found"
				break
			}
		} else {
			// Parse docker inspect output
			stdout = strings.TrimSpace(stdout)
			parts := strings.Split(stdout, "|")
			if len(parts) != 4 {
				result.Status = StatusError
				result.Message = fmt.Sprintf("Unexpected docker inspect output for %s", container)
				result.Duration = time.Since(start)
				return result
			}

			status := parts[0]        // running, exited, created, paused, etc.
			image := parts[1]         // image name
			restartPolicy := parts[2] // no, always, on-failure, unless-stopped
			health := parts[3]        // healthy, unhealthy, starting, none

			// Check state
			if test.State == "running" && status != "running" {
				result.Status = StatusFail
				result.Message = fmt.Sprintf("Container %s is %s, expected running", container, status)
				result.Details[container] = status
				break
			} else if test.State == "stopped" && status == "running" {
				result.Status = StatusFail
				result.Message = fmt.Sprintf("Container %s is running, expected stopped", container)
				result.Details[container] = status
				break
			}

			// Check image if specified
			if test.Image != "" && !strings.Contains(image, test.Image) {
				result.Status = StatusFail
				result.Message = fmt.Sprintf("Container %s image is %s, expected %s", container, image, test.Image)
				result.Details[container] = fmt.Sprintf("image: %s", image)
				break
			}

			// Check restart policy if specified
			if test.RestartPolicy != "" && restartPolicy != test.RestartPolicy {
				result.Status = StatusFail
				result.Message = fmt.Sprintf("Container %s restart policy is %s, expected %s", container, restartPolicy, test.RestartPolicy)
				result.Details[container] = fmt.Sprintf("restart: %s", restartPolicy)
				break
			}

			// Check health if specified
			if test.Health != "" && health != test.Health {
				result.Status = StatusFail
				result.Message = fmt.Sprintf("Container %s health is %s, expected %s", container, health, test.Health)
				result.Details[container] = fmt.Sprintf("health: %s", health)
				break
			}

			// Record container details
			containerInfo := fmt.Sprintf("status: %s", status)
			if test.Image != "" {
				containerInfo += fmt.Sprintf(", image: %s", image)
			}
			if test.RestartPolicy != "" {
				containerInfo += fmt.Sprintf(", restart: %s", restartPolicy)
			}
			if test.Health != "" {
				containerInfo += fmt.Sprintf(", health: %s", health)
			}
			result.Details[container] = containerInfo
		}
	}

	// Build success message if all checks passed
	if result.Status == StatusPass {
		if test.State == "running" {
			result.Message = fmt.Sprintf("All %d containers are running", len(containers))
		} else if test.State == "stopped" {
			result.Message = fmt.Sprintf("All %d containers are stopped", len(containers))
		} else {
			result.Message = fmt.Sprintf("All %d containers exist", len(containers))
		}
		if test.Image != "" {
			result.Message += fmt.Sprintf(" with correct image")
		}
		if test.RestartPolicy != "" {
			result.Message += fmt.Sprintf(" and restart policy")
		}
		if test.Health != "" {
			result.Message += fmt.Sprintf(" and health status")
		}
	}

	result.Duration = time.Since(start)
	return result
}
