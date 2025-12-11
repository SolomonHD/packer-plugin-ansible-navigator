# Design: Navigator Config Surface Refactoring

## Context

The plugin currently supports ~40 configuration options that attempt to control ansible-navigator behavior through multiple mechanisms (CLI flags, environment variables, ansible.cfg files). This complexity arose from trying to support both traditional ansible-navigator usage and Execution Environment (EE) container-based workflows. However, only `ansible-navigator.yml` reliably controls EE container behavior, making most other configuration mechanisms unreliable.

**Key constraint**: ansible-navigator v3+ prioritizes configuration from `ansible-navigator.yml` when present, and this is the only mechanism that reliably propagates settings into EE containers.

## Goals

- **Primary**: Simplify configuration surface by replacing ~40 options with a single `navigator_config` option
- **Reliability**: Use ansible-navigator.yml as the single source of truth for all ansible-navigator settings
- **Usability**: Provide automatic defaults for common EE scenarios (temp directory permissions)
- **Maintainability**: Reduce code complexity by eliminating redundant configuration paths

## Non-Goals

- Backward compatibility with removed options (breaking change is acceptable)
- Supporting both new and old configuration simultaneously
- Implementing configuration translation/migration tooling

## Decisions

### Decision 1: Use HCL map directly for navigator_config

**Choice**: Accept `navigator_config` as `map[string]interface{}` and serialize directly to YAML.

**Rationale**:

- ansible-navigator.yml has a complex, nested structure that maps naturally to HCL's nested map syntax
- Using a flat struct would require dozens of Go fields mirroring ansible-navigator's schema
- HCL maps can express the full ansible-navigator.yml structure without Go code changes for new options
- Simplifies maintenance as ansible-navigator adds new config options

**Alternative considered**: Create typed structs for navigator_config

- Rejected: Would require updating Go structs every time ansible-navigator adds config options
- Rejected: Would limit flexibility for users who want to use newer ansible-navigator features

**Implementation**:

```go
type Config struct {
    // ...
    NavigatorConfig map[string]interface{} `mapstructure:"navigator_config" required:"false"`
    // ...
}
```

### Decision 2: Auto-apply EE defaults only when execution-environment.enabled = true

**Choice**: When `navigator_config.execution-environment.enabled = true` AND user hasn't specified temp dir settings, automatically set:

```yaml
ansible:
  config:
    defaults:
      remote_tmp: "/tmp/.ansible/tmp"
      local_tmp: "/tmp/.ansible-local"
execution-environment:
  environment-variables:
    ANSIBLE_REMOTE_TMP: "/tmp/.ansible/tmp"
    ANSIBLE_LOCAL_TMP: "/tmp/.ansible-local"
```

**Rationale**:

- "Permission denied: /.ansible/tmp" is the most common EE failure
- Non-root container users can't write to `/.ansible`
- Setting these defaults eliminates 90% of EE troubleshooting issues
- User-specified values always take precedence (explicit > defaults)

**Alternative considered**: Require users to always specify temp dirs

- Rejected: Poor user experience; makes EE workflows harder than necessary
- Rejected: Doesn't align with "batteries included" philosophy

### Decision 3: Use ANSIBLE_NAVIGATOR_CONFIG environment variable

**Choice**: Point ansible-navigator at generated config via `ANSIBLE_NAVIGATOR_CONFIG=/path/to/file` environment variable.

**Rationale**:

- Supported by ansible-navigator v3+
- Works for both local execution and EE containers
- Avoids needing to pass `--config` CLI flag (cleaner command construction)
- Environment variables propagate reliably into containers

**Alternative considered**: Use `--config` CLI flag

- Also viable, but environment variable is cleaner
- Avoids adding to already-long command lines

### Decision 4: Generate config in system temp directory with cleanup

**Choice**:

- Generate: `/tmp/packer-navigator-cfg-<uuid>.yml` (or OS temp equivalent)
- Upload to target's staging directory for local provisioner
- Clean up in deferred function after execution (success or failure)

