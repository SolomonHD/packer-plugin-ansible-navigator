# Change: Align play block schema with singular HCL2 naming

## Why

The repository currently documents singular `play` blocks for multi-play configuration in user docs and OpenSpec specs, but the compiled HCL2 schemas still expose a plural `plays` block name. When a user follows the docs and writes `play { ... }` blocks, Packer reports errors such as:

> Blocks of type "play" are not expected here. Did you mean "plays"?

This mismatch between docs/specs and the actual plugin schema creates confusing validation errors and undermines the contract captured in the `local-provisioner-capabilities` and `remote-provisioner-capabilities` specs.

We also want to make `play` the only supported block name going forward and treat any use of legacy `plays` blocks as a hard validation error with a clear migration message.

## What Changes

### HCL2 schema and config wiring

- Update the configuration structs for both `ansible-navigator` (SSH-based) and `ansible-navigator-local` (on-target) provisioners so the multi-play field is wired to a singular `play` block name while remaining a slice in Go.
- Regenerate HCL2 specs so the generated `FlatConfig` and `HCL2Spec` for both provisioners:
  - Use a `BlockListSpec` keyed as `play` (singular) for the multi-play configuration.
  - Do not expose a `plays` key or `TypeName` in the schema for multi-play configuration.

### Validation and error semantics

- Ensure validation logic for both provisioners enforces mutual exclusivity between `playbook_file` and `play` blocks, with error messages that reference `play` (not `plays`).
- Treat any use of legacy `plays` blocks as a hard error:
  - If configuration uses a `plays { ... }` block, the plugin SHALL surface a clear error such as:
    > Blocks of type "plays" are no longer supported. Use repeated `play { ... }` blocks instead.
  - The error MUST be visible both via Packer CLI output and as a returned Go `error`.
- Continue to reject old array-style syntax (`plays = [...]`) with an error that points users at `play { ... }` block syntax.

### Testing

- Add or update unit tests for both provisioners to cover:
  - Successful validation and execution of multiple `play { ... }` blocks.
  - Mutual exclusivity errors for `playbook_file` plus `play` blocks.
  - Hard-error behavior when `plays { ... }` blocks are used.
  - Rejection of `plays = [...]` array syntax with a migration hint.

### Documentation alignment

- Update documentation and examples (README and docs under `docs/`) so:
  - All multi-play configuration uses repeated `play { ... }` blocks.
  - No examples use `plays`, `plays { ... }`, or `plays = [...]`.
  - Any migration notes clearly state that `plays` blocks are no longer supported and must be rewritten as `play` blocks.

## Impact

- **Specs affected**
  - `local-provisioner-capabilities`
  - `remote-provisioner-capabilities`
  - `project-metadata` (for documentation accuracy around configuration syntax)

- **Code and files affected (non-exhaustive)**
  - `provisioner/ansible-navigator/provisioner.go`
  - `provisioner/ansible-navigator/provisioner.hcl2spec.go`
  - `provisioner/ansible-navigator-local/provisioner.go`
  - `provisioner/ansible-navigator-local/provisioner.hcl2spec.go`
  - Unit tests for both provisioners
  - `README.md` and configuration docs under `docs/`

- **Breaking changes**
  - Configurations that currently use `plays { ... }` blocks will begin failing validation and MUST be migrated to `play { ... }` blocks.
  - Configurations that already use singular `play { ... }` blocks (matching the docs and specs) will start validating and running correctly.

## Acceptance criteria

- The generated HCL2 specs for both provisioners expose a singular `play` block (no `plays` key or `TypeName`) for multi-play configuration.
- `Config` and `FlatConfig` types for both provisioners compile with tags that drive the singular `play` block name while preserving slice semantics.
- Validation and error messages for mutual exclusivity and missing configuration reference `playbook_file` and `play` blocks (singular) and do not suggest `plays` in new text.
- A Packer template using `provisioner "ansible-navigator"` and/or `provisioner "ansible-navigator-local"` with multiple `play { ... }` blocks, and no legacy `plays` syntax, passes `packer validate` and `packer build` with the updated plugin.
- Any use of `plays { ... }` or `plays = [...]` produces a clear, actionable migration error pointing to `play { ... }` blocks.
- User-facing documentation contains only `play { ... }` examples for multi-play configuration and clearly describes the singular block semantics.
