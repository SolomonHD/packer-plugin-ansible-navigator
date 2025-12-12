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

### Requirement: Play-Based Execution

The SSH-based provisioner SHALL execute provisioning via one or more ordered `play { ... }` blocks. Each play SHALL specify a `target`.

#### Scenario: At least one play is required

- **GIVEN** a configuration using `provisioner "ansible-navigator"`
- **AND** no `play { ... }` blocks are defined
- **WHEN** the configuration is validated
- **THEN** validation SHALL fail
- **AND** the error message SHALL state that at least one `play` block must be defined

#### Scenario: Playbook target execution

- **GIVEN** a configuration with a `play` block whose `target` ends in `.yml` or `.yaml`
- **WHEN** the provisioner executes
- **THEN** it SHALL run that playbook against the target via SSH

#### Scenario: Role FQDN target execution

- **GIVEN** a configuration with a `play` block whose `target` does not end in `.yml` or `.yaml`
- **WHEN** the provisioner executes
- **THEN** it SHALL treat the target as a role FQDN
- **AND** it SHALL generate a temporary playbook and execute it via SSH

#### Scenario: Ordered execution

- **GIVEN** a configuration with multiple `play { ... }` blocks
- **WHEN** the provisioner executes
- **THEN** it SHALL execute each play in declaration order

### Requirement: Default Navigator Command

The SSH-based provisioner SHALL use ansible-navigator as its default executable, pass `run` as the first argument, and treat the `command` field strictly as the ansible-navigator executable name or path (without embedded arguments).

#### Scenario: Default command construction

- **GIVEN** a Packer configuration using `provisioner "ansible-navigator"`
- **AND** the `command` field is not specified
- **WHEN** the provisioner constructs the command to run ansible-navigator
- **THEN** it SHALL invoke ansible-navigator using `exec.Command("ansible-navigator", "run", ...)` on the local machine
- **AND** configuration SHALL be controlled ONLY via ANSIBLE_NAVIGATOR_CONFIG environment variable
- **AND** NO mode, EE, or other CLI flags SHALL be passed (all settings come from config file)
- **AND** Legacy CLI flags (`--mode`, `--ee`, `--eei`) SHALL NOT be generated

### Requirement: Dependency Management (requirements_file)

The SSH-based provisioner SHALL support dependency installation via an optional `requirements_file` that can define both roles and collections.

#### Scenario: requirements_file installs roles and collections

- **GIVEN** a configuration with `requirements_file = "requirements.yml"`
- **WHEN** the provisioner executes
- **THEN** it SHALL install roles and collections from that file before executing any plays

#### Scenario: requirements_file omitted

- **GIVEN** a configuration with no `requirements_file`
- **WHEN** the provisioner executes
- **THEN** it SHALL proceed without performing dependency installation

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

### Requirement: Remote HOME Expansion for Path-Like Configuration

The SSH-based ansible-navigator provisioner SHALL expand HOME-relative (`~`) paths for a defined set of path-like configuration fields on the local side before validation and execution.

#### Scenario: Expand bare tilde to HOME

- **GIVEN** a configuration for `provisioner "ansible-navigator"`
- **AND** one or more path-like fields (e.g., `requirements_file`, `inventory_file`, `work_dir`, Galaxy-related paths, play `target` when it is a playbook path, and play `vars_files` entries) are set to `"~"`
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
- **WHEN** the provisioner executes ansible-navigator to run one or more plays
- **THEN** it SHALL construct the `exec.Command` environment using the same PATH-prepending behavior as the version check
- **AND** ansible-navigator SHALL be resolved from one of the configured directories

#### Scenario: No change when ansible_navigator_path is unset

- **GIVEN** a configuration for `provisioner "ansible-navigator"` that does not set `ansible_navigator_path`
- **WHEN** the provisioner performs version checks or executes ansible-navigator
- **THEN** it SHALL use the process PATH unchanged
- **AND** existing behavior for locating ansible-navigator SHALL be preserved

### Requirement: Version Check Timeout Configuration

The SSH-based ansible-navigator provisioner SHALL support a configurable timeout for the version check operation to prevent indefinite hangs when ansible-navigator cannot be located or is slow to respond.

#### Scenario: Default timeout value

- **GIVEN** a configuration for `provisioner "ansible-navigator"` that does not specify `version_check_timeout`
- **WHEN** the provisioner prepares for execution
- **THEN** it SHALL use a default timeout of 60 seconds for the version check
- **AND** the version check SHALL fail with a timeout error if it exceeds 60 seconds

