# Tasks: Fix Port Type Coercion in SSH Tunnel Mode

## Implementation Tasks

- [x] Update Port extraction logic in `provisioner.go` line ~1388 to use type switch
  - Handle `int` type (direct assignment)
  - Handle `string` type (parse with `strconv.Atoi`)
  - Provide clear error for unsupported types
- [x] Add port range validation (1-65535)
- [x] Import `strconv` package if not already imported (already present)
- [x] Update error messages to distinguish failure modes:
  - Missing port (nil type handled in default case)
  - Invalid port format (non-numeric string) - wrapped strconv error
  - Out-of-range port - explicit range check
  - Unsupported type - type %T reported

## Validation Tasks

- [x] Run `make generate` to regenerate any HCL2 specs if needed (not needed - no Config changes)
- [x] Run `go build ./...` to verify compilation (SUCCESS)
- [x] Run `go test ./...` to verify no test regressions (SUCCESS - all tests pass)
- [x] Run `make plugin-check` to verify plugin conformance (SUCCESS)
- [ ] Test with Port as `int` (manual or integration test) - requires manual testing
- [ ] Test with Port as `string` (manual or integration test) - requires manual testing

## Documentation Tasks

- [x] Update spec deltas to reflect new Port type handling behavior (already specified in spec.md)
- [x] Ensure error messages are documented in spec scenarios (error scenarios documented in spec.md)
