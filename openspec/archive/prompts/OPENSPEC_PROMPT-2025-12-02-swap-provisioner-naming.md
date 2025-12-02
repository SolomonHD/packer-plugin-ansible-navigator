# OpenSpec change prompt

## Context

The plugin's provisioner naming is inverted from Packer conventions. The official hashicorp/ansible plugin uses:
- `ansible` - runs **from** your local machine, connects **to** target via SSH (default)
- `ansible-local` - runs **on** the target machine with local connection

But this plugin currently has it backwards:
- `ansible-navigator` (DEFAULT) - runs **on** target (should be "-local")
- `ansible-navigator-remote` - runs **from** local machine (should be default)

## Goal

Correct the provisioner naming to align with Packer conventions, making `ansible-navigator` the SSH-based remote execution (default) and `ansible-navigator-local` the on-target local execution.

## Scope

In scope:
- Swap registration names in `main.go`
- Rename directory `provisioner/ansible-navigator` → `provisioner/ansible-navigator-local`
- Rename directory `provisioner/ansible-navigator-remote` → `provisioner/ansible-navigator`
- Update Go package names accordingly
- Update all documentation references
- Add feature parity: missing options from remote should be added to local where applicable

Out of scope:
- New functionality beyond parity
- Changes to core provisioning logic

## Desired behavior

After the change:
- `provisioner "ansible-navigator"` runs ansible-navigator on the local machine, connects to target via SSH
- `provisioner "ansible-navigator-local"` runs ansible-navigator directly on the target machine
- Both provisioners have consistent configuration options where applicable
- Documentation clearly explains when to use each

## Constraints & assumptions

- Assumption: This is a major version bump (v3.0.0) due to breaking change
- Assumption: Existing users of `ansible-navigator` will need to switch to `ansible-navigator-local`
- Constraint: Must follow Packer plugin SDK naming conventions
- Constraint: Go package names cannot contain hyphens

## Acceptance criteria

- [ ] `main.go` registers `plugin.DEFAULT_NAME` → remote provisioner (new `ansible-navigator`)
- [ ] `main.go` registers `"local"` → local provisioner (new `ansible-navigator-local`)
- [ ] Directory structure matches: `provisioner/ansible-navigator/` (remote), `provisioner/ansible-navigator-local/` (local)
- [ ] Go packages named `ansiblenavigator` and `ansiblenavigatorlocal` respectively
- [ ] All docs updated to reflect new naming
- [ ] `go build ./...`, `go test ./...`, and `make plugin-check` pass
- [ ] CHANGELOG documents breaking change with migration guide