#### Scenario: Custom timeout value

- **GIVEN** a configuration with `version_check_timeout = "30s"`
- **WHEN** the version check is performed
- **THEN** it SHALL use a 30-second timeout
- **AND** the timeout SHALL be enforced using `context.WithTimeout` and `exec.CommandContext`

#### Scenario: Timeout uses duration format

- **GIVEN** a configuration with `version_check_timeout` set to a value
- **WHEN** the configuration is prepared
- **THEN** the value SHALL be parsed using `time.ParseDuration`
- **AND** valid duration formats SHALL include "60s", "2m", "1m30s", etc.
- **AND** invalid duration formats SHALL cause configuration validation to fail

#### Scenario: Timeout error message

- **GIVEN** the version check command times out after the configured duration
- **WHEN** the timeout occurs
- **THEN** the error message SHALL indicate:
  - The version check timed out
  - The configured timeout value (e.g., "60s")
  - Suggestions including: use `ansible_navigator_path`, use `command` to specify full path, set `skip_version_check = true`, or increase `version_check_timeout`

#### Scenario: Not-found error remains distinct

- **GIVEN** ansible-navigator cannot be found in PATH before the timeout expires
- **WHEN** the command execution fails with a not-found error
- **THEN** the error message SHALL indicate ansible-navigator was not found
- **AND** the error SHALL not be reported as a timeout
- **AND** the error message SHALL mention both PATH and `ansible_navigator_path` as resolution mechanisms

#### Scenario: Skip version check bypasses timeout

- **GIVEN** a configuration with `skip_version_check = true`
- **WHEN** the provisioner prepares for execution
- **THEN** no version check SHALL be performed
- **AND** the `version_check_timeout` configuration SHALL be ignored

#### Scenario: Version check respects modified PATH

- **GIVEN** a configuration with both `ansible_navigator_path` and `version_check_timeout`
- **WHEN** the version check executes
- **THEN** it SHALL use the modified PATH from `ansible_navigator_path`
- **AND** it SHALL enforce the configured timeout
- **AND** both mechanisms SHALL work together correctly

#### Scenario: asdf installation compatibility

- **GIVEN** ansible-navigator is installed via asdf
- **AND** the configuration uses `command = "~/.asdf/shims/ansible-navigator"` or `ansible_navigator_path = ["~/.asdf/shims"]`
- **WHEN** the version check executes
- **THEN** it SHALL succeed within the timeout period if the shim responds correctly
- **OR** it SHALL timeout with a clear error message if the shim hangs or doesn't respond

### Requirement: Configurable SSH Proxy Bind Address

The SSH-based provisioner SHALL support configuring the bind address for the SSH proxy to allow connections from external containers (e.g., WSL2).

#### Scenario: Default bind address

- **GIVEN** a configuration without `ansible_proxy_bind_address` specified
- **WHEN** the provisioner sets up the SSH proxy
- **THEN** it SHALL bind to `127.0.0.1`
- **AND** this SHALL ensure security by restricting access to localhost

#### Scenario: Custom bind address

- **GIVEN** a configuration with `ansible_proxy_bind_address = "0.0.0.0"`
- **WHEN** the provisioner sets up the SSH proxy
- **THEN** it SHALL bind to `0.0.0.0`
- **AND** this SHALL allow connections from other interfaces (e.g., container bridge)

### Requirement: Configurable Inventory Host Address

The SSH-based provisioner SHALL support configuring the host address used in the generated inventory file to ensure the execution environment can reach the host.

#### Scenario: Default inventory host

- **GIVEN** a configuration without `ansible_proxy_host` specified
- **WHEN** the provisioner generates the inventory file
- **THEN** it SHALL use `127.0.0.1` as the `ansible_host`
- **AND** this SHALL work for local execution environments

#### Scenario: Custom inventory host

- **GIVEN** a configuration with `ansible_proxy_host = "host.containers.internal"`
- **WHEN** the provisioner generates the inventory file
- **THEN** it SHALL use `host.containers.internal` as the `ansible_host`
- **AND** this SHALL allow the containerized execution environment to connect back to the host

### Requirement: Unbuffered Python Output

The SSH-based provisioner SHALL force unbuffered Python output to ensure logs are streamed immediately to Packer, preventing apparent hangs during connection timeouts.

#### Scenario: PYTHONUNBUFFERED injection

