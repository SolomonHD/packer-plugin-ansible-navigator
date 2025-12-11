# remote-provisioner-capabilities Specification Deltas

## REMOVED Requirements

### Requirement: Navigator Mode Support

**Reason**: Configuration via `navigator_mode` field is completely removed. Use `navigator_config.mode` instead.

**Was**: Users could specify `navigator_mode = "stdout"` as a simple string.

**Now**: Must use `navigator_config = { mode = "stdout" }`

### Requirement: Ansible Configuration File Generation

**Reason**: Configuration via `ansible_cfg` map is completely removed. Use `navigator_config.ansible.config` instead.

**Was**: Users could specify `ansible_cfg = { defaults = { ... } }` to control ansible.cfg.

**Now**: Must use `navigator_config = { ansible = { config = { defaults = { ... } } } }`

## MODIFIED Requirements

### Requirement: Configuration Validation

The remote provisioner Config.Validate() method SHALL validate all supported configuration options.

#### Scenario: Comprehensive validation

- **WHEN** Config.Validate() is called
- **THEN** it SHALL validate:
  - One or more `play` blocks are defined
  - Each play has a non-empty `target`
  - SSH connection parameters are valid
  - `navigator_config`, if specified, is a non-empty map
  - Command does not contain embedded arguments
  - **AND** removed options SHALL NOT be validated (they will fail HCL parsing)

#### Scenario: Removed options cause validation errors

- **GIVEN** a configuration attempting to use removed options like `execution_environment = "image:tag"`
- **WHEN** Packer parses the configuration
- **THEN** it SHALL fail with an error indicating the option is not recognized
- **AND** error messages SHOULD guide users to use navigator_config instead
- **AND** error messages SHOULD reference MIGRATION.md for help

### Requirement: Default Navigator Command

The SSH-based provisioner SHALL use ansible-navigator as its default executable, pass `run` as the first argument, and treat the `command` field strictly as the ansible-navigator executable name or path (without embedded arguments).

#### Scenario: Default command construction

- **GIVEN** a Packer configuration using `provisioner "ansible-navigator"`
- **AND** the `command` field is not specified
- **WHEN** the provisioner constructs the command to run ansible-navigator
- **THEN** it SHALL invoke ansible-navigator using `exec.Command("ansible-navigator", "run", ...)` on the local machine
- **AND** configuration SHALL be controlled ONLY via ANSIBLE_NAVIGATOR_CONFIG environment variable
- **AND** NO mode, EE, or other CLI flags SHALL be passed (all settings come from config file)
- **AND** Legacy CLI flags (`--mode`, `--ee`, `--eei`) SHALL NOT be generated
