# OpenSpec Change Prompt

## Context

The packer-plugin-ansible-navigator currently crashes when users try to use the `navigator_config` field because it uses `cty.DynamicPseudoType`, which cannot be serialized over Packer's gRPC plugin RPC protocol. The error occurs during plugin initialization when Packer tries to serialize the HCL2 spec for transmission to the core process.

The `navigator_config` field is meant to provide a modern, declarative way to configure ansible-navigator through a generated `ansible-navigator.yml` file, supporting the full ansible-navigator.yml schema including execution environments, logging settings, and Ansible configuration overrides.

## Goal

Replace the current `map[string]interface{}` + `cty.DynamicPseudoType` approach with explicit Go structs that can be properly serialized over Packer's plugin RPC protocol. This enables users to configure ansible-navigator using native HCL blocks with full type safety and IDE autocomplete support.

## Scope

**In scope:**

- Define explicit Go structs for NavigatorConfig and its nested components (ExecutionEnvironment, AnsibleConfig, LoggingConfig, etc.)
- Update the Config struct to use the new NavigatorConfig type
- Add new types to the go:generate directive for HCL2 spec generation
- Update both provisioner implementations (ansible-navigator and ansible-navigator-local) to use the new structs
- Run make generate to create RPC-safe HCL2 specs
- Update example files to show native HCL block syntax
- Verify the plugin builds, describes, and validates without crashes
- Document the change in migration notes or changelog

**Out of scope:**

- Changing the underlying YAML generation logic (stays the same)
- Modifying how ansible-navigator is executed (no behavioral changes)
- Adding new ansible-navigator features beyond what's already planned
- Changing other unrelated provisioner fields
- Supporting legacy string-based config formats (can be future enhancement)

## Desired Behavior

After this change:

1. Users write native HCL blocks for navigator_config:

```hcl
provisioner "ansible-navigator" {
  navigator_config {
    mode = "stdout"
    
    execution-environment {
      enabled = true
      image   = "my-image:latest"
      pull-policy = "missing"
      
      environment-variables {
        set = {
          ANSIBLE_REMOTE_TMP = "/tmp/.ansible/tmp"
        }
      }
    }
    
    ansible {
      collections = ["integration.common_tasks"]
    }
  }
}
```

2. Packer successfully initializes the plugin without crashes
3. The plugin generates the same `ansible-navigator.yml` file as before
4. HCL2 specs properly serialize over RPC using concrete types
5. Users get IDE autocomplete and early validation for configuration errors

## Constraints & Assumptions

**Constraints:**

- Must use only RPC-serializable cty types (no `cty.DynamicPseudoType`)
- Must maintain backward compatibility where reasonable (existing examples should still work pattern-wise)
- Must follow Packer plugin SDK conventions and x5 API requirements
- Struct definitions must match ansible-navigator.yml schema structure

**Assumptions:**

- ansible-navigator.yml schema is relatively stable and well-defined
- The existing YAML generation code can work with struct-based config
- Users prefer native HCL syntax over string-encoded YAML
- The change will require a minor version bump (breaking change)
- The `generateNavigatorConfigYAML` and `createNavigatorConfigFile` functions can already handle the nested map structure

## Acceptance Criteria

- [ ] NavigatorConfig, ExecutionEnvironment, EnvironmentVariablesConfig, AnsibleConfig, and LoggingConfig structs defined in provisioner.go
- [ ] Config struct updated to use `*NavigatorConfig` instead of `map[string]interface{}`
- [ ] go:generate directive includes all new types
- [ ] make generate runs successfully for both provisioner implementations
- [ ] go build ./... completes without errors
- [ ] Plugin binary passes `./packer-plugin-ansible-navigator describe` test
- [ ] packer validate succeeds with example HCL using native navigator_config blocks
- [ ] No more "unsupported cty.Type conversion" crashes
- [ ] Example files updated to demonstrate new syntax
- [ ] If applicable: both ansible-navigator and ansible-navigator-local provisioners support the new structs

## Suggested Implementation Phases

This change can potentially be split into multiple OpenSpec changes for easier review and testing:

### Option A: Single Atomic Change (Recommended)

**Change: `fix-navigator-config-rpc-serialization`**

- Implement all struct definitions
- Update both provisioners
- Update examples
- Test and verify

**Rationale**: The structs are tightly coupled, and partial implementations won't work (plugin will still crash). Better to do it all at once in a focused change.

### Option B: Multi-Phase (If team prefers smaller changes)

**Phase 1: `add-navigator-config-structs`**

- Define all Go structs in provisioner.go
- Add types to go:generate directive
- Run make generate
- Update tests to compile with new types (but don't use them yet)

**Phase 2: `migrate-to-struct-based-navigator-config`**

- Update Config to use new NavigatorConfig type
- Update both provisioners to use structs
- Update examples

**Rationale**: Separates definition from usage, but Phase 1 alone doesn't provide value and Phase 2 can't merge until Phase 1 is done.

### Option C: Incremental with Fallback

**Phase 1: `add-yaml-string-navigator-config-fallback`**

- Quick fix: Change navigator_config to cty.String
- Add YAML parsing in Go
- Get plugin working immediately

**Phase 2: `add-native-hcl-navigator-config-structs`**

- Add proper structs (long-term solution)
- Deprecate string format but keep it working

**Rationale**: Gets users unblocked quickly with Phase 1, then improves UX with Phase 2. Good if there's urgency.

## Recommendation

**Use Option A (Single Atomic Change)** because:

1. The fix is focused and well-scoped
2. Partial implementations leave the plugin broken
3. The structs are interdependent
4. Testing is simpler with one complete change
5. Follows existing OpenSpec pattern (fix-navigator-config-hcl2-type was similarly scoped)

If there's urgency to unblock users: Use **Option C Phase 1** as an emergency fix, then follow up with **Option C Phase 2** for the proper solution.

## Key Files to Modify

Primary files:

- `/home/solomong/dev/packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/provisioner.go` (struct definitions, Config update, go:generate)
- `/home/solomong/dev/packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/provisioner.hcl2spec.go` (will be regenerated)
- `/home/solomong/dev/packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator-local/provisioner.go` (if it has navigator_config too)
- `/home/solomong/dev/packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator-local/provisioner.hcl2spec.go` (will be regenerated)
- Examples: `/home/solomong/dev/packer/plugins/packer-plugin-ansible-navigator/example/*.pkr.hcl`

Testing files:

- Any test files that use navigator_config

## Related Work

- Current issue: fix-navigator-config-hcl2-type (attempted manual override with DynamicPseudoType - didn't work due to RPC limitations)
- Root cause analysis: DIAGNOSIS.md
- Solution exploration: PACKER_FRIENDLY_FORMATS.md
