# Configuration Reference

> **⚠️ BREAKING CHANGE:** The following configuration options have been REMOVED: `execution_environment`, `ansible_cfg`, `ansible_env_vars`, `ansible_ssh_extra_args`, `extra_arguments`, `navigator_mode`.
>
> **⚠️ BREAKING CHANGE (dependencies):** The dependency-install options have been simplified:
>
> - `roles_cache_dir` → `roles_path`
> - `collections_cache_dir` → `collections_path`
> - `force_update` and `galaxy_force_install` removed (use `galaxy_force` / `galaxy_force_with_deps`)
> - Added: `galaxy_command`, `galaxy_args`
>
> See [Migration Guide](MIGRATION.md) for upgrade instructions.

This document describes the supported configuration surface.

## Core concept: ordered `play {}` blocks (required)

You must define **one or more** `play { ... }` blocks. Plays execute in declaration order.

Each play requires a `target`, which is either:

- a playbook path ending in `.yml` / `.yaml` (example: `"site.yml"`)
- a role FQDN (example: `"namespace.collection.role"`)

```hcl
provisioner "ansible-navigator" {
  play {
    name   = "Base"
    target = "site.yml"
  }

  play {
    name   = "Install Docker"
    target = "geerlingguy.docker"
    become = true
  }
}
```

### Play fields

- `name` (string, optional)
- `target` (string, required)
- `extra_args` (list(string), optional; passed verbatim to `ansible-navigator run` for that play)
- `extra_vars` (map(string), optional)
- `vars_files` (list(string), optional)
- `tags` (list(string), optional)
- `skip_tags` (list(string), optional; remote provisioner only if supported by your version)
- `become` (bool, optional)
- `become_user` (string, optional; remote provisioner only if supported by your version)

Example:

```hcl
provisioner "ansible-navigator" {
  play {
    target     = "site.yml"
    extra_args = ["--check", "--diff"]
  }
}
```

## Dependency installation: `requirements_file` (optional)

To install roles + collections before executing plays, set `requirements_file`.

```hcl
provisioner "ansible-navigator" {
  requirements_file = "./requirements.yml"

  play { target = "site.yml" }
}
```

Example `requirements.yml`:

```yaml
---
collections:
  - name: community.general
    version: ">=7.0.0"

roles:
  - name: geerlingguy.docker
    version: "6.1.0"
```

Related options:

- `roles_path` (string)
- `collections_path` (string)
- `offline_mode` (bool)
- `galaxy_force` (bool)
- `galaxy_force_with_deps` (bool)
- `galaxy_command` (string; defaults to `ansible-galaxy`)
- `galaxy_args` (list(string))

Example with explicit install destinations and Galaxy overrides:

```hcl
provisioner "ansible-navigator" {
  requirements_file = "./requirements.yml"

  roles_path       = "./.ansible/roles"
  collections_path = "./.ansible/collections"

  galaxy_force           = true
  galaxy_force_with_deps = true

  galaxy_command = "ansible-galaxy"
  galaxy_args    = ["--ignore-certs"]

  play { target = "site.yml" }
}
```

## Execution environment options

- `keep_going` (bool; continue running remaining plays after failures)

## Logging

- `structured_logging` (bool; effective when `navigator_config.mode = "json"`)
- `log_output_path` (string; write a summary JSON file)
- `verbose_task_output` (bool)

**Plugin debug output (no separate option):** set `navigator_config.logging.level = "debug"` (case-insensitive).

This single setting controls both:

- ansible-navigator logging behavior, and
- the plugin’s additional diagnostic output (messages prefixed with `[DEBUG]`).

## Navigator configuration: `navigator_config` (optional, recommended for ansible-navigator v3+)

`navigator_config` is an HCL block that maps to a typed configuration struct. When set, the provisioner generates a temporary `ansible-navigator.yml` file and sets `ANSIBLE_NAVIGATOR_CONFIG`.

When you configure `navigator_config.ansible_config` section blocks, the plugin generates an **ansible.cfg** (INI) file and references it from the generated `ansible-navigator.yml` via `ansible.config.path`.

Supported section blocks (initial coverage):

- `defaults`
- `ssh_connection`
- `privilege_escalation`
- `persistent_connection`
- `inventory`
- `paramiko_connection`
- `colors`
- `diff`
- `galaxy`

These blocks are **mutually exclusive** with `navigator_config.ansible_config.config` (path override). If you set `config`, you must not set any section blocks.

Reference: Ansible config file options are documented at <https://docs.ansible.com/ansible/latest/reference_appendices/config.html>.

