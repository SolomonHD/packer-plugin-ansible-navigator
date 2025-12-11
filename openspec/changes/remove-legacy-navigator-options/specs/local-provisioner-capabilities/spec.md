# local-provisioner-capabilities Specification Deltas

## REMOVED Requirements

### Requirement: Execution Environment Support (via string field)

**Reason**: Configuration via string `execution_environment` field is completely removed. Use `navigator_config.execution-environment` object instead.

**Was**: Users could specify `execution_environment = "image:tag"` as a simple string.

**Now**: Must use `navigator_config = { execution-environment = { enabled = true, image = "image:tag" } }`

### Requirement: Navigator Mode Support (via string field)

**Reason**: Configuration via `navigator_mode` field is completely removed. Use `navigator_config.mode` instead.

**Was**: Users could specify `navigator_mode = "stdout"` as a simple string.

**Now**: Must use `navigator_config = { mode = "stdout" }`

### Requirement: Ansible Configuration File Generation and Upload (via ansible_cfg)

**Reason**: Configuration via `ansible_cfg` map is completely removed. Use `navigator_config.ansible.config` instead.

**Was**: Users could specify `ansible_cfg = { defaults = { ... } }` to control ansible.cfg.

**Now**: Must use `navigator_config = { ansible = { config = { defaults = { ... } } } }`

### Requirement: Ansible Environment Variables (via ansible_env_vars)

**Reason**: Configuration via `ansible_env_vars` map is completely removed. Use `navigator_config.execution-environment.environment-variables` instead.

**Was**: Users could specify `ansible_env_vars = { VAR = "value" }`.

**Now**: Must use `navigator_config = { execution-environment = { environment-variables = { VAR = "value" } } }`

### Requirement: Ansible SSH Extra Args Support

**Reason**: `ansible_ssh_extra_args` field is completely removed. Configure via navigator_config or play-level settings.

### Requirement: Extra Arguments Support

**Reason**: `extra_arguments` field is completely removed. Use navigator_config to control all ansible-navigator settings.

### Requirement: Roles Path Configuration

**Reason**: `roles_path` field is completely removed. Use navigator_config or requirements_file for dependency management.

### Requirement: Collections Path Configuration

**Reason**: `collections_path` field is completely removed. Use navigator_config or requirements_file for dependency management.

### Requirement: Galaxy Command Customization

**Reason**: `galaxy_command` field is completely removed. Use requirements_file workflow instead.

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
