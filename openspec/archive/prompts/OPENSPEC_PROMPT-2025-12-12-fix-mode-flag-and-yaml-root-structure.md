# OpenSpec Change Prompt

## Context

The packer-plugin-ansible-navigator has two related bugs causing ansible-navigator to hang during execution:

1. **Missing `--mode` CLI flag** - The plugin generates a correct `ansible-navigator.yml` config file and sets the `ANSIBLE_NAVIGATOR_CONFIG` environment variable, but doesn't pass the `--mode` CLI flag. This causes ansible-navigator to default to interactive mode and hang waiting for terminal input.

2. **Incorrect YAML root structure** - The generated YAML config places settings at the root level instead of nested under the required `ansible-navigator:` key, causing validation errors in ansible-navigator 25.x.

## Goal

Fix both issues so that ansible-navigator executes in non-interactive mode with valid configuration.

## Scope

**In scope:**

- Add `--mode` CLI flag when NavigatorConfig.Mode is set
- Wrap all YAML settings under `ansible-navigator:` root key
- Document that users should use direct ansible-navigator binary paths (not asdf shims)
- Maintain backward compatibility with existing configs
- Ensure proper execution environment container behavior

**Out of scope:**

- Changing the HCL configuration format visible to users
- Adding new configuration options
- Modifying galaxy/collection installation logic
- Changes to the SSH proxy adapter

## Desired Behavior

### 1. CLI Command Generation

When `navigator_config.mode = "stdout"` is specified, the generated command should include `--mode stdout`:

**Before:**

```bash
ansible-navigator run -e packer_build_name="test" -i inventory playbook.yml
# Hangs in interactive mode
```

**After:**

```bash
ansible-navigator run --mode stdout -e packer_build_name="test" -i inventory playbook.yml
# Executes and outputs to stdout
```

### 2. YAML Config Structure

The generated `ansible-navigator.yml` must have settings nested under `ansible-navigator:`  key:

**Before (invalid):**

```yaml
mode: stdout
execution-environment:
  enabled: true
  image: "my-ee:latest"
logging:
  level: debug
```

**After (valid):**

```yaml
ansible-navigator:
  mode: stdout
  execution-environment:
    enabled: true
    image: "my-ee:latest"
  logging:
    level: debug
```

### 3. Documentation Update

Add a troubleshooting section about asdf shim recursion issue and recommend using direct binary paths.

## Constraints & Assumptions

### Assumptions

- Users have ansible-navigator 25.x installed (current version)
- The `ansible-navigator.yml` schema requires the `ansible-navigator:` root key in v25.x+
- CLI flags take precedence over config file settings  
- The plugin already correctly handles `ANSIBLE_NAVIGATOR_CONFIG` environment variable
- Most asdf installations work correctly, but some configurations cause shim recursion loops

### Constraints

- Must not break existing HCL configurations
- Changes should be minimal and surgical
- Must maintain RPC serializability (no `map[string]interface{}` with `cty.DynamicPseudoType`)
- Documentation changes should be clear about when direct binary paths are needed

## Acceptance Criteria

- [x] `--mode` flag is added to ansible-navigator command when `navigator_config.mode` is set
- [ ] Generated YAML has `ansible-navigator:` root key wrapping all settings
- [ ] Validation errors about "unexpected properties" are resolved
- [ ] ansible-navigator executes without hanging in test scenarios
- [ ] Existing HCL configs continue to work without modification
- [ ] Documentation warns about asdf shim recursion and provides workaround
- [ ] All four verification commands pass: `make generate && go build ./... && go test ./... && make plugin-check`

## Technical Notes

### Files Modified

1. **`provisioner/ansible-navigator/provisioner.go`** (lines ~1155-1163)
   - Add conditional logic to prepend `--mode` flag when NavigatorConfig.Mode is set

2. **`provisioner/ansible-navigator/navigator_config.go`** (lines ~69-178)
   - Modify `convertToYAMLStructure()` to wrap settings in `ansible-navigator:` root key

3. **`README.md` or `docs/troubleshooting.md`**
   - Add section about asdf shim recursion issue
   - Recommend using absolute paths like `/home/user/.asdf/installs/python/x.y.z/bin/ansible-navigator`

### Root Cause Analysis

- **Hang Issue #1:** ansible-navigator defaults to interactive mode when `--mode` is not specified on CLI, regardless of config file
- **Hang Issue #2:** asdf shims can create recursive execution loops in some configurations (not a plugin bug)
- **Validation Error:** ansible-navigator v25.x schema validation requires `ansible-navigator:` root key in config YAML

### Testing Strategy

1. Build plugin with fixes
2. Test with `null` builder using simple playbook (no AWS deps)
3. Verify `--mode stdout` appears in command output
4. Verify generated YAML has correct structure
5. Confirm ansible-navigator executes without hanging
6. Test with AWS builder using role FQDN

## Related Issues/PRs

- Related to existing OpenSpec change: `fix-navigator-yaml-pull-policy-structure`
- Both address YAML schema compliance for ansible-navigator v25.x
