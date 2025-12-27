# ConfigMap Tests

Test Kubernetes configmap existence and keys.

## YAML Structure

```yaml
tests:
  kubernetes:
    configmaps:
      - name: string              # Required: Test name
        configmap: string         # Required: ConfigMap name
        namespace: string         # Optional: Namespace (default: from config or "default")
        state: string             # Optional: present or absent (default: present)
        has_keys:                 # Optional: List of keys that must exist
          - key1
          - key2
```

## Fields

### Required

- **name** - Descriptive name for the test
- **configmap** - Name of the ConfigMap to test

### Optional

- **namespace** - Namespace where ConfigMap exists (default: from config or `"default"`)
- **state** - Expected state: `present` or `absent` (default: `present`)
- **has_keys** - List of keys that must exist in `.data` (all must exist)

## Examples

### Basic Existence Check

```yaml
configmaps:
  - name: "App config exists"
    configmap: app-config
    namespace: default
    state: present
```

### Verify Keys Exist

```yaml
configmaps:
  - name: "App config has required keys"
    configmap: app-config
    namespace: default
    state: present
    has_keys:
      - database_url
      - redis_host
      - api_key
```

### Environment-Specific Config

```yaml
configmaps:
  - name: "Production config exists"
    configmap: prod-config
    namespace: production
    state: present
    has_keys:
      - APP_ENV
      - DATABASE_URL
      - CACHE_DRIVER
```

### Ensure ConfigMap Doesn't Exist

```yaml
configmaps:
  - name: "Old config should be removed"
    configmap: deprecated-config
    namespace: default
    state: absent
```

## kubectl Command

The test executes:

```bash
kubectl get configmap <configmap-name> -n <namespace> -o json
```

And validates:
- Exit code 0 = ConfigMap exists
- Exit code 1 with "not found" = ConfigMap absent
- JSON `.data` keys for key validation

## Test Behavior

### State: present

- **Pass** - ConfigMap exists and all specified keys exist in `.data`
- **Fail** - ConfigMap not found or missing required keys
- **Error** - kubectl command fails (cluster unreachable, auth error)

### State: absent

- **Pass** - ConfigMap does not exist
- **Fail** - ConfigMap exists
- **Error** - kubectl command fails

### Key Validation

When `has_keys` is specified:
- All keys must exist in `.data` section
- Key names are case-sensitive
- Only checks key existence, not values
- If any key is missing, test fails

## Common Patterns

### Application Configuration

```yaml
configmaps:
  - name: "App has required configuration"
    configmap: myapp-config
    namespace: default
    state: present
    has_keys:
      - app.properties
      - logging.conf
      - database.yml
```

### Multi-Environment Setup

```yaml
configmaps:
  - name: "Dev environment config"
    configmap: app-config
    namespace: development
    state: present
    has_keys:
      - APP_ENV
      - DEBUG

  - name: "Prod environment config"
    configmap: app-config
    namespace: production
    state: present
    has_keys:
      - APP_ENV
      - CACHE_ENABLED
```

### Feature Flags

```yaml
configmaps:
  - name: "Feature flags configured"
    configmap: feature-flags
    namespace: default
    state: present
    has_keys:
      - enable_new_ui
      - enable_beta_features
      - maintenance_mode
```

### Shared Configuration

```yaml
configmaps:
  - name: "Shared database config"
    configmap: shared-db-config
    namespace: default
    state: present
    has_keys:
      - host
      - port
      - database
```

### Configuration Files

```yaml
configmaps:
  - name: "Nginx config loaded"
    configmap: nginx-config
    namespace: default
    state: present
    has_keys:
      - nginx.conf
      - mime.types
```

## Notes

- **Keys only** - Test validates key existence, not values
- **Case sensitive** - Key names must match exactly
- **Data section** - Only checks `.data`, not `.binaryData`
- **All keys required** - If `has_keys` is specified, all listed keys must exist
- **Namespace defaults** - Uses `kubernetes_namespace` from config if not specified

## Tips

### View ConfigMap Contents

```bash
# List all configmaps
kubectl get configmaps -n default

# View configmap details
kubectl describe configmap app-config -n default

# View configmap data (YAML)
kubectl get configmap app-config -n default -o yaml

# View configmap data (JSON)
kubectl get configmap app-config -n default -o json

# Get specific key value
kubectl get configmap app-config -n default -o jsonpath='{.data.database_url}'
```

### Creating ConfigMaps

```bash
# From literal values
kubectl create configmap app-config \
  --from-literal=APP_ENV=production \
  --from-literal=DEBUG=false

# From file
kubectl create configmap app-config \
  --from-file=config.properties

# From directory
kubectl create configmap app-config \
  --from-file=./config/

# From YAML
kubectl apply -f configmap.yaml
```

### ConfigMap Usage in Pods

ConfigMaps are typically mounted as:
- **Environment variables** - Using `envFrom` or `env`
- **Volume mounts** - Files mounted in container filesystem
- **Command arguments** - Using `$(VAR_NAME)` syntax

Example pod using ConfigMap:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: myapp
spec:
  containers:
  - name: myapp
    image: myapp:latest
    envFrom:
    - configMapRef:
        name: app-config
    volumeMounts:
    - name: config
      mountPath: /etc/config
  volumes:
  - name: config
    configMap:
      name: app-config
```

### Troubleshooting

**ConfigMap not found:**
- Verify name and namespace: `kubectl get cm -A | grep app-config`
- Check spelling and case sensitivity

**Keys missing:**
- View current keys: `kubectl get cm app-config -o jsonpath='{.data}' | jq 'keys'`
- Update ConfigMap: `kubectl edit cm app-config -n default`

**Pod not seeing ConfigMap changes:**
- ConfigMap updates are eventually consistent (can take 1-2 minutes)
- Restart pods to pick up changes: `kubectl rollout restart deployment/myapp`
- Or use immutable ConfigMaps with versioned names

### Best Practices

1. **Version ConfigMaps** - Use names like `app-config-v2` for immutability
2. **Don't store secrets** - Use Secrets for sensitive data
3. **Keep it small** - ConfigMaps have 1MB size limit
4. **Use labels** - Add metadata for organization: `version: "2.1.0"`
5. **Document keys** - Maintain list of required keys in docs

### Immutable ConfigMaps

For production stability, make ConfigMaps immutable:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config-v2
data:
  APP_ENV: production
immutable: true
```

Benefits:
- Prevents accidental changes
- Better performance (no watch overhead)
- Clearer deployment process (new version = new ConfigMap)

### Size Limits

- ConfigMap size limit: **1MB**
- etcd value size limit: **1.5MB** (ConfigMap + metadata)
- For larger data, use external storage or break into multiple ConfigMaps
