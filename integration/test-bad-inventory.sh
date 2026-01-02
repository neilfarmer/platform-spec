#!/bin/bash
set -e

echo "=== Setting up Bad Inventory Demo Environment ==="
echo ""

# Create SSH key if it doesn't exist
if [ ! -f ssh-key ]; then
    echo "Generating SSH key..."
    ssh-keygen -t rsa -b 2048 -f ssh-key -N "" -C "platform-spec-test"
    chmod 600 ssh-key
    chmod 644 ssh-key.pub
fi

echo "Starting 5 SSH containers with docker-compose..."
docker-compose -f docker-compose-demo.yml down -v 2>/dev/null || true
docker-compose -f docker-compose-demo.yml up -d

echo ""
echo "Waiting for all containers to be healthy..."
for i in {1..30}; do
    all_healthy=true
    for j in {1..5}; do
        if ! docker inspect demo-test-host-${j} --format='{{.State.Health.Status}}' 2>/dev/null | grep -q "healthy"; then
            all_healthy=false
            break
        fi
    done

    if [ "$all_healthy" = true ]; then
        echo "All containers are healthy!"
        break
    fi

    echo "Waiting... ($i/30)"
    sleep 2
done

echo ""
echo "Configuring SSH keys and packages..."

# Copy SSH key to all containers (for testuser)
for i in {1..5}; do
    docker cp ssh-key.pub demo-test-host-${i}:/tmp/authorized_keys
    docker exec demo-test-host-${i} sh -c 'mkdir -p /config/.ssh && mv /tmp/authorized_keys /config/.ssh/authorized_keys && chmod 600 /config/.ssh/authorized_keys && chown -R 1000:1000 /config/.ssh'
done

# Configure containers with packages
for i in {1..5}; do
    container_name="demo-test-host-${i}"

    # For first 2 containers, install some packages to make tests pass
    if [ $i -le 2 ]; then
        echo "  Container ${i}: Installing packages (tests will mostly pass)..."
        docker exec ${container_name} apk add --no-cache bash nginx 2>/dev/null || true
        docker exec ${container_name} sh -c 'echo "hello world" > /tmp/test.txt'
    else
        echo "  Container ${i}: Minimal setup (tests will fail)..."
        docker exec ${container_name} apk add --no-cache bash 2>/dev/null || true
    fi
done

echo ""
echo "Creating inventory file..."
cat > inventory-demo.txt <<EOF
# Working hosts (containers on Docker network)
testuser@demo-test-host-1
testuser@demo-test-host-2
testuser@demo-test-host-3
testuser@demo-test-host-4
testuser@demo-test-host-5

# Unreachable hosts (will fail to connect)
root@192.0.2.1
root@192.0.2.2
root@192.0.2.3
root@192.0.2.4
root@192.0.2.5
EOF

echo ""
echo "Creating test spec..."
cat > spec-demo.yaml <<EOF
name: "Bad Inventory Demo"
description: "Mix of passing and failing tests"
timeout: 5

tests:
  packages:
    - name: "Bash installed"
      packages:
        - bash
      state: present

    - name: "Nginx installed"
      packages:
        - nginx
      state: present

  files:
    - name: "/etc/passwd exists"
      path: "/etc/passwd"
      state: present

    - name: "/tmp/test.txt exists"
      path: "/tmp/test.txt"
      state: present

  command_content:
    - name: "Echo test"
      command: "echo hello"
      contains: ["hello"]

    - name: "Test file contains hello"
      command: "cat /tmp/test.txt 2>/dev/null || echo not found"
      contains: ["hello"]
EOF

echo ""
echo "✅ Environment ready!"
echo ""
echo "Running tests from runner container (on same Docker network)..."
echo ""

# Start a runner container on the same network to run the tests
docker run --rm \
    --network platform-spec-demo-net \
    -v "$(pwd)/inventory-demo.txt:/inventory.txt:ro" \
    -v "$(pwd)/spec-demo.yaml:/spec.yaml:ro" \
    -v "$(pwd)/ssh-key:/tmp/ssh-key:ro" \
    -v "$(pwd)/../dist/platform-spec-linux:/platform-spec:ro" \
    alpine:latest \
    sh -c 'mkdir -p /root/.ssh && cp /tmp/ssh-key /root/.ssh/id_rsa && chmod 600 /root/.ssh/id_rsa && /platform-spec test remote \
        --inventory /inventory.txt \
        /spec.yaml \
        -i /root/.ssh/id_rsa \
        -p 2222 \
        --timeout 5 \
        --insecure-ignore-host-key' \
    || true

echo ""
echo ""
echo "=== Cleanup ==="
echo "Stopping containers with docker-compose..."
docker-compose -f docker-compose-demo.yml down -v 2>/dev/null || true
rm -f ssh-key ssh-key.pub inventory-demo.txt spec-demo.yaml

echo ""
echo "✅ Demo completed!"
