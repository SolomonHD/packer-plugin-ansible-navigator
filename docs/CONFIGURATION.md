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

`navigator_config` is an HCL block that maps to a typed configuration struct.
The SSH provisioner generates a temporary `ansible-navigator.yml` on the Packer
host and sets `ANSIBLE_NAVIGATOR_CONFIG`. The on-target provisioner uses CLI
flags for mapped settings and uploads minimal YAML for pull arguments,
playbook artifacts, or collection document cache settings.

When you configure `ansible_config.defaults` and/or `ansible_config.ssh_connection`, the plugin generates an **ansible.cfg** (INI) file. The SSH provisioner references it through YAML `ansible.config.path`; the on-target provisioner uploads it and supplies its path through the existing CLI configuration.

`navigator_config.logging.level` also controls plugin debug output: when set to `"debug"`, the plugin emits a small set of additional diagnostic messages prefixed with `[DEBUG]`.

```hcl
provisioner "ansible-navigator" {
  navigator_config {
    mode = "stdout"

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

### Execution environment options

Both provisioners accept the following fields inside
`navigator_config { execution_environment { ... } }`:

| Field | HCL type | Behavior |
| --- | --- | --- |
| `enabled` | bool | Enable or disable the EE. |
| `image` | string | Container image used for the EE. |
| `container_engine` | string | Select `docker`, `podman`, or `auto`. |
| `container_options` | list(string) | Ordered arguments for container execution. |
| `pull_policy` | string | Navigator pull policy, such as `always`, `missing`, or `never`. |
| `pull_arguments` | list(string) | Ordered arguments for the selected runtime's image pull command. |
| `volume_mounts` | repeated block | Source, destination, and optional mount options. |

```hcl
navigator_config {
  mode = "stdout"
  execution_environment {
    enabled           = true
    image             = "quay.io/ansible/creator-ee:latest"
    container_engine  = "docker"
    container_options = ["--label=packer-example=true"]
    pull_policy       = "missing"
    pull_arguments    = ["--quiet"]

    volume_mounts {
      src     = "/opt/ansible/data"
      dest    = "/runner/data"
      options = "ro"
    }
    volume_mounts {
      src     = "/opt/ansible/config"
      dest    = "/runner/config"
      options = "ro"
    }
  }
}
```

Use repeated `volume_mounts { ... }` blocks rather than
`volume_mounts = [{ ... }]` assignment. User mounts are preserved, and an
existing collections mount source or destination prevents a duplicate automatic
mount. With `ansible-navigator`, Navigator, the runtime, and mount sources are on
the Packer host. With `ansible-navigator-local`, they are on the target; installing
Docker or Navigator on the Packer host does not supply them on the target.

Pull arguments are independent of pull policy. The SSH provisioner's YAML nests
them under `ansible-navigator.execution-environment.pull.arguments`, alongside
`pull.policy` when configured. An arguments-only configuration omits YAML policy;
a policy-only configuration omits arguments; when both are absent, the `pull`
object is omitted. The on-target provisioner uploads minimal YAML containing pull arguments
and references that file from its remote invocation. Pull policy, engine,
container options, and mounts continue to use the existing CLI flags.

Arguments are runtime-specific. For example, a Podman pull flag is not necessarily
a Docker pull flag. Use the selected engine's supported flags and existing
credential store for authenticated registries; keep credentials out of HCL
argument lists.

The SSH provisioner adds `--tty=false` for Docker stdout EEs only when no explicit
TTY option is present. Its bridge host mapping follows the topology rules below;
these additions preserve user options. It skips the automatic collections mount
only when the collections path was defaulted by the plugin, that directory is
absent, and no requirements file is configured. Explicit paths, existing caches,
and requirements-backed caches retain their mounts. This host-cache check does
not apply to the on-target provisioner's target paths.

The [complete example](../example/execution-environment-config.pkr.hcl) separates
host-based SSH and on-target builds. Its comments describe each build's runtime
prerequisites; template validation alone does not establish live on-target EE
qualification.

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

### Execution-environment proxy topology

The remote provisioner resolves the SSH adapter route as one topology before
it starts provisioning:

| Execution topology | Listener default | Inventory host default |
| --- | --- | --- |
| EE disabled | `127.0.0.1` | `127.0.0.1` |
| EE with `--network=host` or `--net=host` | `127.0.0.1` | `127.0.0.1` |
| Docker bridge | `0.0.0.0` | `gateway.docker.internal` |
| Podman bridge | `0.0.0.0` | `host.containers.internal` |

An omitted or `auto` container engine follows Ansible Navigator's preference:
Podman, then Docker. The selected engine is written to the generated Navigator
configuration so runtime selection cannot diverge from proxy routing. Docker
bridge mode also adds
`--add-host=gateway.docker.internal:host-gateway` exactly once. Podman and host
networking do not receive this Docker-specific option.

Explicit `ansible_proxy_bind_address`, `ansible_proxy_host`, and
`container_options` values are preserved. If only one proxy endpoint is set,
the other is inferred from the topology. Known-unreachable combinations, such
as bridged container loopback or a loopback-only host listener, fail during
`Prepare` with corrective guidance.

Bridge mode temporarily binds the adapter to non-loopback host interfaces.
Exposure is limited by the ephemeral listener lifetime and one-time SSH key;
host firewall policy remains applicable. Remote container daemons cannot use a
gateway belonging to another host to reach the local Packer process, so they
require explicit reachable proxy addresses.

---

[← Back to docs index](README.md)
