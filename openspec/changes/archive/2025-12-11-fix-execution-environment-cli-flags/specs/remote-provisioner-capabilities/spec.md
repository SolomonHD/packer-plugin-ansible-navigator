## MODIFIED Requirements

### Requirement: SSH-Based Remote Execution

The default `ansible-navigator` provisioner SHALL run ansible-navigator from the local machine (where Packer is executed) and connect to the target via SSH.

#### Scenario: Default execution mode

- **GIVEN** a Packer configuration using `provisioner "ansible-navigator"`
- **WHEN** the provisioner executes
- **THEN** it SHALL run on the local machine (not the target)
- **AND** it SHALL connect to the target via SSH using the configured communicator
- **AND** this matches the behavior of the official `ansible` provisioner

#### Scenario: Execution environment support

- **GIVEN** a configuration with `execution_environment = "quay.io/ansible/creator-ee:latest"`
- **WHEN** the provisioner executes
- **THEN** it SHALL pass `--ee true --eei quay.io/ansible/creator-ee:latest` flags to ansible-navigator
- **AND** the `--ee` flag SHALL be a boolean (`true`) to enable execution environment mode
- **AND** the `--eei` flag SHALL specify the container image
- **AND** the container SHALL have network access to reach the target via SSH

#### Scenario: No execution environment specified

- **GIVEN** a configuration without `execution_environment` specified
- **WHEN** the provisioner executes
- **THEN** it SHALL NOT pass `--ee` or `--eei` flags
- **AND** ansible-navigator SHALL use its default execution environment behavior

#### Scenario: Execution environment with custom registry

- **GIVEN** a configuration with `execution_environment = "myregistry.io/custom-ansible-ee:v1.0"`
- **WHEN** the provisioner constructs the command
- **THEN** it SHALL pass `--ee true --eei myregistry.io/custom-ansible-ee:v1.0` flags
- **AND** the full image reference SHALL be preserved including registry, repository, and tag
