# connection-mode-configuration Specification

## Purpose
TBD - created by archiving change replace-connection-fields-with-enum. Update Purpose after archive.
## Requirements
### Requirement: Connection Mode Enum Configuration

The remote ansible-navigator provisioner SHALL support a `connection_mode` string field that explicitly defines how Ansible connections are established.

#### Scenario: Connection mode can be configured with explicit value

**Given:** A configuration for `provisioner "ansible-navigator"`  
**When:** The configuration includes `connection_mode = "proxy"`  
**Then:** Parsing SHALL succeed  
**And:** The provisioner SHALL store the connection mode as `"proxy"`

#### Scenario: Connection mode defaults to proxy when unspecified

**Given:** A configuration for `provisioner "ansible-navigator"`  
**And:** `connection_mode` is not specified  
**When:** The provisioner initializes  
**Then:** `connection_mode` SHALL default to `"proxy"`

#### Scenario: Connection mode validation rejects invalid values

**Given:** A configuration with `connection_mode = "invalid"`  
**When:** The configuration is validated  
**Then:** Validation SHALL fail  
**And:** The error message SHALL state valid options: ["proxy", "ssh_tunnel", "direct"]

### Requirement: Proxy Mode Selection

When `connection_mode = "proxy"`, the provisioner SHALL use Packer's SSH proxy adapter for Ansible connections.

#### Scenario: Proxy mode activates SSH proxy adapter

**Given:** A configuration with `connection_mode = "proxy"`  
**When:** [`Provision()`](../../../../../provisioner/ansible-navigator/provisioner.go:1351) is called  
**Then:** It SHALL call [`setupAdapter()`](../../../../../provisioner/ansible-navigator/provisioner.go:1039)  
**And:** It SHALL NOT call [`setupSSHTunnel()`](../../../../../provisioner/ansible-navigator/provisioner.go:1122)  
**And:** The proxy adapter SHALL be used for Ansible connections

#### Scenario: Proxy mode is the default behavior

**Given:** A configuration with no `connection_mode` specified  
**When:** [`Provision()`](../../../../../provisioner/ansible-navigator/provisioner.go:1351) is called  
**Then:** The behavior SHALL match `connection_mode = "proxy"`  
**And:** The proxy adapter SHALL be used

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

### Requirement: Direct Mode Selection

When `connection_mode = "direct"`, the provisioner SHALL connect directly to the target machine without using the proxy adapter or SSH tunnel.

#### Scenario: Direct mode bypasses proxy adapter

**Given:** A configuration with `connection_mode = "direct"`  
**When:** [`Provision()`](../../../../../provisioner/ansible-navigator/provisioner.go:1351) is called  
**Then:** It SHALL NOT call [`setupAdapter()`](../../../../../provisioner/ansible-navigator/provisioner.go:1039)  
**And:** It SHALL NOT call [`setupSSHTunnel()`](../../../../../provisioner/ansible-navigator/provisioner.go:1122)  
**And:** Ansible SHALL use the target's actual IP address and port from `generatedData`

#### Scenario: Direct mode requires valid target host

**Given:** A configuration with `connection_mode = "direct"`  
**And:** `generatedData["Host"]` is empty or missing  
**When:** [`Provision()`](../../../../../provisioner/ansible-navigator/provisioner.go:1351) is called  
**Then:** The provisioner SHALL log a warning  
**And:** It SHALL fallback to proxy mode  
**Or:** It SHALL fail with a clear error message

#### Scenario: Direct mode uses SSH keys from communicator

**Given:** A configuration with `connection_mode = "direct"`  
**And:** `generatedData["ConnType"]` is `"ssh"`  
**When:** Connection credentials are prepared  
**Then:** The provisioner SHALL use `generatedData["SSHPrivateKeyFile"]` or `generatedData["SSHPrivateKey"]`  
**And:** Ansible SHALL authenticate using the communicator's SSH credentials

### Requirement: Connection Mode Integration with Inventory

The connection mode SHALL affect how inventory files are generated to ensure Ansible connects to the correct endpoint.

#### Scenario: Proxy and tunnel modes use local endpoint in inventory

**Given:** A configuration with `connection_mode` set to `"proxy"` or `"ssh_tunnel"`  
**When:** The inventory file is generated  
**Then:** `ansible_host` SHALL be set to `AnsibleProxyHost` (defaults to `"127.0.0.1"`)  
**And:** `ansible_port` SHALL be set to the local port (proxy adapter port or tunnel local port)

#### Scenario: Direct mode uses target endpoint in inventory

**Given:** A configuration with `connection_mode = "direct"`  
**When:** The inventory file is generated  
**Then:** `ansible_host` SHALL be set to `generatedData["Host"]` (target's actual IP/hostname)  
**And:** `ansible_port` SHALL be set to `generatedData["Port"]` (target's actual SSH port)

### Requirement: Connection Configuration Validation

The provisioner SHALL validate `connection_mode` as an enum and conditionally validate related fields based on the selected mode.

**Note**: This replaces the previous validation which enforced mutual exclusivity between `use_proxy` and `ssh_tunnel_mode` boolean fields.

Validation SHALL enforce:

1. `connection_mode` is one of `["proxy", "ssh_tunnel", "direct"]`
2. When `connection_mode = "ssh_tunnel"`, all required bastion fields are provided and valid
3. Default value of `"proxy"` is applied when field is unset

#### Scenario: Invalid connection mode produces clear error

**Given:** A configuration with `connection_mode = "tunnel"` (typo)  
**When:** Validation runs  
**Then:** Validation SHALL fail  
**And:** Error message SHALL list valid values: ["proxy", "ssh_tunnel", "direct"]  
**And:** Error message SHALL show the invalid value provided

#### Scenario: Empty connection mode defaults to proxy

**Given:** A configuration with `connection_mode = ""`  
**When:** Validation runs  
**Then:** `connection_mode` SHALL be set to `"proxy"`  
**And:** Validation SHALL succeed

