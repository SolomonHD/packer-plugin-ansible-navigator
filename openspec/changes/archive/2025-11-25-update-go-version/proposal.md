# Change: Update Go Version to 1.25.3

## Why
The project currently has inconsistent Go version specifications. The `.go-version` file specifies version 1.25.3, but the `go.mod` file still shows version 1.23. To ensure consistency and take advantage of the latest Go features and improvements, we should update the Go version in `go.mod` to match the `.go-version` file.

## What Changes
- Update Go version in `go.mod` from 1.23 to 1.25.3
- Verify that all code is compatible with Go 1.25.3
- No breaking changes expected as this is a minor version update

## Impact
- Affected specs: None (this is a build/tooling change)
- Affected code: go.mod file
- No runtime behavior changes expected