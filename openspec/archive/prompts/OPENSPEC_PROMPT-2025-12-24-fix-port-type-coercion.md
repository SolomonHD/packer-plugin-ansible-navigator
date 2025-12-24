# OpenSpec Prompt: Fix Port Type Coercion Bug in SSH Tunnel Mode

## Context

**Plugin**: packer-plugin-ansible-navigator  
**File**: `/home/solomong/dev/packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/provisioner.go`  
**Lines**: 1382-1391  

When `ssh_tunnel_mode = true` and `use_proxy = false`, the provisioner attempts to create an SSH tunnel to the target host. However, it fails with error:

```
SSH tunnel mode requires a valid target port
```

## Goal

Fix the type assertion bug when extracting the target port from Packer's `generatedData` map. The current code assumes Port is an `int`, but Packer provides it as a `string` when using certain communicators/configurations.

## Scope

### In-Scope

- Fix the Port type extraction to handle both `int` and `string` types
- Add proper error handling for invalid port values
- Maintain backward compatibility with existing configurations
- Add validation that port is non-zero

### Out-of-Scope

- Changes to other provisioner logic
- UX improvements to connection mode configuration (separate prompt)
- Restructuring of bastion configuration (separate prompt)
- Changes to builder-level Port handling in Packer core

## Desired Behavior

The `Provision()` function should:

1. Extract `Port` from `generatedData` as either `int` or `string`
2. Convert string port values to integers using `strconv.Atoi()`
3. Validate the port is between 1 and 65535
4. Provide clear error messages if port is invalid or missing
5. Continue with SSH tunnel establishment using the parsed port

### Expected Code Changes

In `provisioner/ansible-navigator/provisioner.go` around lines 1382-1391:

**Current (buggy):**

```go
target Port, ok := generatedData["Port"].(int)
if !ok || targetPort == 0 {
    return fmt.Errorf("SSH tunnel mode requires a valid target port")
}
```

**Fixed:**

```go
// Handle Port as either int or string (Packer may provide either)
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

if targetPort < 1 || targetPort > 65535 {
    return fmt.Errorf("SSH tunnel mode: port must be between 1-65535, got %d", targetPort)
}
```

## Constraints & Assumptions

- The bug exists in v7.1.0 of the plugin
- Packer's `generatedData["Port"]` type varies by communicator and configuration
- SSH typically uses port 22, but custom ports should be supported (1-65535)
- The fix must not break existing working configurations
- Import `strconv` package if not already imported

## Acceptance Criteria

### Must Have

- [ ] `targetPort` extraction handles both `int` and `string` types from `generatedData["Port"]`
- [ ] Port values are validated to be between 1 and 65535
- [ ] Clear error messages distinguish between:
  - Missing port value
  - Invalid port value (non-numeric string)
  - Out-of-range port value
  - Unsupported type
- [ ] The fix is applied in the `Provision()` function where SSH tunnel mode is checked
- [ ] Code compiles successfully with `go build ./...`
- [ ] No regressions in existing test coverage

### Should Have

- [ ] Error messages include the actual value received for debugging
- [ ] Type switch pattern is idiomatic Go
- [ ] The same Port handling approach is used consistently if Port is accessed elsewhere

### Test Cases

- [ ] Port as int: `generatedData["Port"] = 22` → succeeds with `targetPort = 22`
- [ ] Port as string: `generatedData["Port"] = "22"` → succeeds with `targetPort = 22`
- [ ] Port as string with custom port: `generatedData["Port"] = "2222"` → succeeds with `targetPort = 2222`
- [ ] Port missing: `generatedData["Port"]` not set → fails with clear error
- [ ] Port as invalid string: `generatedData["Port"] = "abc"` → fails with "invalid port value" error
- [ ] Port as zero: `generatedData["Port"] = 0` → fails with "port must be between 1-65535" error
- [ ] Port out of range: `generatedData["Port"] = 99999` → fails with "port must be between 1-65535" error

## Expected Files Touched

- `provisioner/ansible-navigator/provisioner.go` (lines ~1382-1391 in the `Provision()` function)

## Dependencies

None - this is a standalone bug fix that does not depend on other changes.
