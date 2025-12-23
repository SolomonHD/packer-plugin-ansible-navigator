# Proposal: Update SSH Tunnel Documentation

## Overview

Add comprehensive user-facing documentation for SSH tunnel mode, completing the documentation requirements for the SSH tunnel feature implemented in changes `add-ssh-tunnel-config-options`, `implement-ssh-tunnel-establishment`, and `integrate-tunnel-with-inventory`.

## Problem Statement

SSH tunnel functionality (prompts 01-03) has been implemented but lacks user-facing documentation. Users need:

1. **Configuration reference** - Complete documentation of all SSH tunnel configuration fields
2. **Architecture comparison** - Clear explanation of when to use SSH tunnel mode vs proxy adapter mode
3. **Working examples** - Copy-pasteable configuration examples showing both key-based and password-based authentication
4. **Troubleshooting guidance** - Solutions for common tunnel connection failures, especially for WSL2/Docker execution environment scenarios

Without this documentation, users cannot effectively:

- Determine whether SSH tunnel mode solves their networking issues
- Configure bastion/jump host connectivity correctly
- Diagnose and resolve tunnel connection failures
- Understand the architectural differences between proxy and tunnel modes

## Goals

1. Add complete SSH tunnel configuration reference to [`docs/CONFIGURATION.md`](../../../docs/CONFIGURATION.md)
2. Document troubleshooting procedures for SSH tunnel connection failures in [`docs/TROUBLESHOOTING.md`](../../../docs/TROUBLESHOOTING.md)
3. Provide clear architectural comparison between proxy adapter and SSH tunnel modes
4. Include working examples for common use cases (AWS via bastion, lab environments)
5. Document WSL2/Docker-specific networking scenarios where tunnel mode is required

## Non-Goals

- Implementation details (feature is already complete)
- Example playbooks/roles (focus is on provisioner configuration)
- Bastion host setup guides (assume bastion is already configured)
- Network security best practices
- Detailed Ansible configuration beyond tunnel connectivity

## Proposed Changes

### 1. CONFIGURATION.md Updates

Add new section "SSH Tunnel Mode (Bastion/Jump Host)" after existing SSH proxy documentation.

**Content structure:**

1. **When to Use SSH Tunnel Mode** - Decision criteria for choosing tunnel vs proxy
2. **Configuration Fields Table** - Complete reference for all `ssh_tunnel_mode` and `bastion_*` fields
3. **Working Examples** - Copy-pasteable configurations for:
   - AWS EC2 via bastion with key authentication
   - Lab environment with password authentication
4. **Architecture Comparison** - ASCII diagrams showing:
   - Proxy adapter connection flow
   - SSH tunnel connection flow
   - Differences for container-based execution environments
5. **When SSH Tunnel Mode is Required** - Specific scenarios:
   - WSL2 execution environments with container-to-host networking issues
   - Docker on Windows networking unreliability
   - Air-gapped target access requiring bastion
   - Security policy requiring centralized bastion

### 2. TROUBLESHOOTING.md Updates

Add new section "SSH Tunnel Connection Issues" after existing authentication failures section.

**Content structure:**

1. **SSH Tunnel Fails to Establish** - Manual verification commands for:
   - Bastion connectivity testing
   - Manual tunnel establishment
   - Local tunnel endpoint verification
2. **Common Error Messages Table** - Error patterns with causes and solutions:
   - "Failed to connect to bastion" → Network/DNS verification
   - "Failed to authenticate to bastion" → Credential verification
   - "Failed to establish tunnel to target" → Bastion→target connectivity
   - Mutual exclusivity with `use_proxy` configuration conflict
3. **WSL2 Execution Environment Issues** - Step-by-step resolution:
   - Verify execution environment mode
   - Enable SSH tunnel mode
   - Disable proxy adapter if previously set
4. **Key File Permissions** - Permission requirements and verification commands

## Dependencies

This change depends on completed implementation from:

- `add-ssh-tunnel-config-options` - Configuration fields exist
- `implement-ssh-tunnel-establishment` - Tunnel feature works
- `integrate-tunnel-with-inventory` - Integration verified

## Constraints

- **Technical audience**: Assume users understand SSH concepts (keys, tunneling, bastions)
- **Existing documentation style**: Follow current [`CONFIGURATION.md`](../../../docs/CONFIGURATION.md) format
- **Link consistency**: Use relative links to other documentation files
- **Code block syntax**: Use `hcl` syntax highlighting for Packer configs
- **Security emphasis**: Note passwords should use variables, not hardcoded strings
- **HOME expansion**: Document that `~` is expanded in file paths

## Success Criteria

Documentation is complete when:

- [ ] [`docs/CONFIGURATION.md`](../../../docs/CONFIGURATION.md) includes new "SSH Tunnel Mode" section with complete field reference
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

## Open Questions

None - the feature is fully implemented and requirements are clear from the prompt file.

## Related Changes

- Implements requirements from prompt [`openspec/prompts/04-update-documentation.md`](../../../openspec/prompts/04-update-documentation.md)
- Completes documentation sequence after:
  - `add-ssh-tunnel-config-options` (prompt 01)
  - `implement-ssh-tunnel-establishment` (prompt 02)
  - `integrate-tunnel-with-inventory` (prompt 03)
