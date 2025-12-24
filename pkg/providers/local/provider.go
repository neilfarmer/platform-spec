package local

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"syscall"
)

// Provider implements local system testing
type Provider struct{}

// NewProvider creates a new local provider
func NewProvider() *Provider {
	return &Provider{}
}

// ExecuteCommand executes a command on the local system and returns stdout, stderr, and exit code
func (p *Provider) ExecuteCommand(ctx context.Context, command string) (stdout, stderr string, exitCode int, err error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)

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
