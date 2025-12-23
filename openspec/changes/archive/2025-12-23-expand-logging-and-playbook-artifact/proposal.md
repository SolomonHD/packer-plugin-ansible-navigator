# Change: Expand `navigator_config` logging and playbook-artifact configuration

## Why

Users can supply `navigator_config { ... }` to generate an `ansible-navigator.yml`, but the current typed schema only models a subset of the ansible-navigator configuration surface for `logging` and `playbook-artifact`. This leads to “Unsupported argument” errors when users try to configure additional documented options.

## What Changes

- Expand the typed `LoggingConfig` and `PlaybookArtifact` models to match the documented ansible-navigator configuration schema (targeting the project’s “ansible-navigator v3+” baseline).
- Ensure YAML generation includes all supported keys with correct naming:
  - HCL/Go: underscore names (e.g., `save_as`)
  - YAML: hyphenated names (e.g., `save-as`)
- Preserve existing behavior for already-supported fields.

## Non-Goals

- No changes to provisioner registration or provisioner names.
- No changes to execution-environment or ansible-config modeling (handled by other changes in this prompt series).
- No attempt to create or manage playbook artifact files at runtime (ansible-navigator owns artifact creation).

## Impact

- **Affected specs (deltas in this proposal):**
  - `openspec/specs/remote-provisioner-capabilities/spec.md`
  - `openspec/specs/local-provisioner-capabilities/spec.md`
- **Expected implementation areas (not in this proposal workflow):**
  - `provisioner/ansible-navigator/*` and `provisioner/ansible-navigator-local/*` typed config structs and YAML conversion
  - Generated HCL2 spec files (`*.hcl2spec.go`) after regeneration
  - Unit tests validating HCL decoding and YAML structure

## Notes

- This proposal intentionally treats the upstream ansible-navigator documentation as the authoritative list of supported keys, and requires parity with that documentation for the `logging` and `playbook-artifact` sections.
