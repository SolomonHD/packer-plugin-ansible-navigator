# Change: Detect and Handle Version Manager Shims

## Why

ansible-navigator installed via version managers (asdf, rbenv, pyenv) can cause silent hangs due to shim recursion loops. When the plugin executes `ansible-navigator` via `exec.CommandContext`, version manager shims find themselves in PATH and create infinite recursion until the timeout expires. The current 60-second timeout error message doesn't identify shims as the root cause, leaving users to troubleshoot blind.

This affects a significant portion of users who prefer version managers for managing tool installations across projects.

## What Changes

### Detection and Resolution

- Add shim detection function in `getVersion()` before executing ansible-navigator
- Detect common version manager shims by reading file headers for patterns (asdf, rbenv, pyenv)
- Attempt automatic resolution using version manager commands (`asdf which`, `rbenv which`, `pyenv which`)
- Use resolved real binary path for version check when resolution succeeds
- Fail fast with actionable error messages when resolution fails

### Error Messages

- **MODIFIED**: Timeout error messages to include shim troubleshooting guidance
- **ADDED**: Specific error message when shim detected but cannot be resolved
- Both error messages provide clear solutions with example commands and HCL configurations

### Documentation

- **MODIFIED**: `docs/TROUBLESHOOTING.md` to replace outdated shim workaround advice
- **ADDED**: New section explaining automatic shim detection and resolution
- **ADDED**: Examples showing when shims work automatically vs when manual configuration is needed
- **REMOVED**: References to manual shim workarounds as the primary solution

## Impact

### Affected Specs

- `remote-provisioner-capabilities`: Version check behavior changes
- `documentation`: TROUBLESHOOTING.md structure and content changes

### Affected Code

- [`provisioner/ansible-navigator/provisioner.go:623`](../../../provisioner/ansible-navigator/provisioner.go) - `getVersion()` method
- [`docs/TROUBLESHOOTING.md`](../../../docs/TROUBLESHOOTING.md) - Version check troubleshooting section

### Backwards Compatibility

- Fully backwards compatible - only adds detection logic before existing behavior
- Users with manual workarounds (using `command` with full path) unaffected
- Existing timeout and skip_version_check options continue to work
- No breaking changes to HCL configuration surface

## Notes

- Detection overhead is minimal (< 100ms) as it only reads file headers
- Only attempts resolution when shim is detected, avoiding unnecessary overhead
- Fallback to existing behavior if shim detection or resolution fails
- Works on Linux, macOS, and WSL2 where version managers are commonly used
