# OpenSpec change prompt

## Context

The `Play` struct in the Ansible Navigator provisioner currently lacks support for `become_user` and `skip_tags`, which are common Ansible parameters needed for more granular control over playbook execution.

## Goal

Add support for `become_user` and `skip_tags` to the `Play` configuration block so users can specify them in their Packer templates.

## Scope

- In scope:
  - Update `Play` struct in `provisioner/ansible-navigator/provisioner.go`.
  - Update `HCL2Spec` generation (run `go generate` or manually update `provisioner.hcl2spec.go`).
  - Update execution logic to pass `--become-user` and `--skip-tags` to the underlying Ansible command.
- Out of scope:
  - Changes to other structs or global config.

## Desired behavior

- Users can define `become_user = "root"` in a `play` block.
- Users can define `skip_tags = ["tag1", "tag2"]` in a `play` block.
- These values are correctly passed to the ansible execution command as command-line arguments.

## Constraints & assumptions

- Assumption: The project uses `go generate` for HCL2 spec generation.
- Constraint: Must maintain backward compatibility with existing configuration.

## Acceptance criteria

- [ ] `Play` struct includes `BecomeUser` (string) and `SkipTags` ([]string).
- [ ] `HCL2Spec` reflects the new fields.
- [ ] Command construction logic includes `--become-user` and `--skip-tags` when populated.
