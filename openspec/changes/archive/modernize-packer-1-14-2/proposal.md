# Change: Modernize Plugin for Packer 1.14.2 and Later

## Why

The packer-plugin-ansible-navigator needs modernization to align with HashiCorp's latest best practices, Go 1.25.3 features, and Packer 1.14.2 requirements. Current implementation uses patterns from earlier Packer SDK versions and doesn't fully leverage modern Go idioms or latest plugin SDK capabilities.

Key drivers:
- Packer 1.14.2 introduces new plugin protocol features and expectations
- Go 1.25.3 provides improved error handling, context management, and generics
- HashiCorp Plugin SDK v0.6.x includes new testing utilities and validation patterns
- Current codebase uses deprecated patterns (e.g., `ioutil` package)
- Opportunity to improve type safety and reduce runtime errors

## What Changes

### Core Modernization
- Update Go module to explicitly support Go 1.25.3+ features
- Replace deprecated `io/ioutil` with `os` and `io` packages
- Implement structured error wrapping with `%w` verb consistently
- Add context-aware cancellation throughout provisioning pipeline
- Migrate to modern HCL2 validation patterns

### Plugin SDK Alignment
- Update to latest Packer Plugin SDK best practices
- Implement improved plugin metadata and versioning
- Add structured plugin diagnostics
- Enhance HCL2 schema generation with better type annotations
- Implement SDK-recommended testing patterns

### Code Quality Improvements
- Add comprehensive unit test coverage (target: 80%+)
- Implement table-driven tests for configuration validation
- Add integration test fixtures for Packer 1.14.2
- Improve error messages with actionable context
- Standardize logging using structured logging patterns

### Type Safety Enhancements
- Use Go generics where appropriate for type-safe operations
- Strengthen type constraints in configuration structs
- Add validation helper functions with clear contracts
- Implement compile-time interface checks

### Documentation Updates
- Update all godoc comments to meet Go standard
- Add examples for Packer 1.14.2+ features
- Document breaking changes and migration path
- Include performance characteristics in docs

## Impact

### Affected Specs
- `plugin-architecture` - Core plugin structure and initialization
- `configuration-validation` - HCL2 schema and validation logic
- `error-handling` - Error reporting and recovery mechanisms
- `testing-framework` - Test infrastructure and patterns
- `provisioner-lifecycle` - Prepare/Provision/Cancel flow

### Affected Code
- [`main.go`](main.go:1) - Plugin registration and versioning
- [`provisioner/ansible-navigator/provisioner.go`](provisioner/ansible-navigator/provisioner.go:1) - Core provisioner logic
- [`provisioner/ansible-navigator/galaxy.go`](provisioner/ansible-navigator/galaxy.go:1) - Dependency management
- [`provisioner/ansible-navigator-local/provisioner.go`](provisioner/ansible-navigator-local/provisioner.go:1) - Local mode provisioner
- All test files - Modernize test patterns
- [`go.mod`](go.mod:3) - Verify Go version declaration (currently shows 1.25.3)

### Breaking Changes
**None** - This is a backward-compatible internal modernization. External HCL configuration and plugin API remain unchanged.

### Migration Notes
Users running Packer 1.14.2+ will automatically benefit from improvements. No configuration changes required. Users on older Packer versions should upgrade to 1.14.2+ for best results.

## Benefits

1. **Better Error Messages** - Structured error context helps users debug issues faster
2. **Improved Reliability** - Context-aware cancellation prevents resource leaks
3. **Enhanced Testing** - Higher test coverage catches regressions earlier
4. **Future-Proof** - Alignment with latest HashiCorp patterns eases future updates
5. **Performance** - Modern Go patterns reduce allocations and improve efficiency
6. **Developer Experience** - Clear code patterns make contributions easier

## Risks and Mitigation

**Risk**: Subtle behavior changes during refactoring  
**Mitigation**: Comprehensive test suite with before/after validation

**Risk**: Performance regression from additional validation  
**Mitigation**: Benchmark critical paths before and after changes

**Risk**: Edge cases not covered by tests  
**Mitigation**: Retain defensive programming patterns, add extensive logging

## Success Criteria

- [ ] All tests pass on Packer 1.14.2
- [ ] No usage of deprecated `ioutil` package
- [ ] Test coverage ≥ 80%
- [ ] All public functions have godoc comments
- [ ] Zero golangci-lint warnings with strict config
- [ ] Plugin loads and executes successfully in Packer 1.14.2
- [ ] Backward compatibility verified with Packer 1.10.0+