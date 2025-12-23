# ssh-tunnel-runtime Specification

## Purpose

Defines the runtime behavior for SSH tunnel establishment through bastion hosts, including connection setup, port forwarding, lifecycle management, and error handling.

## ADDED Requirements

### Requirement: SSH Tunnel Establishment

The remote ansible-navigator provisioner SHALL establish SSH tunnels through bastion hosts when tunnel mode is enabled.

#### Scenario: Tunnel setup function exists

- **GIVEN** the provisioner source code
- **WHEN** examining `provisioner.go`
- **THEN** a function `setupSSHTunnel()` SHALL exist
- **AND** it SHALL accept parameters for UI, target host, and target port
- **AND** it SHALL return local port number, io.Closer, and error

#### Scenario: Tunnel established using bastion credentials

- **GIVEN** a configuration with valid bastion credentials (key file or password)
- **WHEN** `setupSSHTunnel()` is called
- **THEN** it SHALL establish an SSH connection to the bastion host
- **AND** it SHALL authenticate using the provided credentials (key or password)

#### Scenario: Local port forward created through bastion

- **GIVEN** an established SSH connection to the bastion
- **AND** valid target host and port
- **WHEN** `setupSSHTunnel()` creates the tunnel
- **THEN** it SHALL create a local TCP listener on 127.0.0.1
- **AND** it SHALL forward connections through the bastion to the target

#### Scenario: Allocated port number returned

- **GIVEN** a successful tunnel setup
- **WHEN** `setupSSHTunnel()` returns
- **THEN** the return value SHALL include the actual local port number allocated
- **AND** the port number SHALL be between 1 and 65535

#### Scenario: Cleanup handle returned for tunnel

- **GIVEN** a successful tunnel setup
- **WHEN** `setupSSHTunnel()` returns
- **THEN** the return value SHALL include an io.Closer
- **AND** calling Close() on the io.Closer SHALL terminate the SSH connection
- **AND** calling Close() SHALL release the local port

### Requirement: Authentication Method Support

The provisioner SHALL support both key-based and password-based authentication to bastion hosts.

#### Scenario: Key file authentication succeeds

- **GIVEN** a configuration with `bastion_private_key_file` pointing to a valid SSH private key
- **AND** `bastion_password` is empty
- **WHEN** `setupSSHTunnel()` is called
- **THEN** it SHALL read and parse the private key file using ssh.ParsePrivateKey()
- **AND** it SHALL use the key for authentication to the bastion
- **AND** authentication SHALL succeed if the key is accepted by the bastion

#### Scenario: Password authentication succeeds

- **GIVEN** a configuration with `bastion_password` set
- **AND** `bastion_private_key_file` is empty
- **WHEN** `setupSSHTunnel()` is called
- **THEN** it SHALL use ssh.Password() auth method
- **AND** it SHALL authenticate using the provided password
- **AND** authentication SHALL succeed if the password is correct

#### Scenario: Both key and password provided (key preferred)

- **GIVEN** a configuration with both `bastion_private_key_file` and `bastion_password`
- **WHEN** `setupSSHTunnel()` is called
- **THEN** it SHALL attempt key authentication first
- **AND** if key authentication fails, it SHALL attempt password authentication
- **AND** authentication SHALL succeed if either method is accepted

### Requirement: Port Allocation

The provisioner SHALL implement dynamic port allocation for local tunnel endpoints.

#### Scenario: User-specified port successful

- **GIVEN** `local_port = 5555` in configuration
- **AND** port 5555 is available on 127.0.0.1
- **WHEN** `setupSSHTunnel()` is called
- **THEN** it SHALL bind to 127.0.0.1:5555
- **AND** the returned local port SHALL be 5555

#### Scenario: User-specified port occupied, retry succeeds

- **GIVEN** `local_port = 5555` in configuration
- **AND** port 5555 is occupied
- **AND** port 5556 is available
- **WHEN** `setupSSHTunnel()` is called
- **THEN** it SHALL try port 5555 first
- **AND** it SHALL retry with port 5556
- **AND** the returned local port SHALL be 5556

#### Scenario: System-assigned port when no port specified

- **GIVEN** `local_port` is not specified (0 or unset)
- **WHEN** `setupSSHTunnel()` is called
- **THEN** it SHALL bind to 127.0.0.1:0 (system-assigned port)
- **AND** the returned local port SHALL be the system-assigned port number

#### Scenario: Port allocation failure after retries

- **GIVEN** `local_port = 5555`
- **AND** all 10 ports from 5555 to 5564 are occupied
- **WHEN** `setupSSHTunnel()` is called
- **THEN** it SHALL try ports 5555 through 5564
- **AND** it SHALL return an error "Failed to allocate local port for tunnel"

### Requirement: Integration with Provision Flow

The provisioner SHALL integrate SSH tunnel setup into the provisioning lifecycle.

#### Scenario: Tunnel replaces proxy adapter when enabled

