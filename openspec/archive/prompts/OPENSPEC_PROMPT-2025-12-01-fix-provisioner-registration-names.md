# OpenSpec change prompt

## Context

The `packer-plugin-ansible-navigator` plugin registers provisioners with explicit names (`ansible-navigator`, `ansible-navigator-remote`) in `main.go`. This causes awkward HCL usage because Packer prefixes provisioner names with the `required_plugins` alias, resulting in names like `ansible-navigator-ansible-navigator`.

## Goal

Fix provisioner registration to follow Packer SDK naming conventions so users can reference provisioners with clean names like `ansible-navigator` and `ansible-navigator-local`.

## Scope

- In scope:
  - Update `main.go` provisioner registrations
  - Update documentation to reflect correct usage
  - Verify with `describe` command
- Out of scope:
  - Provisioner logic/behavior changes
  - Adding new provisioners

## Desired behavior

After the change:
- `RegisterProvisioner(plugin.DEFAULT_NAME, ...)` for the local/main provisioner ΓåÆ accessible as `ansible-navigator` in HCL
- `RegisterProvisioner("local", ...)` for local provisioner ΓåÆ accessible as `ansible-navigator-local`  
- `RegisterProvisioner("remote", ...)` for remote provisioner ΓåÆ accessible as `ansible-navigator-remote`
- Running `./plugin describe` shows provisioners as `["-packer-default-plugin-name-" "local" "remote"]` or equivalent

## Constraints & assumptions

- Assume this follows the pattern used by `hashicorp/packer-plugin-ansible` which registers `-packer-default-plugin-name-` and `local`
- The constant `plugin.DEFAULT_NAME` equals `"-packer-default-plugin-name-"`
- Renaming may require a minor version bump (API-compatible but changes user-facing naming)

## Acceptance criteria

- [ ] `main.go` uses `plugin.DEFAULT_NAME` for the primary provisioner
- [ ] Secondary provisioner uses short name (`local` or `remote`) that gets prefixed by Packer
- [ ] `./packer-plugin-ansible-navigator describe` output reflects new registration names
- [ ] `go build ./...` and `make plugin-check` pass
- [ ] README/docs updated with correct HCL usage examples