# Change: Update ansible-navigator YAML Generation to Version 2 Format

## Why

The plugin currently generates ansible-navigator configuration files that trigger version migration prompts in ansible-navigator 25.x, and critically, the `pull_policy = "never"` setting is not being respected, causing Docker to attempt pulling images even when local images should be used. Investigation revealed that ansible-navigator 25.12.0 detects the generated YAML files as "Version 1" format and requires migration to "Version 2" format. Because the files are temporary and deleted quickly, the migration doesn't occur properly, leading to ignored pull-policy settings.

## What Changes

- Research ansible-navigator Version 2 configuration format requirements from official documentation
- Identify the specific differences between Version 1 and Version 2 format (version markers, schema definitions, or structural changes)
- Update [`provisioner/ansible-navigator/navigator_config.go`](../../provisioner/ansible-navigator/navigator_config.go) YAML generation code to produce Version 2 format
- Update [`provisioner/ansible-navigator-local/navigator_config.go`](../../provisioner/ansible-navigator-local/navigator_config.go) YAML generation code to produce Version 2 format
- Ensure pull-policy is correctly formatted for Version 2 to prevent Docker registry pulls when `pull_policy = "never"` is set
- Verify no migration prompts appear with ansible-navigator 25.12.0+
- Add version marker or schema identifier if required by Version 2 format
- Test that local Docker images are used without pull attempts

## Impact

**Affected specs:**
- [`remote-provisioner-capabilities`](../../openspec/specs/remote-provisioner-capabilities/spec.md) - Navigator Config File Generation requirement
- [`local-provisioner-capabilities`](../../openspec/specs/local-provisioner-capabilities/spec.md) - Navigator Config File Generation requirement

**Affected code:**
- `provisioner/ansible-navigator/navigator_config.go` - YAML generation logic
- `provisioner/ansible-navigator-local/navigator_config.go` - YAML generation logic (likely shares code)

**User impact:**
- Eliminates version migration prompts
- Fixes pull-policy behavior so `pull_policy = "never"` actually prevents Docker pull attempts
- No HCL configuration changes required (backward compatible for users)
- Improves reliability when using local-only container images

**Breaking changes:** None - this is a fix to output format; user-facing HCL configuration remains unchanged.
