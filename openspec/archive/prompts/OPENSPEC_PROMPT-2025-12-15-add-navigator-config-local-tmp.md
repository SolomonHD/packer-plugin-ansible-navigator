# OpenSpec Change Prompt

## Context
`navigator_config` is a typed representation of `ansible-navigator.yml` settings.
The plugin currently supports `ansible_config.defaults.remote_tmp`, but execution environments commonly also require `ansible_config.defaults.local_tmp` to avoid non-writable default paths.

## Goal
Add support for `navigator_config.ansible_config.defaults.local_tmp` end-to-end.

## Scope

**In scope:**
- Extend the typed structs to include `local_tmp` under `navigator_config.ansible_config.defaults`.
- Ensure YAML generation includes `local_tmp` when set.
- Regenerate HCL2 specs.
- Update docs/examples.

**Out of scope:**
- Adding new top-level legacy Ansible temp settings outside `navigator_config`.

## Desired Behavior

```hcl
provisioner "ansible-navigator" {
  navigator_config {
    ansible_config {
      defaults {
        remote_tmp = "/tmp/.ansible/tmp"
        local_tmp  = "/tmp/.ansible-local"
      }
    }
  }

  play { target = "site.yml" }
}
```

- Generated `ansible-navigator.yml` includes both values under the correct YAML structure.

## Constraints & Assumptions
- The `navigator_config` field remains the preferred way to set these values.

## Acceptance Criteria
- [ ] `local_tmp` is available in the HCL schema under `navigator_config.ansible_config.defaults.local_tmp`.
- [ ] YAML generation writes `local_tmp` when set.
- [ ] HCL2 specs regenerated.
- [ ] Docs updated.

## Expected areas/files touched
- `provisioner/ansible-navigator/provisioner.go`
- `provisioner/ansible-navigator-local/provisioner.go`
- `provisioner/ansible-navigator/navigator_config.go` (and local equivalent if present)

