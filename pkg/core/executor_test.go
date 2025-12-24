package core

import (
	"context"
	"testing"
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

func TestExecutor_PackageTest(t *testing.T) {
	tests := []struct {
		name         string
		packageTest  PackageTest
		setupMock    func(*MockProvider)
		wantStatus   Status
		wantContains string
	}{
		{
			name: "package installed - dpkg",
			packageTest: PackageTest{
				Name:     "Docker installed",
				Packages: []string{"docker-ce"},
				State:    "present",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("dpkg -l docker-ce 2>/dev/null | grep '^ii'", "ii  docker-ce  5:24.0.7  amd64", "", 0, nil)
			},
			wantStatus:   StatusPass,
			wantContains: "1 packages are installed",
		},
		{
			name: "package not installed",
			packageTest: PackageTest{
				Name:     "Docker installed",
				Packages: []string{"docker-ce"},
				State:    "present",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("dpkg -l docker-ce 2>/dev/null | grep '^ii'", "", "", 1, nil)
				m.SetCommandResult("rpm -q docker-ce 2>/dev/null", "", "", 1, nil)
				m.SetCommandResult("apk info -e docker-ce 2>/dev/null", "", "", 1, nil)
			},
			wantStatus:   StatusFail,
			wantContains: "not installed",
		},
		{
			name: "package absent check - success",
			packageTest: PackageTest{
				Name:     "Telnet removed",
				Packages: []string{"telnet"},
				State:    "absent",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("dpkg -l telnet 2>/dev/null | grep '^ii'", "", "", 1, nil)
				m.SetCommandResult("rpm -q telnet 2>/dev/null", "", "", 1, nil)
				m.SetCommandResult("apk info -e telnet 2>/dev/null", "", "", 1, nil)
			},
			wantStatus:   StatusPass,
			wantContains: "absent as expected",
		},
		{
			name: "package absent check - failure",
			packageTest: PackageTest{
				Name:     "Telnet removed",
				Packages: []string{"telnet"},
				State:    "absent",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("dpkg -l telnet 2>/dev/null | grep '^ii'", "ii  telnet  0.17  amd64", "", 0, nil)
			},
			wantStatus:   StatusFail,
			wantContains: "should be absent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockProvider()
			tt.setupMock(mock)

			spec := &Spec{
				Tests: Tests{
					Packages: []PackageTest{tt.packageTest},
				},
			}

			executor := NewExecutor(spec, mock)
			ctx := context.Background()

			results, err := executor.Execute(ctx)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if len(results.Results) != 1 {
				t.Fatalf("Expected 1 result, got %d", len(results.Results))
			}

			result := results.Results[0]
			if result.Status != tt.wantStatus {
				t.Errorf("Status = %v, want %v", result.Status, tt.wantStatus)
			}

			if tt.wantContains != "" && result.Message != "" {
				if !contains(result.Message, tt.wantContains) {
					t.Errorf("Message %q does not contain %q", result.Message, tt.wantContains)
				}
			}

			if result.Duration == 0 {
				t.Error("Duration should be set")
			}
		})
	}
}

