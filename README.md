# platform-spec

A pluggable infrastructure testing framework that validates system state across multiple platforms using declarative YAML specifications.

## Installation

**Quick Install (macOS & Linux)**

```bash
# Install to /usr/local/bin (default, may require sudo)
curl -sSL https://raw.githubusercontent.com/neilfarmer/platform-spec/main/scripts/install.sh | bash

# Install to custom directory (e.g., ~/.bin)
curl -sSL https://raw.githubusercontent.com/neilfarmer/platform-spec/main/scripts/install.sh | bash -s -- --dir ~/.bin
```

The install script automatically detects your OS and architecture and downloads the latest release.

**Note:** If installing to a custom directory, make sure it's in your `PATH`:
```bash
export PATH="$PATH:$HOME/.bin"  # Add to ~/.bashrc or ~/.zshrc
```

**Manual Installation**

Download the appropriate binary for your platform from the [releases page](https://github.com/neilfarmer/platform-spec/releases/latest):

<details>
<summary>macOS (ARM64)</summary>

```bash
# Replace VERSION with the latest version (e.g., 0.3.1)
VERSION=0.3.1
curl -L https://github.com/neilfarmer/platform-spec/releases/download/v${VERSION}/platform-spec_${VERSION}_darwin_arm64.zip -o platform-spec.zip
unzip platform-spec.zip
sudo mv platform-spec /usr/local/bin/platform-spec
rm platform-spec.zip
```
</details>

<details>
<summary>Linux (AMD64)</summary>

```bash
# Replace VERSION with the latest version (e.g., 0.3.1)
VERSION=0.3.1
curl -L https://github.com/neilfarmer/platform-spec/releases/download/v${VERSION}/platform-spec_${VERSION}_linux_amd64.tar.gz | tar xz
sudo mv platform-spec /usr/local/bin/platform-spec
```
</details>

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
# Test local system
platform-spec test local mytest.yaml

# Test remote system via SSH
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
