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

The local provisioner SHALL expose a consistent set of dependency-install configuration options:

- `requirements_file` (string, optional)
- `offline_mode` (bool, optional)
- `roles_path` (string, optional)
- `collections_path` (string, optional)
- `galaxy_force` (bool, optional)
- `galaxy_force_with_deps` (bool, optional)
- `galaxy_command` (string, optional; defaults to `ansible-galaxy`)
- `galaxy_args` (list(string), optional)

#### Scenario: requirements_file installs roles and collections

- **GIVEN** a configuration with `requirements_file = "requirements.yml"`
- **WHEN** the provisioner executes
- **THEN** it SHALL install roles and collections from that file before executing any plays

#### Scenario: requirements_file omitted

- **GIVEN** a configuration with no `requirements_file`
- **WHEN** the provisioner executes
- **THEN** it SHALL proceed without performing dependency installation

#### Scenario: roles_path exported via ANSIBLE_ROLES_PATH

- **GIVEN** a configuration with `roles_path` set
- **WHEN** the provisioner executes any ansible-galaxy operation and any ansible-navigator play execution on the target
- **THEN** it SHALL ensure `ANSIBLE_ROLES_PATH` is set to the provided `roles_path` value for those operations

#### Scenario: collections_path exported via ANSIBLE_COLLECTIONS_PATHS

- **GIVEN** a configuration with `collections_path` set
- **WHEN** the provisioner executes any ansible-galaxy operation and any ansible-navigator play execution on the target
- **THEN** it SHALL ensure `ANSIBLE_COLLECTIONS_PATHS` is set to the provided `collections_path` value for those operations

#### Scenario: Galaxy command override and extra args

- **GIVEN** a configuration with `requirements_file` set
- **AND** `galaxy_command` set to a custom value
- **AND** `galaxy_args` set to one or more arguments
- **WHEN** the provisioner installs roles and collections
- **THEN** it SHALL invoke Galaxy using the configured `galaxy_command`
- **AND** it SHALL append `galaxy_args` to the constructed Galaxy argument list
- **AND** this behavior SHALL be consistent for both roles install and collections install

#### Scenario: galaxy_force maps to --force

- **GIVEN** a configuration with `galaxy_force = true`
- **AND** `galaxy_force_with_deps` is unset or `false`
- **WHEN** the provisioner invokes ansible-galaxy
- **THEN** it SHALL include `--force`

#### Scenario: galaxy_force_with_deps maps to --force-with-deps and takes precedence

- **GIVEN** a configuration with `galaxy_force_with_deps = true`
- **WHEN** the provisioner invokes ansible-galaxy
- **THEN** it SHALL include `--force-with-deps`
- **AND** it SHALL NOT additionally include `--force`

#### Scenario: offline_mode maps to --offline

- **GIVEN** a configuration with `offline_mode = true`
- **WHEN** the provisioner invokes ansible-galaxy to install from `requirements_file`
- **THEN** it SHALL include `--offline`

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

- **GIVEN** a configuration attempting to use removed options like `execution_environment = "image:tag"` or `work_dir = "/tmp/ansible"`
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

#### Scenario: Ansible config schema compliance

- **GIVEN** a configuration with `navigator_config.ansible_config` set (any combination of supported fields)
- **WHEN** the provisioner generates the ansible-navigator.yml file
- **THEN** the generated YAML MUST NOT contain `defaults` (or any other unexpected keys) under `ansible.config`
- **AND** `ansible.config` MUST contain only the allowed properties: `help`, `path`, and/or `cmdline`

#### Scenario: Mutual exclusivity for ansible_config.config vs nested blocks

- **GIVEN** a configuration with:

  ```hcl
  navigator_config {
    ansible_config {
      config = "/etc/ansible/ansible.cfg"
      defaults {
        remote_tmp = "/tmp/.ansible/tmp"
      }
    }
  }
  ```

- **WHEN** the configuration is validated
- **THEN** validation SHALL fail
- **AND** the error message SHALL state that `ansible_config.config` is mutually exclusive with `ansible_config.defaults` and `ansible_config.ssh_connection`

#### Scenario: Automatic EE defaults are injected per-key when execution environment enabled

- **GIVEN** a configuration with:

  ```hcl
  navigator_config {
    execution_environment {
      enabled = true
      image   = "quay.io/ansible/creator-ee:latest"
      environment_variables {
        set = {
          CUSTOM_VAR = "custom"
        }
      }
    }
  }
  ```

