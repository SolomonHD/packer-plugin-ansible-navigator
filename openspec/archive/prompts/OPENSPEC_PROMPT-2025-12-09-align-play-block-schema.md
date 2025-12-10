# OpenSpec change prompt

## Context

The packer-plugin-ansible-navigator repo documents singular play blocks, but the compiled HCL2 spec still exposes a plural plays block. The generated schema in [`provisioner.hcl2spec.go`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/provisioner.hcl2spec.go) uses plays, while user docs like [`docs/CONFIGURATION.md`](packer/plugins/packer-plugin-ansible-navigator/docs/CONFIGURATION.md) and [`docs/UNIFIED_PLAYS.md`](packer/plugins/packer-plugin-ansible-navigator/docs/UNIFIED_PLAYS.md) describe play blocks. This causes Packer errors such as: “Blocks of type "play" are not expected here. Did you mean "plays"?” when users follow the docs.

## Goal

Align the plugin’s actual HCL2 schema, validation, and runtime behavior with the documented singular play block syntax, and remove all remaining dependency on plays as a block name, while preserving backward compatibility only where explicitly desired.

## Scope

- In scope:
  - Update Go configuration and generated HCL2 spec so the repeatable play configuration uses a singular play block name.
  - Ensure validation, error messages, and execution paths honor play blocks (and, if kept, handle legacy plays syntax in a controlled way).
  - Regenerate any generated files needed to keep schema and docs in sync.
  - Update all user-facing documentation and examples under this repo that still mention plays blocks or syntax.
- Out of scope:
  - Any changes to non-ansible-navigator plugins.
  - New features unrelated to the plays → play rename.

## Desired behavior

- The primary multi-play configuration in HCL is a repeatable singular play block inside `provisioner "ansible-navigator"` and `provisioner "ansible-navigator-local"`.
- The generated HCL2 spec for the remote provisioner in [`provisioner.hcl2spec.go`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/provisioner.hcl2spec.go) uses a BlockListSpec keyed as play (singular), not plays.
- The Go configuration struct and mapstructure tags for the multi-play field match the new block name and continue to support multiple items (slice).
- Validation logic enforces mutual exclusivity between playbook_file (and any playbook_files variant) and play blocks, with clear error messages that reference play, not plays.
- Packer templates that follow the docs and use play { ... } blocks for multiple plays validate and build successfully with the plugin.
- Any decision about accepting legacy plays { } syntax (if supported at all) is explicit, documented, and covered by tests; otherwise, legacy plural naming is rejected with a clear migration error that points to play { }.

## Constraints and assumptions

- Assume the desired long-term public API is singular play blocks, as already described in CHANGELOG 2.0.0+ and existing docs.
- Assume Go code generation for the HCL2 spec continues to be done via the existing mapstructure-to-hcl2 tooling; changes should be applied to the source of truth (the main configuration struct and tags) and then regenerated.
- Assume both remote and local ansible-navigator provisioners should have consistent play semantics and naming.
- Assume existing behavior for how individual plays are executed (target resolution, JSON logging, keep_going, become, etc.) must be preserved; only naming and wiring should change.

## Acceptance criteria

- [ ] The FlatConfig definition and HCL2Spec in [`provisioner.hcl2spec.go`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/provisioner.hcl2spec.go) no longer contain a plays key or TypeName; instead they expose a singular play block name corresponding to the multi-play configuration.
- [ ] The main configuration struct and any related types in the provisioner package (for both remote and local variants) compile with updated mapstructure/hcl tags that drive the new play block name, and go generate / make generate has been run so all generated files are in sync.
- [ ] Validation and error messages for mutual exclusivity and required fields reference playbook_file and play blocks (singular) and do not suggest plays in new error text.
- [ ] All examples and configuration tables under docs/ and README for this plugin consistently show singular play blocks (including any remaining references in .web-docs, docs-partials, and AGENTS), and no example uses plays = [...] or plays { }.
- [ ] A Packer template using `provisioner "ansible-navigator"` with multiple play { ... } blocks and no legacy plays syntax passes packer validate and packer build with the updated plugin binary.
- [ ] If legacy plural plays block syntax is intentionally supported for migration, its behavior and deprecation path are documented, and there are tests that cover both accepted and rejected legacy usage.
