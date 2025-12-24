package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSpec(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{
			name: "valid minimal spec",
			yaml: `version: "1.0"
tests:
  packages:
    - name: "test"
      packages: [bash]
      state: present`,
			wantErr: false,
		},
		{
			name: "valid spec with metadata",
			yaml: `version: "1.0"
metadata:
  name: "Test Suite"
  description: "Test description"
  tags: ["tag1", "tag2"]
tests:
  packages:
    - name: "test"
      packages: [bash]
      state: present`,
			wantErr: false,
		},
		{
			name: "invalid yaml",
			yaml: `this is not valid yaml: [[[`,
			wantErr: true,
		},
		{
			name: "missing package name",
			yaml: `version: "1.0"
tests:
  packages:
    - packages: [bash]
      state: present`,
			wantErr: true,
		},
		{
			name: "invalid package state",
			yaml: `version: "1.0"
tests:
  packages:
    - name: "test"
      packages: [bash]
      state: invalid`,
			wantErr: true,
		},
		{
			name: "missing file path",
			yaml: `version: "1.0"
tests:
  files:
    - name: "test"
      type: file`,
			wantErr: true,
		},
		{
			name: "invalid file type",
			yaml: `version: "1.0"
tests:
  files:
    - name: "test"
      path: /tmp
      type: invalid`,
			wantErr: true,
		},
		{
			name: "valid docker test",
			yaml: `version: "1.0"
tests:
  docker:
    - name: "test container"
      container: nginx
      state: running`,
			wantErr: false,
		},
		{
			name: "docker test missing name",
			yaml: `version: "1.0"
tests:
  docker:
    - container: nginx
      state: running`,
			wantErr: true,
		},
		{
			name: "docker test missing container",
			yaml: `version: "1.0"
tests:
  docker:
    - name: "test"
      state: running`,
			wantErr: true,
		},
		{
			name: "docker test invalid state",
			yaml: `version: "1.0"
tests:
  docker:
    - name: "test"
      container: nginx
      state: invalid`,
			wantErr: true,
		},
		{
			name: "docker test invalid restart policy",
			yaml: `version: "1.0"
tests:
  docker:
    - name: "test"
      container: nginx
      state: running
      restart_policy: invalid`,
			wantErr: true,
		},
		{
			name: "docker test invalid health",
			yaml: `version: "1.0"
tests:
  docker:
    - name: "test"
      container: nginx
      state: running
      health: invalid`,
			wantErr: true,
		},
		{
			name: "valid filesystem test",
			yaml: `version: "1.0"
tests:
  filesystems:
    - name: "root filesystem"
      path: /
      state: mounted`,
			wantErr: false,
		},
		{
			name: "filesystem test missing name",
			yaml: `version: "1.0"
tests:
  filesystems:
    - path: /
      state: mounted`,
			wantErr: true,
		},
		{
			name: "filesystem test missing path",
			yaml: `version: "1.0"
tests:
  filesystems:
    - name: "test"
      state: mounted`,
			wantErr: true,
		},
		{
			name: "filesystem test invalid state",
			yaml: `version: "1.0"
tests:
  filesystems:
    - name: "test"
      path: /
      state: invalid`,
			wantErr: true,
		},
		{
			name: "filesystem test invalid max_usage_percent negative",
			yaml: `version: "1.0"
tests:
  filesystems:
    - name: "test"
      path: /
      state: mounted
      max_usage_percent: -1`,
			wantErr: true,
		},
		{
			name: "filesystem test invalid max_usage_percent over 100",
			yaml: `version: "1.0"
tests:
  filesystems:
    - name: "test"
      path: /
      state: mounted
      max_usage_percent: 101`,
			wantErr: true,
		},
		{
			name: "filesystem test invalid min_size_gb negative",
			yaml: `version: "1.0"
tests:
  filesystems:
    - name: "test"
      path: /
      state: mounted
      min_size_gb: -1`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file with YAML content
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "spec.yaml")
			if err := os.WriteFile(tmpFile, []byte(tt.yaml), 0644); err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			_, err := ParseSpec(tmpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSpec() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSpecValidation(t *testing.T) {
	tests := []struct {
		name    string
		spec    *Spec
		wantErr bool
	}{
		{
			name: "valid package test",
			spec: &Spec{
				Tests: Tests{
					Packages: []PackageTest{
						{
							Name:     "test",
							Packages: []string{"bash"},
							State:    "present",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "package test defaults to present",
			spec: &Spec{
				Tests: Tests{
					Packages: []PackageTest{
						{
							Name:     "test",
							Packages: []string{"bash"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "file test defaults to file type",
			spec: &Spec{
				Tests: Tests{
					Files: []FileTest{
						{
							Name: "test",
							Path: "/tmp",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "spec defaults to version 1.0",
			spec: &Spec{
				Tests: Tests{
					Packages: []PackageTest{
						{
							Name:     "test",
							Packages: []string{"bash"},
							State:    "present",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "docker test defaults to running state",
			spec: &Spec{
				Tests: Tests{
					Docker: []DockerTest{
						{
							Name:      "test",
							Container: "nginx",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid docker test with properties",
			spec: &Spec{
				Tests: Tests{
					Docker: []DockerTest{
						{
							Name:          "test",
							Container:     "nginx",
							State:         "running",
							Image:         "nginx:latest",
							RestartPolicy: "always",
							Health:        "healthy",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "filesystem test defaults to mounted state",
			spec: &Spec{
				Tests: Tests{
					Filesystems: []FilesystemTest{
						{
							Name: "test",
							Path: "/",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid filesystem test with all properties",
			spec: &Spec{
				Tests: Tests{
					Filesystems: []FilesystemTest{
						{
							Name:            "test",
							Path:            "/data",
							State:           "mounted",
							Fstype:          "ext4",
							Options:         []string{"rw", "noexec"},
							MinSizeGB:       100,
							MaxUsagePercent: 80,
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Spec.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check defaults are set
			if !tt.wantErr {
				if tt.spec.Version == "" {
					t.Error("Version should default to 1.0")
				}
				for _, pt := range tt.spec.Tests.Packages {
					if pt.State == "" {
						t.Error("Package state should default to present")
					}
				}
				for _, ft := range tt.spec.Tests.Files {
					if ft.Type == "" {
						t.Error("File type should default to file")
					}
				}
				for _, dt := range tt.spec.Tests.Docker {
					if dt.State == "" {
						t.Error("Docker state should default to running")
					}
				}
				for _, ft := range tt.spec.Tests.Filesystems {
					if ft.State == "" {
						t.Error("Filesystem state should default to mounted")
					}
				}
			}
		})
	}
}
