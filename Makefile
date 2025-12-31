.PHONY: build clean test install release-build deploy-kind-cluster destroy-kind-cluster security-scan security-scan-vuln security-scan-static test-docker test-docker-local test-kubernetes test-integration test-inventory test-jump destroy-test-jump

# Cluster name for kind
KIND_CLUSTER_NAME ?= platform-spec-test

# Verbose mode for tests (set VERBOSE=true to enable)
VERBOSE ?= false
VERBOSE_FLAG :=
ifeq ($(VERBOSE),true)
	VERBOSE_FLAG := --verbose
endif

build:
	mkdir -p dist
	go build -o dist/platform-spec ./cmd/platform-spec

release-build:
	mkdir -p dist
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/platform-spec-darwin-arm64 ./cmd/platform-spec
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/platform-spec-linux-amd64 ./cmd/platform-spec

clean:
	rm -rf dist/
	rm -f coverage.out coverage-filtered.out

test:
	@echo "Running tests with coverage..."
	@go test -v -count=1 -coverprofile=coverage.out -covermode=count ./...
	@echo ""
	@echo "Calculating coverage (excluding cmd/)..."
	@grep -v "cmd/platform-spec" coverage.out > coverage-filtered.out
	@COVERAGE=$$(go tool cover -func=coverage-filtered.out | grep total: | awk '{print $$3}' | sed 's/%//'); \
	echo "Current coverage: $${COVERAGE}%"; \
	echo "Required coverage: 80%"; \
	if [ $$(echo "$${COVERAGE} < 80" | bc) -eq 1 ]; then \
		echo "❌ FAIL: Coverage $${COVERAGE}% is below minimum 80%"; \
		rm -f coverage.out coverage-filtered.out; \
		exit 1; \
	else \
		echo "✅ PASS: Coverage $${COVERAGE}% meets minimum 80%"; \
	fi
	@rm -f coverage.out coverage-filtered.out

install: build
	cp dist/platform-spec $(GOPATH)/bin/platform-spec

# Kind cluster targets

deploy-kind-cluster:
	@echo "Creating kind cluster '$(KIND_CLUSTER_NAME)'..."
	@if kind get clusters 2>/dev/null | grep -q "^$(KIND_CLUSTER_NAME)$$"; then \
		echo "Cluster '$(KIND_CLUSTER_NAME)' already exists"; \
	else \
		kind create cluster --config integration/kubernetes/kind-config.yaml --wait 60s; \
		kubectl wait --for=condition=Ready nodes --all --timeout=60s; \
		echo "Cluster ready!"; \
	fi

destroy-kind-cluster:
	@echo "Deleting kind cluster '$(KIND_CLUSTER_NAME)'..."
	@kind delete cluster --name $(KIND_CLUSTER_NAME) || true

# Docker integration test (auto-detects architecture)
test-docker:
	@echo "Building Linux binary for Docker..."
	@GOOS=linux GOARCH=$$(go env GOARCH) go build -o dist/platform-spec-linux ./cmd/platform-spec
	@echo ""
	@echo "Building Docker test image..."
	@docker build -f integration/docker/Dockerfile.test -t platform-spec-test .
	@echo ""
	@echo "Running tests in Docker container..."
	@docker run --rm platform-spec-test platform-spec test local /test-spec.yaml $(VERBOSE_FLAG)

# Docker integration test for local ARM64 development
test-docker-local:
	@echo "Building Linux ARM64 binary for Docker..."
	@GOOS=linux GOARCH=arm64 go build -o dist/platform-spec-linux ./cmd/platform-spec
	@echo ""
	@echo "Building Docker test image (ARM64)..."
	@docker build -f integration/docker/Dockerfile.test.local -t platform-spec-test-local .
	@echo ""
	@echo "Running tests in Docker container (ARM64)..."
	@docker run --rm platform-spec-test-local platform-spec test local /test-spec.yaml $(VERBOSE_FLAG)

# Kubernetes integration test (matches CI pipeline)
test-kubernetes: deploy-kind-cluster build
	@echo ""
	@echo "Running Kubernetes integration tests..."
	@./dist/platform-spec test kubernetes examples/kubernetes-basic.yaml $(VERBOSE_FLAG)

# Inventory integration test
test-inventory: build
	@echo "=== Running Inventory Integration Test ==="
	@echo ""
	@cd integration && ./test-inventory-realistic.sh

# Run all integration tests (local)
test-integration: test-docker-local test-kubernetes test-inventory
	@echo ""
	@echo "✅ All integration tests completed successfully!"

