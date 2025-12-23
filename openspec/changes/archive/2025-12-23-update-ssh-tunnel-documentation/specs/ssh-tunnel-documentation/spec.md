# SSH Tunnel Documentation Specification

## ADDED Requirements

### Requirement: SSH Tunnel Configuration Reference

[`docs/CONFIGURATION.md`](../../../../../../docs/CONFIGURATION.md) SHALL include comprehensive documentation for SSH tunnel mode configuration with complete field reference, examples, and guidance for choosing between tunnel and proxy modes.

#### Scenario: SSH tunnel section exists with complete field reference

**Given:** A user consults [`docs/CONFIGURATION.md`](../../../../../../docs/CONFIGURATION.md) to configure SSH tunnel mode  
**When:** They navigate to SSH connection configuration  
**Then:** A section titled "SSH Tunnel Mode (Bastion/Jump Host)" SHALL exist after SSH proxy documentation  
**And:** The section SHALL include a table documenting all SSH tunnel configuration fields:

- `ssh_tunnel_mode` (bool, default `false`) - Enable SSH tunnel through bastion (mutually exclusive with `use_proxy = true`)
- `bastion_host` (string, required when tunnel mode enabled) - Bastion/jump host address or FQDN
- `bastion_port` (int, default `22`) - SSH port on bastion host
- `bastion_user` (string, required when tunnel mode enabled) - SSH user for bastion authentication
- `bastion_private_key_file` (string, conditional) - Path to SSH private key for bastion auth
- `bastion_password` (string, conditional) - Password for bastion auth  
**And:** The documentation SHALL state that either `bastion_private_key_file` or `bastion_password` is required when `ssh_tunnel_mode = true`

#### Scenario: When to use guidance helps users choose connection mode

**Given:** A user is deciding between SSH tunnel mode and proxy adapter mode  
**When:** They read the SSH tunnel documentation  
**Then:** The documentation SHALL include a "When to Use SSH Tunnel Mode" section explaining:

- WSL2/Docker execution environment networking issues (container-to-host networking unreliability)
- Direct bastion access requirement (air-gapped targets)
- Security policy requirements (centralized bastion)
- Proxy adapter instability scenarios  
**And:** It SHALL recommend proxy adapter (default) when:
- Running ansible-navigator directly on Packer host (no containers)
- Target has direct network connectivity
- No bastion/jump host required  
**And:** It SHALL recommend SSH tunnel mode when:
- Execution environments in WSL2/Docker show connection issues
- Target requires bastion access
- Proxy adapter networking is unstable

#### Scenario: Architecture comparison diagrams show connection flows

**Given:** A user needs to understand architectural differences  
**When:** They read the SSH tunnel documentation  
**Then:** The documentation SHALL include ASCII diagrams showing:

- **Proxy Adapter Mode**: Ansible → Packer Proxy (127.0.0.1:N) → Target
  - Port randomized, works for direct connections
- **SSH Tunnel Mode**: Ansible → Bastion Host (tunnel) → Target
  - Connects to 127.0.0.1:N, tunnel established through bastion
  - Better for WSL2/container networking  
**And:** The diagrams SHALL clearly highlight container networking differences

#### Scenario: Working examples for key-based authentication

**Given:** A user needs to configure SSH tunnel with key authentication  
**When:** They consult the examples  
**Then:** The documentation SHALL include a complete, copy-pasteable example showing:

- AWS EC2 builder with SSH tunnel configuration
- `ssh_tunnel_mode = true`
- `bastion_host`, `bastion_user`, `bastion_private_key_file` configuration
- Execution environment enabled with container image
- `play { target = "..." }` block  
**And:** The example SHALL be syntactically correct HCL

#### Scenario: Working examples for password-based authentication

**Given:** A user needs to configure SSH tunnel with password authentication (lab environments)  
**When:** They consult the examples  
**Then:** The documentation SHALL include an example showing:

- `ssh_tunnel_mode = true`
- `bastion_host`, `bastion_port`, `bastion_user` configuration
- `bastion_password` using variable reference (e.g., `"${var.bastion_password}"`)  
**And:** The example SHALL include a security note recommending variables over hardcoded passwords

#### Scenario: Mutual exclusivity with use_proxy documented

**Given:** A user has existing `use_proxy = true` configuration  
**When:** They add SSH tunnel mode configuration  
**Then:** The documentation SHALL clearly state that `ssh_tunnel_mode` and `use_proxy` are mutually exclusive  
**And:** It SHALL explain that only one connection mode can be active at a time

#### Scenario: HOME expansion documented for file paths

**Given:** A user configures `bastion_private_key_file`  
**When:** They use `~` in the file path  
**Then:** The documentation SHALL note that `~` is expanded to the user's HOME directory  
**And:** Examples SHALL demonstrate using `~/.ssh/bastion_key` paths

---

### Requirement: SSH Tunnel Troubleshooting Guide

[`docs/TROUBLESHOOTING.md`](../../../../../../docs/TROUBLESHOOTING.md) SHALL include a dedicated section for diagnosing and resolving SSH tunnel connection failures with common error patterns and manual verification procedures.

#### Scenario: SSH tunnel troubleshooting section exists

