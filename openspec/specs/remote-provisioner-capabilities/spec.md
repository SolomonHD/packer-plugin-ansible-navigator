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

The SSH-based provisioner SHALL use ansible-navigator as its default executable, pass `run` as the first argument, and treat the `command` field strictly as the ansible-navigator executable name or path (without embedded arguments).

#### Scenario: Default command when unspecified

- **GIVEN** a Packer configuration using `provisioner "ansible-navigator"`
- **AND** the `command` field is not specified
- **WHEN** the provisioner constructs the command to run ansible-navigator
- **THEN** it SHALL invoke ansible-navigator using `exec.Command("ansible-navigator", "run", ...)` on the local machine
- **AND** any additional arguments (e.g., mode, execution environment, extra arguments) SHALL be passed as subsequent arguments, not embedded in the executable name

#### Scenario: Command treated as executable only

- **GIVEN** a configuration using `provisioner "ansible-navigator"`
- AND** the `command` field is set to a value such as `"my-ansible-navigator"` or `"/opt/tools/ansible-navigator-custom"`
- **WHEN** the provisioner constructs the command to run ansible-navigator
- **THEN** it SHALL treat the `command` value as the executable name or path only
- **AND** it SHALL still pass `run` as the first argument followed by all provisioner-derived arguments

#### Scenario: Embedded arguments in command rejected

- **GIVEN** a configuration using `provisioner "ansible-navigator"`
- **AND** the `command` field contains whitespace or embedded arguments (for example, `"ansible-navigator run"`, `"ansible-navigator --mode json"`, or `"ansible-navigator   run"`)
- **WHEN** the configuration is validated
- **THEN** validation SHALL fail
- **AND** the error message SHALL explain that `command` must be only the executable name or path (no extra arguments)
- **AND** the error message SHALL direct the user to use supported fields such as `extra_arguments` or play-level options for additional flags

#### Scenario: Missing ansible-navigator binary

- **GIVEN** ansible-navigator is not installed on the local machine
- **AND** no valid executable is available at any HOME-expanded `ansible_navigator_path` entry or in PATH
- **WHEN** the provisioner attempts to perform a version check or execute ansible-navigator
- **THEN** it SHALL return a clear error indicating that ansible-navigator is required
- **AND** the error message SHALL mention both PATH and `ansible_navigator_path` as resolution mechanisms

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

### Requirement: Play block naming and validation (remote)

The SSH-based provisioner SHALL expose multi-play configuration via a singular repeatable `play` block and SHALL reject legacy plural or array-based forms that conflict with this naming.

#### Scenario: Singular play block naming

- **GIVEN** a Packer configuration using `provisioner "ansible-navigator"`
- **AND** one or more `play { ... }` blocks are defined
- **WHEN** the configuration is parsed
- **THEN** each `play { ... }` block SHALL be accepted as a repeatable block
- **AND** the resulting configuration SHALL represent the plays as a collection in declaration order

#### Scenario: Legacy plays block rejected

- **GIVEN** a configuration that defines one or more `plays { ... }` blocks
- **WHEN** the configuration is validated
- **THEN** validation SHALL fail
- **AND** the error message SHALL explain that `plays` blocks are no longer supported
- **AND** the error message SHALL direct the user to use repeated `play { ... }` blocks instead

#### Scenario: Array syntax rejected in favor of blocks

- **GIVEN** a configuration that attempts `plays = [...]` or `play = [...]` array-style syntax
- **WHEN** the configuration is validated
- **THEN** validation SHALL fail
- **AND** the error message SHALL state that HCL2 block syntax is required
- **AND** the error message SHALL include an example using `play { ... }` blocks

### Requirement: Remote HOME Expansion for Path-Like Configuration

The SSH-based ansible-navigator provisioner SHALL expand HOME-relative (`~`) paths for a defined set of path-like configuration fields on the local side before validation and execution.

#### Scenario: Expand bare tilde to HOME

- **GIVEN** a configuration for `provisioner "ansible-navigator"`
- **AND** one or more path-like fields (e.g., `playbook_file`, `requirements_file`, `inventory_file`, `work_dir`, Galaxy-related paths) are set to `"~"`
- **WHEN** the configuration is prepared or validated
- **THEN** each `"~"` value SHALL be expanded to the user's HOME directory as reported by the local environment
- **AND** subsequent validation and execution SHALL use the expanded absolute path

#### Scenario: Expand tilde prefix with subdirectory

- **GIVEN** a configuration for `provisioner "ansible-navigator"`
- **AND** one or more path-like fields are set to values like `"~/ansible/site.yml"` or `"~/inventory/hosts"`
- **WHEN** the configuration is prepared or validated
- **THEN** each value SHALL be expanded by replacing the leading `"~/"` with the user's HOME directory plus a path separator
- **AND** file system checks (e.g., `os.Stat`) SHALL operate on the expanded path

#### Scenario: Preserve tilde with explicit username

- **GIVEN** a configuration for `provisioner "ansible-navigator"`
- **AND** a path-like field is set to a value beginning with `"~user/"` for some username other than the current user
- **WHEN** the configuration is prepared or validated
- **THEN** the value SHALL be left unchanged (no multi-user home resolution SHALL be attempted)
- **AND** any resulting validation error SHALL report the path exactly as configured

#### Scenario: Validation operates on expanded paths

- **GIVEN** a configuration that uses `~` or `~/subdir` values in path-like fields
- **WHEN** existing validation helpers check for file or directory existence
- **THEN** they SHALL operate on the HOME-expanded paths
- **AND** any error messages SHALL reference the expanded path that failed validation

### Requirement: Remote PATH Control for ansible-navigator

The SSH-based ansible-navigator provisioner SHALL support an `ansible_navigator_path` configuration option that prepends additional directories to PATH when resolving and running ansible-navigator for version checks and play execution.

#### Scenario: PATH prepended for version check

- **GIVEN** a configuration for `provisioner "ansible-navigator"` with:
  - `ansible_navigator_path = ["~/bin", "/opt/ansible/bin"]`
- **AND** ansible-navigator is installed only under one of these directories
- **WHEN** the provisioner performs a version check using `ansible-navigator --version`
- **THEN** it SHALL construct an environment for the `exec.Command` in which:
  - Each `ansible_navigator_path` entry is HOME-expanded
  - PATH begins with the expanded `ansible_navigator_path` entries joined by the OS path-list separator
  - The original PATH from the process environment is appended after these entries
- **AND** the version check SHALL succeed by locating ansible-navigator in one of the configured directories

#### Scenario: PATH prepended for play execution

- **GIVEN** a configuration for `provisioner "ansible-navigator"` with:
  - `ansible_navigator_path` set to one or more directories
- **AND** ansible-navigator is installed only under one of those directories
- **WHEN** the provisioner executes ansible-navigator to run a playbook or plays
- **THEN** it SHALL construct the `exec.Command` environment using the same PATH-prepending behavior as the version check
- **AND** ansible-navigator SHALL be resolved from one of the configured directories

#### Scenario: No change when ansible_navigator_path is unset

- **GIVEN** a configuration for `provisioner "ansible-navigator"` that does not set `ansible_navigator_path`
- **WHEN** the provisioner performs version checks or executes ansible-navigator
- **THEN** it SHALL use the process PATH unchanged
- **AND** existing behavior for locating ansible-navigator SHALL be preserved

