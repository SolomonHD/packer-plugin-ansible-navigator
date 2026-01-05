# remote-provisioner-capabilities Spec Delta

## ADDED Requirements

### Requirement: ansible-navigator Version 2 Schema Compliance

The remote provisioner SHALL generate ansible-navigator.yml configuration files that include ansible-navigator Version 2 schema markers to ensure immediate recognition by ansible-navigator 25.x without triggering version migration prompts.

#### Scenario: Version 2 schema marker included in generated YAML

- **GIVEN** a configuration with `navigator_config` block containing any settings
- **WHEN** the provisioner generates the `ansible-navigator.yml` file  
- **THEN** the generated YAML SHALL include schema version markers that identify it as Version 2 format
- **AND** ansible-navigator 25.12.0+ SHALL recognize the file as Version 2 format immediately
- **AND** NO version migration prompts SHALL appear

#### Scenario: Pull-policy Version 2 nested structure preserved

- **GIVEN** a configuration with `navigator_config.execution_environment.pull_policy = "never"`
- **WHEN** the provisioner generates the `ansible-navigator.yml` file
- **THEN** the generated YAML SHALL use the Version 2 nested structure:
  ```yaml
  ansible-navigator:
    execution-environment:
      pull:
        policy: never
  ```
- **AND** ansible-navigator SHALL respect the pull-policy setting without attempting registry pulls
- **AND** Docker SHALL NOT attempt to pull images when local images exist

#### Scenario: Version 2 format works with all pull-policy values

- **GIVEN** a configuration with `navigator_config.execution_environment.pull_policy` set to any valid value (`"always"`, `"missing"`, `"never"`, `"tag"`)
- **WHEN** the provisioner generates the `ansible-navigator.yml` file
- **THEN** the generated YAML SHALL use the Version 2 nested `pull.policy` structure
- **AND** ansible-navigator SHALL correctly interpret the pull-policy value
- **AND** Docker pull behavior SHALL match the specified policy

#### Scenario: No legacy Version 1 field names in generated YAML

- **GIVEN** any valid `navigator_config` configuration
- **WHEN** the provisioner generates the `ansible-navigator.yml` file
- **THEN** the generated YAML SHALL NOT contain legacy Version 1 field names or structures
- **AND** ALL field names SHALL conform to Version 2 schema conventions
- **AND** ansible-navigator 25.x SHALL accept the configuration without schema validation warnings

#### Scenario: Version 2 format backward compatible with v24.x

- **GIVEN** a generated ansible-navigator.yml file with Version 2 schema markers
- **WHEN** the file is used with ansible-navigator 24.x or later
- **THEN** ansible-navigator SHALL accept and process the configuration correctly
- **AND** Version 2 format SHALL NOT break compatibility with ansible-navigator 24.x installations

#### Scenario: Temporary config files remain Version 2 format

- **GIVEN** the provisioner creates temporary `/tmp/packer-ansible-navigator-*.yml` files
- **WHEN** these temporary files are generated
- **THEN** they SHALL include Version 2 schema markers
- **AND** ansible-navigator SHALL process them without migration attempts
- **AND** the temporary nature of the files SHALL NOT prevent correct Version 2 recognition
