# local-provisioner-capabilities Specification Deltas

## ADDED Requirements

### Requirement: Navigator Config File Generation

The local provisioner SHALL support generating ansible-navigator.yml configuration files from a declarative HCL map, uploading them to the target, and using them to control all ansible-navigator behavior.

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
- **THEN** it SHALL generate a temporary file named `/tmp/packer-navigator-cfg-<uuid>.yml` (or equivalent in system temp directory) on the LOCAL machine
- **AND** the file SHALL contain valid YAML matching the ansible-navigator.yml schema
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
- **AND** ansible-navigator SHALL use its normal configuration search order on the target

#### Scenario: Empty navigator_config rejected

- **GIVEN** a configuration with `navigator_config = {}`
- **WHEN** the configuration is validated
- **THEN** validation SHALL fail with a clear error message
- **AND** the error SHALL explain that navigator_config must contain at least one configuration section

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

## REMOVED Requirements

### Requirement: Execution Environment Support

**Reason**: Configuration via string `execution_environment` field is replaced by `navigator_config.execution-environment` object.

**Migration**: Users should move from:

```hcl
execution_environment = "quay.io/ansible/creator-ee:latest"
```

To:

```hcl
navigator_config = {
  execution-environment = {
    enabled = true
    image = "quay.io/ansible/creator-ee:latest"
  }
}
```

### Requirement: Navigator Mode

**Reason**: Configuration via `navigator_mode` field is replaced by `navigator_config.mode`.

**Migration**: Users should move from:

```hcl
navigator_mode = "stdout"
```

To:

```hcl
navigator_config = {
  mode = "stdout"
}
```

### Requirement: Ansible Configuration File Generation and Upload

**Reason**: Configuration via `ansible_cfg` map is replaced by `navigator_config.ansible.config`.

**Migration**: Users should move from:

```hcl
ansible_cfg = {
  defaults = {
    remote_tmp = "/tmp/.ansible/tmp"
  }
}
```

To:

```hcl
navigator_config = {
  ansible = {
    config = {
      defaults = {
        remote_tmp = "/tmp/.ansible/tmp"
      }
    }
  }
}
```

## MODIFIED Requirements

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
