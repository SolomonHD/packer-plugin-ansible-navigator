# OpenSpec Change Prompt

## Context

**Plugin**: packer-plugin-ansible-navigator  
**File**: `provisioner/ansible-navigator/provisioner.go`  
**Lines**: 1517-1529  
**Error**: `SSH tunnel mode: Port must be int or string, got type int64 with value 22`

When using `connection_mode = "ssh_tunnel"`, the provisioner extracts the target machine's port from Packer's `generatedData["Port"]`. The current type switch only handles `int` and `string`, but the Packer amazon-ebs builder provides the port as `int64`.

## Goal

Fix the type assertion in the SSH tunnel port extraction to handle all integer types that Packer may provide, including `int64` from the amazon-ebs builder.

## Scope

**In scope:**

- Add `int64` case to the type switch at lines 1517-1529
- Optionally add other integer types (`int32`, `uint`, `uint16`, `uint32`, `uint64`) for robustness
- Validate converted value fits in `int` range

**Out of scope:**

- Changes to other provisioner logic
- Bastion configuration restructuring
- UX improvements to connection mode configuration

## Desired Behavior

The `switch` statement at lines 1517-1529 should:

1. Accept `int`, `int64`, `int32`, and `string` types
2. Convert 64-bit values to `int` safely
3. Continue to validate port is 1-65535 after extraction

## Constraints & Assumptions

- Assumption: Packer amazon-ebs builder provides `generatedData["Port"]` as `int64`
- Assumption: Other Packer builders may provide `int` or `string`
- Constraint: Must maintain backward compatibility with existing configs
- Constraint: Port values cannot exceed 65535

## Acceptance Criteria

- [ ] Build succeeds with SSH tunnel mode using amazon-ebs builder
- [ ] Port as `int64`: `generatedData["Port"] = int64(22)` → succeeds
- [ ] Port as `int`: `generatedData["Port"] = 22` → succeeds
- [ ] Port as `string`: `generatedData["Port"] = "22"` → succeeds
- [ ] Port as `int32`: `generatedData["Port"] = int32(22)` → succeeds
- [ ] `go build ./...` passes
- [ ] `go test ./...` passes
- [ ] `make plugin-check` passes

## Code Location

```go
// provisioner.go lines 1517-1529
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

## Fix Required

Add cases for `int64` and optionally other integer types:

```go
var targetPort int
switch v := generatedData["Port"].(type) {
case int:
    targetPort = v
case int64:
    targetPort = int(v)
case int32:
    targetPort = int(v)
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
