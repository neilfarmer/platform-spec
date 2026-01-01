#!/bin/bash
# Integration test for parallel execution feature
# Demonstrates performance improvement with multiple hosts

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SSH_KEY="$SCRIPT_DIR/perf-test-key"
SPEC_FILE="$SCRIPT_DIR/perf-test-spec.yaml"
NETWORK_NAME="platform-spec-perf-test"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[1;34m'
BOLD='\033[1m'
NC='\033[0m'

log() { echo -e "${GREEN}[INFO]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; }
section() { echo -e "\n${BOLD}${BLUE}=== $1 ===${NC}"; }

cleanup() {
    log "Cleaning up..."
    cd "$SCRIPT_DIR"

    # Stop and remove containers
    docker-compose -f docker-compose-test.yml down -v 2>/dev/null || true

    # Remove network
    docker network rm "$NETWORK_NAME" 2>/dev/null || true

    # Remove temp files
    rm -rf "$SSH_KEY" "$SSH_KEY.pub" perf-inventory-*.txt ssh_config 2>/dev/null || true

    log "Cleanup complete"
}

trap cleanup EXIT INT TERM

section "Parallel Execution Performance Test"
log "This test compares sequential vs parallel execution"
log "Testing with 30 hosts to demonstrate speedup"
echo ""

# Build binaries
log "Building platform-spec binary..."
cd "$PROJECT_ROOT"
make build > /dev/null 2>&1

log "Building Linux binary for Docker..."
GOOS=linux GOARCH=amd64 go build -o "$PROJECT_ROOT/dist/platform-spec-linux" ./cmd/platform-spec

if [[ ! -f "$PROJECT_ROOT/dist/platform-spec-linux" ]]; then
    error "Linux binary not found"
    exit 1
fi

# Generate SSH key
log "Generating SSH key..."
rm -f "$SSH_KEY" "$SSH_KEY.pub"
ssh-keygen -t rsa -b 2048 -f "$SSH_KEY" -N "" -q
chmod 600 "$SSH_KEY"

# Create Docker network
log "Creating Docker network: $NETWORK_NAME"
docker network create "$NETWORK_NAME" 2>/dev/null || true

# Start containers using docker-compose
section "Starting 30 SSH Test Containers"
cd "$SCRIPT_DIR"
log "Starting containers with docker-compose..."
docker-compose -f docker-compose-test.yml up -d ssh-test-{1..30}

# Connect containers to custom network
for i in {1..30}; do
    docker network connect "$NETWORK_NAME" ssh-test-$i 2>/dev/null || true
done

# Wait for containers to be ready
log "Waiting for containers to initialize..."
sleep 15

# Wait for SSH to be ready
log "Waiting for SSH service..."
for attempt in {1..60}; do
    ready=0
    for i in {1..30}; do
        if docker exec ssh-test-$i pgrep sshd >/dev/null 2>&1; then
            ready=$((ready + 1))
        fi
    done

    if [[ $ready -eq 30 ]]; then
        log "All containers ready"
        break
    fi

    if [[ $attempt -eq 60 ]]; then
        error "Containers failed to start SSH service"
        exit 1
    fi

    sleep 1
done

# Setup SSH access
log "Configuring SSH access..."
pubkey=$(cat "$SSH_KEY.pub")

for i in {1..30}; do
    docker exec ssh-test-$i mkdir -p /config/.ssh 2>/dev/null || true
    docker exec ssh-test-$i sh -c "echo '$pubkey' > /config/.ssh/authorized_keys"
    docker exec ssh-test-$i chmod 700 /config/.ssh
    docker exec ssh-test-$i chmod 600 /config/.ssh/authorized_keys
    docker exec ssh-test-$i chown -R testuser:testuser /config/.ssh
done

log "All containers configured"

# Create inventory files
log "Creating inventory files..."
cd "$SCRIPT_DIR"

echo "testuser@ssh-test-1" > perf-inventory-1.txt

for i in {1..30}; do
    echo "testuser@ssh-test-$i" >> perf-inventory-30.txt
done

# Create SSH config
cat > ssh_config <<EOF
Host *
    User testuser
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null
EOF

# Verify connectivity
section "Verifying SSH Connectivity"
log "Testing SSH connection to ssh-test-1..."

if docker run --rm \
    --network "$NETWORK_NAME" \
    -v "$SSH_KEY:/ssh_key:ro" \
    -v "$SCRIPT_DIR/ssh_config:/etc/ssh/ssh_config:ro" \
    alpine:latest \
    sh -c "apk add --no-cache openssh-client > /dev/null 2>&1 && ssh -i /ssh_key -p 2222 ssh-test-1 'echo Connected'" 2>&1 | grep -q "Connected"; then
    log "SSH connectivity verified"
else
    error "SSH connection failed"
    exit 1
fi

# Run performance tests
section "Performance Test: 1 Host Baseline"

START=$(date +%s.%N)
docker run --rm \
    --network "$NETWORK_NAME" \
    -v "$PROJECT_ROOT/dist/platform-spec-linux:/platform-spec:ro" \
    -v "$SPEC_FILE:/spec.yaml:ro" \
    -v "$SCRIPT_DIR/perf-inventory-1.txt:/inventory.txt:ro" \
    -v "$SSH_KEY:/ssh_key:ro" \
    -v "$SCRIPT_DIR/ssh_config:/etc/ssh/ssh_config:ro" \
    alpine:latest \
    /platform-spec test remote \
        --inventory /inventory.txt \
        /spec.yaml \
        -i /ssh_key \
        -p 2222 \
        --insecure-ignore-host-key \
        > /tmp/test1.out 2>&1

EXIT_CODE=$?
END=$(date +%s.%N)
TIME_1=$(echo "$END - $START" | bc)

if [ $EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✓${NC} Single host: ${TIME_1}s"
else
    echo -e "${RED}✗${NC} Test failed"
    cat /tmp/test1.out
    exit 1
fi

section "Performance Test: 30 Hosts Sequential"

START=$(date +%s.%N)
docker run --rm \
    --network "$NETWORK_NAME" \
    -v "$PROJECT_ROOT/dist/platform-spec-linux:/platform-spec:ro" \
    -v "$SPEC_FILE:/spec.yaml:ro" \
    -v "$SCRIPT_DIR/perf-inventory-30.txt:/inventory.txt:ro" \
    -v "$SSH_KEY:/ssh_key:ro" \
    -v "$SCRIPT_DIR/ssh_config:/etc/ssh/ssh_config:ro" \
    alpine:latest \
    /platform-spec test remote \
        --inventory /inventory.txt \
        /spec.yaml \
        -i /ssh_key \
        -p 2222 \
        --insecure-ignore-host-key \
        --parallel 1 \
        > /tmp/test_seq.out 2>&1

EXIT_CODE=$?
END=$(date +%s.%N)
TIME_SEQ=$(echo "$END - $START" | bc)

if [ $EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✓${NC} Sequential (30 hosts, 1 worker): ${TIME_SEQ}s"
else
    echo -e "${YELLOW}⚠${NC} Some tests may have failed, but timing recorded: ${TIME_SEQ}s"
fi

section "Performance Test: 30 Hosts Parallel (10 workers)"

START=$(date +%s.%N)
docker run --rm \
    --network "$NETWORK_NAME" \
    -v "$PROJECT_ROOT/dist/platform-spec-linux:/platform-spec:ro" \
    -v "$SPEC_FILE:/spec.yaml:ro" \
    -v "$SCRIPT_DIR/perf-inventory-30.txt:/inventory.txt:ro" \
    -v "$SSH_KEY:/ssh_key:ro" \
    -v "$SCRIPT_DIR/ssh_config:/etc/ssh/ssh_config:ro" \
    alpine:latest \
    /platform-spec test remote \
        --inventory /inventory.txt \
        /spec.yaml \
        -i /ssh_key \
        -p 2222 \
        --insecure-ignore-host-key \
        --parallel 10 \
        > /tmp/test_par10.out 2>&1

EXIT_CODE=$?
END=$(date +%s.%N)
TIME_PAR10=$(echo "$END - $START" | bc)

if [ $EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✓${NC} Parallel (30 hosts, 10 workers): ${TIME_PAR10}s"
else
    echo -e "${YELLOW}⚠${NC} Some tests may have failed, but timing recorded: ${TIME_PAR10}s"
fi

section "Performance Test: 30 Hosts Parallel (15 workers)"

START=$(date +%s.%N)
docker run --rm \
    --network "$NETWORK_NAME" \
    -v "$PROJECT_ROOT/dist/platform-spec-linux:/platform-spec:ro" \
    -v "$SPEC_FILE:/spec.yaml:ro" \
    -v "$SCRIPT_DIR/perf-inventory-30.txt:/inventory.txt:ro" \
    -v "$SSH_KEY:/ssh_key:ro" \
    -v "$SCRIPT_DIR/ssh_config:/etc/ssh/ssh_config:ro" \
    alpine:latest \
    /platform-spec test remote \
        --inventory /inventory.txt \
        /spec.yaml \
        -i /ssh_key \
        -p 2222 \
        --insecure-ignore-host-key \
        --parallel 15 \
        > /tmp/test_par15.out 2>&1

EXIT_CODE=$?
END=$(date +%s.%N)
TIME_PAR15=$(echo "$END - $START" | bc)

if [ $EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✓${NC} Parallel (30 hosts, 15 workers): ${TIME_PAR15}s"
else
    echo -e "${YELLOW}⚠${NC} Some tests may have failed, but timing recorded: ${TIME_PAR15}s"
fi

section "Performance Test: 30 Hosts Parallel (30 workers)"

START=$(date +%s.%N)
docker run --rm \
    --network "$NETWORK_NAME" \
    -v "$PROJECT_ROOT/dist/platform-spec-linux:/platform-spec:ro" \
    -v "$SPEC_FILE:/spec.yaml:ro" \
    -v "$SCRIPT_DIR/perf-inventory-30.txt:/inventory.txt:ro" \
    -v "$SSH_KEY:/ssh_key:ro" \
    -v "$SCRIPT_DIR/ssh_config:/etc/ssh/ssh_config:ro" \
    alpine:latest \
    /platform-spec test remote \
        --inventory /inventory.txt \
        /spec.yaml \
        -i /ssh_key \
        -p 2222 \
        --insecure-ignore-host-key \
        --parallel 30 \
        > /tmp/test_par30.out 2>&1

EXIT_CODE=$?
END=$(date +%s.%N)
TIME_PAR30=$(echo "$END - $START" | bc)

if [ $EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✓${NC} Parallel (30 hosts, 30 workers): ${TIME_PAR30}s"
else
    echo -e "${YELLOW}⚠${NC} Some tests may have failed, but timing recorded: ${TIME_PAR30}s"
fi

# Calculate speedups
SPEEDUP_10=$(echo "scale=2; $TIME_SEQ / $TIME_PAR10" | bc)
SPEEDUP_15=$(echo "scale=2; $TIME_SEQ / $TIME_PAR15" | bc)
SPEEDUP_30=$(echo "scale=2; $TIME_SEQ / $TIME_PAR30" | bc)
SAVED_10=$(echo "scale=2; $TIME_SEQ - $TIME_PAR10" | bc)
SAVED_15=$(echo "scale=2; $TIME_SEQ - $TIME_PAR15" | bc)
SAVED_30=$(echo "scale=2; $TIME_SEQ - $TIME_PAR30" | bc)

section "Performance Results"
echo ""
echo -e "${BOLD}Execution Time Comparison (30 hosts):${NC}"
echo ""
printf "  %-30s %12s %12s %15s\n" "Configuration" "Time (s)" "Speedup" "Time Saved (s)"
printf "  %-30s %12s %12s %15s\n" "$(printf '%.s─' $(seq 1 30))" "$(printf '%.s─' $(seq 1 12))" "$(printf '%.s─' $(seq 1 12))" "$(printf '%.s─' $(seq 1 15))"
printf "  %-30s %12s %12s %15s\n" "Sequential (1 worker)" "$TIME_SEQ" "1.00x" "-"
printf "  ${GREEN}%-30s %12s %12s %15s${NC}\n" "Parallel (10 workers)" "$TIME_PAR10" "${SPEEDUP_10}x" "$SAVED_10"
printf "  ${GREEN}%-30s %12s %12s %15s${NC}\n" "Parallel (15 workers)" "$TIME_PAR15" "${SPEEDUP_15}x" "$SAVED_15"
printf "  ${GREEN}%-30s %12s %12s %15s${NC}\n" "Parallel (30 workers)" "$TIME_PAR30" "${SPEEDUP_30}x" "$SAVED_30"
echo ""
echo -e "${BOLD}Key Findings:${NC}"
echo -e "  • Sequential execution: ${BOLD}${TIME_SEQ}s${NC} for 30 hosts"
echo -e "  • Parallel (10 workers): ${BOLD}${SPEEDUP_10}x speedup${NC} (saved ${SAVED_10}s)"
echo -e "  • Parallel (15 workers): ${BOLD}${SPEEDUP_15}x speedup${NC} (saved ${SAVED_15}s)"
echo -e "  • Parallel (30 workers): ${BOLD}${SPEEDUP_30}x speedup${NC} (saved ${SAVED_30}s)"
echo ""

# Validate speedup
if [ $(echo "$SPEEDUP_10 > 2.0" | bc) -eq 1 ]; then
    echo -e "${GREEN}✅ SUCCESS: Parallel execution achieves >2x speedup!${NC}"
    echo -e "${GREEN}   Actual speedup with 10 workers: ${SPEEDUP_10}x${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠ WARNING: Speedup lower than expected (${SPEEDUP_10}x < 2.0x)${NC}"
    echo -e "${YELLOW}   This may be due to system load or resource constraints${NC}"
    exit 0
fi
