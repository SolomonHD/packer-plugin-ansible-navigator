# Align Go version to 1.23.4

## Context

The project has a version mismatch:
- `.go-version` specifies `1.23.4` (correct target)
- `go.mod` specifies `go 1.24.0` with `toolchain go1.24.10`

There's also a `replace` directive in `go.mod` that specifically mentions Go 1.23.4 compatibility:
```
// Constrain consul/api to a version compatible with Go 1.23.4
replace github.com/hashicorp/consul/api => github.com/hashicorp/consul/api v1.28.2
```

This suggests the project is intended to use Go 1.23.4, but `go.mod` was accidentally upgraded.

## Goal

Update `go.mod` to align with `.go-version` and use Go 1.23.4 consistently across the project.

## Scope

### In scope:
- Update `go.mod` to specify `go 1.23.4`
- Remove or update the `toolchain` directive (if needed for Go 1.23.4)
- Run `go mod tidy` to ensure dependency compatibility
- Verify the project builds successfully with Go 1.23.4
- Run verification commands: `go build ./...`, `go test ./...`, `make plugin-check`

### Out of scope:
- Changing `.go-version` (already correct at 1.23.4)
- Changing any Go code or functionality
- Updating dependency versions beyond what `go mod tidy` automatically resolves
- Modifying the `replace` directives (they should remain as-is)
- Modifying CI/CD configurations

## Desired behavior

After the change:
- `.go-version` still contains: `1.23.4` (no change needed)
- `go.mod` contains: `go 1.23.4`
- The `toolchain` directive is either removed or set appropriately for Go 1.23.4
- The project compiles without errors using Go 1.23.4
- All tests pass under Go 1.23.4
- The `make plugin-check` target succeeds

## Constraints & assumptions

- The existing `replace` directives in `go.mod` must be preserved (especially the consul/api constraint for Go 1.23.4 compatibility)
- The `retract` directive for v0.0.1 and v0.0.2 must be preserved
- If `go mod tidy` causes dependency conflicts, they should be reported rather than silently ignored
- Go 1.23.4 should be compatible with all current dependencies given the consul/api constraint specifically mentions this version

## Acceptance criteria

- [ ] `go.mod` specifies `go 1.23.4`
- [ ] `toolchain` directive is removed or appropriate for Go 1.23.4
- [ ] All `replace` and `retract` directives remain unchanged
- [ ] `go mod tidy` runs successfully without errors
- [ ] `go build ./...` completes successfully
- [ ] `go test ./...` passes all tests
- [ ] `make plugin-check` succeeds
- [ ] No dependency version conflicts reported