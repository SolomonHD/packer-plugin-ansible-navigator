# local-provisioner-capabilities Spec Delta

## ADDED Requirements

### Requirement: ansible-navigator Version 2 Schema Compliance for Local Provisioner

The local provisioner SHALL generate ansible-navigator.yml configuration files that include ansible-navigator Version 2 schema markers to ensure immediate recognition by ansible-navigator 25.x without triggering version migration prompts.

#### Scenario: Version 2 schema marker included in generated YAML

- **GIVEN** a configuration with `provisioner "ansible-navigator-local"` and `navigator_config` block
- **WHEN** the provisioner generates the `ansible-navigator.yml` file on the target
- **THEN** the generated YAML SHALL include schema version markers that identify it as Version 2 format
- **AND** ansible-navigator 25.12.0+ SHALL recognize the file as Version 2 format immediately
- **AND** NO version migration prompts SHALL appear

#### Scenario: Pull-policy Version 2 nested structure preserved for local provisioner

- **GIVEN** a configuration with `provisioner "ansible-navigator-local"` and `navigator_config.execution_environment.pull_policy = "never"`
- **WHEN** the provisioner generates the `ansible-navigator.yml` file
- **THEN** the generated YAML SHALL use the Version 2 nested structure:
  ```yaml
  ansible-navigator:
    execution-environment:
      pull:
        policy: never
  ```
- **AND** ansible-navigator SHALL respect the pull-policy setting without attempting registry pulls
- **AND** Docker SHALL NOT attempt to pull images when local images exist on the target

#### Scenario: Version 2 format works with all pull-policy values on target

- **GIVEN** a configuration with `provisioner "ansible-navigator-local"` and `navigator_config.execution_environment.pull_policy` set to any valid value
- **WHEN** the provisioner generates the `ansible-navigator.yml` file on the target
- **THEN** the generated YAML SHALL use the Version 2 nested `pull.policy` structure
- **AND** ansible-navigator running on the target SHALL correctly interpret the pull-policy value
- **AND** Docker pull behavior SHALL match the specified policy

#### Scenario: No legacy Version 1 field names in local provisioner generated YAML

- **GIVEN** any valid `navigator_config` configuration for `provisioner "ansible-navigator-local"`
- **WHEN** the provisioner generates the `ansible-navigator.yml` file
- **THEN** the generated YAML SHALL NOT contain legacy Version 1 field names or structures
- **AND** ALL field names SHALL conform to Version 2 schema conventions
- **AND** ansible-navigator 25.x running on the target SHALL accept the configuration without schema validation warnings

#### Scenario: Local provisioner Version 2 format identical to remote provisioner

- **GIVEN** identical `navigator_config` blocks for both `provisioner "ansible-navigator"` and `provisioner "ansible-navigator-local"`
- **WHEN** both provisioners generate their respective `ansible-navigator.yml` files
- **THEN** both generated YAML files SHALL have identical Version 2 schema markers and structure
- **AND** both SHALL be recognized as Version 2 format by ansible-navigator
- **AND** NO differences in version recognition SHALL exist between provisioners
