package core

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Spec represents the parsed YAML specification
type Spec struct {
	Version   string                 `yaml:"version"`
	Metadata  SpecMetadata           `yaml:"metadata"`
	Config    SpecConfig             `yaml:"config"`
	Variables map[string]interface{} `yaml:"variables"`
	Tests     Tests                  `yaml:"tests"`
}

// SpecMetadata contains metadata about the spec
type SpecMetadata struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags"`
}

// SpecConfig contains configuration options
type SpecConfig struct {
	FailFast bool `yaml:"fail_fast"`
	Parallel bool `yaml:"parallel"`
	Timeout  int  `yaml:"timeout"`
}

// Tests contains all test definitions
type Tests struct {
	Packages       []PackageTest        `yaml:"packages"`
	Files          []FileTest           `yaml:"files"`
	Services       []ServiceTest        `yaml:"services"`
	Users          []UserTest           `yaml:"users"`
	Groups         []GroupTest          `yaml:"groups"`
	FileContent    []FileContentTest    `yaml:"file_content"`
	CommandContent []CommandContentTest `yaml:"command_content"`
	Docker         []DockerTest         `yaml:"docker"`
	Filesystems    []FilesystemTest     `yaml:"filesystems"`
	Ping           []PingTest           `yaml:"ping"`
	DNS            []DNSTest            `yaml:"dns"`
	SystemInfo     []SystemInfoTest     `yaml:"systeminfo"`
	HTTP           []HTTPTest           `yaml:"http"`
	Ports          []PortTest           `yaml:"ports"`
}

// PackageTest represents a package installation test
type PackageTest struct {
	Name     string   `yaml:"name"`
	Packages []string `yaml:"packages"`
	State    string   `yaml:"state"` // present, absent
	Version  string   `yaml:"version,omitempty"`
}

// FileTest represents a file/directory test
type FileTest struct {
	Name      string `yaml:"name"`
	Path      string `yaml:"path"`
	Type      string `yaml:"type"` // file, directory
	Owner     string `yaml:"owner,omitempty"`
	Group     string `yaml:"group,omitempty"`
	Mode      string `yaml:"mode,omitempty"`
	Recursive bool   `yaml:"recursive,omitempty"`
}

// ServiceTest represents a service status test
type ServiceTest struct {
	Name     string   `yaml:"name"`
	Service  string   `yaml:"service,omitempty"`
	Services []string `yaml:"services,omitempty"`
	State    string   `yaml:"state"`   // running, stopped
	Enabled  bool     `yaml:"enabled"` // should be enabled on boot
}

// CommandContentTest represents a command output test
type CommandContentTest struct {
	Name     string   `yaml:"name"`
	Command  string   `yaml:"command"`
	Contains []string `yaml:"contains,omitempty"`
	ExitCode int      `yaml:"exit_code,omitempty"`
}

// UserTest represents a user test
type UserTest struct {
	Name   string   `yaml:"name"`
	User   string   `yaml:"user"`
	Groups []string `yaml:"groups,omitempty"`
	Shell  string   `yaml:"shell,omitempty"`
	Home   string   `yaml:"home,omitempty"`
}

// GroupTest represents a group test
type GroupTest struct {
	Name   string   `yaml:"name"`
	Groups []string `yaml:"groups"`
	State  string   `yaml:"state"` // present, absent
}

// FileContentTest represents a file content test
type FileContentTest struct {
	Name     string   `yaml:"name"`
	Path     string   `yaml:"path"`
	Contains []string `yaml:"contains,omitempty"` // strings that must be present
	Matches  string   `yaml:"matches,omitempty"`  // regex pattern to match
}

// DockerTest represents a Docker container test
type DockerTest struct {
	Name          string   `yaml:"name"`
	Container     string   `yaml:"container,omitempty"`
	Containers    []string `yaml:"containers,omitempty"`
	State         string   `yaml:"state"`          // running, stopped, exists
	Image         string   `yaml:"image,omitempty"`
	RestartPolicy string   `yaml:"restart_policy,omitempty"` // no, always, on-failure, unless-stopped
	Health        string   `yaml:"health,omitempty"`         // healthy, unhealthy, starting, none
}

