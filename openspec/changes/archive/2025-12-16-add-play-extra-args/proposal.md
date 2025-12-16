# Change: Add per-play `extra_args` escape hatch

## Why

The provisioners support structured `play {}` blocks, but they lack an escape hatch for uncommon `ansible-navigator run` flags.
Without a generic per-play option, users must request new config fields for each flag, which increases configuration surface area and maintenance burden.

## What Changes

- Add `extra_args` (list(string)) to the `play {}` block schema for both provisioners.
- Define a deterministic argument ordering so `extra_args` behavior is predictable and consistent across provisioners.

## Impact

- **Affected specs (delta in this proposal):**
  - `openspec/specs/remote-provisioner-capabilities/spec.md`
  - `openspec/specs/local-provisioner-capabilities/spec.md`
- **Affected code (implementation phase; not in this workflow):**
  - `provisioner/ansible-navigator/provisioner.go`
  - `provisioner/ansible-navigator-local/provisioner.go`
  - Generated HCL2 spec files for both provisioners
- **Affected docs/examples (implementation phase):**
  - `docs/CONFIGURATION.md` and any relevant examples

## Notes

- `extra_args` values are treated as trusted input (no validation beyond basic type/empty checks).
- This proposal intentionally scopes `extra_args` to the `play {}` block only (no global provisioner-level `extra_args`).

