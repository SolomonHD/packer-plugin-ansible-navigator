# local-provisioner-capabilities Specification Deltas

## REMOVED Requirements

### Requirement: Execution Environment Support

**Reason**: Configuration via string `execution_environment` field is completely removed. Use `navigator_config.execution-environment` object instead.

**Was**: Users could specify `execution_environment = "image:tag"` as a simple string.

**Now**: Must use `navigator_config = { execution-environment = { enabled = true, image = "image:tag" } }`

### Requirement: Navigator Mode

**Reason**: Configuration via `navigator_mode` field is completely removed. Use `navigator_config.mode` instead.

**Was**: Users could specify `navigator_mode = "stdout"` as a simple string.

**Now**: Must use `navigator_config = { mode = "stdout" }`

### Requirement: Ansible Configuration File Generation and Upload

**Reason**: Configuration via `ansible_cfg` map is completely removed. Use `navigator_config.ansible.config` instead.

**Was**: Users could specify `ansible_cfg = { defaults = { ... } }` to control ansible.cfg.

**Now**: Must use `navigator_config = { ansible = { config = { defaults = { ... } } } }`

## MODIFIED Requirements

### Requirement: Configuration Validation

The local provisioner Config.Validate() method SHALL validate all supported configuration options.

#### Scenario: Comprehensive validation

- **WHEN** Config.Validate() is called
- **THEN** it SHALL validate:
  - One or more `play` blocks are defined
  - Each play has a non-empty `target`
  - Any referenced `vars_files` exist on disk (local side)
  - `navigator_config`, if specified, is a non-empty map
  - Command does not contain embedded arguments
  - **AND** removed options SHALL NOT be validated (they will fail HCL parsing)

#### Scenario: Removed options cause validation errors

- **GIVEN** a configuration attempting to use removed options like `execution_environment = "image:tag"`
- **WHEN** Packer parses the configuration
- **THEN** it SHALL fail with an error indicating the option is not recognized
- **AND** error messages SHOULD guide users to use navigator_config instead
- **AND** error messages SHOULD reference MIGRATION.md for help
