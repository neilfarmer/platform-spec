# Platform-Spec: Enhanced Architecture Plan

## Executive Summary

A pluggable infrastructure testing framework in Go that validates system state across multiple platforms (SSH/Linux, AWS, OpenStack, etc.) using declarative YAML specifications.

## Core Architecture

### 1. Plugin System Architecture

```
platform-spec/
├── cmd/
│   └── platform-spec/        # Main CLI entry point
├── pkg/
│   ├── core/                 # Core abstractions
│   │   ├── executor.go       # Test executor interface
│   │   ├── reporter.go       # Result reporting interface
│   │   └── spec.go           # Spec parsing and validation
│   ├── providers/            # Provider plugins
│   │   ├── provider.go       # Provider interface
│   │   ├── ssh/              # SSH provider implementation
│   │   ├── aws/              # AWS provider (future)
│   │   └── openstack/        # OpenStack provider (future)
│   ├── assertions/           # Assertion engine
│   │   ├── assertion.go      # Assertion interface
│   │   ├── package.go        # Package assertions
│   │   ├── file.go           # File/directory assertions
│   │   ├── service.go        # Service assertions
│   │   ├── user.go           # User/group assertions
│   │   └── command.go        # Custom command assertions
│   └── output/               # Output formatters
│       ├── json.go
│       ├── junit.go
│       └── human.go
```

### 2. Key Design Principles

1. **Provider Abstraction**: Each infrastructure type (SSH, AWS, etc.) implements the `Provider` interface
2. **Assertion Plugins**: Each test type (package, file, service) is an independent assertion module
3. **Composable**: Tests can include other test files and reuse common specs
4. **Fail-Fast or Continue**: Configurable behavior on test failures
5. **Parallel Execution**: Tests can run in parallel where safe
6. **Idempotent**: Tests should be read-only and safe to run multiple times

---

## Enhanced YAML Schema

### Basic Structure

```yaml
version: "1.0"
metadata:
  name: "Web Server Infrastructure Test"
  description: "Validates Docker-based web server setup"
  tags: ["production", "docker", "monitoring"]

config:
  fail_fast: false          # Stop on first failure
  parallel: true            # Run tests in parallel when possible
  timeout: 300              # Global timeout in seconds

variables:                  # Reusable variables
  domain: "example.com"
  monitoring_path: "/opt/monitoring"
  required_packages:
    - docker-ce
    - docker-compose-plugin
    - prometheus-node-exporter

includes:                   # Import other test specs
  - ./common/base-system.yaml
  - ./common/security-baseline.yaml

tests:
  # Test definitions here (see sections below)
```

### Test Categories with Examples

#### 1. Package Tests

```yaml
tests:
  packages:
    - name: "Docker packages installed"
      packages:
        - docker-ce>=5:24.0.0
        - docker-compose-plugin
        - containerd.io
      state: present

    - name: "Unwanted packages removed"
      packages:
        - telnet
        - ftp
      state: absent

    - name: "Python version check"
      packages:
        - python3
      version: ">=3.9"
```

#### 2. File/Directory Tests

```yaml
tests:
  files:
    - name: "Monitoring directory structure"
      path: "{{ .variables.monitoring_path }}"
      type: directory
      owner: root
      group: root
      mode: "0755"

    - name: "Docker compose file exists"
      path: "{{ .variables.monitoring_path }}/docker-compose.yml"
      type: file
      owner: root
      mode: "0644"

    - name: "Sensitive config permissions"
      path: /etc/ssl/private
      type: directory
      mode: "0700"
      recursive: false

  file_content:
    - name: "Grafana image version in compose file"
      path: "{{ .variables.monitoring_path }}/docker-compose.yml"
      matches:
        - pattern: 'grafana/grafana:12\.2\.\d+'
          type: regex
        - pattern: "registry.{{ .variables.domain }}"
          type: contains

    - name: "SSH config hardening"
      path: /etc/ssh/sshd_config
      contains:
        - "PermitRootLogin no"
        - "PasswordAuthentication no"
      not_contains:
        - "PermitEmptyPasswords yes"
```

