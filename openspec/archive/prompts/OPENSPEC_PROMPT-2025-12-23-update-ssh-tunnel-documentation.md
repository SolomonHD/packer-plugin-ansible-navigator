# OpenSpec Prompt: Update Documentation for SSH Tunnel Mode

## Context

SSH tunnel mode functionality has been implemented (prompts 01-03). Users need comprehensive documentation to:

1. Understand when to use SSH tunnel mode vs proxy adapter
2. Configure bastion/jump host connectivity
3. Troubleshoot WSL2/Docker execution environment networking issues

The documentation must clearly explain the architecture differences and provide working examples.

## Goal

Add complete user-facing documentation for SSH tunnel mode including:

- Configuration reference
- Architecture comparison (tunnel vs proxy)
- Example configurations
- Troubleshooting guidance for WSL2/Docker scenarios

## Scope

### In Scope

1. **Update CONFIGURATION.md**:
   - Add new section: "SSH Tunnel Mode (Bastion/Jump Host)"
   - Document all `ssh_tunnel_mode` and `bastion_*` configuration fields
   - Provide complete working examples
   - Explain mutual exclusivity with `use_proxy`

2. **Add troubleshooting section**:
   - Create or update troubleshooting documentation
   - Document WSL2/Docker networking issues solved by tunnel mode
   - Provide diagnostic steps for tunnel connection failures
   - Include bastion connectivity verification commands

3. **Add architecture diagrams** (text/ASCII):
   - Show proxy adapter connection flow
   - Show SSH tunnel connection flow
   - Highlight differences for container-based execution environments

### Out of Scope

- Implementation details (already completed in prompts 01-03)
- Example playbooks/roles (focus on provisioner configuration)
- Bastion host setup guides (assume bastion already configured)
- Network security best practices
- Detailed Ansible configuration (beyond tunnel connectivity)

## Desired Behavior

### CONFIGURATION.md Updates

Add new section after existing SSH proxy documentation (~line 240):

#### Section: SSH Tunnel Mode (Bastion/Jump Host)

**Content structure:**

1. **When to Use SSH Tunnel Mode**
   - WSL2/Docker execution environment networking issues
   - Direct bastion access requirement
   - Scenarios where proxy adapter is unreliable

2. **Configuration Fields Table**

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `ssh_tunnel_mode` | bool | No | `false` | Enable SSH tunnel through bastion (mutually exclusive with `use_proxy = true`) |
| `bastion_host` | string | Yes (when tunnel mode enabled) | - | Bastion/jump host address or FQDN |
| `bastion_port` | int | No | `22` | SSH port on bastion host |
| `bastion_user` | string | Yes (when tunnel mode enabled) | - | SSH user for bastion authentication |
| `bastion_private_key_file` | string | Conditional* | - | Path to SSH private key for bastion auth |
| `bastion_password` | string | Conditional* | - | Password for bastion auth |

\* Either `bastion_private_key_file` or `bastion_password` is required when `ssh_tunnel_mode = true`

1. **Working Examples**

```hcl
# Example 1: SSH tunnel with key authentication (AWS via bastion)
source "amazon-ebs" "rhel" {
  ami_name      = "rhel-configured"
  instance_type = "t3.medium"
  region        = "us-east-1"
  
  source_ami_filter {
    filters = {
      name = "RHEL-9*"
    }
    owners      = ["309956199498"]
    most_recent = true
  }
  
  ssh_username = "ec2-user"
}

build {
  sources = ["source.amazon-ebs.rhel"]
  
  provisioner "ansible-navigator" {
    ssh_tunnel_mode         = true
    bastion_host            = "bastion.example.com"
    bastion_user            = "deploy"
    bastion_private_key_file = "~/.ssh/bastion_key"
    
    navigator_config {
      execution_environment {
        enabled = true
        image   = "quay.io/ansible/creator-ee:latest"
      }
    }
    
    play {
      target = "configure-rhel.yml"
    }
  }
}
```

```hcl
# Example 2: SSH tunnel with password auth (testing/lab environments)
provisioner "ansible-navigator" {
  ssh_tunnel_mode  = true
  bastion_host     = "10.0.1.50"
  bastion_port     = 2222
  bastion_user     = "jumpuser"
  bastion_password = "${var.bastion_password}"  # Use variable for security
  
  play {
    target = "test.yml"
  }
}
```

1. **Architecture Comparison**

```
Proxy Adapter Mode (Default):
┌──────────┐          ┌──────────────┐          ┌────────┐
│ Ansible  │─SSH(22)─→│ Packer Proxy │─Comm API→│ Target │
│ (local)  │          │ (127.0.0.1:N)│          │        │
└──────────┘          └──────────────┘          └────────┘
                      ↑ Port randomized
                      ↑ Works for direct connections

SSH Tunnel Mode:
┌──────────┐          ┌─────────┐          ┌────────┐
│ Ansible  │─SSH(N)──→│ Bastion │─SSH(22)─→│ Target │
│ (local)  │   tunnel │  Host   │          │        │
└──────────┘          └─────────┘          └────────┘
     ↑ Connects to 127.0.0.1:N
     ↑ Tunnel established through bastion
     ↑ Better for WSL2/container networking
```

1. **When SSH Tunnel Mode is Required**

SSH tunnel mode solves these specific problems:

- **WSL2 Execution Environments**: When ansible-navigator runs in a WSL2 container and the Packer proxy adapter binds to the Windows host, container-to-host networking can be unreliable due to network namespace isolation.

