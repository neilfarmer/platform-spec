# Kubernetes Provider

The Kubernetes provider tests Kubernetes cluster resources using `kubectl` commands executed locally.

## Overview

The Kubernetes provider enables testing of Kubernetes resources such as namespaces, pods, deployments, services, and configmaps. It executes `kubectl` commands against a Kubernetes cluster specified via kubeconfig.

## Usage

```bash
platform-spec test kubernetes spec.yaml [flags]
```

### Flags

- `--kubeconfig string` - Path to kubeconfig file (default: `~/.kube/config`)
- `--context string` - Kubernetes context to use
- `--namespace string` - Default namespace for tests
- `-o, --output string` - Output format: human, json, junit (default: "human")
- `-v, --verbose` - Verbose output

### Aliases

- `kubernetes`
- `k8s`

## Examples

```bash
# Use default kubeconfig
platform-spec test kubernetes spec.yaml

# Use custom kubeconfig
platform-spec test k8s --kubeconfig=/path/to/config spec.yaml

# Specify context and namespace
platform-spec test k8s --context=production --namespace=staging spec.yaml

# Verbose output
platform-spec test kubernetes spec.yaml --verbose
```

## Connection

The provider uses `kubectl` to communicate with the Kubernetes cluster. It supports:

- **Kubeconfig files** - Standard `~/.kube/config` or custom path via `--kubeconfig`
- **Context selection** - Use `--context` to select a specific cluster context
- **Namespace defaults** - Set default namespace via `--namespace` flag or in spec config

## Authentication

Authentication is handled through your kubeconfig file, which can include:
- Certificate-based authentication
- Token-based authentication
- Cloud provider authentication (AWS, GCP, Azure)
- OIDC authentication

The provider uses whatever authentication is configured in your kubeconfig.

## Test Types

The Kubernetes provider supports 5 test types:

1. **[Namespaces](assertions/namespaces.md)** - Validate namespace existence and labels
2. **[Pods](assertions/pods.md)** - Validate pod status, readiness, images, labels
3. **[Deployments](assertions/deployments.md)** - Validate deployment state, replicas, images
4. **[Services](assertions/services.md)** - Validate service type, ports, selectors
5. **[ConfigMaps](assertions/configmaps.md)** - Validate configmap existence and keys

## Spec Configuration

### Global Config

```yaml
config:
  fail_fast: false
  parallel: false
  timeout: 300
  kubernetes_context: "production"      # Kubernetes context to use
  kubernetes_namespace: "default"       # Default namespace for all tests
```

### Namespace Defaults

Tests inherit the namespace from:
1. Test-level `namespace` field (highest priority)
2. Spec-level `kubernetes_namespace` config
3. Command-line `--namespace` flag
4. Default: `"default"`

## YAML Structure

```yaml
version: "1.0"

metadata:
  name: "Kubernetes Infrastructure Test"
  description: "Validates production cluster resources"

config:
  kubernetes_namespace: default

tests:
  kubernetes:
    namespaces:
      - name: "Production namespace exists"
        namespace: production
        state: present
        labels:
          environment: production

    pods:
      - name: "App pod is running"
        pod: myapp-abc123
        namespace: production
        state: running
        ready: true

    deployments:
      - name: "App deployment is available"
        deployment: myapp
        namespace: production
        state: available
        replicas: 3

    services:
      - name: "App service exists"
        service: myapp-service
        namespace: production
        type: ClusterIP
        ports:
          - port: 8080
            protocol: TCP

    configmaps:
      - name: "App config exists"
        configmap: app-config
        namespace: production
        state: present
        has_keys:
          - database_url
```

## How It Works

The provider executes `kubectl get <resource> -o json` commands and parses the JSON output to validate resource properties. For example:

```bash
# To test a deployment
kubectl get deployment myapp -n production -o json

# Provider then parses the JSON to check:
# - .status.conditions for availability
# - .spec.replicas for replica count
# - .spec.template.spec.containers[].image for images
```

## Error Handling

The provider distinguishes between:

- **Resource not found** (exit code 1) → Test fails
- **kubectl error** (other exit codes) → Test errors
- **Parse error** → Test errors

## Limitations

- Requires `kubectl` installed locally
- Cannot modify resources (read-only testing)
- Pod names must be exact (no glob patterns or selectors yet)
- Does not support custom resources (CRDs) yet

## Best Practices

1. **Use specific contexts** - Avoid testing production accidentally
2. **Set namespace defaults** - Reduce repetition in spec files
3. **Test built-in resources** - Start with namespaces and services
4. **Verify before deploying** - Test manifests before applying to cluster
5. **Use in CI/CD** - Validate cluster state after deployments

## See Also

- [Examples](../../examples/kubernetes-basic.yaml)
- [Integration Tests](../../integration/kubernetes/README.md)
