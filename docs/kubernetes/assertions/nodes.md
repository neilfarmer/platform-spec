# Node Tests

Test Kubernetes node count, health, readiness, and version requirements.

## YAML Structure

```yaml
tests:
  kubernetes:
    nodes:
      - name: string              # Required: Test name
        count: integer            # Optional: Exact node count
        min_count: integer        # Optional: Minimum node count
        min_ready: integer        # Optional: Minimum ready nodes
        min_version: string       # Optional: Minimum kubelet version (e.g., "1.28.0")
        labels:                   # Optional: Label selector
          key: value
```

## Fields

### Required

- **name** - Descriptive name for the test

### Optional

- **count** - Exact number of nodes expected (default: not checked)
- **min_count** - Minimum number of nodes (default: not checked)
- **min_ready** - Minimum number of ready nodes (default: not checked)
- **min_version** - Minimum kubelet version required on all nodes (default: not checked)
- **labels** - Map of labels to filter nodes by

## Examples

### Basic Node Count

```yaml
nodes:
  - name: "Cluster has exactly 3 nodes"
    count: 3
```

### Minimum Nodes

```yaml
nodes:
  - name: "At least 5 nodes available"
    min_count: 5
```

### Ready Nodes

```yaml
nodes:
  - name: "At least 3 nodes ready"
    min_ready: 3
```

### Version Check

```yaml
nodes:
  - name: "All nodes running v1.28+"
    min_version: "1.28.0"
```

### Combined Checks

```yaml
nodes:
  - name: "Production cluster health"
    min_count: 3
    min_ready: 3
    min_version: "1.27.0"
```

### Worker Nodes

```yaml
nodes:
  - name: "Worker nodes available"
    min_count: 2
    min_ready: 2
    labels:
      node-role.kubernetes.io/worker: ""
```

### GPU Nodes

```yaml
nodes:
  - name: "GPU nodes ready"
    min_ready: 2
    labels:
      accelerator: nvidia-tesla-v100
```

## kubectl Command

The test executes:

```bash
kubectl get nodes -o json
```

And validates:
- Total node count against `count` or `min_count`
- `.status.conditions` for Ready status (type=Ready, status=True)
- `.status.nodeInfo.kubeletVersion` for version validation
- `.metadata.labels` for label filtering

## Test Behavior

### Count Validation

When `count` is specified:
- Exact match required
- Fails if node count doesn't match exactly
- Applied after label filtering

When `min_count` is specified:
- Must have at least this many nodes
- Fails if fewer nodes exist
- Applied after label filtering

### Ready Check

When `min_ready` is specified:
- Counts nodes with Ready condition = True
- Only counts nodes that pass label filtering
- Node is ready when condition type="Ready" and status="True"

### Version Validation

When `min_version` is specified:
- All nodes must have kubelet version >= min_version
- Uses basic semantic version comparison
- Checks all nodes (after label filtering)
- Fails if any node has older version

### Label Filtering

When `labels` are specified:
- Only nodes matching ALL labels are counted
- Applied before count/ready/version checks
- Empty string value matches label presence

## Common Patterns

### Basic Cluster Health

```yaml
nodes:
  - name: "Cluster has minimum capacity"
    min_count: 3
    min_ready: 3
```

### Master Nodes

```yaml
nodes:
  - name: "Control plane nodes healthy"
    count: 3
    min_ready: 3
    labels:
      node-role.kubernetes.io/control-plane: ""
```

### Node Pool Validation

```yaml
nodes:
  - name: "General workload pool"
    min_count: 5
    min_ready: 4
    labels:
      pool: general

  - name: "High memory pool"
    min_count: 2
    min_ready: 2
    labels:
      pool: highmem
```

### Version Compliance

```yaml
nodes:
  - name: "All nodes on supported version"
    min_version: "1.27.0"

  - name: "Worker nodes upgraded"
    min_version: "1.28.0"
    labels:
      node-role.kubernetes.io/worker: ""
```

## Notes

- **Version format** - Use semantic version format: `"1.28.0"`, `"1.27.3"`, etc.
- **Label matching** - All labels must match (AND operation)
- **Ready condition** - Based on Kubernetes Ready condition, not just Running
- **Count vs min_count** - Use `count` for exact match, `min_count` for minimum threshold
- **Filtering order** - Labels filter first, then count/ready/version checks apply

## Tips

### Finding Node Information

```bash
# List all nodes
kubectl get nodes

# Get node details
kubectl get nodes -o wide

# Check node labels
kubectl get nodes --show-labels

# Describe specific node
kubectl describe node node-name

# Get kubelet versions
kubectl get nodes -o jsonpath='{.items[*].status.nodeInfo.kubeletVersion}'
```

### Node Labels

Common node labels:
- `node-role.kubernetes.io/control-plane` - Master/control plane nodes
- `node-role.kubernetes.io/worker` - Worker nodes
- `kubernetes.io/hostname` - Node hostname
- `kubernetes.io/arch` - Architecture (amd64, arm64)
- `topology.kubernetes.io/zone` - Availability zone
- `node.kubernetes.io/instance-type` - Cloud instance type

### Readiness vs Running

A node can be in Running state but not Ready:
- Node is communicating with control plane
- But kubelet reports NotReady
- Typically due to resource pressure or failing health checks

Always use `min_ready` for production health checks, not just `min_count`.
