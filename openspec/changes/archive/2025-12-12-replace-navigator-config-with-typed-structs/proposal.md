# Proposal: Replace navigator_config with Typed Structs

## Change ID

`replace-navigator-config-with-typed-structs`

## Problem Statement

The current implementation of the `navigator_config` field uses `map[string]interface{}` with `cty.DynamicPseudoType` in the HCL2 spec. This approach causes the plugin to crash during initialization because `cty.DynamicPseudoType` cannot be serialized over Packer's gRPC plugin RPC protocol.

The error manifests as:

```
unsupported cty.Type conversion from cty.pseudoTypeDynamic
```

This prevents the plugin from being usable when users attempt to configure ansible-navigator through the `navigator_config` field.

## Why

This change is necessary because:

1. **Plugin crashes**: The current `cty.DynamicPseudoType` approach causes RPC serialization failures, making the plugin non-functional
2. **No type safety**: Users get no validation or autocomplete for navigator_config
3. **Poor debugging**: Errors are cryptic and occur at runtime rather than during validation
4. **Against best practices**: Packer SDK documentation recommends typed structs over dynamic types

## What Changes

Replace the `map[string]interface{}` + `cty.DynamicPseudoType` approach with explicit, strongly-typed Go structs that:

1. Can be properly serialized over Packer's gRPC RPC protocol
2. Provide type safety and validation
3. Enable IDE autocomplete and early error detection
4. Support the full ansible-navigator.yml configuration schema
5. Maintain the same YAML generation behavior

## Proposed Solution

Define explicit Go struct types that mirror the ansible-navigator.yml schema structure:

- `NavigatorConfig` - Root configuration container
- `ExecutionEnvironment` - Execution environment settings
- `EnvironmentVariablesConfig` - Environment variable configuration
- `AnsibleConfig` - Ansible-specific configuration
- `AnsibleConfigInner` - Inner config section
- `AnsibleConfigDefaults` - Ansible defaults
- `AnsibleConfigConnection` - SSH connection settings  
- `LoggingConfig` - Logging configuration
- `PlaybookArtifact` - Playbook artifact settings
- `CollectionDocCache` - Collection documentation cache settings

These structs will:

- Use proper `mapstructure` tags for HCL parsing
- Include all commonly-used ansible-navigator.yml options
- Be included in the `go:generate` directive for HCL2 spec generation
- Replace the current `map[string]interface{}` field in the `Config` struct
- Generate RPC-safe HCL2 specs using concrete cty types

## User Experience

### Before (Current - Crashes)

```hcl
provisioner "ansible-navigator" {
  navigator_config = {
    mode = "stdout"
    execution-environment = {
      enabled = true
      image   = "quay.io/ansible/creator-ee:latest"
    }
  }
}
# Results in: unsupported cty.Type conversion error
```

### After (Proposed - Works)

```hcl
provisioner "ansible-navigator" {
  navigator_config {
    mode = "stdout"
    
    execution_environment {
      enabled     = true
      image       = "quay.io/ansible/creator-ee:latest"
      pull_policy = "missing"
      
      environment_variables {
        ANSIBLE_REMOTE_TMP = "/tmp/.ansible/tmp"
      }
    }
    
    ansible_config {
      defaults {
        remote_tmp = "/tmp/.ansible/tmp"
      }
    }
  }
}
# Works correctly with native HCL blocks
```

## Implementation Impact

### Files Modified

- `provisioner/ansible-navigator/provisioner.go` - Add struct definitions, update Config, update go:generate
- `provisioner/ansible-navigator/provisioner.hcl2spec.go` - Regenerated via `make generate`
- `provisioner/ansible-navigator-local/provisioner.go` - Add struct definitions, update Config, update go:generate
- `provisioner/ansible-navigator-local/provisioner.hcl2spec.go` - Regenerated via `make generate`
- `example/*.pkr.hcl` - Update examples to use native HCL block syntax

### Breaking Changes

This is a **breaking change** requiring:

- HCL syntax change from map assignment to block syntax
- Field name changes (hyphens to underscores for HCL compatibility)
- Minor version bump (e.g., 0.x.0 → 0.y.0)

### Migration Path

Users will need to:

1. Update `navigator_config = { ... }` to `navigator_config { ... }` (block syntax)
2. Replace hyphenated keys with underscored keys (e.g., `execution-environment` → `execution_environment`)
3. Ensure nested structures use block syntax where appropriate

## Non-Goals

- Changing the underlying YAML generation logic
- Modifying ansible-navigator execution behavior  
- Adding new ansible-navigator features
- Supporting legacy string-based config formats
- Changing other provisioner configuration fields

## Alternatives Considered

### 1. Use YAML string field

**Approach**: Change `navigator_config` to accept a YAML string
**Pros**: Quick fix, no struct definitions needed
**Cons**: No type safety, no validation, poor UX, no autocomplete

### 2. Keep DynamicPseudoType with workaround

**Approach**: Attempt to implement custom serialization for DynamicPseudoType
**Pros**: Maintains current HCL syntax
**Cons**: Not supported by Packer SDK, would require forking/patching SDK

### 3. Multi-phase migration

**Approach**: Support both approaches during transition period
**Pros**: Smoother migration for users
**Cons**: Increased complexity, more code to maintain, delayed cleanup

**Recommendation**: Use the typed struct approach (proposed solution) as it provides the best long-term developer experience and aligns with Packer plugin best practices.

## Success Criteria

- [ ] All struct types defined with proper mapstructure tags
- [ ] Config struct updated to use `*NavigatorConfig` type
- [ ] go:generate directive includes all new types
- [ ] `make generate` completes successfully for both provisioners
- [ ] `go build ./...` completes without errors
- [ ] Plugin binary passes `describe` test without crashes
- [ ] `packer validate` succeeds with example HCL using typed navigator_config
- [ ] Example files demonstrate new syntax
- [ ] Both provisioners (ansible-navigator and ansible-navigator-local) support the new structs

## Related Work

- Previous attempt: `fix-navigator-config-hcl2-type` (archived) - Attempted manual override with DynamicPseudoType
- Root cause analysis: `DIAGNOSIS.md` in test directory
- Solution exploration: `PACKER_FRIENDLY_FORMATS.md` in test directory