#### 3. Service Tests

```yaml
tests:
  services:
    - name: "Docker daemon running"
      service: docker
      state: running
      enabled: true

    - name: "Firewall active"
      service: ufw
      state: running

    - name: "Legacy services disabled"
      services:
        - telnet
        - rsh
      state: stopped
      enabled: false
```

#### 4. User/Group Tests

```yaml
tests:
  users:
    - name: "App user exists"
      user: appuser
      groups: [docker, www-data]
      shell: /bin/bash
      home: /home/appuser

    - name: "Root login restrictions"
      user: root
      shell: /bin/bash  # Not /bin/sh or disabled

  groups:
    - name: "Required groups exist"
      groups:
        - docker
        - monitoring
      state: present
```

#### 5. Port/Network Tests

```yaml
tests:
  ports:
    - name: "Web server listening"
      port: 443
      protocol: tcp
      state: listening
      process: nginx

    - name: "Database not exposed"
      port: 5432
      protocol: tcp
      state: not_listening
      interface: "0.0.0.0"  # Should not listen on all interfaces
```

#### 6. Process Tests

```yaml
tests:
  processes:
    - name: "Nginx master process"
      process: nginx
      count: ">=1"
      user: root

    - name: "No zombie processes"
      state: zombie
      count: 0
```

#### 7. Kernel/System Tests

```yaml
tests:
  kernel:
    - name: "Kernel version minimum"
      version: ">=5.15"

  sysctl:
    - name: "IP forwarding enabled"
      key: net.ipv4.ip_forward
      value: "1"

    - name: "Security settings"
      settings:
        net.ipv4.conf.all.rp_filter: "1"
        kernel.randomize_va_space: "2"
```

#### 8. Custom Command Tests

```yaml
tests:
  commands:
    - name: "Docker containers running"
      command: "docker ps --format '{{.Names}}'"
      contains:
        - grafana
        - prometheus

    - name: "Disk space check"
      command: "df -h / | tail -1 | awk '{print $5}' | sed 's/%//'"
      assert:
        type: numeric
        operator: "<"
        value: 80

    - name: "Certificate validity"
      command: "openssl x509 -in /etc/ssl/cert.pem -noout -enddate"
      matches: 'notAfter=.*202[6-9]'  # Valid until at least 2026
```

#### 9. Cloud Provider Tests (Future)

```yaml
# AWS Example
tests:
  aws:
    ec2_instances:
      - name: "Web servers running"
        filters:
          tag:Environment: production
          tag:Role: web
        state: running
        count: ">=2"

    s3_buckets:
      - name: "Backup bucket encryption"
        bucket: my-backups
        encryption: enabled
        versioning: enabled
        public_access: blocked

    iam_policies:
      - name: "No overly permissive policies"
        policy_arn: "arn:aws:iam::123456789:policy/MyPolicy"
        not_contains_actions:
          - "s3:*"
          - "*:*"
```

---

## Provider Interface Design

```go
// pkg/providers/provider.go
package providers

import "context"

type Provider interface {
    // Connect establishes connection to the target
    Connect(ctx context.Context) error

    // Close cleans up resources
    Close() error

    // ExecuteAssertion runs a specific assertion
    ExecuteAssertion(ctx context.Context, assertion Assertion) (*Result, error)

    // GetCapabilities returns what this provider supports
    GetCapabilities() []string
}

type ConnectionConfig struct {
    Type       string                 // "ssh", "aws", "openstack"
    Parameters map[string]interface{} // Provider-specific params
}

type Assertion interface {
    Type() string          // "package", "file", "service", etc.
    Validate() error       // Validate assertion config
    Description() string   // Human-readable description
}

type Result struct {
    Name     string
    Status   Status  // Pass, Fail, Skip, Error
    Message  string
    Duration time.Duration
    Details  map[string]interface{}
}
```

