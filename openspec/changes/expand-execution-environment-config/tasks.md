# Implementation Tasks

## Preparation
- [ ] Review ansible-navigator v3.x documentation for execution-environment schema
- [ ] Identify complete list of missing fields from current [`ExecutionEnvironment`](../../provisioner/ansible-navigator/provisioner.go:92) struct

## Struct Definition Updates
- [ ] Add `container_engine` field to [`ExecutionEnvironment`](../../provisioner/ansible-navigator/provisioner.go:92) struct
- [ ] Add `container_options` field to [`ExecutionEnvironment`](../../provisioner/ansible-navigator/provisioner.go:92) struct
- [ ] Add `pull_arguments` field to [`ExecutionEnvironment`](../../provisioner/ansible-navigator/provisioner.go:92) struct
- [ ] Add corresponding fields to local provisioner's [`ExecutionEnvironment`](../../provisioner/ansible-navigator-local/provisioner.go:1) struct (if separate)
- [ ] Add documentation comments for all new struct fields

## YAML Generation Updates
- [ ] Update [`convertToYAMLStructure()`](../../provisioner/ansible-navigator/navigator_config.go:160) to emit `container-engine` field
- [ ] Update [`convertToYAMLStructure()`](../../provisioner/ansible-navigator/navigator_config.go:160) to emit `container-options` list
- [ ] Update [`convertToYAMLStructure()`](../../provisioner/ansible-navigator/navigator_config.go:160) to nest `pull.arguments` alongside existing `pull.policy`
- [ ] Ensure pull struct is created correctly when only one of policy/arguments is provided
- [ ] Update local provisioner's YAML generation (if separate implementation)

## HCL2 Spec Generation
- [ ] Verify [`go:generate`](../../provisioner/ansible-navigator/provisioner.go:6) directive includes `ExecutionEnvironment` type
- [ ] Run `make generate` to regenerate all [`*.hcl2spec.go`](../../provisioner/ansible-navigator/*.hcl2spec.go:1) files
- [ ] Verify generated specs include all new fields with correct cty types

## Testing
- [ ] Add unit test for HCL decoding of `container_engine` field
- [ ] Add unit test for HCL decoding of `container_options` list
- [ ] Add unit test for HCL decoding of `pull_arguments` list
- [ ] Add unit test for YAML generation with `container_engine` only
- [ ] Add unit test for YAML generation with `container_options` only
- [ ] Add unit test for YAML generation with `pull_arguments` only
- [ ] Add unit test for YAML generation with `pull_policy` and `pull_arguments` combined
- [ ] Add unit test for YAML generation with ALL new fields configured
- [ ] Add unit test verifying backward compatibility (existing configs unchanged)
- [ ] Add unit test for volume mount deduplication with additional user mounts

## Build Verification
- [ ] Run `make generate` to regenerate HCL2 specs
- [ ] Run `go build ./...` to verify compilation
- [ ] Run `go test ./...` to verify all tests pass
- [ ] Run `make plugin-check` to verify x5 API compliance

## Integration Testing
- [ ] Create example template in [`example/`](../../example/) using new options
- [ ] Run `packer validate` on example template
- [ ] Run `packer build` on example template (if safe to do so)

## Documentation
- [ ] Update [`README.md`](../../README.md) execution environment section with new options
- [ ] Add examples showing `container_engine`, `container_options`, `pull_arguments` usage
- [ ] Update any existing examples if needed for clarity

## Final Verification
- [ ] Verify no breaking changes to existing configurations
- [ ] Verify all acceptance criteria from proposal are met
- [ ] Verify YAML output matches ansible-navigator v3.x schema expectations
