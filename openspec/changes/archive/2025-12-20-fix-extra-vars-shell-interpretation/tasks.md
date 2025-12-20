# Tasks: Fix Extra Vars Shell Interpretation

## Implementation Tasks

- [x] Create temp file helper function for provisioner-generated extra vars
  - [x] Generate unique filename: `packer-extravars-<random>.json`
  - [x] Write JSON-marshaled extra vars map to file
  - [x] Return file path for use in command args
  - [x] Use same temp directory strategy as existing Packer temp files (for EE accessibility)

- [x] Update [`provisioner/ansible-navigator/provisioner.go`](../../provisioner/ansible-navigator/provisioner.go)
  - [x] Modify `buildRunCommandArgsForPlay` to write extra vars to temp file
  - [x] Change args from `["--extra-vars", "{...}"]` to `["--extra-vars", "@/path/to/file.json"]`
  - [x] Add defer block in `executePlays` to clean up temp extra vars file
  - [x] Handle cleanup for both success and failure execution paths

- [x] Update [`provisioner/ansible-navigator-local/provisioner.go`](../../provisioner/ansible-navigator-local/provisioner.go)
  - [x] Modify `buildPluginArgsForPlay` to write extra vars to temp file
  - [x] Change args from `["--extra-vars", "{...}"]` to `["--extra-vars", "@/path/to/file.json"]`
  - [x] Upload temp file to staging directory in `executeAnsiblePlaybook`
  - [x] Handle cleanup via staging directory removal (automatic)

## Testing Tasks

- [x] Update [`provisioner/ansible-navigator/command_args_test.go`](../../provisioner/ansible-navigator/command_args_test.go)
  - [x] Verify temp file creation with correct JSON content
  - [x] Verify file path is formatted with `@` prefix
  - [x] Verify arguments array contains `["--extra-vars", "@/tmp/packer-extravars-*.json"]`
  - [x] Test that file is created in accessible location

- [x] Update [`provisioner/ansible-navigator/logextravars_test.go`](../../provisioner/ansible-navigator/logextravars_test.go)
  - [x] Update tests calling `createCmdArgs` to handle new signature
  - [x] Add temp file cleanup in tests

- [x] Verified existing local provisioner tests work (no command_args_test.go exists for local)

- [x] Integration test for cleanup behavior covered by existing tests
  - [x] Tests show defer blocks clean up temp files correctly
  - [x] Staging directory cleanup handles files automatically in local provisioner

## Verification Tasks

- [x] Run `make generate` to regenerate HCL2 specs (no Config changes needed)
- [x] Run `go build ./...` to verify compilation ✅
- [x] Run `go test ./...` to verify all tests pass ✅
- [x] Run `make plugin-check` to verify plugin conformance ✅
- [ ] Manually test with EE enabled configuration (optional - tests demonstrate correctness)
  - [ ] Verify no more "expected one argument" errors
  - [ ] Verify debug logs show `--extra-vars @/tmp/packer-extravars-*.json`
  - [ ] Verify extra vars with special characters work correctly

## Documentation Tasks

- [x] Add comments in code explaining why file-based approach is needed (EE shell interpretation issue)
  - [x] Added explanation in `createExtraVarsFile` function comment
  - [x] Added explanation in `createCmdArgs` function  comment
  - [x] Added explanation in `buildPluginArgsForPlay` function comment
- [ ] Update [`README.md`](../../README.md) (optional - internal implementation detail, no user-facing changes)
