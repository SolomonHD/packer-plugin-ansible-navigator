# OpenSpec Change Prompt

## Context

The Packer `ansible-navigator` provisioner has accumulated ~40 configuration options, many of which are redundant or don't work reliably when using Execution Environments (EE containers). The `ansible_cfg` and `ansible_env_vars` options fail to propagate settings inside EE containers because:

- `ansible.cfg` is a host-side file that the container often ignores
- Environment variables set on the host don't propagate into the container
- Only `ansible-navigator.yml` reliably controls EE container behavior

This change makes `ansible-navigator.yml` the **primary and only** configuration mechanism for ansible-navigator settings.

## Goal

Replace redundant configuration options with a single `navigator_config` option that generates an `ansible-navigator.yml` file. Remove all legacy options that duplicate this functionality.

## Scope

**In scope:**

- Add `navigator_config` option that maps directly to `ansible-navigator.yml` structure
- Remove `ansible_cfg`, `ansible_env_vars`, `ansible_ssh_extra_args`, `extra_arguments`
- Remove `execution_environment`, `navigator_mode` (move to `navigator_config`)
- Remove `roles_path`, `collections_path`, `galaxy_command` (move to `navigator_config` or unnecessary)
- Update both provisioners (remote and local) with the new config surface
- Update all documentation with new examples
- Clean up generated config file after execution

**Out of scope:**

- Backward compatibility shims
- New play semantics or galaxy behavior changes
- SSH proxy/inventory options (keep as-is for remote provisioner)

## Desired Behavior

- `navigator_config` accepts a structured HCL map matching `ansible-navigator.yml` schema
- Plugin generates a temporary `ansible-navigator.yml` and points navigator at it via `ANSIBLE_NAVIGATOR_CONFIG` or `--config`
- When `execution-environment.enabled = true`, default temp dir env vars are set automatically to prevent `/.ansible/tmp` permission failures
- Plugin cleans up the temporary config file after execution

## Constraints & Assumptions

- Assume ansible-navigator supports `ANSIBLE_NAVIGATOR_CONFIG` environment variable
- Backward compatibility is NOT required
- Completely remove all references to removed options from code and docs
- All examples in docs must use the new `navigator_config` approach

## Acceptance Criteria

- [ ] New `navigator_config` option generates `ansible-navigator.yml` from HCL
- [ ] Removed entirely: `ansible_cfg`, `ansible_env_vars`, `ansible_ssh_extra_args`, `extra_arguments`, `execution_environment`, `navigator_mode`, `roles_path`, `collections_path`, `galaxy_command`
- [ ] EE runs work without `/.ansible/tmp` permission failures when using `navigator_config`
- [ ] All docs updated: CONFIGURATION.md, EXAMPLES.md, README.md, TROUBLESHOOTING.md
- [ ] No references to removed options in codebase
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] `make plugin-check` passes

## Example: New Config Surface

```hcl
provisioner "ansible-navigator" {
  navigator_config = {
    ansible = {
      config = {
        defaults = {
          remote_tmp = "/tmp/.ansible/tmp"
          local_tmp  = "/tmp/.ansible-local"
        }
      }
    }
    
    execution-environment = {
      enabled = true
      image = "quay.io/ansible/creator-ee:latest"
      pull-policy = "missing"
      environment-variables = {
        ANSIBLE_REMOTE_TMP = "/tmp/.ansible/tmp"
        ANSIBLE_LOCAL_TMP  = "/tmp/.ansible-local"
      }
    }
    
    mode = "stdout"
  }
  
  requirements_file = "./requirements.yml"
  
  play {
    name   = "Configure Server"
    target = "site.yml"
  }
}
```

## Retained Options

| Category | Options |
|----------|---------|
| **Navigator Config** | `navigator_config` |
| **Plays** | `play { }` blocks |
| **Dependencies** | `requirements_file`, `force_update`, `offline_mode`, `galaxy_force_install`, `collections_cache_dir`, `roles_cache_dir` |
| **Command** | `command`, `ansible_navigator_path` |
| **Behavior** | `keep_going`, `structured_logging`, `log_output_path`, `verbose_task_output` |
| **SSH (remote only)** | `use_proxy`, `local_port`, `inventory_*`, `groups`, `empty_groups`, `host_alias`, `limit`, `user`, SSH key options |
| **Version** | `skip_version_check`, `version_check_timeout` |
