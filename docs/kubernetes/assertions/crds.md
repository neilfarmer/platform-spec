# CRD Tests

Test Kubernetes CustomResourceDefinition (CRD) existence.

## YAML Structure

```yaml
tests:
  kubernetes:
    crds:
      - name: string              # Required: Test name
        crd: string               # Required: CRD name
        state: string             # Optional: present, absent (default: present)
```

## Fields

### Required

- **name** - Descriptive name for the test
- **crd** - Full CRD name (e.g., `certificates.cert-manager.io`)

### Optional

- **state** - Expected state: `present` or `absent` (default: `present`)

## Examples

### Basic CRD Existence

```yaml
crds:
  - name: "cert-manager CRDs installed"
    crd: certificates.cert-manager.io
    state: present
```

### Multiple CRDs

```yaml
crds:
  - name: "Istio VirtualService CRD exists"
    crd: virtualservices.networking.istio.io

  - name: "Istio DestinationRule CRD exists"
    crd: destinationrules.networking.istio.io

  - name: "Istio Gateway CRD exists"
    crd: gateways.networking.istio.io
```

### Verify CRD Removed

```yaml
crds:
  - name: "Old monitoring CRD removed"
    crd: servicemonitors.monitoring.coreos.com
    state: absent
```

### cert-manager

```yaml
crds:
  - name: "Certificate CRD installed"
    crd: certificates.cert-manager.io

  - name: "CertificateRequest CRD installed"
    crd: certificaterequests.cert-manager.io

  - name: "Issuer CRD installed"
    crd: issuers.cert-manager.io

  - name: "ClusterIssuer CRD installed"
    crd: clusterissuers.cert-manager.io
```

### Prometheus Operator

```yaml
crds:
  - name: "ServiceMonitor CRD exists"
    crd: servicemonitors.monitoring.coreos.com

  - name: "Prometheus CRD exists"
    crd: prometheuses.monitoring.coreos.com

  - name: "PrometheusRule CRD exists"
    crd: prometheusrules.monitoring.coreos.com
```

### Argo CD

```yaml
crds:
  - name: "Application CRD exists"
    crd: applications.argoproj.io

  - name: "AppProject CRD exists"
    crd: appprojects.argoproj.io
```

## kubectl Command

The test executes:

```bash
kubectl get crd <crd-name> -o json 2>&1
```

And validates:
- Exit code 0 = CRD exists
- Exit code 1 with "not found" = CRD absent
- Other errors = test error

## Test Behavior

### State: present

When `state: present`:
- CRD must exist in cluster
- Test passes if `kubectl get crd` succeeds
- Test fails if CRD not found

### State: absent

When `state: absent`:
- CRD must NOT exist in cluster
- Test passes if CRD not found
- Test fails if CRD exists

### Error Handling

- Connection errors return test status ERROR
- Permission errors return test status ERROR
- "Not found" is expected for `state: absent`

## Common Patterns

### Operator Installation

```yaml
crds:
  - name: "cert-manager installed"
    crd: certificates.cert-manager.io

  - name: "Istio installed"
    crd: virtualservices.networking.istio.io

  - name: "Prometheus Operator installed"
    crd: servicemonitors.monitoring.coreos.com
```

### Migration Cleanup

```yaml
crds:
  - name: "New CRD version installed"
    crd: myresources.v2.example.com
    state: present

  - name: "Old CRD version removed"
    crd: myresources.v1.example.com
    state: absent
```

### Multi-Component Check

```yaml
# Verify all Istio CRDs installed
crds:
  - name: "Istio VirtualService"
    crd: virtualservices.networking.istio.io

  - name: "Istio DestinationRule"
    crd: destinationrules.networking.istio.io

  - name: "Istio ServiceEntry"
    crd: serviceentries.networking.istio.io

  - name: "Istio Gateway"
    crd: gateways.networking.istio.io

  - name: "Istio Sidecar"
    crd: sidecars.networking.istio.io
```

## Notes

- **CRD names** - Use full name with group: `<plural>.<group>`
- **Not resource instances** - Tests CRD existence, not custom resource instances
- **Operator dependencies** - Check CRDs exist before checking custom resources
- **Version handling** - CRD name doesn't include version (versions are in CRD spec)

## Tips

### Finding CRD Names

```bash
# List all CRDs
kubectl get crds

# Search for specific CRDs
kubectl get crds | grep cert-manager

# Get CRD details
kubectl get crd certificates.cert-manager.io -o yaml

# View CRD versions and schema
kubectl describe crd certificates.cert-manager.io
```

### CRD Name Format

CRD names follow the pattern: `<plural>.<group>`

Examples:
- `certificates.cert-manager.io` - Certificate resource in cert-manager.io group
- `virtualservices.networking.istio.io` - VirtualService in networking.istio.io group
- `prometheuses.monitoring.coreos.com` - Prometheus in monitoring.coreos.com group

### Checking Custom Resources

Once CRD exists, test actual custom resource instances:

```bash
# After CRD test passes, check instances
kubectl get certificates -n default
kubectl get virtualservices -n istio-system
```

### Common CRDs

**cert-manager:**
- `certificates.cert-manager.io`
- `issuers.cert-manager.io`
- `clusterissuers.cert-manager.io`

**Istio:**
- `virtualservices.networking.istio.io`
- `destinationrules.networking.istio.io`
- `gateways.networking.istio.io`

**Prometheus Operator:**
- `servicemonitors.monitoring.coreos.com`
- `prometheuses.monitoring.coreos.com`
- `prometheusrules.monitoring.coreos.com`

**Argo CD:**
- `applications.argoproj.io`
- `appprojects.argoproj.io`
