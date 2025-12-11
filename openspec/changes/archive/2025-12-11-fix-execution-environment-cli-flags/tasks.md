# Tasks: Fix execution environment CLI flag generation

## 1. Remote Provisioner Code Changes

- [x] 1.1 Update `executePlays()` in `provisioner/ansible-navigator/provisioner.go` to generate `--ee true --eei <image>` instead of `--execution-environment <image>` when `ExecutionEnvironment` is set
- [x] 1.2 Update `executeSinglePlaybook()` in `provisioner/ansible-navigator/provisioner.go` with the same fix
- [x] 1.3 Update any related inline comments that mention the old `--execution-environment` flag syntax

## 2. Local Provisioner Code Changes

- [x] 2.1 Update command construction in `provisioner/ansible-navigator-local/provisioner.go` to generate `--ee true --eei <image>` instead of `--execution-environment <image>`
- [x] 2.2 Update any related inline comments that mention the old flag syntax

## 3. Test Updates

- [x] 3.1 Update or add unit tests in `provisioner/ansible-navigator/provisioner_test.go` to verify the new CLI flag generation (tests passed)
- [x] 3.2 Update or add unit tests in `provisioner/ansible-navigator-local/provisioner_test.go` if applicable (tests passed)

## 4. Documentation Updates

- [x] 4.1 Update `docs/CONFIGURATION.md` execution_environment section to explain the underlying CLI flags used (`--ee true --eei <image>`)
- [x] 4.2 Review and update any other documentation that mentions deployment/execution details

## 5. Verification

- [x] 5.1 Run `make generate` to regenerate HCL2 specs (if Config struct changes - not expected)
- [x] 5.2 Run `go build ./...` to verify compilation
- [x] 5.3 Run `go test ./...` to verify tests pass
- [x] 5.4 Manual verification with a real ansible-navigator v3+ installation (optional but recommended)
