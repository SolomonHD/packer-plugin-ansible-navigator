# local-provisioner-capabilities Spec Delta

## ADDED Requirements

### Requirement: Navigator Config Nested Structure Support

The local provisioner's `navigator_config` field SHALL accept arbitrary nested map structures to support the full ansible-navigator.yml configuration schema.

#### Scenario: Nested execution-environment configuration

- **GIVEN** a configuration with `provisioner "ansible-navigator-local"`
- **AND** `navigator_config` contains nested execution-environment settings:

  ```hcl
  navigator_config = {
    execution-environment = {
      enabled = true
      image   = "quay.io/ansible/creator-ee:latest"
      environment-variables = {
        set = {
          ANSIBLE_REMOTE_TMP = "/tmp/.ansible/tmp"
        }
      }
    }
  }
  ```

- **WHEN** Packer validates the configuration
- **THEN** validation SHALL succeed
- **AND** the nested structure SHALL be preserved in the Go `map[string]interface{}` representation

#### Scenario: Deeply nested ansible config section

- **GIVEN** a configuration with multiple levels of nesting:

  ```hcl
  navigator_config = {
    ansible = {
      config = {
        defaults = {
          host_key_checking = "False"
          remote_tmp = "/tmp/.ansible/tmp"
        }
        ssh_connection = {
          pipelining = "True"
          timeout = "30"
        }
      }
    }
  }
  ```

- **WHEN** Packer validates the configuration
- **THEN** validation SHALL succeed
- **AND** all nested levels SHALL be accessible in the config struct

#### Scenario: Mixed types in nested config

- **GIVEN** a navigator_config with boolean, string, and nested map values:

  ```hcl
  navigator_config = {
    mode = "stdout"
    execution-environment = {
      enabled = true
      image   = "my-ee:latest"
    }
  }
  ```

- **WHEN** Packer parses the configuration
- **THEN** it SHALL accept the mixed types
- **AND** boolean values SHALL be preserved as booleans
- **AND** string values SHALL be preserved as strings
- **AND** nested maps SHALL be preserved as nested maps

#### Scenario: Backward compatibility with flat string maps

- **GIVEN** a configuration using navigator_config as a flat map:

  ```hcl
  navigator_config = {
    mode = "stdout"
  }
  ```

- **WHEN** Packer validates the configuration
- **THEN** validation SHALL succeed
- **AND** flat map usage SHALL continue to work without changes

#### Scenario: HCL2 spec uses DynamicPseudoType

- **GIVEN** the generated HCL2 spec for the local provisioner
- **WHEN** examining the spec for `navigator_config`
- **THEN** it SHALL use `cty.DynamicPseudoType` as the attribute type
- **AND** it SHALL NOT use `cty.Map(cty.String)` or other restrictive types

#### Scenario: Validation error shows useful message for invalid types

- **GIVEN** a configuration with an invalid type in navigator_config (e.g. a number where only string/bool/map are valid)
- **WHEN** Packer validates the configuration
- **THEN** it SHALL fail with a clear error message
- **AND** the error SHALL indicate the field and type mismatch
