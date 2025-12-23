# OpenSpec Prompt: Add SSH Tunnel Configuration Options

## Context

The ansible-navigator provisioner currently uses a Packer SSH proxy adapter to enable Ansible connectivity. However, when running execution environments in WSL2 or Docker, the container-to-host networking through this adapter is unreliable due to network isolation.

The plugin needs a new connectivity mode that bypasses the Packer adapter entirely by establishing SSH tunnels directly through a bastion/jump host to reach the target.

## Goal

Add configuration schema and validation for SSH tunnel mode, including:

- Mode selection (proxy adapter vs SSH tunnel)
- Bastion host connection parameters
- SSH key configuration for bastion authentication

## Scope

### In Scope

1. Add new configuration fields to the [`Config`](../../provisioner/ansible-navigator/provisioner.go:317) struct:
   - `ssh_tunnel_mode` (bool) - Enable SSH tunnel mode (mutually exclusive with proxy adapter)
   - `bastion_host` (string) - Bastion/jump host address (required when `ssh_tunnel_mode = true`)
   - `bastion_port` (int) - Bastion SSH port (defaults to 22)
   - `bastion_user` (string) - SSH user for bastion (required when `ssh_tunnel_mode = true`)
   - `bastion_private_key_file` (string) - Path to SSH private key for bastion authentication
   - `bastion_password` (string) - Password for bastion (alternative to key-based auth)

2. Update [`Config.Validate()`](../../provisioner/ansible-navigator/provisioner.go:533) to enforce:
   - `ssh_tunnel_mode` and `use_proxy` are mutually exclusive
   - When `ssh_tunnel_mode = true`, require `bastion_host` and `bastion_user`
   - When `ssh_tunnel_mode = true`, require either `bastion_private_key_file` or `bastion_password`
   - Validate `bastion_private_key_file` exists if specified
   - Validate `bastion_port` is in valid range (1-65535)

3. Update the HCL2 spec generation directive at [line 6](../../provisioner/ansible-navigator/provisioner.go:6) to include new Config fields.

4. Run `make generate` to update `.hcl2spec.go` files after adding fields.

### Out of Scope

- SSH tunnel establishment logic (separate prompt)
- Inventory generation changes (separate prompt)  
- Documentation updates (separate prompt)
- Automatic bastion discovery
- Support for SSH agent forwarding (may be added later)
- ProxyJump/ProxyCommand alternatives (focus on explicit tunnel first)

## Desired Behavior

### Configuration Examples

**Basic SSH tunnel mode with key authentication:**

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

**SSH tunnel mode with password authentication:**

```hcl
provisioner "ansible-navigator" {
  ssh_tunnel_mode    = true
  bastion_host       = "10.0.1.100"
  bastion_port       = 2222
  bastion_user       = "jumpuser"
  bastion_password   = "${var.bastion_password}"
  
  play {
    target = "site.yml"
  }
}
```

**Invalid: SSH tunnel mode with proxy adapter (should fail validation):**

```hcl
provisioner "ansible-navigator" {
  ssh_tunnel_mode = true
  use_proxy       = true  # ERROR: mutually exclusive
  
  play {
    target = "site.yml"
  }
}
```

### Validation Behavior

| Configuration | Expected Result |
|---------------|-----------------|
| `ssh_tunnel_mode = true`, no `bastion_host` | Validation error: "bastion_host is required when ssh_tunnel_mode is true" |
| `ssh_tunnel_mode = true`, no `bastion_user` | Validation error: "bastion_user is required when ssh_tunnel_mode is true" |
| `ssh_tunnel_mode = true`, no key/password | Validation error: "either bastion_private_key_file or bastion_password is required" |
| `ssh_tunnel_mode = true`, `use_proxy = true` | Validation error: "ssh_tunnel_mode and use_proxy are mutually exclusive" |
| `bastion_port = 99999` | Validation error: "bastion_port must be between 1 and 65535" |
| `bastion_private_key_file = "/nonexistent"` | Validation error: "bastion_private_key_file: /nonexistent does not exist" |

## Constraints & Assumptions

1. **No behavior changes**: This prompt only adds configuration schema - no runtime behavior changes
2. **Preserve existing proxy adapter**: Default behavior remains unchanged (proxy adapter active)
3. **Path expansion**: Apply HOME expansion to `bastion_private_key_file` using existing [`expandUserPath()`](../../provisioner/ansible-navigator/provisioner.go:1986) function
4. **File validation**: Use existing [`validateFileConfig()`](../../provisioner/ansible-navigator/provisioner.go:1808) pattern for key file validation
5. **Sensitive data**: Treat `bastion_password` as sensitive (do not log in plaintext)

## Acceptance Criteria

- [ ] New configuration fields added to [`Config`](../../provisioner/ansible-navigator/provisioner.go:317) struct with proper mapstructure tags
- [ ] [`Config.Validate()`](../../provisioner/ansible-navigator/provisioner.go:533) enforces mutual exclusivity between `ssh_tunnel_mode` and `use_proxy`
- [ ] [`Config.Validate()`](../../provisioner/ansible-navigator/provisioner.go:533) requires `bastion_host`, `bastion_user`, and auth credentials when `ssh_tunnel_mode = true`
- [ ] [`Config.Validate()`](../../provisioner/ansible-navigator/provisioner.go:533) validates port range for `bastion_port`
- [ ] [`Config.Validate()`](../../provisioner/ansible-navigator/provisioner.go:533) validates existence of `bastion_private_key_file` if specified
- [ ] HOME expansion applied to `bastion_private_key_file` in [`Provisioner.Prepare()`](../../provisioner/ansible-navigator/provisioner.go:673)
- [ ] HCL2 spec generation includes all new fields (update [`go:generate`](../../provisioner/ansible-navigator/provisioner.go:6) directive)
- [ ] `make generate` runs successfully to update `.hcl2spec.go`
- [ ] `go build ./...` compiles without errors
- [ ] Configuration with `ssh_tunnel_mode = true` validates successfully when all required fields are present
- [ ] Configuration with `ssh_tunnel_mode = true` and `use_proxy = true` fails validation with clear error message

## Files Expected to Change

- [`provisioner/ansible-navigator/provisioner.go`](../../provisioner/ansible-navigator/provisioner.go) - Add fields to Config, update Validate() and Prepare()
- `provisioner/ansible-navigator/provisioner.hcl2spec.go` - Generated by `make generate`

## Dependencies

None - this is the first prompt in the sequence.

## Next Steps After Completion

Proceed to prompt `02-implement-ssh-tunnel-establishment.md` to implement the SSH tunnel connection logic.
