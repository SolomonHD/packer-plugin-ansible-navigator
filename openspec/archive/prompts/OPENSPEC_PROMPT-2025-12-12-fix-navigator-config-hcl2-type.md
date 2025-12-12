# OpenSpec Change Prompt

## Context

The `packer-plugin-ansible-navigator` has a `navigator_config` variable that allows users to provide structured ansible-navigator.yml configuration as a map. The Go code correctly handles nested maps (`map[string]interface{}`), but the HCL2 spec was generated incorrectly, resulting in Packer rejecting nested configuration structures.

This causes validation errors when users try to use the feature as intended in the documentation and tests.

## Goal

Fix the HCL2 spec generation for the `navigator_config` field in both provisioners so that it properly accepts nested map structures, allowing users to specify full ansible-navigator.yml configuration including nested objects like `execution-environment`, `ansible.config.defaults`, etc.

## Scope

**In scope:**

- Fix HCL2 spec for `navigator_config` in `provisioner/ansible-navigator/provisioner.go`
- Fix HCL2 spec for `navigator_config` in `provisioner/ansible-navigator-local/provisioner.go`
- Regenerate HCL2 spec files using `make generate` to apply the fix
- Verify the fix with `go build ./...` and `make plugin-check`

**Out of scope:**

- Changes to the YAML generation logic (already working correctly)
- Changes to example files
- Documentation updates (separate task)
- New features beyond fixing the type spec

## Desired Behavior

After the fix, users should be able to write Packer HCL like:

```hcl
provisioner "ansible-navigator" {
  play {
    name   = "Example"
    target = "playbook.yml"
  }
  
  navigator_config = {
    mode = "stdout"
    
    execution-environment = {
      enabled = true
      image   = "my-ee:latest"
      
      environment-variables = {
        set = {
          ANSIBLE_REMOTE_TMP = "/tmp/.ansible/tmp"
          ANSIBLE_LOCAL_TMP  = "/tmp/.ansible-local"
        }
      }
    }
    
    ansible = {
      config = {
        defaults = {
          host_key_checking = "False"
        }
      }
    }
  }
}
```

And have Packer successfully validate and run the configuration.

## Problem

The current HCL2 spec in `provisioner.hcl2spec.go` defines `navigator_config` as:

```go
&hcldec.AttrSpec{Name: "navigator_config", Type: cty.Map(cty.String), Required: false}
```

This `cty.Map(cty.String)` type only accepts a flat map of strings. It rejects nested objects.

The Go struct definition is correct:

```go
NavigatorConfig map[string]interface{} `mapstructure:"navigator_config"`
```

But the generated HCL2 spec doesn't match what `map[string]interface{}` can actually handle.

## Root Cause

The `packer-sdc mapstructure-to-hcl2` code generator doesn't automatically recognize that `map[string]interface{}` should support deeply nested structures. It defaults to `cty.Map(cty.String)`.

## Solution Approach

There are two possible approaches:

### Option A: Use hcldec.BlockSpec with DynamicAttr (Recommended)

Change the HCL2 spec to use `hcldec.BlockSpec` or `DynamicAttr` which properly handles nested, arbitrary structures:

```go
"navigator_config": &hcldec.BlockAttrsSpec{
    TypeName:    "navigator_config",
    ElementType: cty.DynamicPseudoType,
    Required:    false,
},
```

or

```go
"navigator_config": &hcldec.AttrSpec{
    Name:     "navigator_config",
    Type:     cty.DynamicPseudoType,
    Required: false,
},
```

### Option B: Manual Override in FlatConfig

Override the HCL2Spec method to provide a custom spec for navigator_config that handles nested types properly.

### Recommendation

Use **Option A** with `cty.DynamicPseudoType` in an `AttrSpec`. This is the simplest fix and aligns with how Packer handles other arbitrary nested configuration (like `extra_arguments`).

## Testing Strategy

After implementing the fix:

1. Edit the struct tag or add manual spec override
2. Run `make generate` to regenerate `.hcl2spec.go` files
3. Run `go build ./...` to verify compilation
4. Run `make plugin-check` to verify plugin compatibility
5. Test with the example config above using `packer validate`

## Acceptance Criteria

- [ ] `packer validate` accepts nested `navigator_config` structures (no type mismatch errors)
- [ ] `go build ./...` completes successfully
- [ ] `make plugin-check` passes
- [ ] The fix applies to both `ansible-navigator` and `ansible-navigator-local` provisioners
- [ ] No breaking changes to existing flat-map usage of `navigator_config`
- [ ] Generated HCL2 spec files are committed (not accidentally ignored)

## Constraints & Assumptions

**Assumptions:**

- The YAML generation logic in `navigator_config.go` is working correctly and doesn't need changes
- Test files in `navigator_config_test.go` demonstrate the expected Go-level behavior
- The issue is purely in the HCL→Go type mapping, not in the YAML output

**Constraints:**

- Must use packer-sdc code generation where possible (don't hand-write entire `.hcl2spec.go` files)
- Must maintain backward compatibility with any existing configs that might use flat string maps
- Changes must pass existing test suite
- Must follow packer-plugin-sdk conventions for custom type specs

## Related Files

Primary files to modify:

- `provisioner/ansible-navigator/provisioner.go` (struct definition or HCL2Spec method)
- `provisioner/ansible-navigator-local/provisioner.go` (struct definition or HCL2Spec method)

Generated files (will be updated by `make generate`):

- `provisioner/ansible-navigator/provisioner.hcl2spec.go`
- `provisioner/ansible-navigator-local/provisioner.hcl2spec.go`

Reference files (demonstrate correct behavior):

- `provisioner/ansible-navigator/navigator_config_test.go` (shows nested maps work in Go)
- `provisioner/ansible-navigator/navigator_config.go` (processes nested maps correctly)

## Additional Context

This issue surfaced when a user tried to use the documented `navigator_config` feature with nested configuration like:

```hcl
navigator_config = {
  execution-environment = {
    enabled = true
    image = "..."
  }
}
```

Packer rejected it with:

```
Incorrect attribute value type; Inappropriate value for attribute "navigator_config": 
element "execution-environment": string required.
```

The Go code and tests show this should work - it's purely an HCL2 spec generation issue.
