# Tasks: Fix Provisioner Registration Names

## 1. Implementation

- [x] 1.1 Update `main.go` to use `plugin.DEFAULT_NAME` for local provisioner registration
- [x] 1.2 Update `main.go` to use `"remote"` for remote provisioner registration
- [x] 1.3 Verify import of `plugin.DEFAULT_NAME` constant from SDK

## 2. Verification

- [x] 2.1 Run `go build ./...` to confirm compilation
- [x] 2.2 Run `./packer-plugin-ansible-navigator describe` and verify output shows `["-packer-default-plugin-name-", "remote"]`
- [x] 2.3 Run `make plugin-check` to validate plugin conformance

## 3. Documentation

- [x] 3.1 Update README.md HCL examples to show correct provisioner usage (verified existing examples are correct - user-facing HCL unchanged)
- [x] 3.2 Update any inline code comments in `main.go` explaining the registration pattern

## 4. Version Bump

- [x] 4.1 Bump minor version in `version/VERSION` (1.4.0 → 1.5.0)
- [x] 4.2 Add CHANGELOG entry documenting the registration fix