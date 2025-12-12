# Tasks

## 1. Code refactoring and cleanup

- [ ] Review navigator_config.go in both provisioners for refactoring opportunities
- [ ] Extract common YAML generation logic if duplicated between provisioners
- [ ] Clean up any temporary workarounds from Phase 3 migration
- [ ] Ensure consistent error handling patterns
- [ ] Review and optimize data structure usage
- [ ] Add code comments explaining complex logic

## 2. Error message improvements

- [ ] Add validation for common navigator_config mistakes:
  - `execution-environment.enabled = true` without `image` specified
  - Empty nested sections that should have values
  - Invalid YAML structures
- [ ] Enhance error messages with specific field paths (e.g., "navigator_config.execution-environment.image")
- [ ] Include suggested fixes in error messages
- [ ] Add links to documentation for complex errors
- [ ] Test error messages for clarity and helpfulness

## 3. Performance optimization

- [ ] Profile YAML generation performance with complex configs
- [ ] Optimize if bottlenecks identified
- [ ] Consider caching parsed navigator_config if called multiple times
- [ ] Ensure temp file operations are efficient
- [ ] Benchmark before and after optimizations

## 4. Enhanced documentation

- [ ] Add comprehensive examples to README.md:
  - Basic EE configuration
  - Advanced EE with custom environment variables
  - Non-EE configuration
  - Complex multi-section configs
- [ ] Expand docs/EXAMPLES.md with real-world scenarios:
  - Enterprise proxy setup with EE
  - Custom ansible.cfg settings via navigator_config
  - Pull policy and container registry configuration
- [ ] Add TROUBLESHOOTING.md section for navigator_config issues:
  - Common mistakes and fixes
  - How to debug YAML generation
  - Where to find generated config file
  - EE permission issues

## 5. Example scenarios

- [ ] Add example: Basic EE configuration
- [ ] Add example: EE with environment variables
- [ ] Add example: Custom temp directory locations
- [ ] Add example: SSH connection tuning via ansible.config
- [ ] Add example: Multiple ansible.cfg sections
- [ ] Add example: Pull policy configuration
- [ ] Add example: Using navigator_config without EE (local ansible)

## 6. Testing enhancements

- [ ] Add tests for common configuration patterns
- [ ] Add tests for error message quality
- [ ] Add performance benchmarks for YAML generation
- [ ] Add integration test examples
- [ ] Verify all example configs actually work

## 7. Verification

- [ ] Run `go build ./...` and verify compilation succeeds
- [ ] Run `go test ./...` and verify all tests pass
- [ ] Run benchmarks and compare to baseline
- [ ] Manually test example configurations
- [ ] Review documentation for completeness and accuracy
