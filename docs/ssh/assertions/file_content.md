# File Content Assertions

Check file contents for strings or regex patterns.

## Schema

```yaml
tests:
  file_content:
    - name: "Test description"
      path: "/path/to/file"      # required
      contains: [str1, str2]      # optional - strings to search for
      matches: "regex pattern"    # optional - regex pattern to match
```

At least one of `contains` or `matches` must be specified.

## Examples

**File contains string:**
```yaml
tests:
  file_content:
    - name: "Config has log driver setting"
      path: /etc/docker/daemon.json
      contains:
        - "log-driver"
```

**File contains multiple strings:**
```yaml
tests:
  file_content:
    - name: "Docker config complete"
      path: /etc/docker/daemon.json
      contains:
        - "log-driver"
        - "metrics-addr"
        - "storage-driver"
```

**File matches regex pattern:**
```yaml
tests:
  file_content:
    - name: "SSH root login disabled"
      path: /etc/ssh/sshd_config
      matches: "^PermitRootLogin no$"
```

**Combined contains and matches:**
```yaml
tests:
  file_content:
    - name: "App config valid"
      path: /etc/app/config.yml
      contains:
        - "database:"
        - "redis:"
      matches: "^version: [0-9]+"
```

## Notes

- Uses `grep -F` for literal string matching (no regex interpretation in `contains`)
- Uses `grep -E` for extended regex in `matches`
- File must exist and be readable
- All strings in `contains` must be found for test to pass
