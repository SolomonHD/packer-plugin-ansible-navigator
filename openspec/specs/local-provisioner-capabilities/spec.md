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
- **AND** the error message SHALL direct the user to use supported fields such as play-level options or `navigator_config` for additional flags

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
  - **AND** removed options SHALL NOT be validated (they will fail HCL parsing)

#### Scenario: Removed options cause validation errors

- **GIVEN** a configuration attempting to use removed options like `execution_environment = "image:tag"`
- **WHEN** Packer parses the configuration
- **THEN** it SHALL fail with an error indicating the option is not recognized
- **AND** error messages SHOULD guide users to use navigator_config instead
- **AND** error messages SHOULD reference MIGRATION.md for help

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

### Requirement: Navigator Config File Generation

The local provisioner SHALL support generating ansible-navigator.yml configuration files from a declarative HCL map, uploading them to the target, and using them to control ansible-navigator behavior. The generated YAML SHALL conform to the ansible-navigator v24+/v25+ schema, including proper nested structure for execution environment fields.

#### Scenario: User-specified navigator_config generates YAML file

- **GIVEN** a configuration with:

  ```hcl
  navigator_config {
    mode = "stdout"
    execution_environment {
      enabled     = true
      image       = "quay.io/ansible/creator-ee:latest"
      pull_policy = "missing"
    }
    ansible_config {
      defaults {
        remote_tmp = "/tmp/.ansible/tmp"
      }
    }
  }
  ```

