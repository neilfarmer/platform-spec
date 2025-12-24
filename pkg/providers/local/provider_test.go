package local

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestNewProvider(t *testing.T) {
	provider := NewProvider()

	if provider == nil {
		t.Fatal("NewProvider() returned nil")
	}
}

func TestExecuteCommand(t *testing.T) {
	tests := []struct {
		name           string
		command        string
		wantStdout     string
		wantExitCode   int
		wantErr        bool
		stdoutContains string
	}{
		{
			name:         "successful command",
			command:      "echo hello",
			wantStdout:   "hello\n",
			wantExitCode: 0,
			wantErr:      false,
		},
		{
			name:         "command with exit code",
			command:      "exit 42",
			wantExitCode: 42,
			wantErr:      false,
		},
		{
			name:         "command that fails",
			command:      "false",
			wantExitCode: 1,
			wantErr:      false,
		},
		{
			name:         "command with output",
			command:      "printf 'line1\\nline2'",
			wantStdout:   "line1\nline2",
			wantExitCode: 0,
			wantErr:      false,
		},
		{
			name:           "command with stderr",
			command:        "echo error >&2",
			wantExitCode:   0,
			wantErr:        false,
			stdoutContains: "",
		},
	}

	provider := NewProvider()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, _, exitCode, err := provider.ExecuteCommand(ctx, tt.command)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if exitCode != tt.wantExitCode {
				t.Errorf("ExecuteCommand() exitCode = %v, want %v", exitCode, tt.wantExitCode)
			}

			if tt.wantStdout != "" && stdout != tt.wantStdout {
				t.Errorf("ExecuteCommand() stdout = %q, want %q", stdout, tt.wantStdout)
			}

			if tt.stdoutContains != "" && !strings.Contains(stdout, tt.stdoutContains) {
				t.Errorf("ExecuteCommand() stdout does not contain %q, got %q", tt.stdoutContains, stdout)
			}
		})
	}
}

func TestExecuteCommandWithContext(t *testing.T) {
	provider := NewProvider()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Command that sleeps longer than the timeout
	_, _, exitCode, err := provider.ExecuteCommand(ctx, "sleep 10")

	// Context timeout should either return error or non-zero exit code
	if err == nil && exitCode == 0 {
		t.Error("ExecuteCommand() should have failed for timed out command")
	}
}
