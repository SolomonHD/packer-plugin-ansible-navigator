# Execution Environment Configuration

## ADDED Requirements

### Requirement: Execution Environment Support
The provisioner SHALL support specifying a custom execution environment container image for ansible-navigator to use when running playbooks.

#### Scenario: User specifies execution environment
- **WHEN** a user configures `execution_environment = "quay.io/ansible/creator-ee:latest"` in their Packer template
- **THEN** the provisioner SHALL pass `--execution-environment quay.io/ansible/creator-ee:latest` to the ansible-navigator command

#### Scenario: User does not specify execution environment
- **WHEN** a user does not configure the `execution_environment` field
- **THEN** the provisioner SHALL NOT include the `--execution-environment` flag in the ansible-navigator command
- **AND** ansible-navigator SHALL use its default execution environment selection behavior

#### Scenario: User specifies custom registry execution environment
- **WHEN** a user configures `execution_environment = "my-registry.example.com/ansible-ee:v2.0"` with a custom container registry
- **THEN** the provisioner SHALL pass `--execution-environment my-registry.example.com/ansible-ee:v2.0` to the ansible-navigator command

### Requirement: Execution Environment Flag Placement
The provisioner SHALL place the `--execution-environment` flag in the correct position within the ansible-navigator command arguments.

#### Scenario: Flag placement with other arguments
- **WHEN** the execution environment is specified along with other ansible-navigator arguments
- **THEN** the `--execution-environment` flag SHALL be placed after the `--mode` flag but before the playbook path argument
- **AND** the command structure SHALL follow: `ansible-navigator run --mode <mode> --execution-environment <image> [other-args] -i <inventory> <playbook>`

### Requirement: Configuration Field Documentation
The provisioner SHALL provide clear documentation for the `execution_environment` configuration field.

#### Scenario: Documentation describes field purpose
- **WHEN** a user reads the configuration documentation
- **THEN** they SHALL find a description explaining that execution_environment specifies the container image for ansible-navigator to use
- **AND** the documentation SHALL include example values such as `"quay.io/ansible/creator-ee:latest"`

#### Scenario: Documentation shows usage examples
- **WHEN** a user reads the examples documentation
- **THEN** they SHALL find at least one complete example demonstrating execution environment configuration
- **AND** the example SHALL show both HCL and JSON template syntax

### Requirement: Backward Compatibility
The addition of the execution_environment field SHALL maintain full backward compatibility with existing Packer templates.

#### Scenario: Existing templates without execution_environment field
- **WHEN** an existing Packer template that does not use the execution_environment field is run
- **THEN** the provisioner SHALL execute successfully without requiring the field
- **AND** the behavior SHALL be identical to previous plugin versions (using ansible-navigator's default)