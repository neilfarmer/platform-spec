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

	return nil
}
