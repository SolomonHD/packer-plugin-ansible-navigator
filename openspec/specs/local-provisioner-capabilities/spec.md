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

The local provisioner SHALL execute provisioning via one or more ordered `play { ... }` blocks. Each play SHALL specify a `target` and MAY specify play-level settings (e.g., tags, become, vars_files). The provisioner SHALL process plays in declaration order.

#### Scenario: At least one play is required

- **GIVEN** a configuration using `provisioner "ansible-navigator-local"`
- **AND** no `play { ... }` blocks are defined
- **WHEN** the configuration is validated
- **THEN** validation SHALL fail
- **AND** the error message SHALL state that at least one `play` block must be defined

#### Scenario: Playbook target execution

- **GIVEN** a configuration with a `play` block whose `target` ends in `.yml` or `.yaml`
- **WHEN** the provisioner executes
- **THEN** it SHALL run that playbook via `ansible-navigator run` on the target

#### Scenario: Role FQDN target execution

- **GIVEN** a configuration with a `play` block whose `target` does not end in `.yml` or `.yaml`
- **WHEN** the provisioner executes
- **THEN** it SHALL treat the target as a role FQDN
- **AND** it SHALL generate a temporary playbook and execute it via `ansible-navigator run`

#### Scenario: Ordered execution

- **GIVEN** a configuration with multiple `play { ... }` blocks
- **WHEN** the provisioner executes
- **THEN** it SHALL execute each play in declaration order

#### Scenario: Invalid play array assignment syntax

- **GIVEN** a configuration attempting array-assignment syntax for repeatable play configuration
- **WHEN** Packer parses the configuration
- **THEN** it SHALL return an error indicating block syntax is required
- **AND** the error message SHALL suggest using `play { }` block format

#### Scenario: Multiple plays using repeated blocks

- **GIVEN** a configuration with multiple `play` blocks
- **WHEN** the configuration is parsed
- **THEN** each `play { }` block SHALL be processed in declaration order
- **AND** each block SHALL support independent configuration (name, target, extra_vars, become, tags, vars_files, etc.)

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

### Requirement: Dependency Management (requirements_file)

The local provisioner SHALL support dependency installation via an optional `requirements_file` that can define both roles and collections.

#### Scenario: requirements_file installs roles and collections

- **GIVEN** a configuration with `requirements_file = "requirements.yml"`
- **WHEN** the provisioner executes
- **THEN** it SHALL install roles and collections from that file before executing any plays

#### Scenario: requirements_file omitted

- **GIVEN** a configuration with no `requirements_file`
- **WHEN** the provisioner executes
- **THEN** it SHALL proceed without performing dependency installation

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

The local provisioner Config.Validate() method SHALL validate all supported configuration options.

#### Scenario: Comprehensive validation

- **WHEN** Config.Validate() is called
- **THEN** it SHALL validate:
  - One or more `play` blocks are defined
  - Each play has a non-empty `target`
  - Any referenced `vars_files` exist on disk (local side)
  - `navigator_config`, if specified, is a non-empty map
  - Command does not contain embedded arguments
  - All legacy options are still validated (no breaking changes)

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
- **THEN** the block SHALL use singular naming (`play`)
- **AND** multiple items SHALL be expressed as repeated singular blocks

### Requirement: Local HOME Expansion for Path-Like Configuration

The on-target ansible-navigator provisioner SHALL expand HOME-relative (`~`) paths for a defined set of path-like configuration fields on the local side before validation and remote command construction.

#### Scenario: Expand bare tilde to HOME

- **GIVEN** a configuration for `provisioner "ansible-navigator-local"`
- **AND** one or more local-side path-like fields (e.g., `command`, `ansible_navigator_path` entries, `requirements_file`, `work_dir`, play `target` when it is a playbook path, and play `vars_files` entries) are set to `"~"`
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

### Requirement: Version Check Timeout Configuration (Future)

The on-target ansible-navigator provisioner SHALL include a `version_check_timeout` configuration field for consistency with the remote provisioner, reserved for future use when version checking is implemented.

#### Scenario: Configuration field present

- **GIVEN** a configuration for `provisioner "ansible-navigator-local"`
- **WHEN** the configuration includes `version_check_timeout = "60s"`
- **THEN** the field SHALL be accepted in the configuration
- **AND** it SHALL be parsed as a valid duration string
- **AND** HCL2 spec SHALL include the field definition

