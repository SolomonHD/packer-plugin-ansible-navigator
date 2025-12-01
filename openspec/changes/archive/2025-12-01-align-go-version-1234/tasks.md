# Tasks: Align Go Version to 1.23.4

## 1. Update go.mod
- [x] 1.1 Change `go 1.24.0` to `go 1.23.4` in go.mod
- [x] 1.2 Remove the `toolchain go1.24.10` directive
- [x] 1.3 Verify all `replace` directives are preserved
- [x] 1.4 Verify the `retract` directive for v0.0.1 and v0.0.2 is preserved

## 2. Dependency Resolution
- [x] 2.1 Run `go mod tidy` to resolve dependencies for Go 1.23.4
- [x] 2.2 If conflicts arise, document them and investigate resolution
  - Conflicts arose due to transitive dependencies requiring Go 1.24+
  - Resolution: Created clean go.mod with only direct dependencies, ran `go mod tidy -compat=1.23`
  - Dependencies automatically resolved to compatible versions (e.g., golang.org/x/crypto v0.38.0)

## 3. Verification
- [x] 3.1 Run `go build ./...` to verify compilation succeeds
- [x] 3.2 Run `go test ./...` to verify all tests pass
- [x] 3.3 Run `make plugin-check` to verify plugin compatibility

## 4. Documentation
- [x] 4.1 Verify README.md Go version references are consistent (if any)
  - Fixed incorrect references from "1.25.3+" to "1.23.4+" in badge and requirements
- [x] 4.2 Update CHANGELOG.md if needed
  - No CHANGELOG update required - this is a build tooling alignment, not a feature/fix release