# Tasks: Fix ansible.config Schema Validation Error

## 1. Update Config Structs

- [x] 1.1 Modify `AnsibleConfig` struct in `provisioner/ansible-navigator/provisioner.go`:
  - Remove `Defaults *AnsibleConfigDefaults` field (or mark deprecated)
  - Remove `SSHConnection *AnsibleConfigConnection` field (or mark deprecated)
  - Add `Path string` field (path to ansible.cfg)
  - Add `Help bool` field (show help flag)
  - Add `Cmdline string` field (additional command-line arguments)
- [x] 1.2 Modify `AnsibleConfig` struct in `provisioner/ansible-navigator-local/provisioner.go` (same changes as 1.1)
- [x] 1.3 Remove or deprecate `AnsibleConfigDefaults` and `AnsibleConfigConnection` struct definitions if no longer used
- [x] 1.4 Update `//go:generate` directive to remove deleted struct types

## 2. Implement ansible.cfg Generation

- [x] 2.1 Create `generateAnsibleCfg()` function in `provisioner/ansible-navigator/navigator_config.go`:
  - Accept Ansible configuration settings (remote_tmp, host_key_checking, ssh_timeout, pipelining, etc.)
  - Generate INI-formatted ansible.cfg content
  - Return string with INI content
- [x] 2.2 Create `createAnsibleCfgFile()` function in `provisioner/ansible-navigator/navigator_config.go`:
  - Accept INI content string
  - Create temporary file with pattern `packer-ansible-cfg-*.cfg`
  - Write content to file
  - Return absolute path
- [x] 2.3 Duplicate functions in `provisioner/ansible-navigator-local/navigator_config.go`

## 3. Update YAML Generation Logic

- [x] 3.1 Modify `convertToYAMLStructure()` in `provisioner/ansible-navigator/navigator_config.go`:
  - Remove lines 108-129 that generate `config.defaults` and `config.ssh_connection`
  - Add logic to generate valid `ansible.config` structure:
    - `help` (if set)
    - `path` (if temp ansible.cfg is generated)
    - `cmdline` (if set)
  - Keep the `config.Config` string field handling (line 105-107) for user-provided ansible.cfg path
- [x] 3.2 Modify `convertToYAMLStructure()` in `provisioner/ansible-navigator-local/navigator_config.go` (same changes as 3.1)

## 4. Update EE Defaults Logic

- [x] 4.1 Modify `generateNavigatorConfigYAML()` in `provisioner/ansible-navigator/navigator_config.go`:
  - When EE is enabled, generate ansible.cfg content instead of setting struct fields
  - Call `generateAnsibleCfg()` with default values
  - Call `createAnsibleCfgFile()` to write temp file
  - Store the temp file path in `AnsibleConfig.Path`
  - Return both nav config YAML and ansible.cfg path (modify signature)
- [x] 4.2 Modify `generateNavigatorConfigYAML()` in `provisioner/ansible-navigator-local/navigator_config.go` (same changes as 4.1)

## 5. Update Provisioner Integration

- [x] 5.1 Modify `executeAnsible()` in `provisioner/ansible-navigator/provisioner.go`:
  - Capture ansible.cfg path from `generateNavigatorConfigYAML()` (now returns multiple values)
  - Add ansible.cfg cleanup to deferred  cleanup function
  - Track ansible.cfg path for cleanup on error
- [x] 5.2 Modify local provisioner's ansible execution (similar logic in `provisioner/ansible-navigator-local/provisioner.go`)
- [x] 5.3 For local provisioner: Upload generated ansible.cfg to staging directory if needed

## 6. Regenerate HCL2 Specs

- [x] 6.1 Run `make generate` from project root to regenerate `.hcl2spec.go` files
- [x] 6.2 Verify generated files reflect struct changes

## 7. Update Tests

- [ ] 7.1 Update `provisioner/ansible-navigator/navigator_config_test.go`:
  - Add tests for ansible.cfg generation
  - Add tests verifying `ansible.config` only contains valid properties
  - Remove tests for deprecated `defaults` and `ssh_connection` fields
- [ ] 7.2 Update local provisioner tests similarly
- [ ] 7.3 Add integration test validating generated ansible-navigator.yml passes schema validation

## 8. Documentation Updates

- [ ] 8.1 Update README.md to document the change
- [ ] 8.2 Update docs/CONFIGURATION.md:
  - Document new `ansible_config.path`, `ansible_config.help`, `ansible_config.cmdline` fields
  - Mark `ansible_config.defaults` and `ansible_config.ssh_connection` as removed
  - Add migration guidance
- [ ] 8.3 Update docs/EXAMPLES.md with corrected navigator_config examples
- [ ] 8.4 Add CHANGELOG.md entry documenting the breaking change and migration path

## 9. Validation

- [x] 9.1 Run `go build ./...` - must succeed ✓ PASSED
- [ ] 9.2 Run `go test ./...` - all tests must pass (FAILED - tests need updates for new struct, see section 7)
- [ ] 9.3 Run `make plugin-check` - must pass (pending)
- [ ] 9.4 Manual test: Run plugin with EE enabled, verify generated ansible-navigator.yml passes schema validation (pending)
- [ ] 9.5 Manual test: Verify generated ansible.cfg contains expected settings (pending)
