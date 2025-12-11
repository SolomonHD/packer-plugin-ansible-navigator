# Tasks

## 1. Remove deprecated config fields

- [ ] Remove `ansible_cfg` field from both provisioner Config structs
- [ ] Remove `ansible_env_vars` field from both provisioner Config structs
- [ ] Remove `ansible_ssh_extra_args` field from both provisioner Config structs
- [ ] Remove `extra_arguments` field from both provisioner Config structs
- [ ] Remove `execution_environment` field from both provisioner Config structs
- [ ] Remove `navigator_mode` field from both provisioner Config structs
- [ ] Remove `roles_path` field from both provisioner Config structs
- [ ] Remove `collections_path` field from both provisioner Config structs
- [ ] Remove `galaxy_command` field from both provisioner Config structs
- [ ] Run `make generate` to regenerate HCL2 specs after field removal

## 2. Remove processing code for deprecated options

- [ ] Remove ansible_cfg file generation logic
- [ ] Remove ansible_env_vars processing logic
- [ ] Remove ansible_ssh_extra_args handling
- [ ] Remove extra_arguments CLI flag generation
- [ ] Remove execution_environment string handling
- [ ] Remove navigator_mode CLI flag (`--mode`) generation
- [ ] Remove `--ee` CLI flag generation
- [ ] Remove `--eei` CLI flag generation
- [ ] Clean up any helper functions only used by removed options

## 3. Update galaxy.go

- [ ] Update galaxy.go to not reference removed `roles_path` field
- [ ] Update galaxy.go to not reference removed `collections_path` field
- [ ] Update galaxy.go to not reference removed `galaxy_command` field
- [ ] Ensure galaxy functionality works with requirements_file only

## 4. Update validation

- [ ] Remove validation for `ansible_cfg`
- [ ] Remove validation for `ansible_env_vars`
- [ ] Remove validation for `ansible_ssh_extra_args`
- [ ] Remove validation for `extra_arguments`
- [ ] Remove validation for `execution_environment` string
- [ ] Remove validation for `navigator_mode`
- [ ] Remove validation for `roles_path`
- [ ] Remove validation for `collections_path`
- [ ] Remove validation for `galaxy_command`
- [ ] Ensure validation still covers all retained options

## 5. Update tests

- [ ] Update unit tests to use navigator_config instead of removed options
- [ ] Remove tests specific to removed options
- [ ] Add tests verifying removed options cause validation errors
- [ ] Ensure all tests pass with `go test ./...`

## 6. Documentation cleanup

- [ ] Remove all references to `ansible_cfg` from documentation
- [ ] Remove all references to `ansible_env_vars` from documentation
- [ ] Remove all references to `ansible_ssh_extra_args` from documentation
- [ ] Remove all references to `extra_arguments` from documentation
- [ ] Remove all references to `execution_environment` string from documentation
- [ ] Remove all references to `navigator_mode` from documentation
- [ ] Remove all references to `roles_path` from documentation
- [ ] Remove all references to `collections_path` from documentation
- [ ] Remove all references to `galaxy_command` from documentation
- [ ] Update README.md with breaking change notice
- [ ] Update CONFIGURATION.md to only document supported options
- [ ] Update EXAMPLES.md to only show navigator_config examples
- [ ] Keep MIGRATION.md as reference for users upgrading from old versions

## 7. Error messages

- [ ] Add helpful error messages when removed options are detected
- [ ] Error messages should point users to MIGRATION.md
- [ ] Error messages should show which removed option was found
- [ ] Error messages should show the navigator_config equivalent

## 8. Verification

- [ ] Run `make generate` and verify HCL2 specs are correct
- [ ] Run `go build ./...` and verify compilation succeeds
- [ ] Run `go test ./...` and verify all tests pass
- [ ] Run `make plugin-check` and verify plugin conformance
- [ ] Manual test with navigator_config to ensure full functionality
- [ ] Manual test with old config to verify clear error messages
