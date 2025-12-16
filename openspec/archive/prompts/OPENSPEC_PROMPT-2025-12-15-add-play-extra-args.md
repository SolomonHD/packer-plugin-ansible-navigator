# OpenSpec Change Prompt

## Context
The provisioners support structured `play {}` blocks, but they do not provide an escape hatch for uncommon `ansible-navigator run` flags.
Without a generic per-play option, users must request new config fields for each flag, which will cause configuration sprawl.

## Goal
Add a per-play `extra_args` list that is appended to the `ansible-navigator run` invocation for that play.

## Scope

**In scope:**
- Add `extra_args` (list(string)) to the `play {}` block schema for both provisioners.
- Define and document ordering/precedence relative to plugin-generated args.
- Regenerate HCL2 spec files.
- Update docs and examples.

**Out of scope:**
- Global `extra_args` for the whole provisioner.
- Legacy compatibility with removed options.

## Desired Behavior

```hcl
provisioner "ansible-navigator" {
  play {
    target = "site.yml"
    extra_args = ["--check", "--diff"]
  }
}
```

- The provisioner MUST pass the `extra_args` for that play verbatim.
- The spec MUST define deterministic ordering, e.g.:
  1) `ansible-navigator run` (and enforced `--mode` behavior)
  2) play-level `extra_args`
  3) plugin-generated inventory/extra-vars/etc
  4) playbook/role target

## Constraints & Assumptions
- `extra_args` values are trusted (no validation beyond basic type/empty checks).
- Ordering must be consistent across both provisioners.

## Acceptance Criteria
- [ ] `play.extra_args` exists in both provisioners’ HCL schema.
- [ ] `play.extra_args` affects the executed command for that play.
- [ ] HCL2 specs regenerated.
- [ ] Docs updated with at least one example.

## Expected areas/files touched
- `provisioner/ansible-navigator/provisioner.go`
- `provisioner/ansible-navigator-local/provisioner.go`
- `docs/CONFIGURATION.md`

