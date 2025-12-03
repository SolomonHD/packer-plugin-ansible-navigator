# 🚀 Packer Plugin Ansible Navigator

> Modern Ansible provisioning for HashiCorp Packer using **ansible-navigator** for containerized execution environments

[![Apache License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.23.4+-blue.svg)](.go-version)
[![Packer Plugin SDK](https://img.shields.io/badge/Packer%20Plugin%20SDK-v0.6.4-blue.svg)](go.mod)

## 🎯 Why Use This Plugin?

This plugin extends HashiCorp Packer to leverage **Ansible Navigator's** containerized execution environments, providing:

- ✅ **Containerized Ansible** - Run playbooks in isolated, reproducible environments
- ✅ **Collection-Based Workflows** - Direct execution of Ansible Collection plays
- ✅ **Enhanced Error Handling** - Clear per-play failure reporting
- ✅ **Structured Logging** - JSON event streaming for CI/CD integration
- ✅ **Modern Ansible Features** - Full compatibility with execution environments

## 📋 Requirements

- **Go:** 1.23.4+ (for development/building from source)
- **Packer:** ≥ 1.10.0 (required for plugin protocol x5 support)
- **Packer Plugin SDK:** v0.6.4+ (automatically managed via Go modules)
- **Ansible Navigator:** Latest version recommended (runtime dependency)

## � Quick Start

### Installation

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

Run `packer init` to install the plugin automatically.

> **📖 Full Installation Guide:** See [docs/INSTALLATION.md](docs/INSTALLATION.md) for all installation methods

### Basic Usage

#### Example 1: Simple Playbook

```hcl
provisioner "ansible-navigator" {
  playbook_file = "site.yml"
}
```

#### Example 2: Using Collection Plays

```hcl
provisioner "ansible-navigator" {
  play {
    target = "community.general.docker_container"
  }
  play {
    target = "ansible.posix.firewalld"
  }
  collections = [
    "community.general:>=5.0.0",
    "ansible.posix:1.5.4"
  ]
}
```

#### Example 3: Production Setup with JSON Logging

```hcl
provisioner "ansible-navigator" {
  play {
    name   = "Security Hardening"
    target = "baseline.security.harden"
    become = true
  }
  
  play {
    name   = "Application Deployment"
    target = "app.webapp.deploy"
    extra_vars = {
      environment = "production"
      version     = "${var.app_version}"
    }
  }
  
  # Enable structured logging for CI/CD
  navigator_mode = "json"
  structured_logging = true
  log_output_path = "./logs/deployment.json"
  
  # Global arguments applied to all plays
  extra_arguments = ["--verbose"]
}
```

## 🔥 Common Use Cases

### 1. Building Container Images

```hcl
source "docker" "app" {
  image = "ubuntu:22.04"
  commit = true
}

build {
  sources = ["source.docker.app"]
  
  provisioner "ansible-navigator" {
    play {
      name   = "Build Container App"
      target = "containers.docker.build_app"
      extra_vars = {
        app_name      = "webapp"
        build_version = "${var.build_number}"
      }
    }
    collections = ["community.docker:3.4.0"]
  }
}
```

### 2. Cloud VM Configuration

```hcl
provisioner "ansible-navigator" {
  playbook_file = "cloud-init.yml"
  
  # Use specific execution environment
  execution_environment = "quay.io/ansible/creator-ee:latest"
  
  groups = ["webservers", "database"]
  inventory_file = "inventory/cloud.ini"
}
```

### 3. Multi-Stage Application Deployment

```hcl
provisioner "ansible-navigator" {
  play {
    name   = "Base Configuration"
    target = "infra.base.configure"
  }
  
  play {
    name   = "Security Hardening"
    target = "infra.security.harden"
    become = true
  }
  
  play {
    name   = "Database Setup"
    target = "app.database.install"
    extra_vars = {
      db_engine  = "postgresql"
      db_version = "14"
    }
  }
  
  play {
    name       = "Application Deployment"
    target     = "app.webserver.deploy"
    vars_files = ["app_config.yml"]
  }
  
  requirements_file = "./requirements.yml"
  
  # Continue on individual play failure for debugging
  keep_going = true
}
```

## 📚 Documentation

| Document | Description |
|----------|-------------|
| [📦 Installation Guide](docs/INSTALLATION.md) | All installation methods and requirements |
| [⚙️ Configuration Reference](docs/CONFIGURATION.md) | Complete list of options and parameters |
| [🎨 Examples Gallery](docs/EXAMPLES.md) | Real-world examples and use cases |
| [🐛 Troubleshooting](docs/TROUBLESHOOTING.md) | Common issues and solutions |
| [📊 JSON Logging](docs/JSON_LOGGING.md) | Structured logging for automation |
| [🎭 Collection Plays](docs/UNIFIED_PLAYS.md) | Using Ansible Collection plays |

## 🛠️ Key Features

### Dual Invocation Mode

Choose between traditional playbooks or modern collection plays:

```hcl
# Option A: Traditional playbook
playbook_file = "site.yml"

# Option B: Collection plays (mutually exclusive)
play {
  name       = "Play Name"
  target     = "namespace.collection.play_name"
  extra_vars = {}  # Optional per-play variables
}
```

### Enhanced Error Reporting

```
✅ Running play: infra.base.configure
✅ Running play: infra.security.harden
❌ Play 'app.database.install' failed (exit code 2)
   └─ Check logs for task: "Install PostgreSQL"
   └─ Failed hosts: db-server-01, db-server-02
```

### Execution Environments

```hcl
# Use certified execution environments
execution_environment = "quay.io/ansible/creator-ee:latest"

# Or custom environments
execution_environment = "myregistry.io/ansible-ee:custom"
```

## 🚦 Quick Reference

### Essential Configuration Options

| Option | Description | Example |
|--------|-------------|---------|
| `playbook_file` | Path to Ansible playbook | `"site.yml"` |
| `play` | Play block configuration (repeatable) | See [Collection Plays](docs/UNIFIED_PLAYS.md) |
| `collections` | Collections to install | `["community.general:5.0.0"]` |
| `execution_environment` | Container image for ansible-navigator | `"quay.io/ansible/creator-ee"` |
| `inventory_file` | Ansible inventory | `"./inventory/hosts"` |
| `extra_arguments` | Additional ansible-navigator args | `["--extra-vars", "key=value"]` |

> **📖 Complete Reference:** See [docs/CONFIGURATION.md](docs/CONFIGURATION.md)

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

### Development Quick Start

```bash
# Clone and build
git clone https://github.com/solomonhd/packer-plugin-ansible-navigator.git
cd packer-plugin-ansible-navigator
make dev

# Run tests
make test

# Build for release
make build
```

## 📜 License

This project is licensed under the [Apache License 2.0](LICENSE).

## 🆘 Support

- 🐛 **Issues:** [GitHub Issues](https://github.com/solomonhd/packer-plugin-ansible-navigator/issues)
- 💬 **Discussions:** [GitHub Discussions](https://github.com/solomonhd/packer-plugin-ansible-navigator/discussions)
- 📖 **Documentation:** [Full Docs](docs/)

## 🏗️ Project Status

**Current Version:** 3.1.0
**Status:** Production Ready
**Maintained by:** SolomonHD

---

Made with ❤️ for the Ansible and Packer communities
