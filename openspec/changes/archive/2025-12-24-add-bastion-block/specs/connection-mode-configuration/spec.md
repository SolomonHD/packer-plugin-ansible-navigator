# connection-mode-configuration Specification Delta

## MODIFIED Requirements

### Requirement: SSH Tunnel Mode Selection

When `connection_mode = "ssh_tunnel"`, the provisioner SHALL establish an SSH tunnel through a bastion host instead of using the proxy adapter.

#### Scenario: SSH tunnel mode requires bastion configuration

- **GIVEN** a configuration with `connection_mode = "ssh_tunnel"`
- **AND** no `bastion { }` block is provided or `bastion.host` is empty
- **WHEN** the configuration is validated
- **THEN** validation SHALL fail
- **AND** the error message SHALL state "bastion.host is required when connection_mode='ssh_tunnel'"

#### Scenario: SSH tunnel mode requires bastion user

- **GIVEN** a configuration with `connection_mode = "ssh_tunnel"`
- **AND** a `bastion { }` block with `host` is provided
- **AND** `bastion.user` is not provided
- **WHEN** the configuration is validated
- **THEN** validation SHALL fail
- **AND** the error message SHALL state "bastion.user is required when connection_mode='ssh_tunnel'"

#### Scenario: SSH tunnel mode requires authentication credentials

- **GIVEN** a configuration with `connection_mode = "ssh_tunnel"`
- **AND** a `bastion { }` block with `host` and `user` specified
- **AND** neither `bastion.private_key_file` nor `bastion.password` are provided
- **WHEN** the configuration is validated
- **THEN** validation SHALL fail
- **AND** the error SHALL require either `bastion.private_key_file` or `bastion.password`

#### Scenario: SSH tunnel mode with valid configuration succeeds validation

- **GIVEN** a configuration with:
  - `connection_mode = "ssh_tunnel"`
  - `bastion { }` block with `host` provided
  - `bastion { }` block with `user` provided
  - Either `bastion.private_key_file` or `bastion.password` provided in the block
- **WHEN** the configuration is validated
- **THEN** validation SHALL succeed

#### Scenario: SSH tunnel mode establishes tunnel

- **GIVEN** a configuration with `connection_mode = "ssh_tunnel"`
- **AND** All required bastion fields are provided in the `bastion { }` block
- **WHEN** [`Provision()`](../../../../../provisioner/ansible-navigator/provisioner.go:1351) is called
- **THEN** It SHALL NOT call [`setupAdapter()`](../../../../../provisioner/ansible-navigator/provisioner.go:1039)
- **AND** It SHALL call [`setupSSHTunnel()`](../../../../../provisioner/ansible-navigator/provisioner.go:1122) with target host and port
- **AND** `generatedData["Host"]` SHALL be overridden to `"127.0.0.1"`
- **AND** `generatedData["Port"]` SHALL be overridden to the tunnel's local port

#### Scenario: Legacy flat bastion fields still supported with deprecation warning

- **GIVEN** a configuration with `connection_mode = "ssh_tunnel"`
- **AND** legacy flat fields `bastion_host`, `bastion_user`, `bastion_private_key_file` are provided
- **AND** no `bastion { }` block is present
- **WHEN** the provisioner prepares for execution
- **THEN** the flat fields SHALL be migrated to the nested `bastion { }` structure
- **AND** a deprecation warning SHALL be displayed
- **AND** validation and provisioning SHALL proceed using the migrated values
