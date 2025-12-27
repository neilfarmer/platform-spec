# Helm Tests

Test Helm release status and pod health.

## YAML Structure

```yaml
tests:
  kubernetes:
    helm:
      - name: string              # Required: Test name
        release: string           # Required: Helm release name
        namespace: string         # Optional: Namespace (default: "default")
        state: string             # Optional: deployed, failed, etc. (default: "deployed")
        all_pods_ready: boolean   # Optional: Check all release pods are ready
```

## Fields

### Required

- **name** - Descriptive name for the test
- **release** - Helm release name

### Optional

- **namespace** - Namespace where release is installed (default: `"default"`)
- **state** - Expected release state (default: `"deployed"`)
  - Valid states: `deployed`, `uninstalling`, `superseded`, `failed`, `uninstalled`, `pending-install`, `pending-upgrade`, `pending-rollback`
- **all_pods_ready** - If true, verify all pods from release are ready and healthy (default: false)

## Examples

### Basic Release Check

```yaml
helm:
  - name: "Prometheus deployed"
    release: prometheus
    namespace: monitoring
    state: deployed
```

### Release with Pod Health

```yaml
helm:
  - name: "Grafana fully operational"
    release: grafana
    namespace: monitoring
    state: deployed
    all_pods_ready: true
```

### Multiple Releases

```yaml
helm:
  - name: "Nginx ingress deployed"
    release: nginx-ingress
    namespace: ingress-nginx
    state: deployed

  - name: "cert-manager deployed"
    release: cert-manager
    namespace: cert-manager
    state: deployed
    all_pods_ready: true
```

### Check Failed Release

```yaml
helm:
  - name: "Verify rollback worked"
    release: myapp
    namespace: production
    state: deployed
```

### Monitoring Stack

```yaml
helm:
  - name: "Prometheus deployed and healthy"
    release: prometheus
    namespace: monitoring
    state: deployed
    all_pods_ready: true

  - name: "Grafana deployed and healthy"
    release: grafana
    namespace: monitoring
    state: deployed
    all_pods_ready: true

  - name: "Alertmanager deployed and healthy"
    release: alertmanager
    namespace: monitoring
    state: deployed
    all_pods_ready: true
```

### Service Mesh

```yaml
helm:
  - name: "Istio base deployed"
    release: istio-base
    namespace: istio-system
    state: deployed

  - name: "Istiod deployed and ready"
    release: istiod
    namespace: istio-system
    state: deployed
    all_pods_ready: true
```

## helm/kubectl Commands

The test executes:

```bash
# Get release status
helm list -n <namespace> -o json

# When all_pods_ready is true, also checks:
kubectl get pods -n <namespace> -l app.kubernetes.io/instance=<release> -o json
```

And validates:
- Release exists in `helm list` output
- `.status` matches expected state
- When `all_pods_ready: true`:
  - All pods have phase = Running
  - All containers are ready
  - No pods in CrashLoopBackOff, ImagePullBackOff, or ErrImagePull

## Test Behavior

### State Validation

The release status must match the expected state:
- **deployed** - Successfully deployed and active
- **failed** - Deployment or upgrade failed
- **pending-install** - Installation in progress
- **pending-upgrade** - Upgrade in progress
- **uninstalling** - Being removed
- **uninstalled** - Removed but tracked in history
- **superseded** - Replaced by newer release
- **pending-rollback** - Rollback in progress

### Pod Health Check

When `all_pods_ready: true`:
1. Finds all pods with label `app.kubernetes.io/instance=<release-name>`
2. Checks each pod:
   - Phase must be "Running"
   - All containers must have `ready: true`
   - No waiting containers with CrashLoopBackOff
   - No waiting containers with ImagePullBackOff or ErrImagePull
3. Fails if:
   - No pods found for release
   - Any pod not running
   - Any container not ready
   - Any pod in crashloop or image pull error

### Error Detection

Automatically detects common failure states:
- **CrashLoopBackOff** - Container repeatedly crashing
- **ImagePullBackOff** - Cannot pull container image
- **ErrImagePull** - Image pull error
- **Pending** - Pod cannot be scheduled

## Common Patterns

### Basic Deployment Verification

```yaml
helm:
  - name: "App deployed"
    release: myapp
    namespace: production
    state: deployed
```

### Full Health Check

```yaml
helm:
  - name: "App deployed and all pods healthy"
    release: myapp
    namespace: production
    state: deployed
    all_pods_ready: true
```

### Monitoring Multiple Releases

```yaml
helm:
  - name: "Ingress controller healthy"
    release: nginx-ingress
    namespace: ingress-nginx
    all_pods_ready: true

  - name: "Database deployed"
    release: postgresql
    namespace: database
    state: deployed

  - name: "Application stack healthy"
    release: myapp
    namespace: production
    all_pods_ready: true
```

### Upgrade Verification

```yaml
# After helm upgrade
helm:
  - name: "Upgrade completed successfully"
    release: myapp
    namespace: production
    state: deployed
    all_pods_ready: true
```

## Notes

- **Release names** - Exact match required
- **Namespace** - Must match where release was installed
- **Pod labels** - Uses standard `app.kubernetes.io/instance` label
- **Health timeout** - Checks current state, doesn't wait for pods to become ready
- **Multiple namespaces** - Each release test can specify different namespace

## Tips

### Finding Helm Releases

```bash
# List all releases
helm list -A

# List releases in specific namespace
helm list -n monitoring

# Get release details
helm status prometheus -n monitoring

# View release history
helm history prometheus -n monitoring
```

### Checking Pod Health

```bash
# Find pods for a release
kubectl get pods -l app.kubernetes.io/instance=prometheus -n monitoring

# Check pod status
kubectl describe pod <pod-name> -n monitoring

# View pod logs
kubectl logs <pod-name> -n monitoring

# Check crashloop pods
kubectl get pods --field-selector=status.phase=Running -n monitoring
```

### Common Release States

**Normal states:**
- `deployed` - Healthy, running release
- `superseded` - Old version, replaced by upgrade

**Problem states:**
- `failed` - Deployment/upgrade failed
- `pending-install` - Stuck installing
- `pending-upgrade` - Stuck upgrading

**Transitional states:**
- `pending-rollback` - Rolling back
- `uninstalling` - Being removed

### Pod Health vs Release Status

A release can be `deployed` but have unhealthy pods:
- Helm considers release deployed after applying manifests
- Pods may still be starting, crashing, or pulling images
- Use `all_pods_ready: true` to verify actual pod health

### Debugging Failed Checks

If `all_pods_ready: true` fails:

```bash
# Check pod status
kubectl get pods -l app.kubernetes.io/instance=<release> -n <namespace>

# Describe problematic pods
kubectl describe pod <pod-name> -n <namespace>

# Check events
kubectl get events -n <namespace> --sort-by='.lastTimestamp'

# View container logs
kubectl logs <pod-name> -n <namespace> --all-containers
```

### Label Selectors

Helm releases use `app.kubernetes.io/instance` label by default. If your chart uses custom labels, the pod check may not find pods. Verify with:

```bash
kubectl get pods -n <namespace> --show-labels
```
