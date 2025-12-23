# Proposal: Integrate SSH Tunnel with Inventory Generation

## Overview

Validate and ensure that generated Ansible inventory files correctly use SSH tunnel connection details (`127.0.0.1:<tunnel_port>`) when `ssh_tunnel_mode = true`, allowing Ansible to connect through the established tunnel transparently. Add observability through debug logging to confirm tunnel integration.

## Background

SSH tunnel establishment (prompt 02) creates a local port forward at `127.0.0.1:<tunnel_port>` that connects through the bastion to the target machine. The tunnel setup updates `generatedData["Host"]` and `generatedData["Port"]` before inventory generation.

Inventory generation uses these `generatedData` values via template variables. This change validates the integration and adds debug logging for troubleshooting.

## Goals

1. **Validation**: Confirm that inventory generation correctly uses tunnel connection details when tunnel mode is active
2. **Observability**: Add debug logging to show when tunnel parameters are being used for inventory
3. **Edge cases**: Ensure custom inventory templates and proxy adapter paths continue to work correctly

## Proposed Changes

### 1. Debug Logging for Tunnel Mode

Add debug logging in [`Provision()`](../../../provisioner/ansible-navigator/provisioner.go) or [`createInventoryFile()`](../../../provisioner/ansible-navigator/provisioner.go) to show when SSH tunnel mode affects inventory generation.

**Log when tunnel mode active**:

```
[DEBUG] SSH tunnel mode active: inventory will use tunnel endpoint
[DEBUG] Tunnel connection: 127.0.0.1:<tunnel_port>
[DEBUG] Target user: <target_user>
[DEBUG] Target SSH key: <target_key_path>
```

Use existing `debugf()` helper (line 98 of provisioner.go) - debug output is controlled by `navigator_config.logging.level = "debug"`.

### 2. Validation Checks

No code changes expected for core inventory logic. The change validates:

- `generatedData["Host"]` contains `127.0.0.1` when tunnel active
- `generatedData["Port"]` contains tunnel port when tunnel active
- `p.config.User` is target machine user (not bastion user)
- Extra vars file contains `ansible_ssh_private_key_file` pointing to target key (not bastion key)

### 3. Edge Case Confirmation

Verify existing behaviors:

- **Custom inventory templates**: `inventory_file_template` uses `{{ .Host }}` and `{{ .Port }}` - automatically picks up tunnel values
- **Proxy adapter mode**: `use_proxy = true` path remains unchanged (mutually exclusive with tunnel mode)
- **Direct connection**: `use_proxy = false, ssh_tunnel_mode = false` path remains unchanged

## Non-Goals

- Modifying inventory template syntax
- Changing tunnel establishment logic (completed in prompt 02)
- Adding configuration options (completed in prompt 01)
- Documentation updates (prompt 04)
- Supporting multiple concurrent tunnels

## Capabilities

### SSH Tunnel Inventory Integration

When SSH tunnel mode is enabled, the generated inventory file uses the tunnel's local endpoint instead of the target's original address, allowing Ansible to connect transparently through the tunnel.

#### Requirements

- Inventory host address reflects tunnel endpoint (`127.0.0.1`)
- Inventory port reflects tunnel port (not standard SSH port 22)  
- User credentials reference target machine (not bastion)
- SSH key references target machine (not bastion)
- Debug logging shows tunnel connection details
- Custom inventory templates work correctly with tunnel
- No regression in proxy adapter or direct connection modes

## Testing Strategy

### Manual Testing

1. **SSH tunnel mode with EC2 builder**:
   - Configure bastion settings
   - Verify generated inventory contains `ansible_host=127.0.0.1`
   - Verify generated inventory contains `ansible_port=<tunnel_port>`
   - Verify `ansible_user=<ec2_user>` (not bastion user)

2. **Custom inventory template with tunnel**:
   - Use `inventory_file_template` with placeholders
   - Verify template renders with tunnel values

3. **Debug mode**:
   - Set `navigator_config.logging.level = "debug"`
   - Verify debug output shows tunnel connection details

### Regression Testing

- Proxy adapter mode: verify inventory still uses proxy adapter values
- Direct connection mode: verify inventory still uses direct connection values
- Build commands: `go build ./...`, `go test ./...`, `make plugin-check`

## Implementation Notes

### File Changes

- [`provisioner/ansible-navigator/provisioner.go`](../../../provisioner/ansible-navigator/provisioner.go) - Add debug logging only

### Dependencies

- Requires prompt 01 (SSH tunnel configuration schema)
- Requires prompt 02 (SSH tunnel establishment and generatedData updates)
- Blocks prompt 04 (documentation)

### Validation Commands

```bash
# Verify compilation
go build ./...

# Run tests
go test ./...

# Plugin conformance check  
make plugin-check

# Verify build
make build
```

## Success Criteria

- Debug logging shows tunnel parameters when `ssh_tunnel_mode = true` and debug mode enabled
- Generated inventory has `ansible_host=127.0.0.1` when tunnel active
- Generated inventory has `ansible_port=<tunnel_port>` when tunnel active
- User credentials reference target (not bastion)
- Custom inventory templates work with tunnel mode
- No regression in proxy adapter or direct connection modes
- All build and test commands pass

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Tunnel setup incomplete | Inventory uses wrong values | Validate prompt 02 implementation first |
| User vs bastion confusion | Authentication failures | Clear debug logging, validate User field |
| Template compatibility | Custom templates break | Validate with existing template test cases |

## Open Questions

None - requirements are clear from prompt file.
