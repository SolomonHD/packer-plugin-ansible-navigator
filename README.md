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

## Quick Start

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

#### Example 1: Simple Playbook (via `play` block)

```hcl
provisioner "ansible-navigator" {
  play {
    target = "site.yml"
  }
}
```

#### Example 2: Using role FQDNs + `requirements_file`

```hcl
provisioner "ansible-navigator" {
  # Unified requirements file for both roles + collections
  requirements_file = "./requirements.yml"

  play {
    target = "community.general.docker_container"
  }
  play {
    target = "ansible.posix.firewalld"
  }
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
  navigator_config {
    mode = "json"
  }
  structured_logging = true
  log_output_path = "./logs/deployment.json"
  verbose_task_output = true
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
    requirements_file = "./requirements.yml"

    play {
      name   = "Build Container App"
      target = "containers.docker.build_app"
      extra_vars = {
        app_name      = "webapp"
        build_version = "${var.build_number}"
      }
    }
  }
}
```

### 2. Cloud VM Configuration

```hcl
provisioner "ansible-navigator" {
  # Use specific execution environment
  navigator_config {
    execution_environment {
      enabled = true
      image = "quay.io/ansible/creator-ee:latest"
    }
  }
  
  groups = ["webservers", "database"]
  inventory_file = "inventory/cloud.ini"

  play {
    target = "cloud-init.yml"
  }
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
| [🔄 Migration Guide](docs/MIGRATION.md) | **⚠️ Migrate from deprecated options** |
| [🐛 Troubleshooting](docs/TROUBLESHOOTING.md) | Common issues and solutions |
| [📊 JSON Logging](docs/JSON_LOGGING.md) | Structured logging for automation |
| [🎭 Collection Plays](docs/UNIFIED_PLAYS.md) | Using Ansible Collection plays |

## 🛠️ Key Features

### Canonical Configuration (ordered plays + optional dependencies + optional ansible.cfg)

```hcl
provisioner "ansible-navigator" {
  # Optional: install roles + collections before running any plays
  requirements_file = "./requirements.yml"

  # Recommended: use navigator_config for execution environments and ansible.cfg
  navigator_config {
    execution_environment {
      enabled = true
      image = "quay.io/ansible/creator-ee:latest"
    }
    ansible_config {
      defaults {
        remote_tmp = "/tmp/.ansible/tmp"
      }
    }
  }

  # Required: one or more ordered play blocks
  play {
    name   = "Base configuration"
    target = "site.yml" # playbook path (.yml/.yaml)
  }

  play {
    name   = "Install Docker"
    target = "geerlingguy.docker" # role FQDN
    become = true
  }
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
# Recommended: Use navigator_config for execution environments
navigator_config {
  execution_environment {
    enabled = true
    image = "quay.io/ansible/creator-ee:latest"
  }
}

# Or custom environments
navigator_config {
  execution_environment {
    enabled = true
    image = "myregistry.io/ansible-ee:custom"
  }
}
```

### Automatic ansible.cfg generation and collections mounting (execution environments)

When using execution environments, the plugin automatically:

1. **Sets safe temp directories** to avoid `/.ansible/tmp` permission errors
2. **Mounts collections as volumes** so they're accessible inside the container
3. **Configures `ANSIBLE_COLLECTIONS_PATH`** to point to the mounted collections

**Recommended approach using navigator_config:**

```hcl
provisioner "ansible-navigator" {
  # Install collections before running plays
  requirements_file = "./requirements.yml"
  
  navigator_config {
    execution_environment {
      enabled = true
      image = "quay.io/ansible/creator-ee:latest"
    }
    ansible_config {
      defaults {
        remote_tmp = "/tmp/.ansible/tmp"
      }
    }
  }
  
  play {
    # Collections are automatically mounted and accessible
    target = "community.general.some_role"
  }
}
```

**What happens automatically:**

- Collections installed to `~/.packer.d/ansible_collections_cache/ansible_collections` are mounted read-only into the container at `/tmp/.packer_ansible/collections`
- `ANSIBLE_COLLECTIONS_PATH` is set inside the container to `/tmp/.packer_ansible/collections`
- Temp directories are configured to use `/tmp/.ansible/tmp` to avoid permission errors

> **⚠️ BREAKING CHANGE (v4.0.0):** The following options have been REMOVED: `execution_environment`, `ansible_cfg`, `ansible_env_vars`, `ansible_ssh_extra_args`, `extra_arguments`, `navigator_mode`, `roles_path`, `collections_path`, `galaxy_command`. Use `navigator_config` instead. See [Migration Guide](docs/MIGRATION.md) for upgrade instructions.

### Modern Navigator Configuration (navigator_config)

For advanced users, the `navigator_config` option provides direct control over ansible-navigator settings via a generated `ansible-navigator.yml` file. This is the recommended approach for ansible-navigator v3+ and provides the most reliable configuration experience:

```hcl
provisioner "ansible-navigator" {
  navigator_config {
    mode = "stdout"

    # Additional top-level ansible-navigator settings (v3.x)
    format                    = "yaml" # or "json"
    time_zone                  = "America/New_York"
    inventory_columns          = ["name", "address"]
    collection_doc_cache_path  = "/tmp/collection-doc-cache"

    color {
      enable = true
      osc4   = true
    }

    editor {
      command = "vim"
      console = true
    }

    images {
      details = ["everything"]
    }
    execution_environment {
      enabled = true
      image = "quay.io/ansible/creator-ee:latest"
      pull_policy = "missing"

      # New (v3.x EE schema parity): container engine + runtime/pull options
      container_engine  = "podman" # or "docker" / "auto"
      container_options = ["--net=host"]
      pull_arguments    = ["--tls-verify=false"]
    }
    ansible_config {
      defaults {
        remote_tmp = "/tmp/.ansible/tmp"
        host_key_checking = false
      }
    }
  }
  
  play {
    target = "site.yml"
  }
}
```

**Key benefits of navigator_config:**

- Aligns with ansible-navigator v3+ best practices
- Reliably controls execution environment behavior  
- Automatically sets safe defaults for EE temp directories when `execution-environment.enabled = true`
- Single source of truth for ansible-navigator settings

**Precedence:** When both `navigator_config` and legacy options (like `execution_environment`, `navigator_mode`) are present, `navigator_config` takes precedence.

## 🚦 Quick Reference

### Essential Configuration Options

| Option | Description | Example |
|--------|-------------|---------|
| `play` | Play block configuration (repeatable) | See [Collection Plays](docs/UNIFIED_PLAYS.md) |
| `requirements_file` | Install roles + collections from a unified requirements file | `"./requirements.yml"` |
| `navigator_config` | Modern declarative ansible-navigator.yml configuration (recommended for v3+) | See example above |
| `inventory_file` | Ansible inventory | `"./inventory/hosts"` |

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

**Current Version:** 4.1.0
**Status:** Production Ready
**Maintained by:** SolomonHD

---

Made with ❤️ for the Ansible and Packer communities
