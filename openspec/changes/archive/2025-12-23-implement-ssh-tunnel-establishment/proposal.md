# Proposal: Implement SSH Tunnel Establishment

## Context

This change implements the actual SSH tunnel establishment logic that was configured in change `add-ssh-tunnel-config-options`. When [`ssh_tunnel_mode = true`](../../provisioner/ansible-navigator/provisioner.go:535), the provisioner must create a local SSH port forward through the bastion host to the target machine, bypassing the Packer SSH proxy adapter.

## Problem Statement

The configuration schema for SSH tunnel mode exists, but the provisioner still lacks the runtime logic to:

1. Establish an SSH connection to the bastion host
2. Create a local port forward through the bastion to the target
3. Manage tunnel lifecycle (setup and cleanup)
4. Integrate the tunnel into the provisioning flow

Without this implementation, setting [`ssh_tunnel_mode = true`](../../provisioner/ansible-navigator/provisioner.go:535) will validate but fail at runtime when provisioning begins.

## Goals

Implement SSH tunnel establishment, lifecycle management, and error handling to enable direct SSH connectivity through a bastion host.

## Non-Goals

- Inventory generation changes (addressed in `03-integrate-tunnel-with-inventory.md`)
- Documentation updates (addressed in `04-update-documentation.md`)
- SSH agent forwarding support
- Connection persistence/reconnection logic
- Proxy protocol support (SOCKS5, HTTP CONNECT)

## Proposed Solution

### 1. Add [`setupSSHTunnel()`](../../provisioner/ansible-navigator/provisioner.go) function

Create a new function similar to existing [`setupAdapter()`](../../provisioner/ansible-navigator/provisioner.go:1039) that:

- Accepts bastion connection parameters from the config
- Accepts target host/port from `generatedData`
- Establishes SSH connection to bastion using [`golang.org/x/crypto/ssh`](../../provisioner/ansible-navigator/provisioner.go:39)
- Creates local port forward through bastion to target
- Returns local port number and [`io.Closer`] for cleanup

### 2. Integrate tunnel into [`Provision()`](../../provisioner/ansible-navigator/provisioner.go:1185) flow

Modify the provisioning flow to:

- Check [`p.config.SSHTunnelMode`](../../provisioner/ansible-navigator/provisioner.go:535)
- Call `setupSSHTunnel()` instead of `setupAdapter()` when tunnel mode is true
- Override `generatedData["Host"]` and `generatedData["Port"]` to use tunnel endpoint (`127.0.0.1:<localPort>`)
- Store tunnel port in [`p.config.LocalPort`](../../provisioner/ansible-navigator/provisioner.go:392) for inventory generation
- Defer tunnel cleanup via [`io.Closer.Close()`]

### 3. Port allocation strategy

Follow the same pattern as [`setupAdapter()`](../../provisioner/ansible-navigator/provisioner.go:1039):

- If [`p.config.LocalPort`](../../provisioner/ansible-navigator/provisioner.go:392) is specified, try up to 10 ports starting from that value
- Otherwise, let the system assign a port (bind to `:0`)
- Store allocated port back in [`p.config.LocalPort`](../../provisioner/ansible-navigator/provisioner.go:392)

### 4. Authentication methods

Support both key-based and password-based bastion authentication:

- **Key file**: Use [`ssh.ParsePrivateKey()`] on [`p.config.BastionPrivateKeyFile`](../../provisioner/ansible-navigator/provisioner.go:552)
- **Password**: Use [`ssh.Password()`] auth method with [`p.config.BastionPassword`](../../provisioner/ansible-navigator/provisioner.go:556)
- **Both**: Attempt key first, fall back to password

### 5. Error handling

Provide clear, actionable error messages for common failures:

- Bastion unreachable → `"Failed to connect to bastion host <host>:<port>: <error>"`
- Bastion auth failure → `"Failed to authenticate to bastion: <error>"`
- Invalid key file → `"Failed to parse bastion private key: <error>"`
- Target unreachable from bastion → `"Failed to establish tunnel to target <host>:<port>: <error>"`
- Port allocation failure → Try up to 10 ports, then `"Failed to allocate local port for tunnel"`

## Implementation Details

### SSH Tunnel Architecture

```
Packer Machine                 Bastion Host              Target Machine
    |                               |                          |
    |--SSH connect----------------->|                          |
    |  (bastion_user@bastion_host)  |                          |
    |                               |                          |
    |--Setup local forward--------->|                          |
    |  (127.0.0.1:N -> target:22)   |                          |
    |                               |                          |
    |                               |--SSH forward------------>|
    |                               |  (to target_host:22)     |
    |                               |                          |
    [Ansible connects to 127.0.0.1:N, traffic tunnels through bastion]
```

### Function Signature

```go
func (p *Provisioner) setupSSHTunnel(
    ui packersdk.Ui,
    targetHost string,
    targetPort int,
) (localPort int, tunnel io.Closer, err error)
```

### Integration Points

1. **[`Provision()`](../../provisioner/ansible-navigator/provisioner.go:1185)**: Choose between proxy adapter and SSH tunnel
2. **[`createInventoryFile()`](../../provisioner/ansible-navigator/provisioner.go:1123)**: Uses [`p.config.LocalPort`](../../provisioner/ansible-navigator/provisioner.go:392) (will contain tunnel port)
3. **Target credentials**: SSH tunnel only provides network path - Ansible still needs target SSH credentials from communicator or [`ansible_ssh_private_key_file`]

## Dependencies

- Change `add-ssh-tunnel-config-options` (already completed - schema exists)
- All bastion configuration fields must exist in [`Config`](../../provisioner/ansible-navigator/provisioner.go:320)

## Risks and Mitigations

| Risk | Mitigation |
|------|-----------|
| Tunnel cleanup on error | Use defer immediately after successful tunnel establishment |
| Port conflicts | Implement retry logic (up to 10 ports) |
| Auth method confusion | Test both key and password auth methods explicitly |
| Target unreachable from bastion | Provide clear error with target host/port |

## Validation

- Go code compiles (`go build ./...`)
- Existing tests pass (`go test ./...`)
- New `setupSSHTunnel()` function tested with mock SSH server (future work)
- Manual testing with real bastion and target hosts

## Open Questions

None - this is a straightforward implementation of a well-defined SSH tunneling pattern.
