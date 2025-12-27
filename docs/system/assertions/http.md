# HTTP Assertions

Test HTTP endpoints for availability, status codes, and response content.

## Schema

```yaml
tests:
  http:
    - name: "Test description"
      url: "http://example.com"
      status_code: 200           # Optional, default: 200
      contains: ["string"]       # Optional, strings in response body
      method: "GET"              # Optional, default: GET
      insecure: false            # Optional, skip TLS verification
      follow_redirects: false    # Optional, follow HTTP redirects
```

## Fields

| Field | Required | Default | Description |
|-------|----------|---------|-------------|
| `name` | Yes | - | Test description |
| `url` | Yes | - | Full HTTP/HTTPS URL to test |
| `status_code` | No | 200 | Expected HTTP status code |
| `contains` | No | - | List of strings that must be in response body |
| `method` | No | GET | HTTP method (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS) |
| `insecure` | No | false | Skip TLS certificate verification |
| `follow_redirects` | No | false | Follow HTTP redirects (3xx responses) |

## Implementation

Uses `curl` to make HTTP requests:
- `-s`: Silent mode
- `-w "\n%{http_code}"`: Append status code to output
- `-X METHOD`: Specify HTTP method
- `-k`: Skip TLS verification (if insecure: true)

## Examples

**Basic endpoint check:**
```yaml
tests:
  http:
    - name: "Web server running"
      url: http://localhost:8080
```

**Check specific status code:**
```yaml
tests:
  http:
    - name: "API health endpoint"
      url: https://api.example.com/health
      status_code: 200
```

**Validate response content:**
```yaml
tests:
  http:
    - name: "API returns expected data"
      url: https://api.example.com/status
      contains:
        - "status"
        - "healthy"
```

**POST request:**
```yaml
tests:
  http:
    - name: "API accepts POST"
      url: https://api.example.com/webhook
      method: POST
      status_code: 202
```

**Self-signed certificate:**
```yaml
tests:
  http:
    - name: "Internal API with self-signed cert"
      url: https://internal-api.local
      insecure: true
```

**Follow redirects:**
```yaml
tests:
  http:
    - name: "Site with redirect"
      url: http://example.com
      follow_redirects: true
      status_code: 200  # Final status after following redirects
```

**Multiple endpoints:**
```yaml
tests:
  http:
    - name: "Frontend accessible"
      url: http://localhost:3000
      contains:
        - "<html>"

    - name: "API health check"
      url: http://localhost:8080/health
      status_code: 200
      contains:
        - "ok"

    - name: "Metrics endpoint"
      url: http://localhost:9090/metrics
      contains:
        - "# HELP"
```

## Notes

- Test passes if status code matches and all `contains` strings are found in response body
- For `contains` checks, the entire response body is searched (case-sensitive)
- By default, does NOT follow redirects - set `follow_redirects: true` to follow 3xx responses
- When `follow_redirects: true`, the status code checked is the final response after all redirects
- Does not verify response headers (only status code and body)
- Timeout is controlled by global timeout setting in config section
- Requires `curl` to be installed on the target system
- Use `insecure: true` only for development/testing with self-signed certificates
