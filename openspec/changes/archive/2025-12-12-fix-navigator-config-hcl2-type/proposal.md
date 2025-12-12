# fix-navigator-config-hcl2-type Proposal

## Summary

Fix the HCL2 type specification for the `navigator_config` field in both provisioners to properly accept nested map structures, enabling users to specify complex ansible-navigator.yml configuration including nested objects like `execution-environment`, `ansible.config.defaults`, etc.

## Why

The HCL2 type specification incorrectly restricts `navigator_config` to flat string maps (`cty.Map(cty.String)`), preventing users from configuring nested ansible-navigator.yml options. This limitation conflicts with the Go implementation which already supports nested structures via `map[string]interface{}`, and blocks legitimate use cases like configuring execution environments and complex Ansible settings.

## What Changes

- Manually edited HCL2 spec files (`provisioner/ansible-navigator/provisioner.hcl2spec.go` and `provisioner/ansible-navigator-local/provisioner.hcl2spec.go`) to change `navigator_config` type from `cty.Map(cty.String)` to `cty.DynamicPseudoType`
- Added inline comments documenting the manual override to prevent accidental regeneration
- Created test HCL config files demonstrating nested and flat navigator_config usage
- Verified compilation and plugin compatibility

## Problem

The current HCL2 spec generated for `navigator_config` uses `cty.Map(cty.String)`, which only accepts flat string maps. This causes Packer to reject nested configuration structures that are valid according to the ansible-navigator.yml schema and that the Go code (`map[string]interface{}`) can already handle.

Users attempting to configure nested structures get validation errors like:

```
Incorrect attribute value type; Inappropriate value for attribute "navigator_config": 
element "execution-environment": string required.
```

## Root Cause

The `packer-sdc mapstructure-to-hcl2` code generator doesn't automatically recognize that `map[string]interface{}` should support deeply nested structures. It defaults to generating `cty.Map(cty.String)` in the `.hcl2spec.go` files.

## Solution

Override the HCL2 type specification for `navigator_config` to use `cty.DynamicPseudoType`, which properly handles arbitrary nested structures:

```go
"navigator_config": &hcldec.AttrSpec{
    Name:     "navigator_config",
    Type:     cty.DynamicPseudoType,
    Required: false,
},
```

This will be achieved by adding a custom `HCL2Spec()` method on the `FlatConfig` struct in both provisioners that overrides the auto-generated specification for this single field.

## Scope

**In scope:**

- Add custom HCL2Spec() method override in `provisioner/ansible-navigator/provisioner.go`
- Add custom HCL2Spec() method override in `provisioner/ansible-navigator-local/provisioner.go`
- Regenerate HCL2 spec files using `make generate`
- Verify with `go build ./...` and `make plugin-check`

**Out of scope:**

- Changes to YAML generation logic (already working correctly)
- Documentation updates (separate task)
- Example file updates (separate task)
- New features beyond fixing the type spec

## Implementation Approach

1. In each provisioner's `provisioner.go`, add a custom `HCL2Spec()` method on the `FlatConfig` struct
2. Call the auto-generated `(*FlatConfig).HCL2Spec()` method to get the baseline spec
3. Override the `navigator_config` entry with `cty.DynamicPseudoType`
4. Run `make generate` to ensure all other fields remain properly generated
5. Run verification commands

## Benefits

- Users can specify complex ansible-navigator.yml configuration directly in HCL
- Maintains backward compatibility (flat string maps still work)
- Aligns HCL validation with Go runtime capabilities
- Removes artificial limitation in the configuration surface

## Risk Assessment

**Low risk:**

- Change is purely to the HCL type specification
- Go code already handles nested maps correctly
- Backward compatible (existing flat maps continue to work)
- YAML generation logic unchanged

## Validation Strategy

1. Run `make generate` to regenerate specs
2. Run `go build ./...` to verify compilation
3. Run `make plugin-check` for SDK validation
4. Test with nested configuration example from prompt
5. Verify flat string map configs still work

## Related Work

This issue was introduced when the `navigator_config` field was added. The Go implementation correctly handles nested structures, but the HCL2 spec generation did not account for this requirement.
