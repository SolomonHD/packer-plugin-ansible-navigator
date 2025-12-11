# local-provisioner-capabilities Specification

## Purpose

Defines the expected behavior and configuration options for the local `ansible-navigator` provisioner.
## Requirements

<!-- Requirements will be added through change proposals -->

### Requirement: Default Command

The on-target provisioner SHALL use ansible-navigator as its default executable, invoke `ansible-navigator run` via remote shell command, and treat the `command` field strictly as the ansible-navigator executable name or path (without embedded arguments).

#### Scenario: Default command when unspecified

- **GIVEN** a Packer configuration using `provisioner "ansible-navigator-local"`
- **AND** the `command` field is not specified
- **WHEN** the provisioner constructs the remote shell command to run ansible-navigator
- **THEN** it SHALL invoke `ansible-navigator run` on the target machine
- **AND** any additional arguments (e.g., mode, execution environment, extra arguments) SHALL be passed as separate arguments, not embedded in the executable name

#### Scenario: Command treated as executable only

- **GIVEN** a configuration using `provisioner "ansible-navigator-local"`
- **AND** the `command` field is set to a value such as `"my-ansible-navigator"` or `"/opt/tools/ansible-navigator-custom"`
- **WHEN** the provisioner constructs the remote shell command
- **THEN** it SHALL treat the `command` value as the executable name or path only
- **AND** it SHALL still pass `run` as the first argument followed by all provisioner-derived arguments

#### Scenario: Embedded arguments in command rejected

- **GIVEN** a configuration using `provisioner "ansible-navigator-local"`
- **AND** the `command` field contains whitespace or embedded arguments (for example, `"ansible-navigator run"`, `"ansible-navigator --mode json"`, or `"ansible-navigator   run"`)
- **WHEN** the configuration is validated
- **THEN** validation SHALL fail
- **AND** the error message SHALL explain that `command` must be only the executable name or path (no extra arguments)
- **AND** the error message SHALL direct the user to use supported fields such as `extra_arguments` or play-level options for additional flags

#### Scenario: Missing ansible-navigator binary on target

- **GIVEN** ansible-navigator is not installed on the target machine
- **AND** no valid executable is available at any configured `ansible_navigator_path` entry or in the target's default PATH
- **WHEN** the provisioner attempts to execute ansible-navigator via the remote shell command
- **THEN** it SHALL return a clear error indicating that ansible-navigator is required on the target
- **AND** the error message SHALL mention both PATH and `ansible_navigator_path` as resolution mechanisms

### Requirement: Play-Based Execution

The local provisioner SHALL support both traditional playbook files and modern collection-based plays through mutually exclusive configuration options. The play configuration MUST use HCL2 block syntax with the singular block name `play` (repeated `play { }` blocks), following standard HCL idioms for repeatable blocks. Legacy plural or array-based configuration forms for plays MUST NOT be accepted.

#### Scenario: Playbook file execution

- **GIVEN** a configuration with `playbook_file = "site.yml"`
- **AND** no `play` blocks are specified
- **WHEN** the provisioner executes
- **THEN** it SHALL run the specified playbook file

#### Scenario: Collection plays execution with block syntax

- **GIVEN** a configuration with one or more `play` blocks using HCL2 block syntax
- **AND** `playbook_file` is not specified
- **WHEN** the provisioner executes
- **THEN** it SHALL execute each play in sequence
- **AND** for role FQCNs, it SHALL generate temporary playbooks

#### Scenario: Invalid play array assignment syntax

- **GIVEN** a configuration attempting `play = [...]` or `plays = [...]` array assignment syntax
- **WHEN** Packer parses the configuration
- **THEN** it SHALL return an error indicating block syntax is required
- **AND** the error message SHALL suggest using `play { }` block format

#### Scenario: Multiple plays using repeated blocks

- **GIVEN** a configuration with multiple `play` blocks
- **WHEN** the configuration is parsed
- **THEN** each `play { }` block SHALL be processed in declaration order
- **AND** each block SHALL support independent configuration (name, target, extra_vars, become, become_user, tags, skip_tags, etc.)

#### Scenario: Both playbook_file and play specified

- **GIVEN** a configuration with both `playbook_file` and `play` blocks
- **WHEN** the configuration is validated
- **THEN** it SHALL return an error: "you may specify only one of `playbook_file` or `play`"

#### Scenario: Neither playbook_file nor play specified