func TestExecutor_FileTest(t *testing.T) {
	tests := []struct {
		name         string
		fileTest     FileTest
		setupMock    func(*MockProvider)
		wantStatus   Status
		wantContains string
	}{
		{
			name: "directory exists with correct permissions",
			fileTest: FileTest{
				Name:  "App directory",
				Path:  "/opt/app",
				Type:  "directory",
				Owner: "appuser",
				Group: "appuser",
				Mode:  "0755",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("stat -c '%F:%U:%G:%a' /opt/app 2>/dev/null || echo 'notfound'", "directory:appuser:appuser:755", "", 0, nil)
			},
			wantStatus:   StatusPass,
			wantContains: "correct properties",
		},
		{
			name: "file does not exist",
			fileTest: FileTest{
				Name: "Config file",
				Path: "/etc/myapp/config.yml",
				Type: "file",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("stat -c '%F:%U:%G:%a' /etc/myapp/config.yml 2>/dev/null || echo 'notfound'", "notfound", "", 0, nil)
			},
			wantStatus:   StatusFail,
			wantContains: "does not exist",
		},
		{
			name: "wrong file type",
			fileTest: FileTest{
				Name: "Should be directory",
				Path: "/opt/app",
				Type: "directory",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("stat -c '%F:%U:%G:%a' /opt/app 2>/dev/null || echo 'notfound'", "regular file:root:root:644", "", 0, nil)
			},
			wantStatus:   StatusFail,
			wantContains: "expected directory",
		},
		{
			name: "wrong owner",
			fileTest: FileTest{
				Name:  "Config file",
				Path:  "/etc/app/config",
				Type:  "file",
				Owner: "appuser",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("stat -c '%F:%U:%G:%a' /etc/app/config 2>/dev/null || echo 'notfound'", "regular file:root:root:644", "", 0, nil)
			},
			wantStatus:   StatusFail,
			wantContains: "owner is root, expected appuser",
		},
		{
			name: "wrong permissions",
			fileTest: FileTest{
				Name: "Private directory",
				Path: "/opt/secrets",
				Type: "directory",
				Mode: "0700",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("stat -c '%F:%U:%G:%a' /opt/secrets 2>/dev/null || echo 'notfound'", "directory:user:user:755", "", 0, nil)
			},
			wantStatus:   StatusFail,
			wantContains: "mode is 755, expected 700",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockProvider()
			tt.setupMock(mock)

			spec := &Spec{
				Tests: Tests{
					Files: []FileTest{tt.fileTest},
				},
			}

			executor := NewExecutor(spec, mock)
			ctx := context.Background()

			results, err := executor.Execute(ctx)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if len(results.Results) != 1 {
				t.Fatalf("Expected 1 result, got %d", len(results.Results))
			}

			result := results.Results[0]
			if result.Status != tt.wantStatus {
				t.Errorf("Status = %v, want %v", result.Status, tt.wantStatus)
			}

			if tt.wantContains != "" && result.Message != "" {
				if !contains(result.Message, tt.wantContains) {
					t.Errorf("Message %q does not contain %q", result.Message, tt.wantContains)
				}
			}
		})
	}
}

func TestExecutor_FailFast(t *testing.T) {
	mock := NewMockProvider()
	mock.SetCommandResult("dpkg -l package1 2>/dev/null | grep '^ii'", "", "", 1, nil)
	mock.SetCommandResult("rpm -q package1 2>/dev/null", "", "", 1, nil)
	mock.SetCommandResult("apk info -e package1 2>/dev/null", "", "", 1, nil)

	spec := &Spec{
		Config: SpecConfig{
			FailFast: true,
		},
		Tests: Tests{
			Packages: []PackageTest{
				{Name: "Test 1", Packages: []string{"package1"}, State: "present"},
				{Name: "Test 2", Packages: []string{"package2"}, State: "present"},
				{Name: "Test 3", Packages: []string{"package3"}, State: "present"},
			},
		},
	}

	executor := NewExecutor(spec, mock)
	ctx := context.Background()

	results, err := executor.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// With fail_fast, should stop after first failure
	if len(results.Results) != 1 {
		t.Errorf("Expected 1 result (fail fast), got %d", len(results.Results))
	}

	if results.Results[0].Status != StatusFail {
		t.Errorf("First result should be failed")
	}
}

func TestExecutor_MultipleTests(t *testing.T) {
	mock := NewMockProvider()
	mock.SetCommandResult("dpkg -l docker-ce 2>/dev/null | grep '^ii'", "ii  docker-ce", "", 0, nil)
	mock.SetCommandResult("stat -c '%F:%U:%G:%a' /opt/app 2>/dev/null || echo 'notfound'", "directory:root:root:755", "", 0, nil)

	spec := &Spec{
		Metadata: SpecMetadata{
			Name: "Multi Test",
		},
		Tests: Tests{
			Packages: []PackageTest{
				{Name: "Docker installed", Packages: []string{"docker-ce"}, State: "present"},
			},
			Files: []FileTest{
				{Name: "App dir", Path: "/opt/app", Type: "directory"},
			},
		},
	}

	executor := NewExecutor(spec, mock)
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
		if r.Status != StatusPass {
			t.Errorf("Test %q failed: %v", r.Name, r.Message)
		}
	}
}

