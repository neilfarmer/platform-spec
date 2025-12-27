# StatefulSet Tests

Test Kubernetes StatefulSet existence, availability, and replica validation.

## YAML Structure

```yaml
tests:
  kubernetes:
    statefulsets:
      - name: string              # Required: Test name
        statefulset: string       # Required: StatefulSet name
        namespace: string         # Optional: Namespace (default: "default")
        state: string             # Optional: available, exists (default: "available")
        replicas: integer         # Optional: Expected replica count
        ready_replicas: integer   # Optional: Expected ready replica count
```

## Fields

### Required

- **name** - Descriptive name for the test
- **statefulset** - StatefulSet name

### Optional

- **namespace** - Namespace where StatefulSet exists (default: `"default"`)
- **state** - Expected state: `available` or `exists` (default: `available`)
  - `available` - All replicas are ready
  - `exists` - StatefulSet exists in any state
- **replicas** - Expected total replica count (default: not checked)
- **ready_replicas** - Expected ready replica count (default: not checked)

## Examples

### Basic StatefulSet Availability

```yaml
statefulsets:
  - name: "Database is available"
    statefulset: postgres
    namespace: production
    state: available
```

### Multiple StatefulSets

```yaml
statefulsets:
  - name: "Postgres cluster"
    statefulset: postgres
    namespace: default

  - name: "Redis cluster"
    statefulset: redis
    namespace: default

  - name: "Cassandra cluster"
    statefulset: cassandra
    namespace: data
```

### StatefulSet Exists (Any State)

```yaml
statefulsets:
  - name: "Database exists"
    statefulset: postgres
    namespace: production
    state: exists
```

### StatefulSet with Replica Count

```yaml
statefulsets:
  - name: "Postgres has 3 replicas"
    statefulset: postgres
    namespace: production
    state: available
    replicas: 3
```

### StatefulSet with Ready Replica Count

```yaml
statefulsets:
  - name: "At least 2 replicas ready"
    statefulset: postgres
    namespace: production
    state: exists
    ready_replicas: 2
```

### Complete StatefulSet Validation

```yaml
statefulsets:
  - name: "Production database fully operational"
    statefulset: postgres
    namespace: production
    state: available
    replicas: 3
    ready_replicas: 3
```

### Cluster Validation

```yaml
statefulsets:
  - name: "Cassandra cluster available"
    statefulset: cassandra
    namespace: data
    state: available
    replicas: 5

  - name: "Redis cluster available"
    statefulset: redis
    namespace: cache
    state: available
    replicas: 3

  - name: "Kafka cluster available"
    statefulset: kafka
    namespace: streaming
    state: available
    replicas: 3
```

### Partial Availability Check

```yaml
statefulsets:
  - name: "Database partially ready during rollout"
    statefulset: postgres
    namespace: production
    state: exists
    replicas: 5
    ready_replicas: 3
```

## kubectl Command

The test executes:

```bash
kubectl get statefulset <statefulset-name> -n <namespace> -o json 2>&1
```

And validates:
- Exit code 0 = StatefulSet exists
- Exit code 1 with "not found" = StatefulSet absent
- `.spec.replicas` for desired replica count
- `.status.readyReplicas` for ready replica count
- `.status.currentReplicas` for current replica count

## Test Behavior

### State: available

When `state: available`:
- StatefulSet must exist
- All desired replicas must be ready
- Checks that `readyReplicas == spec.replicas`
- Test fails if any replica is not ready
- Useful for production readiness checks

### State: exists

When `state: exists`:
- StatefulSet must exist
- Any state is acceptable (including partially ready)
- Does not require all replicas to be ready
- Useful for checking StatefulSet presence during rollouts

### Replica Count Validation

When `replicas` is specified:
- StatefulSet's desired replica count (`.spec.replicas`) must match exactly
- Validates cluster size configuration
- Test fails if replica count doesn't match

### Ready Replica Validation

When `ready_replicas` is specified:
- StatefulSet's ready replica count must match exactly
- Checks `.status.readyReplicas`
- Useful for partial availability checks
- Test fails if ready count doesn't match

## Common Patterns

### Database Clusters

```yaml
statefulsets:
  - name: "PostgreSQL cluster available"
    statefulset: postgres
    namespace: production
    state: available
    replicas: 3

  - name: "MySQL cluster available"
    statefulset: mysql
    namespace: production
    state: available
    replicas: 3

  - name: "MongoDB replica set available"
    statefulset: mongodb
    namespace: production
    state: available
    replicas: 5
```

### Cache Clusters

