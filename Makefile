.PHONY: help build clean install dev test \
	build-linux build-windows build-darwin \
	package-deb package-rpm package-aur \
	install-deps check-deps

# Version extraction from git tags
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo "0.0.0")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Application name
APP_NAME := serial-mate
BUILD_DIR := build/bin
DIST_DIR := dist

# Platform-specific settings
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
NC := \033[0m # No Color

##@ General

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

check-deps: ## Check if required dependencies are installed
	@echo "$(GREEN)Checking dependencies...$(NC)"
	@command -v go >/dev/null 2>&1 || { echo "$(RED)Error: go is not installed$(NC)"; exit 1; }
	@command -v node >/dev/null 2>&1 || { echo "$(RED)Error: node is not installed$(NC)"; exit 1; }
	@command -v npm >/dev/null 2>&1 || { echo "$(RED)Error: npm is not installed$(NC)"; exit 1; }
	@command -v wails >/dev/null 2>&1 || { echo "$(YELLOW)Warning: wails is not installed. Run 'make install-deps' to install it$(NC)"; }
	@echo "$(GREEN)All core dependencies are installed!$(NC)"

install-deps: ## Install build dependencies (Wails CLI)
	@echo "$(GREEN)Installing Wails CLI...$(NC)"
	go install github.com/wailsapp/wails/v2/cmd/wails@latest
	@echo "$(GREEN)Wails CLI installed successfully!$(NC)"

##@ Development

dev: ## Run development server
	wails dev

test: ## Run tests
	go test -v ./...

clean: ## Clean build artifacts
	@echo "$(GREEN)Cleaning build artifacts...$(NC)"
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -rf frontend/dist
	rm -rf frontend/node_modules/.vite
	@echo "$(GREEN)Clean complete!$(NC)"

##@ Build

build: check-deps ## Build for current platform
	@echo "$(GREEN)Building $(APP_NAME) v$(VERSION) for $(GOOS)/$(GOARCH)...$(NC)"
	wails build -clean

build-linux: ## Build for Linux
	@echo "$(GREEN)Building for Linux...$(NC)"
	wails build -platform linux/amd64 -clean

build-windows: ## Build for Windows
	@echo "$(GREEN)Building for Windows...$(NC)"
	wails build -platform windows/amd64 -clean

build-darwin: ## Build for macOS (Universal)
	@echo "$(GREEN)Building for macOS...$(NC)"
	wails build -platform darwin/universal -clean

##@ Packaging

package-deb-webkit40: build-linux ## Build Debian package with webkit 4.0
	@echo "$(GREEN)Creating Debian package (webkit 4.0)...$(NC)"
	@mkdir -p $(DIST_DIR)
	@command -v nfpm >/dev/null 2>&1 || { echo "$(RED)Error: nfpm is not installed. Install from https://nfpm.goreleaser.com/$(NC)"; exit 1; }
	VERSION=$(VERSION) nfpm package --packager deb --config packaging/nfpm/linux-webkit40.yaml --target $(DIST_DIR)/
	@echo "$(GREEN)Debian package created: $(DIST_DIR)/$(APP_NAME)_$(VERSION)_amd64.deb$(NC)"

package-deb-webkit41: build-linux ## Build Debian package with webkit 4.1
	@echo "$(GREEN)Creating Debian package (webkit 4.1)...$(NC)"
	@mkdir -p $(DIST_DIR)
	@command -v nfpm >/dev/null 2>&1 || { echo "$(RED)Error: nfpm is not installed. Install from https://nfpm.goreleaser.com/$(NC)"; exit 1; }
	VERSION=$(VERSION) nfpm package --packager deb --config packaging/nfpm/linux-webkit41.yaml --target $(DIST_DIR)/
	@echo "$(GREEN)Debian package created: $(DIST_DIR)/$(APP_NAME)_$(VERSION)_amd64.deb$(NC)"

package-deb: package-deb-webkit41 ## Build Debian package (default: webkit 4.1)

package-rpm-webkit40: build-linux ## Build RPM package with webkit 4.0
	@echo "$(GREEN)Creating RPM package (webkit 4.0)...$(NC)"
	@mkdir -p $(DIST_DIR)
	@command -v nfpm >/dev/null 2>&1 || { echo "$(RED)Error: nfpm is not installed. Install from https://nfpm.goreleaser.com/$(NC)"; exit 1; }
	VERSION=$(VERSION) nfpm package --packager rpm --config packaging/nfpm/linux-webkit40.yaml --target $(DIST_DIR)/
	@echo "$(GREEN)RPM package created: $(DIST_DIR)/$(APP_NAME)-$(VERSION).x86_64.rpm$(NC)"

package-rpm-webkit41: build-linux ## Build RPM package with webkit 4.1
	@echo "$(GREEN)Creating RPM package (webkit 4.1)...$(NC)"
	@mkdir -p $(DIST_DIR)
	@command -v nfpm >/dev/null 2>&1 || { echo "$(RED)Error: nfpm is not installed. Install from https://nfpm.goreleaser.com/$(NC)"; exit 1; }
	VERSION=$(VERSION) nfpm package --packager rpm --config packaging/nfpm/linux-webkit41.yaml --target $(DIST_DIR)/
	@echo "$(GREEN)RPM package created: $(DIST_DIR)/$(APP_NAME)-$(VERSION).x86_64.rpm$(NC)"

package-rpm: package-rpm-webkit41 ## Build RPM package (default: webkit 4.1)

package-aur: ## Prepare AUR package files
	@echo "$(GREEN)Preparing AUR package...$(NC)"
	@mkdir -p $(DIST_DIR)/aur
	@cp packaging/aur/PKGBUILD $(DIST_DIR)/aur/
	@cp packaging/aur/PKGBUILD-git $(DIST_DIR)/aur/
	@cd $(DIST_DIR)/aur && sed -i "s/pkgver=.*/pkgver=$(VERSION)/" PKGBUILD
	@echo "$(GREEN)AUR package files prepared in $(DIST_DIR)/aur/$(NC)"
	@echo "$(YELLOW)Note: Update sha256sums in PKGBUILD before publishing to AUR$(NC)"
	@echo "$(YELLOW)Run 'cd $(DIST_DIR)/aur && makepkg --printsrcinfo > .SRCINFO' to generate .SRCINFO$(NC)"

package-windows: build-windows ## Package Windows executable
	@echo "$(GREEN)Packaging Windows executable...$(NC)"
	@mkdir -p $(DIST_DIR)
	@cp $(BUILD_DIR)/$(APP_NAME).exe $(DIST_DIR)/$(APP_NAME)-$(VERSION)-windows-amd64.exe
	@echo "$(GREEN)Windows package created: $(DIST_DIR)/$(APP_NAME)-$(VERSION)-windows-amd64.exe$(NC)"

package-macos: build-darwin ## Package macOS application
	@echo "$(GREEN)Packaging macOS application...$(NC)"
	@mkdir -p $(DIST_DIR)
	@cd $(BUILD_DIR) && zip -r ../../$(DIST_DIR)/$(APP_NAME)-$(VERSION)-macos-universal.app.zip $(APP_NAME).app
	@echo "$(GREEN)macOS package created: $(DIST_DIR)/$(APP_NAME)-$(VERSION)-macos-universal.app.zip$(NC)"

package-all-linux: package-deb-webkit40 package-deb-webkit41 package-rpm-webkit40 package-rpm-webkit41 ## Build all Linux packages

package-all: package-deb package-rpm package-aur package-windows package-macos ## Build all packages (requires all platform builds)

##@ Installation

install: build-linux ## Install on local system (Linux only)
	@echo "$(GREEN)Installing $(APP_NAME)...$(NC)"
	sudo install -Dm755 $(BUILD_DIR)/$(APP_NAME) /usr/bin/$(APP_NAME)
	sudo install -Dm644 packaging/linux/$(APP_NAME).desktop /usr/share/applications/$(APP_NAME).desktop
	sudo install -Dm644 build/appicon.png /usr/share/icons/hicolor/512x512/apps/$(APP_NAME).png
	@echo "$(GREEN)Installation complete!$(NC)"

uninstall: ## Uninstall from local system (Linux only)
	@echo "$(GREEN)Uninstalling $(APP_NAME)...$(NC)"
	sudo rm -f /usr/bin/$(APP_NAME)
	sudo rm -f /usr/share/applications/$(APP_NAME).desktop
	sudo rm -f /usr/share/icons/hicolor/512x512/apps/$(APP_NAME).png
	@echo "$(GREEN)Uninstallation complete!$(NC)"

##@ Information

version: ## Display version information
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo "GOOS: $(GOOS)"
	@echo "GOARCH: $(GOARCH)"
