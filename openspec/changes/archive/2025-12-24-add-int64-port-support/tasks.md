# Tasks: Add int64/int32 Support for SSH Tunnel Port Extraction

## Implementation Tasks

- [x] Add `int64` case to the type switch at [`provisioner.go:1517-1529`](../../provisioner/ansible-navigator/provisioner.go:1517) with conversion to `int`
- [x] Add `int32` case to the type switch with conversion to `int`
- [x] Add `int16` case to the type switch with conversion to `int`
- [x] Add `int8` case to the type switch with conversion to `int`
- [x] Add `uint` case to the type switch with range validation (max 65535) and conversion to `int`
- [x] Add `uint64` case to the type switch with range validation (max 65535) and conversion to `int`
- [x] Add `uint32` case to the type switch with range validation (max 65535) and conversion to `int`
- [x] Add `uint16` case to the type switch with conversion to `int` (no range check needed - inherently ≤ 65535)
- [x] Add `uint8` case to the type switch with conversion to `int` (no range check needed - inherently ≤ 255)
- [x] Update the `default` case error message to indicate numeric types are expected

## Testing Tasks

- [x] Add unit test for port as `int64(22)` - verify successful extraction
- [x] Add unit test for port as `int32(2222)` - verify successful extraction
- [x] Add unit test for port as `int16(3333)` - verify successful extraction
- [x] Add unit test for port as `int8(80)` - verify successful extraction
- [x] Add unit test for port as `uint(8080)` - verify successful extraction
- [x] Add unit test for port as `uint64(443)` - verify successful extraction
- [x] Add unit test for port as `uint32(8000)` - verify successful extraction
- [x] Add unit test for port as `uint16(9090)` - verify successful extraction
- [x] Add unit test for port as `uint8(80)` - verify successful extraction
- [x] Add unit test for invalid string port `"invalid"` - verify error
- [x] Add unit test for unsigned port exceeding max: `uint64(70000)` - verify range error
- [x] Add unit test for negative port: `int64(-22)` - verify range error
- [x] Add unit test for zero port: `int64(0)` - verify range error
- [x] Add unit test for unsupported type (e.g., `float64`, `bool`) - verify type error
- [x] Verify backward compatibility with existing `int` and `string` test cases

## Validation Tasks

- [x] Run `go build ./...` to verify code compiles
- [x] Run unit tests for new port extraction logic - all tests pass
- [ ] Run `go test ./...` to verify all tests pass (note: pre-existing tests have unrelated bastion config failures)
- [ ] Run `make plugin-check` to verify plugin conformance
- [ ] Run `make generate` to verify HCL2 spec consistency (not needed - no Config structs changed)
- [ ] Test with actual amazon-ebs builder configuration providing `int64` port
- [ ] Verify error messages are clear and actionable

## Documentation Tasks

- [x] Update inline code comments to document integer type support (done in provisioner.go:1516)
- [ ] Add example to documentation showing compatibility with amazon-ebs builder
- [x] Update CHANGELOG.md with bugfix entry

## Dependencies

None - this is a standalone code change.

## Notes

- The conversion from signed/unsigned integer types to `int` is safe for port numbers since the valid range (1-65535) fits well within all these types
- `uint16` inherently cannot exceed 65535, so no range check is needed
- `uint8` maxes at 255, so no range check is needed
- Larger unsigned types (`uint`, `uint32`, `uint64`) require explicit range checks before conversion
- The existing port range validation (1-65535) at line 1532 will catch any remaining edge cases after type conversion
