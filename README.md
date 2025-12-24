# platform-spec

Infrastructure testing and verification tool.

## Installation

### Download pre-built binary

**macOS (ARM64)**
```bash
curl -L https://github.com/neilfarmer/platform-spec/releases/download/v0.0.1/platform-spec_0.0.1_darwin_arm64.zip -o platform-spec.zip
unzip platform-spec.zip
sudo mv platform-spec /usr/local/bin/platform-spec
rm platform-spec.zip
```

**Linux (AMD64)**
```bash
curl -L https://github.com/neilfarmer/platform-spec/releases/download/v0.0.1/platform-spec_0.0.1_linux_amd64.tar.gz | tar xz
sudo mv platform-spec /usr/local/bin/platform-spec
```

**Note**: Replace `v0.0.1` with the desired version, or check the [releases page](https://github.com/neilfarmer/platform-spec/releases) for the latest version.

### Build from source

```bash
# Using make
make build

# Or manually
go build -o dist/platform-spec ./cmd/platform-spec
```

## Usage

```bash
# SSH testing
dist/platform-spec test ssh -i ~/.ssh/key.pem ubuntu@192.168.1.100 spec.yaml

# With custom port
dist/platform-spec test ssh -i ~/.ssh/key.pem -p 2222 ubuntu@host spec.yaml

# Multiple spec files
dist/platform-spec test ssh -i ~/.ssh/key.pem ubuntu@host spec1.yaml spec2.yaml

# JSON output
dist/platform-spec test ssh -i ~/.ssh/key.pem ubuntu@host spec.yaml -o json

# Dry run
dist/platform-spec test ssh -i ~/.ssh/key.pem ubuntu@host spec.yaml --dry-run
```

## Providers

- **SSH**: Connect to Linux systems via SSH (in development)
- **AWS**: AWS infrastructure testing (planned)
- **OpenStack**: OpenStack infrastructure testing (planned)
