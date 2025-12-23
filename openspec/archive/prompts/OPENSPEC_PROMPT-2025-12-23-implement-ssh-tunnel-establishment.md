# OpenSpec Prompt: Implement SSH Tunnel Establishment

## Context

Configuration schema for SSH tunnel mode has been added (prompt 01). Now we need to implement the actual SSH tunnel establishment logic that creates a local forwarding tunnel through a bastion host to the target machine.

When `ssh_tunnel_mode = true`, instead of using the Packer SSH proxy adapter, the provisioner should:

1. Establish an SSH connection to the bastion host
2. Create a local port forward from `127.0.0.1:<random_port>` through the bastion to `<target_host>:<target_port>`
3. Use this tunnel for Ansible connectivity

## Goal

Implement SSH tunnel establishment, lifecycle management, and error handling to enable direct SSH connectivity through a bastion host.

## Scope

### In Scope

1. **Add tunnel establishment function**:
   - Create `setupSSHTunnel()` function similar to existing [`setupAdapter()`](../../provisioner/ansible-navigator/provisioner.go:967)
   - Accept bastion config (host, port, user, auth credentials)
   - Accept target config (from generatedData: Host, Port)
   - Return local tunnel port and cleanup function

2. **SSH connection management**:
   - Parse bastion credentials (key file or password)
   - Establish SSH client connection to bastion
   - Configure local port forwarding through bastion to target
   - Handle connection errors with clear messages

3. **Integrate tunnel into provisioning flow**:
   - Modify [`Provision()`](../../provisioner/ansible-navigator/provisioner.go:1113) to choose between proxy adapter and SSH tunnel based on `ssh_tunnel_mode`
   - Store tunnel port for inventory generation
   - Ensure tunnel cleanup on exit (success or failure)

4. **Port allocation**:
   - Use dynamic port allocation (similar to [`setupAdapter()`](../../provisioner/ansible-navigator/provisioner.go:967) behavior)
   - Try up to 10 ports starting from `local_port` if specified, otherwise use system-assigned port
   - Store allocated port in `p.config.LocalPort` for inventory generation

### Out of Scope

- Inventory generation changes (handled in prompt 03)
- Documentation (handled in prompt 04)
- SSH agent forwarding support
- Connection persistence/reconnection logic (assume single-use tunnel per provisioning run)
- Proxy protocol support (SOCKS5, HTTP CONNECT)

## Desired Behavior

### SSH Tunnel Establishment Flow

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
// setupSSHTunnel establishes an SSH tunnel through a bastion host to the target.
// Returns the local port number and an error if tunnel establishment fails.
// The caller must defer tunnel.Close() to clean up the connection.
func (p *Provisioner) setupSSHTunnel(
    ui packersdk.Ui,
    targetHost string,
    targetPort int,
) (localPort int, tunnel io.Closer, err error)
```

### Error Scenarios

| Scenario | Expected Behavior |
|----------|-------------------|
| Bastion unreachable (network error) | Return error: "Failed to connect to bastion host <host>:<port>: <error>" |
| Bastion auth failure (wrong key/password) | Return error: "Failed to authenticate to bastion: <error>" |
| Bastion key file invalid format | Return error: "Failed to parse bastion private key: <error>" |
| Target unreachable from bastion | Return error: "Failed to establish tunnel to target <host>:<port>: <error>" |
| Local port allocation failure | Try up to 10 ports, then return error: "Failed to allocate local port for tunnel" |

### Integration with Provision()

```go
func (p *Provisioner) Provision(...) error {
    // ... existing setup ...
    
    privKeyFile := ""
    var tunnelCloser io.Closer
    
    if p.config.SSHTunnelMode {
        // SSH tunnel mode: establish tunnel through bastion
        ui.Say("Setting up SSH tunnel through bastion host...")
        
        targetHost := generatedData["Host"].(string)
        targetPort := generatedData["Port"].(int)
        
        localPort, tunnel, err := p.setupSSHTunnel(ui, targetHost, targetPort)
        if err != nil {
            return fmt.Errorf("failed to setup SSH tunnel: %w", err)
        }
        tunnelCloser = tunnel
        
        // Override connection details to use tunnel
        p.config.LocalPort = localPort
        generatedData["Host"] = "127.0.0.1"
        generatedData["Port"] = localPort
        
        // Use communicator SSH key for target auth
        privKeyFile = generatedData["SSHPrivateKeyFile"].(string)
        
        defer func() {
            ui.Say("Closing SSH tunnel...")
            tunnelCloser.Close()
        }()
        
    } else if !p.config.UseProxy.False() {
        // Existing proxy adapter path
        pkf, err := p.setupAdapterFunc(ui, comm)
        // ... existing proxy code ...
    }
    
    // ... rest of provisioning ...
}
```

## Constraints & Assumptions

1. **Use golang.org/x/crypto/ssh package**: Already imported at [line 39](../../provisioner/ansible-navigator/provisioner.go:39)
2. **Key parsing**: Use `ssh.ParsePrivateKey()` for reading `bastion_private_key_file`
3. **Password auth**: Use `ssh.Password()` auth method when `bastion_password` is specified
4. **Port forwarding**: Use `ssh.Client.Listen()` and `ssh.Client.Dial()` for local forwarding
5. **Error handling**: Wrap errors with context using `fmt.Errorf("context: %w", err)`
6. **Logging**: Use `ui.Say()` for user-facing messages, `log.Printf()` for debug output
7. **Target credentials**: SSH tunnel only provides network path - Ansible still needs target SSH credentials (from communicator or `ansible_ssh_private_key_file`)

## Acceptance Criteria

- [ ] `setupSSHTunnel()` function implemented with signature matching desired behavior
- [ ] Function successfully parses SSH private key from `bastion_private_key_file`
- [ ] Function successfully authenticates using `bastion_password` when key file not specified
- [ ] Function establishes SSH connection to bastion host
- [ ] Function creates local port forward through bastion to target
- [ ] Function returns allocated local port number
- [ ] Function returns io.Closer for cleanup
- [ ] [`Provision()`](../../provisioner/ansible-navigator/provisioner.go:1113) modified to call `setupSSHTunnel()` when `ssh_tunnel_mode = true`
- [ ] [`Provision()`](../../provisioner/ansible-navigator/provisioner.go:1113) skips proxy adapter setup when `ssh_tunnel_mode = true`
- [ ] Tunnel cleanup (Close) is deferred after successful tunnel establishment
- [ ] Connection errors include helpful context (bastion host, port, target)
- [ ] `go build ./...` compiles without errors
- [ ] `go test ./...` passes (no existing tests should break)

## Files Expected to Change

- [`provisioner/ansible-navigator/provisioner.go`](../../provisioner/ansible-navigator/provisioner.go) - Add setupSSHTunnel(), modify Provision()

## Dependencies

- Prompt `01-add-ssh-tunnel-config-options.md` must be completed
- Configuration schema for `ssh_tunnel_mode`, `bastion_*` fields must exist

## Next Steps After Completion

Proceed to prompt `03-integrate-tunnel-with-inventory.md` to ensure inventory generation uses tunnel connection details correctly.
