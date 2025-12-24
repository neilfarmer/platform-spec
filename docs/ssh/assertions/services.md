# Service Assertions

Check service status and enabled state.

## Schema

```yaml
tests:
  services:
    - name: "Test description"
      service: "servicename"       # single service
      # OR
      services: [svc1, svc2]       # multiple services
      state: running|stopped       # required
      enabled: true|false          # optional
```

## Service Manager

Uses `systemctl` (systemd) to check service status.

## Examples

**Service running:**
```yaml
tests:
  services:
    - name: "Docker running"
      service: docker
      state: running
```

**Service running and enabled:**
```yaml
tests:
  services:
    - name: "Docker running and enabled"
      service: docker
      state: running
      enabled: true
```

**Service stopped:**
```yaml
tests:
  services:
    - name: "Telnet disabled"
      service: telnet
      state: stopped
```

**Multiple services:**
```yaml
tests:
  services:
    - name: "Critical services running"
      services:
        - docker
        - nginx
        - postgresql
      state: running
      enabled: true

    - name: "Unwanted services stopped"
      services:
        - telnet
        - ftp
      state: stopped
```

## Common Services

| Service | Description |
|---------|-------------|
| docker | Docker daemon |
| nginx | Nginx web server |
| apache2 / httpd | Apache web server |
| postgresql | PostgreSQL database |
| mysql / mariadb | MySQL/MariaDB database |
| ssh / sshd | SSH server |
| ufw | Uncomplicated Firewall |
| firewalld | Firewalld |
