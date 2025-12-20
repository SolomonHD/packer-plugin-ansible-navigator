# Implementation Tasks

## Task Breakdown

### Phase 1: Code Changes

- [x] **Task 1.1**: Update ansible-navigator provisioner argument construction
  - File: `provisioner/ansible-navigator/provisioner.go`
  - Convert all `args = append(args, "--flag", value)` to `args = append(args, fmt.Sprintf("--flag=%s", value))`
  - Target patterns: `--extra-vars`, `--ssh-extra-args`, `--limit`, `--mode`
  - Preserve boolean flags: `--become`, `--force`
  - Preserve positional arguments
  - Completed: All flag+value patterns converted to `--flag=value` syntax

- [x] **Task 1.2**: Update ansible-navigator-local provisioner argument construction
  - File: `provisioner/ansible-navigator-local/provisioner.go`
  - Convert all flag-value patterns to `--flag=value` syntax
  - Target patterns: `--tags`, `--become`, `--limit`
  - Preserve boolean flags and positional arguments
  - Completed: All flag+value patterns converted including `-c=local` and `-i=inventory`

- [x] **Task 1.3**: Update ansible-navigator galaxy manager argument construction
  - File: `provisioner/ansible-navigator/galaxy.go`
  - Convert short flag patterns: `-p path` → `-p=path`
  - Convert long flag patterns if present
  - Preserve boolean-only flags: `--offline`, `--force`, `--force-with-deps`
  - Completed: Converted `-r` and `-p` flags to `=` syntax

- [x] **Task 1.4**: Update ansible-navigator-local galaxy manager argument construction
  - File: `provisioner/ansible-navigator-local/galaxy.go`
  - Apply same patterns as Task 1.3
  - Ensure consistency across both provisioner variants
  - Completed: Converted `-r` and `-p` flags to `=` syntax

### Phase 2: Validation

- [x] **Task 2.1**: Run code generation
  - Execute: `make generate`
  - Verify: HCL2 spec files regenerate without errors
  - No changes expected in generated files (internal implementation only)
  - Completed: Code generation successful

- [x] **Task 2.2**: Build verification
  - Execute: `go build ./...`
  - Verify: All packages compile successfully
  - Resolve: Any compilation errors from argument formatting
  - Completed: Build successful with no errors

- [x] **Task 2.3**: Test suite execution
  - Execute: `go test ./...`
  - Verify: All existing tests pass with new format
  - Update test assertions if they explicitly check argument array structure
  - Completed: All tests pass after updating test assertions to expect `--flag=value` format

- [x] **Task 2.4**: Plugin compatibility check
  - Execute: `make plugin-check`
  - Verify: Plugin passes SDK conformance validation
  - Verify: Describe command output is unchanged
  - Completed: Plugin check passed successfully

### Phase 3: Integration Testing

- [ ] **Task 3.1**: Manual integration test
  - Build plugin locally
  - Use test HCL configuration with:
    - Multiple plays with tags
    - Extra vars file
    - SSH extra args
    - Limit filter
  - Run `packer build` and verify success
  - Check generated commands in debug output
  - Verify no "argument expected" or "unknown flag" errors

- [ ] **Task 3.2**: Edge case testing
  - Test with value containing spaces: `--limit=web01 OR web02`
  - Test with value starting with hyphen: `--ssh-extra-args=-o IdentitiesOnly=yes`
  - Test with special characters in extra vars file path
  - Verify all edge cases work correctly

### Phase 4: Documentation

- [ ] **Task 4.1**: Update any internal documentation
  - Check if any README or design docs mention argument construction
  - Update examples if present
  - Note: User-facing HCL syntax unchanged (transparent change)

## Dependencies

- **Task 1.2** depends on **Task 1.1** (follow same pattern)
- **Task 1.4** depends on **Task 1.3** (follow same pattern)
- **Phase 2** depends on **Phase 1** completion
- **Phase 3** depends on **Phase 2** passing
- **Phase 4** can proceed in parallel with **Phase 3**

## Rollback Plan

If issues are discovered:
1. Revert commits for the change
2. Original `--flag value` syntax is still valid
3. No configuration changes needed by users
4. Tests will catch any ansible-navigator incompatibilities

## Success Criteria

All tasks marked complete AND:
- Plugin builds successfully
- All tests pass
- Plugin check passes
- Manual integration test succeeds
- No regression in functionality