#### Scenario: Default value for consistency

- **GIVEN** a configuration that does not specify `version_check_timeout`
- **WHEN** the provisioner prepares for execution
- **THEN** it SHALL default to "60s" for consistency with the remote provisioner
- **AND** the value SHALL be stored in the Config struct

#### Scenario: No version check currently performed

- **GIVEN** any configuration for the local provisioner
- **WHEN** the provisioner executes
- **THEN** no version check SHALL currently be performed on the target machine
- **AND** the `version_check_timeout` field SHALL be available for future implementation
- **AND** the field SHALL not cause errors when specified

#### Scenario: Documentation notes future use

- **GIVEN** user documentation for the local provisioner
- **WHEN** describing the `version_check_timeout` field
- **THEN** it SHALL note that version checking is not currently implemented for the local provisioner
- **AND** it SHALL indicate the field is reserved for future use
- **AND** it SHALL maintain consistency with remote provisioner documentation

### Requirement: Ansible Configuration File Generation and Upload

The local provisioner SHALL support generating a temporary ansible.cfg file from a declarative configuration map, uploading it to the target machine, and configuring the remote shell command to use it via ANSIBLE_CONFIG environment variable.

#### Scenario: User-specified ansible.cfg sections

- **GIVEN** a configuration with `ansible_cfg` set to a map containing section names as keys and key-value pairs as content:

  ```hcl
  ansible_cfg = {
    defaults = {
      remote_tmp = "/tmp/.ansible/tmp"
      host_key_checking = "False"
    }
    ssh_connection = {
      pipelining = "True"
    }
  }
  ```

- **WHEN** the provisioner prepares for execution
- **THEN** it SHALL generate a temporary file named `/tmp/packer-ansible-cfg-<random>.ini` (or equivalent in system temp directory) on the LOCAL machine
- **AND** the file SHALL contain valid INI format with sections and key-value pairs
- **AND** the file SHALL be uploaded to the staging directory on the TARGET machine
- **AND** the local temporary file SHALL be added to the cleanup list

#### Scenario: Generated INI file format

- **GIVEN** a configuration with multiple sections in `ansible_cfg`
- **WHEN** the temporary ansible.cfg file is generated
- **THEN** it SHALL use standard INI format with `[section_name]` headers
- **AND** each key-value pair SHALL be written as `key = value`
- **AND** sections SHALL be separated by blank lines for readability
- **AND** all string values SHALL be written without quotes

#### Scenario: Upload ansible.cfg to target staging directory

- **GIVEN** a generated ansible.cfg file at local path `/tmp/packer-ansible-cfg-ABC123.ini`
- **WHEN** the provisioner uploads files to the target
- **THEN** it SHALL upload the ansible.cfg file to `<staging_directory>/ansible.cfg`
- **AND** the upload SHALL occur after staging directory creation
- **AND** the upload SHALL occur before ansible-navigator execution

#### Scenario: ANSIBLE_CONFIG set in remote shell command

- **GIVEN** an uploaded ansible.cfg file at `<staging_directory>/ansible.cfg`
- **WHEN** the provisioner constructs the remote shell command to run ansible-navigator
- **THEN** it SHALL prepend `ANSIBLE_CONFIG=<staging_directory>/ansible.cfg` to the environment variables in the remote command
- **AND** this SHALL occur for all ansible-navigator executions (version check if implemented, play execution)

#### Scenario: Cleanup after provisioning success

- **GIVEN** a generated local ansible.cfg file
- **WHEN** provisioning completes successfully
- **THEN** the local temporary ansible.cfg file SHALL be deleted
- **AND** if `clean_staging_directory` is true, the uploaded ansible.cfg SHALL be removed with the staging directory

#### Scenario: Cleanup after provisioning failure

- **GIVEN** a generated local ansible.cfg file
- **WHEN** a play or playbook execution fails
- **THEN** the local temporary ansible.cfg file SHALL still be deleted
- **AND** the cleanup SHALL occur before the error is returned to the caller

#### Scenario: Default configuration for execution environments

- **GIVEN** a configuration with `execution_environment` set to a container image
- **AND** `ansible_cfg` is NOT explicitly specified
- **WHEN** the provisioner prepares for execution
- **THEN** it SHALL automatically apply default ansible.cfg settings:

  ```hcl
  ansible_cfg = {
    defaults = {
      remote_tmp = "/tmp/.ansible/tmp"
      local_tmp  = "/tmp/.ansible-local"
    }
  }
  ```

