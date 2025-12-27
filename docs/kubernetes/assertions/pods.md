# Pod Tests

Test Kubernetes pod status, readiness, images, and labels.

## YAML Structure

```yaml
tests:
  kubernetes:
    pods:
      - name: string              # Required: Test name
        pod: string               # Required: Pod name
        namespace: string         # Optional: Namespace (default: from config or "default")
        state: string             # Optional: running, pending, succeeded, failed, exists (default: running)
        ready: boolean            # Optional: All containers must be ready
        image: string             # Optional: Container image contains this string
        labels:                   # Optional: Expected labels
          key: value
```

## Fields

### Required

- **name** - Descriptive name for the test
- **pod** - Exact name of the pod to test

### Optional

- **namespace** - Namespace where pod exists (default: from config or `"default"`)
- **state** - Expected pod phase: `running`, `pending`, `succeeded`, `failed`, or `exists` (default: `running`)
- **ready** - If true, all containers must be ready (default: not checked)
- **image** - Container image must contain this string (default: not checked)
- **labels** - Map of labels that must exist on the pod

## Examples

### Basic Running Check

```yaml
pods:
  - name: "Nginx pod is running"
    pod: nginx-7c5ddbdf54-abc12
    namespace: default
    state: running
```

### Check Readiness

```yaml
pods:
  - name: "App pod is ready"
    pod: myapp-deployment-5d7f8b9c-xyz
    namespace: production
    state: running
    ready: true
```

### Verify Image

```yaml
pods:
  - name: "App uses correct image version"
    pod: myapp-abc123
    namespace: production
    state: running
    image: "myapp:v2.1.0"
```

### Check Labels

```yaml
pods:
  - name: "Pod has correct labels"
    pod: worker-pod-123
    namespace: default
    state: running
    labels:
      app: worker
      version: "2.0"
      tier: backend
```

### Job Completion

```yaml
pods:
  - name: "Backup job completed"
    pod: backup-job-abc123
    namespace: default
    state: succeeded
```

### Check Any State

```yaml
pods:
  - name: "Pod exists (any state)"
    pod: debug-pod
    namespace: default
    state: exists
```

## kubectl Command

The test executes:

```bash
kubectl get pod <pod-name> -n <namespace> -o json
```

And validates:
- `.status.phase` for pod state (Running, Pending, Succeeded, Failed)
- `.status.containerStatuses[].ready` for container readiness
- `.spec.containers[].image` for image validation (substring match)
- `.metadata.labels` for label validation

## Test Behavior

### States

- **running** - Pod phase must be "Running"
- **pending** - Pod phase must be "Pending"
- **succeeded** - Pod phase must be "Succeeded" (completed successfully)
- **failed** - Pod phase must be "Failed"
- **exists** - Pod exists in any state

### Ready Check

When `ready: true`:
- All containers in `.status.containerStatuses` must have `ready: true`
- If any container is not ready, test fails
- If no container statuses exist, test fails

### Image Check

When `image` is specified:
- At least one container image must **contain** the specified string
- Uses substring matching (e.g., `image: nginx` matches `nginx:1.21-alpine`)
- Case-sensitive matching

## Common Patterns

### Deployment Pods

```yaml
# Note: Pod names are generated, get actual name first:
# kubectl get pods -n default -l app=nginx

pods:
  - name: "Nginx replica 1 running"
    pod: nginx-deployment-7c5ddbdf54-abc12
    namespace: default
    state: running
    ready: true
    labels:
      app: nginx
      pod-template-hash: 7c5ddbdf54
```

### StatefulSet Pods

```yaml
pods:
  - name: "Database pod 0"
    pod: postgres-0
    namespace: default
    state: running
    ready: true
    image: postgres:14

  - name: "Database pod 1"
    pod: postgres-1
    namespace: default
    state: running
    ready: true
```

### Job Pods

```yaml
pods:
  - name: "Migration job completed"
    pod: migrate-db-12345
    namespace: default
    state: succeeded
```

### Init Container Checks

```yaml
pods:
  - name: "Pod running after init"
    pod: app-with-init-abc123
    namespace: default
    state: running
    ready: true
```

## Notes

- **Pod names are exact** - No wildcard or selector support yet
- **Get pod names first** - Use `kubectl get pods` to find exact names
- **Pod names change** - Deployment/ReplicaSet pods have random suffixes
- **Image matching** - Substring match, so `nginx` matches `nginx:1.21-alpine`
- **Ready vs Running** - A pod can be Running but not Ready
- **Namespace defaults** - Uses `kubernetes_namespace` from config if not specified

## Tips

### Finding Pod Names

```bash
# List pods by label
kubectl get pods -l app=nginx -n default

# Get specific pod
kubectl get pod nginx-abc123 -n default -o wide

# Watch pod status
kubectl get pods -w -n default
```

### Dynamic Pod Names

For deployment pods with random suffixes, consider testing the deployment instead:

```yaml
deployments:
  - name: "Nginx deployment available"
    deployment: nginx-deployment
    namespace: default
    state: available
    replicas: 3
```

This is more stable than testing individual pods.
