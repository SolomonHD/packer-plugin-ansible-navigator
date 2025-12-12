# Design: Fix Navigator Config HCL2 Struct Type Mapping

## Context

The packer-plugin-ansible-navigator uses typed Go structs for `navigator_config` to avoid RPC serialization issues with `map[string]interface{}` and `cty.DynamicPseudoType`. However, certain mapstructure patterns don't translate correctly through `packer-sdc mapstructure-to-hcl2`.

### Current Issues

1. **`EnvironmentVariablesConfig` with `mapstructure:",remain"`**:

   ```go
   type EnvironmentVariablesConfig struct {
       Variables map[string]string `mapstructure:",remain"`
   }
   ```

   The `,remain` tag is meant to capture all extra fields into the Variables map. However, the HCL2 generator creates:

   ```go
   func (*FlatEnvironmentVariablesConfig) HCL2Spec() map[string]hcldec.Spec {
       s := map[string]hcldec.Spec{
           "variables": &hcldec.AttrSpec{Name: "variables", Type: cty.Map(cty.String), Required: false},
       }
       return s
   }
   ```

   This requires `environment_variables { variables = { KEY = "value" } }` syntax.

2. **`AnsibleConfig` with `mapstructure:",squash"`**:

   ```go
   type AnsibleConfig struct {
       Config string `mapstructure:"config"`
       Inner  *AnsibleConfigInner `mapstructure:",squash"`
   }
   ```

   The `,squash` causes Inner's fields to be "bubbled up", but HCL2Spec only generates `config`:

   ```go
   func (*FlatAnsibleConfig) HCL2Spec() map[string]hcldec.Spec {
       s := map[string]hcldec.Spec{
           "config": &hcldec.AttrSpec{Name: "config", Type: cty.String, Required: false},
       }
       return s
   }
   ```

   The `defaults` and `ssh_connection` blocks are lost.

## Goals / Non-Goals

### Goals

- Fix HCL parsing so all `navigator_config` nested blocks work correctly
- Maintain RPC serialization safety (no `cty.DynamicPseudoType`)
- Follow existing project conventions for typed structs
- Ensure `packer validate` passes for all example configs
- Ensure YAML generation to ansible-navigator.yml continues working

### Non-Goals

- Supporting arbitrary/unknown ansible-navigator.yml keys (users must update plugin for new options)
- Backward compatibility with previous syntax (breaking change is acceptable per project.md)
- Manual editing of `.hcl2spec.go` files (prefer generation)

## Decisions

### Decision 1: Replace `mapstructure:",remain"` with Explicit Fields

**Choice**: Define explicit fields for environment variables instead of using `",remain"`.

**Rationale**:

- The `",remain"` tag does not translate to HCL2 spec generation
- Explicit fields provide IDE autocomplete and type checking
- Common environment variables are well-known and can be enumerated

**Alternative considered**: Override HCL2Spec() to use `hcldec.BlockAttrsSpec`

- Rejected: Would require manual editing of generated files or complex custom code

**New struct design**:

```go
type EnvironmentVariablesConfig struct {
    // Pass-through environment variables (list of var names to pass from host)
    Pass []string `mapstructure:"pass"`
    // Set environment variables (explicit key-value pairs)
    Set map[string]string `mapstructure:"set"`
}
```

This matches the actual ansible-navigator.yml structure:

```yaml
execution-environment:
  environment-variables:
    pass:
      - SSH_AUTH_SOCK
    set:
      ANSIBLE_REMOTE_TMP: "/tmp/.ansible/tmp"
```

### Decision 2: Remove `mapstructure:",squash"` from AnsibleConfig

**Choice**: Replace squash with explicit nested block.

**Rationale**:

- The `,squash` tag causes HCL2 generator to lose nested structure
- Explicit nesting provides clearer HCL block syntax

**Alternative considered**: Flatten all fields into AnsibleConfig

- Rejected: Would lose the logical grouping of defaults vs ssh_connection

**New struct design**:

```go
type AnsibleConfig struct {
    // Path to ansible.cfg file
    Config string `mapstructure:"config"`
    // Defaults section
    Defaults *AnsibleConfigDefaults `mapstructure:"defaults"`
    // SSH connection section  
    SSHConnection *AnsibleConfigConnection `mapstructure:"ssh_connection"`
}
```

Remove `AnsibleConfigInner` entirely.

### Decision 3: Update go:generate Directive

**Choice**: Ensure all relevant types are listed in the directive.

Updated directive:

```go
//go:generate packer-sdc mapstructure-to-hcl2 -type Config,Play,PathEntry,NavigatorConfig,ExecutionEnvironment,EnvironmentVariablesConfig,AnsibleConfig,AnsibleConfigDefaults,AnsibleConfigConnection,LoggingConfig,PlaybookArtifact,CollectionDocCache
```

Note: `AnsibleConfigInner` is removed from the list since it no longer exists.

## Risks / Trade-offs

### Risk 1: Breaking Change for Users

**Mitigation**:

- Document in MIGRATION.md
- Provide example files showing new syntax
- Error messages should guide users

### Risk 2: YAML Generation May Need Updates

**Mitigation**:

- Test YAML generation with new struct layout
- Update `generateNavigatorConfigYAML()` if needed to match ansible-navigator.yml schema

### Risk 3: IDE Autocomplete Changes

**Impact**: Users' IDE suggestions will change
**Mitigation**: Clear documentation and examples

## Migration Plan

1. Update struct definitions in both provisioners
2. Remove `AnsibleConfigInner` type
3. Run `make generate`
4. Update YAML generation logic if needed
5. Update example files
6. Update MIGRATION.md

### Rollback

If issues arise, revert to previous struct definitions. The breaking change is intentional and acceptable per project.md.

## Open Questions

1. Should we add more explicit fields to `EnvironmentVariablesConfig.Set` for common variables like `ANSIBLE_REMOTE_TMP`?
   - **Decision**: No, keep it as a generic map. Users can set any variables.

2. Should we validate that `Set` keys are valid environment variable names?
   - **Decision**: No, defer validation to ansible-navigator.
