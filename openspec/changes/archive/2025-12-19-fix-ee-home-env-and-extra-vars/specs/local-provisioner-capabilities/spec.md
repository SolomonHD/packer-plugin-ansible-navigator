## MODIFIED Requirements

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

## ADDED Requirements

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

