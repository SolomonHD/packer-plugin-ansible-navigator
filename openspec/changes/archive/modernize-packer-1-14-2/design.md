# Design Document: Packer 1.14.2 Modernization

## Context

The packer-plugin-ansible-navigator was forked from HashiCorp's original ansible plugin and has evolved with custom features. It currently works well but uses patterns from earlier Go and Packer SDK versions. With Packer 1.14.2 released and Go 1.23 available, we have an opportunity to modernize the codebase while maintaining backward compatibility.

**Key Constraints:**
- Must maintain backward compatibility with existing HCL configurations
- Should work with Packer 1.10.0+ but optimized for 1.14.2+
- Cannot introduce breaking changes to user-facing API
- Must preserve all existing functionality

**Stakeholders:**
- Plugin users running Packer builds
- Maintainers developing new features
- CI/CD systems relying on the plugin

## Goals / Non-Goals

### Goals
1. Replace all deprecated Go standard library usage (io/ioutil)
2. Implement modern error handling with proper error chains
3. Add comprehensive test coverage (80%+)
4. Improve type safety and validation
5. Align with latest Packer Plugin SDK patterns
6. Fix go.mod version declaration (currently incorrect 1.25.3)

### Non-Goals
1. Changing HCL configuration syntax (backward compatibility)
2. Rewriting core ansible-navigator invocation logic
3. Adding new features beyond modernization
4. Changing plugin name or registration
5. Modifying execution environment behavior

## Technical Decisions

### Decision 1: Replace io/ioutil Package
**Choice:** Direct replacement with os/io packages  
**Rationale:**
- io/ioutil deprecated since Go 1.16, removed in Go 1.23+
- Direct replacements are straightforward: ReadFile→os.ReadFile, etc.
- No behavior change, just API update

**Alternatives Considered:**
- Keep using io/ioutil and stay on older Go: Rejected, limits future Go features
- Create wrapper functions: Rejected, adds unnecessary abstraction

**Implementation:**
```go
// Before
data, err := ioutil.ReadFile(path)

// After  
data, err := os.ReadFile(path)
```

### Decision 2: Centralized Error Handling
**Choice:** Use %w verb for error wrapping, create structured error types  
**Rationale:**
- Preserves error chains for debugging
- Allows errors.Is() and errors.As() to work correctly
- Standard Go 1.13+ pattern

**Alternatives Considered:**
- String concatenation: Rejected, loses error chain
- Custom error wrapper type: Rejected, unnecessary with %w

**Implementation:**
```go
// Before
return fmt.Errorf("failed to execute: %s", err.Error())

// After
return fmt.Errorf("failed to execute ansible-navigator run %s: %w", playbook, err)
```

### Decision 3: Validation Centralization
**Choice:** All validation in Config.Validate() method called from Prepare()  
**Rationale:**
- Single location makes validation logic discoverable
- Easier to test comprehensively
- Follows SDK patterns for provisioners

**Alternatives Considered:**
- Validation scattered throughout: Current state, hard to maintain
- Validation in separate package: Over-engineering for this scale

### Decision 4: Go Module Version Fix
**Choice:** Change go.mod from "go 1.25.3" to "go 1.23"  
**Rationale:**
- Go 1.25 doesn't exist yet; this is clearly an error
- Go 1.23 is the current/targeted version per .go-version file
- Must be fixed for module to be valid

### Decision 5: Test Strategy
**Choice:** Table-driven tests for validation, integration tests for provisioning  
**Rationale:**
- Table-driven tests scale well for multiple validation scenarios
- Integration tests verify end-to-end functionality
- Matches HashiCorp plugin testing patterns

**Test Structure:**
```go
func TestConfigValidate(t *testing.T) {
    tests := []struct {
        name    string
        config  Config
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid playbook file",
            config: Config{PlaybookFile: "test.yml"},
            wantErr: false,
        },
        // ... more cases
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            // assertions
        })
    }
}
```

## Implementation Phases

### Phase 1: Foundation (Week 1)
- Fix go.mod version
- Replace io/ioutil usage
- Update imports

### Phase 2: Error Handling (Week 1-2)
- Add error wrapping throughout
- Create structured error types
- Improve error messages

### Phase 3: Validation (Week 2)
- Consolidate validation logic
- Add validation helpers
- Improve HCL2 schema annotations

### Phase 4: Testing (Week 2-3)
- Write table-driven validation tests
- Add mock implementations
- Create integration test fixtures
- Achieve 80% coverage target

### Phase 5: Documentation & Polish (Week 3)
- Add godoc to all exports
- Run linters and fix issues
- Update CHANGELOG
- Integration testing with Packer 1.14.2

## Risks / Trade-offs

### Risk: Undetected Behavior Changes
**Impact:** Medium - Could break existing workflows  
**Mitigation:**
- Comprehensive test suite with real ansible-navigator commands
- Beta testing period with early adopters
- Extensive acceptance testing

### Risk: Performance Regression
**Impact:** Low - Additional validation overhead  
**Mitigation:**
- Benchmark critical paths
- Validation overhead is negligible for typical use
- Most time spent in ansible-navigator, not plugin code

### Risk: Go 1.23 Adoption
**Impact:** Low - Users may not have Go 1.23  
**Mitigation:**
- Most users install pre-built binaries, don't need Go
- CI builds with Go 1.23, binaries work on older systems
- Documentation mentions Go 1.23 requirement for development

### Trade-off: Test Coverage vs Development Time
**Decision:** Target 80% coverage, not 100%  
**Rationale:**
- 80% captures most critical paths
- 100% has diminishing returns
- Some error paths are hard to trigger artificially

## Migration Plan

### For Users
**No action required** - Backward compatible changes only

### For Developers
1. Update development environment to Go 1.23
2. Run `go mod tidy` after pulling changes
3. Review updated testing patterns
4. Follow new error handling patterns in PRs

### Rollback Strategy
If critical issues found post-release:
1. Previous version remains available via version pinning
2. Git tag can be used to build older version
3. No breaking changes means simple rollback

## Open Questions

1. **Should we add context.Context to more functions?**
   - Deferred to follow-up - adds complexity
   - Current cancellation via done channel works

2. **Should we extract more managers beyond GalaxyManager?**
   - Deferred - current structure is clear enough
   - Could be future refactoring if needed

3. **Should we use generics for type-safe helpers?**
   - Deferred - current code doesn't have strong use case
   - Would be premature optimization

4. **Target test coverage percentage?**
   - Decision: 80% minimum
   - Critical paths must have 100% coverage

## Success Metrics

- ✅ All deprecated API usage removed
- ✅ go.mod declares valid Go version
- ✅ Test coverage ≥ 80%
- ✅ Zero golangci-lint strict warnings
- ✅ Plugin loads successfully in Packer 1.14.2
- ✅ All existing acceptance tests pass
- ✅ Backward compatibility verified with Packer 1.10.0