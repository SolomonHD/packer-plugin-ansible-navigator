# ssh-tunnel-inventory-integration Specification Delta

## Purpose

Define how the ansible-navigator provisioner integrates SSH tunnel connection details into inventory generation, ensuring Ansible connects through established tunnels transparently.

## ADDED Requirements

### Requirement: Inventory Uses Tunnel Endpoint When Tunnel Mode Active

The remote ansible-navigator provisioner SHALL generate inventory files that reference the SSH tunnel's local endpoint (`127.0.0.1:<tunnel_port>`) when `ssh_tunnel_mode = true`.

#### Scenario: Generated inventory contains tunnel host

- **GIVEN** a configuration with `ssh_tunnel_mode = true`
- **AND** SSH tunnel has been successfully established
- **AND** `setupSSHTunnel()` has updated `generatedData["Host"] = "127.0.0.1"`
- **WHEN** `createInventoryFile()` generates the inventory
- **THEN** the generated inventory file SHALL contain `ansible_host=127.0.0.1`
- **AND** the value SHALL NOT be the target machine's original IP address

#### Scenario: Generated inventory contains tunnel port

- **GIVEN** a configuration with `ssh_tunnel_mode = true`
- **AND** SSH tunnel allocated local port 54321
- **AND** `setupSSHTunnel()` has updated `generatedData["Port"] = 54321`
- **WHEN** `createInventoryFile()` generates the inventory
- **THEN** the generated inventory file SHALL contain `ansible_port=54321`
- **AND** the port SHALL NOT be 22 (standard SSH port)
- **AND** the port SHALL match the tunnel's actual local port

#### Scenario: Inventory template variables use tunnel values

- **GIVEN** SSH tunnel mode is active
- **AND** `generatedData["Host"]` and `generatedData["Port"]` have been set to tunnel values
- **WHEN** inventory file is generated using default or custom template
- **THEN** template variables `{{ .Host }}` and `{{ .Port }}` SHALL expand to tunnel values
- **AND** no special template logic is needed (templates use generatedData automatically)

### Requirement: User Context References Target Machine

The remote ansible-navigator provisioner SHALL ensure inventory files reference the target machine's user credentials, not the bastion's user credentials.

#### Scenario: Inventory user is target machine user

- **GIVEN** a configuration with:
  - `ssh_tunnel_mode = true`
  - `bastion_user = "jump-user"`
  - Target machine SSH user (from communicator) is "ec2-user"
- **WHEN** inventory file is generated
- **THEN** the inventory SHALL contain `ansible_user=ec2-user`
- **AND** the inventory SHALL NOT contain `ansible_user=jump-user`
- **AND** `p.config.User` SHALL reference the target machine user

#### Scenario: SSH key references target machine key

- **GIVEN** a configuration with:
  - `ssh_tunnel_mode = true`
  - `bastion_private_key_file = "/path/to/bastion_key"`
  - Target machine SSH key (from communicator) at "/path/to/target_key"
- **WHEN** extra vars file is generated
- **THEN** extra vars SHALL contain `ansible_ssh_private_key_file=/path/to/target_key`
- **AND** extra vars SHALL NOT contain the bastion key path
- **AND** `generatedData["SSHPrivateKeyFile"]` SHALL reference target key

### Requirement: Custom Inventory Templates Work With Tunnel

The remote ansible-navigator provisioner SHALL support custom inventory templates (`inventory_file_template`) when SSH tunnel mode is enabled.

#### Scenario: Custom template uses tunnel values

- **GIVEN** a configuration with:

  ```hcl
  ssh_tunnel_mode = true
  inventory_file_template = "{{ .HostAlias }} ansible_host={{ .Host }} ansible_port={{ .Port }} ansible_user={{ .User }}\n"
  ```

- **AND** tunnel established with local port 12345
- **WHEN** inventory file is generated using the custom template
- **THEN** the template SHALL be rendered with:
  - `{{ .Host }}` = `127.0.0.1`
  - `{{ .Port }}` = `12345`
  - `{{ .User }}` = target machine user
- **AND** the rendered output SHALL be valid Ansible inventory format

#### Scenario: Template variables available

- **GIVEN** SSH tunnel mode active
- **WHEN** custom inventory template is processed
- **THEN** all standard template variables SHALL be available:
  - `.HostAlias`
  - `.Host` (tunnel endpoint: 127.0.0.1)
  - `.Port` (tunnel port)
  - `.User` (target user)
- **AND** template execution SHALL succeed without errors

### Requirement: Debug Logging for Tunnel Inventory Integration

The remote ansible-navigator provisioner SHALL emit debug-only diagnostic messages showing when SSH tunnel mode is affecting inventory generation.

