package system

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestExecutor_SystemInfoTest(t *testing.T) {
	tests := []struct {
		name         string
		systemInfo   core.SystemInfoTest
		setupMock    func(*core.MockProvider)
		wantStatus   core.Status
		wantContains string
	}{
		{
			name: "OS matches",
			systemInfo: core.SystemInfoTest{
				Name: "Ubuntu OS",
				OS:   "ubuntu",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("grep '^ID=' /etc/os-release 2>/dev/null | cut -d= -f2 | tr -d '\"'", "ubuntu", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "matches all specified criteria",
		},
		{
			name: "OS does not match",
			systemInfo: core.SystemInfoTest{
				Name: "Ubuntu OS",
				OS:   "ubuntu",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("grep '^ID=' /etc/os-release 2>/dev/null | cut -d= -f2 | tr -d '\"'", "debian", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "OS is 'debian', expected 'ubuntu'",
		},
		{
			name: "OS version exact match",
			systemInfo: core.SystemInfoTest{
				Name:         "Ubuntu 20.04",
				OSVersion:    "20.04",
				VersionMatch: "exact",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("grep '^VERSION_ID=' /etc/os-release 2>/dev/null | cut -d= -f2 | tr -d '\"'", "20.04", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "matches all specified criteria",
		},
		{
			name: "OS version prefix match",
			systemInfo: core.SystemInfoTest{
				Name:         "Ubuntu 20.x",
				OSVersion:    "20",
				VersionMatch: "prefix",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("grep '^VERSION_ID=' /etc/os-release 2>/dev/null | cut -d= -f2 | tr -d '\"'", "20.04", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "matches all specified criteria",
		},
		{
			name: "OS version does not match",
			systemInfo: core.SystemInfoTest{
				Name:         "Ubuntu 22.04",
				OSVersion:    "22.04",
				VersionMatch: "exact",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("grep '^VERSION_ID=' /etc/os-release 2>/dev/null | cut -d= -f2 | tr -d '\"'", "20.04", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "OS version is '20.04', expected '22.04'",
		},
		{
			name: "Architecture matches",
			systemInfo: core.SystemInfoTest{
				Name: "x86_64 arch",
				Arch: "x86_64",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("uname -m 2>/dev/null", "x86_64", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "matches all specified criteria",
		},
		{
			name: "Architecture does not match",
			systemInfo: core.SystemInfoTest{
				Name: "arm64 arch",
				Arch: "arm64",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("uname -m 2>/dev/null", "x86_64", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "Architecture is 'x86_64', expected 'arm64'",
		},
		{
			name: "Kernel version exact match",
			systemInfo: core.SystemInfoTest{
				Name:          "Kernel 5.15.0",
				KernelVersion: "5.15.0-76-generic",
				VersionMatch:  "exact",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("uname -r 2>/dev/null", "5.15.0-76-generic", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "matches all specified criteria",
		},
		{
			name: "Kernel version prefix match",
			systemInfo: core.SystemInfoTest{
				Name:          "Kernel 5.15",
				KernelVersion: "5.15",
				VersionMatch:  "prefix",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("uname -r 2>/dev/null", "5.15.0-76-generic", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "matches all specified criteria",
		},
		{
			name: "Kernel version does not match",
			systemInfo: core.SystemInfoTest{
				Name:          "Kernel 6.0",
				KernelVersion: "6.0",
				VersionMatch:  "prefix",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("uname -r 2>/dev/null", "5.15.0-76-generic", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "Kernel version is '5.15.0-76-generic', expected '6.0'",
		},
		{
			name: "Hostname matches",
			systemInfo: core.SystemInfoTest{
				Name:     "webserver hostname",
				Hostname: "webserver",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("hostname -s 2>/dev/null", "webserver", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "matches all specified criteria",
		},
		{
			name: "Hostname does not match",
			systemInfo: core.SystemInfoTest{
				Name:     "webserver hostname",
				Hostname: "webserver",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("hostname -s 2>/dev/null", "dbserver", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "Hostname is 'dbserver', expected 'webserver'",
		},
		{
			name: "FQDN matches",
			systemInfo: core.SystemInfoTest{
				Name: "webserver FQDN",
				FQDN: "webserver.example.com",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("hostname -f 2>/dev/null", "webserver.example.com", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "matches all specified criteria",
		},
		{
			name: "FQDN does not match",
			systemInfo: core.SystemInfoTest{
				Name: "webserver FQDN",
				FQDN: "webserver.example.com",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("hostname -f 2>/dev/null", "dbserver.example.com", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "FQDN is 'dbserver.example.com', expected 'webserver.example.com'",
		},
		{
			name: "Multiple fields match",
			systemInfo: core.SystemInfoTest{
				Name:         "Full system check",
				OS:           "ubuntu",
				OSVersion:    "20.04",
				Arch:         "x86_64",
				Hostname:     "webserver",
				VersionMatch: "exact",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("grep '^ID=' /etc/os-release 2>/dev/null | cut -d= -f2 | tr -d '\"'", "ubuntu", "", 0, nil)
				m.SetCommandResult("grep '^VERSION_ID=' /etc/os-release 2>/dev/null | cut -d= -f2 | tr -d '\"'", "20.04", "", 0, nil)
				m.SetCommandResult("uname -m 2>/dev/null", "x86_64", "", 0, nil)
				m.SetCommandResult("hostname -s 2>/dev/null", "webserver", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "matches all specified criteria",
		},
		{
			name: "Multiple fields with one failure",
			systemInfo: core.SystemInfoTest{
				Name:         "System check with mismatch",
				OS:           "ubuntu",
				OSVersion:    "22.04",
				Arch:         "x86_64",
				VersionMatch: "exact",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("grep '^ID=' /etc/os-release 2>/dev/null | cut -d= -f2 | tr -d '\"'", "ubuntu", "", 0, nil)
				m.SetCommandResult("grep '^VERSION_ID=' /etc/os-release 2>/dev/null | cut -d= -f2 | tr -d '\"'", "20.04", "", 0, nil)
				m.SetCommandResult("uname -m 2>/dev/null", "x86_64", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "OS version is '20.04', expected '22.04'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := core.NewMockProvider()
			tt.setupMock(mock)

			ctx := context.Background()

			result := executeSystemInfoTest(ctx, mock, tt.systemInfo)

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
