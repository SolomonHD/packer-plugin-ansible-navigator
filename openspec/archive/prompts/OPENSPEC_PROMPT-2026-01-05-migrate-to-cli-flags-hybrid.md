# OpenSpec Prompt: Migrate ansible-navigator Configuration to CLI Flags (Hybrid Approach)

## Context

The plugin currently generates temporary YAML configuration files (`ansible-navigator.yml`) and passes them to ansible-navigator via the `--settings` flag. This approach has encountered persistent issues:

1. **Version Format Issues**: ansible-navigator 25.x detects YAML as "Version 1" format and triggers migration prompts
2. **Configuration Ignored**: Settings like `pull_policy = "never"` are being ignored despite correct YAML structure
3. **Temporary File Management**: Complexity in creating, tracking, and cleaning up temp files
4. **YAML Schema Dependency**: Vulnerable to ansible-navigator YAML schema changes

## Goal

Refactor the plugin to use **ansible-navigator CLI flags as the primary configuration method**, with minimal YAML generation only for advanced features that lack CLI flag equivalents.

## Scope

### In Scope

1. **Primary CLI Flag Generation**:
   - Mode (`--mode stdout|interactive`)
   - Execution environment image (`--execution-environment-image`)
   - Pull policy (`--pull-policy never|missing|always|tag`)
   - Container engine (`--execution-environment-container-engine docker|podman`)
   - Container options (`--container-options`)
   - Environment variables (`--eev KEY=VALUE`, repeatable)
   - Volume mounts (`--evm src:dest:options`, repeatable)
   - Logging (`--log-level`, `--log-file`, `--log-append`)
   - ansible.cfg path (`--ansible-config`)
   - EE enable/disable (`--execution-environment true|false`)

2. **Fallback YAML Generation** (only when needed):
   - `playbook-artifact` settings (if user configures them)
   - `collection-doc-cache` settings (if user configures them)
   - Any other advanced features not mappable to CLI flags

3. **Hybrid Command Construction**:
   - Build CLI flags for all supported settings
   - Generate minimal YAML only for unmapped settings
   - Pass YAML via `--settings` only if fallback needed

4. **Update Both Provisioners**:
   - `provisioner/ansible-navigator/` (remote)
   - `provisioner/ansible-navigator-local/` (local)

### Out of Scope

- Changes to HCL configuration structure (user-facing config remains unchanged)
- Modifications to Config structs or HCL2 specs
- Changes to ansible.cfg generation (separate concern)

## Desired Behavior

### Command Construction Logic

```
IF all navigator_config settings have CLI flag equivalents THEN
    ansible-navigator run playbook.yml \
      --mode stdout \
      --execution-environment-image quay.io/ansible/creator-ee:latest \
      --pull-policy never \
      --eev ANSIBLE_REMOTE_TMP=/tmp/.ansible/tmp \
      --evm /host/path:/container/path:ro \
      --log-level info
ELSE
    Generate minimal YAML with only unmapped settings
    ansible-navigator run playbook.yml \
      --settings /tmp/packer-navigator-cfg-minimal-XXX.yml \
      --mode stdout \
      --execution-environment-image ... \
      ...
END IF
```

### CLI Flag Mapping

| NavigatorConfig Field | CLI Flag | Repeatable | Notes |
|----------------------|----------|------------|-------|
| `Mode` | `--mode` | No | stdout, interactive |
| `ExecutionEnvironment.Enabled` | `--execution-environment` | No | true/false |
| `ExecutionEnvironment.Image` | `--execution-environment-image` | No | |
| `ExecutionEnvironment.ContainerEngine` | `--execution-environment-container-engine` | No | docker, podman |
| `ExecutionEnvironment.ContainerOptions` | `--container-options` | No | Space-separated list |
| `ExecutionEnvironment.PullPolicy` | `--pull-policy` | No | never, missing, always, tag |
| `ExecutionEnvironment.EnvironmentVariables.Set` | `--eev KEY=VALUE` | Yes | One flag per variable |
| `ExecutionEnvironment.EnvironmentVariables.Pass` | `--eev KEY` | Yes | Pass through from host |
| `ExecutionEnvironment.VolumeMounts` | `--evm src:dest:options` | Yes | One flag per mount |
| `Logging.Level` | `--log-level` or `--ll` | No | |
| `Logging.File` | `--log-file` or `--lf` | No | |
| `Logging.Append` | `--log-append` or `--la` | No | true/false |
| `AnsibleConfig.Config` | `--ansible-config` | No | Path to ansible.cfg |

**Unmapped (require YAML if configured)**:
- `PlaybookArtifact.*`
- `CollectionDocCache.*`

