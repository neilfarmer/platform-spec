# platform-spec Usage Guide

Complete reference for YAML schema, output formats, and examples.

## Providers

### Local Provider

Test your local system.

**Quick Example:**

```bash
platform-spec test local spec.yaml
```

Runs all tests against the local machine. Supports all 8 assertion types (packages, files, services, users, groups, file_content, command_content, docker).

**Available Assertions:** See [SSH Provider assertions](docs/ssh/README.md) - all work identically for local testing.

### SSH Provider

Test Linux systems via SSH connection.

**Full Documentation:** [docs/ssh/README.md](docs/ssh/README.md)

**Quick Example:**

```bash
platform-spec test ssh ubuntu@host spec.yaml
```

See [SSH Provider docs](docs/ssh/README.md) for authentication, connection options, and all available tests.

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

The following assertions work for both Local and SSH providers:

- [Package Assertions](docs/ssh/assertions/packages.md) - Check if packages are installed/absent
- [File Assertions](docs/ssh/assertions/files.md) - Validate file/directory properties
- [Service Assertions](docs/ssh/assertions/services.md) - Check service status and enabled state
- [User Assertions](docs/ssh/assertions/users.md) - Validate user properties and group membership
- [Group Assertions](docs/ssh/assertions/groups.md) - Check if groups exist or are absent
- [File Content Assertions](docs/ssh/assertions/file_content.md) - Check file contents for strings or regex patterns
- [Command Content Assertions](docs/ssh/assertions/command_content.md) - Execute commands and validate output or exit codes
- [Docker Assertions](docs/ssh/assertions/docker.md) - Check Docker container status and properties

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
