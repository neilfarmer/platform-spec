# Secret Tests

Test Kubernetes Secret existence, type, and key validation.

## YAML Structure

```yaml
tests:
  kubernetes:
    secrets:
      - name: string              # Required: Test name
        secret: string            # Required: Secret name
        namespace: string         # Optional: Namespace (default: "default")
        state: string             # Optional: present, absent (default: "present")
        type: string              # Optional: Secret type validation
        has_keys: []string        # Optional: Keys that must exist in data
```

## Fields

### Required

- **name** - Descriptive name for the test
- **secret** - Secret name

### Optional

- **namespace** - Namespace where secret exists (default: `"default"`)
- **state** - Expected state: `present` or `absent` (default: `present`)
- **type** - Secret type to validate (default: not checked)
  - Valid types: `Opaque`, `kubernetes.io/service-account-token`, `kubernetes.io/dockercfg`, `kubernetes.io/dockerconfigjson`, `kubernetes.io/basic-auth`, `kubernetes.io/ssh-auth`, `kubernetes.io/tls`, `bootstrap.kubernetes.io/token`
- **has_keys** - List of keys that must exist in the secret's data (default: not checked)

## Examples

### Basic Secret Existence

```yaml
secrets:
  - name: "Database password exists"
    secret: db-password
    namespace: production
    state: present
```

### Multiple Secrets

```yaml
secrets:
  - name: "App credentials exist"
    secret: app-creds
    namespace: default

  - name: "API keys exist"
    secret: api-keys
    namespace: default

  - name: "TLS certificate exists"
    secret: tls-cert
    namespace: ingress
```

### Verify Secret Removed

```yaml
secrets:
  - name: "Old credentials removed"
    secret: deprecated-creds
    namespace: production
    state: absent
```

### TLS Secret

```yaml
secrets:
  - name: "TLS certificate valid"
    secret: tls-cert
    namespace: ingress
    state: present
    type: kubernetes.io/tls
    has_keys:
      - tls.crt
      - tls.key
```

### Docker Registry Secret

```yaml
secrets:
  - name: "Registry credentials configured"
    secret: regcred
    namespace: default
    type: kubernetes.io/dockerconfigjson
    has_keys:
      - .dockerconfigjson
```

### Basic Auth Secret

```yaml
secrets:
  - name: "Basic auth configured"
    secret: basic-auth
    namespace: default
    type: kubernetes.io/basic-auth
    has_keys:
      - username
      - password
```

### SSH Auth Secret

```yaml
secrets:
  - name: "SSH key exists"
    secret: git-ssh-key
    namespace: ci-cd
    type: kubernetes.io/ssh-auth
    has_keys:
      - ssh-privatekey
```

### Opaque Secret with Custom Keys

```yaml
secrets:
  - name: "App configuration complete"
    secret: app-config
    namespace: production
    type: Opaque
    has_keys:
      - database-url
      - api-key
      - encryption-key
```

### Service Account Token

```yaml
secrets:
  - name: "Service account token exists"
    secret: default-token-abc12
    namespace: default
    type: kubernetes.io/service-account-token
```

## kubectl Command

The test executes:

```bash
kubectl get secret <secret-name> -n <namespace> -o json 2>&1
```

And validates:
- Exit code 0 = Secret exists
- Exit code 1 with "not found" = Secret absent
- `.type` for secret type validation
- `.data` keys for key validation (keys are base64 encoded in secrets)

## Test Behavior

### State: present

When `state: present`:
- Secret must exist in the namespace
- Test passes if `kubectl get secret` succeeds
- Test fails if secret not found

### State: absent

When `state: absent`:
- Secret must NOT exist in the namespace
- Test passes if secret not found
- Test fails if secret exists

### Type Validation

When `type` is specified:
- Secret's `.type` field must match exactly
- Common types have specific use cases (see Secret Types section)
- Fails if type doesn't match

### Key Validation

When `has_keys` is specified:
- All listed keys must exist in `.data`
- Keys are case-sensitive
- Only checks key existence, not values
- Secret data is base64 encoded but test checks keys only

## Secret Types

### Opaque (Default)

Generic secret type for arbitrary data:
```yaml
secrets:
  - name: "Generic secret"
    secret: my-secret
    type: Opaque
    has_keys:
      - api-key
      - database-url
```

### kubernetes.io/tls

TLS certificates and private keys:
```yaml
secrets:
  - name: "TLS cert for ingress"
    secret: tls-secret
    type: kubernetes.io/tls
    has_keys:
      - tls.crt
      - tls.key
```

Required keys: `tls.crt`, `tls.key`

### kubernetes.io/dockerconfigjson

Docker registry authentication:
```yaml
secrets:
  - name: "Private registry access"
    secret: regcred
    type: kubernetes.io/dockerconfigjson
    has_keys:
      - .dockerconfigjson
```

Required key: `.dockerconfigjson`

### kubernetes.io/basic-auth

HTTP basic authentication:
```yaml
secrets:
  - name: "HTTP basic auth"
    secret: basic-auth
    type: kubernetes.io/basic-auth
    has_keys:
      - username
      - password
```

Required keys: `username`, `password`

### kubernetes.io/ssh-auth

SSH private key authentication:
```yaml
secrets:
  - name: "Git SSH key"
    secret: git-ssh
    type: kubernetes.io/ssh-auth
    has_keys:
      - ssh-privatekey
```

Required key: `ssh-privatekey`