// FilesystemTest represents a filesystem/mount point test
type FilesystemTest struct {
	Name            string   `yaml:"name"`
	Path            string   `yaml:"path"`
	State           string   `yaml:"state"`                      // mounted, unmounted
	Fstype          string   `yaml:"fstype,omitempty"`           // ext4, xfs, tmpfs, etc.
	Options         []string `yaml:"options,omitempty"`          // rw, ro, noexec, nosuid, etc.
	MinSizeGB       int      `yaml:"min_size_gb,omitempty"`      // minimum size in GB
	MaxUsagePercent int      `yaml:"max_usage_percent,omitempty"` // maximum usage percentage
}

// PingTest represents a network reachability test
type PingTest struct {
	Name string `yaml:"name"`
	Host string `yaml:"host"`
}

// DNSTest represents a DNS resolution test
type DNSTest struct {
	Name string `yaml:"name"`
	Host string `yaml:"host"`
}

// SystemInfoTest represents a system information validation test
type SystemInfoTest struct {
	Name           string `yaml:"name"`
	OS             string `yaml:"os,omitempty"`              // operating system name
	OSVersion      string `yaml:"os_version,omitempty"`      // OS version
	Arch           string `yaml:"arch,omitempty"`            // architecture (x86_64, aarch64, etc.)
	KernelVersion  string `yaml:"kernel_version,omitempty"`  // kernel version
	Hostname       string `yaml:"hostname,omitempty"`        // short hostname
	FQDN           string `yaml:"fqdn,omitempty"`            // fully qualified domain name
	VersionMatch   string `yaml:"version_match,omitempty"`   // "exact" or "prefix" (default: exact)
}

// HTTPTest represents an HTTP endpoint test
type HTTPTest struct {
	Name            string   `yaml:"name"`
	URL             string   `yaml:"url"`
	StatusCode      int      `yaml:"status_code,omitempty"`      // expected status code (default: 200)
	Contains        []string `yaml:"contains,omitempty"`         // strings that must be in response body
	Method          string   `yaml:"method,omitempty"`           // HTTP method (default: GET)
	Insecure        bool     `yaml:"insecure,omitempty"`         // skip TLS verification (default: false)
	FollowRedirects bool     `yaml:"follow_redirects,omitempty"` // follow HTTP redirects (default: false)
}

// PortTest represents a port/socket listening test
type PortTest struct {
	Name     string `yaml:"name"`
	Port     int    `yaml:"port"`
	Protocol string `yaml:"protocol,omitempty"` // tcp or udp (default: tcp)
	State    string `yaml:"state,omitempty"`    // listening or closed (default: listening)
}

// ParseSpec parses a YAML spec file
func ParseSpec(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec file: %w", err)
	}

	var spec Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := spec.Validate(); err != nil {
		return nil, fmt.Errorf("spec validation failed: %w", err)
	}

	return &spec, nil
}

