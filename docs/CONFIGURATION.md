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
- `extra_vars` (map(string), optional)
- `vars_files` (list(string), optional)
- `tags` (list(string), optional)
- `skip_tags` (list(string), optional; remote provisioner only if supported by your version)
- `become` (bool, optional)
- `become_user` (string, optional; remote provisioner only if supported by your version)

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

## Navigator configuration: `navigator_config` (optional, recommended for ansible-navigator v3+)

`navigator_config` is a map that maps directly to the `ansible-navigator.yml` configuration schema. When set, the provisioner generates a temporary `ansible-navigator.yml` file and sets `ANSIBLE_NAVIGATOR_CONFIG` environment variable.

```hcl
provisioner "ansible-navigator" {
  navigator_config = {
    mode = "stdout"
    execution-environment = {
      enabled = true
      image = "quay.io/ansible/creator-ee:latest"
      pull-policy = "missing"
    }
    ansible = {
      config = {
        defaults = {
          remote_tmp = "/tmp/.ansible/tmp"
          local_tmp = "/tmp/.ansible-local"
          host_key_checking = "False"
        }
      }
    }
  }

  play { target = "site.yml" }
}
```

**Key benefits:**

- Aligns with ansible-navigator v3+ best practices
- Reliably controls execution environment container behavior
- Automatically sets safe EE temp directory defaults when `execution-environment.enabled = true`
- Single source of truth for ansible-navigator settings

**Automatic EE defaults:** When `execution-environment.enabled = true` and temp directory settings are not explicitly provided, the plugin automatically adds:

- `ansible.config.defaults.remote_tmp = "/tmp/.ansible/tmp"`
- `ansible.config.defaults.local_tmp = "/tmp/.ansible-local"`
- `execution-environment.environment-variables.ANSIBLE_REMOTE_TMP`
- `execution-environment.environment-variables.ANSIBLE_LOCAL_TMP`

This prevents "Permission denied: /.ansible/tmp" errors in EE containers.

## Command and PATH handling

- `command` must be the ansible-navigator executable **only** (no embedded args). Example: `"ansible-navigator"`, `"/usr/local/bin/ansible-navigator"`.
- `ansible_navigator_path` prepends directories to `PATH` when locating/running ansible-navigator.

## Remote provisioner (SSH-based) inventory and connection options

The `ansible-navigator` provisioner supports inventory generation and optional proxy behavior. Common options include:

- `inventory_file`, `inventory_directory`, `inventory_file_template`
- `groups`, `empty_groups`, `host_alias`, `limit`
- `use_proxy`, `local_port`, `ansible_proxy_bind_address`, `ansible_proxy_host`
- `ssh_host_key_file`, `ssh_authorized_key_file`, `sftp_command`
- `skip_version_check`, `version_check_timeout`

---

[← Back to docs index](README.md)
