# remote-provisioner-capabilities Specification Deltas

## ADDED Requirements

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
