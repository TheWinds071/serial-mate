# Serial Mate AUR Packages

This directory contains PKGBUILD files for Arch Linux AUR (Arch User Repository).

## Available Packages

### serial-mate
The stable release package that builds from tagged releases.

**Installation:**
```bash
git clone https://aur.archlinux.org/serial-mate.git
cd serial-mate
makepkg -si
```

Or using an AUR helper like `yay`:
```bash
yay -S serial-mate
```

### serial-mate-git
The development package that builds from the latest git commit.

**Installation:**
```bash
git clone https://aur.archlinux.org/serial-mate-git.git
cd serial-mate-git
makepkg -si
```

Or using an AUR helper like `yay`:
```bash
yay -S serial-mate-git
```

## For Maintainers

### Updating PKGBUILD for New Release

1. Update version and checksums:
```bash
./update-aur.sh 1.0.0
```

2. Test the PKGBUILD locally:
```bash
makepkg -si
```

3. Generate .SRCINFO:
```bash
makepkg --printsrcinfo > .SRCINFO
```

4. Commit and push to AUR:
```bash
git clone ssh://aur@aur.archlinux.org/serial-mate.git aur-repo
cd aur-repo
cp ../PKGBUILD .
cp ../.SRCINFO .
git add PKGBUILD .SRCINFO
git commit -m "Update to version 1.0.0"
git push
```

### Manual Update Process

If you prefer to update manually:

1. Edit `PKGBUILD` and update `pkgver`
2. Download the source tarball:
   ```bash
   wget https://github.com/TheWinds071/serial-mate/archive/v1.0.0.tar.gz
   ```
3. Calculate SHA256:
   ```bash
   sha256sum v1.0.0.tar.gz
   ```
4. Update `sha256sums` in PKGBUILD
5. Generate .SRCINFO:
   ```bash
   makepkg --printsrcinfo > .SRCINFO
   ```

## Dependencies

### Runtime Dependencies
- `gtk3`: GTK 3 toolkit
- `webkit2gtk`: WebKit rendering engine

### Build Dependencies
- `go>=1.21`: Go programming language
- `nodejs>=20`: Node.js runtime
- `npm`: Node package manager
- `git`: Version control (for -git package only)

## Building from PKGBUILD

```bash
# Install build dependencies
sudo pacman -S go nodejs npm gtk3 webkit2gtk

# Build and install
cd packaging/aur
makepkg -si
```

## Notes

- The package installs to `/usr/bin/serial-mate`
- Desktop file is installed to `/usr/share/applications/`
- Icon is installed to `/usr/share/icons/hicolor/512x512/apps/`
- The build process uses Wails to compile the application
- Build time may take several minutes depending on your system

## Troubleshooting

### Build Fails with "wails: command not found"

The PKGBUILD automatically installs Wails during the prepare stage. If you see this error, try:
```bash
makepkg -C
```

### WebKit Version Issues

The package depends on `webkit2gtk` which should provide the appropriate version for your Arch system. If you encounter issues, ensure your system is up to date:
```bash
sudo pacman -Syu
```

## Links

- **Upstream**: https://github.com/TheWinds071/serial-mate
- **AUR Package**: https://aur.archlinux.org/packages/serial-mate
- **AUR Git Package**: https://aur.archlinux.org/packages/serial-mate-git