- **WHEN** the provisioner executes `ansible-navigator` for play execution
- **THEN** it SHALL inject `PYTHONUNBUFFERED=1` into the environment variables
- **AND** this SHALL ensure that Python output is flushed immediately

### Requirement: Navigator Config File Generation

The remote provisioner SHALL support generating ansible-navigator.yml configuration files from a declarative HCL map and using them to control ansible-navigator behavior.

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
- **THEN** it SHALL generate a temporary file named `/tmp/packer-navigator-cfg-<uuid>.yml` (or equivalent in system temp directory)
- **AND** the file SHALL contain valid YAML matching the ansible-navigator.yml schema
- **AND** the file path SHALL be recorded for cleanup

#### Scenario: ANSIBLE_NAVIGATOR_CONFIG set in environment

- **GIVEN** a generated ansible-navigator.yml file at a known path
- **WHEN** the provisioner executes ansible-navigator commands
- **THEN** it SHALL set `ANSIBLE_NAVIGATOR_CONFIG=/path/to/file` in the environment
- **AND** this SHALL occur for all ansible-navigator executions

#### Scenario: Cleanup after provisioning

- **GIVEN** a generated ansible-navigator.yml file
- **WHEN** provisioning completes (success or failure)
- **THEN** the temporary ansible-navigator.yml file SHALL be deleted

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
- **AND** ansible-navigator SHALL use its normal configuration search order

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

### Requirement: Configuration Validation

The remote provisioner Config.Validate() method SHALL validate all supported configuration options.

#### Scenario: Comprehensive validation

- **WHEN** Config.Validate() is called
- **THEN** it SHALL validate:
  - One or more `play` blocks are defined
  - Each play has a non-empty `target`
  - SSH connection parameters are valid
  - `navigator_config`, if specified, is a non-empty map
  - Command does not contain embedded arguments
  - **AND** removed options SHALL NOT be validated (they will fail HCL parsing)

#### Scenario: Removed options cause validation errors

- **GIVEN** a configuration attempting to use removed options like `execution_environment = "image:tag"`
- **WHEN** Packer parses the configuration
- **THEN** it SHALL fail with an error indicating the option is not recognized
- **AND** error messages SHOULD guide users to use navigator_config instead
- **AND** error messages SHOULD reference MIGRATION.md for help

### Requirement: Navigator Config Nested Structure Support

The remote provisioner's `navigator_config` field SHALL use explicit Go struct types with proper HCL2 spec generation to support the ansible-navigator.yml configuration schema while ensuring RPC serializability.

#### Scenario: Navigator config uses typed structs

- **GIVEN** the remote provisioner implementation
- **WHEN** examining the Config struct definition
- **THEN** the `NavigatorConfig` field SHALL be defined as `*NavigatorConfig` (pointer to struct type)
- **AND** it SHALL NOT use `map[string]interface{}`
- **AND** the `NavigatorConfig` type SHALL be defined as a Go struct with proper mapstructure tags

#### Scenario: Struct types support nested configuration

- **GIVEN** a configuration with `provisioner "ansible-navigator"`
- **AND** `navigator_config` block with nested execution environment settings:

  ```hcl
  navigator_config {
    mode = "stdout"
    
    execution_environment {
      enabled     = true
      image       = "quay.io/ansible/creator-ee:latest"
      pull_policy = "missing"
      
      environment_variables {
        ANSIBLE_REMOTE_TMP = "/tmp/.ansible/tmp"
      }
    }
  }
  ```

- **WHEN** Packer parses and validates the configuration
- **THEN** validation SHALL succeed
- **AND** all nested structures SHALL be properly parsed into their respective struct types
- **AND** field names SHALL use underscores (not hyphens) for HCL compatibility

#### Scenario: HCL2 spec uses RPC-serializable types

- **GIVEN** the generated HCL2 spec for the remote provisioner
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
  - `ansible_config` block with nested `defaults` and `ssh_connection` sections
  - `logging` configuration options
  - `playbook_artifact` settings
  - `collection_doc_cache` settings

#### Scenario: YAML generation works with typed structs

- **GIVEN** a configuration with typed `navigator_config`
- **WHEN** the provisioner generates the ansible-navigator.yml file
- **THEN** the YAML generation SHALL work correctly with the struct-based config
- **AND** the generated YAML SHALL match the expected ansible-navigator.yml schema
- **AND** nested structures SHALL be preserved in the YAML output

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

