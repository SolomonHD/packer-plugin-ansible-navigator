# Proposal: Add int64/int32 Support for SSH Tunnel Port Extraction

## Overview

Extend the SSH tunnel port extraction logic to handle all integer types (int64, int32, int16, int8, uint, uint64, uint32, uint16, uint8) that Packer builders may provide for the `Port` field in `generatedData`.

## Problem Statement

When using `connection_mode = "ssh_tunnel"`, the provisioner extracts the target machine's port from Packer's `generatedData["Port"]` at lines 1517-1529 of [`provisioner.go`](../../provisioner/ansible-navigator/provisioner.go:1517). The current type switch only handles `int` and `string`:

```go
var targetPort int
switch v := generatedData["Port"].(type) {
case int:
    targetPort = v
case string:
    var err error
    targetPort, err = strconv.Atoi(v)
    if err != nil {
        return fmt.Errorf("SSH tunnel mode: invalid port value %q: %w", v, err)
    }
default:
    return fmt.Errorf("SSH tunnel mode: Port must be int or string, got type %T with value %v", v, v)
}
```

However, the Packer amazon-ebs builder (and potentially other builders) provides `Port` as `int64`, causing the provisioner to fail with:

```
SSH tunnel mode: Port must be int or string, got type int64 with value 22
```

## Proposed Solution

Add cases for all Go integer types to the type switch:

1. Handle `int64`, `int32`, `int16`, `int8` by converting to `int`
2. Handle unsigned types (`uint`, `uint64`, `uint32`, `uint16`, `uint8`) with range validation
3. Validate all converted values fit within valid port range (1-65535)
4. Maintain backward compatibility with existing `int` and `string` support

```go
var targetPort int
switch v := generatedData["Port"].(type) {
case int:
    targetPort = v
case int64:
    targetPort = int(v)
case int32:
    targetPort = int(v)
case int16:
    targetPort = int(v)
case int8:
    targetPort = int(v)
case uint:
    if v > 65535 {
        return fmt.Errorf("SSH tunnel mode: port value %d exceeds maximum 65535", v)
    }
    targetPort = int(v)
case uint64:
    if v > 65535 {
        return fmt.Errorf("SSH tunnel mode: port value %d exceeds maximum 65535", v)
    }
    targetPort = int(v)
case uint32:
    if v > 65535 {
        return fmt.Errorf("SSH tunnel mode: port value %d exceeds maximum 65535", v)
    }
    targetPort = int(v)
case uint16:
    targetPort = int(v)
case uint8:
    targetPort = int(v)
case string:
    var err error
    targetPort, err = strconv.Atoi(v)
    if err != nil {
        return fmt.Errorf("SSH tunnel mode: invalid port value %q: %w", v, err)
    }
default:
    return fmt.Errorf("SSH tunnel mode: Port must be a numeric or string type, got type %T with value %v", v, v)
}

// Existing validation continues
if targetPort < 1 || targetPort > 65535 {
    return fmt.Errorf("SSH tunnel mode: port must be between 1-65535, got %d", targetPort)
}
```

## Impact

### Benefits

- SSH tunnel mode will work with all Packer builders regardless of Port type
- Eliminates type-related failures when using amazon-ebs and other builders
- Maintains full backward compatibility
- Provides clear validation and error messages

### Risks

- Minimal risk: only affects the error path for unsupported types
- No API or configuration changes required
- No behavior changes for working configurations

## Scope

### In-Scope

- Port type extraction logic at lines 1517-1529 in `provisioner.go`
- Range validation for unsigned integer types
- Updated error messages
- Unit tests for all integer type cases

### Out-of-Scope

- Changes to other provisioner logic
- Connection mode configuration improvements
- Bastion configuration changes
- Changes to Packer core Port handling

## Dependencies

None. This is a standalone robustness improvement.

## Assumptions

- **Packer builders may provide Port as any Go integer type** (int, int64, int32, uint, etc.)
- The amazon-ebs builder specifically provides Port as `int64`
- Port values from builders will always represent valid TCP port numbers (1-65535)
- String representation is still needed for some communicators

## Related Specs

- [`ssh-tunnel-runtime`](specs/ssh-tunnel-runtime/spec.md) - Will define Port type handling requirements
