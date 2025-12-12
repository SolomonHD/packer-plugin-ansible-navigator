# Tasks: Phase 4 - Post-Removal Cleanup

## 1. Remove Dead Code from Remote Provisioner

- [x] 1.1 Remove `generateAnsibleCfg()` function from `provisioner/ansible-navigator/provisioner.go` (lines ~1499-1532)
- [x] 1.2 Remove `createTempAnsibleCfg()` function from `provisioner/ansible-navigator/provisioner.go` (lines ~1535-1553)

## 2. Remove Dead Code from Local Provisioner

- [x] 2.1 Remove `generateAnsibleCfg()` function from `provisioner/ansible-navigator-local/provisioner.go` (lines ~701-735)
- [x] 2.2 Remove `createTempAnsibleCfg()` function from `provisioner/ansible-navigator-local/provisioner.go` (lines ~738-756)

## 3. Update Documentation - JSON_LOGGING.md

- [x] 3.1 Replace `navigator_mode = "json"` with `navigator_config = { mode = "json" }` in first example (line ~21)
- [x] 3.2 Replace `navigator_mode = "json"` with `navigator_config = { mode = "json" }` in local example (line ~41)
- [x] 3.3 Update configuration table row from `navigator_mode` to `navigator_config.mode` (line ~61)
- [x] 3.4 Replace `navigator_mode = "json"` in debug example (line ~124)
- [x] 3.5 Replace `navigator_mode = "json"` in error handling example (line ~148)
- [x] 3.6 Replace `navigator_mode = "json"` in performance example (line ~165)
- [x] 3.7 Update explanatory text references to `navigator_mode` (lines ~7, ~215, ~228)

## 4. Update Documentation - README.md

- [x] 4.1 Replace `navigator_mode = "json"` with `navigator_config = { mode = "json" }` in JSON logging example (line ~107)

## 5. Update Documentation - TROUBLESHOOTING.md

- [x] 5.1 Update guidance about passing flags from `navigator_mode`/`extra_arguments` to `navigator_config` (line ~33)

## 6. Update Documentation - MIGRATION.md

- [x] 6.1 Change `execution_environment` status from "⚠️ Deprecated" to "❌ Removed" (line ~20)
- [x] 6.2 Change `navigator_mode` status from "⚠️ Deprecated" to "❌ Removed" (line ~21)
- [x] 6.3 Change `ansible_cfg` status from "⚠️ Deprecated" to "❌ Removed" (line ~22)
- [x] 6.4 Change `ansible_env_vars` status (line ~23)
- [x] 6.5 Change `ansible_ssh_extra_args` status from "⚠️ Deprecated" to "❌ Removed" (line ~24)
- [x] 6.6 Change `extra_arguments` status from "⚠️ Deprecated" to "❌ Removed" (line ~25)
- [x] 6.7 Change `roles_path` status from "⚠️ Deprecated" to "❌ Removed" (line ~26)
- [x] 6.8 Change `collections_path` status from "⚠️ Deprecated" to "❌ Removed"
- [x] 6.9 Change `galaxy_command` status from "⚠️ Deprecated" to "❌ Removed"

## 7. Update OpenSpec Base Specs

- [x] 7.1 Update `openspec/specs/local-provisioner-capabilities/spec.md`:
  - Remove `extra_arguments` reference in command validation scenario (line ~37)
  - Remove legacy precedence scenario (lines ~378-399)
- [x] 7.2 Update `openspec/specs/remote-provisioner-capabilities/spec.md`:
  - Remove `execution_environment` scenarios (lines ~19-40)
  - Remove legacy precedence scenario (lines ~391-411)

## 8. Verification

- [x] 8.1 Run `go build ./...` to verify no compilation errors
- [x] 8.2 Run `go test ./...` to verify all tests pass
- [x] 8.3 Run `make plugin-check` to verify plugin conformance
- [x] 8.4 Run `grep -r "generateAnsibleCfg\|createTempAnsibleCfg" --include="*.go"` to verify functions removed
- [x] 8.5 Run `grep -r "navigator_mode\s*=" docs/` to verify no legacy syntax in docs

## Dependencies

- Tasks 1.x and 2.x can be parallelized (dead code removal)
- Tasks 3.x-7.x can be parallelized (documentation updates)
- Tasks 8.x must run after all other tasks complete
