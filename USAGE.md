# platform-spec Usage Guide

Complete reference for YAML schema, output formats, and examples.

## Providers

### Local Provider

Test your local system.

**Quick Example:**

```bash
platform-spec test local spec.yaml
```

Runs all tests against the local machine. Supports all 14 assertion types.

**Available Assertions:** See [System Test assertions](docs/system/README.md) - all work identically for local testing (14 assertion types).

### Remote Provider

Test remote systems via SSH connection.

**Full Documentation:** [docs/system/README.md](docs/system/README.md)

**Quick Example:**

```bash
platform-spec test remote ubuntu@host spec.yaml
```

**Authentication:**

The remote provider supports two authentication methods:

1. **SSH Key File** (recommended):
   ```bash
   platform-spec test remote ubuntu@host spec.yaml -i ~/.ssh/id_rsa
   ```

2. **SSH Agent**:
   ```bash
   # Ensure SSH agent is running with your key loaded
   eval $(ssh-agent)
   ssh-add ~/.ssh/id_rsa
   platform-spec test remote ubuntu@host spec.yaml
   ```

**Connection Options:**

```bash
# Custom SSH port
platform-spec test remote ubuntu@host spec.yaml -p 2222

# Connection timeout
platform-spec test remote ubuntu@host spec.yaml -t 60

# Verbose output
platform-spec test remote ubuntu@host spec.yaml --verbose
```

See [System Test docs](docs/system/README.md) for all available tests.

### AWS Provider

_Planned - not yet implemented_

### OpenStack Provider

_Planned - not yet implemented_

## YAML Spec Schema

### Complete Schema

```yaml
version: "1.0"

metadata:
  name: "Test Suite Name"
  description: "Description of what this tests"
  tags: ["tag1", "tag2"]

config:
  fail_fast: false # Stop on first failure (default: false)
  parallel: false # Run tests in parallel (default: false)
  timeout: 300 # Global timeout in seconds (default: 300)

variables:
  key: "value" # Variables for future use

tests:
  packages: [] # Package installation tests
  files: [] # File/directory tests
  services: [] # Service status tests
  users: [] # User tests
  groups: [] # Group tests
  file_content: [] # File content tests
  command_content: [] # Command output tests
  docker: [] # Docker container tests
  filesystems: [] # Filesystem mount tests
  ping: [] # Network reachability tests
  dns: [] # DNS resolution tests
  systeminfo: [] # System information validation tests
  http: [] # HTTP endpoint tests
  ports: [] # Port listening tests
```

### Metadata Section

Optional metadata about your test suite:

```yaml
metadata:
  name: "Production Web Server Validation"
  description: "Validates Docker and monitoring setup"
  tags: ["production", "docker", "monitoring"]
```

### Config Section

Optional configuration for test execution:

```yaml
config:
  fail_fast: true # Stop on first test failure
  parallel: false # Enable parallel test execution
  timeout: 600 # Global timeout in seconds
```

### Assertion Types

The following assertions work for both Local and Remote providers:

- [Package Assertions](docs/system/assertions/packages.md) - Check if packages are installed/absent
- [File Assertions](docs/system/assertions/files.md) - Validate file/directory properties
- [Service Assertions](docs/system/assertions/services.md) - Check service status and enabled state
- [User Assertions](docs/system/assertions/users.md) - Validate user properties and group membership
- [Group Assertions](docs/system/assertions/groups.md) - Check if groups exist or are absent
- [File Content Assertions](docs/system/assertions/file_content.md) - Check file contents for strings or regex patterns
- [Command Content Assertions](docs/system/assertions/command_content.md) - Execute commands and validate output or exit codes
- [Docker Assertions](docs/system/assertions/docker.md) - Check Docker container status and properties
- [Filesystem Assertions](docs/system/assertions/filesystems.md) - Check filesystem mount status, type, options, and disk usage
- [Ping Assertions](docs/system/assertions/ping.md) - Check network reachability using ICMP ping
- [DNS Assertions](docs/system/assertions/dns.md) - Check DNS resolution for hostnames
- [System Info Assertions](docs/system/assertions/systeminfo.md) - Validate system properties (OS, architecture, kernel, hostname)
- [HTTP Assertions](docs/system/assertions/http.md) - Test HTTP endpoints for availability, status codes, and response content
- [Port Assertions](docs/system/assertions/ports.md) - Check that network ports are in the expected listening or closed state

## Output

### Human-Readable Format

```
Platform-Spec Test Results
==========================

Spec: Basic System Test
Target: ubuntu@192.168.1.100

✓ Docker packages installed (0.52s)
✓ /opt directory exists (0.11s)
✗ /opt/monitoring/docker-compose.yml file exists (0.10s)
  Path /opt/monitoring/docker-compose.yml does not exist

Tests: 2 passed, 1 failed, 0 skipped, 0 errors
Duration: 0.73s
Status: FAILED
```

**Exit Codes:**

- `0` - All tests passed
- `1` - One or more tests failed

**Symbols:**

- `✓` Passed | `✗` Failed | `○` Skipped | `⚠` Error

## Complete Example

```yaml
version: "1.0"

metadata:
  name: "Docker Host Validation"
  description: "Validates Docker installation and directory structure"
  tags: ["docker", "production"]

config:
  fail_fast: false
  timeout: 300

tests:
  packages:
    - name: "Docker packages installed"
      packages:
        - docker-ce
        - docker-compose-plugin
        - containerd.io
      state: present

    - name: "Unwanted packages absent"
      packages:
        - docker.io # Ubuntu's docker package
      state: absent

  files:
    - name: "Docker data directory"
      path: /var/lib/docker
      type: directory
      owner: root
      group: root
      mode: "0710"

    - name: "Application directories"
      path: /opt/app
      type: directory
      owner: appuser
      group: appuser
      mode: "0755"

    - name: "Docker compose file"
      path: /opt/app/docker-compose.yml
      type: file
      owner: appuser
      group: appuser
      mode: "0644"

    - name: "Sensitive config directory"
      path: /opt/app/secrets
      type: directory
      owner: appuser
      group: appuser
      mode: "0700"
```

## Best Practices

- **Descriptive Names**: Use clear, specific test names
- **Multiple Specs**: Separate concerns into different spec files
- **Version Control**: Store specs alongside infrastructure code
- **Start Simple**: Begin with package and file checks
- **Test Often**: Run during development and deployment
