# Maintainer's Guide

This guide is for maintainers of Serial Mate who need to manage releases and packages.

## Release Process

### 1. Prepare Release

1. Update version numbers if needed (Wails uses git tags for versioning)
2. Update CHANGELOG or release notes
3. Ensure all tests pass
4. Ensure the build system is working correctly:
   ```bash
   make check-deps
   make build
   make test
   ```

### 2. Create Git Tag

Create and push a git tag to trigger the release workflow:

```bash
# Create annotated tag
git tag -a v1.0.0 -m "Release version 1.0.0"

# Push tag to remote
git push origin v1.0.0
```

### 3. Automated Release

The GitHub Actions workflow (`.github/workflows/release.yml`) will automatically:

1. **Build binaries** for all platforms:
   - Windows (amd64)
   - macOS (Universal - Apple Silicon + Intel)
   - Linux (amd64)

2. **Create packages**:
   - `.deb` packages for Ubuntu 22.04 (webkit 4.0) and Ubuntu 24.04 (webkit 4.1)
   - `.rpm` packages for Fedora 40 (webkit 4.0) and Fedora 41 (webkit 4.1)
   - AUR package files (`PKGBUILD` and `PKGBUILD-git`)

3. **Create GitHub Release** with:
   - All binaries and packages
   - Auto-generated changelog
   - Release notes

### 4. Publish to AUR

After the GitHub release is created, manually publish to AUR:

#### First Time Setup

```bash
# Clone AUR repositories
git clone ssh://aur@aur.archlinux.org/serial-mate.git aur-serial-mate
git clone ssh://aur@aur.archlinux.org/serial-mate-git.git aur-serial-mate-git
```

#### Update AUR Packages

For **serial-mate** (stable):

```bash
cd aur-serial-mate

# Download and extract AUR package files from GitHub release
VERSION=1.0.0
wget https://github.com/TheWinds071/serial-mate/releases/download/v${VERSION}/serial-mate-aur-${VERSION}.tar.gz
tar -xzf serial-mate-aur-${VERSION}.tar.gz
cp PKGBUILD ../aur-serial-mate/

# Generate .SRCINFO (requires Arch Linux or Arch-based system)
makepkg --printsrcinfo > .SRCINFO

# Commit and push
git add PKGBUILD .SRCINFO
git commit -m "Update to version ${VERSION}"
git push
```

For **serial-mate-git** (development):

```bash
cd aur-serial-mate-git

# Copy PKGBUILD-git
wget https://github.com/TheWinds071/serial-mate/releases/download/v${VERSION}/serial-mate-aur-${VERSION}.tar.gz
tar -xzf serial-mate-aur-${VERSION}.tar.gz
cp PKGBUILD-git ../aur-serial-mate-git/PKGBUILD

# Generate .SRCINFO
makepkg --printsrcinfo > .SRCINFO

# Commit and push
git add PKGBUILD .SRCINFO
git commit -m "Update PKGBUILD"
git push
```

### 5. Update Documentation

After release:

1. Update README if needed
2. Announce release on relevant channels
3. Update any external documentation

## Manual Package Building

### Local Package Testing

Before releasing, you can test package creation locally:

#### Debian Package

```bash
# For webkit 4.0 (Ubuntu 22.04)
make package-deb-webkit40

# For webkit 4.1 (Ubuntu 24.04+)
make package-deb-webkit41

# Test installation
sudo dpkg -i dist/serial-mate_*.deb
serial-mate --version
```

#### RPM Package

```bash
# For webkit 4.0 (Fedora 40)
make package-rpm-webkit40

# For webkit 4.1 (Fedora 41+)
make package-rpm-webkit41

# Test installation (on Fedora/RHEL)
sudo dnf install dist/serial-mate-*.rpm
serial-mate --version
```

#### AUR Package

```bash
make package-aur

# Test on Arch Linux
cd dist/aur
makepkg -si
serial-mate --version
```

## Troubleshooting

### GitHub Actions Failures

1. **Build failures**: Check the build logs in GitHub Actions
2. **Package creation failures**: Usually related to missing dependencies or incorrect nfpm configuration
3. **Release upload failures**: Check permissions and GitHub token

### AUR Issues

1. **SHA256 mismatch**: Update the checksum in PKGBUILD using `makepkg -g`
2. **Build failures**: Test locally with `makepkg -si`
3. **Dependency issues**: Verify dependencies are available in Arch repos

### Version Conflicts

The version is automatically extracted from git tags. Ensure:
- Tags follow semantic versioning (v1.0.0)
- Tags are pushed to remote
- CI has access to fetch tags (`fetch-depth: '0'` in workflow)

## Package Maintenance

### Dependencies

#### Runtime Dependencies
- **All Linux**: `gtk3`, `webkit2gtk`
- **Windows**: No external runtime dependencies (statically linked)
- **macOS**: No external runtime dependencies

#### Build Dependencies
- **Go**: >= 1.21
- **Node.js**: >= 20
- **npm**: Comes with Node.js
- **Wails**: Installed via `go install`
- **nfpm**: For creating deb/rpm packages

### Updating Dependencies

When updating project dependencies:

1. Update `go.mod` and `go.sum`:
   ```bash
   go get -u ./...
   go mod tidy
   ```

2. Update frontend dependencies:
   ```bash
   cd frontend
   npm update
   cd ..
   ```

3. Test build:
   ```bash
   make clean
   make build
   make test
   ```

4. Update package dependencies in nfpm configs if needed:
   - `packaging/nfpm/linux-webkit40.yaml`
   - `packaging/nfpm/linux-webkit41.yaml`

### WebKit Version Migration

If a new major WebKit version is released:

1. Create new nfpm config (e.g., `linux-webkit42.yaml`)
2. Add new matrix entry in `.github/workflows/release.yml`
3. Update BUILDING.md with new distribution versions
4. Update README with installation instructions

## Security

### Code Signing

#### macOS
- Requires Apple Developer account
- Configure signing in Xcode or via command line
- For CI/CD, store certificates in GitHub Secrets

#### Windows
- Requires code signing certificate
- Can use DigiCert, Sectigo, or other CAs
- For CI/CD, store certificate in GitHub Secrets

### Dependency Scanning

Regular security practices:
- Run `go mod tidy` regularly
- Use `npm audit` for frontend dependencies
- Monitor GitHub security alerts
- Update dependencies promptly

## Continuous Integration

### Workflows

1. **release.yml**: Triggered on git tags, builds and releases all packages
2. **test-build.yml**: Triggered on PRs, tests Makefile targets

### Adding New Platforms

To add support for a new platform:

1. Add build target in Makefile:
   ```makefile
   build-newplatform:
       wails build -platform newplatform/arch -clean
   ```

2. Add packaging target if needed:
   ```makefile
   package-newplatform:
       # Package commands
   ```

3. Update GitHub Actions workflow with new matrix entry

4. Update documentation (BUILDING.md, README.md)

## Resources

- [Wails Documentation](https://wails.io/docs/)
- [nfpm Documentation](https://nfpm.goreleaser.com/)
- [AUR Guidelines](https://wiki.archlinux.org/title/AUR_submission_guidelines)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)

## Getting Help

- GitHub Issues: https://github.com/TheWinds071/serial-mate/issues
- Wails Discord: https://discord.gg/BrRSWTaxRK
