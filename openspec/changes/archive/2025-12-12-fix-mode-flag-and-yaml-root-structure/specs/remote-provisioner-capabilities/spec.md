# remote-provisioner-capabilities

## ADDED Requirements

### Requirement: Mode CLI Flag Support for Remote Provisioner

The SSH-based provisioner SHALL pass the `--mode` CLI flag to ansible-navigator when `navigator_config.mode` is configured, ensuring ansible-navigator runs in the specified mode and does not hang waiting for interactive input.

#### Scenario: Mode flag added when navigator_config.mode is set

- **GIVEN** a configuration with `provisioner "ansible-navigator"`
- **AND** `navigator_config.mode = "stdout"` is specified
- **WHEN** the provisioner constructs the ansible-navigator command
- **THEN** it SHALL include `--mode stdout` flag in the command arguments
- **AND** the flag SHALL be inserted after the `run` subcommand
- **AND** the flag SHALL appear before playbook-specific arguments (inventory, extra vars, etc.)

#### Scenario: Mode flag not added when mode not configured

- **GIVEN** a configuration with `provisioner "ansible-navigator"`
- **AND** NO `navigator_config` block is specified
- **OR** `navigator_config` is specified but does NOT include `mode` field
- **WHEN** the provisioner constructs the ansible-navigator command
- **THEN** it SHALL NOT include a `--mode` flag
- **AND** ansible-navigator SHALL use its default behavior (config file or built-in defaults)

#### Scenario: Mode flag prevents interactive hang

- **GIVEN** a configuration with `navigator_config.mode = "stdout"`
- **AND** ansible-navigator is run in a non-interactive environment (Packer build)
- **WHEN** the provisioner executes ansible-navigator
- **THEN** ansible-navigator SHALL execute in stdout mode
- **AND** it SHALL NOT wait for terminal input
- **AND** it SHALL output playbook execution results to stdout
- **AND** the provisioning SHALL complete without hanging

### Requirement: YAML Root Structure for Remote Provisioner

The remote provisioner SHALL generate ansible-navigator.yml configuration files with a top-level `ansible-navigator:` root key wrapping all settings to conform to ansible-navigator v25.x schema requirements and prevent validation errors.

#### Scenario: Generated YAML wraps all settings under ansible-navigator key

- **GIVEN** a configuration with `navigator_config` block containing any settings
- **WHEN** the provisioner generates the `ansible-navigator.yml` file
- **THEN** the YAML SHALL have a top-level `ansible-navigator:` key
- **AND** ALL configuration settings SHALL be nested under this key
- **AND** the structure SHALL conform to ansible-navigator v25.x schema

#### Scenario: Mode setting nested under root key

- **GIVEN** a configuration with `navigator_config.mode = "stdout"`
- **WHEN** the YAML is generated
- **THEN** the output SHALL be:

  ```yaml
  ansible-navigator:
    mode: stdout
  ```

- **AND** NOT the flat structure:

  ```yaml
  mode: stdout
  ```

#### Scenario: Multiple configuration sections nested correctly

- **GIVEN** a configuration with multiple navigator_config sections (mode, execution_environment, ansible_config, logging)
- **WHEN** the YAML is generated
- **THEN** ALL sections SHALL be nested under `ansible-navigator:` root key
- **AND** nested structure SHALL be preserved for execution-environment, ansible, and other complex settings

#### Scenario: Schema validation passes with root key

- **GIVEN** a generated ansible-navigator.yml file with `ansible-navigator:` root key
- **WHEN** ansible-navigator processes the configuration file
- **THEN** it SHALL pass schema validation
- **AND** it SHALL NOT report "Additional properties are not allowed" errors
- **AND** all configuration settings SHALL be recognized and applied

#### Scenario: Execution environment pull policy with root key

- **GIVEN** a configuration with `execution_environment.pull_policy = "missing"`
- **WHEN** the ansible-navigator.yml YAML is generated
- **THEN** the YAML SHALL contain:

  ```yaml
  ansible-navigator:
    execution-environment:
      pull:
        policy: missing
  ```

- **AND** the generated YAML SHALL pass ansible-navigator's built-in schema validation
- **AND** ansible-navigator SHALL accept the config file without "Additional properties" errors

#### Scenario: ConvertToYAMLStructure wraps in root key

- **GIVEN** the `convertToYAMLStructure()` function implementation
- **WHEN** it converts NavigatorConfig to YAML-compatible structure
- **THEN** it SHALL create a top-level map with key `"ansible-navigator"`
- **AND** the value SHALL be a map containing all navigator settings
- **AND** nested structures SHALL be preserved within this wrapped structure
- **AND** field name conversions (underscores to hyphens) SHALL occur correctly

#### Scenario: Empty navigator_config produces minimal YAML

- **GIVEN** a `navigator_config {}` block with no fields set
- **WHEN** the YAML is generated
- **THEN** it SHALL produce:

  ```yaml
  ansible-navigator: {}
  ```

- **OR** an equivalent minimal structure
- **AND** ansible-navigator SHALL accept this as valid configuration

#### Scenario: Backward compatibility maintained

- **GIVEN** an existing Packer configuration using `navigator_config` block
- **AND** the configuration was written before this change
- **WHEN** the configuration is used with the updated plugin
- **THEN** it SHALL continue to work without modification
- **AND** the YAML SHALL be generated with proper root structure automatically
- **AND** no user action SHALL be required to adopt the new structure
