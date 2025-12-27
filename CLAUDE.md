# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**platform-spec** is a pluggable infrastructure testing framework that validates system state across multiple platforms using declarative YAML specifications. It follows a provider pattern where different infrastructure backends (Remote, Local, Kubernetes, AWS, OpenStack) can execute tests defined in a single YAML spec format.

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
go test -v ./pkg/providers/remote

# Build and run locally
./dist/platform-spec test remote ubuntu@host spec.yaml
```

### Testing the Binary
```bash
# Build first
make build

# Test with example spec
./dist/platform-spec test remote user@hostname examples/basic.yaml

# Use flags for authentication and connection
./dist/platform-spec test remote user@host spec.yaml -i ~/.ssh/key -p 2222 --verbose
```

## Architecture

### Core Components

The codebase follows a **plugin-based architecture** with clear separation between framework, providers, plugins, and CLI:

**1. Provider Interface (pkg/core/executor.go)**
The central abstraction is the `Provider` interface with a single method:
```go
type Provider interface {
    ExecuteCommand(ctx context.Context, command string) (stdout, stderr string, exitCode int, err error)
}
```

All infrastructure providers (Remote, Local, Kubernetes) implement this interface. Providers handle **how** to execute commands (locally, via SSH to remote systems, via kubectl).

**2. Plugin Interface (pkg/core/executor.go)**
The plugin system defines **what** tests to execute:
```go
type Plugin interface {
    Execute(ctx context.Context, spec *Spec, provider Provider, failFast bool) ([]Result, bool)
}
```

Two plugins are currently implemented:
- **SystemPlugin** (pkg/core/system/): OS-level tests that work on any system (local or remote)
- **KubernetesPlugin** (pkg/core/kubernetes/): Kubernetes-specific tests using kubectl

The `Executor` orchestrates test execution by iterating through registered plugins, passing each the spec and provider.

**3. Spec Definition (pkg/core/spec.go)**
Defines 14 test types organized into two categories:

**System Tests** (handled by SystemPlugin):
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
- `PortTest` - Port/socket listening state validation (TCP/UDP)

**Kubernetes Tests** (handled by KubernetesPlugin):
- `KubernetesNamespaceTest` - Namespace existence and labels
- `KubernetesPodTest` - Pod state, readiness, images, and labels
- `KubernetesDeploymentTest` - Deployment availability, replicas, and images
- `KubernetesServiceTest` - Service type, ports, and selectors
- `KubernetesConfigMapTest` - ConfigMap existence and keys

Each test type includes YAML validation with defaults (e.g., `state: present`, `type: file`, `state: running` for docker, `state: mounted` for filesystems, `version_match: exact` for systeminfo, `status_code: 200` and `method: GET` for HTTP, `protocol: tcp` and `state: listening` for ports).

**4. Test Execution Flow (pkg/core/executor.go)**
The executor uses a plugin-based execution model:
1. Executor iterates through registered plugins (System, Kubernetes)
2. Each plugin executes its tests in a defined order
3. Each test produces a `Result` with `Status` (passed/failed/skipped/error), message, duration, and details map
4. The `FailFast` config option stops execution on first failure (checked after each test, not after each plugin)

**5. Providers**

Three providers are currently implemented:

**Local Provider (pkg/providers/local/provider.go)**
- Executes commands on the local system using `os/exec`
- Simple implementation with no connection overhead
- Works with SystemPlugin for local OS testing
- Usage: `platform-spec test local spec.yaml`

**Remote Provider (pkg/providers/remote/provider.go)**
- Connects to remote systems via SSH protocol
- Supports SSH key files (`-i` flag) and SSH Agent (`SSH_AUTH_SOCK` env var)
- Uses golang.org/x/crypto/ssh for connections
- `ParseTarget()` helper parses "user@host" format
- Works with SystemPlugin for remote OS testing
- Usage: `platform-spec test remote user@host spec.yaml`

**Kubernetes Provider (pkg/providers/kubernetes/provider.go)**
- Executes kubectl commands against a Kubernetes cluster
- Supports kubeconfig files, context selection, and namespace override
- Works with both SystemPlugin (for kubectl exec) and KubernetesPlugin
- Usage: `platform-spec test kubernetes spec.yaml --kubeconfig ~/.kube/config`

**6. Output Formatters (pkg/output/human.go)**
Human-readable format with ASCII symbols (✓/✗/○/⚠), duration tracking, and summary counts.

### Package Organization

```
cmd/platform-spec/     # CLI layer (Cobra commands)
├── main.go           # Entry point
├── root.go           # Root command setup
├── test.go           # Subcommands: local, remote, kubernetes
└── version.go        # Version info

pkg/core/             # Core framework
├── executor.go       # Executor, Provider interface, Plugin interface
├── spec.go           # YAML parsing and test type definitions
├── types.go          # Status, Result, TestResults structs
├── mock_provider.go  # Mock provider for testing
├── system/           # System plugin (OS-level tests)
│   ├── plugin.go     # SystemPlugin implementation
│   ├── file.go       # File tests
│   ├── package.go    # Package tests
│   ├── service.go    # Service tests
│   ├── user.go       # User tests
│   ├── group.go      # Group tests
│   ├── docker.go     # Docker tests
│   ├── filesystem.go # Filesystem tests
│   ├── ping.go       # Ping tests
│   ├── dns.go        # DNS tests
│   ├── http.go       # HTTP tests
│   ├── port.go       # Port tests
│   ├── systeminfo.go # System info tests
│   ├── file_content.go     # File content tests
│   ├── command_content.go  # Command content tests
│   └── *_test.go     # Tests for each module
└── kubernetes/       # Kubernetes plugin
    ├── plugin.go     # KubernetesPlugin implementation
    ├── kubernetes.go # All Kubernetes test implementations
    └── kubernetes_test.go

