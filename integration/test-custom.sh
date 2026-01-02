#!/usr/bin/env bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

NUM_CONTAINERS=50
JUMP_HOST="testuser@ssh-jump"
JUMP_PORT=2222

echo -e "${BLUE}================================${NC}"
echo -e "${BLUE}Custom Integration Test${NC}"
echo -e "${BLUE}================================${NC}"
echo ""
echo "Architecture: runner --> ssh-jump --> ssh-custom-1..${NUM_CONTAINERS}"
echo ""

cleanup() {
    echo -e "${YELLOW}Cleaning up...${NC}"
    docker compose -f docker-compose-custom.yml down -v 2>/dev/null || true
    rm -f ssh-test-key ssh-test-key.pub
}
trap cleanup EXIT

echo -e "${YELLOW}Building Linux binary...${NC}"
cd ..
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/platform-spec-linux ./cmd/platform-spec
chmod +x dist/platform-spec-linux
cd integration

if [ ! -f ssh-test-key ]; then
    echo -e "${YELLOW}Generating SSH key...${NC}"
    ssh-keygen -t rsa -b 2048 -f ssh-test-key -N "" -C "platform-spec-custom-test"
    chmod 600 ssh-test-key
    chmod 644 ssh-test-key.pub
fi

echo -e "${YELLOW}Starting containers...${NC}"
docker compose -f docker-compose-custom.yml up -d --wait

echo -e "${YELLOW}Enabling TCP forwarding on jump host...${NC}"
docker exec ssh-jump sh -c "sed -i '/AllowTcpForwarding/d' /config/sshd/sshd_config && echo 'AllowTcpForwarding yes' >> /config/sshd/sshd_config"
docker exec ssh-jump pkill -HUP sshd.pam || true
sleep 2
echo -e "${GREEN}TCP forwarding enabled${NC}"

echo -e "${YELLOW}Verifying connectivity...${NC}"
READY=0
for i in $(seq 1 ${NUM_CONTAINERS}); do
    docker exec ssh-jump nc -z "ssh-custom-$i" 2222 2>/dev/null && ((READY++))
done
echo -e "${GREEN}${READY}/${NUM_CONTAINERS} containers reachable${NC}"
echo ""

echo -e "${BLUE}Test 1: Single Host via Jump${NC}"
docker exec custom-test-runner platform-spec test remote \
    testuser@ssh-custom-1 -p 2222 -i /root/.ssh/id_rsa \
    -J ${JUMP_HOST} --jump-port ${JUMP_PORT} --jump-identity /root/.ssh/id_rsa \
    --insecure-ignore-host-key /spec.yaml --verbose
echo ""

echo -e "${BLUE}Test 2: Sequential (50 hosts)${NC}"
START_SEQ=$(date +%s)
docker exec custom-test-runner platform-spec test remote \
    --inventory /inventory.txt -p 2222 -i /root/.ssh/id_rsa \
    -J ${JUMP_HOST} --jump-port ${JUMP_PORT} --jump-identity /root/.ssh/id_rsa \
    --insecure-ignore-host-key /spec.yaml
SEQ_TIME=$(($(date +%s) - START_SEQ))
echo -e "${GREEN}Sequential: ${SEQ_TIME}s${NC}"
echo ""

echo -e "${BLUE}Test 3: Parallel 10 workers${NC}"
START=$(date +%s)
docker exec custom-test-runner platform-spec test remote \
    --inventory /inventory.txt -p 2222 -i /root/.ssh/id_rsa \
    -J ${JUMP_HOST} --jump-port ${JUMP_PORT} --jump-identity /root/.ssh/id_rsa \
    --insecure-ignore-host-key --parallel 10 /spec.yaml
PAR10_TIME=$(($(date +%s) - START))
echo -e "${GREEN}Parallel 10: ${PAR10_TIME}s${NC}"
echo ""

echo -e "${BLUE}Test 4: Parallel 25 workers${NC}"
START=$(date +%s)
docker exec custom-test-runner platform-spec test remote \
    --inventory /inventory.txt -p 2222 -i /root/.ssh/id_rsa \
    -J ${JUMP_HOST} --jump-port ${JUMP_PORT} --jump-identity /root/.ssh/id_rsa \
    --insecure-ignore-host-key --parallel 25 /spec.yaml
PAR25_TIME=$(($(date +%s) - START))
echo -e "${GREEN}Parallel 25: ${PAR25_TIME}s${NC}"
echo ""

echo -e "${BLUE}Test 5: Parallel 50 workers${NC}"
START=$(date +%s)
docker exec custom-test-runner platform-spec test remote \
    --inventory /inventory.txt -p 2222 -i /root/.ssh/id_rsa \
    -J ${JUMP_HOST} --jump-port ${JUMP_PORT} --jump-identity /root/.ssh/id_rsa \
    --insecure-ignore-host-key --parallel 50 /spec.yaml
PAR50_TIME=$(($(date +%s) - START))
echo -e "${GREEN}Parallel 50: ${PAR50_TIME}s${NC}"
echo ""

echo -e "${BLUE}================================${NC}"
echo -e "${BLUE}Performance Summary${NC}"
echo -e "${BLUE}================================${NC}"
echo "50 containers x 30 tests = 1500 test executions"
echo "All via jump host"
echo ""
echo "Sequential:    ${SEQ_TIME}s"
echo "Parallel 10:   ${PAR10_TIME}s"
echo "Parallel 25:   ${PAR25_TIME}s"
echo "Parallel 50:   ${PAR50_TIME}s"
echo ""
echo -e "${GREEN}All tests passed!${NC}"
