# System Information Assertions

Validate system properties including OS, architecture, kernel version, and hostname.

## Schema

```yaml
tests:
  systeminfo:
    - name: "Test description"
      os: "ubuntu"                    # optional - OS name
      os_version: "22.04"             # optional - OS version
      arch: "x86_64"                  # optional - architecture
      kernel_version: "5.15"          # optional - kernel version
      hostname: "web01"               # optional - short hostname
      fqdn: "web01.example.com"       # optional - FQDN
      version_match: exact            # optional - "exact" or "prefix" (default: exact)
```

## Implementation

Gathers system information from standard Linux commands:
- OS info: `/etc/os-release` (ID and VERSION_ID fields)
- Architecture: `uname -m`
- Kernel version: `uname -r`
- Hostname: `hostname -s`
- FQDN: `hostname -f`

Test passes if all specified fields match the actual system values.

## Version Matching

The `version_match` field controls how `os_version` and `kernel_version` are compared:

- `exact` (default): Versions must match exactly
  - `"22.04"` only matches `"22.04"`
- `prefix`: Versions match if actual starts with expected
  - `"22.04"` matches `"22.04"`, `"22.04.1"`, `"22.04.2"`, etc.
  - `"5.15"` matches `"5.15"`, `"5.15.0"`, `"5.15.0-91-generic"`, etc.

## Examples

**Basic OS validation:**
```yaml
tests:
  systeminfo:
    - name: "Ubuntu 22.04 system"
      os: ubuntu
      os_version: "22.04"
```

**Exact version matching:**
```yaml
tests:
  systeminfo:
    - name: "Exact Ubuntu 22.04.0"
      os: ubuntu
      os_version: "22.04.0"
      version_match: exact
```

**Prefix version matching:**
```yaml
tests:
  systeminfo:
    - name: "Ubuntu 22.04.x"
      os: ubuntu
      os_version: "22.04"
      version_match: prefix  # matches 22.04, 22.04.1, 22.04.2, etc.
```

**Architecture validation:**
```yaml
tests:
  systeminfo:
    - name: "x86_64 architecture"
      arch: x86_64
```

**Hostname validation:**
```yaml
tests:
  systeminfo:
    - name: "Production web server"
      hostname: web01
      fqdn: web01.prod.example.com
```

**Full system validation:**
```yaml
tests:
  systeminfo:
    - name: "Production server requirements"
      os: ubuntu
      os_version: "22.04"
      arch: x86_64
      kernel_version: "5.15"
      hostname: web01
      fqdn: web01.prod.example.com
      version_match: prefix
```

**Multiple systems:**
```yaml
tests:
  systeminfo:
    - name: "Web servers are Ubuntu"
      os: ubuntu
      arch: x86_64

    - name: "Database servers are RHEL"
      os: rhel
      os_version: "8"
```

## Notes

- All fields are optional - specify only what you want to validate
- Test passes if all specified fields match
- OS names come from `/etc/os-release` ID field (typically lowercase: ubuntu, debian, rhel, centos, alpine, etc.)
- All gathered system info is included in test result details for reference
- Default matching is exact - use `version_match: prefix` for more flexible version checking
- Architecture values vary by system: `x86_64`, `aarch64`, `armv7l`, etc.
