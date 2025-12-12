# Tasks

## Implementation Tasks

- [x] **Add --mode CLI flag for remote provisioner**
  - Modify `provisioner/ansible-navigator/provisioner.go` command construction
  - Check if `NavigatorConfig != nil && NavigatorConfig.Mode != ""`
  - If true, insert `"--mode", config.NavigatorConfig.Mode` after `"run"` argument
  - Position before playbook-specific arguments (inventory, extra vars, etc.)

- [x] **Add --mode CLI flag for local provisioner**
  - Modify `provisioner/ansible-navigator-local/provisioner.go` remote shell command construction
  - Use same conditional logic as remote provisioner
  - Insert `--mode` flag after `run` in the remote command string

- [x] **Wrap YAML in ansible-navigator root key (remote)**
  - Modify `provisioner/ansible-navigator/navigator_config.go` `convertToYAMLStructure()` function
  - Change return value to wrap entire structure in `map[string]interface{}{"ansible-navigator": <existing_structure>}`
  - Ensure nested structures (execution-environment, ansible, logging, etc.) remain intact
  - Verify field name conversions (underscores to hyphens) still work correctly

- [x] **Wrap YAML in ansible-navigator root key (local)**
  - Modify `provisioner/ansible-navigator-local/navigator_config.go` `convertToYAMLStructure()` function
  - Apply identical wrapping logic as remote provisioner
  - Ensure consistency between local and remote YAML generation

- [ ] **Add documentation about asdf shim recursion**
  - Create or update `docs/troubleshooting.md`
  - Document asdf shim recursion issue (some installations cause loops)
  - Recommend using `command` with full path: `/home/user/.asdf/installs/python/x.y.z/bin/ansible-navigator`
  - Recommend using `ansible_navigator_path` to prepend directories: `ansible_navigator_path = ["~/.asdf/installs/python/3.11.6/bin"]`
  - Note that most asdf installations work fine, issue is configuration-specific

- [ ] **Add documentation link from README**
  - Update main `README.md` if it doesn't already link to troubleshooting docs
  - Add reference to troubleshooting guide in configuration examples section

## Validation Tasks

- [x] **Regenerate HCL2 specs**
  - Run `make generate` in plugin directory
  - Verify no changes needed (Config structs unchanged by this fix)

- [x] **Build all packages**
  - Run `go build ./...` to verify compilation
  - Check for any errors related to command construction or YAML generation

- [x] **Run unit tests**
  - Run `go test ./...` to execute all tests
  - Add/update tests if needed for:
    - Mode flag insertion logic
    - YAML root key wrapping
    - Various navigator_config combinations

- [x] **Run plugin conformance check**
  - Run `make plugin-check` to validate x5 API compliance
  - Ensure no regressions from changes

## Testing Tasks

- [ ] **Test with null builder and simple playbook**
  - Create minimal test config with `build.null` source
  - Use `navigator_config.mode = "stdout"`
  - Verify `--mode stdout` appears in command output
  - Verify ansible-navigator executes without hanging
  - Check generated YAML has `ansible-navigator:` root key

- [ ] **Test generated YAML structure with various configs**
  - Test with execution_environment enabled
  - Test with ansible_config defaults
  - Test with logging configuration
  - Verify all cases produce valid YAML with root key

- [ ] **Test backward compatibility**
  - Use existing HCL config files (if available in examples/)
  - Verify they work without modification
  - Confirm no breaking changes to user-facing config

- [ ] **Test without navigator_config**
  - Verify plugin works when navigator_config is omitted
  - Ensure no --mode flag is added in this case
  - Ensure no YAML file is generated

- [ ] **Test with AWS builder and role FQDN**
  - Optional: test with real AWS infrastructure if available
  - Verify behavior with role execution (generates temp playbook)
  - Confirm mode flag and YAML root key work in real scenario

## Documentation Tasks

- [ ] **Update MIGRATION.md if needed**
  - Check if MIGRATION.md needs updates about YAML structure changes
  - Note that user-facing HCL doesn't change
  - Clarify that this is a transparent fix for generated internal files

- [ ] **Update CHANGELOG.md**
  - Add entry for bugfix:
    - "Fixed: ansible-navigator hanging - now passes --mode CLI flag when configured"
    - "Fixed: Generated ansible-navigator.yml now uses proper v25.x schema with root key"
  - Note backward compatibility maintained

## Completion Criteria

All tasks above must be completed and:

- [x] All four verification commands pass: `make generate && go build ./... && go test ./... && make plugin-check`
- [ ] Test execution with `mode = "stdout"` completes without hanging
- [ ] Generated YAML files conform to ansible-navigator v25.x schema
- [ ] Schema validation errors about "unexpected properties" are resolved
- [ ] Existing configurations continue to work without changes
