package core

import "context"

// MockProvider is a mock implementation of the Provider interface for testing
type MockProvider struct {
	commands map[string]mockCommandResult
}

type mockCommandResult struct {
	stdout   string
	stderr   string
	exitCode int
	err      error
}

// NewMockProvider creates a new MockProvider
func NewMockProvider() *MockProvider {
	return &MockProvider{
		commands: make(map[string]mockCommandResult),
	}
}

// SetCommandResult sets the result for a given command
func (m *MockProvider) SetCommandResult(command string, stdout, stderr string, exitCode int, err error) {
	m.commands[command] = mockCommandResult{
		stdout:   stdout,
		stderr:   stderr,
		exitCode: exitCode,
		err:      err,
	}
}

// ExecuteCommand executes a command and returns the mocked result
func (m *MockProvider) ExecuteCommand(ctx context.Context, command string) (stdout, stderr string, exitCode int, err error) {
	if result, ok := m.commands[command]; ok {
		return result.stdout, result.stderr, result.exitCode, result.err
	}
	return "", "", 0, nil
}
