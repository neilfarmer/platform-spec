# Package Assertions

Check if packages are installed or absent.

## Schema

```yaml
tests:
  packages:
    - name: "Test description"
      packages: [package1, package2]
      state: present|absent
```

## Supported Package Managers

- **dpkg** - Debian, Ubuntu
- **rpm** - RHEL, CentOS, Fedora
- **apk** - Alpine

Package manager is auto-detected.

## Examples

**Packages installed:**
```yaml
tests:
  packages:
    - name: "Docker packages installed"
      packages: [docker-ce, docker-compose-plugin]
      state: present
```

**Packages NOT installed:**
```yaml
tests:
  packages:
    - name: "Unwanted packages removed"
      packages: [telnet, ftp]
      state: absent
```

**Multiple checks:**
```yaml
tests:
  packages:
    - name: "Web server packages"
      packages: [nginx, certbot]
      state: present

    - name: "Legacy packages removed"
      packages: [apache2]
      state: absent
```

## Common Package Names

| Software | Ubuntu/Debian | RHEL/CentOS | Alpine |
|----------|---------------|-------------|--------|
| Docker | docker-ce | docker-ce | docker |
| Python 3 | python3 | python3 | python3 |
| Nginx | nginx | nginx | nginx |
