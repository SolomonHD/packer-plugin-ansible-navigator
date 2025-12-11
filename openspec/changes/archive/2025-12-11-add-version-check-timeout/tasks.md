# Implementation Tasks

## 1. Configuration Changes

- [x] 1.1 Add `version_check_timeout` field to remote provisioner Config struct
- [x] 1.2 Add `version_check_timeout` field to local provisioner Config struct
- [x] 1.3 Set default value to "60s" in Prepare() for both provisioners
- [x] 1.4 Add validation for timeout format (must be valid duration string)

## 2. Version Check Implementation

- [x] 2.1 Update `getVersion()` in ansible-navigator provisioner to use `context.WithTimeout`
- [x] 2.2 Replace `exec.Command` with `exec.CommandContext` in getVersion()
- [x] 2.3 Parse timeout value using `time.ParseDuration`
- [x] 2.4 Handle timeout errors distinctly from not-found errors

## 3. Error Messages

- [x] 3.1 Create timeout-specific error message with suggestions
- [x] 3.2 Maintain existing not-found error message format
- [x] 3.3 Include configured timeout value in timeout error
- [x] 3.4 Suggest ansible_navigator_path, skip_version_check, and timeout increase

## 4. HCL2 Spec Generation

- [x] 4.1 Run `make generate` to regenerate provisioner.hcl2spec.go for remote provisioner
- [x] 4.2 Run `make generate` to regenerate provisioner.hcl2spec.go for local provisioner
- [x] 4.3 Verify new field appears in both generated spec files
- [x] 4.4 Commit regenerated files (will be done with final commit)

## 5. Documentation

- [x] 5.1 Add version_check_timeout documentation to remote provisioner
- [x] 5.2 Add version_check_timeout documentation to local provisioner
- [x] 5.3 Create asdf-specific configuration examples
- [x] 5.4 Add troubleshooting section for version check hangs
- [x] 5.5 Document interaction with skip_version_check option

## 6. Validation & Testing

- [x] 6.1 Run `make generate` to ensure HCL2 specs are current
- [x] 6.2 Run `go build ./...` to verify compilation
- [x] 6.3 Run `go test ./...` to verify existing tests pass
- [x] 6.4 Run `make plugin-check` to verify x5 API compliance
- [ ] 6.5 Manual test: Verify default timeout (60s) works correctly
- [ ] 6.6 Manual test: Verify custom timeout (e.g., 10s) works correctly
- [ ] 6.7 Manual test: Verify skip_version_check bypasses timeout
- [ ] 6.8 Manual test: Verify clear error on timeout vs not-found
