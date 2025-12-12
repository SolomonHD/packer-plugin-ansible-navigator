# Change: Remove Legacy Navigator Options (Breaking)

## Why

After giving users time to migrate (Phase 1 added `navigator_config`, Phase 2 added deprecation warnings), this phase removes the deprecated configuration options entirely. This simplifies the codebase and eliminates redundant configuration paths.

**Removed options:**

- `ansible_cfg` - Replaced by `navigator_config.ansible.config`
- `ansible_env_vars` - Replaced by `navigator_config.execution-environment.environment-variables`
- `ansible_ssh_extra_args` - Use navigator_config or play-level options
- `extra_arguments` - Use navigator_config
- `execution_environment` - Replaced by `navigator_config.execution-environment`
- `navigator_mode` - Replaced by `navigator_config.mode`
- `roles_path` - Use navigator_config or requirements_file
- `collections_path` - Use navigator_config or requirements_file
- `galaxy_command` - Unnecessary with requirements_file

This is a **BREAKING CHANGE**. Configurations using removed options will fail validation.

## What Changes

### Removed

All deprecated configuration fields are completely removed:

- **Config structs**: Fields removed from both provisioner Config structs
- **Processing code**: All code handling deprecated options removed
- **CLI flag generation**: No longer generate `--mode`, `--ee`, `--eei` flags
- **Validation**: No longer validate removed fields
- **Documentation**: References to removed options deleted

### Retained

The following options remain unchanged:

- **Navigator Config**: `navigator_config` (primary configuration method)
- **Plays**: `play { }` blocks (repeatable, ordered)
- **Dependencies**: `requirements_file`, `force_update`, `offline_mode`, `galaxy_force_install`, `collections_cache_dir`, `roles_cache_dir`
- **Command**: `command`, `ansible_navigator_path`
- **Behavior**: `keep_going`, `structured_logging`, `log_output_path`, `verbose_task_output`
- **SSH (remote only)**: `use_proxy`, `local_port`, `inventory_*`, `groups`, `empty_groups`, `host_alias`, `limit`, `user`, SSH key options, `ansible_proxy_bind_address`, `ansible_proxy_host`
- **Version**: `skip_version_check`, `version_check_timeout`

## Impact

### Affected Specs

- `local-provisioner-capabilities` - Remove deprecated requirements
- `remote-provisioner-capabilities` - Remove deprecated requirements

### Affected Code

- `provisioner/ansible-navigator/provisioner.go` - Remove deprecated fields and processing code
- `provisioner/ansible-navigator-local/provisioner.go` - Remove deprecated fields and processing code
- `provisioner/*/provisioner.hcl2spec.go` - Regenerate after removing fields
- `provisioner/*/galaxy.go` - Update to not use removed fields
- All test files - Update to use navigator_config
- All documentation files

### Breaking Changes

**YES - Major breaking change**:

- Configurations using any removed option will fail validation with clear error messages
- No backward compatibility or migration shims
- Users must have migrated to `navigator_config` before upgrading
- Error messages will guide users to MIGRATION.md for help

### Migration Requirements

**Users MUST migrate before upgrading to this version:**

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

See MIGRATION.md for complete migration guide.

## Next Phase

Phase 4 will optimize and polish the `navigator_config` implementation now that legacy options are removed.
