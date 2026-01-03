package system

import (
	"context"

	"github.com/neilfarmer/platform-spec/pkg/core"
)

// SystemPlugin handles all system-level tests (packages, files, services, users, groups, etc.)
type SystemPlugin struct{}

// NewSystemPlugin creates a new system plugin
func NewSystemPlugin() *SystemPlugin {
	return &SystemPlugin{}
}

// Execute runs all system-level tests
func (p *SystemPlugin) Execute(ctx context.Context, spec *core.Spec, provider core.Provider, failFast bool, callback core.ResultCallback) ([]core.Result, bool) {
	var results []core.Result
	shouldStop := false

	// Execute package tests
	for _, test := range spec.Tests.Packages {
		result := executePackageTest(ctx, provider, test)
		results = append(results, result)
		if callback != nil {
			callback(result)
		}
		if failFast && result.Status == core.StatusFail {
			return results, true
		}
	}

	// Execute file tests
	for _, test := range spec.Tests.Files {
		result := executeFileTest(ctx, provider, test)
		results = append(results, result)
		if callback != nil {
			callback(result)
		}
		if failFast && result.Status == core.StatusFail {
			return results, true
		}
	}

	// Execute service tests
	for _, test := range spec.Tests.Services {
		result := executeServiceTest(ctx, provider, test)
		results = append(results, result)
		if callback != nil {
			callback(result)
		}
		if failFast && result.Status == core.StatusFail {
			return results, true
		}
	}

	// Execute user tests
	for _, test := range spec.Tests.Users {
		result := executeUserTest(ctx, provider, test)
		results = append(results, result)
		if callback != nil {
			callback(result)
		}
		if failFast && result.Status == core.StatusFail {
			return results, true
		}
	}

	// Execute group tests
	for _, test := range spec.Tests.Groups {
		result := executeGroupTest(ctx, provider, test)
		results = append(results, result)
		if callback != nil {
			callback(result)
		}
		if failFast && result.Status == core.StatusFail {
			return results, true
		}
	}

	// Execute file content tests
	for _, test := range spec.Tests.FileContent {
		result := executeFileContentTest(ctx, provider, test)
		results = append(results, result)
		if callback != nil {
			callback(result)
		}
		if failFast && result.Status == core.StatusFail {
			return results, true
		}
	}

	// Execute command content tests
	for _, test := range spec.Tests.CommandContent {
		result := executeCommandContentTest(ctx, provider, test)
		results = append(results, result)
		if callback != nil {
			callback(result)
		}
		if failFast && result.Status == core.StatusFail {
			return results, true
		}
	}

	// Execute Docker tests
	for _, test := range spec.Tests.Docker {
		result := executeDockerTest(ctx, provider, test)
		results = append(results, result)
		if callback != nil {
			callback(result)
		}
		if failFast && result.Status == core.StatusFail {
			return results, true
		}
	}

	// Execute filesystem tests
	for _, test := range spec.Tests.Filesystems {
		result := executeFilesystemTest(ctx, provider, test)
		results = append(results, result)
		if callback != nil {
			callback(result)
		}
		if failFast && result.Status == core.StatusFail {
			return results, true
		}
	}

	// Execute ping tests
	for _, test := range spec.Tests.Ping {
		result := executePingTest(ctx, provider, test)
		results = append(results, result)
		if callback != nil {
			callback(result)
		}
		if failFast && result.Status == core.StatusFail {
			return results, true
		}
	}

	// Execute DNS tests
	for _, test := range spec.Tests.DNS {
		result := executeDNSTest(ctx, provider, test)
		results = append(results, result)
		if callback != nil {
			callback(result)
		}
		if failFast && result.Status == core.StatusFail {
			return results, true
		}
	}

	// Execute system info tests
	for _, test := range spec.Tests.SystemInfo {
		result := executeSystemInfoTest(ctx, provider, test)
		results = append(results, result)
		if callback != nil {
			callback(result)
		}
		if failFast && result.Status == core.StatusFail {
			return results, true
		}
	}

	// Execute HTTP tests
	for _, test := range spec.Tests.HTTP {
		result := executeHTTPTest(ctx, provider, test)
		results = append(results, result)
		if callback != nil {
			callback(result)
		}
		if failFast && result.Status == core.StatusFail {
			return results, true
		}
	}

	// Execute port tests
	for _, test := range spec.Tests.Ports {
		result := executePortTest(ctx, provider, test)
		results = append(results, result)
		if callback != nil {
			callback(result)
		}
		if failFast && result.Status == core.StatusFail {
			return results, true
		}
	}

	return results, shouldStop
}
