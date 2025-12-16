# Change: Warn when `skip_version_check` makes `version_check_timeout` ineffective

## Why

The plugin supports both `skip_version_check` and `version_check_timeout`. When `skip_version_check = true`, the version check is bypassed and `version_check_timeout` becomes ineffective. Today this creates a configuration footgun because users can set both and receive no feedback that one setting is ignored.

## What Changes

- Emit a **user-visible warning** when:
  - `skip_version_check = true`, **and**
  - `version_check_timeout` is **explicitly configured** by the user.
- Ensure the warning is surfaced in Packer UI output (not only debug logs).
- Extend the local provisioner (`ansible-navigator-local`) configuration surface to include `skip_version_check` for parity, even though the local provisioner does not currently perform a version check.

## Scope

### In scope

- Warning emitted during configuration validation / prepare for:
  - `provisioner "ansible-navigator"` (remote)
  - `provisioner "ansible-navigator-local"` (local)
- Warning is non-fatal (does not fail the build).

### Out of scope

- Changing default timeout behavior.
- Removing or deprecating either `skip_version_check` or `version_check_timeout`.
- Adding local provisioner version-check execution.

## Impact

### Affected OpenSpec capabilities

- `remote-provisioner-capabilities` (add warning behavior requirement)
- `local-provisioner-capabilities` (add `skip_version_check` config surface + warning behavior requirement)

### Expected areas/files touched (implementation)

- `provisioner/ansible-navigator/provisioner.go`
- `provisioner/ansible-navigator-local/provisioner.go`

## Notes

- The warning MUST only trigger when `version_check_timeout` was explicitly set by the user. Because `version_check_timeout` has a default value, implementation may need to distinguish "unset" vs "set" (e.g., pointer field or decode metadata), otherwise the warning would trigger for every `skip_version_check = true` configuration.