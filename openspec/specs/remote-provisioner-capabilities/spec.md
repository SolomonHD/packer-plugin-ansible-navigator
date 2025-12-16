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

The SSH-based provisioner SHALL expose a consistent set of dependency-install configuration options:

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
- **WHEN** the provisioner executes any ansible-galaxy operation and any ansible-navigator play execution
- **THEN** it SHALL set `ANSIBLE_ROLES_PATH` to the provided `roles_path` value

#### Scenario: collections_path exported via ANSIBLE_COLLECTIONS_PATHS

- **GIVEN** a configuration with `collections_path` set
- **WHEN** the provisioner executes any ansible-galaxy operation and any ansible-navigator play execution
- **THEN** it SHALL set `ANSIBLE_COLLECTIONS_PATHS` to the provided `collections_path` value

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
- **AND** one or more path-like fields (e.g., `requirements_file`, `inventory_file`, Galaxy-related paths, play `target` when it is a playbook path, and play `vars_files` entries) are set to `"~"`
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

The SSH-based provisioner SHALL support a configurable timeout for the version check operation to prevent indefinite hangs when ansible-navigator cannot be located or is slow to respond, with enhanced error messages for version manager shim troubleshooting.

#### Scenario: Default timeout with shim detection

- **GIVEN** a configuration for `provisioner "ansible-navigator"` without `version_check_timeout` specified
- **AND** `skip_version_check` is not set or is `false`
- **WHEN** the provisioner performs version check
- **THEN** it SHALL first attempt shim detection and resolution
- **AND** if a shim is detected, it SHALL use the resolved path
- **AND** it SHALL apply a default timeout of 60 seconds to the version check
- **AND** if the timeout expires, the error message SHALL include shim troubleshooting guidance

#### Scenario: Shim resolution bypasses timeout issues

- **GIVEN** ansible-navigator is installed via asdf and would normally cause timeout
- **WHEN** the provisioner detects the asdf shim and successfully resolves to the real binary
- **THEN** the version check SHALL use the real binary path
- **AND** the version check SHALL complete within normal timeframes (< 5 seconds)
- **AND** no timeout error SHALL occur

#### Scenario: Manual command configuration bypasses shim detection

- **GIVEN** a configuration with `command = "/absolute/path/to/ansible-navigator"`
- **WHEN** the provisioner performs version check
- **THEN** shim detection MAY be skipped if the path is absolute and points to a real binary
- **OR** shim detection MAY still run but resolution SHALL use the configured absolute path
- **AND** existing manual configurations SHALL continue to work without changes

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

The remote provisioner SHALL support generating ansible-navigator.yml configuration files from a declarative HCL map and using them to control ansible-navigator behavior. The generated YAML SHALL conform to the ansible-navigator v24+/v25+ schema, including proper nested structure for execution environment fields.

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
- **THEN** it SHALL generate a temporary file named `/tmp/packer-navigator-cfg-<uuid>.yml` (or equivalent in system temp directory)
- **AND** the file SHALL contain valid YAML conforming to the ansible-navigator.yml v25+ schema
- **AND** the execution-environment pull policy SHALL be generated as nested structure: `pull: { policy: missing }`
- **AND** the file SHALL NOT use flat `pull-policy: missing` syntax (which is rejected by ansible-navigator v25+)
- **AND** the local temporary file SHALL be added to the cleanup list

#### Scenario: ANSIBLE_NAVIGATOR_CONFIG set in command execution

- **GIVEN** a generated ansible-navigator.yml file at a temporary path
- **WHEN** the provisioner executes ansible-navigator
- **THEN** it SHALL prepend `ANSIBLE_NAVIGATOR_CONFIG=<temp_path>` to the environment variables
- **AND** this SHALL occur for all ansible-navigator executions

#### Scenario: Cleanup after provisioning

