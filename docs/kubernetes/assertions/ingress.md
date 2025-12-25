# Ingress Tests

Test Kubernetes Ingress existence, hosts, TLS configuration, and ingress class validation.

## YAML Structure

```yaml
tests:
  kubernetes:
    ingress:
      - name: string              # Required: Test name
        ingress: string           # Required: Ingress name
        namespace: string         # Optional: Namespace (default: "default")
        state: string             # Optional: present, absent (default: "present")
        hosts: []string           # Optional: Expected hosts
        tls: boolean              # Optional: Check if TLS is configured
        ingress_class: string     # Optional: Expected ingress class
```

## Fields

### Required

- **name** - Descriptive name for the test
- **ingress** - Ingress name

### Optional

- **namespace** - Namespace where ingress exists (default: `"default"`)
- **state** - Expected state: `present` or `absent` (default: `present`)
- **hosts** - List of hosts that must be configured in the ingress rules
- **tls** - Whether TLS should be configured (default: not checked)
- **ingress_class** - Expected ingress class name (checks both `spec.ingressClassName` and annotation)

## Examples

### Basic Ingress Existence

```yaml
ingress:
  - name: "API ingress exists"
    ingress: api-ingress
    namespace: production
    state: present
```

### Multiple Ingresses

```yaml
ingress:
  - name: "Web app ingress"
    ingress: webapp
    namespace: default

  - name: "API ingress"
    ingress: api
    namespace: default

  - name: "Admin ingress"
    ingress: admin
    namespace: admin-namespace
```

### Verify Ingress Removed

```yaml
ingress:
  - name: "Old ingress removed"
    ingress: deprecated-ingress
    namespace: production
    state: absent
```

### Ingress with Host Validation

```yaml
ingress:
  - name: "API ingress has correct hosts"
    ingress: api-ingress
    namespace: production
    hosts:
      - api.example.com
      - api-v2.example.com
```

### Ingress with TLS

```yaml
ingress:
  - name: "Web ingress has TLS"
    ingress: web-ingress
    namespace: production
    hosts:
      - www.example.com
    tls: true
```

### Ingress with Specific Class

```yaml
ingress:
  - name: "Internal ingress uses nginx"
    ingress: internal-api
    namespace: default
    ingress_class: nginx
```

### Complete Ingress Validation

```yaml
ingress:
  - name: "Production web ingress fully configured"
    ingress: web-ingress
    namespace: production
    state: present
    hosts:
      - www.example.com
      - example.com
    tls: true
    ingress_class: nginx
```

### Multi-tenant Ingress Setup

```yaml
ingress:
  - name: "Tenant A ingress"
    ingress: tenant-a
    namespace: tenant-a
    hosts:
      - tenant-a.example.com
    tls: true
    ingress_class: nginx

  - name: "Tenant B ingress"
    ingress: tenant-b
    namespace: tenant-b
    hosts:
      - tenant-b.example.com
    tls: true
    ingress_class: nginx
```

## kubectl Command

The test executes:

```bash
kubectl get ingress <ingress-name> -n <namespace> -o json 2>&1
```

And validates:
- Exit code 0 = Ingress exists
- Exit code 1 with "not found" = Ingress absent
- `.spec.rules[].host` for host validation
- `.spec.tls` array for TLS configuration
- `.spec.ingressClassName` or annotation `kubernetes.io/ingress.class` for ingress class

## Test Behavior

### State: present

When `state: present`:
- Ingress must exist in the namespace
- Test passes if `kubectl get ingress` succeeds
- Test fails if ingress not found

### State: absent

When `state: absent`:
- Ingress must NOT exist in the namespace
- Test passes if ingress not found
- Test fails if ingress exists

### Host Validation

When `hosts` is specified:
- All listed hosts must be configured in `.spec.rules`
- Hosts are matched exactly (case-sensitive)
- Test fails if any expected host is missing
- Extra hosts not in the list are ignored

### TLS Validation

When `tls: true`:
- Ingress must have at least one TLS configuration in `.spec.tls`
- Test fails if `.spec.tls` is empty or missing
- Does not validate specific TLS hosts or secret names

### Ingress Class Validation

When `ingress_class` is specified:
- Checks `.spec.ingressClassName` first (Kubernetes 1.18+)
- Falls back to annotation `kubernetes.io/ingress.class` (older API)
- Must match exactly (case-sensitive)
- Test fails if class doesn't match or is missing

## Common Patterns

### Public Web Applications

```yaml
ingress:
  - name: "Main website ingress"
    ingress: website
    namespace: production
    hosts:
      - www.example.com
      - example.com
    tls: true
    ingress_class: nginx

  - name: "Blog ingress"
    ingress: blog
    namespace: production
    hosts:
      - blog.example.com
    tls: true
    ingress_class: nginx
```

### API Gateways

```yaml
ingress:
  - name: "Public API gateway"
    ingress: api-public
    namespace: api
    hosts:
      - api.example.com
    tls: true
    ingress_class: nginx

  - name: "Internal API gateway"
    ingress: api-internal
    namespace: api
    hosts:
      - api.internal.example.com
    ingress_class: internal-nginx
```