- **AND** these defaults SHALL prevent "Permission denied: /.ansible" errors for non-root container users

#### Scenario: User-provided ansible.cfg overrides automatic defaults

- **GIVEN** a configuration with `execution_environment` set
- **AND** `ansible_cfg` is explicitly specified by the user
- **WHEN** the provisioner prepares for execution
- **THEN** it SHALL use the user-specified `ansible_cfg` settings
- **AND** it SHALL NOT apply automatic defaults
- **AND** the user's configuration SHALL take full precedence

#### Scenario: No ansible.cfg generation when not configured

- **GIVEN** a configuration without `execution_environment`
- **AND** `ansible_cfg` is not specified
- **WHEN** the provisioner prepares for execution
- **THEN** it SHALL NOT generate a temporary ansible.cfg file
- **AND** it SHALL NOT upload an ansible.cfg file to the target
- **AND** it SHALL NOT set the ANSIBLE_CONFIG environment variable in the remote command
- **AND** Ansible SHALL use its normal configuration search order on the target

#### Scenario: Ansible's normal config file precedence on target

- **GIVEN** the target machine has an existing `ansible.cfg` file in a directory where ansible-navigator will execute
- **AND** the provisioner has uploaded a generated ansible.cfg to the staging directory
- **AND** ANSIBLE_CONFIG environment variable points to the uploaded file
- **WHEN** ansible-navigator executes on the target
- **THEN** the ANSIBLE_CONFIG environment variable SHALL take precedence
- **AND** the uploaded ansible.cfg SHALL be used
- **AND** any existing ansible.cfg files in default search locations SHALL be ignored

#### Scenario: Empty or malformed ansible_cfg rejected

- **GIVEN** a configuration with `ansible_cfg` set to an empty map `{}`
- **OR** `ansible_cfg` set to a non-map type
- **WHEN** the configuration is validated
- **THEN** validation SHALL fail with a clear error message
- **AND** the error SHALL explain that `ansible_cfg` must be a map of section names to key-value pairs

#### Scenario: Invalid section or key names allowed

- **GIVEN** a configuration with `ansible_cfg` containing section or key names not recognized by Ansible
- **WHEN** the provisioner generates the ansible.cfg file
- **THEN** it SHALL write them to the file as provided
- **AND** it SHALL NOT validate against Ansible's known options
- **AND** Ansible itself SHALL handle unknown options per its normal behavior on the target

#### Scenario: Special characters in values preserved

- **GIVEN** a configuration with `ansible_cfg` containing values with special characters (spaces, quotes, etc.)
- **WHEN** the INI file is generated
- **THEN** values SHALL be written literally without additional quoting
- **AND** special characters SHALL be preserved as-is
- **AND** bash-style quoting or escaping SHALL NOT be applied

#### Scenario: Path resolution for ansible.cfg in remote command

- **GIVEN** an uploaded ansible.cfg at `<staging_directory>/ansible.cfg`
- **AND** `staging_directory` is `/tmp/packer-provisioner-ansible-local/<uuid>`
- **WHEN** the remote shell command is constructed
- **THEN** the ANSIBLE_CONFIG path SHALL use the full remote staging directory path
- **AND** the path SHALL be constructed using `filepath.ToSlash()` for consistency with other remote paths

### Requirement: Navigator Config File Generation

The local provisioner SHALL support generating ansible-navigator.yml configuration files from a declarative HCL map, uploading them to the target, and using them to control ansible-navigator behavior.

#### Scenario: User-specified navigator_config generates YAML file

- **GIVEN** a configuration with:

  ```hcl
  navigator_config = {
    mode = "stdout"
    execution-environment = {
      enabled = true
      image = "quay.io/ansible/creator-ee:latest"
      pull-policy = "missing"
    }
    ansible = {
      config = {
        defaults = {
          remote_tmp = "/tmp/.ansible/tmp"
        }
      }
    }
  }
  ```

- **WHEN** the provisioner prepares for execution
- **THEN** it SHALL generate a temporary file named `/tmp/packer-navigator-cfg-<uuid>.yml` (or equivalent in system temp directory) on the LOCAL machine
- **AND** the file SHALL contain valid YAML matching the ansible-navigator.yml schema
- **AND** the file SHALL be uploaded to the staging directory on the TARGET machine
- **AND** the local temporary file SHALL be added to the cleanup list

