# Change: Regenerate HCL2 Spec Files

## Why

The `ansible_navigator_path` configuration field was added to the Config struct but the HCL2 spec files (`.hcl2spec.go`) were not regenerated. When Packer parses HCL templates, it uses the generated `FlatConfig` struct and `HCL2Spec()` function to determine which attributes are valid. Since `ansible_navigator_path` is missing from the generated files, Packer silently ignores this attribute.

This is a **build process fix**, not a feature change - the Go code is correct, but the generated files are stale.

## What Changes

- Regenerate `provisioner/ansible-navigator/provisioner.hcl2spec.go`
- Regenerate `provisioner/ansible-navigator-local/provisioner.hcl2spec.go` (if applicable)
- Verify all Config struct fields appear in generated FlatConfig structs

## Impact

- **Affected specs**: None (this is a build tooling fix, not a spec change)
- **Affected code**: `*.hcl2spec.go` files (auto-generated)
- **User impact**: `ansible_navigator_path` attribute will be recognized in Packer HCL templates

## Notes

This change uses `--skip-specs` when archiving since no capability specs are modified. The underlying capability already exists in code; we are simply synchronizing the generated HCL2 parsing layer.
