## MODIFIED Requirements

### Requirement: Navigator Config Nested Structure Support

The local provisioner's `navigator_config` field SHALL use explicit Go struct types with proper HCL2 spec generation to support the ansible-navigator.yml configuration schema while ensuring RPC serializability.

#### Scenario: Navigator config uses typed structs

- **GIVEN** the local provisioner implementation
- **WHEN** examining the Config struct definition
- **THEN** the `NavigatorConfig` field SHALL be defined as `*NavigatorConfig` (pointer to struct type)
- **AND** it SHALL NOT use `map[string]interface{}`
- **AND** the `NavigatorConfig` type SHALL be defined as a Go struct with proper mapstructure tags

#### Scenario: Struct types support nested configuration

- **GIVEN** a configuration with `provisioner "ansible-navigator-local"`
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

- **GIVEN** the generated HCL2 spec for the local provisioner
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