func TestExecutor_ServiceTest(t *testing.T) {
	tests := []struct {
		name         string
		serviceTest  ServiceTest
		setupMock    func(*MockProvider)
		wantStatus   Status
		wantContains string
	}{
		{
			name: "service running and enabled",
			serviceTest: ServiceTest{
				Name:    "Docker running",
				Service: "docker",
				State:   "running",
				Enabled: true,
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("systemctl is-active docker 2>/dev/null", "active", "", 0, nil)
				m.SetCommandResult("systemctl is-enabled docker 2>/dev/null", "enabled", "", 0, nil)
			},
			wantStatus:   StatusPass,
			wantContains: "running",
		},
		{
			name: "service not running",
			serviceTest: ServiceTest{
				Name:    "Docker running",
				Service: "docker",
				State:   "running",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("systemctl is-active docker 2>/dev/null", "inactive", "", 3, nil)
			},
			wantStatus:   StatusFail,
			wantContains: "not running",
		},
		{
			name: "service stopped check - success",
			serviceTest: ServiceTest{
				Name:    "Telnet stopped",
				Service: "telnet",
				State:   "stopped",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("systemctl is-active telnet 2>/dev/null", "inactive", "", 3, nil)
			},
			wantStatus:   StatusPass,
			wantContains: "stopped",
		},
		{
			name: "service stopped check - failure",
			serviceTest: ServiceTest{
				Name:    "Telnet stopped",
				Service: "telnet",
				State:   "stopped",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("systemctl is-active telnet 2>/dev/null", "active", "", 0, nil)
			},
			wantStatus:   StatusFail,
			wantContains: "should be stopped",
		},
		{
			name: "service not enabled",
			serviceTest: ServiceTest{
				Name:    "Docker enabled",
				Service: "docker",
				State:   "running",
				Enabled: true,
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("systemctl is-active docker 2>/dev/null", "active", "", 0, nil)
				m.SetCommandResult("systemctl is-enabled docker 2>/dev/null", "disabled", "", 1, nil)
			},
			wantStatus:   StatusFail,
			wantContains: "not enabled",
		},
		{
			name: "multiple services",
			serviceTest: ServiceTest{
				Name:     "Critical services running",
				Services: []string{"docker", "nginx"},
				State:    "running",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("systemctl is-active docker 2>/dev/null", "active", "", 0, nil)
				m.SetCommandResult("systemctl is-enabled docker 2>/dev/null", "enabled", "", 0, nil)
				m.SetCommandResult("systemctl is-active nginx 2>/dev/null", "active", "", 0, nil)
				m.SetCommandResult("systemctl is-enabled nginx 2>/dev/null", "enabled", "", 0, nil)
			},
			wantStatus:   StatusPass,
			wantContains: "2 services",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockProvider()
			tt.setupMock(mock)

			spec := &Spec{
				Tests: Tests{
					Services: []ServiceTest{tt.serviceTest},
				},
			}

			executor := NewExecutor(spec, mock)
			ctx := context.Background()

			results, err := executor.Execute(ctx)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if len(results.Results) != 1 {
				t.Fatalf("Expected 1 result, got %d", len(results.Results))
			}

			result := results.Results[0]
			if result.Status != tt.wantStatus {
				t.Errorf("Status = %v, want %v", result.Status, tt.wantStatus)
			}

			if tt.wantContains != "" && result.Message != "" {
				if !contains(result.Message, tt.wantContains) {
					t.Errorf("Message %q does not contain %q", result.Message, tt.wantContains)
				}
			}
		})
	}
}

