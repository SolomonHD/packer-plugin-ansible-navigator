# Proposal: Add SSH Tunnel Configuration Options

## Overview

Add configuration schema and validation for SSH tunnel mode to the ansible-navigator provisioner. This enables direct SSH tunneling through a bastion host as an alternative to the Packer SSH proxy adapter, addressing connectivity issues in WSL2 and Docker execution environment scenarios.

## Problem Statement

The current ansible-navigator provisioner uses the Packer SSH proxy adapter to enable Ansible connectivity to targets. When running execution environments in WSL2 or Docker containers, container-to-host networking through this adapter is unreliable due to network isolation:

- Containers cannot reliably reach the proxy adapter listening on `127.0.0.1`
- The `host.containers.internal` workaround doesn't work consistently across environments
- Container networking introduces latency and complexity

## Proposed Solution

Introduce a new connectivity mode that bypasses the Packer adapter by establishing SSH tunnels directly through a bastion/jump host to reach the target. This change adds the configuration schema only—no runtime behavior changes.

### Configuration Fields

Add the following fields to the [`Config`](../../provisioner/ansible-navigator/provisioner.go) struct:

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `ssh_tunnel_mode` | `bool` | No | `false` | Enable SSH tunnel mode (mutually exclusive with proxy adapter) |
| `bastion_host` | `string` | Conditional | - | Bastion/jump host address (required when `ssh_tunnel_mode = true`) |
| `bastion_port` | `int` | No | `22` | Bastion SSH port |
| `bastion_user` | `string` | Conditional | - | SSH user for bastion (required when `ssh_tunnel_mode = true`) |
| `bastion_private_key_file` | `string` | Conditional | - | Path to SSH private key for bastion authentication |
| `bastion_password` | `string` | Conditional | - | Password for bastion (alternative to key-based auth) |

### Validation Rules

Enforce the following validation rules in [`Config.Validate()`](../../provisioner/ansible-navigator/provisioner.go):

1. **Mutual exclusivity**: `ssh_tunnel_mode` and `use_proxy` cannot both be `true`
2. **Required fields**: When `ssh_tunnel_mode = true`:
   - `bastion_host` must be non-empty
   - `bastion_user` must be non-empty
   - Either `bastion_private_key_file` OR `bastion_password` must be provided
3. **File existence**: If `bastion_private_key_file` is specified, the file must exist
4. **Port range**: `bastion_port` must be between 1 and 65535

### Path Handling

Apply HOME expansion (`~` and `~/path`) to `bastion_private_key_file` using the existing [`expandUserPath()`](../../provisioner/ansible-navigator/provisioner.go) function in [`Provisioner.Prepare()`](../../provisioner/ansible-navigator/provisioner.go).

## Scope

### In Scope

- Configuration field additions to [`Config`](../../provisioner/ansible-navigator/provisioner.go) struct
- Validation logic in [`Config.Validate()`](../../provisioner/ansible-navigator/provisioner.go)
- Path expansion in [`Provisioner.Prepare()`](../../provisioner/ansible-navigator/provisioner.go)
- HCL2 spec generation updates (via `make generate`)

### Out of Scope

- SSH tunnel establishment logic (covered in prompt 02)
- Inventory generation changes (covered in prompt 03)
- Documentation updates (covered in prompt 04)
- Automatic bastion discovery
- SSH agent forwarding support
- ProxyJump/ProxyCommand alternatives

## Impact

### User Experience

Users will be able to configure SSH tunnel mode in their Packer templates:

```hcl
provisioner "ansible-navigator" {
  ssh_tunnel_mode         = true
  bastion_host            = "bastion.example.com"
  bastion_user            = "deploy"
  bastion_private_key_file = "~/.ssh/bastion_key"
  
  play {
    target = "site.yml"
  }
}
```

### Backward Compatibility

- **Preserved**: Default behavior remains unchanged (proxy adapter active)
- **No breaking changes**: Existing configurations without `ssh_tunnel_mode` continue to work
- **New validation**: Configurations attempting to use both `ssh_tunnel_mode` and `use_proxy` will fail validation with a clear error message

## Dependencies

None. This is the first prompt in the SSH tunnel feature sequence.

## Next Steps

After this proposal is implemented:

1. Prompt 02: Implement SSH tunnel establishment logic
2. Prompt 03: Integrate tunnel with inventory generation
3. Prompt 04: Update documentation

## References

- [Prompt file](../../openspec/prompts/01-add-ssh-tunnel-config-options.md)
- [Config struct](../../provisioner/ansible-navigator/provisioner.go)
- [Existing path expansion function](../../provisioner/ansible-navigator/provisioner.go)
