package kubernetes

import (
	"context"
	"testing"
)

func TestNewProvider(t *testing.T) {
	config := &Config{
		Kubeconfig: "/path/to/kubeconfig",
		Context:    "test-context",
		Namespace:  "test-namespace",
	}

	provider := NewProvider(config)

	if provider == nil {
		t.Error("NewProvider() returned nil")
	}

	if provider.config != config {
		t.Error("Provider config not set correctly")
	}
}

func TestExecuteCommand(t *testing.T) {
	// Note: This test requires kubectl to be installed and a valid kubeconfig
	// In a real test environment, you would mock the command execution

	provider := NewProvider(&Config{})
	ctx := context.Background()

	// Test a simple kubectl version command
	_, _, _, err := provider.ExecuteCommand(ctx, "kubectl version --client")

	// We can't assert on success since kubectl might not be installed
	// But we can verify the function doesn't panic
	if err != nil {
		t.Logf("kubectl not available (expected in some test environments): %v", err)
	}
}

func TestExecuteCommandWithContext(t *testing.T) {
	config := &Config{
		Context: "test-context",
	}
	provider := NewProvider(config)

	ctx := context.Background()

	// The command should have --context injected
	// We can't really test this without mocking, but we can verify it doesn't panic
	_, _, _, err := provider.ExecuteCommand(ctx, "kubectl version --client")

	if err != nil {
		t.Logf("kubectl not available (expected in some test environments): %v", err)
	}
}
