# Port Assertions

Test that network ports/sockets are in the expected listening or closed state.

## Schema

```yaml
tests:
  ports:
    - name: "Test description"
      port: 80
      protocol: "tcp"      # Optional, default: tcp
      state: "listening"   # Optional, default: listening
```

## Fields

| Field | Required | Default | Description |
|-------|----------|---------|-------------|
| `name` | Yes | - | Test description |
| `port` | Yes | - | Port number (1-65535) |
| `protocol` | No | tcp | Protocol type: `tcp` or `udp` |
| `state` | No | listening | Expected state: `listening` or `closed` |

## Implementation

Uses `ss` (socket statistics) command to check port states:
- TCP ports: `ss -tln | grep -E ':PORT\s'`
- UDP ports: `ss -uln | grep -E ':PORT\s'`

The test passes if the actual port state matches the expected state.

## Examples

**Basic web server port:**
```yaml
tests:
  ports:
    - name: "HTTP port listening"
      port: 80
```

**HTTPS port:**
```yaml
tests:
  ports:
    - name: "HTTPS port listening"
      port: 443
      protocol: tcp
      state: listening
```

**Database port:**
```yaml
tests:
  ports:
    - name: "PostgreSQL listening"
      port: 5432
      protocol: tcp
```

**UDP service:**
```yaml
tests:
  ports:
    - name: "DNS port listening"
      port: 53
      protocol: udp
```

**Verify port is closed:**
```yaml
tests:
  ports:
    - name: "Telnet port closed"
      port: 23
      state: closed
```

**Multiple ports:**
```yaml
tests:
  ports:
    - name: "SSH port listening"
      port: 22
      protocol: tcp

    - name: "HTTP port listening"
      port: 80
      protocol: tcp

    - name: "HTTPS port listening"
      port: 443
      protocol: tcp

    - name: "MySQL port listening"
      port: 3306
      protocol: tcp

    - name: "Old FTP port closed"
      port: 21
      state: closed
```

**Common service ports:**
```yaml
tests:
  ports:
    # Web services
    - name: "Nginx HTTP"
      port: 80

    - name: "Nginx HTTPS"
      port: 443

    # Databases
    - name: "PostgreSQL"
      port: 5432

    - name: "MySQL"
      port: 3306

    - name: "Redis"
      port: 6379

    # Monitoring
    - name: "Prometheus"
      port: 9090

    - name: "Grafana"
      port: 3000

    # Message queues
    - name: "RabbitMQ"
      port: 5672

    - name: "Kafka"
      port: 9092
```

## Notes

- Test passes when the actual port state matches the expected state
- `listening` state means the port is bound and accepting connections
- `closed` state means the port is not listening (useful for security checks)
- Requires `ss` command to be available on the target system
- Only checks if the port is listening, not if it's reachable from external networks
- Does not test actual connectivity or service functionality
- For UDP services, some may not show as listening even when running (connectionless protocol behavior)
- Timeout is controlled by global timeout setting in config section