- **WHEN** the provisioner generates the ansible-navigator.yml file
- **THEN** it SHALL ensure execution-environment environment-variables include safe defaults **per missing key**, at minimum:

  ```yaml
  execution-environment:
    environment-variables:
      set:
        ANSIBLE_REMOTE_TMP: "/tmp/.ansible/tmp"
        ANSIBLE_LOCAL_TMP: "/tmp/.ansible-local"
  ```

- **AND** it SHALL NOT remove or overwrite unrelated user-provided keys (e.g., `CUSTOM_VAR`)

#### Scenario: Automatic EE HOME/XDG defaults when not explicitly set or passed through

- **GIVEN** a configuration with `navigator_config.execution_environment.enabled = true`
- **AND** the user has NOT provided values for any of the following via `execution_environment.environment_variables.set`:
  - `HOME`
  - `XDG_CACHE_HOME`
  - `XDG_CONFIG_HOME`
- **AND** the user has NOT requested pass-through of any of the following via `execution_environment.environment_variables.pass`:
  - `HOME`
  - `XDG_CACHE_HOME`
  - `XDG_CONFIG_HOME`
- **WHEN** the provisioner generates the ansible-navigator.yml file
- **THEN** it SHALL set the following defaults under `execution-environment.environment-variables.set`:
  - `HOME=/tmp`
  - `XDG_CACHE_HOME=/tmp/.cache`
  - `XDG_CONFIG_HOME=/tmp/.config`

#### Scenario: User-provided env var values are not overridden

- **GIVEN** a configuration with `navigator_config.execution_environment.enabled = true`
- **AND** the user has explicitly set one or more env var values under `execution_environment.environment_variables.set`
- **WHEN** the provisioner generates the ansible-navigator.yml file
- **THEN** the provisioner SHALL preserve the user's values
- **AND** it SHALL NOT overwrite user-set `HOME` or `XDG_*` values
- **AND** it SHALL NOT overwrite user-set `ANSIBLE_*` values

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

### Requirement: Per-Play extra_args escape hatch (local provisioner)

The on-target provisioner SHALL support a per-play `extra_args` field (list(string)) that is appended to the `ansible-navigator run` invocation for that play.

#### Scenario: play.extra_args is accepted in HCL schema

- **GIVEN** a configuration using `provisioner "ansible-navigator-local"`
- **AND** a `play {}` block includes `extra_args = ["--check", "--diff"]`
- **WHEN** Packer parses the configuration
- **THEN** parsing SHALL succeed
- **AND** the provisioner configuration SHALL include the `extra_args` values for that play

#### Scenario: play.extra_args is passed verbatim

- **GIVEN** a configuration using `provisioner "ansible-navigator-local"` with a play:

  ```hcl
  play {
    target     = "site.yml"
    extra_args = ["--check", "--diff"]
  }
  ```

- **WHEN** the provisioner constructs the remote shell command for that play
- **THEN** it SHALL include both `--check` and `--diff` as command arguments
- **AND** it SHALL not rewrite, split, or validate the `extra_args` values beyond basic type handling

#### Scenario: Deterministic argument ordering includes extra_args

- **GIVEN** a configuration using `provisioner "ansible-navigator-local"` with:
  - one `play {}` block
  - `navigator_config.mode = "stdout"`
  - `play.extra_args = ["--check", "--diff"]`
- **WHEN** the provisioner constructs the remote shell command arguments
- **THEN** the argument ordering SHALL be deterministic and consistent across executions:
  1. `run` subcommand
  2. enforced mode flag behavior (when configured), inserted immediately after `run`
  3. play-level `extra_args`
  4. provisioner-generated arguments (inventory, extra vars, tags, etc.)
  5. the play target (playbook path or generated playbook/role invocation)

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

### Requirement: Provide `skip_version_check` configuration field (parity)

The on-target provisioner SHALL include a `skip_version_check` configuration field for parity with the remote provisioner, even though local version checks are not currently performed.

#### Scenario: Configuration field present

Given: A configuration for `provisioner "ansible-navigator-local"` including `skip_version_check = true`
When: Packer parses the configuration
Then: Parsing succeeds and the field is accepted (non-fatal)

### Requirement: Warn when `version_check_timeout` is ineffective due to `skip_version_check`

When users explicitly configure `version_check_timeout` but also set `skip_version_check = true`, the plugin SHALL emit a user-visible warning indicating that the timeout is ignored.

#### Scenario: Warning when skip_version_check=true and version_check_timeout explicitly set

