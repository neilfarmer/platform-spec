# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**platform-spec** is a pluggable infrastructure testing framework that validates system state across multiple platforms using declarative YAML specifications. It follows a provider pattern where different infrastructure backends (SSH, AWS, OpenStack) can execute tests defined in a single YAML spec format.

## Commandments

When working in this codebase, follow these rules:

1. **Ask First**: Always prompt the user for clarifying questions before making assumptions about implementation details or requirements.

2. **Test Everything**: When code is added, changed, or removed, corresponding tests must be added, changed, or removed. No code changes without test changes.

3. **Light Documentation**: Keep documentation light and straightforward. Focus on clarity and brevity over comprehensiveness.

## Essential Commands

### Development
```bash
make build          # Build to dist/platform-spec
make test           # Run all tests with -v -count=1
make install        # Build and install to $GOPATH/bin
make clean          # Remove dist/ directory
make release-build  # Cross-compile for darwin-arm64 and linux-amd64
```

### Running Tests
```bash
# Run all tests
go test -v -count=1 ./...

# Run tests for a specific package
go test -v ./pkg/core
go test -v ./pkg/providers/ssh

# Build and run locally
./dist/platform-spec test ssh ubuntu@host spec.yaml
```

### Testing the Binary
```bash
# Build first
make build

# Test with example spec
./dist/platform-spec test ssh user@hostname examples/basic.yaml

# Use flags for authentication and connection
./dist/platform-spec test ssh user@host spec.yaml -i ~/.ssh/key -p 2222 --verbose
```

## Architecture

### Core Components

The codebase follows a layered architecture with clear separation between framework, providers, and CLI:

**1. Provider Interface (pkg/core/executor.go)**
The central abstraction is the `Provider` interface with a single method:
```go
type Provider interface {
    ExecuteCommand(ctx context.Context, command string) (stdout, stderr string, exitCode int, err error)
}
```

All infrastructure providers (SSH, AWS, OpenStack) implement this interface. The `Executor` coordinates test execution by calling provider methods.

**2. Spec Definition (pkg/core/spec.go)**
Defines 13 test types that work across all providers:
- `PackageTest` - Package installation state (dpkg/rpm/apk detection)
- `FileTest` - File/directory properties (ownership, permissions, type)
- `ServiceTest` - Service status (running/stopped, enabled/disabled via systemd)
- `UserTest` - User properties (shell, home, group membership)
- `GroupTest` - Group existence
- `FileContentTest` - File content matching (strings or regex)
- `CommandContentTest` - Command execution and output validation
- `DockerTest` - Docker container status and properties (image, restart policy, health)
- `FilesystemTest` - Filesystem mount status, type, options, and disk usage
- `PingTest` - Network reachability using ICMP ping
- `DNSTest` - DNS resolution validation
- `SystemInfoTest` - System information validation (OS, architecture, kernel, hostname)
- `HTTPTest` - HTTP endpoint testing (status code, response content validation)

Each test type includes YAML validation with defaults (e.g., `state: present`, `type: file`, `state: running` for docker, `state: mounted` for filesystems, `version_match: exact` for systeminfo, `status_code: 200` and `method: GET` for HTTP).

**3. Test Execution Flow (pkg/core/executor.go)**
The executor runs tests sequentially in this order:
1. Packages → 2. Files → 3. Services → 4. Users → 5. Groups → 6. File Content → 7. Commands → 8. Docker → 9. Filesystems → 10. Ping → 11. DNS → 12. SystemInfo → 13. HTTP

Each test produces a `Result` with `Status` (passed/failed/skipped/error), message, duration, and details map. The `FailFast` config option stops execution on first failure.

**4. Providers**

Two providers are currently implemented:

**Local Provider (pkg/providers/local/provider.go)**
- Executes commands on the local system using `os/exec`
- Simple implementation with no connection overhead
- Supports all 13 test types
- Usage: `platform-spec test local spec.yaml`

**SSH Provider (pkg/providers/ssh/provider.go)**
- Supports SSH key files (`-i` flag) and SSH Agent (`SSH_AUTH_SOCK` env var)
- Uses golang.org/x/crypto/ssh for connections
- `ParseTarget()` helper parses "user@host" format
- **Security Note**: Currently uses `InsecureIgnoreHostKey()` - marked as TODO for production use
- Usage: `platform-spec test ssh user@host spec.yaml`

**5. Output Formatters (pkg/output/human.go)**
Human-readable format with ASCII symbols (✓/✗/○/⚠), duration tracking, and summary counts.

