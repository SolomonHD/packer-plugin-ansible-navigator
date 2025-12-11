# Change: Deprecate Legacy Navigator Options (Warning Phase)

## Why

Phase 1 added the `navigator_config` field, providing users a better way to configure ansible-navigator. This phase marks legacy configuration options as deprecated to give users a clear migration window before removal.

**Deprecated options:**

- `ansible_cfg` - Use `navigator_config.ansible.config` instead
- `ansible_env_vars` - Use `navigator_config.execution-environment.environment-variables` instead
- `ansible_ssh_extra_args` - Use navigator_config or play-level options instead
- `extra_arguments` - Use navigator_config instead
- `execution_environment` - Use `navigator_config.execution-environment` instead
- `navigator_mode` - Use `navigator_config.mode` instead
- `roles_path` - Use navigator_config or requirements_file instead
- `collections_path` - Use navigator_config or requirements_file instead
- `galaxy_command` - Unnecessary with requirements_file

This is a **non-breaking** change. All deprecated options continue to work, but now emit runtime warnings.

## What Changes

### Modified

- **Documentation**: All references to deprecated options updated with deprecation notices
- **Field comments**: Deprecation notices added to Config struct field comments
- **Runtime warnings**: When deprecated fields are used, warnings logged to Packer output
- **Migration guide**: Added to documentation showing how to migrate each deprecated option

### Retained

All existing functionality remains unchanged:

- Deprecated options still work exactly as before
- `navigator_config` continues to take precedence when both present
- No behavioral changes to how the provisioners execute

## Impact

### Affected Specs

- No spec changes - This is a documentation and warning-only change

### Affected Code

- `provisioner/ansible-navigator/provisioner.go` - Add deprecation warnings to Config struct comments
- `provisioner/ansible-navigator-local/provisioner.go` - Add deprecation warnings to Config struct comments
- `provisioner/ansible-navigator/provisioner.go` - Add runtime warning logging when deprecated fields used
- `provisioner/ansible-navigator-local/provisioner.go` - Add runtime warning logging when deprecated fields used
- All documentation files (README.md, CONFIGURATION.md, EXAMPLES.md)

### Breaking Changes

**No breaking changes** - This is a warning-only phase. All functionality remains intact.

### Warning Messages

Users will see warnings like:

```
Warning: 'execution_environment' is deprecated and will be removed in a future version. 
Use 'navigator_config.execution-environment' instead. 
See MIGRATION.md for details.
```

## Migration Timeline

This deprecation provides users with a grace period:

1. **Current release**: Warnings added, options still work
2. **1-2 releases later**: Options removed (Phase 3)

Users should update their configurations during this window.

## Next Phase

After sufficient warning period (1-2 releases), Phase 3 will remove the deprecated options entirely.