#### Scenario: ANSIBLE_NAVIGATOR_CONFIG set in remote shell command

- **GIVEN** an uploaded ansible-navigator.yml file at `<staging_directory>/ansible-navigator.yml`
- **WHEN** the provisioner constructs the remote shell command to run ansible-navigator
- **THEN** it SHALL prepend `ANSIBLE_NAVIGATOR_CONFIG=<staging_directory>/ansible-navigator.yml` to the environment variables
- **AND** this SHALL occur for all ansible-navigator executions on the target

#### Scenario: Cleanup after provisioning

- **GIVEN** a generated local ansible-navigator.yml file
- **WHEN** provisioning completes (success or failure)
- **THEN** the local temporary ansible-navigator.yml file SHALL be deleted
- **AND** if `clean_staging_directory` is true, the uploaded file SHALL be removed with the staging directory

#### Scenario: Automatic EE defaults when execution environment enabled

- **GIVEN** a configuration with:

  ```hcl
  navigator_config = {
    execution-environment = {
      enabled = true
      image = "quay.io/ansible/creator-ee:latest"
    }
  }
  ```

- **AND** the user has NOT specified `ansible.config.defaults.remote_tmp` or `ansible.config.defaults.local_tmp`
- **AND** the user has NOT specified `execution-environment.environment-variables`
- **WHEN** the provisioner generates the ansible-navigator.yml file
- **THEN** it SHALL automatically include:

  ```yaml
  ansible:
    config:
      defaults:
        remote_tmp: "/tmp/.ansible/tmp"
        local_tmp: "/tmp/.ansible-local"
  execution-environment:
    environment-variables:
      ANSIBLE_REMOTE_TMP: "/tmp/.ansible/tmp"
      ANSIBLE_LOCAL_TMP: "/tmp/.ansible-local"
  ```

- **AND** these defaults SHALL prevent "Permission denied: /.ansible/tmp" errors in EE containers

#### Scenario: User-provided values override automatic defaults

- **GIVEN** a configuration with `execution-environment.enabled = true`
- **AND** the user has explicitly specified temp directory settings in navigator_config
- **WHEN** the provisioner generates the ansible-navigator.yml file
- **THEN** it SHALL use the user-specified values
- **AND** it SHALL NOT apply automatic defaults
- **AND** the user's configuration SHALL take full precedence

#### Scenario: No config file generation when navigator_config not specified

- **GIVEN** a configuration without `navigator_config`
- **WHEN** the provisioner prepares for execution
- **THEN** it SHALL NOT generate an ansible-navigator.yml file
- **AND** it SHALL NOT set the ANSIBLE_NAVIGATOR_CONFIG environment variable
- **AND** ansible-navigator SHALL use its normal configuration search order on the target

#### Scenario: navigator_config takes precedence over legacy options

- **GIVEN** a configuration with both legacy options and navigator_config:

  ```hcl
  execution_environment = "legacy-image:latest"
  navigator_mode = "json"
  
  navigator_config = {
    execution-environment = {
      enabled = true
      image = "new-image:latest"
    }
    mode = "stdout"
  }
  ```

- **WHEN** ansible-navigator is executed
- **THEN** the settings from `navigator_config` SHALL be used
- **AND** the legacy options SHALL be ignored in favor of navigator_config
- **AND** the generated config file SHALL use "new-image:latest" and "stdout" mode

#### Scenario: Complex nested structure preserved

- **GIVEN** a configuration with deeply nested navigator_config:

  ```hcl
  navigator_config = {
    ansible = {
      config = {
        defaults = {
          remote_tmp = "/tmp/.ansible/tmp"
          host_key_checking = "False"
        }
        ssh_connection = {
          pipelining = "True"
          timeout = "30"
        }
      }
    }
    execution-environment = {
      enabled = true
      image = "custom:latest"
      environment-variables = {
        ANSIBLE_REMOTE_TMP = "/custom/tmp"
        CUSTOM_VAR = "value"
      }
    }
  }
  ```

- **WHEN** the YAML file is generated
- **THEN** the nested structure SHALL be preserved exactly
- **AND** all keys and values SHALL be written correctly

