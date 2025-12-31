#!/bin/bash

# Realistic integration test for inventory file functionality
# This test properly demonstrates the inventory feature by:
# 1. Creating a custom Docker network
# 2. Starting 3 containers all accessible on port 2222
# 3. Running platform-spec from within a container on the same network
# 4. Using container names in the inventory file

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SSH_KEY="$SCRIPT_DIR/test_key"
INVENTORY_FILE="$SCRIPT_DIR/inventory-hosts.txt"
SPEC_FILE="$SCRIPT_DIR/test-spec.yaml"
NETWORK_NAME="platform-spec-test-net"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

function log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

function log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

function cleanup() {
    log_info "Cleaning up..."

    docker-compose -f "$SCRIPT_DIR/docker-compose.yml" down -v 2>/dev/null || true
    docker network rm "$NETWORK_NAME" 2>/dev/null || true
    rm -f "$SSH_KEY" "$SSH_KEY.pub" "$SCRIPT_DIR/ssh_config" 2>/dev/null || true

    log_info "Cleanup complete"
}

trap cleanup EXIT INT TERM

function build_binary() {
    log_info "Building binary for local platform..."
    cd "$PROJECT_ROOT"
    make build

    log_info "Building Linux binary for Docker container..."
    GOOS=linux GOARCH=amd64 go build -o "$PROJECT_ROOT/dist/platform-spec-linux" ./cmd/platform-spec

    if [[ ! -f "$PROJECT_ROOT/dist/platform-spec-linux" ]]; then
        log_error "Linux binary not found"
        exit 1
    fi

    log_info "Binaries built successfully"
}

function generate_ssh_key() {
    log_info "Generating SSH key..."
    rm -f "$SSH_KEY" "$SSH_KEY.pub"
    ssh-keygen -t rsa -b 2048 -f "$SSH_KEY" -N "" -q
    chmod 600 "$SSH_KEY"
    log_info "SSH key generated"
}

function start_containers() {
    log_info "Creating Docker network: $NETWORK_NAME"
    docker network create "$NETWORK_NAME" 2>/dev/null || true

    log_info "Starting Docker containers..."
    cd "$SCRIPT_DIR"
    docker-compose up -d

    # Connect containers to custom network
    for i in 1 2 3; do
        docker network connect "$NETWORK_NAME" platform-spec-test-$i 2>/dev/null || true
    done

    log_info "Waiting for containers to be ready..."
    sleep 5

    for i in {1..30}; do
        if docker exec platform-spec-test-1 pgrep sshd >/dev/null 2>&1 && \
           docker exec platform-spec-test-2 pgrep sshd >/dev/null 2>&1 && \
           docker exec platform-spec-test-3 pgrep sshd >/dev/null 2>&1; then
            log_info "All containers are ready"
            return 0
        fi
        sleep 1
    done

    log_error "Containers failed to start properly"
    exit 1
}

function setup_ssh_access() {
    log_info "Setting up SSH access to containers..."
    local pubkey=$(cat "$SSH_KEY.pub")

    for i in 1 2 3; do
        docker exec platform-spec-test-$i mkdir -p /config/.ssh
        docker exec platform-spec-test-$i sh -c "echo '$pubkey' >> /config/.ssh/authorized_keys"
        docker exec platform-spec-test-$i chmod 700 /config/.ssh
        docker exec platform-spec-test-$i chmod 600 /config/.ssh/authorized_keys
        docker exec platform-spec-test-$i chown -R testuser:testuser /config/.ssh
    done

    log_info "SSH access configured"
}

function create_inventory_file() {
    log_info "Creating inventory file with container names..."

    # Note: Even though the design said hosts-only, the parser accepts user@host format
    # since it just validates no whitespace. We use this for testing.
    cat > "$INVENTORY_FILE" <<EOF
# Test inventory with Docker container names
# All accessible on port 2222 via custom Docker network
# Using testuser@host format for authentication
testuser@platform-spec-test-1
testuser@platform-spec-test-2
testuser@platform-spec-test-3
EOF

    log_info "Inventory file created:"
    cat "$INVENTORY_FILE"
}

function test_with_inventory() {
    log_info "========================================="
    log_info "Testing with inventory file"
    log_info "========================================="

    # Run platform-spec from within a container on the same network
    # This allows it to resolve container names
    log_info "Running platform-spec with inventory file from within Docker network..."

    # Note: We need to pass the username since inventory mode defaults to 'root'
    # but our containers use 'testuser'
    # We do this by creating a wrapper inventory with testuser@ prefix
    # OR by using SSH config
    # For simplicity, let's create a modified inventory file with user@host format

    # Actually, the current implementation doesn't support user@host in inventory
    # So we need to use a workaround: create user-specific SSH config or modify the code
    # For now, let's use SSH config approach

    log_info "Creating SSH config for testuser..."
    cat > "$SCRIPT_DIR/ssh_config" <<EOF
Host *
    User testuser
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null
EOF

    docker run --rm \
        --network "$NETWORK_NAME" \
        -v "$PROJECT_ROOT/dist/platform-spec-linux:/platform-spec:ro" \
        -v "$SPEC_FILE:/spec.yaml:ro" \
        -v "$INVENTORY_FILE:/inventory.txt:ro" \
        -v "$SSH_KEY:/ssh_key:ro" \
        -v "$SCRIPT_DIR/ssh_config:/etc/ssh/ssh_config:ro" \
        alpine:latest \
        /platform-spec test remote \
        --inventory /inventory.txt \
        /spec.yaml \
        -i /ssh_key \
        -p 2222 \
        --insecure-ignore-host-key \
        --no-color

    if [[ $? -eq 0 ]]; then
        log_info ""
        log_info "✓ Inventory test PASSED!"
        log_info ""
        log_info "Successfully tested 3 hosts using inventory file:"
        log_info "  - platform-spec-test-1"
        log_info "  - platform-spec-test-2"
        log_info "  - platform-spec-test-3"
        return 0
    else
        log_error "✗ Inventory test FAILED"
        return 1
    fi
}

main() {
    log_info "========================================="
    log_info "Realistic Inventory Integration Test"
    log_info "========================================="

    build_binary
    generate_ssh_key
    start_containers
    setup_ssh_access
    create_inventory_file

    log_info ""
    test_with_inventory

    log_info ""
    log_info "========================================="
    log_info "Test complete!"
    log_info "========================================="
}

main "$@"
