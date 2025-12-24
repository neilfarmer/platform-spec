# platform-spec

A pluggable infrastructure testing framework that validates system state across multiple platforms using declarative YAML specifications.

## Installation

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

See [releases page](https://github.com/neilfarmer/platform-spec/releases) for other versions.

## Quick Start

Create a spec file `mytest.yaml`:

```yaml
version: "1.0"
tests:
  packages:
    - name: "Docker installed"
      packages: [docker-ce]
      state: present
  files:
    - name: "Config directory exists"
      path: /etc/myapp
      type: directory
```

Run the tests:

```bash
platform-spec test ssh ubuntu@myhost mytest.yaml
```

See [USAGE.md](USAGE.md) for complete documentation.

## Roadmap

### Phase 1: Core Framework ✅

- SSH provider with agent support
- Package assertions (dpkg, rpm, apk)
- File/directory assertions
- Human-readable output

### Phase 2: Enhanced Assertions (In Progress)

- Service status testing
- User/group testing
- Custom command assertions
- File content matching

### Phase 3: Advanced Features

- JSON and JUnit output formats
- Parallel execution
- Variable substitution

### Phase 4: Cloud Providers

- AWS provider (EC2, S3, IAM, RDS)
- OpenStack provider
- Kubernetes provider
