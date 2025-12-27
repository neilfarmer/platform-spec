# PersistentVolumeClaim Tests

Test Kubernetes PersistentVolumeClaim (PVC) existence, status, storage class, and capacity validation.

## YAML Structure

```yaml
tests:
  kubernetes:
    pvcs:
      - name: string              # Required: Test name
        pvc: string               # Required: PVC name
        namespace: string         # Optional: Namespace (default: "default")
        state: string             # Optional: present, absent (default: "present")
        status: string            # Optional: Expected status (Bound, Pending, Lost)
        storage_class: string     # Optional: Expected storage class name
        min_capacity: string      # Optional: Minimum capacity (e.g., "100Gi")
```

## Fields

### Required

- **name** - Descriptive name for the test
- **pvc** - PersistentVolumeClaim name

### Optional

- **namespace** - Namespace where PVC exists (default: `"default"`)
- **state** - Expected state: `present` or `absent` (default: `present`)
- **status** - Expected PVC status: `Bound`, `Pending`, or `Lost` (default: not checked)
- **storage_class** - Expected storage class name (default: not checked)
- **min_capacity** - Minimum required capacity with units (e.g., `"100Gi"`, `"1Ti"`) (default: not checked)

## Examples

### Basic PVC Existence

```yaml
pvcs:
  - name: "Database volume exists"
    pvc: postgres-data
    namespace: production
    state: present
```

### Multiple PVCs

```yaml
pvcs:
  - name: "Database PVC"
    pvc: postgres-data
    namespace: default

  - name: "File storage PVC"
    pvc: file-storage
    namespace: default

  - name: "Cache PVC"
    pvc: redis-data
    namespace: default
```

### Verify PVC Removed

```yaml
pvcs:
  - name: "Old volume removed"
    pvc: deprecated-data
    namespace: production
    state: absent
```

### PVC with Status Check

```yaml
pvcs:
  - name: "Database volume is bound"
    pvc: postgres-data
    namespace: production
    status: Bound
```

### PVC with Storage Class

```yaml
pvcs:
  - name: "Fast SSD storage"
    pvc: app-data
    namespace: production
    storage_class: fast-ssd
```

### PVC with Capacity Check

```yaml
pvcs:
  - name: "Database has sufficient storage"
    pvc: postgres-data
    namespace: production
    min_capacity: "100Gi"
```

### Complete PVC Validation

```yaml
pvcs:
  - name: "Production database volume fully validated"
    pvc: postgres-data
    namespace: production
    state: present
    status: Bound
    storage_class: ssd-storage
    min_capacity: "500Gi"
```

### Storage Tier Validation

```yaml
pvcs:
  - name: "High-performance storage"
    pvc: database-primary
    namespace: production
    storage_class: premium-ssd
    min_capacity: "1Ti"
    status: Bound

  - name: "Standard storage"
    pvc: backups
    namespace: production
    storage_class: standard-hdd
    min_capacity: "5Ti"
    status: Bound
```

## kubectl Command

The test executes:

```bash
kubectl get pvc <pvc-name> -n <namespace> -o json 2>&1
```

And validates:
- Exit code 0 = PVC exists
- Exit code 1 with "not found" = PVC absent
- `.status.phase` for status validation (Bound, Pending, Lost)
- `.spec.storageClassName` for storage class
- `.status.capacity.storage` for capacity validation

## Test Behavior

### State: present

When `state: present`:
- PVC must exist in the namespace
- Test passes if `kubectl get pvc` succeeds
- Test fails if PVC not found

### State: absent

When `state: absent`:
- PVC must NOT exist in the namespace
- Test passes if PVC not found
- Test fails if PVC exists

### Status Validation

When `status` is specified:
- PVC's `.status.phase` must match exactly
- Valid statuses: `Bound`, `Pending`, `Lost`
- **Bound** - PVC is bound to a PersistentVolume
- **Pending** - PVC is waiting for a PersistentVolume
- **Lost** - PVC lost its bound PersistentVolume
- Test fails if status doesn't match

### Storage Class Validation

When `storage_class` is specified:
- PVC's `.spec.storageClassName` must match exactly
- Case-sensitive matching
- Test fails if storage class doesn't match or is missing

### Capacity Validation

When `min_capacity` is specified:
- PVC's actual capacity (`.status.capacity.storage`) must be >= min_capacity
- Supports units: `Ki`, `Mi`, `Gi`, `Ti` (binary) and `K`, `M`, `G`, `T` (decimal)
- Test fails if capacity is less than minimum
- Test fails if PVC is not bound (capacity not available)

