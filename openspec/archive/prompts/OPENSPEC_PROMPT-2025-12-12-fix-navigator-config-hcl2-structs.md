# OpenSpec Change Prompt

## Context

The `packer-plugin-ansible-navigator` plugin has HCL2 type specification issues causing `environment_variables` blocks and other nested configuration structs to fail during Packer parsing. The plugin defines typed Go structs for `navigator_config` but several design patterns conflict with how `packer-sdc mapstructure-to-hcl2` generates HCL2 specs.

## Goal

Fix all HCL2 variable and nested block type issues in the `navigator_config` structs across both provisioners (`ansible-navigator` and `ansible-navigator-local`) so users can properly configure execution environments with environment variables and other nested settings.

## Scope

**In scope:**

- `EnvironmentVariablesConfig` struct and its HCL2 spec
- `AnsibleConfig` struct and nested `AnsibleConfigInner` with `mapstructure:",squash"` usage
- All nested config structs: `ExecutionEnvironment`, `LoggingConfig`, `PlaybookArtifact`, `CollectionDocCache`
- Both provisioner files: `provisioner/ansible-navigator/provisioner.go` and `provisioner/ansible-navigator-local/provisioner.go`
- Regeneration of `.hcl2spec.go` files via `make generate`
- Example files in `example/` directory

**Out of scope:**

- Changes to the `Play` struct or play execution logic
- Changes to SSH/WinRM proxy or communicator code
- Changes to GalaxyManager or dependency installation
- Changes to non-HCL configuration parsing

## Desired Behavior

1. Users can write `environment_variables` blocks with key-value pairs:

   ```hcl
   environment_variables {
     ANSIBLE_REMOTE_TMP = "/tmp/.ansible/tmp"
     MY_CUSTOM_VAR = "value"
   }
   ```

2. All nested blocks parse correctly without errors:
   - `execution_environment { ... }`
   - `ansible_config { ... }`
   - `logging { ... }`
   - `playbook_artifact { ... }`
   - `collection_doc_cache { ... }`

3. Go code can access parsed values through the typed structs
4. YAML generation in `navigator_config.go` continues to work correctly

## Constraints & Assumptions

- **Constraint:** Must use typed structs (not `map[string]interface{}`) to avoid RPC serialization crashes with `cty.DynamicPseudoType`
- **Constraint:** Generated `.hcl2spec.go` files must be regenerated via `make generate` after any struct changes
- **Assumption:** The `mapstructure:",remain"` tag on `EnvironmentVariablesConfig.Variables` is the root cause of the mismatch - this pattern doesn't translate well to HCL2 specs
- **Assumption:** The `mapstructure:",squash"` tag on `AnsibleConfig.Inner` may be causing the nested block to lose its structure
- **Assumption:** Both provisioner files have identical struct definitions and both need the same fixes

## Acceptance Criteria

- [ ] `packer validate` succeeds on example configs using `environment_variables` blocks with arbitrary key-value pairs
- [ ] `packer validate` succeeds on example configs using full `navigator_config` with all nested blocks
- [ ] `go build ./...` succeeds after struct modifications
- [ ] `go test ./...` passes (including any existing navigator_config tests)
- [ ] `make plugin-check` passes
- [ ] Example files in `example/` directory use correct syntax matching the fixed structs
- [ ] YAML generation produces valid `ansible-navigator.yml` content
