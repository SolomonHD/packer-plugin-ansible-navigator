# Proposal: Fix ansible-navigator Hang Issues

## Problem Statement

The packer-plugin-ansible-navigator causes ansible-navigator to hang during execution due to two related bugs:

1. **Missing `--mode` CLI flag** - The plugin generates a correct `ansible-navigator.yml` config file and sets the `ANSIBLE_NAVIGATOR_CONFIG` environment variable, but doesn't pass the `--mode` CLI flag when `NavigatorConfig.Mode` is set. This causes ansible-navigator to default to interactive mode and hang waiting for terminal input.

2. **Incorrect YAML root structure** - The generated YAML config places settings at the root level instead of nested under the required `ansible-navigator:` key, causing validation errors and unexpected behavior in ansible-navigator v25.x.

## Goals

1. Add `--mode` CLI flag to ansible-navigator command when `NavigatorConfig.Mode` is configured
2. Wrap all YAML settings under the `ansible-navigator:` root key to conform to ansible-navigator v25.x schema
3. Document workaround for asdf shim recursion issues (recommend using direct binary paths)
4. Maintain backward compatibility with existing HCL configurations

## Non-Goals

- Changing the user-visible HCL configuration format
- Adding new configuration options beyond what's needed for the fixes
- Modifying galaxy/collection installation logic
- Changes to the SSH proxy adapter implementation

## Scope

### In Scope

- **CLI command generation**: Add conditional `--mode` flag when `navigator_config.mode` is set (for both provisioners)
- **YAML structure**: Modify `convertToYAMLStructure()` to wrap settings under `ansible-navigator:` root key
- **Documentation**: Add troubleshooting section about asdf shim recursion and recommend direct binary paths
- **Testing**: Ensure changes work with both local and remote provisioners
- **Validation**: All four verification commands must pass: `make generate && go build ./... && go test ./... && make plugin-check`

### Out of Scope

- User-facing HCL block structure changes
- New configuration fields or capabilities
- Galaxy/collection dependency management changes
- SSH proxy implementation changes
- Changes to provisioner registration or plugin initialization

## Success Criteria

- [ ] `--mode` flag is added to ansible-navigator command when `navigator_config.mode` is set
- [ ] Generated YAML has `ansible-navigator:` root key wrapping all settings
- [ ] Validation errors about "unexpected properties" are resolved
- [ ] ansible-navigator executes without hanging in test scenarios
- [ ] Existing HCL configs continue to work without modification
- [ ] Documentation warns about asdf shim recursion and provides workaround
- [ ] All four verification commands pass: `make generate && go build ./... && go test ./... && make plugin-check`

## Implementation Notes

### Files to Modify

1. **`provisioner/ansible-navigator/provisioner.go`** (approx. lines 1155-1163)
   - Add conditional logic to prepend `--mode <value>` flag when `NavigatorConfig.Mode` is set
   - Insert after `"run"` argument, before other ansible-navigator arguments

2. **`provisioner/ansible-navigator-local/provisioner.go`** (similar location)
   - Same --mode flag logic for local provisioner consistency

3. **`provisioner/ansible-navigator/navigator_config.go`** (lines 69-178)
   - Modify `convertToYAMLStructure()` to wrap all settings in `ansible-navigator:` root key
   - Ensure nested structure is preserved (execution-environment, ansible, etc.)

4. **`provisioner/ansible-navigator-local/navigator_config.go`** (similar logic)
   - Same YAML root key wrapping for local provisioner

5. **`README.md` or `docs/troubleshooting.md`**
   - Add section about asdf shim recursion issue
   - Recommend using absolute paths like `/home/user/.asdf/installs/python/x.y.z/bin/ansible-navigator` or `ansible_navigator_path` configuration

### Root Cause Analysis

- **Hang Issue #1:** ansible-navigator defaults to interactive mode when `--mode` is not specified on CLI, regardless of config file settings. The config file alone is insufficient.
- **Hang Issue #2:** asdf shims can create recursive execution loops in some configurations (not a plugin bug, but worth documenting)
- **Validation Error:** ansible-navigator v25.x schema validation requires `ansible-navigator:` root key in config YAML. Flat structure at root causes "Additional properties" validation errors.

## Related Changes

- Related to existing OpenSpec change: `fix-navigator-yaml-pull-policy-structure` which addresses YAML schema compliance for pull policy nested structure
- Both changes address ansible-navigator v25.x schema compliance

## Links

- Acceptance criteria partially complete: `--mode` flag logic may already be implemented (marked as [x] in prompt)
- Schema reference: ansible-navigator v25.x requires top-level `ansible-navigator:` key
