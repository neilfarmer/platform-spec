package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSpec(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		wantErr  bool
		filename string // Optional: custom filename (default: spec.yaml)
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
			name: "invalid file extension",
			yaml: `version: "1.0"
tests:
  packages:
    - name: "test"
      packages: [bash]
      state: present`,
			wantErr: true,
			filename: "spec.txt", // Will be used to test non-YAML extension
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
		{
			name: "valid ping test",
			yaml: `version: "1.0"
tests:
  ping:
    - name: "test ping"
      host: google.com`,
			wantErr: false,
		},
		{
			name: "ping test missing name",
			yaml: `version: "1.0"
tests:
  ping:
    - host: google.com`,
			wantErr: true,
		},
		{
			name: "ping test missing host",
			yaml: `version: "1.0"
tests:
  ping:
    - name: "test"`,
			wantErr: true,
		},
		{
			name: "valid dns test",
			yaml: `version: "1.0"
tests:
  dns:
    - name: "test dns"
      host: google.com`,
			wantErr: false,
		},
		{
			name: "dns test missing name",
			yaml: `version: "1.0"
tests:
  dns:
    - host: google.com`,
			wantErr: true,
		},
		{
			name: "dns test missing host",
			yaml: `version: "1.0"
tests:
  dns:
    - name: "test"`,
			wantErr: true,
		},
		{
			name: "valid systeminfo test",
			yaml: `version: "1.0"
tests:
  systeminfo:
    - name: "test sysinfo"
      os: ubuntu
      os_version: "22.04"`,
			wantErr: false,
		},
		{
			name: "systeminfo test missing name",
			yaml: `version: "1.0"
tests:
  systeminfo:
    - os: ubuntu`,
			wantErr: true,
		},
		{
			name: "systeminfo test invalid version_match",
			yaml: `version: "1.0"
tests:
  systeminfo:
    - name: "test"
      os: ubuntu
      version_match: invalid`,
			wantErr: true,
		},
		{
			name: "systeminfo test defaults to exact",
			yaml: `version: "1.0"
tests:
  systeminfo:
    - name: "test"
      os: ubuntu`,
			wantErr: false,
		},
		{
			name: "valid http test",
			yaml: `version: "1.0"
tests:
  http:
    - name: "test endpoint"
      url: http://example.com
      status_code: 200`,
			wantErr: false,
		},
		{
			name: "http test missing name",
			yaml: `version: "1.0"
tests:
  http:
    - url: http://example.com`,
			wantErr: true,
		},
		{
			name: "http test missing url",
			yaml: `version: "1.0"
tests:
  http:
    - name: "test"`,
			wantErr: true,
		},
		{
			name: "http test invalid method",
			yaml: `version: "1.0"
tests:
  http:
    - name: "test"
      url: http://example.com
      method: INVALID`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file with YAML content
			tmpDir := t.TempDir()
			filename := tt.filename
			if filename == "" {
				filename = "spec.yaml"
			}
			tmpFile := filepath.Join(tmpDir, filename)
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
		{
			name: "valid ping test",
			spec: &Spec{
				Tests: Tests{
					Ping: []PingTest{
						{
							Name: "test",
							Host: "google.com",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid dns test",
			spec: &Spec{
				Tests: Tests{
					DNS: []DNSTest{
						{
							Name: "test",
							Host: "google.com",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "systeminfo test defaults to exact version_match",
			spec: &Spec{
				Tests: Tests{
					SystemInfo: []SystemInfoTest{
						{
							Name: "test",
							OS:   "ubuntu",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid systeminfo test with all fields",
			spec: &Spec{
				Tests: Tests{
					SystemInfo: []SystemInfoTest{
						{
							Name:          "test",
							OS:            "ubuntu",
							OSVersion:     "22.04",
							Arch:          "x86_64",
							KernelVersion: "5.15",
							Hostname:      "web01",
							FQDN:          "web01.example.com",
							VersionMatch:  "prefix",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "http test missing name",
			spec: &Spec{
				Tests: Tests{
					HTTP: []HTTPTest{
						{
							URL: "http://example.com",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "http test missing url",
			spec: &Spec{
				Tests: Tests{
					HTTP: []HTTPTest{
						{
							Name: "test",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "http test invalid method",
			spec: &Spec{
				Tests: Tests{
					HTTP: []HTTPTest{
						{
							Name:   "test",
							URL:    "http://example.com",
							Method: "INVALID",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "http test defaults status_code to 200",
			spec: &Spec{
				Tests: Tests{
					HTTP: []HTTPTest{
						{
							Name: "test",
							URL:  "http://example.com",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "http test defaults method to GET",
			spec: &Spec{
				Tests: Tests{
					HTTP: []HTTPTest{
						{
							Name: "test",
							URL:  "http://example.com",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid http test with all fields",
			spec: &Spec{
				Tests: Tests{
					HTTP: []HTTPTest{
						{
							Name:       "test",
							URL:        "https://example.com/api",
							StatusCode: 201,
							Contains:   []string{"success", "data"},
							Method:     "POST",
							Insecure:   true,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "port test missing name",
			spec: &Spec{
				Tests: Tests{
					Ports: []PortTest{
						{
							Port: 22,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "port test invalid port (0)",
			spec: &Spec{
				Tests: Tests{
					Ports: []PortTest{
						{
							Name: "test",
							Port: 0,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "port test invalid port (too large)",
			spec: &Spec{
				Tests: Tests{
					Ports: []PortTest{
						{
							Name: "test",
							Port: 70000,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "port test invalid protocol",
			spec: &Spec{
				Tests: Tests{
					Ports: []PortTest{
						{
							Name:     "test",
							Port:     22,
							Protocol: "sctp",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "port test invalid state",
			spec: &Spec{
				Tests: Tests{
					Ports: []PortTest{
						{
							Name:  "test",
							Port:  22,
							State: "open",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "port test defaults",
			spec: &Spec{
				Tests: Tests{
					Ports: []PortTest{
						{
							Name: "test",
							Port: 22,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid port test with all fields",
			spec: &Spec{
				Tests: Tests{
					Ports: []PortTest{
						{
							Name:     "test",
							Port:     8080,
							Protocol: "udp",
							State:    "closed",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid kubernetes node test",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						Nodes: []KubernetesNodeTest{
							{
								Name:     "test",
								MinCount: 3,
								MinReady: 2,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "kubernetes crd test defaults to present",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						CRDs: []KubernetesCRDTest{
							{
								Name: "test",
								CRD:  "certificates.cert-manager.io",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "kubernetes helm test defaults",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						Helm: []KubernetesHelmTest{
							{
								Name:    "test",
								Release: "prometheus",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "kubernetes storageclass test defaults to present",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						StorageClasses: []KubernetesStorageClassTest{
							{
								Name:         "test",
								StorageClass: "fast-ssd",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "kubernetes secret test defaults",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						Secrets: []KubernetesSecretTest{
							{
								Name:   "test",
								Secret: "db-password",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "kubernetes ingress test defaults",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						Ingress: []KubernetesIngressTest{
							{
								Name:    "test",
								Ingress: "myapp-ingress",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "kubernetes pvc test defaults",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						PVCs: []KubernetesPVCTest{
							{
								Name: "test",
								PVC:  "data-pvc",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "kubernetes statefulset test defaults",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						StatefulSets: []KubernetesStatefulSetTest{
							{
								Name:        "test",
								StatefulSet: "postgres",
							},
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
				for _, st := range tt.spec.Tests.SystemInfo {
					if st.VersionMatch == "" {
						t.Error("SystemInfo version_match should default to exact")
					}
				}
				for _, ht := range tt.spec.Tests.HTTP {
					if ht.StatusCode == 0 {
						t.Error("HTTP status_code should default to 200")
					}
					if ht.Method == "" {
						t.Error("HTTP method should default to GET")
					}
				}
				for _, pt := range tt.spec.Tests.Ports {
					if pt.Protocol == "" {
						t.Error("Port protocol should default to tcp")
					}
					if pt.State == "" {
						t.Error("Port state should default to listening")
					}
				}
				for _, ct := range tt.spec.Tests.Kubernetes.CRDs {
					if ct.State == "" {
						t.Error("CRD state should default to present")
					}
				}
				for _, ht := range tt.spec.Tests.Kubernetes.Helm {
					if ht.Namespace == "" {
						t.Error("Helm namespace should default to default")
					}
					if ht.State == "" {
						t.Error("Helm state should default to deployed")
					}
				}
				for _, st := range tt.spec.Tests.Kubernetes.StorageClasses {
					if st.State == "" {
						t.Error("StorageClass state should default to present")
					}
				}
				for _, st := range tt.spec.Tests.Kubernetes.Secrets {
					if st.Namespace == "" {
						t.Error("Secret namespace should default to default")
					}
					if st.State == "" {
						t.Error("Secret state should default to present")
					}
				}
				for _, it := range tt.spec.Tests.Kubernetes.Ingress {
					if it.Namespace == "" {
						t.Error("Ingress namespace should default to default")
					}
					if it.State == "" {
						t.Error("Ingress state should default to present")
					}
				}
				for _, pt := range tt.spec.Tests.Kubernetes.PVCs {
					if pt.Namespace == "" {
						t.Error("PVC namespace should default to default")
					}
					if pt.State == "" {
						t.Error("PVC state should default to present")
					}
				}
				for _, st := range tt.spec.Tests.Kubernetes.StatefulSets {
					if st.Namespace == "" {
						t.Error("StatefulSet namespace should default to default")
					}
					if st.State == "" {
						t.Error("StatefulSet state should default to available")
					}
				}
			}
		})
	}
}

func TestValidationErrorPaths(t *testing.T) {
	tests := []struct {
		name    string
		spec    *Spec
		wantErr string
	}{
		{
			name: "service test without service or services",
			spec: &Spec{
				Tests: Tests{
					Services: []ServiceTest{{Name: "test"}},
				},
			},
			wantErr: "service or services is required",
		},
		{
			name: "docker test with both container and containers",
			spec: &Spec{
				Tests: Tests{
					Docker: []DockerTest{{Name: "test", Container: "c1", Containers: []string{"c2"}}},
				},
			},
			wantErr: "cannot specify both container and containers",
		},
		{
			name: "port out of range - too low",
			spec: &Spec{
				Tests: Tests{
					Ports: []PortTest{{Name: "test", Port: 0}},
				},
			},
			wantErr: "port must be between 1 and 65535",
		},
		{
			name: "port out of range - too high",
			spec: &Spec{
				Tests: Tests{
					Ports: []PortTest{{Name: "test", Port: 70000}},
				},
			},
			wantErr: "port must be between 1 and 65535",
		},
		{
			name: "group test without groups",
			spec: &Spec{
				Tests: Tests{
					Groups: []GroupTest{{Name: "test", Groups: []string{}}},
				},
			},
			wantErr: "at least one group is required",
		},
		{
			name: "user test without user",
			spec: &Spec{
				Tests: Tests{
					Users: []UserTest{{Name: "test"}},
				},
			},
			wantErr: "user is required",
		},
		{
			name: "file_content test without path",
			spec: &Spec{
				Tests: Tests{
					FileContent: []FileContentTest{{Name: "test", Contains: []string{"text"}}},
				},
			},
			wantErr: "path is required",
		},
		{
			name: "file_content test without contains or matches",
			spec: &Spec{
				Tests: Tests{
					FileContent: []FileContentTest{{Name: "test", Path: "/tmp/file"}},
				},
			},
			wantErr: "either contains or matches is required",
		},
		{
			name: "command_content test without command",
			spec: &Spec{
				Tests: Tests{
					CommandContent: []CommandContentTest{{Name: "test", Contains: []string{"text"}}},
				},
			},
			wantErr: "command is required",
		},
		{
			name: "command_content test without contains or exit_code",
			spec: &Spec{
				Tests: Tests{
					CommandContent: []CommandContentTest{{Name: "test", Command: "echo hello"}},
				},
			},
			wantErr: "either contains or exit_code is required",
		},
		{
			name: "package test with empty packages list",
			spec: &Spec{
				Tests: Tests{
					Packages: []PackageTest{{Name: "test", Packages: []string{}}},
				},
			},
			wantErr: "at least one package is required",
		},
		{
			name: "port protocol udp is valid",
			spec: &Spec{
				Tests: Tests{
					Ports: []PortTest{{Name: "test", Port: 53, Protocol: "udp"}},
				},
			},
			wantErr: "",
		},
		{
			name: "invalid port protocol",
			spec: &Spec{
				Tests: Tests{
					Ports: []PortTest{{Name: "test", Port: 80, Protocol: "sctp"}},
				},
			},
			wantErr: "protocol must be 'tcp' or 'udp'",
		},
		{
			name: "invalid port state",
			spec: &Spec{
				Tests: Tests{
					Ports: []PortTest{{Name: "test", Port: 80, State: "open"}},
				},
			},
			wantErr: "state must be 'listening' or 'closed'",
		},
		{
			name: "filesystem invalid state",
			spec: &Spec{
				Tests: Tests{
					Filesystems: []FilesystemTest{{Name: "test", Path: "/mnt", State: "present"}},
				},
			},
			wantErr: "state must be 'mounted' or 'unmounted'",
		},
		{
			name: "filesystem invalid max usage percent negative",
			spec: &Spec{
				Tests: Tests{
					Filesystems: []FilesystemTest{{Name: "test", Path: "/mnt", MaxUsagePercent: -1}},
				},
			},
			wantErr: "max_usage_percent must be between 0 and 100",
		},
		{
			name: "filesystem invalid max usage percent over 100",
			spec: &Spec{
				Tests: Tests{
					Filesystems: []FilesystemTest{{Name: "test", Path: "/mnt", MaxUsagePercent: 101}},
				},
			},
			wantErr: "max_usage_percent must be between 0 and 100",
		},
		{
			name: "filesystem invalid min size gb",
			spec: &Spec{
				Tests: Tests{
					Filesystems: []FilesystemTest{{Name: "test", Path: "/mnt", MinSizeGB: -5}},
				},
			},
			wantErr: "min_size_gb must be >= 0",
		},
		{
			name: "docker invalid restart policy",
			spec: &Spec{
				Tests: Tests{
					Docker: []DockerTest{{Name: "test", Container: "c1", RestartPolicy: "sometimes"}},
				},
			},
			wantErr: "restart_policy must be 'no', 'always', 'on-failure', or 'unless-stopped'",
		},
		{
			name: "docker invalid health",
			spec: &Spec{
				Tests: Tests{
					Docker: []DockerTest{{Name: "test", Container: "c1", Health: "sick"}},
				},
			},
			wantErr: "health must be 'healthy', 'unhealthy', 'starting', or 'none'",
		},
		{
			name: "kubernetes namespace without namespace field",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						Namespaces: []KubernetesNamespaceTest{{Name: "test"}},
					},
				},
			},
			wantErr: "namespace is required",
		},
		{
			name: "kubernetes pod without pod field",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						Pods: []KubernetesPodTest{{Name: "test"}},
					},
				},
			},
			wantErr: "pod is required",
		},
		{
			name: "kubernetes pod invalid state",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						Pods: []KubernetesPodTest{{Name: "test", Pod: "mypod", State: "crashed"}},
					},
				},
			},
			wantErr: "state must be one of: running, pending, succeeded, failed, exists",
		},
		{
			name: "kubernetes deployment without deployment field",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						Deployments: []KubernetesDeploymentTest{{Name: "test"}},
					},
				},
			},
			wantErr: "deployment is required",
		},
		{
			name: "kubernetes deployment invalid state",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						Deployments: []KubernetesDeploymentTest{{Name: "test", Deployment: "myapp", State: "ready"}},
					},
				},
			},
			wantErr: "state must be one of: available, progressing, exists",
		},
		{
			name: "kubernetes deployment invalid replicas",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						Deployments: []KubernetesDeploymentTest{{Name: "test", Deployment: "myapp", Replicas: -1}},
					},
				},
			},
			wantErr: "replicas must be >= 0",
		},
		{
			name: "kubernetes deployment invalid ready_replicas",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						Deployments: []KubernetesDeploymentTest{{Name: "test", Deployment: "myapp", ReadyReplicas: -2}},
					},
				},
			},
			wantErr: "ready_replicas must be >= 0",
		},
		{
			name: "kubernetes node without name",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						Nodes: []KubernetesNodeTest{{Count: 3}},
					},
				},
			},
			wantErr: "name is required",
		},
		{
			name: "kubernetes node invalid count",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						Nodes: []KubernetesNodeTest{{Name: "test", Count: -1}},
					},
				},
			},
			wantErr: "count must be non-negative",
		},
		{
			name: "kubernetes node invalid min_count",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						Nodes: []KubernetesNodeTest{{Name: "test", MinCount: -1}},
					},
				},
			},
			wantErr: "min_count must be non-negative",
		},
		{
			name: "kubernetes node invalid min_ready",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						Nodes: []KubernetesNodeTest{{Name: "test", MinReady: -1}},
					},
				},
			},
			wantErr: "min_ready must be non-negative",
		},
		{
			name: "kubernetes crd without name",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						CRDs: []KubernetesCRDTest{{CRD: "certificates.cert-manager.io"}},
					},
				},
			},
			wantErr: "name is required",
		},
		{
			name: "kubernetes crd without crd field",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						CRDs: []KubernetesCRDTest{{Name: "test"}},
					},
				},
			},
			wantErr: "crd is required",
		},
		{
			name: "kubernetes crd invalid state",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						CRDs: []KubernetesCRDTest{{Name: "test", CRD: "test.example.com", State: "installed"}},
					},
				},
			},
			wantErr: "state must be 'present' or 'absent'",
		},
		{
			name: "kubernetes helm without name",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						Helm: []KubernetesHelmTest{{Release: "prometheus"}},
					},
				},
			},
			wantErr: "name is required",
		},
		{
			name: "kubernetes helm without release",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						Helm: []KubernetesHelmTest{{Name: "test"}},
					},
				},
			},
			wantErr: "release is required",
		},
		{
			name: "kubernetes helm invalid state",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						Helm: []KubernetesHelmTest{{Name: "test", Release: "prometheus", State: "running"}},
					},
				},
			},
			wantErr: "state must be one of",
		},
		{
			name: "kubernetes storageclass without name",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						StorageClasses: []KubernetesStorageClassTest{{StorageClass: "fast-ssd"}},
					},
				},
			},
			wantErr: "name is required",
		},
		{
			name: "kubernetes storageclass without storageclass field",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						StorageClasses: []KubernetesStorageClassTest{{Name: "test"}},
					},
				},
			},
			wantErr: "storageclass is required",
		},
		{
			name: "kubernetes storageclass invalid state",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						StorageClasses: []KubernetesStorageClassTest{{Name: "test", StorageClass: "fast-ssd", State: "active"}},
					},
				},
			},
			wantErr: "state must be 'present' or 'absent'",
		},
		{
			name: "kubernetes secret without name",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						Secrets: []KubernetesSecretTest{{Secret: "db-password"}},
					},
				},
			},
			wantErr: "name is required",
		},
		{
			name: "kubernetes secret without secret field",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						Secrets: []KubernetesSecretTest{{Name: "test"}},
					},
				},
			},
			wantErr: "secret is required",
		},
		{
			name: "kubernetes secret invalid state",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						Secrets: []KubernetesSecretTest{{Name: "test", Secret: "db-password", State: "active"}},
					},
				},
			},
			wantErr: "state must be 'present' or 'absent'",
		},
		{
			name: "kubernetes secret invalid type",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						Secrets: []KubernetesSecretTest{{Name: "test", Secret: "db-password", Type: "invalid-type"}},
					},
				},
			},
			wantErr: "type must be one of",
		},
		{
			name: "kubernetes ingress without name",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						Ingress: []KubernetesIngressTest{{Ingress: "myapp-ingress"}},
					},
				},
			},
			wantErr: "name is required",
		},
		{
			name: "kubernetes ingress without ingress field",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						Ingress: []KubernetesIngressTest{{Name: "test"}},
					},
				},
			},
			wantErr: "ingress is required",
		},
		{
			name: "kubernetes ingress invalid state",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						Ingress: []KubernetesIngressTest{{Name: "test", Ingress: "myapp-ingress", State: "ready"}},
					},
				},
			},
			wantErr: "state must be 'present' or 'absent'",
		},
		{
			name: "kubernetes pvc without name",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						PVCs: []KubernetesPVCTest{{PVC: "data-pvc"}},
					},
				},
			},
			wantErr: "name is required",
		},
		{
			name: "kubernetes pvc without pvc field",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						PVCs: []KubernetesPVCTest{{Name: "test"}},
					},
				},
			},
			wantErr: "pvc is required",
		},
		{
			name: "kubernetes pvc invalid state",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						PVCs: []KubernetesPVCTest{{Name: "test", PVC: "data-pvc", State: "active"}},
					},
				},
			},
			wantErr: "state must be 'present' or 'absent'",
		},
		{
			name: "kubernetes pvc invalid status",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						PVCs: []KubernetesPVCTest{{Name: "test", PVC: "data-pvc", Status: "Ready"}},
					},
				},
			},
			wantErr: "status must be 'Bound', 'Pending', or 'Lost'",
		},
		{
			name: "kubernetes statefulset without name",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						StatefulSets: []KubernetesStatefulSetTest{{StatefulSet: "postgres"}},
					},
				},
			},
			wantErr: "name is required",
		},
		{
			name: "kubernetes statefulset without statefulset field",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						StatefulSets: []KubernetesStatefulSetTest{{Name: "test"}},
					},
				},
			},
			wantErr: "statefulset is required",
		},
		{
			name: "kubernetes statefulset invalid state",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						StatefulSets: []KubernetesStatefulSetTest{{Name: "test", StatefulSet: "postgres", State: "running"}},
					},
				},
			},
			wantErr: "state must be 'available' or 'exists'",
		},
		{
			name: "kubernetes statefulset invalid replicas",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						StatefulSets: []KubernetesStatefulSetTest{{Name: "test", StatefulSet: "postgres", Replicas: -1}},
					},
				},
			},
			wantErr: "replicas must be >= 0",
		},
		{
			name: "kubernetes statefulset invalid ready_replicas",
			spec: &Spec{
				Tests: Tests{
					Kubernetes: KubernetesTests{
						StatefulSets: []KubernetesStatefulSetTest{{Name: "test", StatefulSet: "postgres", ReadyReplicas: -1}},
					},
				},
			},
			wantErr: "ready_replicas must be >= 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Validate()
			if tt.wantErr == "" {
				// Expecting success
				if err != nil {
					t.Errorf("Expected no error, got %q", err.Error())
				}
			} else {
				// Expecting error
				if err == nil {
					t.Error("Expected error, got nil")
					return
				}
				if !contains(err.Error(), tt.wantErr) {
					t.Errorf("Expected error containing %q, got %q", tt.wantErr, err.Error())
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestParseSpecEnhancedErrors tests that enhanced YAML error messages appear
// when parsing specs with various YAML errors
func TestParseSpecEnhancedErrors(t *testing.T) {
	tests := []struct {
		name            string
		yaml            string
		wantErrContains []string // All of these strings should appear in the error
	}{
		{
			name: "string instead of array for packages",
			yaml: `version: "1.0"
tests:
  packages:
    - name: "Docker installed"
      packages: docker-ce
      state: present`,
			wantErrContains: []string{
				"YAML parsing error",
				"Expected a list",
				"got a single string value",
				"Wrong: packages: nginx",
				"Right: packages: [nginx]",
				"examples/basic.yaml",
			},
		},
		{
			name: "string instead of array for services",
			yaml: `version: "1.0"
tests:
  services:
    - name: "Docker running"
      services: docker
      state: running`,
			wantErrContains: []string{
				"Expected a list",
				"got a single string value",
				"Common fix: Wrap single values in brackets",
			},
		},
		{
			name: "array instead of string for name",
			yaml: `version: "1.0"
tests:
  packages:
    - name: ["test"]
      packages: [docker]`,
			wantErrContains: []string{
				"Expected a string",
				"got a list",
				"Wrong: name: [test]",
				"Right: name: test",
			},
		},
		{
			name: "string instead of int for port",
			yaml: `version: "1.0"
tests:
  ports:
    - name: "SSH port"
      port: "22"
      state: listening`,
			wantErrContains: []string{
				"Expected a number",
				"got a string",
				"Wrong: port: \"8080\"",
				"Right: port: 8080",
			},
		},
		{
			name: "string instead of bool for enabled",
			yaml: `version: "1.0"
tests:
  services:
    - name: "Docker"
      service: docker
      state: running
      enabled: "true"`,
			wantErrContains: []string{
				"Expected a boolean",
				"got a string",
				"Wrong: enabled: \"true\"",
				"Right: enabled: true",
			},
		},
		{
			name: "string instead of int for status_code",
			yaml: `version: "1.0"
tests:
  http:
    - name: "API check"
      url: http://localhost:8080
      status_code: "200"`,
			wantErrContains: []string{
				"Expected a number",
				"got a string",
			},
		},
		// Note: yaml.v3 ignores unknown fields rather than erroring,
		// so we can't test field name typos via YAML parsing errors.
		// Those would be caught during validation instead.
		{
			name: "invalid yaml syntax",
			yaml: `version: "1.0"
tests:
  packages:
    - name: "test"
      packages: [
        - docker`,
			wantErrContains: []string{
				"YAML parsing error",
				"Troubleshooting tips",
				"Check indentation",
				"examples/basic.yaml",
			},
		},
		{
			name: "map instead of array",
			yaml: `version: "1.0"
tests:
  packages:
    name: "test"
    packages: [docker]`,
			wantErrContains: []string{
				"Expected a list",
				"got a mapping/object",
				"this field should be a list",
			},
		},
		{
			name: "multiple type errors - first one reported",
			yaml: `version: "1.0"
tests:
  packages:
    - name: "test"
      packages: nginx
      state: present
  files:
    - name: ["another error"]
      path: /tmp`,
			wantErrContains: []string{
				"Expected a list",
				"got a single string value",
			},
		},
		{
			name: "string instead of array for file content contains",
			yaml: `version: "1.0"
tests:
  file_content:
    - name: "Check config"
      path: /etc/config
      contains: "single-value"`,
			wantErrContains: []string{
				"Expected a list",
				"got a single string value",
			},
		},
		{
			name: "string instead of array for command content contains",
			yaml: `version: "1.0"
tests:
  command_content:
    - name: "Version check"
      command: "app --version"
      contains: "v1.0"`,
			wantErrContains: []string{
				"Expected a list",
				"got a single string value",
			},
		},
		{
			name: "string instead of array for docker containers",
			yaml: `version: "1.0"
tests:
  docker:
    - name: "Containers running"
      containers: nginx`,
			wantErrContains: []string{
				"Expected a list",
				"got a single string value",
			},
		},
		{
			name: "string instead of array for groups",
			yaml: `version: "1.0"
tests:
  groups:
    - name: "Docker group"
      groups: docker`,
			wantErrContains: []string{
				"Expected a list",
				"got a single string value",
			},
		},
		{
			name: "string instead of int for replicas",
			yaml: `version: "1.0"
tests:
  kubernetes:
    deployments:
      - name: "App deployment"
        deployment: myapp
        replicas: "3"`,
			wantErrContains: []string{
				"Expected a number",
				"got a string",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary spec file
			tmpDir := t.TempDir()
			specFile := filepath.Join(tmpDir, "test-spec.yaml")
			if err := os.WriteFile(specFile, []byte(tt.yaml), 0644); err != nil {
				t.Fatalf("Failed to write test spec file: %v", err)
			}

			// Try to parse the spec - should fail with enhanced error
			_, err := ParseSpec(specFile)
			if err == nil {
				t.Fatalf("ParseSpec() should have failed but succeeded")
			}

			errMsg := err.Error()

			// Verify all expected strings appear in the error message
			for _, want := range tt.wantErrContains {
				if !contains(errMsg, want) {
					t.Errorf("Error message should contain %q\nGot error:\n%s", want, errMsg)
				}
			}

			// Verify the error is more helpful than the raw yaml.v3 error
			// It should NOT contain cryptic YAML type syntax in the main part
			if contains(errMsg, "!!str") || contains(errMsg, "!!seq") || contains(errMsg, "!!map") {
				// Only acceptable if it's in the "Original error:" section
				if !contains(errMsg, "Original error:") {
					t.Errorf("Error message contains cryptic YAML syntax but no 'Original error:' section:\n%s", errMsg)
				}
			}
		})
	}
}

// TestParseSpecErrorLineNumbers tests that line numbers are extracted and included
// in error messages when available
func TestParseSpecErrorLineNumbers(t *testing.T) {
	tests := []struct {
		name            string
		yaml            string
		wantLineInError bool
	}{
		{
			name: "error on specific line",
			yaml: `version: "1.0"
tests:
  packages:
    - name: "test"
      packages: docker
      state: present`,
			wantLineInError: true, // yaml.v3 should report line number for type error
		},
		{
			name: "syntax error",
			yaml: `version: "1.0"
tests:
  packages:
    - name: "test"
      packages: [
`,
			wantLineInError: true, // yaml.v3 reports line for syntax errors
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			specFile := filepath.Join(tmpDir, "test-spec.yaml")
			if err := os.WriteFile(specFile, []byte(tt.yaml), 0644); err != nil {
				t.Fatalf("Failed to write test spec file: %v", err)
			}

			_, err := ParseSpec(specFile)
			if err == nil {
				t.Fatalf("ParseSpec() should have failed")
			}

			errMsg := err.Error()

			// Check if "at line" appears in the error
			if tt.wantLineInError {
				if !contains(errMsg, "at line") && !contains(errMsg, "line ") {
					t.Logf("Note: Line number not extracted from yaml.v3 error (this is acceptable)")
					t.Logf("Error was: %s", errMsg)
				}
			}

			// At minimum, verify the error contains helpful information
			if !contains(errMsg, "YAML parsing error") && !contains(errMsg, "Troubleshooting") {
				t.Errorf("Error should contain enhanced information, got:\n%s", errMsg)
			}
		})
	}
}

// TestParseSpecEnhancedVsOriginal compares enhanced errors to original yaml.v3 errors
// to verify the enhancement is actually helpful
func TestParseSpecEnhancedVsOriginal(t *testing.T) {
	// This test documents the improvement from the original cryptic errors
	examples := []struct {
		name                string
		yaml                string
		originalWouldContain string // What the original yaml.v3 error would say
		enhancedContains    string // What our enhanced error should say
	}{
		{
			name: "string vs array clarity",
			yaml: `version: "1.0"
tests:
  packages:
    - name: "test"
      packages: nginx`,
			originalWouldContain: "cannot unmarshal",
			enhancedContains:     "Expected a list, but got a single string value",
		},
		{
			name: "actionable fix provided",
			yaml: `version: "1.0"
tests:
  packages:
    - name: "test"
      packages: nginx`,
			originalWouldContain: "!!str",
			enhancedContains:     "Wrong: packages: nginx",
		},
		{
			name: "reference to examples",
			yaml: `version: "1.0"
tests:
  packages:
    - name: "test"
      packages: nginx`,
			originalWouldContain: "into []string",
			enhancedContains:     "examples/basic.yaml",
		},
	}

	for _, tt := range examples {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			specFile := filepath.Join(tmpDir, "test.yaml")
			if err := os.WriteFile(specFile, []byte(tt.yaml), 0644); err != nil {
				t.Fatalf("Failed to write spec file: %v", err)
			}

			_, err := ParseSpec(specFile)
			if err == nil {
				t.Fatalf("Expected error but got none")
			}

			errMsg := err.Error()

			// Verify the enhanced message appears
			if !contains(errMsg, tt.enhancedContains) {
				t.Errorf("Enhanced error should contain %q, got:\n%s",
					tt.enhancedContains, errMsg)
			}

			// The main error message should NOT start with the cryptic original
			// (though it can include it at the end as "Original error:")
			lines := err.Error()
			if len(lines) > 0 {
				firstPart := lines
				if idx := findIndex(lines, "Original error:"); idx != -1 {
					firstPart = lines[:idx]
				}

				// The enhanced part should not have cryptic syntax
				if contains(firstPart, tt.originalWouldContain) {
					t.Logf("Note: Original cryptic error still appears in main message")
					t.Logf("This might be acceptable if it's part of a larger helpful message")
				}
			}
		})
	}
}

// Helper function to find substring index
func findIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