```yaml
statefulsets:
  - name: "Redis cluster"
    statefulset: redis
    namespace: cache
    state: available
    replicas: 3

  - name: "Memcached cluster"
    statefulset: memcached
    namespace: cache
    state: available
    replicas: 3
```

### Message Queue Clusters

```yaml
statefulsets:
  - name: "Kafka cluster"
    statefulset: kafka
    namespace: streaming
    state: available
    replicas: 3

  - name: "RabbitMQ cluster"
    statefulset: rabbitmq
    namespace: messaging
    state: available
    replicas: 3

  - name: "NATS cluster"
    statefulset: nats
    namespace: messaging
    state: available
    replicas: 3
```

### NoSQL Databases

```yaml
statefulsets:
  - name: "Cassandra cluster"
    statefulset: cassandra
    namespace: data
    state: available
    replicas: 5

  - name: "Elasticsearch cluster"
    statefulset: elasticsearch
    namespace: logging
    state: available
    replicas: 3

  - name: "Consul cluster"
    statefulset: consul
    namespace: service-mesh
    state: available
    replicas: 5
```

### High Availability Setup

```yaml
statefulsets:
  - name: "Primary database cluster"
    statefulset: postgres-primary
    namespace: production
    state: available
    replicas: 3
    ready_replicas: 3

  - name: "Read replica cluster"
    statefulset: postgres-replica
    namespace: production
    state: available
    replicas: 5
    ready_replicas: 5
```

### Multi-Environment Clusters

```yaml
statefulsets:
  - name: "Production database"
    statefulset: postgres
    namespace: production
    state: available
    replicas: 5

  - name: "Staging database"
    statefulset: postgres
    namespace: staging
    state: available
    replicas: 3

  - name: "Development database"
    statefulset: postgres
    namespace: development
    state: available
    replicas: 1
```

### Rollout Monitoring

```yaml
# Before rollout
statefulsets:
  - name: "All replicas ready before update"
    statefulset: postgres
    namespace: production
    state: available
    replicas: 5
    ready_replicas: 5

# During rollout
statefulsets:
  - name: "Minimum replicas available during rollout"
    statefulset: postgres
    namespace: production
    state: exists
    replicas: 5
    ready_replicas: 3

# After rollout
statefulsets:
  - name: "All replicas ready after update"
    statefulset: postgres
    namespace: production
    state: available
    replicas: 5
    ready_replicas: 5
```

## Notes

- **Ordered deployment** - StatefulSets deploy pods in order (pod-0, pod-1, pod-2, etc.)
- **Stable network identity** - Each pod has a persistent hostname
- **Stable storage** - Each pod can have its own PersistentVolumeClaim
- **Graceful updates** - Rolling updates happen in reverse order
- **Namespace scope** - StatefulSets are namespace-scoped resources
- **Available vs Exists** - Use `available` for production checks, `exists` for rollouts

## Tips

### Finding StatefulSets

```bash
# List all StatefulSets in namespace
kubectl get statefulset -n default

# Get StatefulSet details
kubectl get statefulset postgres -n default -o yaml

# Describe StatefulSet
kubectl describe statefulset postgres -n default

# List StatefulSets across all namespaces
kubectl get statefulset -A

# Show StatefulSet with replicas
kubectl get statefulset -n default -o custom-columns=NAME:.metadata.name,DESIRED:.spec.replicas,CURRENT:.status.currentReplicas,READY:.status.readyReplicas
```

### Creating StatefulSets for Testing

```bash
# Create basic StatefulSet
cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: default
spec:
  serviceName: postgres
  replicas: 3
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:14
        ports:
        - containerPort: 5432
        env:
        - name: POSTGRES_PASSWORD
          value: example
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 10Gi
EOF
```

### Checking StatefulSet Status

```bash
# Get replica counts
kubectl get statefulset postgres -n default -o jsonpath='{.spec.replicas}'
kubectl get statefulset postgres -n default -o jsonpath='{.status.readyReplicas}'
kubectl get statefulset postgres -n default -o jsonpath='{.status.currentReplicas}'

# Check if all replicas are ready
DESIRED=$(kubectl get statefulset postgres -n default -o jsonpath='{.spec.replicas}')
READY=$(kubectl get statefulset postgres -n default -o jsonpath='{.status.readyReplicas}')
if [ "$DESIRED" -eq "$READY" ]; then echo "Available"; else echo "Not Available"; fi

# List pods for StatefulSet
kubectl get pods -n default -l app=postgres

# Check pod readiness
kubectl get pods -n default -l app=postgres -o jsonpath='{.items[*].status.conditions[?(@.type=="Ready")].status}'
```