### Package Organization

```
cmd/platform-spec/     # CLI layer (Cobra commands)
├── main.go           # Entry point
├── root.go           # Root command setup
├── test.go           # Subcommands: ssh, aws, openstack
└── version.go        # Version info

pkg/core/             # Core framework
├── types.go          # Status, Result, TestResults structs
├── spec.go           # YAML parsing and 7 test type definitions
└── executor.go       # Executor + all test execution logic

pkg/providers/        # Infrastructure providers
└── ssh/
    └── provider.go   # SSH implementation of Provider interface

pkg/assertions/       # Legacy (being phased out - logic moved to executor.go)
├── package.go
└── file.go

pkg/output/           # Output formatters
└── human.go          # Human-readable output (JSON/JUnit planned)
```

## Key Implementation Patterns

### Adding a New Test Type

1. Define struct in `pkg/core/spec.go` with YAML tags
2. Add field to `Tests` struct in `pkg/core/spec.go`
3. Add validation logic in `Spec.Validate()` method
4. Implement `execute<TestType>Test()` method in `pkg/core/executor.go`
5. Call the method from `Executor.Execute()` with fail-fast check
6. Update documentation in `docs/ssh/assertions/`

### Adding a New Provider

1. Create new package under `pkg/providers/<name>/`
2. Implement `Provider` interface with `ExecuteCommand()`
3. Add subcommand in `cmd/platform-spec/test.go`
4. Test against all 13 assertion types in the spec

### Remote Command Patterns

The executor uses standard Linux commands via the provider:
- Package detection: `dpkg -l`, `rpm -q`, `apk info -e`
- File info: `stat -c '%F:%U:%G:%a'`
- Service status: `systemctl is-active`, `systemctl is-enabled`
- User info: `id -u`, `id -g`, `getent passwd`
- Groups: `id -Gn`, `getent group`
- File content: `grep -F` (fixed strings), `grep -E` (regex)
- Command exec: Direct execution with stdout/stderr capture
- Docker: `docker inspect --format` with template for status, image, restart policy, health
- Filesystems: `findmnt --noheadings --output TARGET,FSTYPE,OPTIONS,SIZE,USED,USE%` for mount info, `df -BG --output=size` for disk size
- Ping: `ping -c 1 -W 5` for single ICMP packet with 5 second timeout
- DNS: `dig +short` or fallback to `getent hosts` for hostname resolution
- SystemInfo: `/etc/os-release` for OS info, `uname -m` for architecture, `uname -r` for kernel, `hostname -s/-f` for hostname/FQDN
- HTTP: `curl -s -w "\n%{http_code}"` for HTTP requests with status code extraction, supports `-X METHOD` and `-k` for insecure TLS

All commands include `2>/dev/null` for error suppression and fallback checks.

## Development Workflow

### Phase Roadmap
The project follows a phased development approach (see README.md):
- **Phase 1** (Complete): Core framework, SSH provider, package/file assertions
- **Phase 2** (In Progress): Service, user, group, file content, command assertions
- **Phase 3** (Planned): JSON/JUnit output, parallel execution, variable substitution
- **Phase 4** (Planned): AWS and OpenStack providers

### Testing Strategy
- Unit tests for each package (pkg/core/*_test.go, pkg/providers/ssh/*_test.go)
- Tests use `-count=1` to disable caching for fresh runs
- No integration tests with live SSH connections yet (opportunity for improvement)

### Cross-Platform Builds
The project targets:
- macOS: `darwin/arm64` (Apple Silicon)
- Linux: `linux/amd64`

GoReleaser handles releases with automatic changelog generation (excludes docs/test/chore commits).

## Important Notes

### Security Considerations
- SSH host key verification is currently disabled (`InsecureIgnoreHostKey`) - this is a known TODO
- No secrets should be in YAML specs - use SSH agent for authentication
- Command execution uses `shellQuote()` to escape single quotes in grep patterns

### Context and Timeout
All provider commands accept `context.Context` for timeout and cancellation support. The global timeout defaults to 300 seconds (configurable in spec YAML).

### Parallel Execution
Currently marked as TODO in spec config. All tests run sequentially. When implementing:
- Use goroutines with WaitGroup for each test type
- Protect `TestResults.Results` slice with mutex
- Respect `FailFast` flag with context cancellation

### Variable Substitution
The `variables` section in YAML spec is parsed but not used. Planned for Phase 3 to allow templating in test definitions.
