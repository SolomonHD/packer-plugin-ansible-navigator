# local-provisioner-capabilities

## ADDED Requirements

### Requirement: Mode CLI Flag Support for Local Provisioner

The local provisioner SHALL pass the `--mode` CLI flag in the remote shell command when `navigator_config.mode` is configured, preventing ansible-navigator on the target from hanging in interactive mode.

#### Scenario: Mode flag added in remote shell command

- **GIVEN** a configuration with `provisioner "ansible-navigator-local"`
- **AND** `navigator_config.mode = "stdout"` is specified
- **WHEN** the provisioner constructs the remote shell command for ansible-navigator
- **THEN** it SHALL include `--mode stdout` flag in the command
- **AND** the flag SHALL be positioned after the `run` subcommand
- **AND** the flag SHALL appear before playbook/target arguments

#### Scenario: Mode flag not added when mode not configured

- **GIVEN** a configuration with `provisioner "ansible-navigator-local"`
- **AND** NO `navigator_config.mode` is specified
- **WHEN** the provisioner constructs the remote shell command
- **THEN** it SHALL NOT include a `--mode` flag
- **AND** ansible-navigator on the target SHALL use its default behavior

#### Scenario: Mode flag prevents hang on target

- **GIVEN** a configuration with `navigator_config.mode = "stdout"`
- **AND** ansible-navigator is run on the target in a non-interactive SSH session
- **WHEN** the provisioner executes
- **THEN** ansible-navigator SHALL execute in stdout mode on the target
- **AND** it SHALL NOT wait for terminal input
- **AND** execution SHALL complete without hanging

### Requirement: YAML Root Structure for Local Provisioner

The local provisioner SHALL generate ansible-navigator.yml configuration files with a top-level `ansible-navigator:` root key wrapping all settings to conform to ansible-navigator v25.x schema requirements, preventing validation errors.

#### Scenario: Generated YAML wraps all settings under ansible-navigator key

- **GIVEN** a configuration with `provisioner "ansible-navigator-local"`
- **AND** `navigator_config` block with any settings
- **WHEN** the provisioner generates the ansible-navigator.yml file
- **THEN** the YAML SHALL have `ansible-navigator:` root key
- **AND** all settings SHALL be nested under this key
- **AND** the file SHALL be uploaded to the target staging directory

#### Scenario: Consistent YAML structure with remote provisioner

- **GIVEN** identical `navigator_config` blocks in local and remote provisioner configs
- **WHEN** both provisioners generate YAML files
- **THEN** the YAML content SHALL be identical
- **AND** both SHALL use `ansible-navigator:` root key
- **AND** both SHALL produce valid ansible-navigator v25.x configuration

#### Scenario: ConvertToYAMLStructure implementation matches remote

- **GIVEN** the local provisioner's `convertToYAMLStructure()` function
- **WHEN** it converts NavigatorConfig to YAML-compatible structure
- **THEN** it SHALL create a top-level map with key `"ansible-navigator"`
- **AND** the value SHALL be a map containing all navigator settings
- **AND** the implementation SHALL be identical to the remote provisioner's function
- **AND** nested structures SHALL be preserved within this wrapped structure

#### Scenario: Schema validation passes for local provisioner YAML

- **GIVEN** a generated ansible-navigator.yml file with `ansible-navigator:` root key
- **AND** the file is uploaded to the target
- **WHEN** ansible-navigator on the target processes the configuration file
- **THEN** it SHALL pass schema validation
- **AND** it SHALL NOT report "Additional properties are not allowed" errors
- **AND** all configuration settings SHALL be recognized and applied

#### Scenario: Backward compatibility for local provisioner configs

- **GIVEN** an existing Packer configuration using `provisioner "ansible-navigator-local"` with `navigator_config`
- **AND** the configuration was written before this change
- **WHEN** the configuration is used with the updated plugin
- **THEN** it SHALL continue to work without modification
- **AND** the YAML SHALL be generated with proper root structure automatically
