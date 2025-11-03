# ⚙️ Configuration Reference

Complete configuration options for the Packer Plugin Ansible Navigator.

## Table of Contents

- [Core Options](#core-options)
- [Playbook Configuration](#playbook-configuration)
- [Collection Management](#collection-management)
- [Inventory Options](#inventory-options)
- [Execution Environment](#execution-environment)
- [Communication Settings](#communication-settings)
- [Advanced Options](#advanced-options)
- [Environment Variables](#environment-variables)
- [JSON Logging Options](#json-logging-options)

## Core Options

### playbook_file

**Type:** `string`  
**Required:** Either this or `plays` must be specified  
**Conflicts with:** `plays`

Path to the Ansible playbook file.

```hcl
provisioner "ansible-navigator" {
  playbook_file = "site.yml"
}
```

### plays

**Type:** `[]object`
**Required:** Either this or `playbook_file` must be specified
**Conflicts with:** `playbook_file`

Array of play configuration objects. Each play can have its own variables, tags, and settings.

**Play Object Fields:**
- `name` (string): Display name for the play
- `target` (string): **Required** - Either a playbook path or collection play FQDN
- `extra_vars` (map[string]string): Variables specific to this play
- `vars_files` ([]string): Variable files to load for this play
- `tags` ([]string): Tags to execute for this play
- `become` (bool): Use privilege escalation for this play

```hcl
provisioner "ansible-navigator" {
  plays = [
    {
      name = "Setup Server"
      target = "namespace.collection.play_name"
      extra_vars = {
        environment = "production"
        region = "us-east-1"
      }
    },
    {
      name = "Configure Services"
      target = "community.general.setup_server"
      vars_files = ["vars/services.yml"]
      become = true
    }
  ]
}
```

## Playbook Configuration

### extra_arguments

**Type:** `[]string`  
**Default:** `[]`

Additional arguments to pass to ansible-navigator.

```hcl
provisioner "ansible-navigator" {
  extra_arguments = [
    "--extra-vars", "environment=production",
    "--extra-vars", "version=2.0.0",
    "--forks", "10",
    "--timeout", "30"
  ]
}
```

### ansible_env_vars

**Type:** `[]string`  
**Default:** `[]`

Environment variables to set during playbook execution.

```hcl
provisioner "ansible-navigator" {
  ansible_env_vars = [
    "ANSIBLE_HOST_KEY_CHECKING=False",
    "ANSIBLE_SSH_ARGS='-o ForwardAgent=yes'",
    "ANSIBLE_TIMEOUT=30",
    "ANSIBLE_GATHERING=smart"
  ]
}
```

### user

**Type:** `string`  
**Default:** Current user

The user to use for the connection. This defaults to the user Packer is running as.

```hcl
provisioner "ansible-navigator" {
  user = "ansible"
}
```

## Collection Management

### collections

**Type:** `[]string`  
**Default:** `[]`  
**Conflicts with:** `requirements_file`

List of Ansible collections to install before running plays.

```hcl
provisioner "ansible-navigator" {
  collections = [
    # From Galaxy
    "community.general:>=5.0.0",
    "ansible.posix:1.5.4",
    
    # From Git
    "git+https://github.com/org/collection.git,main",
    
    # From local path
    "myorg.mycollection@/local/path/to/collection"
  ]
}
```

### requirements_file

**Type:** `string`  
**Conflicts with:** `collections`

Path to Ansible requirements file (requirements.yml).

```hcl
provisioner "ansible-navigator" {
  requirements_file = "./requirements.yml"
}
```

### galaxy_file

**Type:** `string`  
**Deprecated:** Use `requirements_file` instead

Legacy option for Galaxy requirements. Still supported for backward compatibility.

```hcl
provisioner "ansible-navigator" {
  galaxy_file = "./requirements.yml"
}
```

### collections_cache_dir

**Type:** `string`  
**Default:** `~/.ansible/collections`

Directory to cache downloaded collections.

```hcl
provisioner "ansible-navigator" {
  collections_cache_dir = "~/.packer.d/ansible_collections"
}
```

### collections_force_update

**Type:** `bool`  
**Default:** `false`

Force update of collections even if they exist in cache.

```hcl
provisioner "ansible-navigator" {
  collections_force_update = true
}
```

### collections_offline

**Type:** `bool`  
**Default:** `false`

Use only locally cached collections (for air-gapped environments).

```hcl
provisioner "ansible-navigator" {
  collections_offline = true
  collections_cache_dir = "/mnt/shared/collections"
}
```

### galaxy_command

**Type:** `string`  
**Default:** `ansible-galaxy`

Override the ansible-galaxy command (useful for custom installations).

```hcl
provisioner "ansible-navigator" {
  galaxy_command = "/usr/local/bin/ansible-galaxy"
}
```

## Inventory Options

### inventory_file

**Type:** `string`

Path to the Ansible inventory file.

```hcl
provisioner "ansible-navigator" {
  inventory_file = "./inventory/hosts.ini"
}
```

### inventory_directory

**Type:** `string`

Path to the Ansible inventory directory.

```hcl
provisioner "ansible-navigator" {
  inventory_directory = "./inventory/"
}
```

### groups

**Type:** `[]string`  
**Default:** `[]`

List of group names to add the host to in the Ansible inventory.

```hcl
provisioner "ansible-navigator" {
  groups = ["webservers", "production", "us-east-1"]
}
```

### empty_groups

**Type:** `[]string`  
**Default:** `[]`

List of group names to create in the inventory even if empty.

```hcl
provisioner "ansible-navigator" {
  empty_groups = ["dbservers", "cache", "monitoring"]
}
```

### host_alias

**Type:** `string`

Alias for the host in the Ansible inventory.

```hcl
provisioner "ansible-navigator" {
  host_alias = "web-server-01"
}
```

### limit

**Type:** `string`

Limit playbook execution to specific hosts/groups.

```hcl
provisioner "ansible-navigator" {
  limit = "webservers:&production"
}
```

## Execution Environment

### execution_environment

**Type:** `string`  
**Default:** `quay.io/ansible/creator-ee:latest`

Container image to use as the execution environment.

```hcl
provisioner "ansible-navigator" {
  execution_environment = "quay.io/ansible/creator-ee:latest"
}
```

### work_dir

**Type:** `string`  
**Default:** Current directory

Working directory for ansible-navigator.

```hcl
provisioner "ansible-navigator" {
  work_dir = "/tmp/ansible-work"
}
```

### navigator_mode

**Type:** `string`  
**Default:** `stdout`  
**Options:** `stdout`, `interactive`, `json`

Output mode for ansible-navigator.

```hcl
provisioner "ansible-navigator" {
  navigator_mode = "json"  # For structured logging
}
```

### keep_going

**Type:** `bool`  
**Default:** `false`

Continue executing remaining plays even if one fails.

```hcl
provisioner "ansible-navigator" {
  keep_going = true
}
```

## Communication Settings

### use_proxy

**Type:** `bool`  
**Default:** `true`

Use the proxy settings from the Packer communicator.

```hcl
provisioner "ansible-navigator" {
  use_proxy = false
}
```

### local_port

**Type:** `uint`

Local port to use for SSH connections. If not specified, Packer will choose a random available port.

```hcl
provisioner "ansible-navigator" {
  local_port = 2222
}
```

### ssh_host_key_file

**Type:** `string`

Path to the SSH host key file.

```hcl
provisioner "ansible-navigator" {
  ssh_host_key_file = "~/.ssh/known_hosts"
}
```

### ssh_authorized_key_file

**Type:** `string`

Path to the authorized key file to use for SSH.

```hcl
provisioner "ansible-navigator" {
  ssh_authorized_key_file = "~/.ssh/authorized_keys"
}
```

### sftp_command

**Type:** `string`

Override the SFTP command.

```hcl
provisioner "ansible-navigator" {
  sftp_command = "/usr/libexec/openssh/sftp-server -e"
}
```

## Advanced Options

### command

**Type:** `string`  
**Default:** `ansible-navigator run`

Override the ansible-navigator command.

```hcl
provisioner "ansible-navigator" {
  command = "/custom/path/ansible-navigator run"
}
```

### skip_version_check

**Type:** `bool`  
**Default:** `false`

Skip ansible-navigator version verification.

```hcl
provisioner "ansible-navigator" {
  skip_version_check = true
}
```

### pause_before

**Type:** `duration string`  
**Default:** `0`

Duration to pause before running the provisioner.

```hcl
provisioner "ansible-navigator" {
  pause_before = "10s"
}
```

### max_retries

**Type:** `int`  
**Default:** `0`

Maximum number of retries if the provisioner fails.

```hcl
provisioner "ansible-navigator" {
  max_retries = 3
}
```

### timeout

**Type:** `duration string`

Overall timeout for the provisioner.

```hcl
provisioner "ansible-navigator" {
  timeout = "30m"
}
```

## Environment Variables

### System Environment Variables

The plugin respects these system environment variables:

```bash
# Ansible configuration
export ANSIBLE_CONFIG=/path/to/ansible.cfg
export ANSIBLE_ROLES_PATH=/custom/roles
export ANSIBLE_COLLECTIONS_PATH=/custom/collections

# Execution environment
export ANSIBLE_NAVIGATOR_IMAGE=custom/ee:latest
export ANSIBLE_NAVIGATOR_PULL_POLICY=never

# Proxy settings
export HTTP_PROXY=http://proxy.example.com:8080
export HTTPS_PROXY=http://proxy.example.com:8080
export NO_PROXY=localhost,127.0.0.1
```

### Setting Variables in Configuration

```hcl
provisioner "ansible-navigator" {
  ansible_env_vars = [
    "ANSIBLE_STDOUT_CALLBACK=json",
    "ANSIBLE_LOAD_CALLBACK_PLUGINS=True",
    "ANSIBLE_FORCE_COLOR=True",
    "ANSIBLE_DEPRECATION_WARNINGS=False"
  ]
}
```

## JSON Logging Options

### structured_logging

**Type:** `bool`  
**Default:** `false`

Enable structured JSON event parsing.

```hcl
provisioner "ansible-navigator" {
  structured_logging = true
  navigator_mode = "json"
}
```

### log_output_path

**Type:** `string`

Path to write structured JSON summary file.

```hcl
provisioner "ansible-navigator" {
  log_output_path = "./logs/ansible-summary.json"
}
```

### verbose_task_output

**Type:** `bool`  
**Default:** `false`

Include detailed task output in logs.

```hcl
provisioner "ansible-navigator" {
  verbose_task_output = true
}
```

## Configuration Examples

### Minimal Configuration

```hcl
provisioner "ansible-navigator" {
  playbook_file = "site.yml"
}
```

### Production Configuration

```hcl
provisioner "ansible-navigator" {
  plays = [
    {
      name = "Security Hardening"
      target = "baseline.security.harden"
      become = true
    },
    {
      name = "Application Configuration"
      target = "app.deployment.configure"
      extra_vars = {
        deployment_type = "production"
      }
    }
  ]
  
  requirements_file = "./requirements.yml"
  
  execution_environment = "registry.company.com/ansible-ee:production"
  
  groups = ["production", "web-tier"]
  
  ansible_env_vars = [
    "ANSIBLE_HOST_KEY_CHECKING=False",
    "ANSIBLE_GATHERING=smart",
    "ANSIBLE_CACHE_PLUGIN=jsonfile",
    "ANSIBLE_CACHE_PLUGIN_CONNECTION=/tmp/ansible-cache"
  ]
  
  extra_arguments = [
    "--extra-vars", "environment=production",
    "--extra-vars", "@secrets.yml",
    "--vault-password-file", ".vault-pass",
    "--forks", "20"
  ]
  
  navigator_mode = "json"
  structured_logging = true
  log_output_path = "./logs/deployment.json"
  
  timeout = "1h"
  max_retries = 2
}
```

### Development Configuration

```hcl
provisioner "ansible-navigator" {
  playbook_file = "dev-playbook.yml"
  
  collections = [
    "myorg.custom@../local-collection"
  ]
  
  collections_force_update = true
  
  extra_arguments = [
    "--extra-vars", "debug=true",
    "--verbose"
  ]
  
  keep_going = true  # Continue on errors for debugging
}
```

### Air-Gapped Configuration

```hcl
provisioner "ansible-navigator" {
  plays = [
    {
      name = "Offline Deployment"
      target = "app.deploy.offline"
      extra_vars = {
        offline_mode = "true"
      }
    }
  ]
  
  collections_offline = true
  collections_cache_dir = "/mnt/shared/offline-collections"
  
  execution_environment = "localhost:5000/ansible-ee:latest"
  
  ansible_env_vars = [
    "ANSIBLE_GALAXY_SERVER_LIST=",  # Disable Galaxy
    "ANSIBLE_GALAXY_IGNORE=True"
  ]
}
```

## Configuration Validation

The plugin validates configuration at runtime and will fail with clear error messages:

### Mutual Exclusivity Errors

```
Error: You may specify only one of `playbook_file` or `plays`
```

### Required Field Errors

```
Error: Either `playbook_file` or `plays` must be defined
```

### Invalid Value Errors

```
Error: navigator_mode must be one of: stdout, interactive, json
```

## Best Practices

1. **Use Execution Environments**: Always specify a consistent execution environment for reproducible builds
2. **Enable Structured Logging**: Use JSON logging for CI/CD pipelines
3. **Cache Collections**: Use `collections_cache_dir` to speed up builds
4. **Set Timeouts**: Define reasonable timeouts to prevent hanging builds
5. **Use Requirements Files**: For complex dependencies, use `requirements_file` instead of inline `collections`
6. **Version Pin Collections**: Always specify versions for production builds
7. **Separate Environments**: Use different configurations for development vs production

## Debugging Configuration

Enable verbose output to debug configuration issues:

```hcl
provisioner "ansible-navigator" {
  playbook_file = "debug.yml"
  
  extra_arguments = [
    "--verbose",  # or -v, -vv, -vvv for more verbosity
    "--check"     # Dry run mode
  ]
  
  ansible_env_vars = [
    "ANSIBLE_DEBUG=True",
    "ANSIBLE_VERBOSITY=4"
  ]
}
```

---

[← Installation Guide](INSTALLATION.md) | [Examples Gallery →](EXAMPLES.md)