## Storage Size Units

### Binary Units (IEC)
- `Ki` - Kibibytes (1024 bytes)
- `Mi` - Mebibytes (1024 Ki)
- `Gi` - Gibibytes (1024 Mi)
- `Ti` - Tebibytes (1024 Gi)
- `Pi` - Pebibytes (1024 Ti)

### Decimal Units (SI)
- `K` - Kilobytes (1000 bytes)
- `M` - Megabytes (1000 K)
- `G` - Gigabytes (1000 M)
- `T` - Terabytes (1000 G)
- `P` - Petabytes (1000 T)

**Note:** Kubernetes primarily uses binary units (Gi, Ti) for storage.

## Common Patterns

### Database Storage

```yaml
pvcs:
  - name: "PostgreSQL data volume"
    pvc: postgres-data
    namespace: production
    storage_class: ssd-storage
    status: Bound
    min_capacity: "500Gi"

  - name: "MySQL data volume"
    pvc: mysql-data
    namespace: production
    storage_class: ssd-storage
    status: Bound
    min_capacity: "200Gi"
```

### Application Storage Tiers

```yaml
pvcs:
  - name: "Hot storage (SSD)"
    pvc: app-hot-data
    namespace: production
    storage_class: premium-ssd
    status: Bound
    min_capacity: "100Gi"

  - name: "Warm storage (HDD)"
    pvc: app-warm-data
    namespace: production
    storage_class: standard-hdd
    status: Bound
    min_capacity: "1Ti"

  - name: "Cold storage (Archive)"
    pvc: app-archive
    namespace: production
    storage_class: archive
    status: Bound
    min_capacity: "10Ti"
```

### Multi-Environment Storage

```yaml
pvcs:
  - name: "Production database"
    pvc: db-data
    namespace: production
    storage_class: premium-ssd
    status: Bound
    min_capacity: "1Ti"

  - name: "Staging database"
    pvc: db-data
    namespace: staging
    storage_class: standard-ssd
    status: Bound
    min_capacity: "500Gi"

  - name: "Development database"
    pvc: db-data
    namespace: development
    storage_class: standard-hdd
    status: Bound
    min_capacity: "100Gi"
```

### StatefulSet Storage

```yaml
pvcs:
  - name: "Cassandra node 0"
    pvc: data-cassandra-0
    namespace: production
    storage_class: local-ssd
    status: Bound
    min_capacity: "500Gi"

  - name: "Cassandra node 1"
    pvc: data-cassandra-1
    namespace: production
    storage_class: local-ssd
    status: Bound
    min_capacity: "500Gi"

  - name: "Cassandra node 2"
    pvc: data-cassandra-2
    namespace: production
    storage_class: local-ssd
    status: Bound
    min_capacity: "500Gi"
```

### Backup and Restore Validation

```yaml
pvcs:
  - name: "Backup volume ready"
    pvc: backup-volume
    namespace: backup
    storage_class: standard-hdd
    status: Bound
    min_capacity: "5Ti"

  - name: "Restore volume prepared"
    pvc: restore-volume
    namespace: restore
    storage_class: fast-ssd
    status: Bound
    min_capacity: "1Ti"
```

## Notes

- **Capacity units** - Use binary units (Gi, Ti) for consistency with Kubernetes
- **Status phases** - PVC status can change over time (Pending → Bound)
- **Storage class** - Must exist in cluster before PVC can be bound
- **Dynamic provisioning** - Storage classes enable automatic PV provisioning
- **Namespace scope** - PVCs are namespace-scoped resources
- **Capacity check** - Only works when PVC is Bound (capacity not available in Pending state)

## Tips

### Finding PVCs

```bash
# List all PVCs in namespace
kubectl get pvc -n default

# Get PVC details
kubectl get pvc postgres-data -n default -o yaml

# Describe PVC (shows events)
kubectl describe pvc postgres-data -n default

# List PVCs across all namespaces
kubectl get pvc -A

# Show PVC with capacity
kubectl get pvc -n default -o custom-columns=NAME:.metadata.name,STATUS:.status.phase,CAPACITY:.status.capacity.storage,STORAGECLASS:.spec.storageClassName
```

### Creating PVCs for Testing

```bash
# Create basic PVC
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: app-data
  namespace: default
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 100Gi
  storageClassName: standard-ssd
EOF

# Create PVC with specific storage class
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: postgres-data
  namespace: production
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 500Gi
  storageClassName: premium-ssd
EOF
```

### Checking PVC Configuration

