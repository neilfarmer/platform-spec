# Deployment Tests

Test Kubernetes deployment status, replicas, and images.

## YAML Structure

```yaml
tests:
  kubernetes:
    deployments:
      - name: string              # Required: Test name
        deployment: string        # Required: Deployment name
        namespace: string         # Optional: Namespace (default: from config or "default")
        state: string             # Optional: available, progressing, exists (default: available)
        replicas: integer         # Optional: Desired replica count
        ready_replicas: integer   # Optional: Ready replica count
        image: string             # Optional: Container image contains this string
```

## Fields

### Required

- **name** - Descriptive name for the test
- **deployment** - Name of the deployment to test

### Optional

- **namespace** - Namespace where deployment exists (default: from config or `"default"`)
- **state** - Expected state: `available`, `progressing`, or `exists` (default: `available`)
- **replicas** - Expected value of `.spec.replicas` (default: not checked)
- **ready_replicas** - Expected value of `.status.readyReplicas` (default: not checked)
- **image** - Container image must contain this string (default: not checked)

## Examples

### Basic Availability Check

```yaml
deployments:
  - name: "Nginx deployment is available"
    deployment: nginx-deployment
    namespace: default
    state: available
```

### Check Replica Count

```yaml
deployments:
  - name: "App has 3 replicas"
    deployment: myapp
    namespace: production
    state: available
    replicas: 3
    ready_replicas: 3
```

### Verify Image Version

```yaml
deployments:
  - name: "App uses v2.1.0"
    deployment: myapp
    namespace: production
    state: available
    image: "myapp:v2.1.0"
```

### Rolling Update in Progress

```yaml
deployments:
  - name: "Deployment is updating"
    deployment: myapp
    namespace: production
    state: progressing
```

### Any State

```yaml
deployments:
  - name: "Deployment exists"
    deployment: old-app
    namespace: default
    state: exists
```

## kubectl Command

The test executes:

```bash
kubectl get deployment <deployment-name> -n <namespace> -o json
```

And validates:
- `.status.conditions` for availability and progressing state
- `.spec.replicas` for desired replica count
- `.status.readyReplicas` for ready replica count
- `.spec.template.spec.containers[].image` for image validation

## Test Behavior

### States

**available**
- `.status.conditions` must contain type=`Available` with status=`True`
- Test passes if deployment is fully available
- All replicas are ready and available

**progressing**
- `.status.conditions` must contain type=`Progressing` with status=`True`
- Indicates deployment is rolling out or updating
- May still be available while progressing

**exists**
- Deployment resource exists (any state)
- No condition checking

### Replica Checks

**replicas** (desired count):
- Checks `.spec.replicas`
- Must match exactly
- Only checked if `replicas > 0`

**ready_replicas** (ready count):
- Checks `.status.readyReplicas`
- Must match exactly
- Only checked if `ready_replicas > 0`

### Image Check

- Checks all containers in `.spec.template.spec.containers[].image`
- At least one image must **contain** the specified string
- Substring matching (e.g., `image: nginx` matches `nginx:1.21-alpine`)

## Common Patterns

### Production Validation

```yaml
deployments:
  - name: "Web app is available with 5 replicas"
    deployment: web-app
    namespace: production
    state: available
    replicas: 5
    ready_replicas: 5
    image: "web-app:v3.2.1"

  - name: "API is available with 3 replicas"
    deployment: api-server
    namespace: production
    state: available
    replicas: 3
    ready_replicas: 3
```

### High Availability

```yaml
deployments:
  - name: "HA deployment has minimum replicas"
    deployment: critical-service
    namespace: production
    state: available
    replicas: 3
    ready_replicas: 3
```

### Blue-Green Deployment

```yaml
deployments:
  - name: "Green deployment ready"
    deployment: myapp-green
    namespace: production
    state: available
    replicas: 3
    ready_replicas: 3

  - name: "Blue deployment still exists"
    deployment: myapp-blue
    namespace: production
    state: exists
```

### Canary Deployment

```yaml
deployments:
  - name: "Stable deployment at 90%"
    deployment: myapp-stable
    namespace: production
    state: available
    replicas: 9
    ready_replicas: 9

  - name: "Canary deployment at 10%"
    deployment: myapp-canary
    namespace: production
    state: available
    replicas: 1
    ready_replicas: 1
```

## Notes

- **Available vs Ready** - Available means deployment rollout is complete, ready means pods are ready
- **Replicas** - If not specified, any replica count is acceptable
- **Image matching** - Substring match on any container image
- **State** - `available` is most common for production validation
- **Namespace defaults** - Uses `kubernetes_namespace` from config if not specified

## Tips

### Check Deployment Status

```bash
# View deployment
kubectl get deployment myapp -n production

# Check replica status
kubectl get deployment myapp -n production -o jsonpath='{.status}'

# View conditions
kubectl describe deployment myapp -n production
```

### Common States

- **Available=True, Progressing=True** - Healthy and rolling out
- **Available=True, Progressing=False** - Healthy and stable
- **Available=False, Progressing=True** - Not available, rolling out
- **Available=False, Progressing=False** - Failed or degraded

### Replica Mismatch

If `ready_replicas` < `replicas`:
- Pods may be starting
- Pods may be failing readiness checks
- Insufficient cluster resources
- Image pull errors

Use `kubectl describe deployment` to investigate.
