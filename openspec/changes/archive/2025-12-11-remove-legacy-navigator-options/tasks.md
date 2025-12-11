# Tasks

## 1. Remove deprecated config fields

- [x] Remove `ansible_cfg` field from both provisioner Config structs
- [x] Remove `ansible_env_vars` field from both provisioner Config structs
- [x] Remove `ansible_ssh_extra_args` field from both provisioner Config structs
- [x] Remove `extra_arguments` field from both provisioner Config structs
- [x] Remove `execution_environment` field from both provisioner Config structs
- [x] Remove `navigator_mode` field from both provisioner Config structs
- [x] Remove `roles_path` field from both provisioner Config structs
- [x] Remove `collections_path` field from both provisioner Config structs
- [x] Remove `galaxy_command` field from both provisioner Config structs
- [x] Run `make generate` to regenerate HCL2 specs after field removal

## 2. Remove processing code for deprecated options

- [x] Remove ansible_cfg file generation logic
- [x] Remove ansible_env_vars processing logic
- [x] Remove ansible_ssh_extra_args handling
- [x] Remove extra_arguments CLI flag generation
- [x] Remove execution_environment string handling
- [x] Remove navigator_mode CLI flag (`--mode`) generation
- [x] Remove `--ee` CLI flag generation
- [x] Remove `--eei` CLI flag generation
- [x] Clean up any helper functions only used by removed options (logDeprecationWarnings removed)

## 3. Update galaxy.go

- [x] Update galaxy.go to not reference removed `roles_path` field
- [x] Update galaxy.go to not reference removed `collections_path` field
- [x] Update galaxy.go to not reference removed `galaxy_command` field (hardcoded to "ansible-galaxy")
- [x] Ensure galaxy functionality works with requirements_file only

## 4. Update validation

- [x] Remove validation for `ansible_cfg`
- [x] Remove validation for `ansible_env_vars` (never had specific validation)
- [x] Remove validation for `ansible_ssh_extra_args` (never had specific validation)
- [x] Remove validation for `extra_arguments` (never had specific validation)
- [x] Remove validation for `execution_environment` string (never had specific validation)
- [x] Remove validation for `navigator_mode`
- [x] Remove validation for `roles_path` (never had specific validation)
- [x] Remove validation for `collections_path` (never had specific validation)
- [x] Remove validation for `galaxy_command` (never had specific validation)
- [x] Ensure validation still covers all retained options

## 5. Update tests

- [x] Update unit tests to use navigator_config instead of removed options
- [x] Remove tests specific to removed options
- [x] Add tests verifying removed options cause validation errors
- [x] Ensure all tests pass with `go test ./...`

## 6. Documentation cleanup

- [x] Remove all references to `ansible_cfg` from documentation
- [x] Remove all references to `ansible_env_vars` from documentation
- [x] Remove all references to `ansible_ssh_extra_args` from documentation
- [x] Remove all references to `extra_arguments` from documentation
- [x] Remove all references to `execution_environment` string from documentation
- [x] Remove all references to `navigator_mode` from documentation
- [x] Remove all references to `roles_path` from documentation
- [x] Remove all references to `collections_path` from documentation
- [x] Remove all references to `galaxy_command` from documentation
- [x] Update README.md with breaking change notice
- [x] Update CONFIGURATION.md to only document supported options
- [x] Update EXAMPLES.md to only show navigator_config examples
- [x] Keep MIGRATION.md as reference for users upgrading from old versions
- [x] Updated .web-docs to use navigator_config instead of deprecated options

## 7. Error messages

- [x] Error handling via HCL parsing (fields removed from Config, so HCL parser rejects them)
- [x] Documentation updated with breaking change notices pointing to MIGRATION.md
- [x] MIGRATION.md provides navigator_config equivalents for all removed options

Note: Since fields are completely removed from Config structs, Packer's HCL parser will reject unknown options before validation runs. Error messages come from HCL parser (e.g., "An argument named 'execution_environment' is not expected here"). Users are directed to MIGRATION.md via breaking change notices in README, CONFIGURATION.md, and EXAMPLES.md.

## 8. Verification

- [x] Run `make generate` and verify HCL2 specs are correct
- [x] Run `go build ./...` and verify compilation succeeds
- [x] Run `go test ./...` and verify all tests pass
- [x] Run `make plugin-check` and verify plugin conformance
- [ ] Manual test with navigator_config to ensure full functionality
- [ ] Manual test with old config to verify HCL parser rejects removed options
