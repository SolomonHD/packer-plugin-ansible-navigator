# OpenSpec Prompts Index

This directory contains OpenSpec prompt files for improving the packer-plugin-ansible-navigator.

## Prompt Sequence

Execute these prompts in order:

1. **[01-fix-port-type-coercion.md](01-fix-port-type-coercion.md)** - Fix Port type assertion bug in SSH tunnel mode
2. **[02-add-connection-mode-enum.md](02-add-connection-mode-enum.md)** - Replace use_proxy/ssh_tunnel_mode with single connection_mode enum
3. **[03-add-bastion-block.md](03-add-bastion-block.md)** - Restructure bastion configuration as nested block with sub-variables

## Context

Plugin: `/home/solomong/dev/packer/plugins/packer-plugin-ansible-navigator`

These changes address:

- Bug: SSH tunnel mode fails with "requires a valid target port" due to type assertion
- UX: Confusing double-negative with `use_proxy=false` to enable tunnel mode  
- Config: Flat bastion_* variables should be nested for clarity
