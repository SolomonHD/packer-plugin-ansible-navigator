# Change: Align Go Version to 1.23.4

## Why
The project has a Go version mismatch that needs to be corrected:
- `.go-version` specifies `1.23.4` (the intended target version)
- `go.mod` incorrectly specifies `go 1.24.0` with `toolchain go1.24.10`
- A `replace` directive in `go.mod` explicitly mentions Go 1.23.4 compatibility for `consul/api`

This indicates the `go.mod` was accidentally upgraded and needs to be downgraded to align with the project's intended Go version.

## What Changes
- Update `go.mod` from `go 1.24.0` to `go 1.23.4`
- Remove or adjust the `toolchain go1.24.10` directive (not needed for Go 1.23.x)
- Preserve all existing `replace` directives (especially the consul/api constraint)
- Preserve the `retract` directive for v0.0.1 and v0.0.2
- Run `go mod tidy` to ensure dependency compatibility
- Verify build and tests pass with Go 1.23.4

## Impact
- Affected specs: `build-tooling` (modifying Go version requirement from 1.25.3 to 1.23.4)
- Affected code: `go.mod` file only
- No runtime behavior changes expected
- Dependencies should remain compatible due to existing replace directives

## Constraints
- `.go-version` remains unchanged at 1.23.4
- All `replace` and `retract` directives must be preserved
- Project must compile and pass all tests after the change
- `make plugin-check` must succeed