`navigator_config.logging.level` also controls plugin debug output: when set to `"debug"`, the plugin emits a small set of additional diagnostic messages prefixed with `[DEBUG]`.

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

    # When set to "debug", enables additional plugin diagnostics (prefixed with [DEBUG])
    logging {
      level = "debug"
    }

    execution_environment {
      enabled     = true
      image       = "quay.io/ansible/creator-ee:latest"
      pull_policy = "missing"
    }

    # Configure Ansible temp directories (written to generated ansible.cfg)
    ansible_config {
      defaults {
        remote_tmp = "/tmp/.ansible/tmp"
        local_tmp  = "/tmp/.ansible-local"
      }
    }
  }

  play { target = "site.yml" }
}
```

### Supported additional top-level settings

The following top-level `navigator_config` keys are supported and will be emitted into the generated `ansible-navigator.yml` under the `ansible-navigator:` root:

- `format` (string)
- `time_zone` (string)
- `inventory_columns` (list(string))
- `collection_doc_cache_path` (string)
- `color { enable, osc4 }`
- `editor { command, console }`
- `images { details }`

**Key benefits:**

- Aligns with ansible-navigator v3+ best practices
- Reliably controls execution environment container behavior
- Automatically sets safe EE temp directory defaults when `execution-environment.enabled = true`
- Single source of truth for ansible-navigator settings

**Automatic EE defaults:** When `execution_environment.enabled = true` and temp directory settings are not explicitly provided, the plugin automatically adds:

- `ansible_config.defaults.remote_tmp = "/tmp/.ansible/tmp"`
- `ansible_config.defaults.local_tmp = "/tmp/.ansible-local"`
- `execution_environment.environment_variables.set.ANSIBLE_REMOTE_TMP`
- `execution_environment.environment_variables.set.ANSIBLE_LOCAL_TMP`

This prevents "Permission denied: /.ansible/tmp" errors in EE containers.

## Command and PATH handling

- `command` must be the ansible-navigator executable **only** (no embedded args). Example: `"ansible-navigator"`, `"/usr/local/bin/ansible-navigator"`.
- `ansible_navigator_path` prepends directories to `PATH` when locating/running ansible-navigator.

## Remote provisioner (SSH-based) inventory and connection options

The `ansible-navigator` provisioner supports inventory generation and optional connectivity modes. Common options include:

- `inventory_file`, `inventory_directory`, `inventory_file_template`
- `groups`, `empty_groups`, `host_alias`, `limit`
- `use_proxy`, `local_port`, `ansible_proxy_bind_address`, `ansible_proxy_host`
- `ssh_host_key_file`, `ssh_authorized_key_file`, `sftp_command`
- `skip_version_check`, `version_check_timeout`

### SSH Tunnel Mode (Bastion/Jump Host)

SSH tunnel mode provides an alternative to the Packer SSH proxy adapter for scenarios where direct SSH tunneling through a bastion or jump host is required.

#### When to Use SSH Tunnel Mode

Choose SSH tunnel mode over the default proxy adapter when:

- **WSL2 execution environments**: Container-to-host networking issues prevent the proxy adapter from working reliably
- **Docker on Windows**: Networking unreliability with the proxy adapter
- **Air-gapped targets**: Target instances only accessible through a bastion host
- **Security policies**: Organizational requirements mandate all SSH connections route through a centralized jump host
- **Network segmentation**: Target network is isolated and requires bastion connectivity

Choose the default proxy adapter (`use_proxy`) when:

- Running Packer on Linux/macOS (non-WSL2)
- No bastion/jump host infrastructure exists
- Direct network connectivity to targets
- Simpler configuration is preferred

**IMPORTANT**: `ssh_tunnel_mode` and `use_proxy` are **mutually exclusive** - you cannot enable both simultaneously.

#### Configuration Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `ssh_tunnel_mode` | bool | No | `false` | Enable SSH tunnel mode. When `true`, establishes SSH tunnels through a bastion to reach targets. **Mutually exclusive with `use_proxy`.** |
| `bastion_host` | string | Yes when `ssh_tunnel_mode=true` | - | Hostname or IP address of the bastion/jump host |
| `bastion_port` | int | No | `22` | SSH port for bastion connection. Must be between 1-65535. |
| `bastion_user` | string | Yes when `ssh_tunnel_mode=true` | - | SSH username for bastion authentication |
| `bastion_private_key_file` | string | **Either this OR `bastion_password`** | - | Path to SSH private key for bastion auth. Supports `~` for home directory expansion. |
| `bastion_password` | string | **Either this OR `bastion_private_key_file`** | - | Password for bastion authentication. **Use variables, not hardcoded strings.** |

#### Architecture Comparison

**Proxy Adapter Mode (default):**

```
┌──────────────┐                  ┌─────────────────┐
│              │   Packer SSH     │                 │
│  Packer Host │ ◄──────────────► │  Target Host    │
│              │  (direct conn)   │                 │
└──────────────┘                  └─────────────────┘
       │                                   ▲
       │                                   │
       │         Proxy adapter forwards    │
       │         SSH traffic from          │
       └───────────ansible-navigator───────┘
```

**SSH Tunnel Mode:**

```
┌──────────────┐       SSH         ┌──────────────┐       SSH        ┌─────────────────┐
│              │      Tunnel       │              │     Connection    │                 │
│  Packer Host │ ◄───────────────► │ Bastion Host │ ◄──────────────► │  Target Host    │
│              │   (to bastion)    │              │  (bastion→target) │                 │
└──────────────┘                   └──────────────┘                   └─────────────────┘
       │                                                                        ▲
       │                                                                        │
       │              ansible-navigator connects via tunnel endpoint            │
       └────────────────────────────────────────────────────────────────────────┘
                           (localhost:random_port → bastion → target)
