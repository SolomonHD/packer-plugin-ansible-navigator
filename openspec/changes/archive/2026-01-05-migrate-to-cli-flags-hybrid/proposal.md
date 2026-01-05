# Change: Migrate ansible-navigator Configuration to CLI Flags (Hybrid Approach)

## Why

The plugin currently generates temporary YAML configuration files (`ansible-navigator.yml`) and passes them via the `--settings` flag. This approach has encountered persistent issues:

1. **Version Format Issues**: ansible-navigator 25.x detects YAML as "Version 1" format and triggers migration prompts
2. **Configuration Ignored**: Settings like `pull_policy = "never"` are being ignored despite correct YAML structure
3. **Temporary File Management**: Complexity in creating, tracking, and cleaning up temp files across provisioners
4. **YAML Schema Dependency**: Vulnerable to ansible-navigator YAML schema changes

CLI flags are the preferred, stable interface for ansible-navigator configuration and are less prone to version-specific issues.

## What Changes

- **Refactor command construction** to use CLI flags as the primary configuration method
- **Add CLI flag builder** functions for mapping NavigatorConfig fields to ansible-navigator flags
- **Implement hybrid strategy**: generate minimal YAML only for advanced features lacking CLI flag equivalents (playbook-artifact, collection-doc-cache)
- **Maintain backward compatibility**: existing HCL configurations work without changes
- **Update both provisioners**: ansible-navigator (remote) and ansible-navigator-local

**BREAKING**: None. User-facing HCL configuration remains unchanged.

## Impact

### Affected Specs
- [`command-argument-construction`](../../specs/command-argument-construction/spec.md) - New CLI flag generation requirements
- [`remote-provisioner-capabilities`](../../specs/remote-provisioner-capabilities/spec.md) - Updated navigator configuration handling
- [`local-provisioner-capabilities`](../../specs/local-provisioner-capabilities/spec.md) - Updated navigator configuration handling

### Affected Code
- `provisioner/ansible-navigator/provisioner.go` - Command construction logic
- `provisioner/ansible-navigator/navigator_config.go` - Add CLI flag builder, minimal YAML generator
- `provisioner/ansible-navigator-local/provisioner.go` - Command construction logic  
- `provisioner/ansible-navigator-local/navigator_config.go` - Add CLI flag builder, minimal YAML generator
- Tests for both provisioners

### Benefits
- Eliminates 90%+ of YAML file generation
- Resolves Version 2 format issues by avoiding YAML in common cases
- Improves debugging with explicit CLI commands visible in logs
- Simplifies maintenance by reducing YAML schema dependency
- Fixes pull-policy ignored bug
