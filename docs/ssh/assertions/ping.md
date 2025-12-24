# Ping Assertions

Check network reachability using ICMP ping.

## Schema

```yaml
tests:
  ping:
    - name: "Test description"
      host: "hostname or IP"
```

## Implementation

Uses `ping -c 1 -W 5 <host>` to send a single ICMP packet with 5 second timeout.

## Examples

**Ping internal server:**
```yaml
tests:
  ping:
    - name: "Database server reachable"
      host: db.internal
```

**Ping external service:**
```yaml
tests:
  ping:
    - name: "Internet connectivity"
      host: 8.8.8.8
```

**Multiple hosts:**
```yaml
tests:
  ping:
    - name: "App server reachable"
      host: app.example.com

    - name: "Cache server reachable"
      host: redis.internal

    - name: "External API reachable"
      host: api.github.com
```

## Notes

- Tests pass if the host responds to ICMP ping (exit code 0)
- Uses 1 packet with 5 second timeout for faster results
- Requires ICMP to be allowed by firewalls between source and destination
- Some hosts may block ICMP ping for security - use DNS assertions as alternative
- On Linux, ping requires NET_RAW capability or is setuid root