**Given:** A user encounters SSH tunnel connection failures  
**When:** They consult [`docs/TROUBLESHOOTING.md`](../../../../../../docs/TROUBLESHOOTING.md)  
**Then:** A section titled "SSH Tunnel Connection Issues" SHALL exist after the "Authentication failures" section

#### Scenario: Bastion connectivity verification documented

**Given:** A user's SSH tunnel fails to establish  
**When:** They need to verify bastion connectivity  
**Then:** The documentation SHALL provide manual verification commands:

```bash
# Verify bastion connectivity
ssh -i ~/.ssh/bastion_key bastion_user@bastion_host

# Test tunnel manually
ssh -i ~/.ssh/bastion_key -L 127.0.0.1:9999:target_ip:22 bastion_user@bastion_host

# In another terminal, test local tunnel endpoint
ssh -i ~/.ssh/target_key -p 9999 target_user@127.0.0.1
```

#### Scenario: Common error messages table documents failures

**Given:** A user encounters a specific SSH tunnel error  
**When:** They search the troubleshooting documentation  
**Then:** A table SHALL document common error patterns with causes and solutions:

- "Failed to connect to bastion" → Verify bastion_host is resolvable and reachable (DNS/network issue)
- "Failed to authenticate to bastion" → Check bastion_private_key_file path and permissions (wrong credentials)
- "Failed to establish tunnel to target" → Verify bastion → target connectivity
- "ssh_tunnel_mode and use_proxy are mutually exclusive" → Remove `use_proxy = true` or disable `ssh_tunnel_mode` (configuration conflict)

#### Scenario: WSL2 execution environment resolution steps documented

**Given:** A user uses execution environments in WSL2 and experiences connectivity issues  
**When:** They consult the troubleshooting guide  
**Then:** The documentation SHALL provide step-by-step resolution:

1. **Verify Execution Environment Mode** - Show example confirming `execution_environment.enabled = true`
2. **Enable SSH Tunnel Mode** - Show example with `ssh_tunnel_mode = true` and bastion configuration
3. **Disable Proxy Adapter** - Instruct removing or commenting out `use_proxy = true`  
**And:** Each step SHALL include HCL configuration examples

#### Scenario: Key file permissions documented

**Given:** A user's bastion key authentication fails  
**When:** They check troubleshooting guidance  
**Then:** The documentation SHALL explain key file permission requirements:

```bash
# Ensure bastion key has correct permissions
chmod 600 ~/.ssh/bastion_key

# Verify key format
ssh-keygen -l -f ~/.ssh/bastion_key
```

---

### Requirement: WSL2/Docker Execution Environment Scenarios

The documentation SHALL specifically address WSL2 and Docker execution environment networking scenarios where SSH tunnel mode resolves container-to-host connectivity issues.

#### Scenario: WSL2 networking issues explained

**Given:** A user runs ansible-navigator in a WSL2 container  
**When:** The Packer proxy adapter binds to the Windows host  
**Then:** The documentation SHALL explain:

- Container-to-host networking can be unreliable due to network namespace isolation
- Proxy adapter binding to Windows host may not be reachable from WSL2 container
- SSH tunnel mode resolves this by establishing the tunnel from inside the execution environment

#### Scenario: Docker on Windows networking issues explained

**Given:** A user uses Docker Desktop on Windows with execution environments  
**When:** They experience connection failures with proxy adapter  
**Then:** The documentation SHALL explain:

- Similar container-to-host networking issues as WSL2
- Execution environments running in Docker have network namespace separation
- SSH tunnel mode provides reliable connectivity through bastion

#### Scenario: Proxy adapter recommended for non-container scenarios

**Given:** A user runs ansible-navigator directly on the Packer host (no execution environment)  
**When:** They read the documentation  
**Then:** It SHALL clearly state proxy adapter is the recommended default for:

- Direct execution (no containers)
- Targets with direct network connectivity
- When no bastion is required

#### Scenario: Container-based scenarios clearly identified

**Given:** A user has execution environments enabled  
**When:** They experience connection issues  
**Then:** The documentation SHALL identify this as a container-based execution scenario  
**And:** It SHALL recommend trying SSH tunnel mode if connectivity is problematic  
**And:** Examples SHALL show `navigator_config.execution_environment.enabled = true` in tunnel mode configurations

---

### Requirement: Security Best Practices

The documentation SHALL emphasize secure credential handling for SSH tunnel configuration.

#### Scenario: Password variables recommended over hardcoded values

**Given:** A user configures `bastion_password`  
**When:** They view the password authentication example  
**Then:** The example SHALL use a variable reference: `bastion_password = "${var.bastion_password}"`  
**And:** It SHALL include a comment: `# Use variable for security`  
**And:** It SHALL NOT show hardcoded password strings

#### Scenario: Key file permissions emphasized

**Given:** A user configures `bastion_private_key_file`  
**When:** They read the troubleshooting documentation  
**Then:** It SHALL explain that SSH keys require restrictive permissions (600)  
**And:** It SHALL provide commands to set correct permissions

#### Scenario: Credential storage warnings

**Given:** A user reads authentication configuration  
**When:** They consider how to store credentials  
**Then:** The documentation SHALL warn against:

- Committing credentials to version control
- Hardcoding passwords in templates  
**And:** It SHALL recommend:
- Using Packer variables with secret values
- Using CI/CD secret variables
- Using SSH keys with proper permissions
