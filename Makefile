.PHONY: build clean test install release-build

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
	go test ./...

install: build
	cp dist/platform-spec $(GOPATH)/bin/platform-spec