### Multiple Environments

```yaml
ingress:
  - name: "Production ingress"
    ingress: app
    namespace: production
    hosts:
      - app.example.com
    tls: true
    ingress_class: nginx

  - name: "Staging ingress"
    ingress: app
    namespace: staging
    hosts:
      - app-staging.example.com
    tls: true
    ingress_class: nginx

  - name: "Development ingress"
    ingress: app
    namespace: development
    hosts:
      - app-dev.example.com
    ingress_class: nginx
```

### SSL/TLS Enforcement

```yaml
ingress:
  - name: "Secure web app"
    ingress: secure-app
    namespace: production
    hosts:
      - secure.example.com
    tls: true

  - name: "Payment gateway"
    ingress: payments
    namespace: production
    hosts:
      - pay.example.com
    tls: true
    ingress_class: nginx
```

### Ingress Migration

```yaml
# Before migration
ingress:
  - name: "Old ingress controller"
    ingress: web
    namespace: default
    ingress_class: traefik

# During migration (both active)
ingress:
  - name: "Old controller still active"
    ingress: web-traefik
    namespace: default
    ingress_class: traefik
    state: present

  - name: "New controller ready"
    ingress: web-nginx
    namespace: default
    ingress_class: nginx
    state: present

# After migration
ingress:
  - name: "New controller active"
    ingress: web
    namespace: default
    ingress_class: nginx

  - name: "Old controller removed"
    ingress: web-traefik
    namespace: default
    state: absent
```

## Notes

- **Ingress Class** - Modern Kubernetes (1.18+) uses `.spec.ingressClassName`, older versions use annotations
- **Namespace scope** - Ingresses are namespace-scoped resources
- **Host matching** - Hosts must match exactly, wildcards are treated as literals in tests
- **TLS validation** - Only checks if TLS is configured, not certificate validity
- **Multiple rules** - An ingress can have multiple rules with different hosts

## Tips

### Finding Ingresses

```bash
# List all ingresses in namespace
kubectl get ingress -n default

# Get ingress details
kubectl get ingress api-ingress -n default -o yaml

# Describe ingress
kubectl describe ingress api-ingress -n default

# List ingresses across all namespaces
kubectl get ingress -A
```

### Creating Ingresses for Testing

```bash
# Create basic ingress
kubectl create ingress web \
  --rule="www.example.com/*=web-service:80" \
  -n default

# Create ingress from YAML
cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: api-ingress
  namespace: production
spec:
  ingressClassName: nginx
  rules:
  - host: api.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: api-service
            port:
              number: 80
  tls:
  - hosts:
    - api.example.com
    secretName: api-tls-cert
EOF
```

### Checking Ingress Configuration

```bash
# Get ingress class
kubectl get ingress api-ingress -n default -o jsonpath='{.spec.ingressClassName}'

# Get hosts from ingress
kubectl get ingress api-ingress -n default -o jsonpath='{.spec.rules[*].host}'

# Check TLS configuration
kubectl get ingress api-ingress -n default -o jsonpath='{.spec.tls[*].hosts}'

# Get ingress controller annotation (older API)
kubectl get ingress api-ingress -n default -o jsonpath='{.metadata.annotations.kubernetes\.io/ingress\.class}'
```

### Common Ingress Classes

- `nginx` - NGINX Ingress Controller (most common)
- `traefik` - Traefik Ingress Controller
- `alb` - AWS Application Load Balancer
- `gce` - Google Cloud Load Balancer
- `haproxy` - HAProxy Ingress Controller
- `istio` - Istio Gateway
- `contour` - Contour Ingress Controller

### Troubleshooting

```bash
# Check ingress controller logs
kubectl logs -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx

# Verify backend service exists
kubectl get svc -n default

# Check ingress events
kubectl describe ingress api-ingress -n default | grep Events -A 10

# Test ingress connectivity
curl -H "Host: api.example.com" http://<ingress-controller-ip>/

# Verify TLS certificate
kubectl get secret api-tls-cert -n production -o jsonpath='{.data.tls\.crt}' | base64 -d | openssl x509 -text -noout
```

### Ingress Annotations

Common annotations to check manually (not validated by tests):

```yaml
metadata:
  annotations:
    # SSL redirect
    nginx.ingress.kubernetes.io/ssl-redirect: "true"

    # Rate limiting
    nginx.ingress.kubernetes.io/limit-rps: "10"

    # CORS
    nginx.ingress.kubernetes.io/enable-cors: "true"

    # Rewrite target
    nginx.ingress.kubernetes.io/rewrite-target: /

    # Basic auth
    nginx.ingress.kubernetes.io/auth-type: basic
    nginx.ingress.kubernetes.io/auth-secret: basic-auth
```

### Best Practices

1. **Use TLS** - Always enable TLS for production ingresses
2. **Ingress class** - Explicitly specify ingress class to avoid ambiguity
3. **Namespace organization** - Group related ingresses in the same namespace
4. **Host naming** - Use consistent DNS naming conventions
5. **Cert management** - Use cert-manager for automatic certificate management
6. **Path-based routing** - Use multiple paths in one ingress when possible to reduce resources
