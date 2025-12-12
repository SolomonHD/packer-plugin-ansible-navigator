# Change: Fix ansible-navigator YAML pull-policy Structure

## Why

The YAML generation code in both provisioners outputs an incorrect structure for `pull-policy` that is incompatible with ansible-navigator v24+/v25+. The code currently outputs `pull-policy` as a flat field, but ansible-navigator requires a nested `pull.policy` structure. This causes validation errors like "Additional properties are not allowed ('pull-policy' was unexpected)" and can cause ansible-navigator to hang or fail silently in non-TTY environments like Packer builds.

## What Changes

- Fix the `convertToYAMLStructure()` function in both provisioners to output the correct nested `pull.policy` structure
- Verify all other execution-environment fields follow the correct ansible-navigator v25+ schema
- Ensure generated YAML passes ansible-navigator's built-in validation

## Impact

- Affected specs: `local-provisioner-capabilities`, `remote-provisioner-capabilities` (both have identical YAML generation logic)
- Affected code:
  - `provisioner/ansible-navigator/navigator_config.go` (line 83)
  - `provisioner/ansible-navigator-local/navigator_config.go` (line 83)
- User-facing HCL configuration syntax remains unchanged (users continue to use `pull_policy` field)
- Breaking: None - This is a bug fix that restores compatibility with ansible-navigator v25+
