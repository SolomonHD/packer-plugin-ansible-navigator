## ADDED Requirements

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

## MODIFIED Requirements

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
