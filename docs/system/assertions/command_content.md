# Command Content Assertions

Execute custom commands and validate output content or exit codes.

## Schema

```yaml
tests:
  command_content:
    - name: "Test description"
      command: "command to run"   # required
      contains: [str1, str2]      # optional - strings in stdout
      exit_code: 0                # optional - expected exit code
```

At least one of `contains` or `exit_code` must be specified.

## Examples

**Check command output contains string:**
```yaml
tests:
  command_content:
    - name: "Docker filesystem mounted"
      command: df -h
      contains:
        - "/var/lib/docker"
```

**Check multiple strings in output:**
```yaml
tests:
  command_content:
    - name: "Docker info shows config"
      command: docker info
      contains:
        - "Server Version"
        - "Storage Driver"
        - "Logging Driver"
```

**Check exit code:**
```yaml
tests:
  command_content:
    - name: "Backup script succeeds"
      command: /opt/scripts/backup.sh --dry-run
      exit_code: 0
```

**Combined output and exit code:**
```yaml
tests:
  command_content:
    - name: "Service active check"
      command: systemctl status nginx
      exit_code: 0
      contains:
        - "active (running)"
```

## Notes

- Command is executed via SSH on the remote system
- `contains` checks stdout only (not stderr)
- Exit code 0 is not validated unless explicitly specified with non-zero value or when Contains is empty
- Commands run as the connecting user (no sudo by default)