// Validate validates the spec
func (s *Spec) Validate() error {
	if s.Version == "" {
		s.Version = "1.0"
	}

	// Validate package tests
	for i := range s.Tests.Packages {
		pt := &s.Tests.Packages[i]
		if pt.Name == "" {
			return fmt.Errorf("package test %d: name is required", i)
		}
		if len(pt.Packages) == 0 {
			return fmt.Errorf("package test '%s': at least one package is required", pt.Name)
		}
		if pt.State == "" {
			pt.State = "present"
		}
		if pt.State != "present" && pt.State != "absent" {
			return fmt.Errorf("package test '%s': state must be 'present' or 'absent'", pt.Name)
		}
	}

	// Validate file tests
	for i := range s.Tests.Files {
		ft := &s.Tests.Files[i]
		if ft.Name == "" {
			return fmt.Errorf("file test %d: name is required", i)
		}
		if ft.Path == "" {
			return fmt.Errorf("file test '%s': path is required", ft.Name)
		}
		if ft.Type == "" {
			ft.Type = "file"
		}
		if ft.Type != "file" && ft.Type != "directory" {
			return fmt.Errorf("file test '%s': type must be 'file' or 'directory'", ft.Name)
		}
	}

	// Validate service tests
	for i, st := range s.Tests.Services {
		if st.Name == "" {
			return fmt.Errorf("service test %d: name is required", i)
		}
		if st.Service == "" && len(st.Services) == 0 {
			return fmt.Errorf("service test '%s': service or services is required", st.Name)
		}
		if st.State != "running" && st.State != "stopped" {
			return fmt.Errorf("service test '%s': state must be 'running' or 'stopped'", st.Name)
		}
	}

	// Validate user tests
	for i, ut := range s.Tests.Users {
		if ut.Name == "" {
			return fmt.Errorf("user test %d: name is required", i)
		}
		if ut.User == "" {
			return fmt.Errorf("user test '%s': user is required", ut.Name)
		}
	}

	// Validate group tests
	for i := range s.Tests.Groups {
		gt := &s.Tests.Groups[i]
		if gt.Name == "" {
			return fmt.Errorf("group test %d: name is required", i)
		}
		if len(gt.Groups) == 0 {
			return fmt.Errorf("group test '%s': at least one group is required", gt.Name)
		}
		if gt.State == "" {
			gt.State = "present"
		}
		if gt.State != "present" && gt.State != "absent" {
			return fmt.Errorf("group test '%s': state must be 'present' or 'absent'", gt.Name)
		}
	}

	// Validate file content tests
	for i, fct := range s.Tests.FileContent {
		if fct.Name == "" {
			return fmt.Errorf("file_content test %d: name is required", i)
		}
		if fct.Path == "" {
			return fmt.Errorf("file_content test '%s': path is required", fct.Name)
		}
		if len(fct.Contains) == 0 && fct.Matches == "" {
			return fmt.Errorf("file_content test '%s': either contains or matches is required", fct.Name)
		}
	}

	// Validate command content tests
	for i, ct := range s.Tests.CommandContent {
		if ct.Name == "" {
			return fmt.Errorf("command_content test %d: name is required", i)
		}
		if ct.Command == "" {
			return fmt.Errorf("command_content test '%s': command is required", ct.Name)
		}
		if len(ct.Contains) == 0 && ct.ExitCode == 0 {
			return fmt.Errorf("command_content test '%s': either contains or exit_code is required", ct.Name)
		}
	}

	// Validate docker tests
	for i := range s.Tests.Docker {
		dt := &s.Tests.Docker[i]
		if dt.Name == "" {
			return fmt.Errorf("docker test %d: name is required", i)
		}
		if dt.Container == "" && len(dt.Containers) == 0 {
			return fmt.Errorf("docker test '%s': container or containers is required", dt.Name)
		}
		if dt.Container != "" && len(dt.Containers) > 0 {
			return fmt.Errorf("docker test '%s': cannot specify both container and containers", dt.Name)
		}
		if dt.State == "" {
			dt.State = "running"
		}
		if dt.State != "running" && dt.State != "stopped" && dt.State != "exists" {
			return fmt.Errorf("docker test '%s': state must be 'running', 'stopped', or 'exists'", dt.Name)
		}
		// Validate restart policy if specified
		if dt.RestartPolicy != "" {
			validPolicies := map[string]bool{"no": true, "always": true, "on-failure": true, "unless-stopped": true}
			if !validPolicies[dt.RestartPolicy] {
				return fmt.Errorf("docker test '%s': restart_policy must be 'no', 'always', 'on-failure', or 'unless-stopped'", dt.Name)
			}
		}
		// Validate health if specified
		if dt.Health != "" {
			validHealth := map[string]bool{"healthy": true, "unhealthy": true, "starting": true, "none": true}
			if !validHealth[dt.Health] {
				return fmt.Errorf("docker test '%s': health must be 'healthy', 'unhealthy', 'starting', or 'none'", dt.Name)
			}
		}
	}

	// Validate filesystem tests
	for i := range s.Tests.Filesystems {
		ft := &s.Tests.Filesystems[i]
		if ft.Name == "" {
			return fmt.Errorf("filesystem test %d: name is required", i)
		}
		if ft.Path == "" {
			return fmt.Errorf("filesystem test '%s': path is required", ft.Name)
		}
		if ft.State == "" {
			ft.State = "mounted"
		}
		if ft.State != "mounted" && ft.State != "unmounted" {
			return fmt.Errorf("filesystem test '%s': state must be 'mounted' or 'unmounted'", ft.Name)
		}
		if ft.MaxUsagePercent < 0 || ft.MaxUsagePercent > 100 {
			return fmt.Errorf("filesystem test '%s': max_usage_percent must be between 0 and 100", ft.Name)
		}
		if ft.MinSizeGB < 0 {
			return fmt.Errorf("filesystem test '%s': min_size_gb must be >= 0", ft.Name)
		}
	}

	// Validate ping tests
	for i, pt := range s.Tests.Ping {
		if pt.Name == "" {
			return fmt.Errorf("ping test %d: name is required", i)
		}
		if pt.Host == "" {
			return fmt.Errorf("ping test '%s': host is required", pt.Name)
		}
	}

	// Validate DNS tests
	for i, dt := range s.Tests.DNS {
		if dt.Name == "" {
			return fmt.Errorf("dns test %d: name is required", i)
		}
		if dt.Host == "" {
			return fmt.Errorf("dns test '%s': host is required", dt.Name)
		}
	}

	// Validate systeminfo tests
	for i := range s.Tests.SystemInfo {
		st := &s.Tests.SystemInfo[i]
		if st.Name == "" {
			return fmt.Errorf("systeminfo test %d: name is required", i)
		}
		// Set default version_match to "exact"
		if st.VersionMatch == "" {
			st.VersionMatch = "exact"
		}
		// Validate version_match value
		if st.VersionMatch != "exact" && st.VersionMatch != "prefix" {
			return fmt.Errorf("systeminfo test '%s': version_match must be 'exact' or 'prefix'", st.Name)
		}
	}

	// Validate HTTP tests
	for i := range s.Tests.HTTP {
		ht := &s.Tests.HTTP[i]
		if ht.Name == "" {
			return fmt.Errorf("http test %d: name is required", i)
		}
		if ht.URL == "" {
			return fmt.Errorf("http test '%s': url is required", ht.Name)
		}
		// Set default status code to 200
		if ht.StatusCode == 0 {
			ht.StatusCode = 200
		}
		// Set default method to GET
		if ht.Method == "" {
			ht.Method = "GET"
		}
		// Validate method
		validMethods := map[string]bool{"GET": true, "POST": true, "PUT": true, "DELETE": true, "PATCH": true, "HEAD": true, "OPTIONS": true}
		if !validMethods[ht.Method] {
			return fmt.Errorf("http test '%s': method must be one of GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS", ht.Name)
		}
	}

	// Validate port tests
	for i := range s.Tests.Ports {
		pt := &s.Tests.Ports[i]
		if pt.Name == "" {
			return fmt.Errorf("port test %d: name is required", i)
		}
		if pt.Port <= 0 || pt.Port > 65535 {
			return fmt.Errorf("port test '%s': port must be between 1 and 65535", pt.Name)
		}
		// Set default protocol to tcp
		if pt.Protocol == "" {
			pt.Protocol = "tcp"
		}
		// Validate protocol
		if pt.Protocol != "tcp" && pt.Protocol != "udp" {
			return fmt.Errorf("port test '%s': protocol must be 'tcp' or 'udp'", pt.Name)
		}
		// Set default state to listening
		if pt.State == "" {
			pt.State = "listening"
		}
		// Validate state
		if pt.State != "listening" && pt.State != "closed" {
			return fmt.Errorf("port test '%s': state must be 'listening' or 'closed'", pt.State)
		}
	}

	return nil
}
