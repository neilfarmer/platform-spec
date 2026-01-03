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

echo -e "${BLUE}Running Custom Test - ${NUM_CONTAINERS} containers${NC}"
echo ""
echo -e "${YELLOW}Command:${NC}"
echo "docker exec custom-test-runner platform-spec test remote \\"
echo "    --inventory /inventory.txt -p 2222 -i /root/.ssh/id_rsa \\"
echo "    -J ${JUMP_HOST} --jump-port ${JUMP_PORT} --jump-identity /root/.ssh/id_rsa \\"
echo "    --insecure-ignore-host-key --parallel 10 /spec.yaml"
echo ""
echo "Testing ${NUM_CONTAINERS} containers with 30 tests each = $((NUM_CONTAINERS * 30)) total test executions"
echo ""

START=$(date +%s)
docker exec custom-test-runner platform-spec test remote \
    --inventory /inventory.txt -p 2222 -i /root/.ssh/id_rsa \
    -J ${JUMP_HOST} --jump-port ${JUMP_PORT} --jump-identity /root/.ssh/id_rsa \
    --insecure-ignore-host-key --parallel 10 /spec.yaml
ELAPSED=$(($(date +%s) - START))

echo ""
echo -e "${GREEN}Test completed: ${NUM_CONTAINERS} containers, $((NUM_CONTAINERS * 30)) tests in ${ELAPSED}s${NC}"
