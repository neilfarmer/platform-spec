.PHONY: build clean test test-coverage install release-build deploy-kind-cluster destroy-kind-cluster

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
	rm -f coverage.out

test:
	@echo "Running tests with coverage..."
	@go test -v -count=1 -coverprofile=coverage.out -covermode=count ./...
	@echo ""
	@echo "Calculating coverage (excluding cmd/ and pkg/providers/remote)..."
	@grep -v "cmd/platform-spec" coverage.out | grep -v "pkg/providers/remote" > coverage-filtered.out
	@COVERAGE=$$(go tool cover -func=coverage-filtered.out | grep total: | awk '{print $$3}' | sed 's/%//'); \
	echo "Current coverage: $${COVERAGE}%"; \
	echo "Required coverage: 80%"; \
	if [ $$(echo "$${COVERAGE} < 80" | bc) -eq 1 ]; then \
		echo "❌ FAIL: Coverage $${COVERAGE}% is below minimum 80%"; \
		exit 1; \
	else \
		echo "✅ PASS: Coverage $${COVERAGE}% meets minimum 80%"; \
	fi
	@rm -f coverage-filtered.out

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
