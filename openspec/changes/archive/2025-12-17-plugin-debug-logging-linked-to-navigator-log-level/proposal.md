# Change: Link Plugin Debug Logging to `navigator_config.logging.level`

## Why

Users running the provisioners under Dagger want more diagnostic output when troubleshooting, but they do not want a separate plugin-specific log level knob.

The configuration already includes a natural “log level” input via `navigator_config.logging.level` (intended for ansible-navigator). This change explicitly links that field to the plugin’s debug output so users can enable diagnostics without adding new configuration surface.

## What Changes

- Add a single internal mechanism to determine whether plugin debug mode is enabled.
- Define plugin debug mode as enabled **if and only if** `navigator_config.logging.level` is `"debug"` (case-insensitive).
- When enabled, emit additional diagnostic output via Packer UI, prefixed with `[DEBUG]`.
- Apply consistently to both provisioners:
  - `ansible-navigator` (SSH-based)
  - `ansible-navigator-local` (on-target)
- Document the linkage prominently: `navigator_config.logging.level` controls both ansible-navigator logging and plugin debug output.

## Non-Goals

- Adding a new plugin config field such as `debug = true` or `log_level`.
- Changing the meaning of `structured_logging` / JSON parsing behavior.
- Changing ansible-navigator behavior beyond its existing config generation.

## Impact

- Affected specs:
  - `remote-provisioner-capabilities`
  - `local-provisioner-capabilities`
  - `documentation`
- Affected implementation areas (for follow-on implementation task):
  - [`provisioner/ansible-navigator/provisioner.go`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/provisioner.go:1)
  - `provisioner/ansible-navigator-local/provisioner.go`
  - [`docs/CONFIGURATION.md`](packer/plugins/packer-plugin-ansible-navigator/docs/CONFIGURATION.md:1)
