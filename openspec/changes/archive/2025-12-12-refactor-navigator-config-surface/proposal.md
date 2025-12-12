# Change: Replace ~40 Config Options with Single navigator_config

## Why

The `ansible-navigator` provisioners have accumulated ~40 configuration options, many of which are redundant or don't work reliably when using Execution Environments (EE containers). The core issue: only `ansible-navigator.yml` reliably controls EE container behavior. The current approach using `ansible_cfg`, `ansible_env_vars`, `execution_environment`, `navigator_mode`, and dozens of other options creates a complex, unreliable configuration surface.

**Problems with current configuration:**

- `ansible.cfg` is a host-side file that containers often ignore
- Environment variables set on the host don't propagate into containers
- Multiple ways to configure the same thing (e.g., `execution_environment` string vs. `navigator_config.execution-environment` object)
- Users must understand which settings work in EE vs. non-EE modes
- Configuration complexity makes troubleshooting difficult

Making `ansible-navigator.yml` the primary configuration mechanism aligns with ansible-navigator v3+ best practices and eliminates configuration ambiguity.

## What Changes

**BREAKING CHANGE**: This is a major simplification of the configuration surface.

### Added

- **`navigator_config`**: New HCL map option that maps directly to `ansible-navigator.yml` schema
  - Supports full ansible-navigator.yml structure including:
    - `ansible` section (config overrides, playbook settings)
    - `execution-environment` object (enabled, image, pull-policy, environment-variables)
    - `mode` (stdout, json, yaml, interactive)
    - All other ansible-navigator.yml options
  - Plugin generates temporary `ansible-navigator.yml` file
  - Points navigator at it via `ANSIBLE_NAVIGATOR_CONFIG` environment variable
  - When `execution-environment.enabled = true`, auto-sets default temp dir env vars to prevent `/.ansible/tmp` permission failures
  - Plugin cleans up the config file after execution

### Removed

The following options are **removed entirely** (no migration path):

- `ansible_cfg` (replaced by `navigator_config.ansible.config`)
- `ansible_env_vars` (replaced by `navigator_config.execution-environment.environment-variables`)
- `ansible_ssh_extra_args` (use navigator_config or play options)
- `extra_arguments` (use navigator_config)
- `execution_environment` (replaced by `navigator_config.execution-environment`)
- `navigator_mode` (replaced by `navigator_config.mode`)
- `roles_path` (use navigator_config or requirements_file)
- `collections_path` (use navigator_config or requirements_file)
- `galaxy_command` (unnecessary with requirements_file)

### Retained

The following options remain unchanged:

- **Navigator Config**: `navigator_config` (new)
- **Plays**: `play { }` blocks (repeatable, ordered)
- **Dependencies**: `requirements_file`, `force_update`, `offline_mode`, `galaxy_force_install`, `collections_cache_dir`, `roles_cache_dir`
- **Command**: `command`, `ansible_navigator_path`
- **Behavior**: `keep_going`, `structured_logging`, `log_output_path`, `verbose_task_output`
- **SSH (remote only)**: `use_proxy`, `local_port`, `inventory_*`, `groups`, `empty_groups`, `host_alias`, `limit`, `user`, SSH key options, `ansible_proxy_bind_address`, `ansible_proxy_host`
- **Version**: `skip_version_check`, `version_check_timeout`

## Impact

### Affected Specs

- `local-provisioner-capabilities` - Major configuration removal and addition
- `remote-provisioner-capabilities` - Major configuration removal and addition

### Affected Code

- `provisioner/ansible-navigator/provisioner.go` - Remove options, add navigator_config processing
- `provisioner/ansible-navigator-local/provisioner.go` - Remove options, add navigator_config processing
- `provisioner/*/provisioner.hcl2spec.go` - Regenerate after removing fields
- All documentation files (README.md, docs/CONFIGURATION.md, docs/EXAMPLES.md, docs/TROUBLESHOOTING.md)

### Breaking Changes

**Yes - Major breaking change**:

- ~40 configuration options removed
- No backward compatibility
- Users must migrate to `navigator_config` approach
- All existing configurations using removed options will fail validation

### Migration

Users updating from previous versions need to:

1. Replace `execution_environment = "image"` with:

   ```hcl
   navigator_config = {
     execution-environment = {
       enabled = true
       image = "image"
     }
   }
   ```

2. Replace `navigator_mode = "stdout"` with:

   ```hcl
   navigator_config = {
     mode = "stdout"
   }
   ```

3. Replace `ansible_cfg = {...}` with:

   ```hcl
   navigator_config = {
     ansible = {
       config = {
         defaults = { ... }
       }
     }
   }
   ```

4. Replace `ansible_env_vars = {...}` with:

   ```hcl
   navigator_config = {
     execution-environment = {
       environment-variables = { ... }
     }
   }
   ```