func TestExecutor_UserTest(t *testing.T) {
	tests := []struct {
		name         string
		userTest     UserTest
		setupMock    func(*MockProvider)
		wantStatus   Status
		wantContains string
	}{
		{
			name: "user exists",
			userTest: UserTest{
				Name: "Ubuntu user exists",
				User: "ubuntu",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("id -u ubuntu 2>/dev/null && id -g ubuntu 2>/dev/null && getent passwd ubuntu 2>/dev/null", "1000\n1000\nubuntu:x:1000:1000:Ubuntu:/home/ubuntu:/bin/bash", "", 0, nil)
			},
			wantStatus:   StatusPass,
			wantContains: "exists",
		},
		{
			name: "user does not exist",
			userTest: UserTest{
				Name: "Test user exists",
				User: "testuser",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("id -u testuser 2>/dev/null && id -g testuser 2>/dev/null && getent passwd testuser 2>/dev/null", "", "", 1, nil)
			},
			wantStatus:   StatusFail,
			wantContains: "does not exist",
		},
		{
			name: "user with correct shell",
			userTest: UserTest{
				Name:  "Ubuntu shell",
				User:  "ubuntu",
				Shell: "/bin/bash",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("id -u ubuntu 2>/dev/null && id -g ubuntu 2>/dev/null && getent passwd ubuntu 2>/dev/null", "1000\n1000\nubuntu:x:1000:1000:Ubuntu:/home/ubuntu:/bin/bash", "", 0, nil)
			},
			wantStatus:   StatusPass,
			wantContains: "exists",
		},
		{
			name: "user with wrong shell",
			userTest: UserTest{
				Name:  "Ubuntu shell",
				User:  "ubuntu",
				Shell: "/bin/zsh",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("id -u ubuntu 2>/dev/null && id -g ubuntu 2>/dev/null && getent passwd ubuntu 2>/dev/null", "1000\n1000\nubuntu:x:1000:1000:Ubuntu:/home/ubuntu:/bin/bash", "", 0, nil)
			},
			wantStatus:   StatusFail,
			wantContains: "shell",
		},
		{
			name: "user with correct home",
			userTest: UserTest{
				Name: "Ubuntu home",
				User: "ubuntu",
				Home: "/home/ubuntu",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("id -u ubuntu 2>/dev/null && id -g ubuntu 2>/dev/null && getent passwd ubuntu 2>/dev/null", "1000\n1000\nubuntu:x:1000:1000:Ubuntu:/home/ubuntu:/bin/bash", "", 0, nil)
			},
			wantStatus:   StatusPass,
			wantContains: "exists",
		},
		{
			name: "user with wrong home",
			userTest: UserTest{
				Name: "Ubuntu home",
				User: "ubuntu",
				Home: "/root",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("id -u ubuntu 2>/dev/null && id -g ubuntu 2>/dev/null && getent passwd ubuntu 2>/dev/null", "1000\n1000\nubuntu:x:1000:1000:Ubuntu:/home/ubuntu:/bin/bash", "", 0, nil)
			},
			wantStatus:   StatusFail,
			wantContains: "home",
		},
		{
			name: "user with groups",
			userTest: UserTest{
				Name:   "Ubuntu groups",
				User:   "ubuntu",
				Groups: []string{"sudo", "docker"},
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("id -u ubuntu 2>/dev/null && id -g ubuntu 2>/dev/null && getent passwd ubuntu 2>/dev/null", "1000\n1000\nubuntu:x:1000:1000:Ubuntu:/home/ubuntu:/bin/bash", "", 0, nil)
				m.SetCommandResult("id -Gn ubuntu 2>/dev/null", "ubuntu sudo docker", "", 0, nil)
			},
			wantStatus:   StatusPass,
			wantContains: "exists",
		},
		{
			name: "user missing group",
			userTest: UserTest{
				Name:   "Ubuntu groups",
				User:   "ubuntu",
				Groups: []string{"sudo", "wheel"},
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("id -u ubuntu 2>/dev/null && id -g ubuntu 2>/dev/null && getent passwd ubuntu 2>/dev/null", "1000\n1000\nubuntu:x:1000:1000:Ubuntu:/home/ubuntu:/bin/bash", "", 0, nil)
				m.SetCommandResult("id -Gn ubuntu 2>/dev/null", "ubuntu sudo docker", "", 0, nil)
			},
			wantStatus:   StatusFail,
			wantContains: "not in group",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockProvider()
			tt.setupMock(mock)

			spec := &Spec{
				Tests: Tests{
					Users: []UserTest{tt.userTest},
				},
			}

			executor := NewExecutor(spec, mock)
			ctx := context.Background()

			results, err := executor.Execute(ctx)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if len(results.Results) != 1 {
				t.Fatalf("Expected 1 result, got %d", len(results.Results))
			}

			result := results.Results[0]
			if result.Status != tt.wantStatus {
				t.Errorf("Status = %v, want %v", result.Status, tt.wantStatus)
			}

			if tt.wantContains != "" && result.Message != "" {
				if !contains(result.Message, tt.wantContains) {
					t.Errorf("Message %q does not contain %q", result.Message, tt.wantContains)
				}
			}
		})
	}
}