```

**Execution Environment Container Differences:**

When using `execution_environment.enabled = true`, the connectivity differences become more apparent:

**Proxy Adapter in EE Container:**

- Container attempts to reach Packer host's proxy endpoint
- WSL2/Docker on Windows: Often fails due to networking issues between container and host
- Requires container→host network route to be reliable

**SSH Tunnel in EE Container:**

- Plugin establishes tunnel **before** starting ansible-navigator
- Container connects to tunnel endpoint on localhost
- No container→host networking issues
- Tunnel remains active for all plays

#### When SSH Tunnel Mode is Required

SSH tunnel mode is **required** in these specific scenarios:

1. **WSL2 with execution environments**: When running Packer in WSL2 with `execution_environment.enabled = true`, the proxy adapter often fails due to container-to-host networking limitations. The SSH tunnel mode establishes connectivity before the container starts, avoiding these issues.

2. **Docker Desktop on Windows**: Similar to WSL2, Docker networking on Windows can be unreliable when containers need to reach the host's proxy endpoint.

3. **Air-gapped infrastructure**: When target instances have no direct network route from the Packer host and all SSH connections must route through a bastion.

4. **Compliance requirements**: When security policies mandate that all SSH connections to production infrastructure route through audited jump hosts.

#### Examples

##### Example 1: AWS EC2 via Bastion (Key Authentication)

```hcl
source "amazon-ebs" "example" {
  # ... AWS source configuration ...
}

build {
  provisioner "ansible-navigator" {
    # Enable SSH tunnel mode (disables proxy adapter)
    ssh_tunnel_mode = true
    
    # Bastion host configuration
    bastion_host             = "bastion.aws.example.com"
    bastion_port             = 22  # Default, can be omitted
    bastion_user             = "ec2-user"
    bastion_private_key_file = "~/.ssh/bastion-key.pem"  # ~ expands to home directory
    
    # Execution environment with tunnel mode (recommended for WSL2)
    navigator_config {
      execution_environment {
        enabled     = true
        image       = "quay.io/ansible/creator-ee:latest"
        pull_policy = "missing"
      }
    }
    
    # Your plays
    play {
      target = "site.yml"
    }
  }
}
```

##### Example 2: Lab Environment (Password Authentication)

```hcl
# IMPORTANT: Define password as a variable, never hardcode sensitive values
variable "bastion_pass" {
  type      = string
  sensitive = true
}

build {
  provisioner "ansible-navigator" {
    ssh_tunnel_mode  = true
    bastion_host     = "jumphost.lab.example.com"
    bastion_port     = 2222
    bastion_user     = "deploy"
    bastion_password = var.bastion_pass  # Use variable for password
    
    play {
      target = "configure.yml"
    }
  }
}
```

Run with:

```bash
packer build -var="bastion_pass=YourPassword" template.pkr.hcl
```

##### Example 3: WSL2 with Execution Environment (Troubleshooting)

If you encounter "connection refused" or "network unreachable" errors with the proxy adapter in WSL2:

```hcl
build {
  provisioner "ansible-navigator" {
    # Switch from proxy to tunnel mode
    # use_proxy = true  ← REMOVE THIS (mutually exclusive)
    
    ssh_tunnel_mode          = true
    bastion_host             = "bastion.example.com"
    bastion_user             = "ubuntu"
    bastion_private_key_file = "~/.ssh/id_ed25519"
    
    navigator_config {
      execution_environment {
        enabled = true
        image   = "quay.io/ansible/creator-ee:latest"
      }
    }
    
    play {
      target = "myplaybook.yml"
    }
  }
}
```

#### Important Notes

1. **Home directory expansion**: File paths using `~` are automatically expanded to your home directory:

   ```hcl
   bastion_private_key_file = "~/.ssh/id_rsa"
   # Expands to: /home/username/.ssh/id_rsa
   ```

2. **Password security**: Always use Packer variables for passwords, never hardcode them:

   ```hcl
   # ❌ BAD - Hardcoded password
   bastion_password = "MyPassword123"
   
   # ✅ GOOD - Use sensitive variable
   bastion_password = var.bastion_pass
   ```

3. **Key file permissions**: SSH private keys must have restrictive permissions (see [TROUBLESHOOTING.md](TROUBLESHOOTING.md#key-file-permissions))

4. **Mutual exclusivity**: Cannot use both `ssh_tunnel_mode = true` and `use_proxy = true` - the plugin will reject this configuration with an error.

5. **Target authentication**: The SSH tunnel provides network connectivity to the target, but Ansible still requires valid credentials for the target host (separate from bastion credentials). Configure target credentials using standard Ansible methods (SSH keys, passwords, etc.).

6. **Multiple targets**: A single tunnel through the bastion can be reused for multiple target hosts if they're all accessible from the bastion.

---

[← Back to docs index](README.md)
