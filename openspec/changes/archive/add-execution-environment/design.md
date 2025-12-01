# Design: Execution Environment Support

## Context

Ansible Navigator is designed to run Ansible playbooks inside containerized execution environments, which provide:
- Consistent Ansible versions and dependencies
- Isolated execution with predictable behavior
- Portable environments that work across different systems
- Pre-packaged collections and Python dependencies

The packer-plugin-ansible-navigator currently does not expose ansible-navigator's `--execution-environment` flag, limiting users to the default execution environment selection.

## Goals / Non-Goals

### Goals
- Enable users to specify custom execution environment container images
- Maintain full backward compatibility with existing templates
- Follow ansible-navigator's command-line interface conventions
- Provide clear documentation with practical examples

### Non-Goals
- Building or managing execution environment images (users are responsible for this)
- Validating that specified images exist in registries (ansible-navigator handles this)
- Supporting execution environment image building within Packer (out of scope)
- Managing container registry authentication (handled by ansible-navigator/container runtime)

## Decisions

### Decision 1: Configuration Field Name
**Choice**: Use `execution_environment` as the field name.

**Rationale**: 
- Matches ansible-navigator's terminology and flag name (`--execution-environment`)
- Aligns with Ansible's official naming conventions
- Clear and self-documenting for users familiar with ansible-navigator

**Alternatives considered**:
- `container_image`: Less specific, could be confused with other container operations
- `ee_image`: Too abbreviated, not self-documenting
- `ansible_ee`: Redundant prefix in the context of an ansible provisioner

### Decision 2: Field Type
**Choice**: String field accepting full container image references.

**Rationale**:
- Simple and flexible - accepts any valid container image reference
- Matches how users typically specify container images (registry/image:tag)
- No need for complex validation - ansible-navigator will handle invalid images
- Allows for pull-by-digest formats (`image@sha256:...`) if needed

**Type signature**:
```go
ExecutionEnvironment string `mapstructure:"execution_environment"`
```

### Decision 3: Command Argument Placement
**Choice**: Place `--execution-environment` flag after `--mode` but before inventory/playbook arguments.

**Command structure**:
```bash
ansible-navigator run --mode <mode> --execution-environment <image> [other-args] -i <inventory> <playbook>
```

**Rationale**:
- Follows ansible-navigator's conventions for option ordering
- Places execution environment specification early in the command (logical ordering)
- Keeps inventory and playbook as final positional arguments (ansible convention)

**Implementation location**: Modify `createCmdArgs()` and `executePlays()`/`executeSinglePlaybook()` methods.

### Decision 4: Default Behavior
**Choice**: When field is empty/unset, do not pass the `--execution-environment` flag.

**Rationale**:
- Preserves ansible-navigator's default behavior
- Maximum backward compatibility
- Users who don't need this feature are not affected
- ansible-navigator has intelligent defaults for execution environment selection

### Decision 5: Documentation Strategy
**Choice**: Document the field inline with code comments and in all user-facing documentation.

**Locations**:
1. Inline Go documentation comment in Config struct
2. Configuration reference documentation
3. Examples documentation with practical use cases
4. README with mention in features section
5. Provisioner-specific MDX documentation

**Example documentation text**:
```
// The container image to use as the execution environment for ansible-navigator.
// Specifies which containerized environment runs the Ansible playbooks.
// When unset, ansible-navigator uses its default execution environment.
// Examples: "quay.io/ansible/creator-ee:latest", "my-registry.io/custom-ee:v1.0"
ExecutionEnvironment string `mapstructure:"execution_environment"`
```

## Risks / Trade-offs

### Risk: Image Availability
- **Risk**: Specified image might not be available in the container registry
- **Mitigation**: ansible-navigator will fail with clear error message; no additional validation needed

### Risk: Image Compatibility
- **Risk**: Custom execution environment might not have required collections/roles
- **Mitigation**: Users are responsible for ensuring their EE images contain necessary dependencies; document this clearly

### Trade-off: No Image Building
- **Trade-off**: Users must build/maintain execution environment images separately
- **Rationale**: Image building is a complex, separate concern best handled by dedicated tools (ansible-builder)
- **Documentation**: Provide links to ansible-builder and execution environment creation guides

## Migration Plan

### Phase 1: Implementation
1. Add the configuration field to Config struct
2. Update command argument generation logic
3. Add unit tests for the new functionality
4. Run `go generate` to update HCL2 specs

### Phase 2: Documentation
1. Update all documentation files
2. Add practical examples showing common use cases
3. Include troubleshooting section for common EE issues

### Phase 3: Testing
1. Test with official Ansible execution environments
2. Test with custom execution environments
3. Verify backward compatibility with existing templates
4. Test error handling when image is unavailable

### Rollback Strategy
- If issues arise, the field can be deprecated without removing it
- Existing templates without the field are unaffected
- No database migrations or data changes required

## Open Questions

1. **Q**: Should we validate the image reference format?
   **A**: No - ansible-navigator will validate and provide clear errors if invalid

2. **Q**: Should we support execution environment environment variables?
   **A**: Defer to future enhancement - ansible-navigator has `--execution-environment-volume-mounts` and `--container-options` flags that could be added later

3. **Q**: Should we provide a list of recommended execution environments?
   **A**: Yes in documentation - link to official Ansible EE images and community recommendations