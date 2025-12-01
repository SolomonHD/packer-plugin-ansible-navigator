# local-provisioner-capabilities Specification Delta

## Purpose

Defines the expected behavior and configuration options for the local `ansible-navigator` provisioner after alignment with the remote provisioner implementation.

## ADDED Requirements

### Requirement: Default Command

The local provisioner SHALL use `ansible-navigator run` as the default command instead of `ansible-playbook`.

#### Scenario: Default command when unspecified
- **GIVEN** a Packer configuration using `provisioner "ansible-navigator"`
- **WHEN** the `command` field is not specified
- **THEN** the provisioner SHALL execute using `ansible-navigator run`
- **AND** the provisioner SHALL support all standard ansible-navigator flags

#### Scenario: Override to legacy ansible-playbook
- **GIVEN** a Packer configuration using `provisioner "ansible-navigator"`
- **WHEN** the `command` field is set to `ansible-playbook`
- **THEN** the provisioner SHALL execute using the specified command
- **AND** backward compatibility with legacy configurations SHALL be maintained

#### Scenario: Missing ansible-navigator binary
- **GIVEN** ansible-navigator is not installed in PATH
- **WHEN** the provisioner attempts to run
- **THEN** it SHALL return an error: "ansible-navigator not found in PATH. Please install it before running this provisioner"

### Requirement: Play-Based Execution

The local provisioner SHALL support both traditional playbook files and modern collection-based plays through mutually exclusive configuration options.

#### Scenario: Playbook file execution
- **GIVEN** a configuration with `playbook_file = "site.yml"`
- **AND** `plays` is not specified
- **WHEN** the provisioner executes
- **THEN** it SHALL run the specified playbook file

#### Scenario: Collection plays execution
- **GIVEN** a configuration with `plays` array containing play definitions
- **AND** `playbook_file` is not specified
- **WHEN** the provisioner executes
- **THEN** it SHALL execute each play in sequence
- **AND** for role FQDNs, it SHALL generate temporary playbooks

#### Scenario: Both playbook_file and plays specified
- **GIVEN** a configuration with both `playbook_file` and `plays`
- **WHEN** the configuration is validated
- **THEN** it SHALL return an error: "you may specify only one of `playbook_file` or `plays`"

#### Scenario: Neither playbook_file nor plays specified
- **GIVEN** a configuration with neither `playbook_file` nor `plays`
- **WHEN** the configuration is validated
- **THEN** it SHALL return an error: "either `playbook_file` or `plays` must be defined"

### Requirement: Navigator Mode

The local provisioner SHALL support configuring the ansible-navigator execution mode.

#### Scenario: Default mode is stdout
- **GIVEN** a configuration without `navigator_mode` specified
- **WHEN** the provisioner executes
- **THEN** it SHALL use mode `stdout` for Packer-safe non-interactive output

#### Scenario: JSON mode with structured logging
- **GIVEN** a configuration with `navigator_mode = "json"`
- **AND** `structured_logging = true`
- **WHEN** the provisioner executes
- **THEN** it SHALL parse JSON events from ansible-navigator output
- **AND** it SHALL provide detailed task-level reporting

#### Scenario: Invalid mode
- **GIVEN** a configuration with `navigator_mode = "invalid_mode"`
- **WHEN** the configuration is validated
- **THEN** it SHALL return an error indicating valid modes are: stdout, json, yaml, interactive

### Requirement: Execution Environment Support

The local provisioner SHALL support specifying containerized execution environments for ansible-navigator.

#### Scenario: Custom execution environment
- **GIVEN** a configuration with `execution_environment = "quay.io/ansible/creator-ee:latest"`
- **WHEN** the provisioner executes
- **THEN** it SHALL pass `--execution-environment` flag to ansible-navigator

#### Scenario: Default execution environment
- **GIVEN** a configuration without `execution_environment` specified
- **WHEN** the provisioner executes
- **THEN** it SHALL use ansible-navigator's default execution environment behavior

### Requirement: Collections Management

The local provisioner SHALL support automatic installation of Ansible collections.

#### Scenario: Inline collections specification
- **GIVEN** a configuration with `collections = ["community.general:>=5.0.0", "ansible.posix:1.5.4"]`
- **WHEN** the provisioner executes
- **THEN** it SHALL install the specified collections before running playbooks

#### Scenario: Requirements file for collections
- **GIVEN** a configuration with `requirements_file = "requirements.yml"`
- **WHEN** the provisioner executes
- **THEN** it SHALL install collections and roles from the requirements file

#### Scenario: Offline mode
- **GIVEN** a configuration with `collections_offline = true`
- **AND** some required collections are not cached
- **WHEN** the provisioner executes
- **THEN** it SHALL fail with a clear error about missing cached collections

### Requirement: Error Handling and Keep Going

The local provisioner SHALL support configurable error handling for multi-play execution.

#### Scenario: Stop on first failure (default)
- **GIVEN** a configuration with multiple plays
- **AND** `keep_going` is not set or is `false`
- **WHEN** a play fails
- **THEN** execution SHALL stop immediately
- **AND** an error SHALL be returned: "Play 'X' failed with exit code 2"

#### Scenario: Continue on failure
- **GIVEN** a configuration with `keep_going = true`
- **WHEN** a play fails
- **THEN** the failure SHALL be logged
- **AND** execution SHALL continue to the next play
- **AND** a message SHALL indicate: "Continuing to next play despite failure (keep_going=true)"

### Requirement: Groups Configuration

The local provisioner SHALL support assigning the target host to Ansible inventory groups.

#### Scenario: Host assigned to groups
- **GIVEN** a configuration with `groups = ["webservers", "production"]`
- **WHEN** the inventory file is generated
- **THEN** the target host SHALL be listed under each specified group

### Requirement: Structured Logging

The local provisioner SHALL support structured JSON output and logging when configured.

#### Scenario: Log output to file
- **GIVEN** a configuration with `structured_logging = true`
- **AND** `log_output_path = "./logs/ansible.json"`
- **WHEN** the provisioner completes
- **THEN** a JSON summary file SHALL be written to the specified path

#### Scenario: Verbose task output
- **GIVEN** a configuration with `structured_logging = true`
- **AND** `verbose_task_output = true`
- **WHEN** a task executes
- **THEN** detailed task output SHALL be included in logs

### Requirement: Configuration Validation

The local provisioner Config.Validate() method SHALL validate all configuration options.

#### Scenario: Comprehensive validation
- **WHEN** Config.Validate() is called
- **THEN** it SHALL validate:
  - Navigator mode is valid (stdout, json, yaml, interactive)
  - Playbook file or plays (not both, at least one required)
  - Playbook files exist on disk
  - Referenced vars_files exist
  - Local port is valid (0-65535)

### Requirement: HCL2 Spec Generation

The local provisioner SHALL have properly generated HCL2 spec files for all configuration options.

#### Scenario: Generate HCL2 spec
- **WHEN** `make generate` is run
- **THEN** `provisioner.hcl2spec.go` SHALL be updated
- **AND** all new Config fields SHALL be represented in the spec

## Related Specifications

- `plugin-registration` - Local provisioner registration as DEFAULT_NAME
- `build-tooling` - Go version and build requirements