- **GIVEN** a configuration with `ssh_tunnel_mode = true`
- **WHEN** Provision() is called
- **THEN** it SHALL NOT call setupAdapter()
- **AND** it SHALL call setupSSHTunnel() instead
- **AND** generatedData["Host"] SHALL be overridden to "127.0.0.1"
- **AND** generatedData["Port"] SHALL be overridden to the tunnel's local port

#### Scenario: Config local port updated after tunnel setup

- **GIVEN** successful tunnel establishment
- **AND** the tunnel allocated local port 54321
- **WHEN** `setupSSHTunnel()` returns
- **THEN** `p.config.LocalPort` SHALL be set to 54321
- **AND** subsequent inventory generation SHALL use port 54321

#### Scenario: Tunnel cleanup on provisioning success

- **GIVEN** a successfully established tunnel
- **AND** provisioning completes successfully
- **WHEN** Provision() returns
- **THEN** the tunnel's Close() method SHALL be called
- **AND** the SSH connection SHALL be terminated
- **AND** the local port SHALL be released

#### Scenario: Tunnel cleanup on provisioning failure

- **GIVEN** a successfully established tunnel
- **AND** provisioning fails with an error
- **WHEN** Provision() returns with error
- **THEN** the tunnel's Close() method SHALL still be called (via defer)
- **AND** the SSH connection SHALL be terminated
- **AND** the local port SHALL be released

### Requirement: Error Handling and Messages

The provisioner SHALL provide clear, actionable error messages for tunnel establishment failures.

#### Scenario: Bastion connection failure

- **GIVEN** `bastion_host = "unreachable.example.com"`
- **WHEN** `setupSSHTunnel()` attempts to connect
- **AND** connection fails with network error
- **THEN** it SHALL return an error containing "Failed to connect to bastion host unreachable.example.com:<port>"
- **AND** the error SHALL include the underlying network error details

#### Scenario: Bastion authentication failure

- **GIVEN** valid bastion host and port
- **AND** incorrect bastion credentials
- **WHEN** `setupSSHTunnel()` attempts to authenticate
- **AND** authentication fails
- **THEN** it SHALL return an error containing "Failed to authenticate to bastion"
- **AND** the error SHALL include authentication method details

#### Scenario: Invalid private key file format

- **GIVEN** `bastion_private_key_file` pointing to a file with invalid SSH key format
- **WHEN** `setupSSHTunnel()` attempts to parse the key
- **AND** ssh.ParsePrivateKey() fails
- **THEN** it SHALL return an error containing "Failed to parse bastion private key"
- **AND** the error SHALL include the parsing error details

#### Scenario: Target unreachable from bastion

- **GIVEN** successful bastion connection
- **AND** target host "10.1.2.3" is unreachable from bastion
- **WHEN** `setupSSHTunnel()` attempts to create port forward
- **AND** forward setup fails
- **THEN** it SHALL return an error containing "Failed to establish tunnel to target 10.1.2.3:<port>"
- **AND** the error SHALL include cause details

### Requirement: Target Credential Independence

SSH tunnel establishment SHALL be independent of target machine SSH credentials.

#### Scenario: Tunnel uses communicator credentials for target

- **GIVEN** SSH tunnel mode enabled
- **AND** tunnel successfully established
- **WHEN** Ansible connects through the tunnel
- **THEN** Ansible SHALL use target credentials from:
  - generatedData["SSHPrivateKeyFile"], OR
  - ansible_ssh_private_key_file extra var, OR
  - SSH agent
- **AND** bastion credentials SHALL NOT be used for target authentication

#### Scenario: Tunnel provides network path only

- **GIVEN** a working SSH tunnel through bastion
- **WHEN** Ansible attempts to connect to the target
- **THEN** network connectivity SHALL be provided via 127.0.0.1:<tunnel_port>
- **AND** target SSH authentication SHALL be handled separately by Ansible
- **AND** target authentication failures SHALL NOT be attributed to tunnel setup

### Requirement: Logging and Diagnostics

The provisioner SHALL provide diagnostic logging for tunnel operations.

#### Scenario: UI message on tunnel setup start

- **GIVEN** `ssh_tunnel_mode = true`
- **WHEN** Provision() begins tunnel setup
- **THEN** it SHALL call ui.Say() with message "Setting up SSH tunnel through bastion host..."

#### Scenario: UI message on tunnel setup success

- **GIVEN** successful tunnel establishment
- **AND** local port 54321 allocated
- **WHEN** `setupSSHTunnel()` returns successfully
- **THEN** it SHALL log an appropriate success message via ui.Say()

#### Scenario: UI message on tunnel cleanup

- **GIVEN** provisioning completes (success or failure)
- **WHEN** tunnel cleanup begins
- **THEN** it SHALL call ui.Say() with message "Closing SSH tunnel..."

#### Scenario: Error details in UI output

- **GIVEN** tunnel setup fails
- **WHEN** error is returned from `setupSSHTunnel()`
- **THEN** Provision() SHALL format the error with context
- **AND** it SHALL return error containing "failed to setup SSH tunnel: <original error>"
