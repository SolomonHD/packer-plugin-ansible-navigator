# Context

The Packer provisioners in this repo emit a mix of user-facing output (via Packer UI) and internal logs.

- User-facing output uses `ui.Say` / `ui.Message` / `ui.Error` (e.g. [`Provisioner.Provision()`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/provisioner.go:959)).
- Some diagnostics use Go’s standard logger (`log.Printf`) (e.g. inventory creation in [`Provisioner.createInventoryFile()`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/provisioner.go:897)).

Users running the provisioner inside Dagger want **more diagnostic output when they increase log level**, but **do not want a new plugin-specific log-level option**.

The repo already has a natural “log level” input: `navigator_config.logging.level` (part of [`LoggingConfig`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/provisioner.go:121)). This is intended for ansible-navigator, but we will explicitly **link** it to the plugin’s debug output to reduce configuration surface.

# Goal

Add a DEBUG logging mode for the plugin that is enabled when (and only when) `navigator_config.logging.level` is set to `"debug"` (case-insensitive), and document that **ansible-navigator and plugin logging are linked**.

# Scope

## In scope

- Add a single internal mechanism to determine whether plugin debug logging is enabled.
- When debug logging is enabled, emit additional **plugin** diagnostic output via Packer UI (prefixed with `[DEBUG]`).
- Update docs to clearly state the linkage: `navigator_config.logging.level` affects both ansible-navigator logging and plugin debug output.
- Apply consistently to both provisioners:
  - SSH-based: [`provisioner/ansible-navigator/`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/provisioner.go:9)
  - On-target: `provisioner/ansible-navigator-local/` (parallel behavior)

## Out of scope

- Adding a new `log_level` / `debug` field to the plugin’s config.
- Changing ansible-navigator behavior beyond existing config generation.
- Changing structured JSON parsing behavior (controlled by `structured_logging` / `verbose_task_output`).

# Desired Behavior

## Debug enablement

The plugin is considered in **debug mode** iff:

- `navigator_config` is set, and
- `navigator_config.logging` is set, and
- `navigator_config.logging.level` equals `"debug"` (case-insensitive).

## Debug output

When debug mode is enabled, the plugin SHALL emit additional diagnostic UI messages including:

- The resolved ansible-navigator executable path decision (final `command` and any PATH prefixing intent).
- Whether `ANSIBLE_NAVIGATOR_CONFIG` is being set (and the path used).
- Key execution decisions that are otherwise “silent”, such as:
  - whether a role target was converted to a generated temporary playbook
  - the absolute playbook path resolution result

All debug output MUST:

- use the Packer UI (`ui.Message`) so it is visible to users in real time
- be prefixed with `[DEBUG]`
- avoid printing secrets (follow the existing sanitization patterns in [`Provisioner.executeAnsibleCommand()`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/provisioner.go:1375))

# Constraints & Assumptions

- Debug output should not materially change behavior; it is informational only.
- Debug output must be deterministic and avoid flooding output.
- The linkage between ansible-navigator logging and plugin debug logging must be documented prominently to reduce confusion.

# Acceptance Criteria

- [ ] Debug mode is enabled when `navigator_config.logging.level = "debug"` (case-insensitive), and disabled otherwise.
- [ ] When debug mode is enabled, the plugin emits additional `[DEBUG]` messages at the key decision points described above.
- [ ] When debug mode is not enabled, these extra debug messages do not appear.
- [ ] Documentation is updated to explicitly state that `navigator_config.logging.level` controls both:
  - ansible-navigator logging, and
  - plugin debug output.
- [ ] Unit tests exist that validate debug enablement and that debug messages are gated correctly.

## Expected files/areas touched

- [`provisioner/ansible-navigator/provisioner.go`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/provisioner.go:1)
- `provisioner/ansible-navigator-local/provisioner.go`
- [`docs/CONFIGURATION.md`](packer/plugins/packer-plugin-ansible-navigator/docs/CONFIGURATION.md:117)

## Dependencies

- None
