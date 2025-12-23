# OpenSpec Prompt: Integrate SSH Tunnel with Inventory Generation

## Context

SSH tunnel establishment has been implemented (prompt 02). The tunnel creates a local port forward at `127.0.0.1:<tunnel_port>` that connects through the bastion to the target machine.

Now we need to ensure that the generated Ansible inventory file references the tunnel endpoint (`127.0.0.1:tunnel_port`) rather than the target machine's original address. This allows Ansible to connect through the tunnel transparently.

## Goal

Modify inventory generation logic to use tunnel connection details when SSH tunnel mode is active, ensuring Ansible connects through the established tunnel.

## Scope

### In Scope

1. **Inventory template data modification**:
   - When `ssh_tunnel_mode = true`, ensure [`createInventoryFile()`](../../provisioner/ansible-navigator/provisioner.go:1051) uses tunnel connection details
   - Verify that ctxData Host/Port values point to the tunnel (should already be set by Provision() after prompt 02)
   - Ensure `ansible_user` is set to the target machine's user, NOT the bastion user

2. **Connection parameter validation**:
   - Add debug logging to show inventory generation is using tunnel parameters
   - Verify SSH key reference points to target key, not bastion key

3. **Edge case handling**:
   - Ensure custom inventory templates (via `inventory_file_template`) work with tunnel mode
   - Verify `use_proxy.False()` conditions don't interfere with tunnel mode

### Out of Scope

- SSH tunnel establishment (completed in prompt 02)
- Configuration schema (completed in prompt 01)
- Documentation updates (prompt 04)
- Inventory template syntax changes (existing templates should work)
- Support for multiple tunnels (single target only)

## Desired Behavior

### Inventory Generation Flow with Tunnel

When `ssh_tunnel_mode = true`:

```
setupSSHTunnel() → LocalPort = 12345
generatedData["Host"] = "127.0.0.1"
generatedData["Port"] = 12345
createInventoryFile() → ansible_host=127.0.0.1 ansible_port=12345 ansible_user=<target_user>
```

### Expected Inventory Content

**With SSH tunnel mode enabled:**

```ini
default ansible_host=127.0.0.1 ansible_port=12345 ansible_user=ec2-user
```

Key points:

- `ansible_host` is `127.0.0.1` (tunnel endpoint, not target IP)
- `ansible_port` is the tunnel port (not 22)
- `ansible_user` is the target machine user (not bastion user)
- SSH key (via extra vars) points to target machine key

**Without SSH tunnel mode (existing behavior):**

```ini
default ansible_host=10.0.1.50 ansible_port=22 ansible_user=ec2-user
```

or with proxy adapter:

```ini
default ansible_host=127.0.0.1 ansible_port=54321 ansible_user=<packer_user>
```

### Debug Output

When plugin debug logging is enabled ([`navigator_config.logging.level = "debug"`](../../provisioner/ansible-navigator/provisioner.go:91)):

```
[DEBUG] SSH tunnel mode active: inventory will use tunnel endpoint
[DEBUG] Tunnel connection: 127.0.0.1:12345
[DEBUG] Target user: ec2-user
[DEBUG] Target SSH key: /path/to/target_key
```

## Constraints & Assumptions

1. **Minimal changes required**: Inventory generation should already work if `Provision()` correctly updates `generatedData["Host"]` and `generatedData["Port"]` (implemented in prompt 02)
2. **User context**: `p.config.User` must reference target machine user, not bastion user
3. **SSH key context**: Extra vars `ansible_ssh_private_key_file` must reference target key, not bastion key
4. **Template compatibility**: Existing inventory templates use `{{ .Host }}` and `{{ .Port }}` placeholders - these will automatically pick up tunnel values
5. **Debug output**: Use existing [`debugf()`](../../provisioner/ansible-navigator/provisioner.go:98) helper for debug logging
6. **Tunnel vs Proxy distinction**: SSH tunnel mode and proxy adapter are mutually exclusive (enforced by validation in prompt 01)

## Acceptance Criteria

- [ ] [`createInventoryFile()`](../../provisioner/ansible-navigator/provisioner.go:1051) generates correct inventory when `ssh_tunnel_mode = true`
- [ ] Generated inventory has `ansible_host=127.0.0.1` when tunnel is active
- [ ] Generated inventory has `ansible_port=<tunnel_port>` when tunnel is active
- [ ] Generated inventory has `ansible_user=<target_user>` when tunnel is active (NOT bastion user)
- [ ] Extra vars file contains `ansible_ssh_private_key_file=<target_key>` (NOT bastion key)
- [ ] Debug logging output shows tunnel connection details when debug mode enabled
- [ ] Custom inventory templates (via `inventory_file_template`) work correctly with tunnel
- [ ] No regression: proxy adapter mode still generates correct inventory
- [ ] No regression: direct connection mode (`use_proxy = false`) still generates correct inventory
- [ ] `go build ./...` compiles without errors
- [ ] `go test ./...` passes

## Files Expected to Change

- [`provisioner/ansible-navigator/provisioner.go`](../../provisioner/ansible-navigator/provisioner.go) - Add debug logging in createInventoryFile() or Provision()

Note: Actual inventory generation logic should NOT need changes if prompt 02 correctly updates `generatedData` before calling `createInventoryFile()`. This prompt primarily validates and adds observability.

## Dependencies

- Prompt `01-add-ssh-tunnel-config-options.md` must be completed
- Prompt `02-implement-ssh-tunnel-establishment.md` must be completed
- `setupSSHTunnel()` must correctly update `generatedData["Host"]` and `generatedData["Port"]`
- `setupSSHTunnel()` must NOT modify `p.config.User` (should remain as target user)

## Validation Strategy

Create test configurations:

1. **SSH tunnel mode with EC2 builder:**

   ```hcl
   source "amazon-ebs" "test" {
     # bastion configured in source
   }
   
   provisioner "ansible-navigator" {
     ssh_tunnel_mode         = true
     bastion_host            = "<bastion_ip>"
     bastion_user            = "jump-user"
     bastion_private_key_file = "~/.ssh/bastion_key"
     
     play { target = "test.yml" }
   }
   ```

   Verify inventory contains:
   - `ansible_host=127.0.0.1`
   - `ansible_port=<tunnel_port>`
   - `ansible_user=<ec2_user>` (from communicator, not "jump-user")

2. **Custom inventory template with tunnel:**

   ```hcl
   provisioner "ansible-navigator" {
     ssh_tunnel_mode = true
     bastion_host    = "bastion.example.com"
     bastion_user    = "bastion"
     bastion_private_key_file = "~/.ssh/bastion"
     
     inventory_file_template = "{{ .HostAlias }} ansible_host={{ .Host }} ansible_port={{ .Port }} ansible_user={{ .User }}\n"
     
     play { target = "test.yml" }
   }
   ```

   Verify template renders with tunnel values.

## Next Steps After Completion

Proceed to prompt `04-update-documentation.md` to document SSH tunnel mode configuration and troubleshooting.
