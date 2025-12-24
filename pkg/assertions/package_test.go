package assertions

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

type mockExecutor struct {
	commands map[string]mockResult
}

type mockResult struct {
	stdout   string
	stderr   string
	exitCode int
	err      error
}

func newMockExecutor() *mockExecutor {
	return &mockExecutor{
		commands: make(map[string]mockResult),
	}
}

func (m *mockExecutor) setResult(command string, stdout, stderr string, exitCode int, err error) {
	m.commands[command] = mockResult{stdout, stderr, exitCode, err}
}

func (m *mockExecutor) ExecuteCommand(ctx context.Context, command string) (string, string, int, error) {
	if result, ok := m.commands[command]; ok {
		return result.stdout, result.stderr, result.exitCode, result.err
	}
	return "", "", 0, nil
}

func TestCheckPackages(t *testing.T) {
	tests := []struct {
		name       string
		test       core.PackageTest
		setupMock  func(*mockExecutor)
		wantStatus core.Status
	}{
		{
			name: "package installed via dpkg",
			test: core.PackageTest{
				Name:     "Docker installed",
				Packages: []string{"docker-ce"},
				State:    "present",
			},
			setupMock: func(m *mockExecutor) {
				m.setResult("dpkg -l docker-ce 2>/dev/null | grep '^ii'", "ii  docker-ce  5:24.0.7", "", 0, nil)
			},
			wantStatus: core.StatusPass,
		},
		{
			name: "package not installed",
			test: core.PackageTest{
				Name:     "Docker installed",
				Packages: []string{"docker-ce"},
				State:    "present",
			},
			setupMock: func(m *mockExecutor) {
				m.setResult("dpkg -l docker-ce 2>/dev/null | grep '^ii'", "", "", 1, nil)
				m.setResult("rpm -q docker-ce 2>/dev/null", "", "", 1, nil)
				m.setResult("apk info -e docker-ce 2>/dev/null", "", "", 1, nil)
			},
			wantStatus: core.StatusFail,
		},
		{
			name: "package absent - success",
			test: core.PackageTest{
				Name:     "Telnet removed",
				Packages: []string{"telnet"},
				State:    "absent",
			},
			setupMock: func(m *mockExecutor) {
				m.setResult("dpkg -l telnet 2>/dev/null | grep '^ii'", "", "", 1, nil)
				m.setResult("rpm -q telnet 2>/dev/null", "", "", 1, nil)
				m.setResult("apk info -e telnet 2>/dev/null", "", "", 1, nil)
			},
			wantStatus: core.StatusPass,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			tt.setupMock(mock)

			ctx := context.Background()
			result := CheckPackages(ctx, mock, tt.test)

			if result.Status != tt.wantStatus {
				t.Errorf("Status = %v, want %v", result.Status, tt.wantStatus)
			}

			if result.Name != tt.test.Name {
				t.Errorf("Name = %v, want %v", result.Name, tt.test.Name)
			}

			if result.Duration == 0 {
				t.Error("Duration should be set")
			}
		})
	}
}

func TestExtractDpkgVersion(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{
			name:   "standard dpkg output",
			output: "ii  docker-ce  5:24.0.7-1~ubuntu  amd64  Docker",
			want:   "5:24.0.7-1~ubuntu",
		},
		{
			name:   "empty output",
			output: "",
			want:   "unknown",
		},
		{
			name:   "malformed output",
			output: "ii docker-ce",
			want:   "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractDpkgVersion(tt.output)
			if got != tt.want {
				t.Errorf("extractDpkgVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
