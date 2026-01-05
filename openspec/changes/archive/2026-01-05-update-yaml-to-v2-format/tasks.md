# Implementation Tasks

## 1. Research ansible-navigator Version 2 Format

- [x] 1.1 Search ansible-navigator GitHub repository for Version 2 configuration format documentation
- [x] 1.2 Review ansible-navigator settings file schema and migration changelog
- [x] 1.3 Find examples of Version 2 format ansible-navigator.yml files
- [x] 1.4 Identify specific differences between Version 1 and Version 2 (version markers, schema identifiers, structural changes)
- [x] 1.5 Document the exact format changes needed for pull-policy to work correctly

## 2. Update YAML Generation Code

- [x] 2.1 Update [`convertToYAMLStructure`](../../provisioner/ansible-navigator/navigator_config.go:160) function to produce Version 2 format
- [x] 2.2 Add version marker or schema identifier if required by Version 2
- [x] 2.3 Update pull-policy generation to match Version 2 requirements
- [x] 2.4 Apply same changes to [`provisioner/ansible-navigator-local/navigator_config.go`](../../provisioner/ansible-navigator-local/navigator_config.go) (verify if code is shared or duplicated)
- [x] 2.5 Ensure all other YAML structure elements conform to Version 2 schema

## 3. Testing and Verification

- [ ] 3.1 Test with ansible-navigator 25.12.0+ to verify no "version migration required" prompts
- [ ] 3.2 Verify generated YAML is recognized as Version 2 format immediately
- [ ] 3.3 Test that `pull_policy = "never"` actually prevents Docker image pulls when local image exists
- [ ] 3.4 Test with `pull_policy = "missing"` to verify it still works correctly
- [ ] 3.5 Verify all existing configuration options continue working (mode, execution-environment settings, ansible_config, logging, etc.)
- [x] 3.6 Run `go build ./...` to verify no compile errors
- [x] 3.7 Run `go test ./...` to verify existing tests pass
- [x] 3.8 Run `make generate` to regenerate HCL2 specs if Config structs changed
- [x] 3.9 Run `make plugin-check` to verify plugin compatibility

## 4. Documentation

- [x] 4.1 Update comments in navigator_config.go to reference Version 2 format
- [x] 4.2 Document any new version marker or schema identifier added
- [x] 4.3 Update inline code comments explaining pull-policy structure

## 5. Validation

- [x] 5.1 Run `openspec validate update-yaml-to-v2-format --strict`
- [x] 5.2 Resolve any validation errors
- [x] 5.3 Confirm all scenarios in spec deltas have proper formatting
