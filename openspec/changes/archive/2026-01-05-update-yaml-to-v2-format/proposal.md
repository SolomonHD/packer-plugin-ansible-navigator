# Change: Update ansible-navigator YAML Generation to Version 2 Format

## Why

The plugin currently generates ansible-navigator configuration files that trigger version migration prompts in ansible-navigator 25.x. This causes operational issues:

1. **Version mismatch warnings**: Generated files are detected as "Version 1" format, triggering migration prompts
2. **Pull-policy not respected**: The `pull_policy = "never"` setting doesn't prevent Docker from attempting registry pulls
3. **Temporary file complications**: Because config files are temp files deleted quickly, migrations don't persist properly

The root cause is missing schema version markers in the generated YAML. While the plugin already uses the correct nested structure for `pull.policy` (Version 2 format), ansible-navigator 25.x requires an explicit version marker to recognize the format immediately without triggering migration prompts.

## What Changes

- Research ansible-navigator Version 2 schema requirements (version markers, structural expectations)
- Update YAML generation in [`navigator_config.go`](../../provisioner/ansible-navigator/navigator_config.go:139) to include Version 2 schema markers
- Verify that generated files are immediately recognized as Version 2 format by ansible-navigator 25.12.0+
- Confirm `pull_policy = "never"` behavior works correctly (Docker uses local images without pull attempts)
- Update tests to validate Version 2 format generation

**Breaking Changes**: None - this is an internal YAML generation update. User-facing HCL configuration remains unchanged.

## Impact

**Affected specs:**
- `remote-provisioner-capabilities` - YAML config generation for ansible-navigator provisioner
- `local-provisioner-capabilities` - YAML config generation for ansible-navigator-local provisioner

**Affected code:**
- [`provisioner/ansible-navigator/navigator_config.go`](../../provisioner/ansible-navigator/navigator_config.go:139)
- [`provisioner/ansible-navigator-local/navigator_config.go`](../../provisioner/ansible-navigator-local/navigator_config.go:1)
- Related test files for YAML generation validation

**User experience improvement:**
- No more version migration warnings during Packer builds
- Pull-policy settings work as expected
- Transparent update - no HCL configuration changes required