- **WHEN** the provisioner prepares for execution
- **THEN** it SHALL generate a temporary file named `/tmp/packer-navigator-cfg-<uuid>.yml` (or equivalent in system temp directory) on the LOCAL machine
- **AND** the file SHALL contain valid YAML conforming to the ansible-navigator.yml v25+ schema
- **AND** the execution-environment pull policy SHALL be generated as nested structure: `pull: { policy: missing }`
- **AND** the file SHALL NOT use flat `pull-policy: missing` syntax (which is rejected by ansible-navigator v25+)
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
  navigator_config {
    execution_environment {
      enabled = true
      image   = "quay.io/ansible/creator-ee:latest"
    }
  }
  ```

- **AND** the user has NOT specified `ansible_config.defaults.remote_tmp`
- **AND** the user has NOT specified `execution_environment.environment_variables`
- **WHEN** the provisioner generates the ansible-navigator.yml file
- **THEN** it SHALL automatically include:

  ```yaml
  ansible:
    config:
      defaults:
        remote_tmp: "/tmp/.ansible/tmp"
  execution-environment:
    environment-variables:
      set:
        ANSIBLE_REMOTE_TMP: "/tmp/.ansible/tmp"
        ANSIBLE_LOCAL_TMP: "/tmp/.ansible-local"
  ```

- **AND** these defaults SHALL prevent "Permission denied: /.ansible/tmp" errors in EE containers

#### Scenario: User-provided values override automatic defaults

- **GIVEN** a configuration with `execution_environment.enabled = true`
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

#### Scenario: Execution environment pull policy generates correct nested structure

- **GIVEN** a configuration with `execution_environment.pull_policy = "missing"`
- **WHEN** the ansible-navigator.yml YAML is generated
- **THEN** the YAML SHALL contain:

  ```yaml
  execution-environment:
    pull:
      policy: missing
  ```

- **AND** it SHALL NOT contain the flat structure `pull-policy: missing`
- **AND** the generated YAML SHALL pass ansible-navigator's built-in schema validation
- **AND** ansible-navigator SHALL accept the config file without "Additional properties" errors

#### Scenario: Complex nested structure preserved

- **GIVEN** a configuration with deeply nested navigator_config:

  ```hcl
  navigator_config {
    ansible_config {
      defaults {
        remote_tmp         = "/tmp/.ansible/tmp"
        host_key_checking  = false
      }
      ssh_connection {
        pipelining  = true
        ssh_timeout = 30
      }
    }
    execution_environment {
      enabled     = true
      image       = "custom:latest"
      pull_policy = "always"
      environment_variables {
        set = {
          ANSIBLE_REMOTE_TMP = "/custom/tmp"
          CUSTOM_VAR         = "value"
        }
      }
    }
  }
  ```

- **WHEN** the YAML file is generated
- **THEN** the nested structure SHALL be preserved exactly
- **AND** all keys and values SHALL be written correctly
- **AND** execution-environment pull policy SHALL use nested `pull.policy` structure

### Requirement: Navigator Config Nested Structure Support

The local provisioner's `navigator_config` field SHALL use explicit Go struct types with proper HCL2 spec generation to support the ansible-navigator.yml configuration schema while ensuring RPC serializability.

#### Scenario: Navigator config uses typed structs

- **GIVEN** the local provisioner implementation
- **WHEN** examining the Config struct definition
- **THEN** the `NavigatorConfig` field SHALL be defined as `*NavigatorConfig` (pointer to struct type)
- **AND** it SHALL NOT use `map[string]interface{}`
- **AND** the `NavigatorConfig` type SHALL be defined as a Go struct with proper mapstructure tags

#### Scenario: Struct types support nested configuration

- **GIVEN** a configuration with `provisioner "ansible-navigator-local"`
- **AND** `navigator_config` block with nested execution environment settings:

  ```hcl
  navigator_config {
    mode = "stdout"
    
    execution_environment {
      enabled     = true
      image       = "quay.io/ansible/creator-ee:latest"
      pull_policy = "missing"
      
      environment_variables {
        pass = ["SSH_AUTH_SOCK"]
        set = {
          ANSIBLE_REMOTE_TMP = "/tmp/.ansible/tmp"
        }
      }
    }
  }
  ```

- **WHEN** Packer parses and validates the configuration
- **THEN** validation SHALL succeed
- **AND** all nested structures SHALL be properly parsed into their respective struct types
- **AND** field names SHALL use underscores (not hyphens) for HCL compatibility

#### Scenario: Environment variables block uses pass/set structure

- **GIVEN** a configuration using `environment_variables` within `execution_environment`:

  ```hcl
  execution_environment {
    environment_variables {
      pass = ["SSH_AUTH_SOCK", "HOME"]
      set = {
        ANSIBLE_REMOTE_TMP = "/tmp/.ansible/tmp"
        MY_CUSTOM_VAR = "value"
      }
    }
  }
  ```

- **WHEN** Packer parses the configuration
- **THEN** `pass` SHALL be parsed as a list of strings (environment variable names to pass through)
- **AND** `set` SHALL be parsed as a map of string key-value pairs to set
- **AND** the generated YAML SHALL produce valid ansible-navigator.yml structure:

  ```yaml
  execution-environment:
    environment-variables:
      pass:
        - SSH_AUTH_SOCK
        - HOME
      set:
        ANSIBLE_REMOTE_TMP: "/tmp/.ansible/tmp"
        MY_CUSTOM_VAR: "value"
  ```

#### Scenario: Ansible config block supports nested defaults and ssh_connection

- **GIVEN** a configuration using `ansible_config` within `navigator_config`:

  ```hcl
  navigator_config {
    ansible_config {
      config = "/path/to/ansible.cfg"
      
      defaults {
        remote_tmp       = "/tmp/.ansible/tmp"
        host_key_checking = false
      }
      
      ssh_connection {
        ssh_timeout = 30
        pipelining  = true
      }
    }
  }
  ```

- **WHEN** Packer parses the configuration
- **THEN** the `ansible_config` block SHALL be parsed with nested `defaults` and `ssh_connection` blocks
- **AND** the struct SHALL NOT use `mapstructure:",squash"` tags that lose nested structure
- **AND** all nested fields SHALL be accessible as proper HCL blocks

#### Scenario: HCL2 spec uses RPC-serializable types

- **GIVEN** the generated HCL2 spec for the local provisioner
- **WHEN** examining the spec for `navigator_config`
- **THEN** it SHALL use concrete cty types (e.g., `cty.Object`, `cty.String`, `cty.Bool`)
- **AND** it SHALL NOT use `cty.DynamicPseudoType`
- **AND** it SHALL NOT use `cty.Map(cty.String)` for nested structures

#### Scenario: Plugin initialization succeeds without RPC errors

- **GIVEN** a configuration using `navigator_config` with nested structures
- **WHEN** Packer initializes the plugin
- **THEN** initialization SHALL complete successfully
- **AND** no "unsupported cty.Type conversion" errors SHALL occur
- **AND** the HCL2 spec SHALL serialize correctly over gRPC

#### Scenario: Structs support all common ansible-navigator.yml fields

- **GIVEN** the NavigatorConfig and related struct definitions
- **WHEN** examining their fields
- **THEN** they SHALL support at minimum:
  - `mode` (string)
  - `execution_environment` block with `enabled`, `image`, `pull_policy`, `environment_variables`
  - `environment_variables` block with `pass` (list), `set` (map)
  - `ansible_config` block with `config`, `defaults`, `ssh_connection` fields
  - `logging` configuration options
  - `playbook_artifact` settings
  - `collection_doc_cache` settings

#### Scenario: YAML generation works with typed structs

- **GIVEN** a configuration with typed `navigator_config`
- **WHEN** the provisioner generates the ansible-navigator.yml file
- **THEN** the YAML generation SHALL work correctly with the struct-based config
- **AND** the generated YAML SHALL match the expected ansible-navigator.yml schema
- **AND** nested structures SHALL be preserved in the YAML output
- **AND** hyphens SHALL be used in YAML keys where required by ansible-navigator

#### Scenario: Validation works with typed config

- **GIVEN** a configuration with typed `navigator_config`
- **WHEN** Config.Validate() is called
- **THEN** it SHALL validate that `navigator_config`, if specified, has valid field values
- **AND** it SHALL provide clear error messages for invalid configurations
- **AND** it SHALL support validation of nested fields

#### Scenario: Block syntax required for navigator_config

- **GIVEN** a configuration attempting to use map assignment syntax for `navigator_config`
- **WHEN** Packer parses the configuration
- **THEN** it SHALL return an error indicating block syntax is required
- **AND** the error message SHALL suggest using `navigator_config { }` block format

#### Scenario: All struct types included in go:generate directive

- **GIVEN** the provisioner source code
- **WHEN** examining the `go:generate` directive
- **THEN** it SHALL include all navigator config struct types needed for HCL2 spec generation
- **AND** `make generate` SHALL successfully generate specs for all types
- **AND** the directive SHALL NOT include removed types like `AnsibleConfigInner`

### Requirement: Mode CLI Flag Support for Local Provisioner

The local provisioner SHALL pass the `--mode` CLI flag in the remote shell command when `navigator_config.mode` is configured, preventing ansible-navigator on the target from hanging in interactive mode.

#### Scenario: Mode flag added in remote shell command

- **GIVEN** a configuration with `provisioner "ansible-navigator-local"`
- **AND** `navigator_config.mode = "stdout"` is specified
- **WHEN** the provisioner constructs the remote shell command for ansible-navigator
- **THEN** it SHALL include `--mode stdout` flag in the command
- **AND** the flag SHALL be positioned after the `run` subcommand
- **AND** the flag SHALL appear before playbook/target arguments

#### Scenario: Mode flag not added when mode not configured

- **GIVEN** a configuration with `provisioner "ansible-navigator-local"`
- **AND** NO `navigator_config.mode` is specified
- **WHEN** the provisioner constructs the remote shell command
- **THEN** it SHALL NOT include a `--mode` flag
- **AND** ansible-navigator on the target SHALL use its default behavior

#### Scenario: Mode flag prevents hang on target

- **GIVEN** a configuration with `navigator_config.mode = "stdout"`
- **AND** ansible-navigator is run on the target in a non-interactive SSH session
- **WHEN** the provisioner executes
- **THEN** ansible-navigator SHALL execute in stdout mode on the target
- **AND** it SHALL NOT wait for terminal input
- **AND** execution SHALL complete without hanging

### Requirement: YAML Root Structure for Local Provisioner

The local provisioner SHALL generate ansible-navigator.yml configuration files with a top-level `ansible-navigator:` root key wrapping all settings to conform to ansible-navigator v25.x schema requirements, preventing validation errors.

#### Scenario: Generated YAML wraps all settings under ansible-navigator key

- **GIVEN** a configuration with `provisioner "ansible-navigator-local"`
- **AND** `navigator_config` block with any settings
- **WHEN** the provisioner generates the ansible-navigator.yml file
- **THEN** the YAML SHALL have `ansible-navigator:` root key
- **AND** all settings SHALL be nested under this key
- **AND** the file SHALL be uploaded to the target staging directory

#### Scenario: Consistent YAML structure with remote provisioner

- **GIVEN** identical `navigator_config` blocks in local and remote provisioner configs
- **WHEN** both provisioners generate YAML files
- **THEN** the YAML content SHALL be identical
- **AND** both SHALL use `ansible-navigator:` root key
- **AND** both SHALL produce valid ansible-navigator v25.x configuration

#### Scenario: ConvertToYAMLStructure implementation matches remote

- **GIVEN** the local provisioner's `convertToYAMLStructure()` function
- **WHEN** it converts NavigatorConfig to YAML-compatible structure
- **THEN** it SHALL create a top-level map with key `"ansible-navigator"`
- **AND** the value SHALL be a map containing all navigator settings
- **AND** the implementation SHALL be identical to the remote provisioner's function
- **AND** nested structures SHALL be preserved within this wrapped structure

#### Scenario: Schema validation passes for local provisioner YAML

- **GIVEN** a generated ansible-navigator.yml file with `ansible-navigator:` root key
- **AND** the file is uploaded to the target
- **WHEN** ansible-navigator on the target processes the configuration file
- **THEN** it SHALL pass schema validation
- **AND** it SHALL NOT report "Additional properties are not allowed" errors
- **AND** all configuration settings SHALL be recognized and applied

#### Scenario: Backward compatibility for local provisioner configs

- **GIVEN** an existing Packer configuration using `provisioner "ansible-navigator-local"` with `navigator_config`
- **AND** the configuration was written before this change
- **WHEN** the configuration is used with the updated plugin
- **THEN** it SHALL continue to work without modification
- **AND** the YAML SHALL be generated with proper root structure automatically

