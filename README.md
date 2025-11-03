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
