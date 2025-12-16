## MODIFIED Requirements

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

