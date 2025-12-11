# local-provisioner-capabilities Specification Deltas

## ADDED Requirements

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
