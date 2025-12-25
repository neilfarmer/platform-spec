# platform-spec

A pluggable infrastructure testing framework that validates system state across multiple platforms using declarative YAML specifications.

## Architecture

platform-spec uses a **plugin-based architecture** that separates test execution from command delivery:

- **Plugins** define WHAT to test (System tests, Kubernetes tests)
- **Providers** define HOW to execute commands (Remote, Local, Kubernetes)
- **Executor** coordinates plugins and providers

This design allows system-level tests (files, packages, services, etc.) to work seamlessly on both local and remote systems, while specialized plugins handle platform-specific resources like Kubernetes.

## Installation

**macOS (ARM64)**

```bash
curl -L https://github.com/neilfarmer/platform-spec/releases/download/v0.0.1/platform-spec_0.0.1_darwin_arm64.zip -o platform-spec.zip
unzip platform-spec.zip
sudo mv platform-spec /usr/local/bin/platform-spec
rm platform-spec.zip
```

**Linux (AMD64)**

```bash
curl -L https://github.com/neilfarmer/platform-spec/releases/download/v0.0.1/platform-spec_0.0.1_linux_amd64.tar.gz | tar xz
sudo mv platform-spec /usr/local/bin/platform-spec
```

See [releases page](https://github.com/neilfarmer/platform-spec/releases) for other versions.

## Quick Start

Create a spec file `mytest.yaml`:

```yaml
version: "1.0"
tests:
  packages:
    - name: "Docker installed"
      packages: [docker-ce]
      state: present
  files:
    - name: "Config directory exists"
      path: /etc/myapp
      type: directory
```

Run the tests:

```bash
# Test local system
platform-spec test local mytest.yaml

# Test remote system via SSH
platform-spec test remote ubuntu@myhost mytest.yaml
```

See [USAGE.md](USAGE.md) for complete documentation.

## Roadmap

### Phase 1: Core Framework ✅

- Remote and Local providers
- Package and file assertions
- Human-readable output

### Phase 2: Plugin Architecture ✅

- **System Plugin**: 14 test types for OS-level validation
  - Packages, files, services, users, groups
  - Docker containers, filesystems
  - Network (ping, DNS, HTTP, ports)
  - System information
  - File and command content matching
- **Kubernetes Plugin**: 5 test types for K8s resources
  - Namespaces, pods, deployments, services, configmaps
- **Kubernetes Provider**: kubectl-based command execution

### Phase 3: Advanced Features

- JSON and JUnit output formats
- Parallel execution
- Variable substitution

### Phase 4: Cloud Providers

- AWS provider (EC2, S3, IAM, RDS)
- OpenStack provider
