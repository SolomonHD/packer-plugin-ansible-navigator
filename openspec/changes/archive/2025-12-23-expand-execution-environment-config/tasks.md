# Implementation Tasks

## Preparation

- [x] Review ansible-navigator v3.x documentation for execution-environment schema
  - Validated the expected YAML shapes (notably `execution-environment.pull.{policy,arguments}` and hyphenated keys like `container-engine`).
- [x] Identify complete list of missing fields from current [`ExecutionEnvironment`](../../provisioner/ansible-navigator/provisioner.go:92) struct
  - `container_engine`, `container_options`, `pull_arguments`, and `volume_mounts` were already present in the struct; gaps were in YAML emission and coverage tests.

## Struct Definition Updates

- [x] Add `container_engine` field to [`ExecutionEnvironment`](../../provisioner/ansible-navigator/provisioner.go:92) struct
- [x] Add `container_options` field to [`ExecutionEnvironment`](../../provisioner/ansible-navigator/provisioner.go:92) struct
- [x] Add `pull_arguments` field to [`ExecutionEnvironment`](../../provisioner/ansible-navigator/provisioner.go:92) struct
- [x] Add corresponding fields to local provisioner's [`ExecutionEnvironment`](../../provisioner/ansible-navigator-local/provisioner.go:1) struct (if separate)
- [x] Add documentation comments for all new struct fields
  - Already present in both provisioners.

## YAML Generation Updates

- [x] Update [`convertToYAMLStructure()`](../../provisioner/ansible-navigator/navigator_config.go:160) to emit `container-engine` field
- [x] Update [`convertToYAMLStructure()`](../../provisioner/ansible-navigator/navigator_config.go:160) to emit `container-options` list
- [x] Update [`convertToYAMLStructure()`](../../provisioner/ansible-navigator/navigator_config.go:160) to nest `pull.arguments` alongside existing `pull.policy`
- [x] Ensure pull struct is created correctly when only one of policy/arguments is provided
- [x] Update local provisioner's YAML generation (if separate implementation)

## HCL2 Spec Generation

- [x] Verify [`go:generate`](../../provisioner/ansible-navigator/provisioner.go:6) directive includes `ExecutionEnvironment` type
- [x] Run `make generate` to regenerate all [`*.hcl2spec.go`](../../provisioner/ansible-navigator/*.hcl2spec.go:1) files
- [x] Verify generated specs include all new fields with correct cty types
  - Confirmed `container_engine` (cty.String), `container_options`/`pull_arguments` (cty.List(cty.String)) appear in generated spec.

## Testing

- [x] Add unit test for HCL decoding of `container_engine` field
- [x] Add unit test for HCL decoding of `container_options` list
- [x] Add unit test for HCL decoding of `pull_arguments` list
- [x] Add unit test for YAML generation with `container_engine` only
- [x] Add unit test for YAML generation with `container_options` only
- [x] Add unit test for YAML generation with `pull_arguments` only
- [x] Add unit test for YAML generation with `pull_policy` and `pull_arguments` combined
- [x] Add unit test for YAML generation with ALL new fields configured
- [x] Add unit test verifying backward compatibility (existing configs unchanged)
  - Existing YAML/EE default behavior tests continue to pass.
- [x] Add unit test for volume mount deduplication with additional user mounts

## Build Verification

- [x] Run `make generate` to regenerate HCL2 specs
- [x] Run `go build ./...` to verify compilation
- [x] Run `go test ./...` to verify all tests pass
- [x] Run `make plugin-check` to verify x5 API compliance

## Integration Testing

- [x] Create example template in [`example/`](../../example/) using new options
  - Updated `example/execution-environment-config.pkr.hcl` to demonstrate `container_engine`, `container_options`, and `pull_arguments`.
- [x] Run `packer validate` on example template
  - Validated via `make dev` + `packer init execution-environment-config.pkr.hcl` + `packer validate execution-environment-config.pkr.hcl`.
- [ ] Run `packer build` on example template (if safe to do so)
  - Not run (docker build side-effects).

## Documentation

- [x] Update [`README.md`](../../README.md) execution environment section with new options
- [x] Add examples showing `container_engine`, `container_options`, `pull_arguments` usage
- [x] Update any existing examples if needed for clarity

## Final Verification

- [x] Verify no breaking changes to existing configurations
- [x] Verify all acceptance criteria from proposal are met
- [x] Verify YAML output matches ansible-navigator v3.x schema expectations
