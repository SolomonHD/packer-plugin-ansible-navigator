# Change: Add Configurable Version Check Timeout

## Why

The ansible-navigator plugin's version check (in `getVersion()`) can hang indefinitely when the binary is not found or doesn't respond. This particularly affects users who install ansible-navigator via asdf, where the shim wrapper may not work correctly in subprocess contexts without proper environment setup.

The current implementation uses `cmd.Output()` without a timeout, causing `packer validate` to hang when ansible-navigator cannot be located or takes too long to respond.

## What Changes

- Add `version_check_timeout` configuration field to both provisioners (remote and local)
- Modify `getVersion()` to use `context.WithTimeout` and `exec.CommandContext`
- Set default timeout to 60 seconds (1 minute)
- Improve error messages to distinguish between timeout, not-found, and other errors
- Add asdf-specific examples to documentation showing recommended configurations
- Regenerate HCL2 spec files to include the new configuration field

## Impact

### Affected Specs

- `remote-provisioner-capabilities` - Add version check timeout requirement
- `local-provisioner-capabilities` - Add version check timeout requirement

### Affected Code

- `provisioner/ansible-navigator/provisioner.go` - Update Config struct and getVersion()
- `provisioner/ansible-navigator-local/provisioner.go` - Update Config struct (no version check currently)
- Both HCL2 spec files after regeneration

### Backward Compatibility

- **Non-Breaking**: Existing configurations without `version_check_timeout` will use the 60-second default
- The existing `skip_version_check` option remains available as an escape hatch

### User Experience

- Users experiencing hangs will now get clear timeout errors within 60 seconds
- Error messages provide actionable solutions (ansible_navigator_path, skip_version_check, custom timeout)
- asdf users get specific documentation on recommended configurations
