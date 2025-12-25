package core_test

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestMockProvider(t *testing.T) {
	mock := core.NewMockProvider()
	if mock == nil {
		t.Fatal("NewMockProvider returned nil")
	}

	// Test SetCommandResult and ExecuteCommand
	mock.SetCommandResult("test command", "test output", "test error", 42, nil)

	ctx := context.Background()
	stdout, stderr, exitCode, err := mock.ExecuteCommand(ctx, "test command")

	if stdout != "test output" {
		t.Errorf("Expected stdout 'test output', got %q", stdout)
	}

	if stderr != "test error" {
		t.Errorf("Expected stderr 'test error', got %q", stderr)
	}

	if exitCode != 42 {
		t.Errorf("Expected exit code 42, got %d", exitCode)
	}

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestMockProvider_UnknownCommand(t *testing.T) {
	mock := core.NewMockProvider()
	ctx := context.Background()

	stdout, stderr, exitCode, err := mock.ExecuteCommand(ctx, "unknown command")

	if stdout != "" {
		t.Errorf("Expected empty stdout, got %q", stdout)
	}

	if stderr != "" {
		t.Errorf("Expected empty stderr, got %q", stderr)
	}

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestMockProvider_MultipleCommands(t *testing.T) {
	mock := core.NewMockProvider()

	mock.SetCommandResult("cmd1", "out1", "err1", 1, nil)
	mock.SetCommandResult("cmd2", "out2", "err2", 2, nil)
	mock.SetCommandResult("cmd3", "out3", "err3", 3, nil)

	ctx := context.Background()

	// Test cmd1
	stdout, stderr, exitCode, err := mock.ExecuteCommand(ctx, "cmd1")
	if stdout != "out1" || stderr != "err1" || exitCode != 1 || err != nil {
		t.Errorf("cmd1 failed: got (%q, %q, %d, %v)", stdout, stderr, exitCode, err)
	}

	// Test cmd2
	stdout, stderr, exitCode, err = mock.ExecuteCommand(ctx, "cmd2")
	if stdout != "out2" || stderr != "err2" || exitCode != 2 || err != nil {
		t.Errorf("cmd2 failed: got (%q, %q, %d, %v)", stdout, stderr, exitCode, err)
	}

	// Test cmd3
	stdout, stderr, exitCode, err = mock.ExecuteCommand(ctx, "cmd3")
	if stdout != "out3" || stderr != "err3" || exitCode != 3 || err != nil {
		t.Errorf("cmd3 failed: got (%q, %q, %d, %v)", stdout, stderr, exitCode, err)
	}
}