func TestExecutor_GroupTest(t *testing.T) {
	tests := []struct {
		name         string
		groupTest    GroupTest
		setupMock    func(*MockProvider)
		wantStatus   Status
		wantContains string
	}{
		{
			name: "group exists",
			groupTest: GroupTest{
				Name:   "Docker group exists",
				Groups: []string{"docker"},
				State:  "present",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("getent group docker 2>/dev/null", "docker:x:999:ubuntu", "", 0, nil)
			},
			wantStatus:   StatusPass,
			wantContains: "groups exist",
		},
		{
			name: "group does not exist",
			groupTest: GroupTest{
				Name:   "Test group exists",
				Groups: []string{"testgroup"},
				State:  "present",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("getent group testgroup 2>/dev/null", "", "", 2, nil)
			},
			wantStatus:   StatusFail,
			wantContains: "does not exist",
		},
		{
			name: "group absent check - success",
			groupTest: GroupTest{
				Name:   "Unwanted group absent",
				Groups: []string{"badgroup"},
				State:  "absent",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("getent group badgroup 2>/dev/null", "", "", 2, nil)
			},
			wantStatus:   StatusPass,
			wantContains: "absent as expected",
		},
		{
			name: "group absent check - failure",
			groupTest: GroupTest{
				Name:   "Unwanted group absent",
				Groups: []string{"docker"},
				State:  "absent",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("getent group docker 2>/dev/null", "docker:x:999:ubuntu", "", 0, nil)
			},
			wantStatus:   StatusFail,
			wantContains: "should be absent",
		},
		{
			name: "multiple groups",
			groupTest: GroupTest{
				Name:   "Required groups exist",
				Groups: []string{"docker", "sudo"},
				State:  "present",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("getent group docker 2>/dev/null", "docker:x:999:ubuntu", "", 0, nil)
				m.SetCommandResult("getent group sudo 2>/dev/null", "sudo:x:27:ubuntu", "", 0, nil)
			},
			wantStatus:   StatusPass,
			wantContains: "2 groups",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockProvider()
			tt.setupMock(mock)

			spec := &Spec{
				Tests: Tests{
					Groups: []GroupTest{tt.groupTest},
				},
			}

			executor := NewExecutor(spec, mock)
			ctx := context.Background()

			results, err := executor.Execute(ctx)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if len(results.Results) != 1 {
				t.Fatalf("Expected 1 result, got %d", len(results.Results))
			}

			result := results.Results[0]
			if result.Status != tt.wantStatus {
				t.Errorf("Status = %v, want %v", result.Status, tt.wantStatus)
			}

			if tt.wantContains != "" && result.Message != "" {
				if !contains(result.Message, tt.wantContains) {
					t.Errorf("Message %q does not contain %q", result.Message, tt.wantContains)
				}
			}
		})
	}
}

