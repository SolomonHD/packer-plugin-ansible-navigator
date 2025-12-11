## MODIFIED Requirements

### Requirement: Execution Environment Support

The local provisioner SHALL support specifying containerized execution environments for ansible-navigator using the modern ansible-navigator v3+ CLI syntax.

#### Scenario: Custom execution environment

- **GIVEN** a configuration with `execution_environment = "quay.io/ansible/creator-ee:latest"`
- **WHEN** the provisioner constructs the remote shell command
- **THEN** it SHALL pass `--ee true --eei quay.io/ansible/creator-ee:latest` flags to ansible-navigator
- **AND** the `--ee` flag SHALL be a boolean (`true`) to enable execution environment mode
- **AND** the `--eei` flag SHALL specify the container image

#### Scenario: Default execution environment

- **GIVEN** a configuration without `execution_environment` specified
- **WHEN** the provisioner executes
- **THEN** it SHALL NOT pass `--ee` or `--eei` flags
- **AND** ansible-navigator SHALL use its default execution environment behavior

#### Scenario: Execution environment with custom registry

- **GIVEN** a configuration with `execution_environment = "myregistry.io/custom-ansible-ee:v1.0"`
- **WHEN** the provisioner constructs the command
- **THEN** it SHALL pass `--ee true --eei myregistry.io/custom-ansible-ee:v1.0` flags
- **AND** the full image reference SHALL be preserved including registry, repository, and tag
