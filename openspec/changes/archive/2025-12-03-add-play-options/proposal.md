# Change: Add become_user and skip_tags to Play configuration

## Why

The `Play` struct currently lacks support for `become_user` and `skip_tags`, which are common Ansible parameters needed for more granular control over playbook execution. Users need these options to execute plays as specific users or to skip certain tasks during provisioning.

## What Changes

- Update `Play` struct to include `BecomeUser` (string) and `SkipTags` ([]string).
- Update `HCL2Spec` generation to reflect the new fields.
- Update execution logic to pass `--become-user` and `--skip-tags` to the underlying Ansible command.

## Impact

- Affected specs: `local-provisioner-capabilities`
- Affected code: `provisioner/ansible-navigator/provisioner.go`