Given: A configuration for `provisioner "ansible-navigator-local"` with `skip_version_check = true` and an explicitly set `version_check_timeout`
When: The provisioner prepares for execution (configuration validation/prepare)
Then: The provisioner prints a non-fatal warning in Packer UI output stating that `version_check_timeout` is ignored when `skip_version_check=true`

#### Scenario: No warning when skip_version_check=false

Given: A configuration for `provisioner "ansible-navigator-local"` with `skip_version_check = false` and an explicitly set `version_check_timeout`
When: The provisioner prepares for execution (configuration validation/prepare)
Then: No warning about `version_check_timeout` being ignored is printed

#### Scenario: No warning when version_check_timeout not explicitly set

Given: A configuration for `provisioner "ansible-navigator-local"` with `skip_version_check = true` and without an explicitly set `version_check_timeout`
When: The provisioner prepares for execution (configuration validation/prepare)
Then: No warning about `version_check_timeout` being ignored is printed

### Requirement: `work_dir` is not supported (local provisioner)

The on-target `ansible-navigator-local` provisioner SHALL NOT support a `work_dir` configuration field.

#### Scenario: `work_dir` rejected at HCL parse time

Given: a configuration using `provisioner "ansible-navigator-local"` that includes `work_dir = "/tmp/ansible"`
When: Packer parses the configuration
Then: parsing SHALL fail with an error indicating `work_dir` is not a recognized argument

### Requirement: `ansible_config.defaults.local_tmp` support (local provisioner)

The on-target provisioner SHALL support configuring Ansible's local temp directory via `navigator_config.ansible_config.defaults.local_tmp`.

#### Scenario: HCL schema accepts local_tmp under ansible_config.defaults

- **GIVEN** a configuration using `provisioner "ansible-navigator-local"` with:

  ```hcl
  navigator_config {
    ansible_config {
      defaults {
        remote_tmp = "/tmp/.ansible/tmp"
        local_tmp  = "/tmp/.ansible-local"
      }
    }
  }
  ```

- **WHEN** Packer parses the configuration
- **THEN** parsing SHALL succeed
- **AND** the parsed configuration SHALL preserve the `local_tmp` value

#### Scenario: local_tmp is written to generated ansible.cfg and referenced from YAML

- **GIVEN** a configuration using `provisioner "ansible-navigator-local"` that sets `navigator_config.ansible_config.defaults.local_tmp`
- **WHEN** the provisioner prepares for execution and generates configuration artifacts
- **THEN** it SHALL generate an ansible.cfg that includes `local_tmp` under `[defaults]`
- **AND** the generated `ansible-navigator.yml` uploaded to the target SHALL reference the generated ansible.cfg via `ansible.config.path`
- **AND** the generated `ansible-navigator.yml` MUST NOT emit `defaults` under `ansible.config` (schema-compliance requirement)

#### Scenario: local_tmp omitted when unset

- **GIVEN** a configuration that does not set `navigator_config.ansible_config.defaults.local_tmp`
- **WHEN** the provisioner generates ansible.cfg
- **THEN** the generated ansible.cfg SHALL NOT include a `local_tmp` entry

### Requirement: Plugin debug mode enablement (local provisioner)

The on-target provisioner SHALL enable plugin debug output if and only if `navigator_config.logging.level` is set to `"debug"` (case-insensitive).

#### Scenario: Debug mode enabled only when navigator_config.logging.level is debug

- **GIVEN** a configuration for `provisioner "ansible-navigator-local"` includes `navigator_config`
- **AND** `navigator_config.logging.level` is set to `"debug"` in any letter casing
- **WHEN** the provisioner determines whether plugin debug mode is enabled
- **THEN** plugin debug mode SHALL be enabled

#### Scenario: Debug mode disabled when logging config is missing or level is not debug

- **GIVEN** a configuration for `provisioner "ansible-navigator-local"` does not set `navigator_config.logging.level` to `"debug"`
- **WHEN** the provisioner determines whether plugin debug mode is enabled
- **THEN** plugin debug mode SHALL be disabled

### Requirement: Plugin debug output format and sink (local provisioner)

When plugin debug mode is enabled, the on-target provisioner SHALL emit additional diagnostic messages via Packer UI using `ui.Message` and prefix each message with `[DEBUG]`.

#### Scenario: Debug messages use Packer UI and are prefixed

- **GIVEN** plugin debug mode is enabled
- **WHEN** the provisioner emits plugin diagnostic messages intended for debugging
- **THEN** the messages SHALL be emitted via Packer UI `ui.Message`
- **AND** each message SHALL be prefixed with `[DEBUG]`

