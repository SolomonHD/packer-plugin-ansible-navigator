# OpenSpec Prompts Index: SSH Tunnel Mode Feature

This index tracks the decomposed prompts for implementing the SSH tunnel mode feature.

## Overview

Add `ssh_tunnel_mode` option to bypass the Packer SSH proxy adapter by establishing direct SSH tunnels through a bastion host. This solves WSL2/Docker container networking reliability issues.

## Prompt Sequence

Execute prompts in numerical order. Each prompt builds on the previous.

### 01-add-ssh-tunnel-config-options.md

**Status**: Pending  
**Dependencies**: None  
**Description**: Add configuration schema for SSH tunnel mode including bastion host settings and mode selection.

### 02-implement-ssh-tunnel-establishment.md

**Status**: Pending  
**Dependencies**: 01  
**Description**: Implement SSH tunnel establishment logic and connection management through bastion host.

### 03-integrate-tunnel-with-inventory.md

**Status**: Pending  
**Dependencies**: 02  
**Description**: Modify inventory generation to use tunnel ports when SSH tunnel mode is active.

### 04-update-documentation.md

**Status**: Pending  
**Dependencies**: 03  
**Description**: Document SSH tunnel mode configuration, bastion setup, and WSL2/container troubleshooting guidance.

## Implementation Notes

- SSH tunnel mode is **optional** and only activates when configured
- Preserves existing proxy adapter behavior as default
- Tunnel mode and proxy adapter are mutually exclusive
- Target use case: WSL2/Docker execution environments where container-to-host networking is unreliable

## Testing Strategy

Each prompt should include:

- Configuration validation tests
- Connection establishment tests  
- Error handling tests
- Integration tests with execution environments

## Related Files

Key files that will be modified:

- [`provisioner/ansible-navigator/provisioner.go`](../../provisioner/ansible-navigator/provisioner.go) - Core provisioner logic
- [`docs/CONFIGURATION.md`](../../docs/CONFIGURATION.md) - User-facing configuration reference
- OpenSpec specs under `openspec/specs/remote-provisioner-capabilities/spec.md`
