# DNS Assertions

Check DNS resolution for hostnames.

## Schema

```yaml
tests:
  dns:
    - name: "Test description"
      host: "hostname to resolve"
```

## Implementation

Uses `dig +short <host>` (preferred) or falls back to `getent hosts <host>`.
Test passes if the hostname resolves to any IP address(es).

## Examples

**Resolve internal hostname:**
```yaml
tests:
  dns:
    - name: "Internal DNS works"
      host: myapp.internal
```

**Resolve external hostname:**
```yaml
tests:
  dns:
    - name: "External DNS works"
      host: google.com
```

**Multiple DNS checks:**
```yaml
tests:
  dns:
    - name: "Database hostname resolves"
      host: db.example.com

    - name: "API hostname resolves"
      host: api.example.com

    - name: "Internal service resolves"
      host: service.local
```

## Notes

- Tests pass if the hostname resolves to one or more IP addresses
- Does not validate which specific IP(s) are returned - only that resolution succeeds
- Uses `dig` if available, falls back to `getent hosts` for compatibility
- Respects system DNS configuration (`/etc/resolv.conf`, `/etc/hosts`)
- Useful for validating DNS propagation and internal DNS configuration
- More reliable than ping for connectivity testing when ICMP is blocked
