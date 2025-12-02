# remote-provisioner-capabilities Specification

## Purpose
TBD - created by archiving change swap-provisioner-naming. Update Purpose after archive.
## Requirements
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
- **THEN** it SHALL use the specified container image on the local machine
- **AND** the container SHALL have network access to reach the target via SSH

### Requirement: Play-Based Execution
The SSH-based provisioner SHALL support both traditional playbook files and modern collection-based plays through mutually exclusive configuration options.

#### Scenario: Playbook file execution
- **GIVEN** a configuration with `playbook_file = "site.yml"`
- **AND** no `play` blocks are specified
- **WHEN** the provisioner executes
- **THEN** it SHALL run the specified playbook against the target via SSH

#### Scenario: Collection plays execution
- **GIVEN** a configuration with one or more `play` blocks
- **AND** `playbook_file` is not specified
- **WHEN** the provisioner executes
- **THEN** it SHALL execute each play in sequence against the target

#### Scenario: Mutual exclusivity validation
- **GIVEN** a configuration with both `playbook_file` and `play` blocks
- **WHEN** the configuration is validated
- **THEN** it SHALL return an error: "you may specify only one of `playbook_file` or `play`"

### Requirement: Default Navigator Command
The SSH-based provisioner SHALL use `ansible-navigator run` as the default command.

#### Scenario: Default command when unspecified
- **GIVEN** a Packer configuration using `provisioner "ansible-navigator"`
- **WHEN** the `command` field is not specified
- **THEN** the provisioner SHALL execute using `ansible-navigator run`

#### Scenario: Missing ansible-navigator binary
- **GIVEN** ansible-navigator is not installed on the local machine
- **WHEN** the provisioner attempts to run
- **THEN** it SHALL return a clear error indicating ansible-navigator is required

### Requirement: Navigator Mode Support
The SSH-based provisioner SHALL support configuring the ansible-navigator execution mode.

#### Scenario: Default mode is stdout
- **GIVEN** a configuration without `navigator_mode` specified
- **WHEN** the provisioner executes
- **THEN** it SHALL use mode `stdout` for Packer-safe non-interactive output

#### Scenario: JSON mode with structured logging
- **GIVEN** a configuration with `navigator_mode = "json"` and `structured_logging = true`
- **WHEN** the provisioner executes
- **THEN** it SHALL parse JSON events and provide detailed task-level reporting

### Requirement: Collections Management
The SSH-based provisioner SHALL support automatic installation of Ansible collections on the local machine.

#### Scenario: Inline collections specification
- **GIVEN** a configuration with `collections = ["community.general:>=5.0.0"]`
- **WHEN** the provisioner executes
- **THEN** it SHALL install the specified collections on the local machine before running

#### Scenario: Requirements file for collections
- **GIVEN** a configuration with `requirements_file = "requirements.yml"`
- **WHEN** the provisioner executes
- **THEN** it SHALL install collections and roles from the requirements file

### Requirement: Inventory Generation
The SSH-based provisioner SHALL generate dynamic inventory for the target host.

#### Scenario: Packer communicator as inventory source
- **WHEN** the provisioner executes
- **THEN** it SHALL generate an inventory with the target host
- **AND** the host SHALL be configured with connection details from the Packer communicator

#### Scenario: Host groups assignment
- **GIVEN** a configuration with `groups = ["webservers", "production"]`
- **WHEN** the inventory is generated
- **THEN** the target host SHALL be assigned to all specified groups