#### Scenario: Debug messages emitted when tunnel mode active

- **GIVEN** a configuration with:
  - `ssh_tunnel_mode = true`
  - `navigator_config.logging.level = "debug"`
- **WHEN** inventory file is about to be generated
- **THEN** the provisioner SHALL emit debug messages via `debugf()` helper including:
  - "[DEBUG] SSH tunnel mode active: inventory will use tunnel endpoint"
  - "[DEBUG] Tunnel connection: 127.0.0.1:<tunnel_port>"
  - "[DEBUG] Target user: <target_user>"
  - "[DEBUG] Target SSH key: <target_key_path>"
- **AND** the messages SHALL be prefixed with `[DEBUG]`
- **AND** the messages SHALL use Packer UI output stream

#### Scenario: No debug output when debug mode disabled

- **GIVEN** a configuration with `ssh_tunnel_mode = true`
- **AND** `navigator_config.logging.level` is NOT set to "debug"
- **WHEN** provisioning executes
- **THEN** the SSH tunnel debug messages SHALL NOT appear in output
- **AND** normal operation messages SHALL still appear

#### Scenario: Debug output location

- **GIVEN** debug mode enabled
- **WHEN** debug messages are emitted
- **THEN** the provisioner SHALL use the existing `debugf()` helper function
- **AND** `debugf()` SHALL check `p.isDebugEnabled()` before emitting
- **AND** messages SHALL go to Packer UI via `ui.Message()`

### Requirement: No Regression in Proxy Adapter Mode

The remote ansible-navigator provisioner SHALL maintain existing proxy adapter behavior when SSH tunnel mode is disabled.

#### Scenario: Proxy adapter inventory unchanged

- **GIVEN** a configuration with:
  - `ssh_tunnel_mode = false` (or unset)
  - `use_proxy = true`
- **WHEN** inventory file is generated
- **THEN** the inventory SHALL use proxy adapter connection details
- **AND** `ansible_host` SHALL be `127.0.0.1` or configured `ansible_proxy_host`
- **AND** `ansible_port` SHALL be the proxy adapter port
- **AND** behavior SHALL be identical to pre-tunnel implementation

#### Scenario: Direct connection inventory unchanged

- **GIVEN** a configuration with:
  - `ssh_tunnel_mode = false` (or unset)
  - `use_proxy = false`
- **WHEN** inventory file is generated
- **THEN** the inventory SHALL use direct connection details
- **AND** `ansible_host` SHALL be the target machine's actual IP/hostname
- **AND** `ansible_port` SHALL be the target's SSH port (default 22)
- **AND** behavior SHALL be identical to pre-tunnel implementation

### Requirement: Inventory Generation Integration Point

The remote ansible-navigator provisioner's `Provision()` method SHALL ensure `generatedData` contains correct tunnel values before calling `createInventoryFile()`.

#### Scenario: generatedData updated before inventory creation

- **GIVEN** SSH tunnel mode is active
- **AND** `setupSSHTunnel()` has completed successfully
- **WHEN** `Provision()` prepares to create inventory
- **THEN** `generatedData["Host"]` SHALL be "127.0.0.1"
- **AND** `generatedData["Port"]` SHALL be the tunnel's local port
- **AND** these updates SHALL occur BEFORE `createInventoryFile()` is called
- **AND** `createInventoryFile()` SHALL read these values from `generatedData`

#### Scenario: createInventoryFile uses current generatedData

- **GIVEN** `createInventoryFile()` is called
- **WHEN** inventory templates are rendered
- **THEN** template data SHALL come from current `generatedData` map
- **AND** NO special tunnel-mode conditional logic SHALL be needed in `createInventoryFile()`
- **AND** the function SHALL work identically for tunnel, proxy, and direct modes

### Requirement: Validation Checks for Tunnel Integration

The remote ansible-navigator provisioner SHALL validate that tunnel mode configuration produces correct inventory without runtime errors.

#### Scenario: Inventory generation succeeds with tunnel

- **GIVEN** a valid configuration with `ssh_tunnel_mode = true`
- **AND** tunnel established successfully
- **WHEN** `createInventoryFile()` is called
- **THEN** inventory file creation SHALL succeed
- **AND** NO errors related to template rendering SHALL occur
- **AND** the generated file SHALL be valid Ansible inventory format

#### Scenario: Tunnel values are non-empty

- **GIVEN** SSH tunnel mode active
- **WHEN** validating `generatedData` before inventory generation
- **THEN** `generatedData["Host"]` SHALL NOT be empty
- **AND** `generatedData["Port"]` SHALL be a valid port number (1-65535)
- **AND** `generatedData["User"]` SHALL NOT be empty
- **AND** if any value is invalid, provisioning SHALL fail with clear error
