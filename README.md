# Packer Plugin Ansible Navigator

The `ansible-navigator` plugin extends HashiCorp [Packer](https://www.packer.io) to enable provisioning of images using **Ansible Navigator** (`ansible-navigator run`), extending beyond traditional Ansible playbook execution. This is an independent project developed under `github.com/SolomonHD/packer-plugin-ansible-navigator`.

## Features

### Dual Invocation Mode

Support for two mutually exclusive configuration paths:

**Option A: Traditional playbook file**
```hcl
provisioner "ansible-navigator" {
  playbook_file = "site.yml"
}
```

**Option B: Collection plays**
```hcl
provisioner "ansible-navigator" {
  plays = [
    "integration.portainer.migrate_node",
    "acme.firewall.configure_rules"
  ]
}
```

### Enhanced Error Handling

- **Per-play failure reporting**: When running multiple plays, clearly identifies which play failed
- **Detailed dependency checking**: Ensures `ansible-navigator` binary is available before execution
- **Rich UI integration**: All output is properly surfaced through both console and programmatic interfaces

### Modern Ansible Integration

- Uses `ansible-navigator run` instead of `ansible-playbook`
- Full SSH and WinRM communicator compatibility
- Integration with HashiCorp's Packer Plugin SDK

## Installation

### Using the `packer init` command (Recommended)

Starting from version 1.7, Packer supports a new `packer init` command allowing automatic installation of Packer plugins. Read the [Packer documentation](https://www.packer.io/docs/commands/init) for more information.

To install this plugin, copy and paste this code into your Packer configuration. Then, run [`packer init`](https://www.packer.io/docs/commands/init).

```hcl
packer {
  required_plugins {
    ansible-navigator = {
      version = ">= 0.1.0"
      source  = "github.com/SolomonHD/ansible-navigator"
    }
  }
}
```

### Installing from GitHub Releases

Download the appropriate binary for your platform from the [GitHub Releases](https://github.com/SolomonHD/packer-plugin-ansible-navigator/releases) page.

```bash
# Example for Linux AMD64
VERSION="0.1.0"
PLATFORM="linux_amd64"
wget "https://github.com/SolomonHD/packer-plugin-ansible-navigator/releases/download/v${VERSION}/packer-plugin-ansible-navigator_v${VERSION}_${PLATFORM}.zip"
unzip "packer-plugin-ansible-navigator_v${VERSION}_${PLATFORM}.zip"

# Install to Packer plugins directory
mkdir -p ~/.packer.d/plugins
mv packer-plugin-ansible-navigator ~/.packer.d/plugins/
chmod +x ~/.packer.d/plugins/packer-plugin-ansible-navigator
```

### Installing from Git Repository

For development or to use the latest unreleased changes, you can install directly from the Git repository:

```bash
# Clone the repository
git clone https://github.com/SolomonHD/packer-plugin-ansible-navigator.git
cd packer-plugin-ansible-navigator

# Checkout a specific version (optional)
git checkout v0.1.0

# Build and install
make dev
```

The `make dev` command builds the plugin and installs it to your local Packer plugins directory.

### Building from Source

If you prefer to build the plugin from sources manually:

```bash
# Clone the repository
git clone https://github.com/SolomonHD/packer-plugin-ansible-navigator.git
cd packer-plugin-ansible-navigator

# Build the plugin
go build -o packer-plugin-ansible-navigator

# Install to Packer plugins directory
mkdir -p ~/.packer.d/plugins
mv packer-plugin-ansible-navigator ~/.packer.d/plugins/
chmod +x ~/.packer.d/plugins/packer-plugin-ansible-navigator
```

**Requirements:**
- Go 1.25.3 or later
- Git

## Configuration

### Required Options

You must specify either `playbook_file` OR `plays`, but not both:

```hcl
# Using a playbook file
provisioner "ansible-navigator" {
  playbook_file = "site.yml"
}

# Using collection plays
provisioner "ansible-navigator" {
  plays = [
    "integration.portainer.migrate_node",
    "acme.firewall.configure_rules"
  ]
}
```

### Common Options

```hcl
provisioner "ansible-navigator" {
  playbook_file = "site.yml"
  
  # Extra arguments to pass to ansible-navigator
  extra_arguments = ["--extra-vars", "Environment=production"]
  
  # Environment variables
  ansible_env_vars = [
    "ANSIBLE_HOST_KEY_CHECKING=False",
    "ANSIBLE_SSH_ARGS='-o ForwardAgent=yes'"
  ]
  
  # Galaxy requirements file
  galaxy_file = "requirements.yml"
  
  # Inventory configuration
  inventory_file = "./inventory/hosts"
  groups = ["webservers", "dbservers"]
}
```

## Advanced Examples

-> **Note:** The `collections` and `requirements_file` options are mutually exclusive. You can specify collections inline using the `collections` array, or reference a requirements file using `requirements_file`, but not both at the same time.

### Using Plays with Managed Collections

Combine collection plays with automatic collection installation:

```hcl
provisioner "ansible-navigator" {
  plays = [
    "community.general.docker_container",
    "ansible.posix.firewalld"
  ]
  
  collections = [
    "community.general:>=5.0.0",
    "ansible.posix:1.5.4"
  ]
  
  extra_arguments = [
    "--extra-vars", "container_name=myapp",
    "--extra-vars", "container_image=nginx:latest"
  ]
}
```

### Collections from Git Repositories

Pull collections directly from Git repositories:

```hcl
provisioner "ansible-navigator" {
  plays = [
    "myorg.infrastructure.deploy",
    "myorg.infrastructure.configure"
  ]
  
  collections = [
    "git+https://github.com/myorg/ansible-collection-infrastructure.git,main",
    "git+https://github.com/acme/ansible-collection-security.git,v2.1.0",
    "community.general:5.11.0"
  ]
  
  collections_cache_dir = "~/.packer.d/collections"
  
  extra_arguments = [
    "--extra-vars", "environment=production"
  ]
}
```

### Multi-Stage Deployment with Custom Collections

Deploy applications using multiple plays from a custom collection:

```hcl
provisioner "ansible-navigator" {
  plays = [
    "integration.portainer.setup_environment",
    "integration.portainer.configure_swarm",
    "integration.portainer.deploy_stack"
  ]
  
  collections = [
    "integration.portainer@/local/path/to/collection",
    "community.docker:3.4.0"
  ]
  
  collections_cache_dir = "~/.packer.d/collections"
  
  extra_arguments = [
    "--extra-vars", "stack_name=production-app",
    "--extra-vars", "replicas=3"
  ]
}
```

### Offline/Air-Gapped Environment

Use pre-cached collections in an offline environment:

```hcl
provisioner "ansible-navigator" {
  plays = [
    "acme.firewall.configure_rules",
    "acme.security.harden_system"
  ]
  
  collections = [
    "acme.firewall:2.1.0",
    "acme.security:1.5.0"
  ]
  
  collections_offline = true
  collections_cache_dir = "/mnt/shared/ansible-collections"
  
  ansible_env_vars = [
    "ANSIBLE_HOST_KEY_CHECKING=False"
  ]
}
```

### Development Workflow with Force Update

Iterate quickly during development by always updating local collections:

```hcl
provisioner "ansible-navigator" {
  plays = ["myorg.app.deploy"]
  
  collections = [
    "myorg.app@../ansible-collections/myorg-app"
  ]
  
  collections_force_update = true
  
  extra_arguments = [
    "--extra-vars", "env=development",
    "--extra-vars", "debug=true"
  ]
}
```

### Complex Production Setup

Production deployment with requirements file and multiple plays:

```hcl
provisioner "ansible-navigator" {
  requirements_file = "./requirements.yml"
  
  plays = [
    "baseline.security.harden_system",
    "baseline.monitoring.setup_agents",
    "app.webapp.deploy"
  ]
  
  collections_cache_dir = "~/.packer.d/ansible_collections_cache"
  
  ansible_env_vars = [
    "ANSIBLE_HOST_KEY_CHECKING=False",
    "ANSIBLE_GATHERING=smart"
  ]
  
  extra_arguments = [
    "--extra-vars", "environment=production",
    "--extra-vars", "version=2.5.1",
    "--extra-vars", "enable_monitoring=true"
  ]
}
```

### Mixed Playbook and Collections

Combine traditional playbook with managed collections:

```hcl
provisioner "ansible-navigator" {
  playbook_file = "site.yml"
  
  collections = [
    "community.general:5.11.0",
    "ansible.windows:2.3.0",
    "ansible.posix:1.5.4"
  ]
  
  galaxy_file = "requirements.yml"
  
  extra_arguments = ["--extra-vars", "tier=web"]
}
```

## Error Handling Examples

### Configuration Validation
```hcl
# This will fail with a clear error message:
provisioner "ansible-navigator" {
  playbook_file = "site.yml"
  plays = ["integration.portainer.migrate_node"]  # ERROR: mutual exclusivity
}
```

### Missing Dependencies
If `ansible-navigator` is not found in PATH, you'll receive:
```
Error: ansible-navigator not found in PATH. Please install it before running this provisioner.
```

### Play Execution Failure
When running multiple plays:
```
ERROR: Play 'integration.portainer.migrate_node' failed (exit code 2)
Aborting remaining plays. Check the above output for the failing play.
```

## Contributing

* If you think you've found a bug in the code or you have a question regarding the usage of this software, please reach out to us by opening an issue in this GitHub repository.
* Contributions to this project are welcome: if you want to add a feature or a fix a bug, please do so by opening a Pull Request in this GitHub repository. In case of feature contribution, we kindly ask you to open an issue to discuss it beforehand.

## License

This project is licensed under the Apache License, Version 2.0. See the [LICENSE](LICENSE) file for details.
