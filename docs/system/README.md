# System Tests

Test operating systems (local or remote) using system-level assertions.

**Note:** System tests work with any provider - Local (testing the current system), SSH (testing remote systems), or Kubernetes (testing via kubectl exec). This documentation covers all system-level test types.

## Usage Examples

```bash
# Test local system
platform-spec test local spec.yaml

# Test remote system via SSH
platform-spec test ssh ubuntu@host spec.yaml

# Test remote system with SSH key
platform-spec test ssh -i ~/.ssh/key.pem ubuntu@host spec.yaml

# Test Kubernetes pod (kubectl exec)
platform-spec test kubernetes spec.yaml --context prod
```

For SSH-specific authentication and connection options, see the [SSH Provider documentation](../providers/ssh.md).

## Available Test Types

System tests cover 14 different types of OS-level validations:

### Package Assertions
Check if packages are installed or absent on the system.

[View Package Assertions →](assertions/packages.md)

### File Assertions
Validate file and directory properties (ownership, permissions, existence).

[View File Assertions →](assertions/files.md)

### Service Assertions
Check service status (running, stopped) and enabled state.

[View Service Assertions →](assertions/services.md)

### User Assertions
Validate user properties including shell, home directory, and group membership.

[View User Assertions →](assertions/users.md)

### Group Assertions
Check if groups exist or are absent.

[View Group Assertions →](assertions/groups.md)

### File Content Assertions
Check file contents for strings or regex patterns.

[View File Content Assertions →](assertions/file_content.md)

### Command Content Assertions
Execute custom commands and validate output or exit codes.

[View Command Content Assertions →](assertions/command_content.md)

### Docker Assertions
Check Docker container status and properties.

[View Docker Assertions →](assertions/docker.md)

### Filesystem Assertions
Check filesystem mount status, type, options, and disk usage.

[View Filesystem Assertions →](assertions/filesystems.md)

### Ping Assertions
Check network reachability using ICMP ping.

[View Ping Assertions →](assertions/ping.md)

### DNS Assertions
Check DNS resolution for hostnames.

[View DNS Assertions →](assertions/dns.md)

### System Info Assertions
Validate system properties including OS, architecture, kernel version, and hostname.

[View System Info Assertions →](assertions/systeminfo.md)

### HTTP Assertions
Test HTTP endpoints for availability, status codes, and response content.

[View HTTP Assertions →](assertions/http.md)

### Port Assertions
Check that network ports/sockets are in the expected listening or closed state.

[View Port Assertions →](assertions/ports.md)

## Requirements

The system under test must have the following commands available:
- **Package managers**: `dpkg`, `rpm`, or `apk` (for package tests)
- **File commands**: `stat`, `test` (for file tests)
- **Content search**: `grep` (for file/command content tests)
- **User/group commands**: `id`, `getent` (for user/group tests)
- **Service commands**: `systemctl` (for service tests)
- **Docker**: `docker inspect` (for Docker tests)
- **Filesystem**: `findmnt`, `df` (for filesystem tests)
- **Network**: `ping`, `dig` or `getent` (for ping/DNS tests)
- **System info**: `uname`, `hostname`, `cat` (for systeminfo tests)
- **HTTP**: `curl` (for HTTP tests)
- **Sockets**: `ss` (for port tests)
- **Shell**: `bash` or compatible

**Note**: When using SSH provider, the SSH user must have permissions to execute these commands.

## Supported Distributions

| Distribution | Package Manager | Tested |
|--------------|----------------|--------|
| Ubuntu/Debian | dpkg | ✅ |
| RHEL/CentOS/Fedora | rpm | ✅ |
| Alpine | apk | ✅ |

## Examples

### Basic System Check
```yaml
version: "1.0"
metadata:
  name: "Basic System Check"

tests:
  packages:
    - name: "Essential packages installed"
      packages: [bash, coreutils]
      state: present

  files:
    - name: "Root filesystem"
      path: /
      type: directory
```

### Production Server Validation
```yaml
version: "1.0"
metadata:
  name: "Production Server Validation"

tests:
  packages:
    - name: "Docker installed"
      packages:
        - docker-ce
        - docker-compose-plugin
      state: present

  files:
    - name: "Application directory"
      path: /opt/myapp
      type: directory
      owner: appuser
      mode: "0755"
```

## Limitations

- Read-only operations only (no state changes to the system)
- No sudo support (commands run as the executing user)
- Single system at a time (no parallel multi-system testing yet)