### kubernetes.io/service-account-token

Service account tokens (usually auto-generated):
```yaml
secrets:
  - name: "SA token exists"
    secret: default-token-abc12
    type: kubernetes.io/service-account-token
```

## Common Patterns

### Application Credentials

```yaml
secrets:
  - name: "Database credentials"
    secret: postgres-creds
    namespace: production
    type: Opaque
    has_keys:
      - username
      - password
      - host

  - name: "API keys"
    secret: external-apis
    namespace: production
    type: Opaque
    has_keys:
      - stripe-api-key
      - sendgrid-api-key
```

### TLS/SSL Certificates

```yaml
secrets:
  - name: "Ingress TLS certificate"
    secret: example-com-tls
    namespace: ingress-nginx
    type: kubernetes.io/tls
    has_keys:
      - tls.crt
      - tls.key

  - name: "Wildcard certificate"
    secret: wildcard-tls
    namespace: ingress-nginx
    type: kubernetes.io/tls
    has_keys:
      - tls.crt
      - tls.key
```

### Docker Registry Access

```yaml
secrets:
  - name: "DockerHub credentials"
    secret: dockerhub-regcred
    namespace: default
    type: kubernetes.io/dockerconfigjson
    has_keys:
      - .dockerconfigjson

  - name: "Private registry access"
    secret: gcr-regcred
    namespace: production
    type: kubernetes.io/dockerconfigjson
    has_keys:
      - .dockerconfigjson
```

### Git/SSH Keys

```yaml
secrets:
  - name: "GitHub SSH key"
    secret: github-ssh
    namespace: ci-cd
    type: kubernetes.io/ssh-auth
    has_keys:
      - ssh-privatekey

  - name: "GitLab deploy key"
    secret: gitlab-deploy-key
    namespace: ci-cd
    type: kubernetes.io/ssh-auth
    has_keys:
      - ssh-privatekey
```

### Migration Cleanup

```yaml
secrets:
  - name: "New credentials installed"
    secret: app-creds-v2
    namespace: production
    state: present

  - name: "Old credentials removed"
    secret: app-creds-v1
    namespace: production
    state: absent
```

## Notes

- **Security** - Tests check existence and structure, not actual secret values
- **Base64 encoding** - Secret data is base64 encoded, but tests only verify keys exist
- **Namespace isolation** - Secrets are namespace-scoped
- **Type validation** - Type must match exactly (case-sensitive)
- **Key names** - Case-sensitive, must match exactly

## Tips

### Finding Secrets

```bash
# List all secrets in namespace
kubectl get secrets -n default

# Get secret details (values are base64 encoded)
kubectl get secret db-password -n default -o yaml

# Decode secret value
kubectl get secret db-password -n default -o jsonpath='{.data.password}' | base64 -d

# Describe secret (doesn't show values)
kubectl describe secret db-password -n default
```

### Creating Secrets for Testing

```bash
# Generic secret from literals
kubectl create secret generic db-password \
  --from-literal=username=admin \
  --from-literal=password=secret123 \
  -n default

# TLS secret from files
kubectl create secret tls tls-cert \
  --cert=path/to/tls.crt \
  --key=path/to/tls.key \
  -n ingress

# Docker registry secret
kubectl create secret docker-registry regcred \
  --docker-server=https://index.docker.io/v1/ \
  --docker-username=myuser \
  --docker-password=mypass \
  --docker-email=my@email.com \
  -n default

# SSH key secret
kubectl create secret generic git-ssh \
  --from-file=ssh-privatekey=~/.ssh/id_rsa \
  --type=kubernetes.io/ssh-auth \
  -n ci-cd
```

### Secret Data Keys

Common key naming conventions:

**TLS Secrets:**
- `tls.crt` - Certificate
- `tls.key` - Private key
- `ca.crt` - CA certificate (optional)

**Basic Auth:**
- `username` - Username
- `password` - Password

**SSH Auth:**
- `ssh-privatekey` - Private key
- `ssh-publickey` - Public key (optional)

**Docker Config:**
- `.dockerconfigjson` - Docker config JSON

**Service Account:**
- `token` - Bearer token
- `ca.crt` - CA certificate
- `namespace` - Namespace

### Security Best Practices

1. **Never commit secrets** to version control
2. **Use external secrets** - Tools like External Secrets Operator, Sealed Secrets
3. **Rotate regularly** - Test both old and new during rotation
4. **Namespace isolation** - Keep secrets in appropriate namespaces
5. **RBAC** - Restrict secret access with RBAC policies
6. **Encryption at rest** - Enable secret encryption in etcd

### Testing During Secret Rotation

```yaml
# During rotation, test both old and new secrets exist
secrets:
  - name: "Current credentials active"
    secret: api-key-v1
    namespace: production
    state: present

  - name: "New credentials ready"
    secret: api-key-v2
    namespace: production
    state: present

# After rotation complete
secrets:
  - name: "New credentials active"
    secret: api-key-v2
    namespace: production
    state: present

  - name: "Old credentials removed"
    secret: api-key-v1
    namespace: production
    state: absent
```

### Checking Secret Usage

```bash
# Find pods using a secret
kubectl get pods -n default -o json | \
  jq -r '.items[] | select(.spec.volumes[]?.secret.secretName=="db-password") | .metadata.name'

# Find which resources reference a secret
kubectl get pods,deployments,statefulsets -n default -o yaml | \
  grep -C 3 "secretName: db-password"
```
