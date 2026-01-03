package system

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestNewSystemPlugin(t *testing.T) {
	plugin := NewSystemPlugin()
	if plugin == nil {
		t.Fatal("NewSystemPlugin returned nil")
	}
}

func TestSystemPlugin_Execute(t *testing.T) {
	mock := core.NewMockProvider()
	// Package test
	mock.SetCommandResult("dpkg -l docker-ce 2>/dev/null | grep '^ii'", "ii  docker-ce", "", 0, nil)
	// File test
	mock.SetCommandResult("stat -c '%F:%U:%G:%a' /etc/hosts 2>/dev/null || echo 'notfound'", "file:root:root:644", "", 0, nil)
	// Service test
	mock.SetCommandResult("systemctl is-active nginx 2>/dev/null", "active", "", 0, nil)
	// User test
	mock.SetCommandResult("id -u testuser 2>/dev/null && id -g testuser 2>/dev/null && getent passwd testuser 2>/dev/null", "1000\n1000\ntestuser:x:1000:1000:Test User:/home/testuser:/bin/bash", "", 0, nil)
	// Group test
	mock.SetCommandResult("getent group testgroup 2>/dev/null", "testgroup:x:1001:", "", 0, nil)
	// File content test
	mock.SetCommandResult("grep -F 'test' /tmp/file 2>/dev/null", "test content", "", 0, nil)
	// Command content test
	mock.SetCommandResult("echo hello", "hello", "", 0, nil)
	// Docker test
	mock.SetCommandResult("docker inspect --format '{{.State.Status}}|{{.Config.Image}}|{{.HostConfig.RestartPolicy.Name}}|{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' testcontainer 2>/dev/null", "running|nginx:latest|always|none", "", 0, nil)
	// Filesystem test
	mock.SetCommandResult("findmnt --noheadings --output TARGET,FSTYPE,OPTIONS,SIZE,USED,USE% --target /mnt 2>/dev/null", "/mnt               ext4   rw,relatime    100G  50G   50%", "", 0, nil)
	// Ping test
	mock.SetCommandResult("ping -c 1 -W 5 8.8.8.8 2>/dev/null", "PING 8.8.8.8", "", 0, nil)
	// DNS test
	mock.SetCommandResult("dig +short 'example.com' 2>/dev/null || getent hosts 'example.com' 2>/dev/null | awk '{print $1}'", "93.184.216.34", "", 0, nil)
	// SystemInfo test - just OS check
	mock.SetCommandResult("grep '^ID=' /etc/os-release 2>/dev/null | cut -d= -f2 | tr -d '\"'", "ubuntu", "", 0, nil)
	// HTTP test
	mock.SetCommandResult("curl -s -w $'\\n%{http_code}' 'http://example.com'", "content\n200", "", 0, nil)
	// Port test
	mock.SetCommandResult("ss -tln | grep -E ':80\\s' || true", "LISTEN    0    128    0.0.0.0:80    0.0.0.0:*", "", 0, nil)

	spec := &core.Spec{
		Tests: core.Tests{
			Packages:       []core.PackageTest{{Name: "Package", Packages: []string{"docker-ce"}, State: "present"}},
			Files:          []core.FileTest{{Name: "File", Path: "/etc/hosts", Type: "file"}},
			Services:       []core.ServiceTest{{Name: "Service", Services: []string{"nginx"}, State: "running"}},
			Users:          []core.UserTest{{Name: "User", User: "testuser"}},
			Groups:         []core.GroupTest{{Name: "Group", Groups: []string{"testgroup"}}},
			FileContent:    []core.FileContentTest{{Name: "FileContent", Path: "/tmp/file", Contains: []string{"test"}}},
			CommandContent: []core.CommandContentTest{{Name: "Command", Command: "echo hello", Contains: []string{"hello"}}},
			Docker:         []core.DockerTest{{Name: "Docker", Containers: []string{"testcontainer"}, State: "running"}},
			Filesystems:    []core.FilesystemTest{{Name: "FS", Path: "/mnt", State: "mounted"}},
			Ping:           []core.PingTest{{Name: "Ping", Host: "8.8.8.8"}},
			DNS:            []core.DNSTest{{Name: "DNS", Host: "example.com"}},
			SystemInfo:     []core.SystemInfoTest{{Name: "SysInfo", OS: "ubuntu"}},
			HTTP:           []core.HTTPTest{{Name: "HTTP", URL: "http://example.com"}},
			Ports:          []core.PortTest{{Name: "Port", Port: 80, Protocol: "tcp", State: "listening"}},
		},
	}

	// Validate spec to set defaults
	if err := spec.Validate(); err != nil {
		t.Fatalf("Spec validation failed: %v", err)
	}

	plugin := NewSystemPlugin()
	ctx := context.Background()

	results, shouldStop := plugin.Execute(ctx, spec, mock, false, nil)

	if len(results) != 14 {
		t.Errorf("Expected 14 results, got %d", len(results))
	}

	if shouldStop {
		t.Error("Should not stop when failFast is false")
	}

	for _, r := range results {
		if r.Status != core.StatusPass {
			t.Errorf("Test %q failed: %v", r.Name, r.Message)
		}
	}
}

func TestSystemPlugin_Execute_FailFast(t *testing.T) {
	mock := core.NewMockProvider()
	// First package fails
	mock.SetCommandResult("dpkg -l missing-package 2>/dev/null | grep '^ii'", "", "", 1, nil)
	mock.SetCommandResult("rpm -q missing-package 2>/dev/null", "", "", 1, nil)
	mock.SetCommandResult("apk info -e missing-package 2>/dev/null", "", "", 1, nil)

	spec := &core.Spec{
		Tests: core.Tests{
			Packages: []core.PackageTest{
				{Name: "Missing package", Packages: []string{"missing-package"}, State: "present"},
				{Name: "Another package", Packages: []string{"other-package"}, State: "present"},
			},
		},
	}

	plugin := NewSystemPlugin()
	ctx := context.Background()

	results, shouldStop := plugin.Execute(ctx, spec, mock, true, nil)

	if len(results) != 1 {
		t.Errorf("Expected 1 result (fail fast), got %d", len(results))
	}

	if !shouldStop {
		t.Error("Should stop when failFast is true and test fails")
	}

	if results[0].Status != core.StatusFail {
		t.Errorf("Expected failure, got %v", results[0].Status)
	}
}