func TestExecutor_FileContentTest(t *testing.T) {
	tests := []struct {
		name            string
		fileContentTest FileContentTest
		setupMock       func(*MockProvider)
		wantStatus      Status
		wantContains    string
	}{
		{
			name: "file contains string",
			fileContentTest: FileContentTest{
				Name:     "Config contains setting",
				Path:     "/etc/app/config.yml",
				Contains: []string{"log-driver"},
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("test -f /etc/app/config.yml && test -r /etc/app/config.yml", "", "", 0, nil)
				m.SetCommandResult("grep -F 'log-driver' /etc/app/config.yml >/dev/null 2>&1", "", "", 0, nil)
			},
			wantStatus:   StatusPass,
			wantContains: "contains",
		},
		{
			name: "file missing string",
			fileContentTest: FileContentTest{
				Name:     "Config contains setting",
				Path:     "/etc/app/config.yml",
				Contains: []string{"missing-option"},
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("test -f /etc/app/config.yml && test -r /etc/app/config.yml", "", "", 0, nil)
				m.SetCommandResult("grep -F 'missing-option' /etc/app/config.yml >/dev/null 2>&1", "", "", 1, nil)
			},
			wantStatus:   StatusFail,
			wantContains: "does not contain",
		},
		{
			name: "file does not exist",
			fileContentTest: FileContentTest{
				Name:     "Config contains setting",
				Path:     "/etc/app/missing.yml",
				Contains: []string{"something"},
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("test -f /etc/app/missing.yml && test -r /etc/app/missing.yml", "", "", 1, nil)
			},
			wantStatus:   StatusFail,
			wantContains: "does not exist",
		},
		{
			name: "file matches regex pattern",
			fileContentTest: FileContentTest{
				Name:    "SSH config secure",
				Path:    "/etc/ssh/sshd_config",
				Matches: "^PermitRootLogin no$",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("test -f /etc/ssh/sshd_config && test -r /etc/ssh/sshd_config", "", "", 0, nil)
				m.SetCommandResult("grep -E '^PermitRootLogin no$' /etc/ssh/sshd_config >/dev/null 2>&1", "", "", 0, nil)
			},
			wantStatus:   StatusPass,
			wantContains: "matches pattern",
		},
		{
			name: "file does not match pattern",
			fileContentTest: FileContentTest{
				Name:    "SSH config check",
				Path:    "/etc/ssh/sshd_config",
				Matches: "^PermitRootLogin yes$",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("test -f /etc/ssh/sshd_config && test -r /etc/ssh/sshd_config", "", "", 0, nil)
				m.SetCommandResult("grep -E '^PermitRootLogin yes$' /etc/ssh/sshd_config >/dev/null 2>&1", "", "", 1, nil)
			},
			wantStatus:   StatusFail,
			wantContains: "does not match",
		},
		{
			name: "file contains multiple strings",
			fileContentTest: FileContentTest{
				Name:     "Config has required settings",
				Path:     "/etc/docker/daemon.json",
				Contains: []string{"log-driver", "metrics-addr", "storage-driver"},
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("test -f /etc/docker/daemon.json && test -r /etc/docker/daemon.json", "", "", 0, nil)
				m.SetCommandResult("grep -F 'log-driver' /etc/docker/daemon.json >/dev/null 2>&1", "", "", 0, nil)
				m.SetCommandResult("grep -F 'metrics-addr' /etc/docker/daemon.json >/dev/null 2>&1", "", "", 0, nil)
				m.SetCommandResult("grep -F 'storage-driver' /etc/docker/daemon.json >/dev/null 2>&1", "", "", 0, nil)
			},
			wantStatus:   StatusPass,
			wantContains: "contains all 3",
		},
		{
			name: "file contains strings and matches pattern",
			fileContentTest: FileContentTest{
				Name:     "Config complete",
				Path:     "/etc/app/config.yml",
				Contains: []string{"setting1", "setting2"},
				Matches:  "^version:",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("test -f /etc/app/config.yml && test -r /etc/app/config.yml", "", "", 0, nil)
				m.SetCommandResult("grep -F 'setting1' /etc/app/config.yml >/dev/null 2>&1", "", "", 0, nil)
				m.SetCommandResult("grep -F 'setting2' /etc/app/config.yml >/dev/null 2>&1", "", "", 0, nil)
				m.SetCommandResult("grep -E '^version:' /etc/app/config.yml >/dev/null 2>&1", "", "", 0, nil)
			},
			wantStatus:   StatusPass,
			wantContains: "contains all 2 strings and matches",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockProvider()
			tt.setupMock(mock)

			spec := &Spec{
				Tests: Tests{
					FileContent: []FileContentTest{tt.fileContentTest},
				},
			}

			executor := NewExecutor(spec, mock)
			ctx := context.Background()

			results, err := executor.Execute(ctx)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if len(results.Results) != 1 {
				t.Fatalf("Expected 1 result, got %d", len(results.Results))
			}

			result := results.Results[0]
			if result.Status != tt.wantStatus {
				t.Errorf("Status = %v, want %v", result.Status, tt.wantStatus)
			}

			if tt.wantContains != "" && result.Message != "" {
				if !contains(result.Message, tt.wantContains) {
					t.Errorf("Message %q does not contain %q", result.Message, tt.wantContains)
				}
			}
		})
	}
}

