# Implementation Tasks

## 1. Shim Detection and Resolution

- [x] 1.1 Create `detectShim()` function that reads file header and detects version manager patterns
- [x] 1.2 Create `resolveShim()` function that attempts to resolve using appropriate version manager command
- [x] 1.3 Add shim detection call in `getVersion()` before executing ansible-navigator command
- [x] 1.4 Use resolved real binary path when shim resolution succeeds
- [x] 1.5 Return descriptive error with solutions when shim detected but resolution fails

## 2. Error Message Enhancements

- [x] 2.1 Update timeout error message in `getVersion()` to include shim troubleshooting section
- [x] 2.2 Ensure all error messages include concrete examples for asdf, rbenv, and pyenv
- [x] 2.3 Verify error messages mention both `command` and `ansible_navigator_path` resolution options

## 3. Testing

- [x] 3.1 Add unit tests for `detectShim()` covering asdf, rbenv, pyenv shim patterns
- [x] 3.2 Add unit tests for `detectShim()` with non-shim files (real binaries, regular scripts)
- [x] 3.3 Add unit tests for `resolveShim()` with mocked version manager commands
- [x] 3.4 Test automatic resolution success path (mock `asdf which` returning valid path)
- [x] 3.5 Test resolution failure path (mock `asdf which` returning error)
- [x] 3.6 Verify shim detection overhead is < 100ms
- [ ] 3.7 Integration test with real asdf installation confirming no hang

## 4. Documentation Updates

- [x] 4.1 Update `docs/TROUBLESHOOTING.md` "Version check hangs / timeouts" section
- [x] 4.2 Add section explaining automatic shim detection and resolution
- [x] 4.3 Add examples showing when shim resolution works automatically
- [x] 4.4 Add examples showing when manual configuration is still needed
- [x] 4.5 Remove or update references positioning manual workarounds as primary solution
- [x] 4.6 Clarify that shims now work automatically in most cases

## 5. Validation

- [x] 5.1 Run `make generate` to regenerate any code (no generation needed - no Config changes)
- [x] 5.2 Run `go build ./...` to verify compilation (passed)
- [x] 5.3 Run `go test ./...` to verify all tests pass (passed - 0.122s)
- [x] 5.4 Run `make plugin-check` to verify plugin conformance (passed)
- [x] 5.5 Run `openspec validate detect-and-handle-version-manager-shims --strict` (valid)
