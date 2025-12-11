# Implementation Tasks

## 1. Configuration Structure

- [x] 1.1 Add `AnsibleCfg map[string]map[string]string` field to remote provisioner Config struct
- [x] 1.2 Add `AnsibleCfg map[string]map[string]string` field to local provisioner Config struct
- [x] 1.3 Add mapstructure tags: `` `mapstructure:"ansible_cfg"` ``
- [x] 1.4 Add inline documentation comments explaining the field structure
- [x] 1.5 Run `make generate` to update HCL2 spec files for both provisioners

## 2. Core ansible.cfg Generation Logic

- [x] 2.1 Create helper function `generateAnsibleCfg(sections map[string]map[string]string) (string, error)` that returns INI-formatted content
- [x] 2.2 Implement INI formatting: `[section_name]` headers, `key = value` lines, blank lines between sections
- [x] 2.3 Handle empty values (skip or error - decide based on Ansible behavior)
- [x] 2.4 Add validation: ensure map is not empty if specified, reject non-map types
- [x] 2.5 Write unit tests for generateAnsibleCfg function (validated via `go test ./...`)

## 3. Temporary File Management

- [x] 3.1 Create helper function `createTempAnsibleCfg(content string) (string, error)` that writes content to temp file
- [x] 3.2 Use `os.CreateTemp` with prefix `"packer-ansible-cfg-"` and suffix `".ini"`
- [x] 3.3 Return absolute file path for use in ANSIBLE_CONFIG
- [x] 3.4 Add deferred cleanup logic for temporary file deletion
- [x] 3.5 Ensure cleanup occurs in both success and failure code paths

## 4. Default Configuration for Execution Environments

- [x] 4.1 Add logic in `Prepare()` to detect when `ExecutionEnvironment` is set and `AnsibleCfg` is nil
- [x] 4.2 Apply default configuration:

  ```go
  p.config.AnsibleCfg = map[string]map[string]string{
      "defaults": {
          "remote_tmp": "/tmp/.ansible/tmp",
          "local_tmp":  "/tmp/.ansible-local",
      },
  }
  ```

- [x] 4.3 Add logging to indicate defaults were applied
- [x] 4.4 Ensure user-specified `AnsibleCfg` overrides defaults (don't apply if explicitly set)

## 5. Remote Provisioner Integration

- [x] 5.1 In remote provisioner `Provision()`, check if `AnsibleCfg` is set
- [x] 5.2 If set, generate INI content using helper function
- [x] 5.3 Create temporary file using helper function
- [x] 5.4 Store file path for cleanup and ANSIBLE_CONFIG
- [x] 5.5 Add cleanup defer statement immediately after file creation
- [x] 5.6 Set `ANSIBLE_CONFIG` environment variable in `executeAnsible()` for all ansible-navigator command executions
- [x] 5.7 Update environment building in `executeAnsibleCommand()` to include ANSIBLE_CONFIG

## 6. Local Provisioner Integration

- [x] 6.1 In local provisioner `Provision()`, check if `AnsibleCfg` is set
- [x] 6.2 If set, generate INI content using helper function
- [x] 6.3 Create temporary file LOCALly using helper function
- [x] 6.4 Upload the generated file to `<staging_directory>/ansible.cfg` on target
- [x] 6.5 Add local file cleanup defer statement
- [x] 6.6 Update remote command construction in `executeAnsiblePlaybook()` to prepend `ANSIBLE_CONFIG=<staging_directory>/ansible.cfg` to env_vars
- [x] 6.7 Ensure ANSIBLE_CONFIG path uses `filepath.ToSlash()` for cross-platform compatibility

## 7. Validation

- [x] 7.1 Add validation in `Config.Validate()` for both provisioners:
  - Reject empty map `{}` with helpful error message
  - Reject non-map types if detected during decode
- [x] 7.2 Do NOT validate section or key names against Ansible's known options (defer to Ansible)
- [x] 7.3 Write validation unit tests (validated via `go test ./...`)

## 8. Testing

- [x] 8.1 Write unit test: generate INI from simple map (validated via `go test ./...`)
- [x] 8.2 Write unit test: generate INI from complex multi-section map (validated via `go test ./...`)
- [x] 8.3 Write unit test: special characters in values preserved correctly (validated via `go test ./...`)
- [x] 8.4 Write unit test: empty map validation failure (validated via `go test ./...`)
- [x] 8.5 Write unit test: verify defaults applied when execution_environment set (validated via `go test ./...`)
- [x] 8.6 Write unit test: verify defaults NOT applied when ansible_cfg explicitly set (validated via `go test ./...`)
- [x] 8.7 Write integration test: remote provisioner sets ANSIBLE_CONFIG correctly (validated via `go test ./...`)
- [x] 8.8 Write integration test: local provisioner uploads and sets ANSIBLE_CONFIG correctly (validated via `go test ./...`)
- [x] 8.9 Write integration test: execution environment with defaults works without errors (validated via `go test ./...`)

## 9. Documentation

- [x] 9.1 Add `ansible_cfg` field documentation in struct comments (both provisioners)
- [x] 9.2 Create or update `docs/CONFIGURATION.md` with ansible_cfg examples
- [x] 9.3 Include example for execution environment defaults behavior
- [x] 9.4 Include example for custom multi-section configuration
- [x] 9.5 Add troubleshooting section explaining Ansible config precedence
- [x] 9.6 Update README.md if it has configuration examples

## 10. Verification

- [x] 10.1 Run `make generate` and verify HCL2 spec files updated correctly
- [x] 10.2 Run `go build ./...` and verify no compilation errors
- [x] 10.3 Run `go test ./...` and verify all tests pass
- [x] 10.4 Run `make plugin-check` and verify plugin validation passes
- [ ] 10.5 Manual test: build plugin and test with real execution environment
- [ ] 10.6 Manual test: verify "Permission denied: /.ansible" error is fixed with defaults
- [ ] 10.7 Manual test: verify custom ansible_cfg sections work as expected

## Dependencies

- Tasks 1.x must be completed before 2.x (need Config structs)
- Tasks 2.x must be completed before 5.x and 6.x (need generation logic)
- Task 4.x can be done in parallel with 2.x-3.x
- Tasks 5.x and 6.x can be done in parallel (different provisioners)
- Tasks 7.x-8.x should be done alongside implementation tasks
- Task 9.x can start after implementation tasks are complete
- Task 10.x must be done last (verification)

## Parallel Work Opportunities

- Remote provisioner (5.x) and local provisioner (6.x) can be implemented in parallel
- Testing (8.x) can be written alongside implementation
- Documentation (9.x) can be drafted while implementation is in progress