### SSH Provider Implementation

```go
// pkg/providers/ssh/provider.go
package ssh

type SSHProvider struct {
    client *ssh.Client
    config *SSHConfig
}

type SSHConfig struct {
    Host       string
    Port       int
    User       string
    KeyPath    string
    Password   string
    Timeout    time.Duration
}

func (p *SSHProvider) Connect(ctx context.Context) error {
    // SSH connection logic
}

func (p *SSHProvider) ExecuteAssertion(ctx context.Context, assertion Assertion) (*Result, error) {
    switch a := assertion.(type) {
    case *PackageAssertion:
        return p.checkPackage(ctx, a)
    case *FileAssertion:
        return p.checkFile(ctx, a)
    // ... other assertion types
    }
}
```

---

## CLI Design

### Command Structure

```bash
# Basic usage
platform-spec test ssh -i ~/.ssh/key.pem ubuntu@192.168.1.100 spec.yaml

# Multiple specs
platform-spec test ssh -i ~/.ssh/key.pem ubuntu@192.168.1.100 spec1.yaml spec2.yaml

# Custom output format
platform-spec test ssh -i ~/.ssh/key.pem ubuntu@192.168.1.100 spec.yaml -o json > results.json

# Parallel testing across multiple hosts
platform-spec test ssh -i ~/.ssh/key.pem -f hosts.txt spec.yaml --parallel

# AWS testing (future)
platform-spec test aws --region us-east-1 --profile prod aws-spec.yaml

# Dry run mode
platform-spec test ssh ubuntu@host spec.yaml --dry-run

# Verbose output
platform-spec test ssh ubuntu@host spec.yaml -v
```

### Configuration File Support

```yaml
# .platform-spec.yaml
providers:
  ssh:
    default_user: ubuntu
    default_key: ~/.ssh/id_rsa
    timeout: 30

  aws:
    region: us-east-1
    profile: default

output:
  format: human
  verbose: false

reporting:
  junit_output: ./test-results/junit.xml
  json_output: ./test-results/results.json
```

---

## Output Formats

### Human-Readable (Default)

```
Platform-Spec Test Results
==========================

Spec: Web Server Infrastructure Test
Host: ubuntu@192.168.1.100

✓ Docker packages installed (0.5s)
✓ Monitoring directory structure (0.1s)
✓ Docker daemon running (0.2s)
✗ Grafana container running (0.3s)
  Expected: grafana container in 'docker ps' output
  Actual: Container not found

Tests: 15 passed, 1 failed, 0 skipped
Duration: 5.2s
Status: FAILED
```

### JSON Output

```json
{
  "metadata": {
    "name": "Web Server Infrastructure Test",
    "timestamp": "2025-12-23T10:30:00Z",
    "duration": 5.2
  },
  "summary": {
    "total": 16,
    "passed": 15,
    "failed": 1,
    "skipped": 0
  },
  "tests": [
    {
      "name": "Docker packages installed",
      "status": "passed",
      "duration": 0.5,
      "assertion_type": "package"
    },
    {
      "name": "Grafana container running",
      "status": "failed",
      "duration": 0.3,
      "assertion_type": "command",
      "message": "Container not found",
      "details": {
        "expected": "grafana in output",
        "actual": "prometheus\nnode-exporter"
      }
    }
  ]
}
```

---

## Implementation Roadmap

### Phase 1: Core Framework (Weeks 1-2)
- [ ] CLI structure with Cobra
- [ ] YAML spec parser with validation
- [ ] Core executor engine
- [ ] Basic SSH provider
- [ ] File and package assertions
- [ ] Human-readable output

### Phase 2: Enhanced Assertions (Weeks 3-4)
- [ ] Service assertions
- [ ] User/group assertions
- [ ] Custom command assertions
- [ ] File content matching (regex, contains)
- [ ] Port/network assertions