```bash
# Get PVC status
kubectl get pvc postgres-data -n default -o jsonpath='{.status.phase}'

# Get storage class
kubectl get pvc postgres-data -n default -o jsonpath='{.spec.storageClassName}'

# Get capacity
kubectl get pvc postgres-data -n default -o jsonpath='{.status.capacity.storage}'

# Get bound PersistentVolume
kubectl get pvc postgres-data -n default -o jsonpath='{.spec.volumeName}'

# Get access modes
kubectl get pvc postgres-data -n default -o jsonpath='{.spec.accessModes}'
```

### Common Storage Classes

Cloud provider examples:

**AWS (EBS)**
- `gp3` - General Purpose SSD (recommended)
- `gp2` - General Purpose SSD (older)
- `io1` - Provisioned IOPS SSD
- `st1` - Throughput Optimized HDD
- `sc1` - Cold HDD

**GCP (GCE)**
- `standard-rwo` - Standard persistent disk
- `ssd-rwo` - SSD persistent disk
- `balanced-rwo` - Balanced persistent disk

**Azure**
- `managed-premium` - Premium SSD
- `managed-standard-ssd` - Standard SSD
- `managed-standard-hdd` - Standard HDD

**On-Premises**
- `local-path` - Local storage
- `nfs` - NFS storage
- `ceph-rbd` - Ceph RBD
- `longhorn` - Longhorn storage

### Troubleshooting

```bash
# Check why PVC is pending
kubectl describe pvc postgres-data -n default

# List available storage classes
kubectl get storageclass

# Check storage class details
kubectl describe storageclass premium-ssd

# Find pods using a PVC
kubectl get pods -n default -o json | \
  jq -r '.items[] | select(.spec.volumes[]?.persistentVolumeClaim.claimName=="postgres-data") | .metadata.name'

# Check PV bound to PVC
PV_NAME=$(kubectl get pvc postgres-data -n default -o jsonpath='{.spec.volumeName}')
kubectl describe pv $PV_NAME

# Check events for PVC
kubectl get events -n default --field-selector involvedObject.name=postgres-data
```

### PVC Expansion

```bash
# Check if storage class allows expansion
kubectl get storageclass premium-ssd -o jsonpath='{.allowVolumeExpansion}'

# Expand PVC (edit spec.resources.requests.storage)
kubectl edit pvc postgres-data -n default

# Monitor expansion progress
kubectl describe pvc postgres-data -n default | grep -A 5 Conditions
```

### Access Modes

- `ReadWriteOnce` (RWO) - Single node read-write
- `ReadOnlyMany` (ROX) - Multiple nodes read-only
- `ReadWriteMany` (RWX) - Multiple nodes read-write
- `ReadWriteOncePod` (RWOP) - Single pod read-write (Kubernetes 1.22+)

### Best Practices

1. **Storage class** - Always specify storage class explicitly
2. **Right-sizing** - Start with appropriate capacity to avoid frequent expansions
3. **Performance tiers** - Use SSD for databases, HDD for backups/archives
4. **Access modes** - Choose correct access mode for workload (most use RWO)
5. **Reclaim policy** - Understand storage class reclaim policy (Retain vs Delete)
6. **Snapshots** - Use VolumeSnapshots for backups before upgrades
7. **Monitoring** - Monitor PVC usage to prevent running out of space
8. **Namespace organization** - Group related PVCs in same namespace
9. **Labels** - Use labels to organize and identify PVCs (app, tier, environment)
10. **Backup strategy** - Have a backup plan for critical persistent data

### Monitoring PVC Usage

```bash
# Check PVC disk usage (requires metrics-server or prometheus)
kubectl top pvc -n default

# Manual check using a debug pod
kubectl run -it --rm debug --image=busybox --restart=Never -- df -h /data \
  --overrides='{"spec":{"volumes":[{"name":"data","persistentVolumeClaim":{"claimName":"postgres-data"}}],"containers":[{"name":"debug","image":"busybox","volumeMounts":[{"name":"data","mountPath":"/data"}]}]}}'
```

### Migration and Cleanup

```yaml
# Verify new PVC ready before removing old
pvcs:
  - name: "New storage provisioned"
    pvc: postgres-data-v2
    namespace: production
    storage_class: premium-ssd
    status: Bound
    min_capacity: "1Ti"

  - name: "Old storage still present"
    pvc: postgres-data-v1
    namespace: production
    state: present

# After migration complete
pvcs:
  - name: "New storage active"
    pvc: postgres-data-v2
    namespace: production
    status: Bound

  - name: "Old storage removed"
    pvc: postgres-data-v1
    namespace: production
    state: absent
```