### Managing StatefulSets

```bash
# Scale StatefulSet
kubectl scale statefulset postgres -n default --replicas=5

# Update StatefulSet image (rolling update)
kubectl set image statefulset/postgres postgres=postgres:15 -n default

# Check rollout status
kubectl rollout status statefulset/postgres -n default

# Rollback StatefulSet
kubectl rollout undo statefulset/postgres -n default

# Restart StatefulSet pods
kubectl rollout restart statefulset/postgres -n default
```

### Troubleshooting

```bash
# Check why StatefulSet is not ready
kubectl describe statefulset postgres -n default

# Check pod status
kubectl get pods -n default -l app=postgres

# Check pod logs
kubectl logs postgres-0 -n default

# Check events
kubectl get events -n default --field-selector involvedObject.name=postgres --sort-by='.lastTimestamp'

# Check PVCs for StatefulSet
kubectl get pvc -n default -l app=postgres

# Check service for StatefulSet
kubectl get svc postgres -n default

# Exec into pod
kubectl exec -it postgres-0 -n default -- bash

# Delete stuck pod (will be recreated)
kubectl delete pod postgres-0 -n default --grace-period=0 --force
```

### Update Strategies

```bash
# Check update strategy
kubectl get statefulset postgres -n default -o jsonpath='{.spec.updateStrategy.type}'

# RollingUpdate (default) - Updates pods in reverse order
# OnDelete - Manual pod deletion required for updates
```

### Pod Management Policies

```bash
# Check pod management policy
kubectl get statefulset postgres -n default -o jsonpath='{.spec.podManagementPolicy}'

# OrderedReady (default) - Waits for each pod to be ready before next
# Parallel - Creates/deletes all pods simultaneously
```

### Persistent Storage

```bash
# List PVCs created by StatefulSet
kubectl get pvc -n default | grep postgres

# Check PVC for specific pod
kubectl get pvc data-postgres-0 -n default

# StatefulSet PVCs are NOT deleted when StatefulSet is deleted
# Manual cleanup required:
kubectl delete pvc -n default -l app=postgres
```

### Common StatefulSet Applications

**Databases:**
- PostgreSQL, MySQL, MongoDB, CockroachDB
- MariaDB, Percona

**Cache:**
- Redis, Memcached

**Message Queues:**
- Kafka, RabbitMQ, NATS, Pulsar

**Search/Analytics:**
- Elasticsearch, OpenSearch, Solr

**Distributed Systems:**
- Cassandra, ScyllaDB, Consul, etcd, ZooKeeper

**Time Series:**
- Prometheus, InfluxDB, TimescaleDB

### Best Practices

1. **Headless service** - Always create a headless service for StatefulSets
2. **Pod disruption budgets** - Use PDBs to prevent all pods from being down
3. **Init containers** - Use init containers for setup/coordination
4. **Readiness probes** - Configure proper readiness probes for accurate ready count
5. **Liveness probes** - Prevent stuck pods with liveness probes
6. **Resource limits** - Set CPU/memory limits for predictable performance
7. **Anti-affinity** - Use pod anti-affinity to spread across nodes
8. **Storage sizing** - Plan PVC size carefully (expansion requires specific storage class support)
9. **Backup strategy** - Implement backup for StatefulSet persistent data
10. **Monitoring** - Monitor replica health, disk usage, and application metrics

### High Availability Checks

```yaml
# Ensure minimum replicas for quorum
statefulsets:
  - name: "Consul has quorum"
    statefulset: consul
    namespace: service-mesh
    state: available
    replicas: 5
    ready_replicas: 5  # Minimum 3 needed for quorum, 5 for fault tolerance

  - name: "etcd cluster has quorum"
    statefulset: etcd
    namespace: kube-system
    state: available
    replicas: 3
    ready_replicas: 3  # Minimum 2 needed for quorum

  - name: "Kafka cluster operational"
    statefulset: kafka
    namespace: streaming
    state: available
    replicas: 3
    ready_replicas: 3
```

### Disaster Recovery Testing

```yaml
# Test cluster survives with reduced capacity
statefulsets:
  - name: "Database degraded but operational"
    statefulset: postgres
    namespace: production
    state: exists
    replicas: 5
    ready_replicas: 3  # Survives 2 node failures

# Test cluster fully recovered
statefulsets:
  - name: "Database fully recovered"
    statefulset: postgres
    namespace: production
    state: available
    replicas: 5
    ready_replicas: 5
```
