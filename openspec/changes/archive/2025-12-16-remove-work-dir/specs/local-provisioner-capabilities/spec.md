## ADDED Requirements

### Requirement: `work_dir` is not supported (local provisioner)

The on-target `ansible-navigator-local` provisioner SHALL NOT support a `work_dir` configuration field.

#### Scenario: `work_dir` rejected at HCL parse time

Given: a configuration using `provisioner "ansible-navigator-local"` that includes `work_dir = "/tmp/ansible"`
When: Packer parses the configuration
Then: parsing SHALL fail with an error indicating `work_dir` is not a recognized argument

## MODIFIED Requirements

### Requirement: Local HOME Expansion for Path-Like Configuration

The on-target ansible-navigator provisioner SHALL expand HOME-relative (`~`) paths for a defined set of path-like configuration fields on the local side before validation and remote command construction.

#### Scenario: Expand bare tilde to HOME

- **GIVEN** a configuration for `provisioner "ansible-navigator-local"`
- **AND** one or more local-side path-like fields (e.g., `command`, `ansible_navigator_path` entries, `requirements_file`, play `target` when it is a playbook path, and play `vars_files` entries) are set to `"~"`
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

### Requirement: Configuration Validation

The local provisioner Config.Validate() method SHALL validate all supported configuration options.

#### Scenario: Comprehensive validation

- **WHEN** Config.Validate() is called
- **THEN** it SHALL validate:
  - One or more `play` blocks are defined
  - Each play has a non-empty `target`
  - Any referenced `vars_files` exist on disk (local side)
  - `navigator_config`, if specified, is a non-empty map
  - Command does not contain embedded arguments
  - **AND** removed options SHALL NOT be validated (they will fail HCL parsing)

#### Scenario: Removed options cause validation errors

- **GIVEN** a configuration attempting to use removed options like `execution_environment = "image:tag"` or `work_dir = "/tmp/ansible"`
- **WHEN** Packer parses the configuration
- **THEN** it SHALL fail with an error indicating the option is not recognized
- **AND** error messages SHOULD guide users to use navigator_config instead
- **AND** error messages SHOULD reference MIGRATION.md for help

