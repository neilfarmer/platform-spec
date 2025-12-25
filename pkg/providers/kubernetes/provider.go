package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// Provider implements Kubernetes testing via kubectl
type Provider struct {
	config *Config
}

// Config holds Kubernetes provider configuration
type Config struct {
	Kubeconfig string
	Context    string
	Namespace  string // default namespace
}

// NewProvider creates a new Kubernetes provider
func NewProvider(config *Config) *Provider {
	return &Provider{
		config: config,
	}
}

// ExecuteCommand executes a kubectl command and returns stdout, stderr, and exit code
func (p *Provider) ExecuteCommand(ctx context.Context, command string) (stdout, stderr string, exitCode int, err error) {
	// Build environment with KUBECONFIG
	env := os.Environ()
	if p.config.Kubeconfig != "" {
		env = append(env, fmt.Sprintf("KUBECONFIG=%s", p.config.Kubeconfig))
	}

	// If context is specified, inject --context into kubectl commands
	if p.config.Context != "" && len(command) >= 7 && command[:7] == "kubectl" {
		command = fmt.Sprintf("kubectl --context=%s%s", p.config.Context, command[7:])
	}

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Env = env

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
				err = nil
			} else {
				return stdout, stderr, -1, fmt.Errorf("failed to get exit status: %w", err)
			}
		} else {
			return stdout, stderr, -1, fmt.Errorf("command execution failed: %w", err)
		}
	} else {
		exitCode = 0
	}

	return stdout, stderr, exitCode, nil
}
