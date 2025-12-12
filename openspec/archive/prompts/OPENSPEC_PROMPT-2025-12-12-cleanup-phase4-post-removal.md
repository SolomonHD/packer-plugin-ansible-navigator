# OpenSpec Change Prompt

## Context

This is Phase 4 of the config surface refactoring per BREAKUP_PLAN.md. Phases 1-3 are complete:

- Phase 1 added `navigator_config` field
- Phase 2 added deprecation warnings
- Phase 3 removed legacy config fields from Config structs

The codebase now has cruft remaining from the legacy removal that needs cleanup.

## Goal

Clean up dead code, fix outdated documentation, and polish the `navigator_config` implementation.

## Scope

**In scope:**

- Remove unused `generateAnsibleCfg()` and `createTempAnsibleCfg()` functions from both provisioners
- Fix doc comments in provisioner.go that reference removed options (`extra_arguments`, `navigator_mode`)
- Update JSON_LOGGING.md to use `navigator_config.mode` instead of `navigator_mode`
- Update docs/README.md examples to use `navigator_config`
- Update docs/TROUBLESHOOTING.md examples to use `navigator_config`
- Update MIGRATION.md status from "Deprecated" to "Removed"
- Add helpful error messages for common `navigator_config` mistakes

**Out of scope:**

- Config struct changes (already complete)
- New features
- Performance optimization (can be separate change if needed)

## Desired Behavior

- No dead code referencing removed legacy options
- All documentation shows only `navigator_config` approach
- Error messages guide users to correct `navigator_config` usage
- `go build ./...`, `go test ./...`, and `make plugin-check` all pass

## Constraints & Assumptions

- Assumption: `generateAnsibleCfg` and `createTempAnsibleCfg` are unused (verify with grep for callers)
- Constraint: Must not change any public API or behavior
- Constraint: MIGRATION.md should be kept as a historical reference for users upgrading

## Acceptance Criteria

- [ ] `generateAnsibleCfg()` and `createTempAnsibleCfg()` removed from both provisioners
- [ ] No doc comments reference `extra_arguments` as a config option
- [ ] JSON_LOGGING.md uses `navigator_config = { mode = "json" }` syntax
- [ ] docs/README.md examples use only `navigator_config`
- [ ] docs/TROUBLESHOOTING.md examples use only `navigator_config`
- [ ] MIGRATION.md table shows "❌ Removed" instead of "⚠️ Deprecated"
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] `make plugin-check` passes