func TestExecutor_CommandContentTest(t *testing.T) {
	tests := []struct {
		name               string
		commandContentTest CommandContentTest
		setupMock          func(*MockProvider)
		wantStatus         Status
		wantContains       string
	}{
		{
			name: "command output contains string",
			commandContentTest: CommandContentTest{
				Name:     "Disk space check",
				Command:  "df -h",
				Contains: []string{"/var/lib/docker"},
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("df -h", "Filesystem      Size  Used Avail Use% Mounted on\n/dev/sda1        50G   20G   30G  40% /\n/dev/sdb1       100G   50G   50G  50% /var/lib/docker", "", 0, nil)
			},
			wantStatus:   StatusPass,
			wantContains: "contains",
		},
		{
			name: "command output missing string",
			commandContentTest: CommandContentTest{
				Name:     "Check for nginx",
				Command:  "ps aux",
				Contains: []string{"nginx"},
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("ps aux", "USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND\nroot         1  0.0  0.1  12345  6789 ?        Ss   10:00   0:01 /sbin/init", "", 0, nil)
			},
			wantStatus:   StatusFail,
			wantContains: "does not contain",
		},
		{
			name: "command exit code check",
			commandContentTest: CommandContentTest{
				Name:     "Service status",
				Command:  "systemctl is-active docker",
				ExitCode: 0,
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("systemctl is-active docker", "active", "", 0, nil)
			},
			wantStatus:   StatusPass,
			wantContains: "executed successfully",
		},
		{
			name: "command wrong exit code",
			commandContentTest: CommandContentTest{
				Name:     "Service should fail",
				Command:  "systemctl is-active nonexistent",
				ExitCode: 3,
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("systemctl is-active nonexistent", "inactive", "", 0, nil)
			},
			wantStatus:   StatusFail,
			wantContains: "exit code is 0, expected 3",
		},
		{
			name: "command multiple contains",
			commandContentTest: CommandContentTest{
				Name:     "Docker info check",
				Command:  "docker info",
				Contains: []string{"Server Version", "Storage Driver", "Logging Driver"},
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("docker info", "Server Version: 24.0.7\nStorage Driver: overlay2\nLogging Driver: json-file\nCgroup Driver: systemd", "", 0, nil)
			},
			wantStatus:   StatusPass,
			wantContains: "contains all 3",
		},
		{
			name: "command with exit code and contains",
			commandContentTest: CommandContentTest{
				Name:     "Success with content",
				Command:  "echo hello",
				ExitCode: 0,
				Contains: []string{"hello"},
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("echo hello", "hello", "", 0, nil)
			},
			wantStatus:   StatusPass,
			wantContains: "contains all 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockProvider()
			tt.setupMock(mock)

			spec := &Spec{
				Tests: Tests{
					CommandContent: []CommandContentTest{tt.commandContentTest},
				},
			}

			executor := NewExecutor(spec, mock)
			ctx := context.Background()

			results, err := executor.Execute(ctx)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if len(results.Results) != 1 {
				t.Fatalf("Expected 1 result, got %d", len(results.Results))
			}

			result := results.Results[0]
			if result.Status != tt.wantStatus {
				t.Errorf("Status = %v, want %v", result.Status, tt.wantStatus)
			}

			if tt.wantContains != "" && result.Message != "" {
				if !contains(result.Message, tt.wantContains) {
					t.Errorf("Message %q does not contain %q", result.Message, tt.wantContains)
				}
			}
		})
	}
}

