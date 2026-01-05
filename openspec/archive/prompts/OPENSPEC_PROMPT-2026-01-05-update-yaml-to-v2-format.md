# OpenSpec Change Prompt

## Context

The packer-plugin-ansible-navigator generates ansible-navigator YAML configuration files during Packer builds. Currently, these generated files are in an outdated format that triggers version migration prompts in ansible-navigator 25.x, and more critically, the pull-policy setting is not being respected, causing Docker to attempt pulling images even when `pull_policy = "never"` is set.

Investigation revealed:
- Plugin generates temporary `/tmp/packer-ansible-navigator-*.yml` config files
- ansible-navigator 25.12.0 detects these as "Version 1" format
- A migration is required to convert to "Version 2" format
- The migration updates the `pull-policy` handling
- Because files are temporary and deleted quickly, migration doesn't occur properly
- This causes Docker to ignore the never-pull policy and attempt registry pulls

## Goal

Update the plugin's YAML generation code to produce ansible-navigator configuration files in the **latest Version 2 format** that:
1. Does not trigger migration prompts
2. Correctly implements pull-policy so local images are used
3. Is fully compatible with ansible-navigator 25.x

## Scope

**In scope:**
- Research ansible-navigator Version 2 configuration format from official documentation and GitHub
- Identify all differences between V1 and V2 format
- Update Go code that generates ansible-navigator.yml files
- Ensure pull-policy is correctly formatted for V2
- Test with ansible-navigator 25.12.0 to confirm no migration needed
- Verify local Docker images are used without pull attempts

**Out of scope:**
- Switching to CLI flags (prefer config file approach)
- Changes to HCL configuration syntax (user-facing config stays same)
- Support for ansible-navigator versions older than 24.x
- Modifications to other provisioner modes (focus on ansible-navigator YAML generation only)

## Desired Behavior

When the plugin generates an ansible-navigator config file:
- ansible-navigator should recognize it as Version 2 format immediately
- No migration prompts should appear
- Setting `pull_policy = "never"` in HCL should result in ansible-navigator using local Docker images only
- All existing configuration options should continue working

## Constraints & Assumptions

- Assumption: ansible-navigator Version 2 format is documented in official ansible-navigator documentation or GitHub repository
- Assumption: The migration process hints at what changed (the output showed "Migration of 'pull-policy'..Updated")
- Constraint: Must maintain backward compatibility with existing Packer HCL configurations
- Constraint: Generated YAML must be directly usable by ansible-navigator 25.x without modification

## Acceptance Criteria

- [ ] Plugin generates YAML files that ansible-navigator 25.12.0+ recognizes as Version 2 format
- [ ] No "version migration required" prompts appear when using generated configs
- [ ] Setting `pull_policy = "never"` results in ansible-navigator using local images without attempting pulls
- [ ] Test Packer build completes successfully without "Unable to find image locally" errors when local image exists
- [ ] Generated YAML follows ansible-navigator Version 2 schema/structure from official documentation

## Research Required

The AI implementing this change should:
1. Search ansible-navigator GitHub repository for Version 2 configuration format documentation
2. Look for ansible-navigator settings file schema and migration changelog
3. Find examples of Version 2 format ansible-navigator.yml files
4. Identify the specific differences in pull-policy handling between V1 and V2
5. Check for any version markers or schema indicators needed in V2 format
