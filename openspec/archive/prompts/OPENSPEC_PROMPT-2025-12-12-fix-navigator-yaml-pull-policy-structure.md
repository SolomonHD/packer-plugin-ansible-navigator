# OpenSpec Change Prompt

## Context

The packer-plugin-ansible-navigator plugin generates temporary `ansible-navigator.yml` configuration files from HCL user input. The YAML generation code in `provisioner/ansible-navigator/navigator_config.go` (specifically the `convertToYAMLStructure()` function) currently outputs an incorrect YAML structure that is incompatible with ansible-navigator 25.x.

**Current bug**: Line 83 outputs `pull-policy` as a flat field:

```yaml
execution-environment:
  pull-policy: missing  # ❌ Wrong - causes validation error
```

**Expected structure** (ansible-navigator 25.x):

```yaml
execution-environment:
  pull:
    policy: missing  # ✅ Correct nested structure
```

This causes ansible-navigator to reject the generated config file with:

```
Error: Additional properties are not allowed ('pull-policy' was unexpected).
```

When running from Packer (non-TTY), this validation error causes ansible-navigator to hang or fail silently, blocking the entire Packer build.

## Goal

Fix the YAML generation logic in `navigator_config.go` to produce ansible-navigator.yml compatible with the **latest ansible-navigator syntax** (v24+/v25+). Specifically:

1. Verify the current field mapping strategy against latest ansible-navigator documentation/schema
2. Fix the `pull-policy` mapping to use the correct nested `pull: { policy: ... }` structure
3. Check for any other fields that may have similar flat-vs-nested mismatches
4. Ensure generated YAML passes ansible-navigator's built-in validation

## Scope

**In scope:**

- Fix `convertToYAMLStructure()` function in `provisioner/ansible-navigator/navigator_config.go`
- Verify all execution-environment related fields against latest ansible-navigator schema
- Update YAML field mapping to match nested structure where required
- Test that generated YAML validates with `ansible-navigator settings --schema`
- Apply same fixes to `provisioner/ansible-navigator-local/navigator_config.go` (parallel implementation)

**Out of scope:**

- Changes to the Go struct definitions (those are fine - they use mapstructure for HCL parsing)
- Changes to HCL2 spec generation
- Changes to the plugin's HCL configuration syntax (user-facing config stays the same)
- Adding new configuration options beyond fixing the YAML output format
- WSL2/Docker networking configuration (different issue)

## Desired Behavior

After the fix:

1. User configures in HCL:

   ```hcl
   navigator_config {
     execution_environment {
       enabled     = true
       image       = "myimage:latest"
       pull_policy = "missing"
     }
   }
   ```

2. Plugin generates valid YAML:

   ```yaml
   ansible-navigator:
     execution-environment:
       enabled: true
       image: "myimage:latest"
       pull:
         policy: missing
   ```

3. ansible-navigator accepts the config without validation errors

4. Packer builds complete successfully when using execution environments

## Constraints & Assumptions

- **Assumption**: The bug is in the YAML generation, not the struct definitions
- **Assumption**: ansible-navigator 24.x and 25.x use the nested `pull: { policy, arguments }` structure (verify this)
- **Constraint**: Must maintain backwards compatibility with the HCL configuration syntax users are already using
- **Constraint**: Must generate YAML that wraps fields under `ansible-navigator:` root key
- **Constraint**: Solution must work for both `provisioner/ansible-navigator` and `provisioner/ansible-navigator-local`

## Reference Information

- **File to fix**: `/home/solomong/dev/packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go`
- **Parallel file**: `/home/solomong/dev/packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator-local/navigator_config.go`
- **Function**: `convertToYAMLStructure()` around line 69-176
- **Bug location**: Line 83 (`eeMap["pull-policy"] = config.ExecutionEnvironment.PullPolicy`)

### Verification command

After fixing, test the generated YAML:

```bash
ansible-navigator settings --sample | grep -A10 "execution-environment"
```

Check that policy is nested under `pull:`:

```yaml
execution-environment:
  pull:
    policy: tag
    arguments:
      - "--tls-verify=false"
```

## Acceptance Criteria

- [ ] `convertToYAMLStructure()` outputs `pull.policy` nested structure, not flat `pull-policy`
- [ ] Generated YAML contains `ansible-navigator:` root key wrapping all fields
- [ ] Generated YAML passes ansible-navigator's validation (no "Additional properties" errors)
- [ ] All other execution-environment fields verified against ansible-navigator 25.x schema
- [ ] Changes applied to both `ansible-navigator` and `ansible-navigator-local` provisioners
- [ ] Packer build with execution-environment enabled completes without hanging
- [ ] No changes required to user-facing HCL configuration syntax
