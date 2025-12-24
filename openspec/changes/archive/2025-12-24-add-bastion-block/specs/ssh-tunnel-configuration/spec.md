# ssh-tunnel-configuration Specification Delta

## ADDED Requirements

### Requirement: Bastion Nested Block Configuration

The remote ansible-navigator provisioner SHALL support a nested `bastion {}` HCL block for configuring SSH bastion (jump host) parameters when using SSH tunnel mode.

#### Scenario: Bastion block can be configured

- **GIVEN** a configuration for `provisioner "ansible-navigator"`
- **AND** the configuration includes a `bastion { }` block with fields
- **WHEN** Packer parses the configuration
- **THEN** parsing SHALL succeed
- **AND** the provisioner SHALL store the bastion configuration in a `BastionConfig` struct

#### Scenario: Bastion host field in block is required for SSH tunnel mode

- **GIVEN** a configuration with `connection_mode = "ssh_tunnel"`
- **AND** a `bastion { }` block is provided
- **AND** the block does not specify `host`
- **WHEN** the configuration is validated
- **THEN** validation SHALL fail
- **AND** the error message SHALL state "bastion.host is required when connection_mode='ssh_tunnel'"

#### Scenario: Bastion user field in block is required for SSH tunnel mode

- **GIVEN** a configuration with `connection_mode = "ssh_tunnel"`
- **AND** a `bastion { }` block with `host` specified
- **AND** the block does not specify `user`
- **WHEN** the configuration is validated
- **THEN** validation SHALL fail
- **AND** the error message SHALL state "bastion.user is required when connection_mode='ssh_tunnel'"

#### Scenario: Bastion port defaults to 22 in nested block

- **GIVEN** a configuration with `connection_mode = "ssh_tunnel"`
- **AND** a `bastion { }` block is provided
- **AND** the block does not specify `port`
- **WHEN** the provisioner initializes
- **THEN** `bastion.port` SHALL default to 22

#### Scenario: Bastion authentication requires credentials in nested block

- **GIVEN** a configuration with `connection_mode = "ssh_tunnel"`
- **AND** a `bastion { }` block with `host` and `user` specified
- **AND** neither `private_key_file` nor `password` is specified in the block
- **WHEN** the configuration is validated
- **THEN** validation SHALL fail
- **AND** the error message SHALL state "either bastion.private_key_file or bastion.password must be provided when connection_mode='ssh_tunnel'"

#### Scenario: Bastion port validation enforces valid range

- **GIVEN** a configuration with `connection_mode = "ssh_tunnel"`
- **AND** a `bastion { }` block with `port = 99999`
- **WHEN** the configuration is validated
- **THEN** validation SHALL fail
- **AND** the error message SHALL state "bastion.port must be between 1 and 65535"

#### Scenario: Valid bastion block configuration accepted

- **GIVEN** a configuration with:

  ```hcl
  connection_mode = "ssh_tunnel"
  bastion {
    host             = "bastion.example.com"
    user             = "deploy"
    private_key_file = "~/.ssh/bastion_key"
    port             = 2222
  }
  ```

- **AND** the file `~/.ssh/bastion_key` exists
- **WHEN** the configuration is validated
- **THEN** validation SHALL succeed

#### Scenario: HOME expansion applied to private_key_file in nested block

- **GIVEN** a bastion block with `private_key_file = "~/.ssh/bastion_key"`
- **WHEN** the provisioner prepares for execution
- **THEN** the `~` SHALL be expanded to the user's HOME directory
- **AND** subsequent validation SHALL use the expanded absolute path

#### Scenario: Enabled field auto-set when host is provided

- **GIVEN** a bastion block with `host = "bastion.example.com"`
- **AND** the `enabled` field is not explicitly set
- **WHEN** the provisioner prepares for execution
- **THEN** `bastion.enabled` SHALL be automatically set to `true`

## MODIFIED Requirements

### Requirement: Bastion Authentication Configuration

The remote ansible-navigator provisioner SHALL support both key-based and password-based authentication to the bastion host using the nested `bastion {}` block or legacy flat fields.

#### Scenario: Either key file or password is required (nested block)

- **GIVEN** a configuration with `connection_mode = "ssh_tunnel"`
- **AND** a `bastion { }` block with `host` and `user`
- **AND** neither `private_key_file` nor `password` is provided in the block
- **WHEN** the configuration is validated
- **THEN** validation SHALL fail
- **AND** the error message SHALL state "either bastion.private_key_file or bastion.password must be provided when connection_mode='ssh_tunnel'"

#### Scenario: Key file authentication is accepted (nested block)

- **GIVEN** a configuration with:
  - `connection_mode = "ssh_tunnel"`
  - bastion block with `host = "bastion.example.com"`
  - bastion block with `user = "deploy"`
  - bastion block with `private_key_file = "~/.ssh/bastion_key"`
- **AND** the file `~/.ssh/bastion_key` exists
- **WHEN** the configuration is validated
- **THEN** validation SHALL succeed

#### Scenario: Password authentication is accepted (nested block)

- **GIVEN** a configuration with:
  - `connection_mode = "ssh_tunnel"`
  - bastion block with `host = "10.0.1.100"`
  - bastion block with `user = "jumpuser"`
  - bastion block with `password = "secret"`
- **WHEN** the configuration is validated
- **THEN** validation SHALL succeed

#### Scenario: Both key and password can be provided (nested block)

- **GIVEN** a configuration with:
  - `connection_mode = "ssh_tunnel"`
  - bastion block with `private_key_file = "~/.ssh/bastion_key"`
  - bastion block with `password = "secret"`
- **WHEN** the configuration is validated
- **THEN** validation SHALL succeed
- **AND** both authentication methods SHALL be available

#### Scenario: Key file must exist (nested block)

- **GIVEN** a configuration with:
  - `connection_mode = "ssh_tunnel"`
  - bastion block with `private_key_file = "/nonexistent/key"`
- **WHEN** the configuration is validated
- **THEN** validation SHALL fail
- **AND** the error message SHALL state "bastion.private_key_file: /nonexistent/key does not exist"

#### Scenario: HOME expansion applied to key file path (nested block)

- **GIVEN** a bastion block with `private_key_file = "~/.ssh/bastion_key"`
- **WHEN** the provisioner prepares for execution
- **THEN** the `~` SHALL be expanded to the user's HOME directory
- **AND** subsequent validation SHALL use the expanded absolute path

#### Scenario: Legacy flat fields still work (deprecated)

- **GIVEN** a configuration with:
  - `connection_mode = "ssh_tunnel"`
  - `bastion_host = "bastion.example.com"`
  - `bastion_user = "deploy"`
  - `bastion_private_key_file = "~/.ssh/bastion_key"`
- **AND** no `bastion { }` block is present
- **AND** the file `~/.ssh/bastion_key` exists
- **WHEN** the configuration is validated
- **THEN** validation SHALL succeed
- **AND** a deprecation warning SHALL be displayed

#### Scenario: Nested block takes precedence over legacy flat fields

- **GIVEN** a configuration with:
  - `connection_mode = "ssh_tunnel"`
  - `bastion_host = "old.example.com"` (legacy)
  - `bastion { host = "new.example.com" }` (nested)
  - All required fields in both formats
- **WHEN** the provisioner prepares for execution
- **THEN** the nested block values SHALL take precedence
- **AND** `p.config.Bastion.Host` SHALL equal `"new.example.com"`
- **AND** a deprecation warning SHALL be displayed about flat fields

## REMOVED Requirements

None. Legacy flat bastion fields are deprecated but not removed for backward compatibility.
