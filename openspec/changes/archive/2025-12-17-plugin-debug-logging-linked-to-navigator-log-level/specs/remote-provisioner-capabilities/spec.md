## ADDED Requirements

### Requirement: Plugin debug mode enablement (remote provisioner)

The SSH-based provisioner SHALL enable plugin debug output if and only if `navigator_config.logging.level` is set to `"debug"` (case-insensitive).

#### Scenario: Debug mode enabled only when navigator_config.logging.level is debug

- **GIVEN** a configuration for `provisioner "ansible-navigator"` includes `navigator_config`
- **AND** `navigator_config.logging.level` is set to `"debug"` in any letter casing
- **WHEN** the provisioner determines whether plugin debug mode is enabled
- **THEN** plugin debug mode SHALL be enabled

#### Scenario: Debug mode disabled when logging config is missing or level is not debug

- **GIVEN** a configuration for `provisioner "ansible-navigator"` does not set `navigator_config.logging.level` to `"debug"`
- **WHEN** the provisioner determines whether plugin debug mode is enabled
- **THEN** plugin debug mode SHALL be disabled

### Requirement: Plugin debug output format and sink (remote provisioner)

When plugin debug mode is enabled, the SSH-based provisioner SHALL emit additional diagnostic messages via Packer UI using `ui.Message` and prefix each message with `[DEBUG]`.

#### Scenario: Debug messages use Packer UI and are prefixed

- **GIVEN** plugin debug mode is enabled
- **WHEN** the provisioner emits plugin diagnostic messages intended for debugging
- **THEN** the messages SHALL be emitted via Packer UI `ui.Message`
- **AND** each message SHALL be prefixed with `[DEBUG]`

#### Scenario: Debug messages are gated off when debug mode is disabled

- **GIVEN** plugin debug mode is disabled
- **WHEN** the provisioner executes a build
- **THEN** the additional `[DEBUG]` diagnostic messages SHALL NOT appear in the Packer UI output

### Requirement: Plugin debug output content (remote provisioner)

When plugin debug mode is enabled, the SSH-based provisioner SHALL emit a small, deterministic set of additional diagnostic messages describing key execution decisions, without printing secrets.

#### Scenario: Debug includes command/path and config-file decisions

- **GIVEN** plugin debug mode is enabled
- **WHEN** the provisioner constructs the ansible-navigator command execution context
- **THEN** it SHALL emit debug messages that include:
  - the resolved ansible-navigator executable decision (final `command` and any PATH-prefixing intent)
  - whether `ANSIBLE_NAVIGATOR_CONFIG` is being set (and the path used)

#### Scenario: Debug includes “silent” play execution decisions

- **GIVEN** plugin debug mode is enabled
- **WHEN** the provisioner resolves each play target for execution
- **THEN** it SHALL emit debug messages that include:
  - whether a role target was converted into a generated temporary playbook
  - the absolute playbook path resolution result (when the play target is a playbook path)

#### Scenario: Debug output avoids printing secrets

- **GIVEN** plugin debug mode is enabled
- **WHEN** the provisioner emits debug messages containing user-provided values
- **THEN** it SHALL avoid printing secrets in debug output
- **AND** it SHALL follow the existing sanitization approach used for command/log output
