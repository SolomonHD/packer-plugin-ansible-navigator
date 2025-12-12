# Proposal: Phase 4 - Post-Removal Cleanup

## Summary

This is Phase 4 of the config surface refactoring per BREAKUP_PLAN.md. Phases 1-3 are complete:

- **Phase 1**: Added `navigator_config` field
- **Phase 2**: Added deprecation warnings
- **Phase 3**: Removed legacy config fields from Config structs

This phase cleans up dead code and documentation cruft remaining from the legacy removal.

## Motivation

The codebase now has:

1. **Dead code**: `generateAnsibleCfg()` and `createTempAnsibleCfg()` functions remain in both provisioners but are never called (orphaned after legacy field removal)
2. **Outdated documentation**: Multiple doc files still show `navigator_mode = "json"` syntax instead of the current `navigator_config = { mode = "json" }` approach
3. **Incorrect status markers**: MIGRATION.md still shows "⚠️ Deprecated" for options that are now completely removed

## Scope

### In Scope

- Remove unused `generateAnsibleCfg()` and `createTempAnsibleCfg()` functions from both provisioners
- Fix doc comments in provisioner.go that reference removed options
- Update JSON_LOGGING.md examples to use `navigator_config.mode` instead of `navigator_mode`
- Update docs/README.md examples to use `navigator_config`
- Update docs/TROUBLESHOOTING.md examples to use `navigator_config`
- Update MIGRATION.md status indicators from "⚠️ Deprecated" to "❌ Removed"
- Update openspec/specs base specs to remove legacy references

### Out of Scope

- Config struct changes (already complete in Phase 3)
- New features
- Performance optimization

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

## Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Dead code removal breaks something | Low | High | Verify with grep that functions are truly uncalled |
| Doc updates introduce errors | Low | Medium | Follow existing navigator_config examples in CONFIGURATION.md |

## Related Changes

- Depends on: Phase 3 completion (remove-legacy-navigator-options, archived 2025-12-11)
- Follows: BREAKUP_PLAN.md Phase 4 specification