### Phase 3: Advanced Features (Weeks 5-6)
- [ ] Variable substitution and templating
- [ ] Include/import support
- [ ] Parallel execution
- [ ] JSON and JUnit output formats
- [ ] Configuration file support

### Phase 4: Additional Providers (Future)
- [ ] AWS provider (EC2, S3, IAM, RDS)
- [ ] OpenStack provider
- [ ] Kubernetes provider
- [ ] Local execution provider (no SSH)

---

## Key Suggestions

### 1. Use Existing Libraries
- **SSH**: `golang.org/x/crypto/ssh`
- **YAML**: `gopkg.in/yaml.v3`
- **CLI**: `github.com/spf13/cobra`
- **Testing utilities**: Consider patterns from `testify`, `goss`, `inspec`

### 2. Error Handling
- Distinguish between test failures (expected) and execution errors (unexpected)
- Provide actionable error messages
- Include debug mode for troubleshooting

### 3. Security Considerations
- Support SSH agent forwarding
- Support bastion/jump hosts
- Never log sensitive data (passwords, keys)
- Support vault integration for secrets

### 4. Testing Strategy
- Unit tests for each assertion type
- Integration tests with Docker containers
- Example specs in `examples/` directory
- CI/CD pipeline validation

### 5. Documentation
- Clear README with examples
- Assertion reference documentation
- Provider-specific guides
- Migration guide from similar tools (ServerSpec, etc.)

---

## Example Complete Spec

```yaml
version: "1.0"

metadata:
  name: "Production Web Server Validation"
  description: "Comprehensive validation for production web servers"
  tags: ["production", "web", "docker"]

config:
  fail_fast: false
  parallel: true
  timeout: 300

variables:
  domain: "example.com"
  app_user: "webapp"
  monitoring_path: "/opt/monitoring"

includes:
  - ./common/base-hardening.yaml
  - ./common/docker-baseline.yaml

tests:
  packages:
    - name: "Required packages installed"
      packages:
        - docker-ce>=5:24.0
        - nginx>=1.18
        - ufw
      state: present

  files:
    - name: "Application directories"
      paths:
        - "{{ .variables.monitoring_path }}"
        - /opt/app
        - /var/log/app
      type: directory
      owner: "{{ .variables.app_user }}"
      mode: "0755"

  file_content:
    - name: "Nginx SSL configuration"
      path: /etc/nginx/sites-enabled/default
      contains:
        - "ssl_protocols TLSv1.2 TLSv1.3;"
        - "ssl_prefer_server_ciphers on;"

  services:
    - name: "Critical services running"
      services:
        - docker
        - nginx
        - ufw
      state: running
      enabled: true

  commands:
    - name: "Docker containers healthy"
      command: "docker ps --filter health=healthy --format '{{.Names}}'"
      contains:
        - grafana
        - prometheus

    - name: "Certificate valid for 30+ days"
      command: "openssl x509 -in /etc/nginx/ssl/cert.pem -noout -checkend 2592000"
      exit_code: 0

  sysctl:
    - name: "Security kernel parameters"
      settings:
        net.ipv4.conf.all.rp_filter: "1"
        net.ipv4.tcp_syncookies: "1"
        kernel.randomize_va_space: "2"
```

---

## Advantages Over Existing Tools

1. **Single Binary**: Unlike ServerSpec (Ruby) or Testinfra (Python), Go produces a single static binary
2. **Cloud-Native**: First-class support for cloud providers (AWS, OpenStack) alongside SSH
3. **Parallel by Default**: Go's concurrency makes parallel testing natural
4. **Strongly Typed**: Go's type system catches errors at compile time
5. **Extensible**: Plugin architecture allows custom providers and assertions

This design gives you a solid foundation to build a powerful, extensible infrastructure testing tool. Start with Phase 1 focusing on SSH + basic assertions, then expand from there.