#### Scenario: Debug messages are gated off when debug mode is disabled

- **GIVEN** plugin debug mode is disabled
- **WHEN** the provisioner executes a build
- **THEN** the additional `[DEBUG]` diagnostic messages SHALL NOT appear in the Packer UI output

### Requirement: Plugin debug output content (local provisioner)

When plugin debug mode is enabled, the on-target provisioner SHALL emit a small, deterministic set of additional diagnostic messages describing key execution decisions, without printing secrets.

#### Scenario: Debug includes command/path and config-file decisions

- **GIVEN** plugin debug mode is enabled
- **WHEN** the provisioner constructs the remote shell command and execution context for ansible-navigator
- **THEN** it SHALL emit debug messages that include:
  - the resolved ansible-navigator executable decision (final `command` and any PATH-prefixing intent)
  - whether `ANSIBLE_NAVIGATOR_CONFIG` is being set (and the path used)

#### Scenario: Debug includes “silent” play execution decisions

- **GIVEN** plugin debug mode is enabled
- **WHEN** the provisioner resolves each play target for execution
- **THEN** it SHALL emit debug messages that include:
  - whether a role target was converted into a generated temporary playbook
  - the absolute playbook path resolution result (when the play target is a playbook path)

#### Scenario: Debug output avoids printing secrets

- **GIVEN** plugin debug mode is enabled
- **WHEN** the provisioner emits debug messages containing user-provided values
- **THEN** it SHALL avoid printing secrets in debug output
- **AND** it SHALL follow the existing sanitization approach used for command/log output

### Requirement: REQ-EE-DEBUG-001 Debug-only EE container-runtime preflight diagnostics (local provisioner)

When plugin debug mode is enabled and `navigator_config.execution_environment.enabled = true`, the on-target provisioner SHALL emit DEBUG-only preflight diagnostics for container-runtime wiring on the target machine (where `ansible-navigator run` will execute).

#### Scenario: Preflight diagnostics are emitted when debug mode and EE are enabled
Given: a configuration for `provisioner "ansible-navigator-local"` where plugin debug mode is enabled
Given: `navigator_config.execution_environment.enabled = true`
When: the provisioner is about to invoke `ansible-navigator run` on the target machine
Then: the provisioner SHALL emit DEBUG-only preflight diagnostics via Packer UI messages
Then: the diagnostics SHALL reflect the target-side execution environment used for `ansible-navigator run`
Then: the diagnostics SHALL include whether `DOCKER_HOST` is unset or, if set, a safe representation of the value (e.g., redacting embedded credentials)
Then: the diagnostics SHALL include whether `/var/run/docker.sock` exists and whether it is a socket
Then: the diagnostics SHALL include whether the `docker` client is available in PATH
Then: the diagnostics SHALL NOT include unrelated environment variables

#### Scenario: Preflight diagnostics are NOT emitted when debug mode is disabled
Given: a configuration for `provisioner "ansible-navigator-local"` where plugin debug mode is disabled
When: the provisioner executes
Then: the new EE/Docker/DinD preflight diagnostics from this change SHALL NOT be emitted

#### Scenario: Preflight diagnostics are NOT emitted when EE is disabled
Given: a configuration for `provisioner "ansible-navigator-local"` where plugin debug mode is enabled
Given: `navigator_config.execution_environment.enabled` is unset or `false`
When: the provisioner executes
Then: the new EE/Docker/DinD preflight diagnostics from this change SHALL NOT be emitted

### Requirement: REQ-EE-DEBUG-002 Debug-only “likely DinD” warning heuristic (local provisioner)

When plugin debug mode is enabled and `navigator_config.execution_environment.enabled = true`, the on-target provisioner SHALL emit a warning-only debug advisory when a `dockerd` process is detected on the target, indicating a likely Docker-in-Docker setup.

#### Scenario: Dockerd presence emits a warning-only advisory
Given: a configuration for `provisioner "ansible-navigator-local"` where plugin debug mode is enabled
Given: `navigator_config.execution_environment.enabled = true`
Given: a `dockerd` process is detected in the same target-side execution environment
When: the provisioner runs the EE preflight checks
Then: the provisioner SHALL emit a warning-labeled debug message (e.g. prefixed with `[DEBUG][WARN]`)
Then: the warning SHALL be advisory only and SHALL NOT hard-fail the build
Then: the warning SHALL include an actionable next step (e.g., advise using a host/remote daemon rather than Docker-in-Docker)

