# Breaking Up refactor-navigator-config-surface

The `refactor-navigator-config-surface` change is too large to implement safely in one session. This document proposes breaking it into 4 smaller, sequential changes.

## Phase 1: Add navigator_config Field (Non-Breaking)

**Change ID**: `add-navigator-config-field`

**Description**: Add the new `navigator_config` field to both provisioners alongside existing options. This is non-breaking - both old and new config methods work.

**Tasks**:

1. Add `NavigatorConfig map[string]interface{}` to both Config structs
2. Implement YAML generation functions (generateNavigatorYAML, applyEEDefaults)
3. Implement temp file generation for ansible-navigator.yml
4. Add ANSIBLE_NAVIGATOR_CONFIG environment variable support
5. Ensure navigator_config takes precedence over legacy CLI flags when both present
6. Run `make generate`
7. Add tests for YAML generation and EE defaults
8. Add documentation showing navigator_config examples
9. Verify with `go build ./...` and `go test ./...`

**Files**:

- `provisioner/ansible-navigator/provisioner.go`
- `provisioner/ansible-navigator-local/provisioner.go`
- New: `provisioner/ansible-navigator/navigator_config.go`
- New: `provisioner/ansible-navigator-local/navigator_config.go`
- Tests and docs

**Benefit**: Users can start migrating to navigator_config while old options still work.

---

## Phase 2: Deprecate Legacy Options (Warning Phase)

**Change ID**: `deprecate-legacy-navigator-options`

**Description**: Mark legacy options as deprecated but still functional. Add deprecation warnings when they're used.

**Tasks**:

1. Add deprecation notices to field comments in Config structs
2. Add runtime warnings when deprecated fields are used
3. Update documentation to recommend navigator_config
4. Add migration examples to docs
5. Update README with deprecation notice

**Files**:

- `provisioner/ansible-navigator/provisioner.go`  
- `provisioner/ansible-navigator-local/provisioner.go`
- All documentation files

**Benefit**: Gives users warning period to migrate. No breaking changes yet.

---

## Phase 3: Remove Legacy Options (Breaking)

**Change ID**: `remove-legacy-navigator-options`

**Description**: Remove deprecated configuration options. This is the breaking change.

**Tasks**:

1. Remove legacy config fields from both provisioners:
   - `ansible_cfg`, `ansible_env_vars`,  `ansible_ssh_extra_args`, `extra_arguments`
   - `execution_environment`, `navigator_mode`
   - `roles_path`, `collections_path`, `galaxy_command`
2. Remove all code processing these options
3. Update validation logic
4. Remove CLI flag generation (--mode, --ee, --eei)
5. Run `make generate`
6. Update all tests to use navigator_config
7. Update all documentation
8. Add migration guide
9. Full verification suite

**Files**:

- Both provisioner files
- `galaxy.go` (update to not use removed fields)
- All test files
- All documentation files

**Benefit**: Clean codebase, single config mechanism.

---

## Phase 4: Cleanup and Optimization

**Change ID**: `optimize-navigator-config`

**Description**: Cleanup and optimize the navigator_config implementation now that legacy options are gone.

**Tasks**:

1. Refactor any remaining cruft from legacy option removal
2. Optimize YAML generation  
3. Add comprehensive error messages for common config mistakes
4. Add more examples and troubleshooting docs
5. Performance testing
6. Final verification

**Files**:

- Both provisioners
- Documentation

**Benefit**: Polish the new implementation.

---

## Implementation Order

1. **Phase 1** (Non-breaking):Use to add navigator_config support
2. **Wait for user feedback/testing**
3. **Phase 2** (Deprecation): Warn users to migrate
4. **Wait 1-2 releases**
5. **Phase 3** (Breaking): Remove old options
6. **Phase 4** (Optimization): Polish

## Why This Approach?

1. **Smaller Changes**: Each phase is manageable (5-10 tasks vs. 39)
2. **Incremental Testing**: Test at each phase before moving forward
3. **Migration Path**: Users have time to adapt (Phases 1-2)
4. **Rollback Possible**: Can revert individual phases if issues found
5. **Less Risk**: Smaller changes = fewer places for bugs

## Decision

To proceed with breakup, manually create these 4 changes in `openspec/changes/` with their own:

- `proposal.md`
- `tasks.md`
- `design.md` (if needed)

Then apply them sequentially using `openspec apply`.
