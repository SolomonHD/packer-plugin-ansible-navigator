# Proposal: Fix Port Type Coercion in SSH Tunnel Mode

## Overview

Fix a type assertion bug in the SSH tunnel establishment code that prevents tunnels from being created when Packer provides the target port as a `string` instead of an `int`.

## Problem Statement

When `ssh_tunnel_mode = true` and `use_proxy = false`, the provisioner attempts to establish an SSH tunnel to the target host. However, the current implementation at line 1388 of [`provisioner.go`](../../provisioner/ansible-navigator/provisioner.go:1388) assumes `generatedData["Port"]` is always an `int`:

```go
targetPort, ok := generatedData["Port"].(int)
if !ok || targetPort == 0 {
    return fmt.Errorf("SSH tunnel mode requires a valid target port")
}
```

Depending on the communicator and configuration, Packer may provide `Port` as a `string`, causing the type assertion to fail and the provisioner to incorrectly report "SSH tunnel mode requires a valid target port".

## Proposed Solution

Update the Port extraction logic to handle both `int` and `string` types using a type switch:

1. Attempt to extract Port as `int` (backwards compatible)
2. If Port is a `string`, parse it using `strconv.Atoi()`
3. Validate the parsed port is within valid range (1-65535)
4. Provide clear error messages for each failure mode

## Impact

### Benefits

- SSH tunnel mode will work with all Packer communicator configurations
- Better error messages distinguish between missing, invalid, and out-of-range ports
- Maintains backward compatibility with existing working configurations

### Risks

- Low risk: only affects the error path that is currently broken
- No API changes to configuration surface

## Scope

### In-Scope

- Port type extraction and validation in the SSH tunnel setup code
- Error message improvements
- Import of `strconv` package if not already present

### Out-of-Scope

- Changes to other provisioner logic
- UX improvements to connection mode configuration (tracked separately)
- Restructuring of bastion configuration (tracked separately)
- Changes to Packer core Port handling

## Dependencies

None. This is a standalone bug fix.

## Related Specs

- [`ssh-tunnel-runtime`](../../specs/ssh-tunnel-runtime/spec.md) - Will be updated with new Port type handling requirements