### Requirement: REQ-EE-DEBUG-003 Preflight checks are safe and non-blocking (local provisioner)

The on-target provisioner's debug-only EE preflight checks SHALL be fast, non-blocking, and SHALL NOT change execution behavior beyond emitting Packer UI debug messages.

#### Scenario: Checks avoid slow/hanging docker operations
Given: a configuration for `provisioner "ansible-navigator-local"` where plugin debug mode is enabled
Given: `navigator_config.execution_environment.enabled = true`
When: the provisioner runs the EE preflight checks
Then: the checks SHALL be fast and non-blocking
Then: the checks SHALL NOT run potentially slow/hanging docker commands such as `docker info`
Then: the checks SHALL NOT change execution behavior beyond emitting debug UI messages

### Requirement: REQ-EXTRA-VARS-001 Provisioner-generated extra-vars are passed as a single JSON object (local)

The on-target provisioner SHALL pass provisioner-generated Ansible extra vars (including Packer-derived variables and automatically added variables like `ansible_ssh_private_key_file`) in a form that cannot produce malformed `-e` / `--extra-vars` usage and cannot shift positional arguments.

#### Scenario: JSON extra-vars cannot produce a standalone -e

- **GIVEN** the provisioner constructs the `ansible-navigator run` argument list
- **WHEN** the provisioner includes provisioner-generated extra vars
- **THEN** it SHALL encode those extra vars as a single JSON object
- **AND** it SHALL pass that JSON object via exactly one `-e`/`--extra-vars` argument pair
- **AND** the argument list SHALL NOT contain a standalone `-e`/`--extra-vars` flag without an argument

#### Scenario: Playbook path is always last and not displaced by extra-vars

- **GIVEN** a play whose target resolves to a playbook path
- **WHEN** the provisioner constructs the final `ansible-navigator run` command arguments
- **THEN** the playbook path argument SHALL be last
- **AND** no extra-vars value (including any JSON string) SHALL appear in the final command position

### Requirement: Show Extra Vars JSON in Output Log (local provisioner)

The on-target local provisioner SHALL support a `show_extra_vars` configuration option that, when enabled, logs the complete extra vars JSON object to Packer UI output before executing ansible-navigator on the target.

#### Scenario: show_extra_vars option accepted in HCL schema

- **GIVEN** a configuration using `provisioner "ansible-navigator-local"`
- **AND** the configuration includes `show_extra_vars = true`
- **WHEN** Packer parses the configuration
- **THEN** parsing SHALL succeed
- **AND** the provisioner configuration SHALL include the `show_extra_vars` value

#### Scenario: Extra vars JSON logged when show_extra_vars is true

- **GIVEN** a configuration using `provisioner "ansible-navigator-local"` with `show_extra_vars = true`
- **AND** the provisioner has constructed extra vars including Packer-injected variables (e.g., `packer_build_name`, `packer_builder_type`)
- **AND** the provisioner may also include user-defined play-level extra vars
- **WHEN** the provisioner prepares to execute ansible-navigator for a play
- **THEN** it SHALL emit the extra vars JSON object to Packer UI via `ui.Message()`
- **AND** the output SHALL be prefixed with a clear identifier (e.g., `[Extra Vars]`)
- **AND** the JSON SHALL be formatted with indentation for human readability

#### Scenario: Sensitive values are redacted in extra vars output

- **GIVEN** a configuration using `provisioner "ansible-navigator-local"` with `show_extra_vars = true`
- **AND** the extra vars include sensitive values (e.g., `ansible_password`)
- **WHEN** the provisioner logs the extra vars JSON
- **THEN** it SHALL redact the `ansible_password` value by replacing it with `*****`
- **AND** any other known sensitive keys SHALL be similarly redacted
- **AND** non-sensitive values SHALL be shown

#### Scenario: Extra vars not logged when show_extra_vars is false or unset

- **GIVEN** a configuration using `provisioner "ansible-navigator-local"`
- **AND** `show_extra_vars` is not set OR `show_extra_vars = false`
- **WHEN** the provisioner executes ansible-navigator
- **THEN** it SHALL NOT emit the extra vars JSON to Packer UI output
- **AND** existing behavior SHALL be preserved

#### Scenario: show_extra_vars defaults to false

- **GIVEN** a configuration using `provisioner "ansible-navigator-local"` without setting `show_extra_vars`
- **WHEN** the provisioner configuration is prepared
- **THEN** `show_extra_vars` SHALL default to `false`
- **AND** no extra vars JSON output SHALL be produced

