# OpenSpec Prompt: Update YAML Generation to Version 2 Format

## Context

The [`packer-plugin-ansible-navigator`](../../README.md:1) generates ansible-navigator YAML configuration files during Packer builds. Currently, these generated files are in an outdated Version 1 format that triggers version migration prompts in ansible-navigator 25.x.

**Critical Issue:** The `pull-policy` setting is not being respected, causing Docker to attempt pulling images even when `pull_policy = "never"` is set in the HCL configuration.

### Investigation Findings

- Plugin generates temporary `/tmp/packer-ansible-navigator-*.yml` config files
- ansible-navigator 25.12.0 detects these as "Version 1" format
- A migration is required to convert to "Version 2" format
- The migration updates the `pull-policy` handling
- Because files are temporary and deleted quickly, migration doesn't occur properly
- This causes Docker to ignore the never-pull policy and attempt registry pulls

### Technical Root Cause

The YAML generation code in [`provisioner/ansible-navigator/provisioner.go`](../../provisioner/ansible-navigator/provisioner.go:1) produces Version 1 format files with outdated structure and key naming conventions. ansible-navigator 25.x expects Version 2 format with updated schema markers and key structures.

## Goal

Update the plugin's YAML generation code to produce ansible-navigator configuration files in the **Version 2 format** that:

1. Does not trigger migration prompts in ansible-navigator 25.x
2. Correctly implements pull-policy so local images are used when specified
3. Is fully compatible with ansible-navigator 25.x and later versions

## Scope

### In Scope

- Research ansible-navigator Version 2 configuration format from official documentation and GitHub
- Identify all structural differences between V1 and V2 format:
  - Schema version markers
  - Key naming conventions (hyphens vs underscores, etc.)
  - Nested structure changes
  - Pull-policy implementation in V2
- Update Go code in [`navigator_config.go`](../../provisioner/ansible-navigator/navigator_config.go:1) that generates `ansible-navigator.yml` files
- Ensure `pull-policy` is correctly formatted for V2 schema
- Update YAML generation in [`provisioner.go`](../../provisioner/ansible-navigator/provisioner.go:1) to match V2 structure
- Test with ansible-navigator 25.12.0 to confirm:
  - No migration prompts appear
  - Local Docker images are used without pull attempts when `pull_policy = "never"`
  - All existing configuration options continue working

### Out of Scope

- Switching from config files to CLI flags (prefer config file approach)
- Changes to HCL configuration syntax (user-facing config interface stays the same)
- Support for ansible-navigator versions older than 24.x
- Modifications to other provisioner modes (focus on YAML generation only)
- Expanding navigator_config beyond current capabilities (that's addressed in prompts 01-04)

## Desired Behavior

### Before (Current V1 Format)

```yaml
# Generated file triggers migration warning
ansible-navigator:
  execution-environment:
    pull-policy: never  # V1 format - may not be respected
```

### After (Target V2 Format)

```yaml
# V2 format - recognized immediately, no migration needed
# (exact structure TBD from ansible-navigator v2 schema research)
ansible-navigator:
  execution-environment:
    pull:
      policy: never  # V2 format - correctly respected
```

### User Experience

When the plugin generates an ansible-navigator config file:
- ✅ ansible-navigator recognizes it as Version 2 format immediately
- ✅ No migration prompts appear
- ✅ Setting `pull_policy = "never"` in HCL results in ansible-navigator using local Docker images only
- ✅ All existing configuration options continue working without HCL changes

## Constraints & Assumptions

### Assumptions

- ansible-navigator Version 2 format is documented in official ansible-navigator documentation or GitHub repository
- The migration process output hints at what changed: "Migration of 'pull-policy'..Updated" suggests structural changes
- Version 2 format includes a schema version marker or uses structural conventions that ansible-navigator detects
- The plugin currently generates minimal configuration - Version 2 conversion should be straightforward

### Constraints

- **Backward compatibility:** Must maintain compatibility with existing Packer HCL configurations
- **No user impact:** Generated YAML structure changes should be transparent to users
- **Version support:** Must work correctly with ansible-navigator 25.x without modification
- **Packer plugin rules:** Must follow patterns in [`packer-modern-plugin-install.md`](../../../../../.local/share/chezmoi/dot_kilocode/rules/packer-modern-plugin-install.md:1) and [`packer-versioning-testing.md`](../../../../../.local/share/chezmoi/dot_kilocode/rules/packer-versioning-testing.md:1)

## Acceptance Criteria

- [ ] Research complete: Version 2 schema documented (source: ansible-navigator GitHub or docs)
- [ ] YAML generation updated: Plugin generates Version 2 format files
- [ ] No migration warnings: ansible-navigator 25.12.0+ recognizes files as Version 2 format
- [ ] Pull-policy works: Setting `pull_policy = "never"` results in ansible-navigator using local images without pull attempts
- [ ] Test success: Packer build completes successfully without "Unable to find image locally" errors when local image exists
- [ ] Schema compliance: Generated YAML follows ansible-navigator Version 2 schema structure
- [ ] Backward compatible: Existing HCL configurations continue to work without modification
- [ ] Verification passed: All four verification commands pass:
  - `make generate`
  - `go build ./...`
  - `go test ./...`
  - `make plugin-check`

## Files Expected to Change

- [`provisioner/ansible-navigator/navigator_config.go`](../../provisioner/ansible-navigator/navigator_config.go:1) - YAML generation logic
- [`provisioner/ansible-navigator/provisioner.go`](../../provisioner/ansible-navigator/provisioner.go:1) - Config file writing
- Test files for YAML generation validation

## Research Required

The implementer should:

1. **Locate Version 2 documentation:**
   - Search ansible-navigator GitHub repository for Version 2 configuration format documentation
   - Check ansible-navigator changelog/migration guides for V1→V2 changes
   - Find examples of Version 2 format `ansible-navigator.yml` files

2. **Identify structural differences:**
   - Schema version markers (if any)
   - Key naming changes: `pull-policy` → `pull.policy` or other patterns
   - Nested structure reorganization
   - Any new required fields

3. **Understand pull-policy specifically:**
   - How V1 format handled `pull-policy`
   - How V2 format handles pull policy (nested structure, different keys)
   - Why V1 format doesn't work correctly with ansible-navigator 25.x

4. **Validate approach:**
   - Create a manual Version 2 format file
   - Test with ansible-navigator 25.12.0 to confirm no migration warnings
   - Verify pull-policy behavior with `never` setting

## Success Metrics

- Zero migration prompts when using plugin-generated configs with ansible-navigator 25.12.0+
- Docker pull operations respect `pull_policy = "never"` setting
- Existing plugin functionality unchanged from user perspective
- All tests pass with Version 2 YAML generation
