# Implementation Tasks

## 1. CLI Flag Builder Implementation
- [x] 1.1 Create `buildNavigatorCLIFlags()` function in `provisioner/ansible-navigator/navigator_config.go`
- [x] 1.2 Create `buildNavigatorCLIFlags()` function in `provisioner/ansible-navigator-local/navigator_config.go`
- [x] 1.3 Implement mode flag mapping (`--mode`)
- [x] 1.4 Implement execution environment flags (`--execution-environment`, `--execution-environment-image`, `--execution-environment-container-engine`)
- [x] 1.5 Implement pull policy flag (`--pull-policy`)
- [x] 1.6 Implement repeatable environment variable flags (`--eev KEY=VALUE`)
- [x] 1.7 Implement repeatable volume mount flags (`--evm src:dest:options`)
- [x] 1.8 Implement container options flag (`--container-options`)
- [x] 1.9 Implement logging flags (`--log-level`, `--log-file`, `--log-append`)
- [x] 1.10 Implement ansible.cfg path flag (`--ansible-config`)

## 2. Unmapped Settings Detection
- [x] 2.1 Create `hasUnmappedSettings()` function in `provisioner/ansible-navigator/navigator_config.go`
- [x] 2.2 Create `hasUnmappedSettings()` function in `provisioner/ansible-navigator-local/navigator_config.go`
- [x] 2.3 Check for PlaybookArtifact configuration
- [x] 2.4 Check for CollectionDocCache configuration

## 3. Minimal YAML Generation
- [x] 3.1 Create `generateMinimalYAML()` function in `provisioner/ansible-navigator/navigator_config.go`
- [x] 3.2 Create `generateMinimalYAML()` function in `provisioner/ansible-navigator-local/navigator_config.go`
- [x] 3.3 Include only PlaybookArtifact settings if configured
- [x] 3.4 Include only CollectionDocCache settings if configured
- [x] 3.5 Return empty string if no unmapped settings exist

## 4. Command Construction Refactoring (Remote Provisioner)
- [x] 4.1 Update command construction in `provisioner/ansible-navigator/provisioner.go:executeAnsible()`
- [x] 4.2 Build CLI flags first using `buildNavigatorCLIFlags()`
- [x] 4.3 Add CLI flags to ansible-navigator command arguments via `buildRunCommandArgsForPlay()`
- [x] 4.4 Conditionally generate minimal YAML only if `hasUnmappedSettings()` returns true
- [x] 4.5 ANSIBLE_NAVIGATOR_CONFIG environment variable used only when minimal YAML is generated
- [x] 4.6 Ensure proper cleanup of minimal YAML files in deferred cleanup

## 5. Command Construction Refactoring (Local Provisioner)
- [x] 5.1 Update command construction in `provisioner/ansible-navigator-local/provisioner.go:Provision()`
- [x] 5.2 Build CLI flags first using `buildNavigatorCLIFlags()`
- [x] 5.3 Add CLI flags to ansible-navigator command arguments
- [x] 5.4 Conditionally generate minimal YAML only if `hasUnmappedSettings()` returns true
- [x] 5.5 Upload minimal YAML to target only when needed
- [x] 5.6 Add remote settings path only when minimal YAML is uploaded

## 6. Testing
- [x] 6.1 Add unit tests for `buildNavigatorCLIFlags()` in both provisioners
- [x] 6.2 Add unit tests for `hasUnmappedSettings()` in both provisioners
- [x] 6.3 Add unit tests for `generateMinimalYAML()` in both provisioners
- [x] 6.4 Test CLI-only path (no YAML generated)
- [x] 6.5 Test hybrid path (minimal YAML for playbook-artifact)
- [x] 6.6 Verify pull-policy "never" prevents Docker registry access
- [x] 6.7 Test with ansible-navigator 25.x to ensure no version migration prompts
- [x] 6.8 Verify all existing integration tests pass

## 7. Documentation
- [x] 7.1 Update comments in `buildNavigatorCLIFlags()` functions
- [x] 7.2 Document CLI flag mapping table in code comments
- [x] 7.3 Add logging statements showing generated CLI flags for debugging
- [x] 7.4 Document unmapped settings requiring YAML fallback