- **Docker on Windows**: Similar container-to-host networking issues when execution environments run in Docker Desktop.

- **Air-Gapped Target Access**: When the target is only accessible through a specific bastion/jump host and cannot be reached directly.

- **Security Policy Requirements**: When infrastructure policy requires all SSH connections to pass through a centralized bastion.

**Use proxy adapter (default) when:**

- Running ansible-navigator directly on the Packer host (no containers)
- Target has direct network connectivity from Packer host
- No bastion/jump host required

**Use SSH tunnel mode when:**

- Execution environments in WSL2/Docker show connection issues
- Target requires bastion access
- Proxy adapter networking is unstable

### Troubleshooting Documentation

Add new file or section: `docs/TROUBLESHOOTING.md` (or update existing if present)

#### Section: SSH Tunnel Connection Issues

**SSH Tunnel Fails to Establish:**

```bash
# Verify bastion connectivity
ssh -i ~/.ssh/bastion_key bastion_user@bastion_host

# Test tunnel manually
ssh -i ~/.ssh/bastion_key -L 127.0.0.1:9999:target_ip:22 bastion_user@bastion_host

# In another terminal, test local tunnel endpoint
ssh -i ~/.ssh/target_key -p 9999 target_user@127.0.0.1
```

**Common Error Messages:**

| Error | Cause | Solution |
|-------|-------|----------|
| "Failed to connect to bastion" | Network/DNS issue | Verify bastion_host is resolvable and reachable |
| "Failed to authenticate to bastion" | Wrong credentials | Check bastion_private_key_file path and permissions |
| "Failed to establish tunnel to target" | Bastion can't reach target | Verify bastion → target connectivity |
| "ssh_tunnel_mode and use_proxy are mutually exclusive" | Configuration conflict | Remove `use_proxy = true` or disable `ssh_tunnel_mode` |

**WSL2 Execution Environment Issues:**

If using execution environments in WSL2 and experiencing connectivity issues:

1. **Verify Execution Environment Mode:**

   ```hcl
   navigator_config {
     execution_environment {
       enabled = true
       # SSH tunnel mode recommended for WSL2 EE
     }
   }
   ```

2. **Enable SSH Tunnel Mode:**

   ```hcl
   ssh_tunnel_mode = true
   bastion_host    = "<your_bastion>"
   bastion_user    = "<bastion_user>"
   bastion_private_key_file = "~/.ssh/bastion_key"
   ```

3. **Disable Proxy Adapter (if previously set):**

   ```hcl
   # Remove or comment out:
   # use_proxy = true
   ```

**Key File Permissions:**

```bash
# Ensure bastion key has correct permissions
chmod 600 ~/.ssh/bastion_key

# Verify key format
ssh-keygen -l -f ~/.ssh/bastion_key
```

## Constraints & Assumptions

1. **Technical audience**: Assume users understand SSH concepts (keys, tunneling, bastions)
2. **Existing documentation structure**: Follow existing [`CONFIGURATION.md`](../../docs/CONFIGURATION.md) format and style
3. **Link consistency**: Use relative links to other documentation files
4. **Code block syntax**: Use `hcl` syntax highlighting for Packer configs
5. **Security emphasis**: Note that passwords should use variables, not hardcoded strings
6. **HOME expansion**: Document that `~` is expanded in file paths

## Acceptance Criteria

- [ ] [`docs/CONFIGURATION.md`](../../docs/CONFIGURATION.md) includes new "SSH Tunnel Mode" section with complete field reference
- [ ] Configuration examples show both key-based and password-based authentication
- [ ] Architecture comparison diagram clearly shows tunnel vs proxy differences
- [ ] "When to Use" guidance helps users choose between tunnel and proxy modes
- [ ] Troubleshooting documentation includes bastion connectivity verification steps
- [ ] Common error messages are documented with solutions
- [ ] WSL2/Docker-specific guidance is included
- [ ] All configuration fields have descriptions and type information
- [ ] Mutual exclusivity with `use_proxy` is clearly documented
- [ ] File paths, syntax, and links are correct
- [ ] Documentation follows existing style and formatting

## Files Expected to Change

- [`docs/CONFIGURATION.md`](../../docs/CONFIGURATION.md) - Add SSH tunnel mode configuration section
- `docs/TROUBLESHOOTING.md` - Add or update with SSH tunnel troubleshooting (create if doesn't exist)

Optional (if TROUBLESHOOTING.md doesn't exist, may add section to README):

- `README.md` - Add link to tunnel mode documentation

## Dependencies

- Prompt `01-add-ssh-tunnel-config-options.md` must be completed (configuration exists)
- Prompt `02-implement-ssh-tunnel-establishment.md` must be completed (feature works)
- Prompt `03-integrate-tunnel-with-inventory.md` must be completed (integration verified)

## Next Steps After Completion

SSH tunnel mode feature is complete. Consider:

- Publishing documentation updates
- Adding example configurations to documentation repository
- Creating troubleshooting decision tree
- Collecting user feedback on bastion connectivity patterns

## Review Checklist

- [ ] Examples are copy-paste ready and syntactically correct
- [ ] Technical terms are used consistently (bastion vs jump host)
- [ ] Security notes are appropriate (password variables, key permissions)
- [ ] Architecture diagrams render correctly in markdown viewers
- [ ] Links to related documentation work
- [ ] Configuration field table is complete and accurate
