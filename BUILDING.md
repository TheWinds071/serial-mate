# Building Serial Mate

This guide covers how to build Serial Mate from source for different platforms.

## Prerequisites

### All Platforms
- **Go**: >= 1.21 ([Download](https://golang.org/dl/))
- **Node.js**: >= 20 ([Download](https://nodejs.org/))
- **npm**: Comes with Node.js
- **Wails CLI**: Install with `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### Platform-Specific Dependencies

#### Linux
- **GTK 3**: `libgtk-3-dev` (Debian/Ubuntu) or `gtk3-devel` (Fedora)
- **WebKitGTK**: `libwebkit2gtk-4.1-dev` or `libwebkit2gtk-4.0-dev` (Debian/Ubuntu), `webkit2gtk4.1-devel` or `webkit2gtk3-devel` (Fedora)
- **nfpm** (for packaging): [Installation guide](https://nfpm.goreleaser.com/install/)

```bash
# Ubuntu 24.04+
sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.1-dev

# Ubuntu 22.04 / Debian 12
sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.0-dev

# Fedora 41+
sudo dnf install -y gtk3-devel webkit2gtk4.1-devel

# Fedora 40
sudo dnf install -y gtk3-devel webkit2gtk3-devel

# Arch Linux
sudo pacman -S gtk3 webkit2gtk
```

#### macOS
Xcode Command Line Tools are required:
```bash
xcode-select --install
```

#### Windows
- **MinGW-w64**: Required for CGO support
- Alternatively, use MSVC Build Tools

## Quick Start

### Using Make (Recommended)

The project includes a comprehensive Makefile for easy building:

```bash
# Check dependencies
make check-deps

# Install Wails CLI if not already installed
make install-deps

# Build for current platform
make build

# Run in development mode
make dev
```

### Manual Build

```bash
# Clone the repository
git clone https://github.com/TheWinds071/serial-mate.git
cd serial-mate

# Install dependencies
go mod download
cd frontend && npm install && cd ..

# Build with Wails
wails build
```

The built application will be in `build/bin/`.

## Building for Specific Platforms

### Linux

```bash
# Build binary
make build-linux

# Or manually
wails build -platform linux/amd64 -clean
```

### Windows

```bash
# Build binary
make build-windows

# Or manually
wails build -platform windows/amd64 -clean
```

### macOS

```bash
# Build universal binary (Apple Silicon + Intel)
make build-darwin

# Or manually
wails build -platform darwin/universal -clean
```

## Creating Distribution Packages

### Debian Package (.deb)

```bash
# For systems with webkit 4.1 (Ubuntu 24.04+)
make package-deb-webkit41

# For systems with webkit 4.0 (Ubuntu 22.04, Debian 12)
make package-deb-webkit40

# Output: dist/serial-mate_VERSION_amd64.deb
```

### RPM Package (.rpm)

```bash
# For systems with webkit 4.1 (Fedora 41+)
make package-rpm-webkit41

# For systems with webkit 4.0 (Fedora 40)
make package-rpm-webkit40

# Output: dist/serial-mate-VERSION.x86_64.rpm
```

### AUR Package (Arch Linux)

```bash
# Prepare AUR package files
make package-aur

# This creates PKGBUILD files in dist/aur/
# Follow packaging/aur/README.md for publishing to AUR
```

### Windows Executable

```bash
make package-windows

# Output: dist/serial-mate-VERSION-windows-amd64.exe
```

### macOS Application Bundle

```bash
make package-macos

# Output: dist/serial-mate-VERSION-macos-universal.app.zip
```

### All Packages

Build all available packages for current platform:

```bash
make package-all
```

## Development

### Development Mode

Run the application in development mode with hot reloading:

```bash
make dev
```

### Running Tests

```bash
make test
```

### Cleaning Build Artifacts

```bash
make clean
```

## Installing Locally (Linux)

Install the application system-wide on Linux:

```bash
sudo make install
```

This will:
- Copy the binary to `/usr/bin/serial-mate`
- Install the desktop file to `/usr/share/applications/`
- Install the icon to `/usr/share/icons/hicolor/512x512/apps/`

To uninstall:

```bash
sudo make uninstall
```

## Makefile Targets Reference

Run `make help` to see all available targets:

```
General:
  help                 Display help information
  check-deps           Check if required dependencies are installed
  install-deps         Install build dependencies (Wails CLI)

Development:
  dev                  Run development server
  test                 Run tests
  clean                Clean build artifacts

Build:
  build                Build for current platform
  build-linux          Build for Linux
  build-windows        Build for Windows
  build-darwin         Build for macOS (Universal)

Packaging:
  package-deb          Build Debian package (webkit 4.1)
  package-deb-webkit40 Build Debian package (webkit 4.0)
  package-deb-webkit41 Build Debian package (webkit 4.1)
  package-rpm          Build RPM package (webkit 4.1)
  package-rpm-webkit40 Build RPM package (webkit 4.0)
  package-rpm-webkit41 Build RPM package (webkit 4.1)
  package-aur          Prepare AUR package files
  package-windows      Package Windows executable
  package-macos        Package macOS application
  package-all-linux    Build all Linux packages
  package-all          Build all packages

Installation:
  install              Install on local system (Linux only)
  uninstall            Uninstall from local system (Linux only)

Information:
  version              Display version information
```

## Troubleshooting

### "wails: command not found"

Install the Wails CLI:
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

Make sure `$GOPATH/bin` or `$HOME/go/bin` is in your PATH:
```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

### "nfpm: command not found"

Install nfpm for creating packages:
```bash
# macOS
brew install goreleaser/tap/nfpm

# Linux
echo "deb [trusted=yes] https://repo.goreleaser.com/apt/ /" | sudo tee /etc/apt/sources.list.d/goreleaser.list
sudo apt update
sudo apt install nfpm

# Or download from GitHub
go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
```

### Linux: "Package webkit2gtk was not found"

Install WebKitGTK development files:
```bash
# Ubuntu/Debian
sudo apt-get install libwebkit2gtk-4.1-dev

# Fedora
sudo dnf install webkit2gtk4.1-devel

# Arch
sudo pacman -S webkit2gtk
```

### Windows: CGO Build Errors

Ensure MinGW-w64 is installed and in your PATH, or use MSVC with CGO enabled.

### macOS: Code Signing Issues

For development builds, you can disable code signing:
```bash
wails build -skipbindings -nosyncgomod
```

For distribution, you'll need an Apple Developer account and valid certificates.

## CI/CD

The project includes GitHub Actions workflows for automated building and releasing. See `.github/workflows/release.yml` for the complete CI/CD pipeline that builds packages for all supported platforms.

## More Information

- **Wails Documentation**: https://wails.io/docs/introduction
- **Go Documentation**: https://golang.org/doc/
- **Vue 3 Documentation**: https://vuejs.org/guide/introduction.html

## Getting Help

If you encounter issues:
1. Check the [GitHub Issues](https://github.com/TheWinds071/serial-mate/issues)
2. Review the [Wails Troubleshooting Guide](https://wails.io/docs/guides/troubleshooting)
3. Open a new issue with details about your environment and the error
