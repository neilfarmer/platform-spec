package assertions

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestCheckFile(t *testing.T) {
	tests := []struct {
		name       string
		test       core.FileTest
		setupMock  func(*mockExecutor)
		wantStatus core.Status
	}{
		{
			name: "directory exists with correct properties",
			test: core.FileTest{
				Name:  "App directory",
				Path:  "/opt/app",
				Type:  "directory",
				Owner: "appuser",
				Group: "appuser",
				Mode:  "0755",
			},
			setupMock: func(m *mockExecutor) {
				m.setResult("stat -c '%F:%U:%G:%a' /opt/app 2>/dev/null || echo 'notfound'", "directory:appuser:appuser:755", "", 0, nil)
			},
			wantStatus: core.StatusPass,
		},
		{
			name: "file does not exist",
			test: core.FileTest{
				Name: "Config file",
				Path: "/etc/app/config.yml",
				Type: "file",
			},
			setupMock: func(m *mockExecutor) {
				m.setResult("stat -c '%F:%U:%G:%a' /etc/app/config.yml 2>/dev/null || echo 'notfound'", "notfound", "", 0, nil)
			},
			wantStatus: core.StatusFail,
		},
		{
			name: "wrong file type",
			test: core.FileTest{
				Name: "Should be directory",
				Path: "/opt/app",
				Type: "directory",
			},
			setupMock: func(m *mockExecutor) {
				m.setResult("stat -c '%F:%U:%G:%a' /opt/app 2>/dev/null || echo 'notfound'", "regular file:root:root:644", "", 0, nil)
			},
			wantStatus: core.StatusFail,
		},
		{
			name: "wrong owner",
			test: core.FileTest{
				Name:  "Config file",
				Path:  "/etc/app/config",
				Type:  "file",
				Owner: "appuser",
			},
			setupMock: func(m *mockExecutor) {
				m.setResult("stat -c '%F:%U:%G:%a' /etc/app/config 2>/dev/null || echo 'notfound'", "regular file:root:root:644", "", 0, nil)
			},
			wantStatus: core.StatusFail,
		},
		{
			name: "wrong permissions",
			test: core.FileTest{
				Name: "Private directory",
				Path: "/opt/secrets",
				Type: "directory",
				Mode: "0700",
			},
			setupMock: func(m *mockExecutor) {
				m.setResult("stat -c '%F:%U:%G:%a' /opt/secrets 2>/dev/null || echo 'notfound'", "directory:user:user:755", "", 0, nil)
			},
			wantStatus: core.StatusFail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			tt.setupMock(mock)

			ctx := context.Background()
			result := CheckFile(ctx, mock, tt.test)

			if result.Status != tt.wantStatus {
				t.Errorf("Status = %v, want %v (message: %s)", result.Status, tt.wantStatus, result.Message)
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

func TestNormalizeFileType(t *testing.T) {
	tests := []struct {
		statType string
		want     string
	}{
		{"directory", "directory"},
		{"regular file", "file"},
		{"regular empty file", "file"},
		{"symbolic link", "symlink"},
		{"Directory", "directory"},
		{"REGULAR FILE", "file"},
	}

	for _, tt := range tests {
		t.Run(tt.statType, func(t *testing.T) {
			got := normalizeFileType(tt.statType)
			if got != tt.want {
				t.Errorf("normalizeFileType(%q) = %v, want %v", tt.statType, got, tt.want)
			}
		})
	}
}

func TestMatchesFileType(t *testing.T) {
	tests := []struct {
		actual   string
		expected string
		want     bool
	}{
		{"directory", "directory", true},
		{"file", "file", true},
		{"symlink", "symlink", true},
		{"Directory", "directory", true},
		{"file", "directory", false},
		{"directory", "file", false},
	}

	for _, tt := range tests {
		t.Run(tt.actual+"_"+tt.expected, func(t *testing.T) {
			got := matchesFileType(tt.actual, tt.expected)
			if got != tt.want {
				t.Errorf("matchesFileType(%q, %q) = %v, want %v", tt.actual, tt.expected, got, tt.want)
			}
		})
	}
}

func TestNormalizeMode(t *testing.T) {
	tests := []struct {
		mode string
		want string
	}{
		{"0755", "755"},
		{"755", "755"},
		{"0o755", "755"},
		{"0644", "644"},
		{"1777", "1777"},
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			got := normalizeMode(tt.mode)
			if got != tt.want {
				t.Errorf("normalizeMode(%q) = %v, want %v", tt.mode, got, tt.want)
			}
		})
	}
}
