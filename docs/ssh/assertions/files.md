# File Assertions

Validate file and directory properties.

## Schema

```yaml
tests:
  files:
    - name: "Test description"
      path: "/path/to/file"
      type: file|directory
      owner: "username"    # optional
      group: "groupname"   # optional
      mode: "0755"         # optional
```

## Permission Modes

Must be quoted octal strings:

| Mode | Permission | Use Case |
|------|------------|----------|
| `"0644"` | rw-r--r-- | Regular files |
| `"0755"` | rwxr-xr-x | Directories, executables |
| `"0700"` | rwx------ | Private directories |
| `"0600"` | rw------- | Private files (SSH keys) |
| `"1777"` | rwxrwxrwt | Sticky bit (/tmp) |

## Examples

**File existence:**
```yaml
tests:
  files:
    - name: "Config file exists"
      path: /etc/myapp/config.yml
      type: file
```

**Directory with permissions:**
```yaml
tests:
  files:
    - name: "Application directory"
      path: /opt/myapp
      type: directory
      owner: appuser
      group: appuser
      mode: "0755"
```

**Secure file permissions:**
```yaml
tests:
  files:
    - name: "SSH private key"
      path: /home/user/.ssh/id_rsa
      type: file
      owner: user
      mode: "0600"
```

**Multiple files:**
```yaml
tests:
  files:
    - name: "App directory"
      path: /opt/app
      type: directory
      owner: appuser
      mode: "0755"

    - name: "App secrets"
      path: /opt/app/secrets
      type: directory
      owner: appuser
      mode: "0700"

    - name: "Config file"
      path: /opt/app/config.yml
      type: file
      owner: appuser
      mode: "0644"
```