### Expected Output Commands

**Example 1: CLI-only (common case)**
```bash
ansible-navigator run /tmp/playbook.yml \
  --mode stdout \
  --execution-environment-image quay.io/ansible/creator-ee:latest \
  --pull-policy never \
  --execution-environment-container-engine docker \
  --eev ANSIBLE_REMOTE_TMP=/tmp/.ansible/tmp \
  --eev HOME=/tmp \
  --evm /home/user/.packer.d/collections:/tmp/.packer_ansible/collections:ro \
  --log-level info
```

**Example 2: Hybrid (rare case with playbook-artifact)**
```bash
ansible-navigator run /tmp/playbook.yml \
  --settings /tmp/packer-navigator-cfg-minimal-abc123.yml \
  --mode stdout \
  --execution-environment-image quay.io/ansible/creator-ee:latest \
  --pull-policy never \
  --log-level info
```

Where minimal YAML contains only:
```yaml
ansible-navigator:
  playbook-artifact:
    enable: true
    save-as: /path/to/artifact.json
```

## Constraints & Assumptions

1. **CLI flags take precedence over YAML**: If both are present, CLI flag wins
2. **Backward compatibility**: Existing HCL configurations must work without changes
3. **Automatic EE defaults**: Continue applying safe EE defaults (ANSIBLE_REMOTE_TMP, etc.)
4. **Collections path mounting**: Maintain current automatic collections mounting behavior
5. **Flag ordering**: ansible-navigator is generally tolerant of flag order
6. **Shell escaping**: Properly escape values containing spaces or special characters
7. **Validation**: Validate that unmapped settings trigger YAML generation

## Acceptance Criteria

1. **Primary Path (CLI-only)**:
   - WHEN user configures only CLI-mappable settings
   - THEN no YAML file is generated
   - AND command uses only CLI flags
   - AND `pull_policy = "never"` correctly prevents Docker pulls
   - AND no version migration prompts appear

2. **Fallback Path (Hybrid)**:
   - WHEN user configures unmapped settings (e.g., `playbook_artifact { enable = true }`)
   - THEN minimal YAML file is generated containing ONLY unmapped settings
   - AND CLI flags are used for all mapped settings
   - AND command includes both `--settings` and other flags

3. **Verification**:
   - Test with ansible-navigator 25.12.0+ shows no YAML version issues
   - Pull policy "never" prevents Docker registry access
   - All existing tests continue passing
   - Generated commands are valid and execute successfully

4. **Code Quality**:
   - Clear separation between CLI flag generation and YAML generation
   - Well-documented mapping logic
   - Minimal YAML generation only when necessary
   - Proper cleanup of any fallback YAML files

## Implementation Notes

### Suggested Approach

1. **Create CLI Flag Builder**:
   - New function: `buildNavigatorCLIFlags(config *NavigatorConfig) []string`
   - Returns array of flag strings
   - Handles repeatable flags (--eev, --evm)

2. **Detect Unmapped Settings**:
   - New function: `hasUnmappedSettings(config *NavigatorConfig) bool`
   - Returns true if PlaybookArtifact or CollectionDocCache configured

3. **Generate Minimal YAML**:
   - New function: `generateMinimalYAML(config *NavigatorConfig) string`
   - Only includes unmapped settings
   - Returns empty string if no unmapped settings

4. **Update Command Construction**:
   - Modify `ProvisionRemote()` and `ProvisionLocal()` (or their helpers)
   - Build CLI flags first
   - Conditionally add `--settings` if YAML needed

### Testing Strategy

1. Unit tests for CLI flag generation
2. Unit tests for unmapped setting detection
3. Integration tests with actual ansible-navigator execution
4. Test pull-policy enforcement with local Docker images
5. Verify no YAML files created in CLI-only path

### Files to Modify

- `provisioner/ansible-navigator/provisioner.go` - Update command construction
- `provisioner/ansible-navigator/navigator_config.go` - Add CLI flag builder, minimal YAML generator
- `provisioner/ansible-navigator-local/provisioner.go` - Update command construction
- `provisioner/ansible-navigator-local/navigator_config.go` - Add CLI flag builder, minimal YAML generator
- Tests for both provisioners

## Related Issues

- Fixes Version 2 format issues by avoiding YAML in common case
- Eliminates pull-policy ignored bug
- Simplifies maintenance by reducing YAML dependency
- Improves debugging with explicit CLI commands

## Success Metrics

- Zero YAML files generated for 90%+ of use cases
- No version migration prompts with ansible-navigator 25.x
- Pull policy correctly enforced
- Command line fully visible in logs for debugging
