# ssh-tunnel-configuration Specification Delta

## Purpose

This capability adds SSH tunnel mode configuration to the ansible-navigator provisioner, enabling direct SSH tunneling through a bastion host as an alternative to the Packer SSH proxy adapter.

## ADDED Requirements

### Requirement: SSH Tunnel Mode Configuration

The remote ansible-navigator provisioner SHALL support an optional SSH tunnel mode that bypasses the Packer SSH proxy adapter by connecting through a bastion/jump host.

#### Scenario: SSH tunnel mode can be enabled

- **GIVEN** a configuration for `provisioner "ansible-navigator"`
- **AND** the configuration includes `ssh_tunnel_mode = true`
- **WHEN** Packer parses the configuration
- **THEN** parsing SHALL succeed
- **AND** the provisioner SHALL store the tunnel mode flag

#### Scenario: SSH tunnel mode is disabled by default

- **GIVEN** a configuration for `provisioner "ansible-navigator"`
- **AND** `ssh_tunnel_mode` is not specified
- **WHEN** the provisioner initializes
- **THEN** `ssh_tunnel_mode` SHALL default to `false`
- **AND** the proxy adapter behavior SHALL remain active

### Requirement: Bastion Host Configuration

The remote ansible-navigator provisioner SHALL support configuration of bastion host connection parameters when SSH tunnel mode is enabled.

#### Scenario: Bastion host and user are required when tunnel mode is enabled

- **GIVEN** a configuration with `ssh_tunnel_mode = true`
- **AND** `bastion_host` is not provided
- **WHEN** the configuration is validated
- **THEN** validation SHALL fail
- **AND** the error message SHALL state "bastion_host is required when ssh_tunnel_mode is true"

#### Scenario: Bastion user is required when tunnel mode is enabled

- **GIVEN** a configuration with `ssh_tunnel_mode = true`
- **AND** `bastion_user` is not provided
- **WHEN** the configuration is validated
- **THEN** validation SHALL fail
- **AND** the error message SHALL state "bastion_user is required when ssh_tunnel_mode is true"

#### Scenario: Bastion port has a reasonable default

- **GIVEN** a configuration with `ssh_tunnel_mode = true`
- **AND** `bastion_port` is not specified
- **WHEN** the provisioner initializes
- **THEN** `bastion_port` SHALL default to 22

#### Scenario: Bastion port must be in valid range

- **GIVEN** a configuration with `ssh_tunnel_mode = true`
- **AND** `bastion_port = 99999`
- **WHEN** the configuration is validated
- **THEN** validation SHALL fail
- **AND** the error message SHALL state "bastion_port must be between 1 and 65535"

#### Scenario: Valid bastion port accepted

- **GIVEN** a configuration with `ssh_tunnel_mode = true`
- **AND** `bastion_port = 2222`
- **WHEN** the configuration is validated
- **THEN** validation SHALL succeed

### Requirement: Bastion Authentication Configuration

The remote ansible-navigator provisioner SHALL support both key-based and password-based authentication to the bastion host.

#### Scenario: Either key file or password is required

- **GIVEN** a configuration with `ssh_tunnel_mode = true`
- **AND** neither `bastion_private_key_file` nor `bastion_password` is provided
- **WHEN** the configuration is validated
- **THEN** validation SHALL fail
- **AND** the error message SHALL state "either bastion_private_key_file or bastion_password is required when ssh_tunnel_mode is true"

#### Scenario: Key file authentication is accepted

- **GIVEN** a configuration with:
  - `ssh_tunnel_mode = true`
  - `bastion_host = "bastion.example.com"`
  - `bastion_user = "deploy"`
  - `bastion_private_key_file = "~/.ssh/bastion_key"`
- **AND** the file `~/.ssh/bastion_key` exists
- **WHEN** the configuration is validated
- **THEN** validation SHALL succeed

#### Scenario: Password authentication is accepted

- **GIVEN** a configuration with:
  - `ssh_tunnel_mode = true`
  - `bastion_host = "10.0.1.100"`
  - `bastion_user = "jumpuser"`
  - `bastion_password = "secret"`
- **WHEN** the configuration is validated
- **THEN** validation SHALL succeed

#### Scenario: Both key and password can be provided

