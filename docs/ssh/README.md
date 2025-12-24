# SSH Provider

Test Linux systems via SSH connection.

**Note:** The assertion types documented here (packages, files, services, etc.) work for both SSH and Local providers. This page covers SSH-specific connection details.

## Authentication

The SSH provider supports:
1. **SSH Key File** - Via `-i` flag
2. **SSH Agent** - Via `SSH_AUTH_SOCK` environment variable (fallback)

```bash
# Using SSH agent
platform-spec test ssh ubuntu@host spec.yaml

# Using key file
platform-spec test ssh -i ~/.ssh/prod.pem ubuntu@host spec.yaml
```

## Connection Options

| Flag | Description | Default |
|------|-------------|---------|
| `-p, --port <int>` | SSH port | 22 |
| `-t, --timeout <int>` | Connection timeout (seconds) | 30 |

## Target Format

```
[user@]hostname
```

**Examples:**
- `ubuntu@192.168.1.100`
- `root@web-server.example.com`
- `web-server.example.com` (defaults to `root`)

## Supported Assertions

The SSH provider supports the following assertion types:

### Package Assertions
Check if packages are installed or absent on the remote system.

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

## Requirements

The SSH user must have permissions to execute:
- Package managers: `dpkg`, `rpm`, or `apk`
- File commands: `stat`, `test`
- Content search: `grep`
- User/group commands: `id`, `getent`
- Service commands: `systemctl`
- Docker commands: `docker inspect` (if using Docker assertions)
- Filesystem commands: `findmnt`, `df` (if using filesystem assertions)
- Shell: `bash` or compatible

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

- No sudo support (commands run as connecting user)
- Single host at a time (no parallel multi-host testing yet)
- Read-only operations only (no state changes)
