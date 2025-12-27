package core_test

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
	k8splugin "github.com/neilfarmer/platform-spec/pkg/core/kubernetes"
	"github.com/neilfarmer/platform-spec/pkg/core/system"
)

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

func NewMockProvider() *MockProvider {
	return &MockProvider{
		commands: make(map[string]mockCommandResult),
	}
}

func (m *MockProvider) SetCommandResult(command string, stdout, stderr string, exitCode int, err error) {
	m.commands[command] = mockCommandResult{
		stdout:   stdout,
		stderr:   stderr,
		exitCode: exitCode,
		err:      err,
	}
}

func (m *MockProvider) ExecuteCommand(ctx context.Context, command string) (stdout, stderr string, exitCode int, err error) {
	if result, ok := m.commands[command]; ok {
		return result.stdout, result.stderr, result.exitCode, result.err
	}
	return "", "", 0, nil
}

func TestExecutor_FailFast(t *testing.T) {
	mock := NewMockProvider()
	mock.SetCommandResult("dpkg -l package1 2>/dev/null | grep '^ii'", "", "", 1, nil)
	mock.SetCommandResult("rpm -q package1 2>/dev/null", "", "", 1, nil)
	mock.SetCommandResult("apk info -e package1 2>/dev/null", "", "", 1, nil)

	spec := &core.Spec{
		Config: core.SpecConfig{
			FailFast: true,
		},
		Tests: core.Tests{
			Packages: []core.PackageTest{
				{Name: "Test 1", Packages: []string{"package1"}, State: "present"},
				{Name: "Test 2", Packages: []string{"package2"}, State: "present"},
				{Name: "Test 3", Packages: []string{"package3"}, State: "present"},
			},
		},
	}

	executor := core.NewExecutor(spec, mock, system.NewSystemPlugin(), k8splugin.NewKubernetesPlugin())
	ctx := context.Background()

	results, err := executor.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// With fail_fast, should stop after first failure
	if len(results.Results) != 1 {
		t.Errorf("Expected 1 result (fail fast), got %d", len(results.Results))
	}

	if results.Results[0].Status != core.StatusFail {
		t.Errorf("First result should be failed")
	}
}

func TestExecutor_MultipleTests(t *testing.T) {
	mock := NewMockProvider()
	mock.SetCommandResult("dpkg -l docker-ce 2>/dev/null | grep '^ii'", "ii  docker-ce", "", 0, nil)
	mock.SetCommandResult("stat -c '%F:%U:%G:%a' /opt/app 2>/dev/null || echo 'notfound'", "directory:root:root:755", "", 0, nil)

	spec := &core.Spec{
		Metadata: core.SpecMetadata{
			Name: "Multi Test",
		},
		Tests: core.Tests{
			Packages: []core.PackageTest{
				{Name: "Docker installed", Packages: []string{"docker-ce"}, State: "present"},
			},
			Files: []core.FileTest{
				{Name: "App dir", Path: "/opt/app", Type: "directory"},
			},
		},
	}

	executor := core.NewExecutor(spec, mock, system.NewSystemPlugin(), k8splugin.NewKubernetesPlugin())
	ctx := context.Background()

	results, err := executor.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(results.Results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results.Results))
	}

	if results.SpecName != "Multi Test" {
		t.Errorf("SpecName = %v, want Multi Test", results.SpecName)
	}

	if results.Duration == 0 {
		t.Error("Duration should be set")
	}

	if results.StartTime.IsZero() {
		t.Error("StartTime should be set")
	}

	for _, r := range results.Results {
		if r.Status != core.StatusPass {
			t.Errorf("Test %q failed: %v", r.Name, r.Message)
		}
	}
}




func TestNewExecutor(t *testing.T) {
	spec := &core.Spec{}
	mock := NewMockProvider()

	executor := core.NewExecutor(spec, mock, system.NewSystemPlugin(), k8splugin.NewKubernetesPlugin())

	if executor == nil {
		t.Fatal("NewExecutor returned nil")
	}

	// Note: We can't access unexported fields (spec, provider, plugins) from an external test package
	// The fact that NewExecutor returns a non-nil executor is sufficient to verify construction
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