- **GIVEN** a configuration with:
  - `ssh_tunnel_mode = true`
  - `bastion_private_key_file = "~/.ssh/bastion_key"`
  - `bastion_password = "secret"`
- **WHEN** the configuration is validated
- **THEN** validation SHALL succeed
- **AND** both authentication methods SHALL be available

#### Scenario: Key file must exist

- **GIVEN** a configuration with:
  - `ssh_tunnel_mode = true`
  - `bastion_private_key_file = "/nonexistent/key"`
- **WHEN** the configuration is validated
- **THEN** validation SHALL fail
- **AND** the error message SHALL state "bastion_private_key_file: /nonexistent/key does not exist"

#### Scenario: HOME expansion applied to key file path

- **GIVEN** a configuration with `bastion_private_key_file = "~/.ssh/bastion_key"`
- **WHEN** the provisioner prepares for execution
- **THEN** the `~` SHALL be expanded to the user's HOME directory
- **AND** subsequent validation SHALL use the expanded absolute path

### Requirement: Mutual Exclusivity with Proxy Adapter

The remote ansible-navigator provisioner SHALL enforce mutual exclusivity between SSH tunnel mode and the Packer SSH proxy adapter mode.

#### Scenario: SSH tunnel mode and proxy adapter cannot both be enabled

- **GIVEN** a configuration with:
  - `ssh_tunnel_mode = true`
  - `use_proxy = true`
- **WHEN** the configuration is validated
- **THEN** validation SHALL fail
- **AND** the error message SHALL state "ssh_tunnel_mode and use_proxy are mutually exclusive"

#### Scenario: SSH tunnel mode with proxy disabled is valid

- **GIVEN** a configuration with:
  - `ssh_tunnel_mode = true`
  - `use_proxy = false` (or unset)
  - All required bastion fields provided
- **WHEN** the configuration is validated
- **THEN** validation SHALL succeed

#### Scenario: Proxy adapter mode with SSH tunnel disabled is valid

- **GIVEN** a configuration with:
  - `ssh_tunnel_mode = false` (or unset)
  - `use_proxy = true`
- **WHEN** the configuration is validated
- **THEN** validation SHALL succeed
- **AND** existing proxy adapter behavior SHALL remain active

### Requirement: HCL2 Schema Generation

The remote ansible-navigator provisioner SHALL include all SSH tunnel configuration fields in the HCL2 spec generation.

#### Scenario: New fields included in go:generate directive

- **GIVEN** the provisioner source code
- **WHEN** examining the `go:generate` directive at the top of the provisioner file
- **THEN** the directive SHALL include the Config type
- **AND** running `make generate` SHALL successfully generate `.hcl2spec.go` files

#### Scenario: Generated spec includes tunnel configuration fields

- **GIVEN** the generated `.hcl2spec.go` file
- **WHEN** examining the HCL2 spec
- **THEN** it SHALL include specifications for all tunnel configuration fields:
  - `ssh_tunnel_mode` (bool)
  - `bastion_host` (string)
  - `bastion_port` (number)
  - `bastion_user` (string)
  - `bastion_private_key_file` (string)
  - `bastion_password` (string)

#### Scenario: HCL parsing accepts tunnel configuration

- **GIVEN** a Packer HCL configuration with SSH tunnel fields specified
- **WHEN** Packer parses the configuration
- **THEN** parsing SHALL succeed
- **AND** all field values SHALL be correctly decoded into the Config struct

### Requirement: Sensitive Data Handling

The remote ansible-navigator provisioner SHALL treat `bastion_password` as sensitive data.

#### Scenario: Password is not logged in plaintext

- **GIVEN** a configuration with `bastion_password = "secret123"`
- **WHEN** the provisioner logs configuration or error messages
- **THEN** the password value SHALL NOT appear in plaintext in any log output
- **AND** sensitive data sanitization SHALL be applied

## Cross-References

This specification delta builds upon:

- [`remote-provisioner-capabilities`](../../../../openspec/specs/remote-provisioner-capabilities/spec.md) - Base provisioner requirements
- [`build-tooling`](../../../../openspec/specs/build-tooling/spec.md) - HCL2 spec generation requirements

Related changes:

- Prompt 02: `implement-ssh-tunnel-establishment` - Will use these configuration fields
- Prompt 03: `integrate-tunnel-with-inventory` - Will use these configuration fields
