# ansible-navigator-config-schema Specification

## Capability

Generate ansible-navigator.yml configuration files that conform to the official ansible-navigator settings schema, specifically ensuring the `ansible.config` section uses only valid properties.

---

## MODIFIED Requirements

### Requirement: Valid ansible.config Structure

The plugin SHALL generate `ansible-navigator.yml` files where the `ansible.config` section contains ONLY the properties defined in the ansible-navigator settings schema: `help`, `path`, and `cmdline`.

#### Scenario: Generate valid ansible.config section

- **GIVEN** a configuration with `navigator_config` block
- **AND** execution environment is enabled requiring Ansible configuration
- **WHEN** the provisioner generates ansible-navigator.yml
- **THEN** the `ansible.config` section SHALL contain ONLY valid properties:
  - `help` (boolean, optional)
  - `path` (string, optional - path to ansible.cfg file)
  - `cmdline` (string, optional - additional command-line arguments)
- **AND** the section SHALL NOT contain `defaults`, `ssh_connection`, or any other unsupported properties
- **AND** the generated YAML SHALL pass ansible-navigator's built-in schema validation

#### Scenario: No nested Ansible configuration in ansible.config

- **GIVEN** any `navigator_config` configuration
- **WHEN** the YAML is generated
- **THEN** there SHALL NOT be nested Ansible configuration sections under `ansible.config`
- **AND** specifically, `ansible.config.defaults` SHALL NOT exist
- **AND** specifically, `ansible.config.ssh_connection` SHALL NOT exist
- **AND** the structure SHALL match:

  ```yaml
  ansible-navigator:
    ansible:
      config:
        help: false              # Optional boolean
        path: "/path/to/ansible.cfg"  # Optional string
        cmdline: "--forks 15"    # Optional string
  ```

---

## ADDED Requirements

### Requirement: Ansible.cfg File Generation

When Ansible-specific configuration is needed (e.g., temp directories, SSH settings), the plugin SHALL generate a separate `ansible.cfg` file and reference it via `ansible.config.path`.

#### Scenario: Generate separate ansible.cfg for EE defaults

- **GIVEN** a configuration with execution environment enabled
- **AND** Ansible configuration is required (e.g., `remote_tmp`, `host_key_checking`)
- **WHEN** the provisioner prepares for execution
- **THEN** it SHALL generate a temporary file named `/tmp/packer-ansible-cfg-<uuid>.cfg`
- **AND** the file SHALL contain valid INI-formatted Ansible configuration:

  ```ini
  [defaults]
  remote_tmp = /tmp/.ansible/tmp
  host_key_checking = False

  [ssh_connection]
  ssh_timeout = 30
  pipelining = True
  ```

- **AND** the ansible-navigator.yml SHALL reference this file:

  ```yaml
  ansible-navigator:
    ansible:
      config:
        path: /tmp/packer-ansible-cfg-<uuid>.cfg
  ```

- **AND** the temp file SHALL be added to the cleanup list

#### Scenario: Cleanup ansible.cfg on completion

- **GIVEN** a generated temporary ansible.cfg file
- **WHEN** the provisioner completes (success or failure)
- **THEN** the temporary ansible.cfg SHALL be removed from the filesystem
- **AND** cleanup SHALL occur regardless of provisioning success or failure

#### Scenario: User-provided ansible.cfg path

- **GIVEN** a configuration with `navigator_config.ansible_config.config = "/custom/path/ansible.cfg"`
- **WHEN** the YAML is generated
- **THEN** the plugin SHALL use the user-provided path
- **AND** SHALL NOT generate a temporary ansible.cfg
- **AND** SHALL NOT apply automatic EE defaults to the user-provided file

---

### Requirement: HCL Configuration Surface Update

The HCL configuration for `navigator_config.ansible_config` SHALL support only the properties defined in the ansible-navigator schema.

#### Scenario: Supported ansible_config fields

- **GIVEN** HCL2 configuration for the provisioner
- **WHEN** user specifies `navigator_config.ansible_config` block
- **THEN** the following fields SHALL be supported:
  - `config` (string) - path to existing ansible.cfg file
  - `path` (string) - alias for `config` (maps to same ansible.config.path)
  - `help` (boolean) - show ansible-config help
  - `cmdline` (string) - additional command-line arguments
- **AND** the following fields SHALL be removed/unsupported:
  - `defaults` (object) - no longer valid per schema
  - `ssh_connection` (object) - no longer valid per schema

#### Scenario: Rejected invalid fields

- **GIVEN** HCL configuration with `navigator_config.ansible_config.defaults`
- **WHEN** the configuration is parsed
- **THEN** Packer SHALL reject the configuration with an error
- **AND** the error message SHALL indicate the field is not supported by ansible-navigator schema
- **AND** SHALL suggest using a separate ansible.cfg file instead

---

### Requirement: Automatic EE Configuration via ansible.cfg

When execution environment is enabled, the plugin SHALL automatically generate ansible.cfg with appropriate defaults if no custom ansible.cfg is provided.

#### Scenario: Auto-generate ansible.cfg for EE

- **GIVEN** `navigator_config.execution_environment.enabled = true`
- **AND** no `navigator_config.ansible_config.config` or `.path` is specified
- **WHEN** the provisioner generates configuration
- **THEN** it SHALL automatically create a temporary ansible.cfg with defaults:
  - `[defaults] remote_tmp = /tmp/.ansible/tmp`
  - `[defaults] host_key_checking = False`
- **AND** SHALL set `ansible.config.path` to point to this temp file
- **AND** environment variables SHALL still be set:
  - `ANSIBLE_REMOTE_TMP=/tmp/.ansible/tmp`
  - `ANSIBLE_LOCAL_TMP=/tmp/.ansible-local`

#### Scenario: Skip auto-generation when user provides ansible.cfg

- **GIVEN** `navigator_config.execution_environment.enabled = true`
- **AND** `navigator_config.ansible_config.config = "/custom/ansible.cfg"`
- **WHEN** the provisioner generates configuration
- **THEN** it SHALL NOT generate a temporary ansible.cfg
- **AND** SHALL use the user-provided path as-is
- **AND** SHALL NOT modify the user-provided file
- **AND** environment variables SHALL still be set for the execution environment

---

## REMOVED Requirements

- Nested Ansible Configuration in ansible.config: The ansible-navigator schema does not support nested Ansible configuration under `ansible.config`. This must be provided via a separate `ansible.cfg` file. Previously, the plugin generated `ansible.config.defaults` and `ansible.config.ssh_connection` sections. Now, the plugin generates a separate `ansible.cfg` file and references it via `ansible.config.path`.
