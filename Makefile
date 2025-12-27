.PHONY: build clean test test-verbose install release-build deploy-kind-cluster destroy-kind-cluster security-scan security-scan-vuln security-scan-static test-docker

# Cluster name for kind
KIND_CLUSTER_NAME ?= platform-spec-test

build:
	mkdir -p dist
	go build -o dist/platform-spec ./cmd/platform-spec

release-build:
	mkdir -p dist
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/platform-spec-darwin-arm64 ./cmd/platform-spec
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/platform-spec-linux-amd64 ./cmd/platform-spec

clean:
	rm -rf dist/

test:
	go test -v -count=1 ./...

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

# Docker integration test
test-docker:
	@echo "Building Linux binary for Docker..."
	@GOOS=linux GOARCH=arm64 go build -o dist/platform-spec-linux ./cmd/platform-spec
	@echo ""
	@echo "Building Docker test image..."
	@docker build -f integration/docker/Dockerfile.test -t platform-spec-test . --quiet
	@echo ""
	@echo "Running tests in Docker container..."
	@docker run --rm platform-spec-test

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
