# Change: Fix Navigator Config HCL2 Struct Type Mapping Issues

## Summary

Fix HCL2 type specification issues in the navigator_config typed structs, specifically the `EnvironmentVariablesConfig` with `mapstructure:",remain"` and `AnsibleConfig` with `mapstructure:",squash"` tags that do not correctly translate to HCL2 parsing.

## Why

The current typed struct approach for `navigator_config` uses mapstructure tags that do not work correctly with `packer-sdc mapstructure-to-hcl2` code generation:

1. **`EnvironmentVariablesConfig` with `mapstructure:",remain"`**: The intent is to capture arbitrary key-value pairs directly inside the `environment_variables` block. However, the generated HCL2 spec creates a nested `variables` attribute, requiring users to write:

   ```hcl
   environment_variables {
     variables = { KEY = "value" }
   }
   ```

   Instead of the intended syntax:

   ```hcl
   environment_variables {
     KEY = "value"
   }
   ```

2. **`AnsibleConfig` with `mapstructure:",squash"` on `Inner`**: The squash tag causes the `AnsibleConfigInner` fields to be "bubbled up" in the FlatConfig, but the generated HCL2Spec only includes the `config` attribute, losing the nested `defaults` and `ssh_connection` block structure.

3. **Overall struct complexity**: The current design attempts to model ansible-navigator.yml's complex nested structure with Go structs, but the mapstructure-to-hcl2 generator has limitations that prevent proper HCL block parsing.

## What Changes

### Option A: Flatten Environment Variables (Recommended)

- Replace `EnvironmentVariablesConfig` with explicit named fields for common environment variables
- Add a generic `additional_env` map for less common variables  
- Remove the `mapstructure:",remain"` pattern that doesn't translate to HCL2

### Option B: Use BlockAttrsSpec Override

- Keep the struct design but override HCL2Spec() methods to use `hcldec.BlockAttrsSpec` for the `environment_variables` block
- This requires manual editing of generated `.hcl2spec.go` files or custom methods

### For AnsibleConfig

- Remove the `mapstructure:",squash"` tag from `AnsibleConfig.Inner`
- Change to explicit `defaults` and `ssh_connection` nested blocks  
- Alternatively, flatten the config into `AnsibleConfig` directly

### Struct changes (both provisioners)

1. Redesign `EnvironmentVariablesConfig` to have explicit fields
2. Fix `AnsibleConfig` squash issue with proper nested block structure
3. Run `make generate` to regenerate HCL2 specs
4. Update example files to reflect correct syntax

## Impact

- **Affected specs**: `remote-provisioner-capabilities`, `local-provisioner-capabilities`
- **Affected code**:
  - `provisioner/ansible-navigator/provisioner.go` (struct definitions)
  - `provisioner/ansible-navigator-local/provisioner.go` (struct definitions)
  - `provisioner/ansible-navigator/*.hcl2spec.go` (generated)
  - `provisioner/ansible-navigator-local/*.hcl2spec.go` (generated)
  - `example/` directory files

## **BREAKING CHANGE**

This change may modify the HCL syntax for `navigator_config` blocks. Users may need to update their configurations if the struct field layout changes.

## Constraints

- Must use typed structs (not `map[string]interface{}`) per project rules
- Generated `.hcl2spec.go` files must use RPC-serializable types (no `cty.DynamicPseudoType`)
- Must be generated via `make generate` (no manual `.hcl2spec.go` editing if possible)
- YAML generation to ansible-navigator.yml must continue to work

## Related Work

- Previous proposal `fix-navigator-config-hcl2-type` addressed the `map[string]interface{}` → `cty.DynamicPseudoType` issue
- Proposal `replace-navigator-config-with-typed-structs` introduced the current typed struct approach
