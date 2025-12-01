# 📦 Installation Guide

This guide covers all methods to install the Packer Plugin for Ansible Navigator.

## Prerequisites

Before installing the plugin, ensure you have:

1. **Packer** >= 1.7.0 installed ([Download Packer](https://www.packer.io/downloads))
2. **Ansible Navigator** installed and in your PATH
3. **Go** >= 1.25.3 (only for building from source)

### Installing Ansible Navigator

```bash
# Using pip (recommended)
pip install ansible-navigator

# Using pipx for isolated environment
pipx install ansible-navigator

# Verify installation
ansible-navigator --version
```

## Installation Methods

### Method 1: Using `packer init` (Recommended) ✨

This is the simplest and recommended installation method.

1. Add the plugin requirement to your Packer template:

```hcl
packer {
  required_plugins {
    ansible-navigator = {
      version = ">= 1.0.0"
      source  = "github.com/solomonhd/ansible-navigator"
    }
  }
}
```

2. Initialize Packer to download and install the plugin:

```bash
packer init your-template.pkr.hcl
```

That's it! Packer will automatically download and install the correct version for your platform.

### Method 2: GitHub Releases 📥

Download pre-built binaries from the [GitHub Releases](https://github.com/solomonhd/packer-plugin-ansible-navigator/releases) page.

#### Linux (AMD64)

```bash
# Set version
VERSION="1.1.0"
PLATFORM="linux_amd64"

# Download
wget "https://github.com/solomonhd/packer-plugin-ansible-navigator/releases/download/v${VERSION}/packer-plugin-ansible-navigator_v${VERSION}_${PLATFORM}.zip"

# Extract
unzip "packer-plugin-ansible-navigator_v${VERSION}_${PLATFORM}.zip"

# Install to Packer plugins directory
mkdir -p ~/.packer.d/plugins
mv packer-plugin-ansible-navigator ~/.packer.d/plugins/
chmod +x ~/.packer.d/plugins/packer-plugin-ansible-navigator

# Verify installation
packer plugins installed
```

#### macOS (Intel)

```bash
VERSION="1.1.0"
PLATFORM="darwin_amd64"

curl -LO "https://github.com/solomonhd/packer-plugin-ansible-navigator/releases/download/v${VERSION}/packer-plugin-ansible-navigator_v${VERSION}_${PLATFORM}.zip"
unzip "packer-plugin-ansible-navigator_v${VERSION}_${PLATFORM}.zip"
mkdir -p ~/.packer.d/plugins
mv packer-plugin-ansible-navigator ~/.packer.d/plugins/
chmod +x ~/.packer.d/plugins/packer-plugin-ansible-navigator
```

#### macOS (Apple Silicon)

```bash
VERSION="1.1.0"
PLATFORM="darwin_arm64"

curl -LO "https://github.com/solomonhd/packer-plugin-ansible-navigator/releases/download/v${VERSION}/packer-plugin-ansible-navigator_v${VERSION}_${PLATFORM}.zip"
unzip "packer-plugin-ansible-navigator_v${VERSION}_${PLATFORM}.zip"
mkdir -p ~/.packer.d/plugins
mv packer-plugin-ansible-navigator ~/.packer.d/plugins/
chmod +x ~/.packer.d/plugins/packer-plugin-ansible-navigator
```

#### Windows

```powershell
# Set version
$VERSION = "1.1.0"
$PLATFORM = "windows_amd64"

# Download
Invoke-WebRequest -Uri "https://github.com/solomonhd/packer-plugin-ansible-navigator/releases/download/v$VERSION/packer-plugin-ansible-navigator_v${VERSION}_${PLATFORM}.zip" -OutFile "packer-plugin-ansible-navigator.zip"

# Extract
Expand-Archive -Path "packer-plugin-ansible-navigator.zip" -DestinationPath "."

# Install (create directory if needed)
$pluginDir = "$env:APPDATA\packer.d\plugins"
New-Item -ItemType Directory -Force -Path $pluginDir
Move-Item -Path "packer-plugin-ansible-navigator.exe" -Destination "$pluginDir\"

# Verify
packer plugins installed
```

### Method 3: Development Build 🔧

For testing latest features or contributing to development:

```bash
# Clone repository
git clone https://github.com/solomonhd/packer-plugin-ansible-navigator.git
cd packer-plugin-ansible-navigator

# Checkout specific branch/tag (optional)
git checkout develop  # or a specific tag like v1.1.0

# Build and install for development
make dev

# This will:
# 1. Build the plugin with development version tag
# 2. Install it to your local Packer plugins directory
# 3. Register it with Packer
```

### Method 4: Building from Source 🛠️

For complete control over the build process:

```bash
# Clone repository
git clone https://github.com/solomonhd/packer-plugin-ansible-navigator.git
cd packer-plugin-ansible-navigator

# Install dependencies
go mod download

# Build the plugin
go build -o packer-plugin-ansible-navigator

# Or build with version information
go build -ldflags="-X github.com/solomonhd/packer-plugin-ansible-navigator/version.Version=1.1.0" \
         -o packer-plugin-ansible-navigator

# Install manually
mkdir -p ~/.packer.d/plugins
mv packer-plugin-ansible-navigator ~/.packer.d/plugins/
chmod +x ~/.packer.d/plugins/packer-plugin-ansible-navigator
```

### Method 5: Using Go Install 🚀

If you have Go installed and configured:

```bash
# Install directly using go
go install github.com/solomonhd/packer-plugin-ansible-navigator@latest

# Copy from Go bin to Packer plugins
cp $(go env GOPATH)/bin/packer-plugin-ansible-navigator ~/.packer.d/plugins/
```

## Plugin Directory Structure

After installation, your Packer plugins directory should look like:

```
~/.packer.d/plugins/
├── packer-plugin-ansible-navigator
└── ... (other plugins)
```

On Windows:
```
%APPDATA%\packer.d\plugins\
├── packer-plugin-ansible-navigator.exe
└── ... (other plugins)
```

## Verifying Installation

### Check Plugin is Installed

```bash
# List all installed plugins
packer plugins installed

# Should show:
# github.com/solomonhd/packer-plugin-ansible-navigator
```

### Test Basic Functionality

Create a test template `test.pkr.hcl`:

```hcl
packer {
  required_plugins {
    ansible-navigator = {
      version = ">= 1.0.0"
      source  = "github.com/solomonhd/ansible-navigator"
    }
  }
}

source "null" "test" {
  communicator = "none"
}

build {
  sources = ["source.null.test"]
  
  provisioner "ansible-navigator" {
    playbook_file = "test-playbook.yml"
  }
}
```

Validate the template:
```bash
packer validate test.pkr.hcl
```

## Upgrading the Plugin

### Using `packer init`

Simply update the version constraint in your template:

```hcl
packer {
  required_plugins {
    ansible-navigator = {
      version = ">= 1.1.0"  # Updated version
      source  = "github.com/solomonhd/ansible-navigator"
    }
  }
}
```

Then run:
```bash
packer init -upgrade your-template.pkr.hcl
```

### Manual Upgrade

1. Download the new version from GitHub Releases
2. Replace the existing plugin file in `~/.packer.d/plugins/`
3. Verify with `packer plugins installed`

## Uninstalling the Plugin

### Remove Plugin File

```bash
# Linux/macOS
rm ~/.packer.d/plugins/packer-plugin-ansible-navigator

# Windows (PowerShell)
Remove-Item "$env:APPDATA\packer.d\plugins\packer-plugin-ansible-navigator.exe"
```

### Clean Plugin Cache (Optional)

```bash
# Remove cached plugin metadata
rm -rf ~/.config/packer/plugins/
```

## Troubleshooting Installation

### Plugin Not Found

If Packer can't find the plugin:

1. Verify the plugin is in the correct directory:
   ```bash
   ls -la ~/.packer.d/plugins/packer-plugin-ansible-navigator
   ```

2. Check file permissions (must be executable):
   ```bash
   chmod +x ~/.packer.d/plugins/packer-plugin-ansible-navigator
   ```

3. Ensure plugin name follows convention:
   - Must start with `packer-plugin-`
   - Binary name must match: `packer-plugin-ansible-navigator`

### Version Conflicts

If you see version conflict errors:

1. Remove all versions of the plugin:
   ```bash
   rm ~/.packer.d/plugins/packer-plugin-ansible-navigator*
   ```

2. Reinstall the desired version using `packer init`

### Platform Compatibility

Ensure you download the correct binary for your platform:

| Platform | Architecture | File Suffix |
|----------|-------------|-------------|
| Linux | x86_64 | `linux_amd64` |
| Linux | ARM64 | `linux_arm64` |
| macOS | Intel | `darwin_amd64` |
| macOS | Apple Silicon | `darwin_arm64` |
| Windows | x86_64 | `windows_amd64` |

### Go Version Issues

If building from source fails:

1. Check Go version:
   ```bash
   go version
   # Should be >= 1.25.3
   ```

2. Update Go if needed:
   ```bash
   # Using Go's official installer
   go install golang.org/dl/go1.25.3@latest
   go1.25.3 download
   ```

## Environment-Specific Setup

### Docker Environments

When using with Docker-based builds:

```bash
# Ensure Docker is running
docker info

# Pull execution environment ahead of time
docker pull quay.io/ansible/creator-ee:latest
```

### CI/CD Pipelines

For automated environments:

```yaml
# GitHub Actions example
- name: Setup Packer
  uses: hashicorp/setup-packer@main
  with:
    version: latest

- name: Install Plugin
  run: packer init config.pkr.hcl
```

### Air-Gapped Environments

For offline installations:

1. Download the plugin binary on a connected machine
2. Transfer to the air-gapped environment
3. Place in `~/.packer.d/plugins/`
4. Ensure execution permissions are set

## Getting Help

If you encounter issues:

1. Check [Troubleshooting Guide](TROUBLESHOOTING.md)
2. Search [GitHub Issues](https://github.com/solomonhd/packer-plugin-ansible-navigator/issues)
3. Ask in [GitHub Discussions](https://github.com/solomonhd/packer-plugin-ansible-navigator/discussions)
4. File a bug report with:
   - Packer version (`packer version`)
   - Plugin version
   - Error messages
   - Minimal reproduction template

---

[← Back to README](../README.md) | [Configuration Guide →](CONFIGURATION.md)