- **GIVEN** a configuration with neither `playbook_file` nor `play` blocks
- **WHEN** the configuration is validated
- **THEN** it SHALL return an error: "either `playbook_file` or `play` must be defined"

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

### Requirement: HCL Block Naming Convention

The local provisioner SHALL follow HCL idioms for block naming, using singular names for blocks that can be repeated.

#### Scenario: Block name follows HCL conventions

- **GIVEN** the provisioner HCL2 spec definition
- **WHEN** defining blocks that represent individual items in a collection
- **THEN** the block SHALL use singular naming (`play` not `plays`)
- **AND** multiple items SHALL be expressed as repeated singular blocks

#### Scenario: Deprecated plural block name (migration)

- **GIVEN** a configuration using the deprecated `plays { }` block name
- **WHEN** Packer parses the configuration
- **THEN** it SHALL return an error indicating the block is not recognized
- **AND** the error message SHALL indicate that `play { }` is the correct syntax

### Requirement: Local HOME Expansion for Path-Like Configuration

The on-target ansible-navigator provisioner SHALL expand HOME-relative (`~`) paths for a defined set of path-like configuration fields on the local side before validation and remote command construction.

#### Scenario: Expand bare tilde to HOME

- **GIVEN** a configuration for `provisioner "ansible-navigator-local"`
- **AND** one or more local-side path-like fields (e.g., `playbook_file`, `playbook_files`, `playbook_paths`, `requirements_file`, `inventory_file`, `galaxy_file`, `role_paths`, `collection_paths`, `group_vars`, `host_vars`, `work_dir`, Galaxy cache directories) are set to `"~"`
- **WHEN** the configuration is prepared or validated
- **THEN** each `"~"` value SHALL be expanded to the user's HOME directory as reported by the local environment
- **AND** subsequent validation and remote command construction SHALL use the expanded absolute path

#### Scenario: Expand tilde prefix with subdirectory

- **GIVEN** a configuration for `provisioner "ansible-navigator-local"`
- **AND** one or more local-side path-like fields are set to values like `"~/ansible/site.yml"` or `"~/inventory/hosts"`
- **WHEN** the configuration is prepared or validated
- **THEN** each value SHALL be expanded by replacing the leading `"~/"` with the user's HOME directory plus a path separator
- **AND** file system checks (e.g., `os.Stat`) SHALL operate on the expanded path

#### Scenario: Preserve tilde with explicit username

- **GIVEN** a configuration for `provisioner "ansible-navigator-local"`
- **AND** a path-like field is set to a value beginning with `"~user/"` for some username other than the current user
- **WHEN** the configuration is prepared or validated
- **THEN** the value SHALL be left unchanged (no multi-user home resolution SHALL be attempted)
- **AND** any resulting validation error SHALL report the path exactly as configured

#### Scenario: Validation operates on expanded paths

- **GIVEN** a configuration that uses `~` or `~/subdir` values in path-like fields
- **WHEN** existing validation helpers check for file or directory existence
- **THEN** they SHALL operate on the HOME-expanded paths
- **AND** any error messages SHALL reference the expanded path that failed validation

### Requirement: Local PATH Control for ansible-navigator

The on-target ansible-navigator provisioner SHALL support an `ansible_navigator_path` configuration option that prepends additional directories to PATH in the remote shell command used to run ansible-navigator.

#### Scenario: PATH prepended in remote shell command

- **GIVEN** a configuration for `provisioner "ansible-navigator-local"` with:
  - `ansible_navigator_path = ["~/bin", "/opt/ansible/bin"]`
- **AND** ansible-navigator is installed on the target machine only under one of these directories
- **WHEN** the provisioner constructs the remote shell command to run ansible-navigator
- **THEN** it SHALL:
  - HOME-expand each `ansible_navigator_path` entry on the local side
  - Construct a PATH override prefix for the remote shell command in the form `PATH="expanded_entry1:expanded_entry2:$PATH"`
  - Place this PATH override at the beginning of the remote shell command invocation
- **AND** ansible-navigator SHALL be resolved from one of the configured directories on the target

#### Scenario: No change when ansible_navigator_path is unset

- **GIVEN** a configuration for `provisioner "ansible-navigator-local"` that does not set `ansible_navigator_path`
- **WHEN** the provisioner constructs the remote shell command to run ansible-navigator
- **THEN** it SHALL not modify the PATH in the remote shell command
- **AND** existing behavior for locating ansible-navigator on the target SHALL be preserved

