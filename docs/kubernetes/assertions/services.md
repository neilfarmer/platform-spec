# Service Tests

Test Kubernetes service type, ports, and selectors.

## YAML Structure

```yaml
tests:
  kubernetes:
    services:
      - name: string              # Required: Test name
        service: string           # Required: Service name
        namespace: string         # Optional: Namespace (default: from config or "default")
        type: string              # Optional: ClusterIP, NodePort, LoadBalancer, ExternalName
        ports:                    # Optional: List of expected ports
          - port: integer         # Port number
            protocol: string      # TCP, UDP, or SCTP (default: TCP)
        selector:                 # Optional: Expected selector labels
          key: value
```

## Fields

### Required

- **name** - Descriptive name for the test
- **service** - Name of the service to test

### Optional

- **namespace** - Namespace where service exists (default: from config or `"default"`)
- **type** - Service type: `ClusterIP`, `NodePort`, `LoadBalancer`, or `ExternalName`
- **ports** - List of ports that must exist (all must match)
- **selector** - Map of selector labels that route traffic to pods

## Examples

### Basic Existence Check

```yaml
services:
  - name: "App service exists"
    service: myapp-service
    namespace: default
```

### Check Service Type

```yaml
services:
  - name: "Internal service is ClusterIP"
    service: database
    namespace: default
    type: ClusterIP

  - name: "External service is LoadBalancer"
    service: web-app
    namespace: default
    type: LoadBalancer
```

### Verify Ports

```yaml
services:
  - name: "Web service exposes port 80"
    service: nginx-service
    namespace: default
    type: ClusterIP
    ports:
      - port: 80
        protocol: TCP
```

### Multiple Ports

```yaml
services:
  - name: "App service exposes HTTP and HTTPS"
    service: web-app
    namespace: production
    type: LoadBalancer
    ports:
      - port: 80
        protocol: TCP
      - port: 443
        protocol: TCP
```

### Check Selector

```yaml
services:
  - name: "Service routes to app pods"
    service: myapp-service
    namespace: default
    type: ClusterIP
    selector:
      app: myapp
      version: v2
```

### Complete Validation

```yaml
services:
  - name: "Production API service configured correctly"
    service: api-service
    namespace: production
    type: ClusterIP
    ports:
      - port: 8080
        protocol: TCP
    selector:
      app: api
      tier: backend
```

## kubectl Command

The test executes:

```bash
kubectl get service <service-name> -n <namespace> -o json
```

And validates:
- `.spec.type` for service type
- `.spec.ports` for port configuration
- `.spec.selector` for pod selector labels

## Test Behavior

### Type Check

- Compares `.spec.type` with expected type
- Types: `ClusterIP`, `NodePort`, `LoadBalancer`, `ExternalName`
- If not specified, any type is acceptable

### Port Check

For each port in the `ports` list:
- Must find a matching port in `.spec.ports`
- Both `port` number and `protocol` must match
- All specified ports must exist (extra ports are ignored)
- Protocol defaults to `TCP` if not specified

### Selector Check

- Compares `.spec.selector` with expected labels
- All specified labels must match exactly
- Extra selector labels are ignored

## Common Patterns

### Internal Services

```yaml
services:
  - name: "Database is internal only"
    service: postgres
    namespace: default
    type: ClusterIP
    ports:
      - port: 5432
        protocol: TCP
    selector:
      app: postgres
```

### External Services

```yaml
services:
  - name: "Web app is externally accessible"
    service: web-app
    namespace: production
    type: LoadBalancer
    ports:
      - port: 80
        protocol: TCP
      - port: 443
        protocol: TCP
```

### Headless Services

```yaml
services:
  - name: "StatefulSet headless service"
    service: postgres-headless
    namespace: default
    type: ClusterIP
    selector:
      app: postgres
```

### Multi-Protocol Services

```yaml
services:
  - name: "DNS service supports TCP and UDP"
    service: kube-dns
    namespace: kube-system
    type: ClusterIP
    ports:
      - port: 53
        protocol: TCP
      - port: 53
        protocol: UDP
```

### Kubernetes API Service

```yaml
services:
  - name: "Kubernetes API is available"
    service: kubernetes
    namespace: default
    type: ClusterIP
    ports:
      - port: 443
        protocol: TCP
```

## Notes

- **Type** - ClusterIP is default for most services
- **Ports** - Service port (external) vs targetPort (pod) - test checks service port
- **Selector** - Must match pod labels for traffic routing
- **Protocols** - TCP, UDP, SCTP supported; TCP is default
- **Namespace defaults** - Uses `kubernetes_namespace` from config if not specified

## Tips

### Check Service Details

```bash
# View service
kubectl get service myapp -n default

# Check service type and ports
kubectl get service myapp -n default -o jsonpath='{.spec.type}: {.spec.ports}'

# View full service spec
kubectl describe service myapp -n default

# Test service connectivity
kubectl run -it --rm debug --image=busybox --restart=Never -- wget -O- http://myapp:8080
```

### Service Types

**ClusterIP** (default):
- Internal cluster IP
- Only accessible within cluster
- Most common for internal services

**NodePort**:
- Exposes service on each node's IP at a static port
- Accessible from outside cluster via `<NodeIP>:<NodePort>`
- Port range: 30000-32767

**LoadBalancer**:
- Creates external load balancer (cloud provider)
- Assigns external IP
- For production external services

**ExternalName**:
- Maps service to DNS name
- No proxy, no selector
- For external services

### Troubleshooting

**Service has no endpoints:**
- Check selector matches pod labels: `kubectl get endpoints myapp -n default`
- Verify pods exist: `kubectl get pods -l app=myapp -n default`

**Wrong service type:**
- Update service: `kubectl edit service myapp -n default`
- Or delete and recreate

**Port not accessible:**
- Check firewall rules (for LoadBalancer/NodePort)
- Verify pod is listening on targetPort
- Check network policies
