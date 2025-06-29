# Release Workflow

This document describes how to create and manage releases for TAI (Terminal AI).

## Overview

TAI uses an automated release workflow powered by GitHub Actions and GoReleaser. The process is triggered by pushing Git tags and automatically builds binaries for multiple platforms, runs tests, and creates GitHub releases.

## Release Process

### 1. Prepare for Release

Before creating a release, ensure:

- [ ] All intended changes are merged to `main` branch
- [ ] Tests are passing: `make test` and `make test-race`
- [ ] Code quality checks pass: `make check`
- [ ] Update any version-specific documentation if needed

### 2. Create a Release Tag

The release process is triggered by pushing a version tag to the repository.

```bash
# Create and push a version tag
git checkout main
git pull origin main
git tag v1.0.0
git push origin v1.0.0
```

**Tag Naming Convention:**
- Use semantic versioning: `v<MAJOR>.<MINOR>.<PATCH>`
- Examples: `v1.0.0`, `v1.2.3`, `v2.0.0-rc1`
- Pre-release tags (containing `-rc`, `-alpha`, `-beta`) are automatically marked as pre-releases

### 3. Automated Release Process

Once you push a version tag, GitHub Actions automatically:

1. **Triggers Release Workflow** (`.github/workflows/release.yml`)
   - Checks out the repository with full history
   - Sets up Go 1.24.4
   - Runs the full test suite (`make test`)

2. **Runs GoReleaser** (`.goreleaser.yml`)
   - Downloads and tidies Go dependencies
   - Builds binaries for multiple platforms:
     - **Linux**: amd64, arm64
     - **macOS**: amd64, arm64
     - **Windows**: amd64 only
   - Creates platform-specific archives:
     - `.tar.gz` for Linux and macOS
     - `.zip` for Windows
   - Generates checksums file
   - Creates GitHub release with changelog

3. **Release Artifacts**
   - Binary archives for each platform
   - `checksums.txt` for verifying downloads
   - Automatic changelog from commit messages
   - Installation instructions in release notes

## Platform Support

### Supported Platforms

| OS      | Architecture | Archive Format |
|---------|-------------|----------------|
| Linux   | amd64       | tar.gz         |
| Linux   | arm64       | tar.gz         |
| macOS   | amd64       | tar.gz         |
| macOS   | arm64       | tar.gz         |
| Windows | amd64       | zip            |

### Binary Naming

Binaries follow this naming convention:
```
tai_<OS>_<ARCH>.<format>
```

Examples:
- `tai_Linux_x86_64.tar.gz`
- `tai_Darwin_arm64.tar.gz`
- `tai_Windows_x86_64.zip`

## Release Configuration

### GoReleaser Configuration

The release process is configured in `.goreleaser.yml`:

- **Builds**: Cross-platform compilation with CGO disabled
- **Archives**: Platform-appropriate compression formats
- **Changelog**: Auto-generated, excluding docs and test commits
- **Release Notes**: Custom template with installation instructions

### GitHub Actions Workflow

The workflow in `.github/workflows/release.yml`:

- **Trigger**: Tags matching `v*` pattern
- **Permissions**: `contents: write` for creating releases
- **Quality Gates**: Full test suite must pass before release
- **Go Version**: Fixed to 1.24.4 for reproducible builds

## Installation Methods

Released binaries can be installed via:

### 1. Direct Download
Download the appropriate binary from GitHub Releases and extract:

```bash
# Example for Linux amd64
curl -L -o tai.tar.gz https://github.com/adamveld12/tai/releases/download/v1.0.0/tai_Linux_x86_64.tar.gz
tar -xzf tai.tar.gz
sudo mv tai /usr/local/bin/
```

### 2. Go Install
Install directly from source using Go:

```bash
go install github.com/adamveld12/tai/cmd/tai@v1.0.0
```

### 3. Build from Source
Clone and build locally:

```bash
git clone https://github.com/adamveld12/tai.git
cd tai
git checkout v1.0.0
make build
```

## Verification

### Checksum Verification

Every release includes a `checksums.txt` file for verifying downloads:

```bash
# Download checksum file
curl -L -o checksums.txt https://github.com/adamveld12/tai/releases/download/v1.0.0/checksums.txt

# Verify your download
sha256sum -c checksums.txt --ignore-missing
```

### Testing Releases

Before announcing a release, test the binaries:

```bash
# Test basic functionality
./tai --help
./tai --version

# Test one-shot mode
echo "Hello, AI!" | ./tai "Please respond"

# Test REPL mode
./tai -repl
```

## Troubleshooting

### Common Issues

**Release workflow fails at test stage:**
- Check that all tests pass locally: `make test-race`
- Review test failures in GitHub Actions logs
- Fix issues and create a new tag

**GoReleaser build failures:**
- Verify cross-compilation works locally: `make build-all`
- Check for platform-specific code issues
- Review GoReleaser logs for dependency problems

**Missing artifacts in release:**
- Confirm `.goreleaser.yml` includes all desired platforms
- Check that builds completed successfully for all targets
- Verify archive creation didn't fail

### Manual Release Recovery

If automated release fails, you can run GoReleaser locally:

```bash
# Install GoReleaser
go install github.com/goreleaser/goreleaser@latest

# Create release (requires GITHUB_TOKEN)
export GITHUB_TOKEN="your_token"
git tag v1.0.0
goreleaser release --clean
```

## Best Practices

### Release Planning

- **Regular Releases**: Cut releases frequently to keep changes manageable
- **Semantic Versioning**: Follow semver for version numbering
- **Release Notes**: Write clear, user-focused release notes
- **Testing**: Always test releases on multiple platforms

### Tag Management

- **Annotated Tags**: Use annotated tags for releases:
  ```bash
  git tag -a v1.0.0 -m "Release version 1.0.0"
  ```
- **Tag Protection**: Consider protecting version tag patterns in GitHub
- **Cleanup**: Remove pre-release tags after final release

### Quality Assurance

- **Pre-release Testing**: Use pre-release tags for release candidates
- **Automated Testing**: Rely on CI/CD pipeline for quality gates
- **Manual Verification**: Test key workflows with release binaries
- **Rollback Plan**: Keep previous release available for quick rollback

## Release Checklist

Use this checklist for each release:

### Pre-Release
- [ ] All features/fixes merged to `main`
- [ ] Tests passing: `make test-race`
- [ ] Code quality checks pass: `make check`
- [ ] Documentation updated if needed
- [ ] Version number decided (semantic versioning)

### Release
- [ ] Create and push version tag: `git tag v1.0.0 && git push origin v1.0.0`
- [ ] Monitor GitHub Actions workflow completion
- [ ] Verify all platform binaries are built
- [ ] Check release notes are generated correctly

### Post-Release
- [ ] Test release binaries on different platforms
- [ ] Update installation documentation if needed
- [ ] Announce release (if applicable)
- [ ] Monitor for user issues/feedback

## Support

For issues with the release process:

1. Check GitHub Actions logs for workflow failures
2. Review GoReleaser documentation: https://goreleaser.com/
3. Test locally with `make build-all` and `make test-race`
4. Create an issue with release workflow problems