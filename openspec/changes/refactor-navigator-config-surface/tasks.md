# Tasks

## 1. Config struct changes

- [ ] 1.1 Add `NavigatorConfig` field (map[string]interface{}) to both provisioner Config structs
- [ ] 1.2 Remove legacy config fields from both provisioners:
  - `ansible_cfg`, `ansible_env_vars`, `ansible_ssh_extra_args`, `extra_arguments`
  - `execution_environment`, `navigator_mode`
  - `roles_path`, `collections_path`, `galaxy_command`
- [ ] 1.3 Run `make generate` to regenerate HCL2 specs after struct changes
- [ ] 1.4 Verify config fields validate correctly via `go test ./...`

## 2. Navigator config file generation

- [ ] 2.1 Implement function to convert HCL `navigator_config` map to YAML ansible-navigator.yml format
- [ ] 2.2 Add logic to generate temporary ansible-navigator.yml file from NavigatorConfig
- [ ] 2.3 Implement automatic defaults when `execution-environment.enabled = true`:
  - Set default `ansible.config.defaults.remote_tmp = "/tmp/.ansible/tmp"`
  - Set default `ansible.config.defaults.local_tmp = "/tmp/.ansible-local"`
  - Set default `execution-environment.environment-variables.ANSIBLE_REMOTE_TMP`
  - Set default `execution-environment.environment-variables.ANSIBLE_LOCAL_TMP`
- [ ] 2.4 Add YAML generation tests verifying structure and defaults

## 3. Execution changes

- [ ] 3.1 Update remote provisioner to use `ANSIBLE_NAVIGATOR_CONFIG` environment variable pointing to generated config
- [ ] 3.2 Update local provisioner to use `ANSIBLE_NAVIGATOR_CONFIG` in remote shell command
- [ ] 3.3 Remove code that processes removed config options (ansible_cfg generation, execution_environment string handling, etc.)
- [ ] 3.4 Ensure config file cleanup happens in deferred functions

## 4. Validation

- [ ] 4.1 Update Config.Validate() to accept navigator_config as optional
- [ ] 4.2 Remove validation for deleted fields
- [ ] 4.3 Add validation that navigator_config, if present, is a valid map structure
- [ ] 4.4 Add validation tests for new navigator_config field

## 5. Command construction

- [ ] 5.1 Remote provisioner: Remove `--mode`, `--ee`, `--eei` CLI flag logic (now controlled by config file)
- [ ] 5.2 Local provisioner: Remove `--mode`, `--ee`, `--eei` from remote shell command
- [ ] 5.3 Verify ansible-navigator uses config file for all settings
- [ ] 5.4 Add tests for command construction with navigator_config

## 6. Documentation

- [ ] 6.1 Update README.md with new configuration examples using navigator_config
- [ ] 6.2 Update docs/CONFIGURATION.md to document navigator_config and remove all references to deleted options
- [ ] 6.3 Update docs/EXAMPLES.md with new navigator_config examples
- [ ] 6.4 Update docs/TROUBLESHOOTING.md to cover navigator_config debugging
- [ ] 6.5 Add migration guide section explaining how to convert from old config to navigator_config
- [ ] 6.6 Verify no references to removed options remain in any documentation

## 7. Tests

- [ ] 7.1 Update existing unit tests to use navigator_config instead of removed options
- [ ] 7.2 Add tests for navigator_config YAML generation
- [ ] 7.3 Add tests for automatic EE defaults
- [ ] 7.4 Add tests for config file cleanup
- [ ] 7.5 Verify all tests pass with `go test ./...`

## 8. Verification

- [ ] 8.1 Run `make generate` and verify HCL2 specs are correct
- [ ] 8.2 Run `go build ./...` and verify compilation succeeds
- [ ] 8.3 Run `go test ./...` and verify all tests pass
- [ ] 8.4 Run `make plugin-check` and verify plugin conformance
- [ ] 8.5 Manual test with execution environment and navigator_config

## 9. OpenSpec maintenance

- [ ] 9.1 Apply spec deltas to openspec/specs after implementation
- [ ] 9.2 Archive this change via `openspec archive refactor-navigator-config-surface --yes`
- [ ] 9.3 Verify with `openspec validate --strict`