# Security scanning - Vulnerability check
security-scan-vuln:
	@echo "Installing govulncheck if needed..."
	@which govulncheck > /dev/null 2>&1 || go install golang.org/x/vuln/cmd/govulncheck@latest
	@echo ""
	@echo "=== Running vulnerability scan (govulncheck) ==="
	@PATH="$(HOME)/go/bin:$(PATH)" govulncheck -show verbose ./...

# Security scanning - Static analysis
security-scan-static:
	@echo "Installing gosec if needed..."
	@which gosec > /dev/null 2>&1 || go install github.com/securego/gosec/v2/cmd/gosec@latest
	@echo ""
	@echo "=== Running static security analysis (gosec) ==="
	@PATH="$(HOME)/go/bin:$(PATH)" gosec -fmt=text ./...

# Security scanning - Run both
security-scan: security-scan-vuln security-scan-static

# Jump host testing - Deploy test environment
test-jump: build
	@echo "=== Setting up Jump Host Test Environment ==="
	@echo ""
	@echo "Creating SSH keypairs (jump and target)..."
	@mkdir -p integration/jump-host/ssh-keys
	@if [ ! -f integration/jump-host/ssh-keys/jump_key ]; then \
		ssh-keygen -t rsa -b 2048 -f integration/jump-host/ssh-keys/jump_key -N "" -C "platform-spec-jump"; \
		chmod 600 integration/jump-host/ssh-keys/jump_key; \
		chmod 644 integration/jump-host/ssh-keys/jump_key.pub; \
		echo "Created jump_key"; \
	else \
		echo "jump_key already exists"; \
	fi
	@if [ ! -f integration/jump-host/ssh-keys/target_key ]; then \
		ssh-keygen -t rsa -b 2048 -f integration/jump-host/ssh-keys/target_key -N "" -C "platform-spec-target"; \
		chmod 600 integration/jump-host/ssh-keys/target_key; \
		chmod 644 integration/jump-host/ssh-keys/target_key.pub; \
		echo "Created target_key"; \
	else \
		echo "target_key already exists"; \
	fi
	@echo ""
	@echo "Building and starting containers..."
	@cd integration/jump-host && docker compose up -d --build
	@echo ""
	@echo "Waiting for containers to be ready..."
	@sleep 5
	@echo ""
	@echo "Copying SSH keys into containers..."
	@docker cp integration/jump-host/ssh-keys/jump_key.pub platform-spec-jump:/tmp/authorized_keys
	@docker exec platform-spec-jump sh -c 'mv /tmp/authorized_keys /home/testuser/.ssh/authorized_keys && chown testuser:testuser /home/testuser/.ssh/authorized_keys && chmod 600 /home/testuser/.ssh/authorized_keys'
	@docker cp integration/jump-host/ssh-keys/target_key.pub platform-spec-target:/tmp/authorized_keys
	@docker exec platform-spec-target sh -c 'mv /tmp/authorized_keys /home/testuser/.ssh/authorized_keys && chown testuser:testuser /home/testuser/.ssh/authorized_keys && chmod 600 /home/testuser/.ssh/authorized_keys'
	@docker exec platform-spec-jump pkill -HUP sshd || true
	@docker exec platform-spec-target pkill -HUP sshd || true
	@sleep 2
	@echo ""
	@echo "Adding jump host to known_hosts..."
	@ssh-keyscan -p 2222 -H localhost >> ~/.ssh/known_hosts 2>/dev/null || true
	@echo ""
	@echo "✅ Jump host test environment is ready!"
	@echo ""
	@echo "Test with SEPARATE keys (jump-key for jump host, target-key for target):"
	@echo "  ./dist/platform-spec test remote -J testuser@localhost --jump-port 2222 --jump-identity integration/jump-host/ssh-keys/jump_key -i integration/jump-host/ssh-keys/target_key testuser@target-host integration/jump-host/spec.yaml --insecure-ignore-host-key --verbose"
	@echo ""
	@echo "Or test direct connections:"
	@echo "  ssh -i integration/jump-host/ssh-keys/jump_key -p 2222 testuser@localhost"
	@echo ""
	@echo "When done, run: make destroy-test-jump"

# Jump host testing - Destroy test environment
destroy-test-jump:
	@echo "=== Tearing down Jump Host Test Environment ==="
	@echo ""
	@echo "Stopping and removing containers..."
	@cd integration/jump-host && docker compose down -v 2>/dev/null || true
	@echo ""
	@echo "Removing SSH keypair..."
	@rm -rf integration/jump-host/ssh-keys
	@echo ""
	@echo "Removing known_hosts entries for localhost:2222..."
	@ssh-keygen -R "[localhost]:2222" 2>/dev/null || true
	@echo ""
	@echo "✅ Jump host test environment destroyed"