pkg/providers/        # Infrastructure providers
├── local/
│   └── provider.go   # Local execution (os/exec)
├── remote/
│   └── provider.go   # Remote execution via SSH
└── kubernetes/
    └── provider.go   # Kubectl execution

pkg/output/           # Output formatters
└── human.go          # Human-readable output (JSON/JUnit planned)
```

## Key Implementation Patterns

### Adding a New Test Type

**For System Tests** (OS-level tests):
1. Define struct in `pkg/core/spec.go` with YAML tags
2. Add field to `Tests` struct in `pkg/core/spec.go`
3. Add validation logic in `Spec.Validate()` method
4. Create new file `pkg/core/system/<testtype>.go` with `execute<TestType>Test()` function
   - Function signature: `func executeXxxTest(ctx context.Context, provider core.Provider, test core.XxxTest) core.Result`
5. Add the test execution to `SystemPlugin.Execute()` in `pkg/core/system/plugin.go`
6. Create comprehensive tests in `pkg/core/system/<testtype>_test.go`
7. Update documentation as needed

**For Kubernetes Tests**:
1. Define struct in `pkg/core/spec.go` with YAML tags (usually under `KubernetesTests`)
2. Add field to `KubernetesTests` struct in `pkg/core/spec.go`
3. Add validation logic in `Spec.Validate()` method
4. Add `executeKubernetes<TestType>Test()` function in `pkg/core/kubernetes/kubernetes.go`
5. Add the test execution to `KubernetesPlugin.Execute()` in `pkg/core/kubernetes/plugin.go`
6. Add test cases in `pkg/core/kubernetes/kubernetes_test.go`

### Adding a New Plugin

1. Create new directory under `pkg/core/<plugin-name>/`
2. Create `plugin.go` implementing the `Plugin` interface:
   ```go
   type MyPlugin struct{}

   func NewMyPlugin() *MyPlugin {
       return &MyPlugin{}
   }

   func (p *MyPlugin) Execute(ctx context.Context, spec *core.Spec, provider core.Provider, failFast bool) ([]core.Result, bool) {
       var results []core.Result
       // Execute tests, checking failFast after each test
       return results, shouldStop
   }
   ```
3. Implement test execution functions in separate files
4. Register plugin in `cmd/platform-spec/test.go` when creating executors
5. Add comprehensive tests

### Adding a New Provider

1. Create new package under `pkg/providers/<name>/`
2. Implement `Provider` interface with `ExecuteCommand()`
3. Add subcommand in `cmd/platform-spec/test.go`
4. Register appropriate plugins when creating the executor
5. Test against both SystemPlugin and KubernetesPlugin test types

### Remote Command Patterns

The SystemPlugin uses standard Linux commands via the provider:
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
- HTTP: `curl -s -w "\n%{http_code}"` for HTTP requests with status code extraction, supports `-X METHOD`, `-k` for insecure TLS, and `-L` for following redirects
- Ports: `ss -tln | grep -E ':PORT\s'` for TCP, `ss -uln | grep -E ':PORT\s'` for UDP socket listening checks

All commands include `2>/dev/null` for error suppression and fallback checks.

The KubernetesPlugin uses kubectl commands via the provider:
- All resources: `kubectl get <resource> <name> -n <namespace> -o json`
- Namespace info: `kubectl get namespace <name> -o json`
- Pod info: `kubectl get pod <name> -n <namespace> -o json`
- Deployment info: `kubectl get deployment <name> -n <namespace> -o json`
- Service info: `kubectl get service <name> -n <namespace> -o json`
- ConfigMap info: `kubectl get configmap <name> -n <namespace> -o json`

All kubectl commands return JSON output which is parsed using helper functions (`getNestedString`, `getNestedMap`, etc.).

## Development Workflow

### Phase Roadmap
The project follows a phased development approach (see README.md):
- **Phase 1** (Complete): Core framework, Remote and Local providers, package/file tests
- **Phase 2** (Complete): All system tests, plugin architecture, Kubernetes provider
- **Phase 3** (Planned): JSON/JUnit output, parallel execution, variable substitution
- **Phase 4** (Planned): AWS and OpenStack providers

### Plugin Architecture
The codebase uses a clean plugin pattern:
- **Separation of Concerns**: Plugins define WHAT to test, Providers define HOW to execute
- **Extensibility**: New plugins can be added without modifying core executor logic
- **Terminology**: "System" refers to OS-level testing (works on local and remote systems)
- **Test Organization**: Each plugin has its own package with dedicated test files
- **Fail-Fast**: Checked after each individual test, not after each plugin

### Testing Strategy
- Unit tests for each plugin (pkg/core/system/*_test.go, pkg/core/kubernetes/*_test.go)
- Unit tests for each provider (pkg/providers/*/*_test.go)
- Core framework tests use external test package (core_test) to avoid import cycles
- All tests use MockProvider for isolated testing
- Tests use `-count=1` to disable caching for fresh runs
- 209 total tests covering all functionality
- No integration tests with live connections yet (opportunity for improvement)

### Cross-Platform Builds
The project targets:
- macOS: `darwin/arm64` (Apple Silicon)
- Linux: `linux/amd64`

GoReleaser handles releases with automatic changelog generation (excludes docs/test/chore commits).

## Important Notes

### Security Considerations
- No secrets should be in YAML specs - use SSH agent or key files for authentication
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
