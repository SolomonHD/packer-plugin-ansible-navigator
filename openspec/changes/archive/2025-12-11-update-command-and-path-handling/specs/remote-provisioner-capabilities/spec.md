## ADDED Requirements

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

## MODIFIED Requirements

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