func TestExecutor_HTTPTest(t *testing.T) {
	tests := []struct {
		name         string
		httpTest     HTTPTest
		setupMock    func(*MockProvider)
		wantStatus   Status
		wantContains string
	}{
		{
			name: "successful GET request with 200",
			httpTest: HTTPTest{
				Name:       "API health check",
				URL:        "http://localhost:8080/health",
				StatusCode: 200,
				Method:     "GET", // Explicitly set to match validation default
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("curl -s -w $'\\n%{http_code}' 'http://localhost:8080/health'", "{\"status\":\"ok\"}\n200", "", 0, nil)
			},
			wantStatus:   StatusPass,
			wantContains: "returned status 200",
		},
		{
			name: "POST request with custom status code",
			httpTest: HTTPTest{
				Name:       "Webhook endpoint",
				URL:        "https://api.example.com/webhook",
				StatusCode: 202,
				Method:     "POST",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("curl -s -X POST -w $'\\n%{http_code}' 'https://api.example.com/webhook'", "{\"accepted\":true}\n202", "", 0, nil)
			},
			wantStatus:   StatusPass,
			wantContains: "returned status 202",
		},
		{
			name: "request with insecure flag",
			httpTest: HTTPTest{
				Name:       "Self-signed cert",
				URL:        "https://internal.local/api",
				StatusCode: 200,
				Method:     "GET",
				Insecure:   true,
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("curl -s -k -w $'\\n%{http_code}' 'https://internal.local/api'", "{\"data\":\"test\"}\n200", "", 0, nil)
			},
			wantStatus:   StatusPass,
			wantContains: "returned status 200",
		},
		{
			name: "response with content validation",
			httpTest: HTTPTest{
				Name:       "API response validation",
				URL:        "http://localhost:3000/status",
				StatusCode: 200,
				Method:     "GET",
				Contains:   []string{"healthy", "version"},
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("curl -s -w $'\\n%{http_code}' 'http://localhost:3000/status'", "{\"status\":\"healthy\",\"version\":\"1.2.3\"}\n200", "", 0, nil)
			},
			wantStatus:   StatusPass,
			wantContains: "with all expected content (2 strings)",
		},
		{
			name: "wrong status code",
			httpTest: HTTPTest{
				Name:       "Expect 200 but got 404",
				URL:        "http://localhost:8080/missing",
				StatusCode: 200,
				Method:     "GET",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("curl -s -w $'\\n%{http_code}' 'http://localhost:8080/missing'", "Not Found\n404", "", 0, nil)
			},
			wantStatus:   StatusFail,
			wantContains: "Status code is 404, expected 200",
		},
		{
			name: "missing content in response",
			httpTest: HTTPTest{
				Name:       "Missing expected string",
				URL:        "http://localhost:8080/api",
				StatusCode: 200,
				Method:     "GET",
				Contains:   []string{"expected_field"},
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("curl -s -w $'\\n%{http_code}' 'http://localhost:8080/api'", "{\"other\":\"data\"}\n200", "", 0, nil)
			},
			wantStatus:   StatusFail,
			wantContains: "Response body missing expected strings: 'expected_field'",
		},
		{
			name: "curl command fails",
			httpTest: HTTPTest{
				Name:       "Connection refused",
				URL:        "http://localhost:9999",
				StatusCode: 200,
				Method:     "GET",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("curl -s -w $'\\n%{http_code}' 'http://localhost:9999'", "", "curl: (7) Failed to connect", 7, nil)
			},
			wantStatus:   StatusFail,
			wantContains: "HTTP request failed",
		},
		{
			name: "redirect (302) when expecting 200",
			httpTest: HTTPTest{
				Name:       "No redirect expected",
				URL:        "https://example.com/old",
				StatusCode: 200,
				Method:     "GET",
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("curl -s -w $'\\n%{http_code}' 'https://example.com/old'", "<html>Moved</html>\n302", "", 0, nil)
			},
			wantStatus:   StatusFail,
			wantContains: "Status code is 302, expected 200",
		},
		{
			name: "follow redirects and get final 200",
			httpTest: HTTPTest{
				Name:            "Follow redirect to final page",
				URL:             "https://example.com/redirect",
				StatusCode:      200,
				Method:          "GET",
				FollowRedirects: true,
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("curl -s -L -w $'\\n%{http_code}' 'https://example.com/redirect'", "<html>Final page</html>\n200", "", 0, nil)
			},
			wantStatus:   StatusPass,
			wantContains: "returned status 200",
		},
		{
			name: "follow redirects with insecure flag",
			httpTest: HTTPTest{
				Name:            "Follow redirect with self-signed cert",
				URL:             "https://internal.local/old",
				StatusCode:      200,
				Method:          "GET",
				FollowRedirects: true,
				Insecure:        true,
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("curl -s -L -k -w $'\\n%{http_code}' 'https://internal.local/old'", "<html>Redirected page</html>\n200", "", 0, nil)
			},
			wantStatus:   StatusPass,
			wantContains: "returned status 200",
		},
		{
			name: "follow redirects with POST method",
			httpTest: HTTPTest{
				Name:            "POST with redirect",
				URL:             "https://api.example.com/create",
				StatusCode:      201,
				Method:          "POST",
				FollowRedirects: true,
			},
			setupMock: func(m *MockProvider) {
				m.SetCommandResult("curl -s -X POST -L -w $'\\n%{http_code}' 'https://api.example.com/create'", "{\"id\":123}\n201", "", 0, nil)
			},
			wantStatus:   StatusPass,
			wantContains: "returned status 201",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockProvider()
			tt.setupMock(mock)

			spec := &Spec{
				Tests: Tests{
					HTTP: []HTTPTest{tt.httpTest},
				},
			}

			// Run validation to set defaults
			if err := spec.Validate(); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			executor := NewExecutor(spec, mock)
			ctx := context.Background()

			results, err := executor.Execute(ctx)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if len(results.Results) != 1 {
				t.Fatalf("Expected 1 result, got %d", len(results.Results))
			}

			result := results.Results[0]
			if result.Status != tt.wantStatus {
				t.Errorf("Status = %v, want %v", result.Status, tt.wantStatus)
			}

			if tt.wantContains != "" && result.Message != "" {
				if !contains(result.Message, tt.wantContains) {
					t.Errorf("Message %q does not contain %q", result.Message, tt.wantContains)
				}
			}
		})
	}
}

func TestNewExecutor(t *testing.T) {
	spec := &Spec{}
	mock := NewMockProvider()

	executor := NewExecutor(spec, mock)

	if executor == nil {
		t.Fatal("NewExecutor returned nil")
	}

	if executor.spec != spec {
		t.Error("Executor spec not set correctly")
	}

	if executor.provider != mock {
		t.Error("Executor provider not set correctly")
	}
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
