# Docker Assertions

Check Docker container status and properties.

## Schema

```yaml
tests:
  docker:
    - name: "Test description"
      container: "containername"      # single container
      # OR
      containers: [cont1, cont2]      # multiple containers
      state: running|stopped|exists   # optional, defaults to running
      image: "image:tag"               # optional
      restart_policy: no|always|on-failure|unless-stopped  # optional
      health: healthy|unhealthy|starting|none              # optional
```

## Implementation

Uses `docker inspect` to check container state and properties.

## Examples

**Container running:**
```yaml
tests:
  docker:
    - name: "Web server running"
      container: nginx-web
      state: running
```

**Container with image check:**
```yaml
tests:
  docker:
    - name: "Nginx container with correct image"
      container: nginx-web
      state: running
      image: "nginx:latest"
```

**Container with restart policy:**
```yaml
tests:
  docker:
    - name: "Production containers auto-restart"
      container: app-prod
      state: running
      restart_policy: always
```

**Container with health check:**
```yaml
tests:
  docker:
    - name: "App container healthy"
      container: myapp
      state: running
      health: healthy
```

**Multiple containers:**
```yaml
tests:
  docker:
    - name: "Core services running"
      containers:
        - nginx-web
        - redis-cache
        - postgres-db
      state: running

    - name: "Test containers stopped"
      containers:
        - test-runner
        - debug-container
      state: stopped
```

**Full validation:**
```yaml
tests:
  docker:
    - name: "Production app validated"
      container: myapp-prod
      state: running
      image: "myapp:v1.2.3"
      restart_policy: unless-stopped
      health: healthy
```

## Notes

- Container names can be obtained with `docker ps --format '{{.Names}}'`
- State defaults to `running` if not specified
- Image matching is partial - `nginx` matches `nginx:latest`, `nginx:1.21`, etc.
- Health status only applies to containers with HEALTHCHECK defined