**Rationale**:

- System temp directory is always writable
- UUID prevents collisions between concurrent Packer runs
- Deferred cleanup ensures no leftover files
- Staging directory for local provisioner keeps files together

**Implementation**:

```go
func (p *Provisioner) generateNavigatorConfig() (string, error) {
    tmpFile, err := os.CreateTemp("", "packer-navigator-cfg-*.yml")
    if err != nil {
        return "", err
    }
    defer tmpFile.Close()
    
    config := p.applyEEDefaults(p.config.NavigatorConfig)
    yamlData, err := yaml.Marshal(config)
    if err != nil {
        return "", err
    }
    
    if _, err := tmpFile.Write(yamlData); err != nil {
        return "", err
    }
    
    return tmpFile.Name(), nil
}
```

### Decision 5: Remove all legacy options without migration shims

**Choice**: Complete removal of ~40 options with no backward compatibility layer.

**Rationale**:

- Plugin is under active development (per project.md: "backward compatibility is not a goal")
- Maintaining dual configuration paths increases complexity
- No versioned releases yet that users depend on
- Clean break simplifies codebase and mental model

**Impact**:

- All existing configurations using removed options will fail validation
- Error messages should guide users to navigator_config equivalents
- Documentation must clearly mark breaking change

## Implementation Approach

### Phase 1: Config struct and generation

1. Add `NavigatorConfig map[string]interface{}` to both provisioners
2. Remove legacy config fields
3. Implement YAML generation with EE defaults logic
4. Add validation for navigator_config field

### Phase 2: Execution integration

1. Generate temp config file during Prepare() phase
2. Set ANSIBLE_NAVIGATOR_CONFIG environment variable
3. Remove legacy CLI flag generation (`--mode`, `--ee`, `--eei`)
4. Implement cleanup in deferred functions

### Phase 3: Testing and documentation

1. Update unit tests to use navigator_config
2. Add tests for YAML generation and defaults
3. Update all documentation
4. Verify plugin-check passes

## Risks / Trade-offs

### Risk: Breaking all existing configurations

**Mitigation**:

- Clear migration guide in documentation
- Detailed error messages pointing to navigator_config
- Example configurations in README showing new approach

### Risk: Users need to understand ansible-navigator.yml schema

**Mitigation**:

- Provide comprehensive examples of common scenarios
- Document EE automatic defaults clearly
- Reference official ansible-navigator documentation

### Risk: YAML serialization edge cases

**Mitigation**:

- Use standard `gopkg.in/yaml.v3` library
- Test with complex nested structures
- Document any known limitations

### Trade-off: Less type safety vs. more flexibility

**Accept**: Using `map[string]interface{}` means less Go type checking, but enables supporting full ansible-navigator.yml schema without code changes.

## Migration Plan

### For users updating from previous versions

**Step 1**: Identify removed options in your configuration

- `ansible_cfg`, `ansible_env_vars`, `execution_environment`, `navigator_mode`, etc.

**Step 2**: Map to navigator_config structure

- Execution environment: `execution-environment.enabled`, `execution-environment.image`
- Mode: `mode`
- Ansible config: `ansible.config.defaults.*`
- Environment variables: `execution-environment.environment-variables`

**Step 3**: Test with new configuration

- Verify ansible-navigator.yml generation by checking logs
- Confirm EE containers work without permission errors
- Validate all plays execute correctly

**No rollback plan**: This is a one-way breaking change. Users must migrate forward.

## Open Questions

- **Q**: Should we validate navigator_config contents against ansible-navigator's schema?
  - **A**: No. Let ansible-navigator validate its own config. We just generate YAML and pass it through. This keeps us decoupled from ansible-navigator's versioning.

- **Q**: Should we support both string and object for execution_environment during transition?
  - **A**: No. Clean break. Remove string form entirely and only support `navigator_config.execution-environment` object.

- **Q**: What happens if user specifies both old and new config?
  - **A**: Not applicable. Old config options are completely removed. Any attempt to use them will fail HCL parsing/validation.
