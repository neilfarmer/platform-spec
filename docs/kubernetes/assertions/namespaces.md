# Namespace Tests

Test Kubernetes namespace existence and properties.

## YAML Structure

```yaml
tests:
  kubernetes:
    namespaces:
      - name: string              # Required: Test name
        namespace: string         # Required: Namespace name
        state: string             # Optional: present or absent (default: present)
        labels:                   # Optional: Expected labels (key-value pairs)
          key: value
```

## Fields

### Required

- **name** - Descriptive name for the test
- **namespace** - Name of the Kubernetes namespace to test

### Optional

- **state** - Expected state: `present` or `absent` (default: `present`)
- **labels** - Map of labels that must exist on the namespace

## Examples

### Basic Existence Check

```yaml
namespaces:
  - name: "Default namespace exists"
    namespace: default
    state: present
```

### Check System Namespaces

```yaml
namespaces:
  - name: "Kube-system namespace exists"
    namespace: kube-system
    state: present

  - name: "Kube-public namespace exists"
    namespace: kube-public
    state: present
```

### Verify Labels

```yaml
namespaces:
  - name: "Production namespace has correct labels"
    namespace: production
    state: present
    labels:
      environment: production
      team: platform
      managed-by: terraform
```

### Ensure Namespace Doesn't Exist

```yaml
namespaces:
  - name: "Old namespace should be deleted"
    namespace: deprecated-app
    state: absent
```

## kubectl Command

The test executes:

```bash
kubectl get namespace <namespace-name> -o json
```

And validates:
- Exit code 0 = namespace exists
- Exit code 1 with "not found" = namespace absent
- JSON `.metadata.labels` for label validation

## Test Behavior

### State: present

- **Pass** - Namespace exists and labels match (if specified)
- **Fail** - Namespace not found or labels don't match
- **Error** - kubectl command fails (cluster unreachable, auth error)

### State: absent

- **Pass** - Namespace does not exist
- **Fail** - Namespace exists
- **Error** - kubectl command fails

## Common Patterns

### Verify All Required Namespaces

```yaml
namespaces:
  - name: "Development namespace"
    namespace: development
    state: present

  - name: "Staging namespace"
    namespace: staging
    state: present

  - name: "Production namespace"
    namespace: production
    state: present
```

### Multi-tenant Validation

```yaml
namespaces:
  - name: "Team A namespace"
    namespace: team-a
    labels:
      team: team-a
      cost-center: "1001"

  - name: "Team B namespace"
    namespace: team-b
    labels:
      team: team-b
      cost-center: "1002"
```

### Environment Segregation

```yaml
namespaces:
  - name: "Production environment"
    namespace: prod
    labels:
      environment: production
      criticality: high
      backup: enabled

  - name: "Test environment"
    namespace: test
    labels:
      environment: test
      criticality: low
```

## Notes

- Namespace tests don't require a default namespace configuration
- Labels must match exactly (partial matching not supported)
- Label validation only works when `state: present`
- System namespaces (default, kube-system, etc.) always exist in clusters
