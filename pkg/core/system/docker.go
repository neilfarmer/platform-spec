package system

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// executeDockerTest executes a Docker container test
func executeDockerTest(ctx context.Context, provider core.Provider, test core.DockerTest) core.Result {
	start := time.Now()
	result := core.Result{
		Name:    test.Name,
		Status:  core.StatusPass,
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
		stdout, _, exitCode, err := provider.ExecuteCommand(ctx, inspectCmd)
		if err != nil {
			result.Status = core.StatusError
			result.Message = fmt.Sprintf("Error inspecting container %s: %v", container, err)
			result.Duration = time.Since(start)
			return result
		}

		if exitCode != 0 {
			// Container doesn't exist
			if test.State == "exists" || test.State == "running" || test.State == "stopped" {
				result.Status = core.StatusFail
				result.Message = fmt.Sprintf("Container %s does not exist", container)
				result.Details[container] = "not found"
				break
			}
		} else {
			// Parse docker inspect output
			stdout = strings.TrimSpace(stdout)
			parts := strings.Split(stdout, "|")
			if len(parts) != 4 {
				result.Status = core.StatusError
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
				result.Status = core.StatusFail
				result.Message = fmt.Sprintf("Container %s is %s, expected running", container, status)
				result.Details[container] = status
				break
			} else if test.State == "stopped" && status == "running" {
				result.Status = core.StatusFail
				result.Message = fmt.Sprintf("Container %s is running, expected stopped", container)
				result.Details[container] = status
				break
			}

			// Check image if specified
			if test.Image != "" && !strings.Contains(image, test.Image) {
				result.Status = core.StatusFail
				result.Message = fmt.Sprintf("Container %s image is %s, expected %s", container, image, test.Image)
				result.Details[container] = fmt.Sprintf("image: %s", image)
				break
			}

			// Check restart policy if specified
			if test.RestartPolicy != "" && restartPolicy != test.RestartPolicy {
				result.Status = core.StatusFail
				result.Message = fmt.Sprintf("Container %s restart policy is %s, expected %s", container, restartPolicy, test.RestartPolicy)
				result.Details[container] = fmt.Sprintf("restart: %s", restartPolicy)
				break
			}

			// Check health if specified
			if test.Health != "" && health != test.Health {
				result.Status = core.StatusFail
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
	if result.Status == core.StatusPass {
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
