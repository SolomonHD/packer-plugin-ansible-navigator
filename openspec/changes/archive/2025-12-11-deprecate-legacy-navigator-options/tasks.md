# Tasks

## 1. Add deprecation notices to Config structs

- [x] Add deprecation comments to ansible_cfg field in both provisioners
- [x] Add deprecation comments to ansible_env_vars field in both provisioners
- [x] Add deprecation comments to ansible_ssh_extra_args field in both provisioners
- [x] Add deprecation comments to extra_arguments field in both provisioners
- [x] Add deprecation comments to execution_environment field in both provisioners
- [x] Add deprecation comments to navigator_mode field in both provisioners
- [x] Add deprecation comments to roles_path field in both provisioners
- [x] Add deprecation comments to collections_path field in both provisioners
- [x] Add deprecation comments to galaxy_command field in both provisioners

## 2. Add runtime deprecation warnings

- [x] Implement logging function to emit deprecation warnings
- [x] Add warning when ansible_cfg is set
- [x] Add warning when ansible_env_vars is set
- [x] Add warning when ansible_ssh_extra_args is set
- [x] Add warning when extra_arguments is set
- [x] Add warning when execution_environment string is set
- [x] Add warning when navigator_mode is set
- [x] Add warning when roles_path is set
- [x] Add warning when collections_path is set
- [x] Add warning when galaxy_command is set
- [x] Ensure warnings include migration guidance and reference to documentation

## 3. Documentation updates

- [x] Update README.md with deprecation notices for affected options
- [x] Update docs/CONFIGURATION.md with deprecation notices
- [x] Mark deprecated options clearly in configuration reference
- [x] Create MIGRATION.md guide with before/after examples for each deprecated option
- [x] Add timeline for deprecation and removal
- [x] Update docs/EXAMPLES.md to primarily show navigator_config examples
- [x] Add deprecation notice banner to relevant documentation sections

## 4. Migration guide content

- [x] Document migration from execution_environment to navigator_config.execution-environment
- [x] Document migration from navigator_mode to navigator_config.mode
- [x] Document migration from ansible_cfg to navigator_config.ansible.config
- [x] Document migration from ansible_env_vars to navigator_config.execution-environment.environment-variables
- [x] Document migration from ansible_ssh_extra_args
- [x] Document migration from extra_arguments
- [x] Document migration from roles_path and collections_path
- [x] Document migration from galaxy_command
- [x] Provide complete working examples for common scenarios

## 5. Verification

- [x] Run `go build ./...` and verify compilation succeeds
- [ ] Manually test with deprecated options to see warning messages
- [ ] Verify warning messages are clear and helpful
- [x] Verify documentation accurately reflects deprecation status
- [x] Verify MIGRATION.md examples are complete and accurate
