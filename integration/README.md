# Integration Tests

This directory contains integration tests for platform-spec that use real SSH connections and Docker containers.

## Prerequisites

- Docker and Docker Compose installed
- SSH client installed
- Bash shell

## Inventory File Integration Test

The `test-inventory.sh` script tests the `--inventory` flag functionality by:

1. Building the platform-spec binary
2. Starting 3 SSH-enabled Docker containers
3. Setting up SSH key-based authentication
4. Testing single-host connections (baseline)
5. Demonstrating multi-host capabilities

### Running the Tests

**Option 1: Realistic Test (Recommended)**

This test properly demonstrates the inventory feature with all hosts on the same port:

```bash
cd integration
./test-inventory-realistic.sh
```

This script:
- Creates a Docker network with 3 SSH containers
- All containers accessible on port 2222
- Runs platform-spec from within the Docker network
- Uses container names in the inventory file
- Shows actual multi-host output with summary

**Option 2: Basic Test**

```bash
cd integration
./test-inventory.sh
```

This test validates connectivity to each container individually.

The script will:
- Build the binary if needed
- Start Docker containers with SSH servers
- Generate SSH keys for authentication
- Test connectivity to each container
- Clean up automatically on exit

### Current Limitations

The current implementation uses a single SSH port for all hosts in an inventory file. This means:

- All hosts must be accessible on the same port (default: 22)
- To test multiple containers on localhost with different ports, you need separate test runs
- In production, this works well when all your infrastructure uses standard SSH port 22

### Real-World Usage

In a real deployment, you would have:

```
# inventory.txt
web-server-01.example.com
web-server-02.example.com
db-server-01.example.com
```

All accessible on port 22, and you would run:

```bash
platform-spec test remote --inventory inventory.txt spec.yaml -i ~/.ssh/key
```

## Docker Compose Setup

The `docker-compose.yml` file creates 3 containers:
- `platform-spec-test-1` on port 2221
- `platform-spec-test-2` on port 2222
- `platform-spec-test-3` on port 2223

Each container runs an SSH server with:
- Username: `testuser`
- Password: `testpass` (for manual testing)
- SSH key authentication (configured by test script)

## Manual Testing

You can also start the containers manually for exploratory testing:

```bash
# Start containers
docker-compose up -d

# Wait for them to be ready
sleep 5

# Test single host
../dist/platform-spec test remote testuser@127.0.0.1 test-spec.yaml \
  -p 2221 \
  --insecure-ignore-host-key \
  --password testpass

# Stop containers
docker-compose down
```

## Cleanup

The test script automatically cleans up:
- Stops and removes containers
- Removes generated SSH keys
- Removes Docker networks

If cleanup fails, you can manually run:

```bash
docker-compose down -v
docker network rm platform-spec-test
rm -f test_key test_key.pub
```

## Using Make

You can also run the integration test via Make:

```bash
# Run just the inventory integration test
make test-inventory

# Run all integration tests (includes inventory test)
make test-integration
```

## Future Enhancements

Potential improvements for integration tests:

1. ✅ **True Multi-Host Test**: Implemented in `test-inventory-realistic.sh`
2. **Test with Jump Host**: Add bastion container to test `-J` flag with inventory
3. **Test Failures**: Intentionally fail some containers to test error handling
4. **Parallel Execution**: Once implemented, test `--parallel N` flag
5. **GitHub Actions CI**: Automate these tests in CI pipeline
