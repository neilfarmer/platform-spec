package system

import (
	"context"
	"testing"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

func TestExecutor_FilesystemTest(t *testing.T) {
	tests := []struct {
		name           string
		filesystemTest core.FilesystemTest
		setupMock      func(*core.MockProvider)
		wantStatus     core.Status
		wantContains   string
	}{
		{
			name: "filesystem is mounted",
			filesystemTest: core.FilesystemTest{
				Name:  "Root filesystem",
				Path:  "/",
				State: "mounted",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("findmnt --noheadings --output TARGET,FSTYPE,OPTIONS,SIZE,USED,USE% --target / 2>/dev/null", "/               ext4   rw,relatime    100G  50G   50%", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "is mounted",
		},
		{
			name: "filesystem is not mounted",
			filesystemTest: core.FilesystemTest{
				Name:  "Data volume",
				Path:  "/data",
				State: "mounted",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("findmnt --noheadings --output TARGET,FSTYPE,OPTIONS,SIZE,USED,USE% --target /data 2>/dev/null", "", "", 1, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "is not mounted",
		},
		{
			name: "filesystem unmounted as expected",
			filesystemTest: core.FilesystemTest{
				Name:  "Removed volume",
				Path:  "/old-data",
				State: "unmounted",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("findmnt --noheadings --output TARGET,FSTYPE,OPTIONS,SIZE,USED,USE% --target /old-data 2>/dev/null", "", "", 1, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "not mounted as expected",
		},
		{
			name: "filesystem mounted but should be unmounted",
			filesystemTest: core.FilesystemTest{
				Name:  "Removed volume",
				Path:  "/old-data",
				State: "unmounted",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("findmnt --noheadings --output TARGET,FSTYPE,OPTIONS,SIZE,USED,USE% --target /old-data 2>/dev/null", "/old-data       ext4   rw,relatime    100G  10G   10%", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "should be unmounted",
		},
		{
			name: "correct filesystem type",
			filesystemTest: core.FilesystemTest{
				Name:   "Data volume",
				Path:   "/data",
				State:  "mounted",
				Fstype: "xfs",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("findmnt --noheadings --output TARGET,FSTYPE,OPTIONS,SIZE,USED,USE% --target /data 2>/dev/null", "/data           xfs    rw,noatime     500G  100G  20%", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "as xfs",
		},
		{
			name: "wrong filesystem type",
			filesystemTest: core.FilesystemTest{
				Name:   "Data volume",
				Path:   "/data",
				State:  "mounted",
				Fstype: "xfs",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("findmnt --noheadings --output TARGET,FSTYPE,OPTIONS,SIZE,USED,USE% --target /data 2>/dev/null", "/data           ext4   rw,noatime     500G  100G  20%", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "expected xfs",
		},
		{
			name: "mount options present",
			filesystemTest: core.FilesystemTest{
				Name:    "Data volume",
				Path:    "/data",
				State:   "mounted",
				Options: []string{"noatime", "rw"},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("findmnt --noheadings --output TARGET,FSTYPE,OPTIONS,SIZE,USED,USE% --target /data 2>/dev/null", "/data           xfs    rw,noatime     500G  100G  20%", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "with correct options",
		},
		{
			name: "mount option missing",
			filesystemTest: core.FilesystemTest{
				Name:    "Data volume",
				Path:    "/data",
				State:   "mounted",
				Options: []string{"noatime", "noexec"},
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("findmnt --noheadings --output TARGET,FSTYPE,OPTIONS,SIZE,USED,USE% --target /data 2>/dev/null", "/data           xfs    rw,noatime     500G  100G  20%", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "missing required mount option",
		},
		{
			name: "minimum size met",
			filesystemTest: core.FilesystemTest{
				Name:      "Data volume",
				Path:      "/data",
				State:     "mounted",
				MinSizeGB: 100,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("findmnt --noheadings --output TARGET,FSTYPE,OPTIONS,SIZE,USED,USE% --target /data 2>/dev/null", "/data           xfs    rw,noatime     500G  100G  20%", "", 0, nil)
				m.SetCommandResult("df -BG --output=size /data | tail -1 | tr -d 'G '", "500", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "is mounted",
		},
		{
			name: "minimum size not met",
			filesystemTest: core.FilesystemTest{
				Name:      "Data volume",
				Path:      "/data",
				State:     "mounted",
				MinSizeGB: 1000,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("findmnt --noheadings --output TARGET,FSTYPE,OPTIONS,SIZE,USED,USE% --target /data 2>/dev/null", "/data           xfs    rw,noatime     500G  100G  20%", "", 0, nil)
				m.SetCommandResult("df -BG --output=size /data | tail -1 | tr -d 'G '", "500", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "minimum required is 1000GB",
		},
		{
			name: "usage below maximum",
			filesystemTest: core.FilesystemTest{
				Name:            "Data volume",
				Path:            "/data",
				State:           "mounted",
				MaxUsagePercent: 80,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("findmnt --noheadings --output TARGET,FSTYPE,OPTIONS,SIZE,USED,USE% --target /data 2>/dev/null", "/data           xfs    rw,noatime     500G  100G  20%", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "is mounted",
		},
		{
			name: "usage above maximum",
			filesystemTest: core.FilesystemTest{
				Name:            "Data volume",
				Path:            "/data",
				State:           "mounted",
				MaxUsagePercent: 50,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("findmnt --noheadings --output TARGET,FSTYPE,OPTIONS,SIZE,USED,USE% --target /data 2>/dev/null", "/data           xfs    rw,noatime     500G  400G  80%", "", 0, nil)
			},
			wantStatus:   core.StatusFail,
			wantContains: "maximum allowed is 50%",
		},
		{
			name: "complex check with all constraints",
			filesystemTest: core.FilesystemTest{
				Name:            "Production data",
				Path:            "/mnt/prod",
				State:           "mounted",
				Fstype:          "ext4",
				Options:         []string{"rw", "noatime"},
				MinSizeGB:       100,
				MaxUsagePercent: 80,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("findmnt --noheadings --output TARGET,FSTYPE,OPTIONS,SIZE,USED,USE% --target /mnt/prod 2>/dev/null", "/mnt/prod       ext4   rw,noatime,relatime   200G  100G  50%", "", 0, nil)
				m.SetCommandResult("df -BG --output=size /mnt/prod | tail -1 | tr -d 'G '", "200", "", 0, nil)
			},
			wantStatus:   core.StatusPass,
			wantContains: "with correct options",
		},
		{
			name: "malformed findmnt output",
			filesystemTest: core.FilesystemTest{
				Name:  "Data volume",
				Path:  "/data",
				State: "mounted",
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("findmnt --noheadings --output TARGET,FSTYPE,OPTIONS,SIZE,USED,USE% --target /data 2>/dev/null", "/data xfs", "", 0, nil)
			},
			wantStatus:   core.StatusError,
			wantContains: "Unexpected findmnt output",
		},
		{
			name: "invalid size format",
			filesystemTest: core.FilesystemTest{
				Name:      "Data volume",
				Path:      "/data",
				State:     "mounted",
				MinSizeGB: 100,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("findmnt --noheadings --output TARGET,FSTYPE,OPTIONS,SIZE,USED,USE% --target /data 2>/dev/null", "/data           xfs    rw,noatime     500G  100G  20%", "", 0, nil)
				m.SetCommandResult("df -BG --output=size /data | tail -1 | tr -d 'G '", "invalid", "", 0, nil)
			},
			wantStatus:   core.StatusError,
			wantContains: "Error parsing filesystem size",
		},
		{
			name: "invalid usage percent format",
			filesystemTest: core.FilesystemTest{
				Name:            "Data volume",
				Path:            "/data",
				State:           "mounted",
				MaxUsagePercent: 80,
			},
			setupMock: func(m *core.MockProvider) {
				m.SetCommandResult("findmnt --noheadings --output TARGET,FSTYPE,OPTIONS,SIZE,USED,USE% --target /data 2>/dev/null", "/data           xfs    rw,noatime     500G  100G  invalid%", "", 0, nil)
			},
			wantStatus:   core.StatusError,
			wantContains: "Error parsing filesystem usage percent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := core.NewMockProvider()
			tt.setupMock(mock)

			ctx := context.Background()

			result := executeFilesystemTest(ctx, mock, tt.filesystemTest)

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
