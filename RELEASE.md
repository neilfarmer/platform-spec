# Release Process

Releases are automated using conventional commits and release-please.

## Commit Format

- `feat: description` - new feature (bumps minor version)
- `fix: description` - bug fix (bumps patch version)
- `feat!: description` - breaking change (bumps major version)
- `docs:`, `test:`, `chore:`, `refactor:`, `perf:` - other changes

## Creating a Release

1. Commit to main with conventional commit messages
2. release-please will create a release PR automatically
3. Review and merge the PR
4. Binaries will be built and attached to the GitHub release

## First Release (v0.0.1)

```bash
git add .
git commit -m "chore: initial release"
git push origin main
git tag v0.0.1
git push origin v0.0.1
```

## Testing release build locally

```bash
make release-build
ls -lh dist/
```

## Manual release (if needed)

```bash
# Build
make release-build

# Create tarballs
cd dist
tar -czf platform-spec-darwin-arm64.tar.gz platform-spec-darwin-arm64
tar -czf platform-spec-linux-amd64.tar.gz platform-spec-linux-amd64

# Upload to GitHub release page manually
```
