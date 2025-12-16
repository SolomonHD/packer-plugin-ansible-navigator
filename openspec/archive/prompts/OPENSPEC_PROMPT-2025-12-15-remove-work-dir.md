# OpenSpec Change Prompt

## Context
The plugin currently exposes `work_dir`, but it is either inconsistently applied (remote) or effectively ignored (local runs in a staging directory).
This creates user confusion and an unnecessary configuration knob.

## Goal
Remove `work_dir` from both provisioners.

## Scope

**In scope:**
- Remove `work_dir` from both Config structs and HCL schemas.
- Remove any code paths that use `work_dir`.
- Regenerate HCL2 specs.
- Update docs/examples.

**Out of scope:**
- Introducing a replacement directory option.

## Desired Behavior
- Users no longer see or can set `work_dir`.
- Execution behavior is unchanged except that `work_dir` is no longer honored.

## Constraints & Assumptions
- Breaking changes are acceptable.

## Acceptance Criteria
- [ ] `work_dir` removed from remote and local config and docs.
- [ ] HCL2 specs regenerated.
- [ ] Build/test still pass after removal.

## Expected areas/files touched
- `provisioner/ansible-navigator/provisioner.go`
- `provisioner/ansible-navigator-local/provisioner.go`
- `docs/CONFIGURATION.md`

