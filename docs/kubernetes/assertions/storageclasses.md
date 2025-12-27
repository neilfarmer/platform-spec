# StorageClass Tests

Test Kubernetes StorageClass existence.

## YAML Structure

```yaml
tests:
  kubernetes:
    storageclasses:
      - name: string              # Required: Test name
        storageclass: string      # Required: StorageClass name
        state: string             # Optional: present, absent (default: present)
```

## Fields

### Required

- **name** - Descriptive name for the test
- **storageclass** - StorageClass name

### Optional

- **state** - Expected state: `present` or `absent` (default: `present`)

## Examples

### Basic StorageClass Check

```yaml
storageclasses:
  - name: "Fast SSD storage available"
    storageclass: fast-ssd
    state: present
```

### Multiple StorageClasses

```yaml
storageclasses:
  - name: "Standard storage exists"
    storageclass: standard

  - name: "Fast SSD storage exists"
    storageclass: fast-ssd

  - name: "Archive storage exists"
    storageclass: slow-hdd
```

### Verify Removal

```yaml
storageclasses:
  - name: "Old storage class removed"
    storageclass: deprecated-storage
    state: absent
```

### AWS EKS

```yaml
storageclasses:
  - name: "GP2 storage class available"
    storageclass: gp2

  - name: "GP3 storage class available"
    storageclass: gp3
```

### GCP GKE

```yaml
storageclasses:
  - name: "Standard storage available"
    storageclass: standard

  - name: "SSD storage available"
    storageclass: standard-rwo
```

### Azure AKS

```yaml
storageclasses:
  - name: "Default storage class"
    storageclass: default

  - name: "Managed premium storage"
    storageclass: managed-premium
```

### k3s / Local Development

```yaml
storageclasses:
  - name: "Local path storage available"
    storageclass: local-path
```

### OpenEBS

```yaml
storageclasses:
  - name: "OpenEBS hostpath storage"
    storageclass: openebs-hostpath

  - name: "OpenEBS Jiva storage"
    storageclass: openebs-jiva-default
```

## kubectl Command

The test executes:

```bash
kubectl get storageclass <storageclass-name> -o json 2>&1
```

And validates:
- Exit code 0 = StorageClass exists
- Exit code 1 with "not found" = StorageClass absent
- Other errors = test error

## Test Behavior

### State: present

When `state: present`:
- StorageClass must exist in cluster
- Test passes if `kubectl get storageclass` succeeds
- Test fails if StorageClass not found

### State: absent

When `state: absent`:
- StorageClass must NOT exist in cluster
- Test passes if StorageClass not found
- Test fails if StorageClass exists

### Error Handling

- Connection errors return test status ERROR
- Permission errors return test status ERROR
- "Not found" is expected for `state: absent`

## Common Patterns

### Cloud Provider Setup

```yaml
# AWS
storageclasses:
  - name: "EBS GP2 available"
    storageclass: gp2
  - name: "EBS GP3 available"
    storageclass: gp3

# GCP
storageclasses:
  - name: "GCE PD available"
    storageclass: standard
  - name: "GCE SSD available"
    storageclass: standard-rwo

# Azure
storageclasses:
  - name: "Azure Disk available"
    storageclass: default
  - name: "Premium SSD available"
    storageclass: managed-premium
```

### Multi-Tier Storage

```yaml
storageclasses:
  - name: "Performance tier - NVMe"
    storageclass: nvme-ssd

  - name: "Standard tier - SSD"
    storageclass: ssd

  - name: "Economy tier - HDD"
    storageclass: hdd

  - name: "Archive tier removed"
    storageclass: tape-archive
    state: absent
```

### CSI Driver Validation

```yaml
storageclasses:
  - name: "Ceph RBD storage class"
    storageclass: ceph-rbd

  - name: "NFS storage class"
    storageclass: nfs-client

  - name: "Longhorn storage class"
    storageclass: longhorn
```

### Migration Validation

```yaml
storageclasses:
  - name: "New storage provisioner installed"
    storageclass: fast-ssd-v2
    state: present

  - name: "Old storage provisioner removed"
    storageclass: fast-ssd-v1
    state: absent
```

## Notes

- **StorageClass names** - Exact match required
- **Default class** - Usually named `default` or `standard`, but varies by provider
- **Provisioner** - StorageClass defines the provisioner (CSI driver)
- **Not PVC check** - Tests StorageClass existence, not PersistentVolumeClaims

## Tips

### Finding StorageClasses

```bash
# List all storage classes
kubectl get storageclass
kubectl get sc  # Short form

# Get details
kubectl describe storageclass standard

# View YAML
kubectl get storageclass standard -o yaml

# Check default storage class
kubectl get storageclass -o jsonpath='{.items[?(@.metadata.annotations.storageclass\.kubernetes\.io/is-default-class=="true")].metadata.name}'
```

### Common StorageClass Names by Provider

**AWS EKS:**
- `gp2` - General purpose SSD (older)
- `gp3` - General purpose SSD (newer, recommended)
- `io1` - Provisioned IOPS SSD
- `st1` - Throughput optimized HDD
- `sc1` - Cold HDD

**GCP GKE:**
- `standard` - Standard persistent disk
- `standard-rwo` - Standard RWO persistent disk
- `premium-rwo` - SSD persistent disk

**Azure AKS:**
- `default` - Azure Disk (Standard HDD)
- `managed-premium` - Premium SSD
- `azurefile` - Azure Files
- `azurefile-premium` - Premium Azure Files

**Local/Development:**
- `local-path` - k3s local path provisioner
- `standard` - Kind default
- `hostpath` - Minikube default

**CSI Drivers:**
- `ceph-rbd` - Ceph block storage
- `cephfs` - Ceph filesystem
- `nfs-client` - NFS provisioner
- `longhorn` - Longhorn distributed storage
- `openebs-hostpath` - OpenEBS local storage

### Checking Storage Provisioners

```bash
# View provisioner for storage class
kubectl get storageclass -o jsonpath='{.items[*].provisioner}'

# Check if CSI driver pods running
kubectl get pods -n kube-system | grep csi
```

### Testing PersistentVolumes

After StorageClass exists, test actual volumes:

```bash
# List PVCs using storage class
kubectl get pvc --all-namespaces -o jsonpath='{range .items[?(@.spec.storageClassName=="fast-ssd")]}{.metadata.namespace}/{.metadata.name}{"\n"}{end}'

# Check PV provisioning
kubectl get pv
```

### Default StorageClass

To check which is the default:

```bash
kubectl get storageclass -o jsonpath='{.items[?(@.metadata.annotations.storageclass\.kubernetes\.io/is-default-class=="true")].metadata.name}'
```

Only one StorageClass should be marked as default. Multiple defaults can cause issues.
