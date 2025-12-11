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
**Required:** Either this or `play` blocks must be specified
**Conflicts with:** `play`

Path to the Ansible playbook file.

```hcl
provisioner "ansible-navigator" {
  playbook_file = "site.yml"
}
```

### play

**Type:** `block` (repeatable)
**Required:** Either this or `playbook_file` must be specified
**Conflicts with:** `playbook_file`

Play block configuration. Multiple plays are defined as repeated `play { }` blocks.

**Play Block Fields:**

- `name` (string): Display name for the play
- `target` (string): **Required** - Either a playbook path or collection play FQDN
- `extra_vars` (map[string]string): Variables specific to this play
- `vars_files` ([]string): Variable files to load for this play
- `tags` ([]string): Tags to execute for this play
- `become` (bool): Use privilege escalation for this play

```hcl
provisioner "ansible-navigator" {
  play {
    name   = "Setup Server"
    target = "namespace.collection.play_name"
    extra_vars = {
      environment = "production"
      region      = "us-east-1"
    }
  }
  
  play {
    name       = "Configure Services"
    target     = "community.general.setup_server"
    vars_files = ["vars/services.yml"]
    become     = true
  }
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
**Default:** _unset (ansible-navigator uses its default)_

The container image to use as the execution environment for ansible-navigator.
Specifies which containerized environment runs the Ansible playbooks.
When unset, ansible-navigator uses its default execution environment.

**Under the hood:** When you specify `execution_environment = "image:tag"`, the plugin generates the CLI flags `--ee true --eei image:tag` for ansible-navigator v3+. This correctly separates the boolean "enable execution environment" flag (`--ee`) from the "execution environment image" flag (`--eei`).

```hcl
provisioner "ansible-navigator" {
  # Use official Ansible execution environment
  execution_environment = "quay.io/ansible/creator-ee:latest"
  # The plugin will generate: --ee true --eei quay.io/ansible/creator-ee:latest
}
```

**Examples:**

```hcl
# Official Ansible EE (latest version)
execution_environment = "quay.io/ansible/creator-ee:latest"

# Official Ansible EE (specific version)
execution_environment = "quay.io/ansible/creator-ee:v0.21.0"

# Custom execution environment
execution_environment = "myregistry.io/custom-ansible-ee:v1.0"

# Custom execution environment with digest
execution_environment = "myregistry.io/ansible-ee@sha256:abc123..."
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
**Default:** `ansible-navigator`

Override the ansible-navigator executable path or name.

**Important:** This field accepts only the executable name or path, without additional arguments. The `run` subcommand is automatically added. Use `extra_arguments` for additional flags.

```hcl
provisioner "ansible-navigator" {
  # Use custom executable path
  command = "/custom/path/ansible-navigator"
  
  # Or use HOME-relative path
  command = "~/bin/ansible-navigator"
}
```

**Invalid (will cause validation error):**

```hcl
provisioner "ansible-navigator" {
  command = "ansible-navigator run"  # ❌ Arguments not allowed
  command = "ansible-navigator --mode json"  # ❌ Flags not allowed
}
```

**Migration from older versions:** If you previously had `command = "ansible-navigator run"`, change it to:

```hcl
command = "ansible-navigator"  # run is now added automatically
```

### ansible_navigator_path

**Type:** `[]string`
**Default:** `[]` (uses system PATH)

Additional directories to prepend to PATH when locating and running ansible-navigator. Useful when ansible-navigator is installed in a non-standard location.

All paths support HOME expansion (`~` becomes your home directory).

```hcl
provisioner "ansible-navigator" {
  ansible_navigator_path = [
    "~/bin",                    # Expands to /home/user/bin
    "/opt/ansible/bin",         # Absolute path
    "~/.local/bin"              # Expands to /home/user/.local/bin
  ]
}
```

**Use Cases:**

- **Custom installation locations:**

  ```hcl
  ansible_navigator_path = ["/opt/custom-ansible/bin"]
  ```

- **Virtual environments:**

  ```hcl
  ansible_navigator_path = ["~/venvs/ansible/bin"]
  ```

- **Multiple possible locations:**

  ```hcl
  ansible_navigator_path = ["~/bin", "/usr/local/bin", "/opt/ansible/bin"]
  ```

**Note:** Directories are searched in the order specified, before the system PATH.

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

## HOME Path Expansion

Many configuration fields support HOME path expansion using the `~` prefix:

**Supported fields:**

- `command` (when it looks like a path)
- `playbook_file`
- `inventory_file`
- `inventory_directory`
- `galaxy_file`
- `requirements_file`
- `ssh_host_key_file`
- `ssh_authorized_key_file`
- `roles_path`
- `collections_path`
- `collections_cache_dir`
- `roles_cache_dir`
- `work_dir`
- `ansible_navigator_path` entries
- Play `target` fields (when pointing to playbook files)
- Play `vars_files` entries

**Examples:**

```hcl
provisioner "ansible-navigator" {
  # Playbook in HOME
  playbook_file = "~/ansible/site.yml"
  
  # Inventory in HOME
  inventory_file = "~/ansible/inventory/hosts"
  
  # Custom cache directories
  collections_cache_dir = "~/.ansible/collections"
  roles_cache_dir = "~/.ansible/roles"
  
  # Working directory
  work_dir = "~/ansible-work"
  
  # Play with HOME-relative playbook
  play {
    target = "~/playbooks/deploy.yml"
    vars_files = ["~/vars/production.yml"]
  }
}
```

**Expansion Rules:**

- `~` → Your home directory (e.g., `/home/user`)
- `~/subdir` → Home with subdirectory (e.g., `/home/user/subdir`)
- `~otheruser/` → Preserved unchanged (multi-user homes not supported)
- Absolute paths (`/path`) → Unchanged
- Relative paths (`./path`) → Unchanged

## Configuration Examples

### Minimal Configuration

```hcl
provisioner "ansible-navigator" {
  playbook_file = "site.yml"
}
```

### Production Configuration with Custom Paths

```hcl
provisioner "ansible-navigator" {
  # Use ansible-navigator from custom location
  command = "~/custom-ansible/bin/ansible-navigator"
  ansible_navigator_path = ["~/custom-ansible/bin", "/opt/ansible/bin"]
  
  play {
    name   = "Security Hardening"
    target = "baseline.security.harden"
    become = true
  }
  
  play {
    name   = "Application Configuration"
    target = "app.deployment.configure"
    extra_vars = {
      deployment_type = "production"
    }
  }
  
  requirements_file = "~/ansible/requirements.yml"
  
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
  log_output_path = "~/logs/deployment.json"
  
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
  play {
    name   = "Offline Deployment"
    target = "app.deploy.offline"
    extra_vars = {
      offline_mode = "true"
    }
  }
  
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
Error: You may specify only one of `playbook_file`/`playbook_files` or `play` blocks
```

### Required Field Errors

```
Error: Either `playbook_file`/`playbook_files` or `play` blocks must be defined
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