- **GIVEN** a generated ansible-navigator.yml file
- **WHEN** provisioning completes (success or failure)
- **THEN** the temporary ansible-navigator.yml file SHALL be deleted

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
- **THEN** it SHALL automatically include execution-environment temp dir environment variables:

  ```yaml
  execution-environment:
    environment-variables:
      set:
        ANSIBLE_REMOTE_TMP: "/tmp/.ansible/tmp"
        ANSIBLE_LOCAL_TMP: "/tmp/.ansible-local"
  ```

- **AND** it SHALL configure Ansible temp directory defaults via an ansible.cfg referenced by `ansible.config.path` (NOT by emitting `defaults` under `ansible.config`)

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
- **AND** ansible-navigator SHALL use its normal configuration search order

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

#### Scenario: Complex nested structure preserved (except ansible.config)

- **GIVEN** a configuration with deeply nested navigator_config:

  ```hcl
  navigator_config {
    ansible_config {
      defaults {
        remote_tmp        = "/tmp/.ansible/tmp"
        host_key_checking = false
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
- **THEN** the nested structure SHALL be preserved for execution-environment and other supported ansible-navigator.yml sections
- **AND** `ansible.config` SHALL remain schema-compliant (help/path/cmdline only)
- **AND** the Ansible defaults and ssh_connection settings SHALL be represented via a generated ansible.cfg referenced by `ansible.config.path`

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

- **GIVEN** a configuration attempting to use removed options like `execution_environment = "image:tag"` or `work_dir = "/tmp/ansible"`
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

### Requirement: Mode CLI Flag Support for Remote Provisioner

The SSH-based provisioner SHALL pass the `--mode` CLI flag to ansible-navigator when `navigator_config.mode` is configured, ensuring ansible-navigator runs in the specified mode and does not hang waiting for interactive input.

#### Scenario: Mode flag added when navigator_config.mode is set

- **GIVEN** a configuration with `provisioner "ansible-navigator"`
- **AND** `navigator_config.mode = "stdout"` is specified
- **WHEN** the provisioner constructs the ansible-navigator command
- **THEN** it SHALL include `--mode stdout` flag in the command arguments
- **AND** the flag SHALL be inserted after the `run` subcommand
- **AND** the flag SHALL appear before playbook-specific arguments (inventory, extra vars, etc.)

#### Scenario: Mode flag not added when mode not configured

- **GIVEN** a configuration with `provisioner "ansible-navigator"`
- **AND** NO `navigator_config` block is specified
- **OR** `navigator_config` is specified but does NOT include `mode` field
- **WHEN** the provisioner constructs the ansible-navigator command
- **THEN** it SHALL NOT include a `--mode` flag
- **AND** ansible-navigator SHALL use its default behavior (config file or built-in defaults)

#### Scenario: Mode flag prevents interactive hang

- **GIVEN** a configuration with `navigator_config.mode = "stdout"`
- **AND** ansible-navigator is run in a non-interactive environment (Packer build)
- **WHEN** the provisioner executes ansible-navigator
- **THEN** ansible-navigator SHALL execute in stdout mode
- **AND** it SHALL NOT wait for terminal input
- **AND** it SHALL output playbook execution results to stdout
- **AND** the provisioning SHALL complete without hanging

### Requirement: Per-Play extra_args escape hatch (remote provisioner)

The SSH-based provisioner SHALL support a per-play `extra_args` field (list(string)) that is appended to the `ansible-navigator run` invocation for that play.

#### Scenario: play.extra_args is accepted in HCL schema

- **GIVEN** a configuration using `provisioner "ansible-navigator"`
- **AND** a `play {}` block includes `extra_args = ["--check", "--diff"]`
- **WHEN** Packer parses the configuration
- **THEN** parsing SHALL succeed
- **AND** the provisioner configuration SHALL include the `extra_args` values for that play

#### Scenario: play.extra_args is passed verbatim

- **GIVEN** a configuration using `provisioner "ansible-navigator"` with a play:

  ```hcl
  play {
    target     = "site.yml"
    extra_args = ["--check", "--diff"]
  }
  ```

- **WHEN** the provisioner constructs the ansible-navigator command for that play
- **THEN** it SHALL include both `--check` and `--diff` as command arguments
- **AND** it SHALL not rewrite, split, or validate the `extra_args` values beyond basic type handling

#### Scenario: Deterministic argument ordering includes extra_args

- **GIVEN** a configuration using `provisioner "ansible-navigator"` with:
  - one `play {}` block
  - `navigator_config.mode = "stdout"`
  - `play.extra_args = ["--check", "--diff"]`
- **WHEN** the provisioner constructs the ansible-navigator command arguments
- **THEN** the argument ordering SHALL be deterministic and consistent across executions:
  1. `run` subcommand
  2. enforced mode flag behavior (when configured), inserted immediately after `run`
  3. play-level `extra_args`
  4. provisioner-generated arguments (inventory, extra vars, tags, etc.)
  5. the play target (playbook path or generated playbook/role invocation)

### Requirement: YAML Root Structure for Remote Provisioner

The remote provisioner SHALL generate ansible-navigator.yml configuration files with a top-level `ansible-navigator:` root key wrapping all settings to conform to ansible-navigator v25.x schema requirements and prevent validation errors.

#### Scenario: Generated YAML wraps all settings under ansible-navigator key

- **GIVEN** a configuration with `navigator_config` block containing any settings
- **WHEN** the provisioner generates the `ansible-navigator.yml` file
- **THEN** the YAML SHALL have a top-level `ansible-navigator:` key
- **AND** ALL configuration settings SHALL be nested under this key
- **AND** the structure SHALL conform to ansible-navigator v25.x schema

#### Scenario: Mode setting nested under root key

- **GIVEN** a configuration with `navigator_config.mode = "stdout"`
- **WHEN** the YAML is generated
- **THEN** the output SHALL be:

  ```yaml
  ansible-navigator:
    mode: stdout
  ```

- **AND** NOT the flat structure:

  ```yaml
  mode: stdout
  ```

#### Scenario: Multiple configuration sections nested correctly

- **GIVEN** a configuration with multiple navigator_config sections (mode, execution_environment, ansible_config, logging)
- **WHEN** the YAML is generated
- **THEN** ALL sections SHALL be nested under `ansible-navigator:` root key
- **AND** nested structure SHALL be preserved for execution-environment, ansible, and other complex settings

#### Scenario: Schema validation passes with root key

- **GIVEN** a generated ansible-navigator.yml file with `ansible-navigator:` root key
- **WHEN** ansible-navigator processes the configuration file
- **THEN** it SHALL pass schema validation
- **AND** it SHALL NOT report "Additional properties are not allowed" errors
- **AND** all configuration settings SHALL be recognized and applied

#### Scenario: Execution environment pull policy with root key

- **GIVEN** a configuration with `execution_environment.pull_policy = "missing"`
- **WHEN** the ansible-navigator.yml YAML is generated
- **THEN** the YAML SHALL contain:

  ```yaml
  ansible-navigator:
    execution-environment:
      pull:
        policy: missing
  ```

- **AND** the generated YAML SHALL pass ansible-navigator's built-in schema validation
- **AND** ansible-navigator SHALL accept the config file without "Additional properties" errors

#### Scenario: ConvertToYAMLStructure wraps in root key

- **GIVEN** the `convertToYAMLStructure()` function implementation
- **WHEN** it converts NavigatorConfig to YAML-compatible structure
- **THEN** it SHALL create a top-level map with key `"ansible-navigator"`
- **AND** the value SHALL be a map containing all navigator settings
- **AND** nested structures SHALL be preserved within this wrapped structure
- **AND** field name conversions (underscores to hyphens) SHALL occur correctly

#### Scenario: Empty navigator_config produces minimal YAML

- **GIVEN** a `navigator_config {}` block with no fields set
- **WHEN** the YAML is generated
- **THEN** it SHALL produce:

  ```yaml
  ansible-navigator: {}
  ```

- **OR** an equivalent minimal structure
- **AND** ansible-navigator SHALL accept this as valid configuration

#### Scenario: Backward compatibility maintained

- **GIVEN** an existing Packer configuration using `navigator_config` block
- **AND** the configuration was written before this change
- **WHEN** the configuration is used with the updated plugin
- **THEN** it SHALL continue to work without modification
- **AND** the YAML SHALL be generated with proper root structure automatically
- **AND** no user action SHALL be required to adopt the new structure

### Requirement: Version Manager Shim Detection

The SSH-based provisioner SHALL detect when `ansible-navigator` resolves to a version manager shim (asdf, rbenv, pyenv) and attempt automatic resolution to prevent silent hangs caused by shim recursion loops.

#### Scenario: asdf shim detected and resolved successfully

- **GIVEN** ansible-navigator is installed via asdf
- **AND** the command resolves to `~/.asdf/shims/ansible-navigator`
- **WHEN** the provisioner performs version check in `getVersion()`
- **THEN** it SHALL detect the file is an asdf shim by reading the file header
- **AND** it SHALL execute `asdf which ansible-navigator` to resolve the real binary path
- **AND** it SHALL use the resolved path for the version check
- **AND** the version check SHALL succeed without hanging

#### Scenario: rbenv shim detected and resolved

- **GIVEN** ansible-navigator is installed via rbenv
- **AND** the command resolves to a rbenv shim file
- **WHEN** the provisioner performs version check
- **THEN** it SHALL detect the rbenv shim pattern in the file header
- **AND** it SHALL execute `rbenv which ansible-navigator` to resolve the real binary
- **AND** it SHALL use the resolved binary path for version check
- **AND** the version check SHALL complete successfully

#### Scenario: pyenv shim detected and resolved

- **GIVEN** ansible-navigator is installed via pyenv
- **AND** the command resolves to a pyenv shim file
- **WHEN** the provisioner performs version check
- **THEN** it SHALL detect the pyenv shim pattern in the file header
- **AND** it SHALL execute `pyenv which ansible-navigator` to resolve the real binary
- **AND** it SHALL use the resolved binary path for version check
- **AND** the version check SHALL complete successfully

#### Scenario: Shim detected but resolution fails

- **GIVEN** ansible-navigator command resolves to an asdf shim
- **AND** `asdf which ansible-navigator` returns an error or empty result
- **WHEN** the provisioner attempts version check
- **THEN** it SHALL fail immediately with error message:

  ```
  Error: ansible-navigator appears to be an asdf shim, but the real binary could not be resolved.
  
  SOLUTIONS:
  1. Verify ansible-navigator is installed:
     $ asdf list ansible
  
  2. Find the actual binary path:
     $ asdf which ansible-navigator
  
  3. Configure the plugin to use the actual binary:
     provisioner "ansible-navigator" {
       command = "/path/to/actual/ansible-navigator"
       # OR
       ansible_navigator_path = ["/path/to/bin/directory"]
     }
  
  4. Alternatively, skip the version check (not recommended):
     provisioner "ansible-navigator" {
       skip_version_check = true
     }
  ```

- **AND** the error message SHALL include the detected version manager name (asdf/rbenv/pyenv)

#### Scenario: Non-shim file detected correctly

- **GIVEN** ansible-navigator is a regular binary (not a shim)
- **WHEN** the provisioner performs shim detection
- **THEN** it SHALL correctly identify the file as NOT a shim
- **AND** it SHALL proceed with normal version check using the binary directly
- **AND** no resolution attempt SHALL be made

#### Scenario: Shim detection is fast

- **GIVEN** any ansible-navigator executable
- **WHEN** the provisioner performs shim detection
- **THEN** the detection overhead SHALL be less than 100 milliseconds
- **AND** it SHALL not significantly impact plugin initialization time

### Requirement: Enhanced Timeout Error Messages

The SSH-based provisioner SHALL provide enhanced error messages when version check times out, including guidance for troubleshooting version manager shim issues.

#### Scenario: Timeout error includes shim troubleshooting

- **WHEN** ansible-navigator version check times out after the configured duration
- **THEN** it SHALL return an error message that includes:

  ```
  ansible-navigator version check timed out after 60s.
  
  COMMON CAUSES:
  1. Version manager shim (asdf/rbenv/pyenv) causing infinite recursion
  2. ansible-navigator not properly installed or not in PATH
  3. ansible-navigator requires additional configuration
  
  TROUBLESHOOTING:
  1. Check if you're using a version manager:
     $ which ansible-navigator
     $ head -1 $(which ansible-navigator)  # Should show shebang
  
  2. If using asdf, find the real binary:
     $ asdf which ansible-navigator
  
  3. Configure the plugin with the actual binary:
     provisioner "ansible-navigator" {
       command = "/home/user/.asdf/installs/ansible/2.9.0/bin/ansible-navigator"
       # OR
       ansible_navigator_path = ["/home/user/.asdf/installs/ansible/2.9.0/bin"]
     }
  
  4. Verify ansible-navigator works independently:
     $ ansible-navigator --version
  
  5. Skip version check if needed (not recommended):
     provisioner "ansible-navigator" {
       skip_version_check = true
     }
  ```

- **AND** it SHALL mention version manager shims as the first common cause

### Requirement: Warn when `version_check_timeout` is ineffective due to `skip_version_check`

When users explicitly configure `version_check_timeout` but also set `skip_version_check = true`, the plugin SHALL emit a user-visible warning indicating that the timeout is ignored.

#### Scenario: Warning when skip_version_check=true and version_check_timeout explicitly set

Given: A configuration for `provisioner "ansible-navigator"` with `skip_version_check = true` and an explicitly set `version_check_timeout`
When: The provisioner prepares for execution (configuration validation/prepare)
Then: The provisioner prints a non-fatal warning in Packer UI output stating that `version_check_timeout` is ignored when `skip_version_check=true`

#### Scenario: No warning when skip_version_check=false

Given: A configuration for `provisioner "ansible-navigator"` with `skip_version_check = false` and an explicitly set `version_check_timeout`
When: The provisioner prepares for execution (configuration validation/prepare)
Then: No warning about `version_check_timeout` being ignored is printed

#### Scenario: No warning when version_check_timeout not explicitly set

Given: A configuration for `provisioner "ansible-navigator"` with `skip_version_check = true` and without an explicitly set `version_check_timeout`
When: The provisioner prepares for execution (configuration validation/prepare)
Then: No warning about `version_check_timeout` being ignored is printed

### Requirement: `work_dir` is not supported (remote provisioner)

The SSH-based `ansible-navigator` provisioner SHALL NOT support a `work_dir` configuration field.

#### Scenario: `work_dir` rejected at HCL parse time

Given: a configuration using `provisioner "ansible-navigator"` that includes `work_dir = "/tmp/ansible"`
When: Packer parses the configuration
Then: parsing SHALL fail with an error indicating `work_dir` is not a recognized argument

### Requirement: `ansible_config.defaults.local_tmp` support (remote provisioner)

The SSH-based provisioner SHALL support configuring Ansible's local temp directory via `navigator_config.ansible_config.defaults.local_tmp`.

#### Scenario: HCL schema accepts local_tmp under ansible_config.defaults

- **GIVEN** a configuration using `provisioner "ansible-navigator"` with:

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

- **GIVEN** a configuration using `provisioner "ansible-navigator"` that sets `navigator_config.ansible_config.defaults.local_tmp`
- **WHEN** the provisioner prepares for execution and generates configuration artifacts
- **THEN** it SHALL generate an ansible.cfg that includes `local_tmp` under `[defaults]`
- **AND** the generated `ansible-navigator.yml` SHALL reference the generated ansible.cfg via `ansible.config.path`
- **AND** the generated `ansible-navigator.yml` MUST NOT emit `defaults` under `ansible.config` (schema-compliance requirement)

#### Scenario: local_tmp omitted when unset

- **GIVEN** a configuration that does not set `navigator_config.ansible_config.defaults.local_tmp`
- **WHEN** the provisioner generates ansible.cfg
- **THEN** the generated ansible.cfg SHALL NOT include a `local_tmp` entry
