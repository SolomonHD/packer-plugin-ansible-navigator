# Change: Add navigator_config Field (Non-Breaking)

## Why

The ansible-navigator provisioners currently support ~40 configuration options, but only `ansible-navigator.yml` reliably controls behavior when using Execution Environment (EE) containers. This phase adds the new `navigator_config` field alongside existing options to provide users a migration path.

**Key benefits of navigator_config:**

- Aligns with ansible-navigator v3+ best practices
- Reliably controls EE container behavior
- Reduces configuration complexity
- Single source of truth for ansible-navigator settings

This is a **non-breaking** change. Both old and new configuration methods work simultaneously, allowing users to migrate at their own pace.

## What Changes

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
  - When both `navigator_config` and legacy options are present, `navigator_config` takes precedence

### Retained

All existing configuration options remain unchanged and functional:

- **Legacy Config**: `ansible_cfg`, `ansible_env_vars`, `ansible_ssh_extra_args`, `extra_arguments`
- **Legacy Navigator**: `execution_environment`, `navigator_mode`
- **Legacy Paths**: `roles_path`, `collections_path`, `galaxy_command`
- **Plays**: `play { }` blocks (repeatable, ordered)
- **Dependencies**: `requirements_file`, `force_update`, `offline_mode`, `galaxy_force_install`, `collections_cache_dir`, `roles_cache_dir`
- **Command**: `command`, `ansible_navigator_path`
- **Behavior**: `keep_going`, `structured_logging`, `log_output_path`, `verbose_task_output`
- **SSH (remote only)**: `use_proxy`, `local_port`, `inventory_*`, `groups`, `empty_groups`, `host_alias`, `limit`, `user`, SSH key options, `ansible_proxy_bind_address`, `ansible_proxy_host`
- **Version**: `skip_version_check`, `version_check_timeout`

## Impact

### Affected Specs

- `local-provisioner-capabilities` - Add navigator_config capability
- `remote-provisioner-capabilities` - Add navigator_config capability

### Affected Code

- `provisioner/ansible-navigator/provisioner.go` - Add NavigatorConfig field and processing logic
- `provisioner/ansible-navigator-local/provisioner.go` - Add NavigatorConfig field and processing logic
- New: `provisioner/ansible-navigator/navigator_config.go` - YAML generation functions
- New: `provisioner/ansible-navigator-local/navigator_config.go` - YAML generation functions
- `provisioner/*/provisioner.hcl2spec.go` - Regenerate after adding field
- Documentation updates (README.md, CONFIGURATION.md, EXAMPLES.md)

### Breaking Changes

**No breaking changes** - This is an additive change. All existing configurations continue to work.

### Precedence Rules

When both legacy options and `navigator_config` are present:

1. `navigator_config` takes precedence for ansible-navigator settings
2. Legacy CLI flags are still generated but overridden by `ANSIBLE_NAVIGATOR_CONFIG`
3. This allows gradual migration without forcing immediate changes

## Migration Path

Users can migrate incrementally:

**Step 1**: Add `navigator_config` alongside existing options:

```hcl
provisioner "ansible-navigator" {
  # Old options still work
  execution_environment = "quay.io/ansible/creator-ee:latest"
  navigator_mode = "stdout"
  
  # New option takes precedence
  navigator_config = {
    execution-environment = {
      enabled = true
      image = "quay.io/ansible/creator-ee:latest"
    }
    mode = "stdout"
  }
}
```

**Step 2**: Test and verify new configuration

**Step 3**: Remove legacy options once confident (in future phase)

## Next Phase

After users have time to migrate, Phase 2 will add deprecation warnings for legacy options.
