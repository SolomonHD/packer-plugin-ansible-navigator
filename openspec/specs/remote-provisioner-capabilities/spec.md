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
- **THEN** it SHALL pass `--ee true --eei quay.io/ansible/creator-ee:latest` flags to ansible-navigator
- **AND** the `--ee` flag SHALL be a boolean (`true`) to enable execution environment mode
- **AND** the `--eei` flag SHALL specify the container image
- **AND** the container SHALL have network access to reach the target via SSH

#### Scenario: No execution environment specified

- **GIVEN** a configuration without `execution_environment` specified
- **WHEN** the provisioner executes
- **THEN** it SHALL NOT pass `--ee` or `--eei` flags
- **AND** ansible-navigator SHALL use its default execution environment behavior

#### Scenario: Execution environment with custom registry

- **GIVEN** a configuration with `execution_environment = "myregistry.io/custom-ansible-ee:v1.0"`
- **WHEN** the provisioner constructs the command
- **THEN** it SHALL pass `--ee true --eei myregistry.io/custom-ansible-ee:v1.0` flags
- **AND** the full image reference SHALL be preserved including registry, repository, and tag

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

### Requirement: Ansible Configuration File Generation

The SSH-based provisioner SHALL support generating a temporary ansible.cfg file from a declarative configuration map, automatically applying defaults for execution environment use cases.

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
- **THEN** it SHALL generate a temporary file named `/tmp/packer-ansible-cfg-<random>.ini` (or equivalent in system temp directory)
- **AND** the file SHALL contain valid INI format with sections and key-value pairs
- **AND** the file path SHALL be added to the cleanup list

#### Scenario: Generated INI file format

- **GIVEN** a configuration with multiple sections in `ansible_cfg`
- **WHEN** the temporary ansible.cfg file is generated
- **THEN** it SHALL use standard INI format with `[section_name]` headers
- **AND** each key-value pair SHALL be written as `key = value`
- **AND** sections SHALL be separated by blank lines for readability
- **AND** all string values SHALL be written without quotes

#### Scenario: ANSIBLE_CONFIG environment variable set

- **GIVEN** a generated ansible.cfg file at path `/tmp/packer-ansible-cfg-ABC123.ini`
- **WHEN** the provisioner executes ansible-navigator
- **THEN** it SHALL set `ANSIBLE_CONFIG=/tmp/packer-ansible-cfg-ABC123.ini` in the command environment
- **AND** the environment variable SHALL be present for both version checks and play execution

#### Scenario: Cleanup after provisioning success

- **GIVEN** a generated ansible.cfg file
- **WHEN** provisioning completes successfully
- **THEN** the temporary ansible.cfg file SHALL be deleted
- **AND** the cleanup SHALL occur in a deferred function to ensure execution even if other cleanup fails

#### Scenario: Cleanup after provisioning failure

- **GIVEN** a generated ansible.cfg file
- **WHEN** a play or playbook execution fails
- **THEN** the temporary ansible.cfg file SHALL still be deleted
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
- **AND** it SHALL NOT set the ANSIBLE_CONFIG environment variable
- **AND** Ansible SHALL use its normal configuration search order

#### Scenario: Ansible's normal config file precedence honored

- **GIVEN** a user has an existing `ansible.cfg` file in their working directory
- **AND** the provisioner has generated a temporary ansible.cfg file
- **WHEN** ansible-navigator executes
- **THEN** Ansible's normal configuration precedence SHALL apply
- **AND** if the user's `ansible.cfg` is in a higher-precedence location, it SHALL take priority
- **AND** the generated file SHALL only be used if no higher-precedence config exists

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
- **AND** Ansible itself SHALL handle unknown options per its normal behavior

#### Scenario: Special characters in values preserved

- **GIVEN** a configuration with `ansible_cfg` containing values with special characters (spaces, quotes, etc.)
- **WHEN** the INI file is generated
- **THEN** values SHALL be written literally without additional quoting
- **AND** special characters SHALL be preserved as-is
- **AND** bash-style quoting or escaping SHALL NOT be applied

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
  - All legacy options are still validated (no breaking changes)

