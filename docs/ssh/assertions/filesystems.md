# Filesystem Assertions

Check filesystem mount status, type, options, and disk usage.

## Schema

```yaml
tests:
  filesystems:
    - name: "Test description"
      path: "/mount/path"               # required
      state: mounted|unmounted          # optional, defaults to mounted
      fstype: "ext4|xfs|tmpfs|..."      # optional
      options: [rw, noexec, nosuid]     # optional mount options
      min_size_gb: 100                  # optional minimum size in GB
      max_usage_percent: 80             # optional maximum usage %
```

## Implementation

Uses `findmnt` to check mount status, filesystem type, and mount options.
Uses `df` to check disk size and usage percentage.

## Examples

**Filesystem mounted:**
```yaml
tests:
  filesystems:
    - name: "Root filesystem mounted"
      path: /
      state: mounted
```

**Filesystem type check:**
```yaml
tests:
  filesystems:
    - name: "Data partition is ext4"
      path: /data
      state: mounted
      fstype: ext4
```

**Mount options validation:**
```yaml
tests:
  filesystems:
    - name: "Tmp partition has noexec"
      path: /tmp
      state: mounted
      options:
        - rw
        - noexec
        - nosuid
```

**Disk space check:**
```yaml
tests:
  filesystems:
    - name: "Data partition has enough space"
      path: /data
      state: mounted
      min_size_gb: 100
      max_usage_percent: 80
```

**Full validation:**
```yaml
tests:
  filesystems:
    - name: "Production data mount validated"
      path: /mnt/data
      state: mounted
      fstype: xfs
      options:
        - rw
        - noatime
      min_size_gb: 500
      max_usage_percent: 75
```

**Check filesystem not mounted:**
```yaml
tests:
  filesystems:
    - name: "Backup mount not active"
      path: /mnt/backup
      state: unmounted
```

## Notes

- Current mount points can be listed with `findmnt` or `df -h`
- State defaults to `mounted` if not specified
- Mount options are checked for presence - actual options may include additional values
- Size validation uses `df` in gigabytes (GB)
- Usage percentage includes reserved blocks (matches `df` output)
- Common filesystem types: `ext4`, `xfs`, `tmpfs`, `nfs`, `btrfs`
- Common mount options: `rw`, `ro`, `noexec`, `nosuid`, `nodev`, `noatime`
