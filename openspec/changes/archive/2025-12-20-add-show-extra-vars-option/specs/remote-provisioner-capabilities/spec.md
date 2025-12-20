## ADDED Requirements

### Requirement: Show Extra Vars JSON in Output Log (remote provisioner)

The SSH-based provisioner SHALL support a `show_extra_vars` configuration option that, when enabled, logs the complete extra vars JSON object to Packer UI output before executing ansible-navigator.

#### Scenario: show_extra_vars option accepted in HCL schema

- **GIVEN** a configuration using `provisioner "ansible-navigator"`
- **AND** the configuration includes `show_extra_vars = true`
- **WHEN** Packer parses the configuration
- **THEN** parsing SHALL succeed
- **AND** the provisioner configuration SHALL include the `show_extra_vars` value

#### Scenario: Extra vars JSON logged when show_extra_vars is true

- **GIVEN** a configuration using `provisioner "ansible-navigator"` with `show_extra_vars = true`
- **AND** the provisioner has constructed extra vars including Packer-injected variables (e.g., `packer_build_name`, `packer_builder_type`)
- **AND** the provisioner may also include user-defined play-level extra vars
- **WHEN** the provisioner prepares to execute ansible-navigator for a play
- **THEN** it SHALL emit the extra vars JSON object to Packer UI via `ui.Message()`
- **AND** the output SHALL be prefixed with a clear identifier (e.g., `[Extra Vars]`)
- **AND** the JSON SHALL be formatted with indentation for human readability

#### Scenario: Sensitive values are redacted in extra vars output

- **GIVEN** a configuration using `provisioner "ansible-navigator"` with `show_extra_vars = true`
- **AND** the extra vars include sensitive values (e.g., `ansible_password`)
- **WHEN** the provisioner logs the extra vars JSON
- **THEN** it SHALL redact the `ansible_password` value by replacing it with `*****`
- **AND** any other known sensitive keys SHALL be similarly redacted
- **AND** non-sensitive values like `ansible_ssh_private_key_file` (path only, not content) SHALL be shown

#### Scenario: Extra vars not logged when show_extra_vars is false or unset

- **GIVEN** a configuration using `provisioner "ansible-navigator"`
- **AND** `show_extra_vars` is not set OR `show_extra_vars = false`
- **WHEN** the provisioner executes ansible-navigator
- **THEN** it SHALL NOT emit the extra vars JSON to Packer UI output
- **AND** existing behavior SHALL be preserved

#### Scenario: show_extra_vars defaults to false

- **GIVEN** a configuration using `provisioner "ansible-navigator"` without setting `show_extra_vars`
- **WHEN** the provisioner configuration is prepared
- **THEN** `show_extra_vars` SHALL default to `false`
- **AND** no extra vars JSON output SHALL be produced
