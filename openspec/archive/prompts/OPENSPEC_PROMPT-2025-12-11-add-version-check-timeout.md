# OpenSpec change prompt

## Context

The ansible-navigator plugin's version check (in `getVersion()`) can hang indefinitely when the binary is not found or doesn't respond. This particularly affects users who install ansible-navigator via asdf, where the shim wrapper may not work correctly in subprocess contexts without proper environment setup.

The current implementation uses `cmd.Output()` without a timeout, causing `packer validate` to hang when ansible-navigator cannot be located or takes too long to respond.

## Goal

Add a configurable timeout for the version check operation to prevent indefinite hangs, with special consideration for asdf installations.

## Scope

### In scope

- Add `version_check_timeout` configuration option to the provisioner Config struct
- Modify `getVersion()` to use `context.WithTimeout` and `exec.CommandContext`
- Set default timeout to 60 seconds (1 minute)
- Improve error messages to distinguish between timeout and not-found errors
- Add asdf-specific examples to documentation showing recommended configurations

### Out of scope

- Changes to other timeout mechanisms
- Modifications to the actual ansible-navigator execution (only version check)
- Changes to PATH resolution logic beyond documentation

## Desired behavior

### Configuration

Users can optionally set a timeout for the version check:

```hcl
provisioner "ansible-navigator" {
  version_check_timeout = "30s"  # Optional, defaults to "60s"
  skip_version_check = true      # Still available as escape hatch
  # ...
}
```

### Version check behavior

- When `version_check_timeout` is set, the version check must complete within that duration
- If timeout is exceeded, return a clear error message indicating:
  - The version check timed out
  - The configured timeout value
  - Suggestions for resolution (ansible_navigator_path, skip_version_check, or increase timeout)
- If binary is not found (non-timeout error), return the existing error message
- Default timeout of 60 seconds should be reasonable for most installations

### Error messages

On timeout:

```
Error: ansible-navigator version check timed out after 60s. 
This might indicate ansible-navigator is not properly installed or not in PATH.
Solutions:
  - Use 'ansible_navigator_path' to specify additional directories
  - Use 'command' to specify the full path to ansible-navigator
  - Set 'skip_version_check = true' to bypass this check
  - Increase 'version_check_timeout' if your system needs more time
```

On not found (existing):

```
Error: ansible-navigator not found in PATH. Please install it before running this provisioner.
You can use ansible_navigator_path to specify additional directories: <error details>
```

### Documentation additions

Add to installation/configuration docs:

**For asdf users:**

```hcl
# Recommended configuration for asdf-managed ansible-navigator
provisioner "ansible-navigator" {
  # Option 1: Direct path to asdf shim (recommended)
  command = "~/.asdf/shims/ansible-navigator"
  
  # Option 2: Add asdf shims to PATH
  ansible_navigator_path = ["~/.asdf/shims"]
  
  # Optional: Adjust timeout if needed (default: 60s)
  version_check_timeout = "30s"
  
  # Or skip version check entirely if issues persist
  # skip_version_check = true
}
```

**Troubleshooting section:**
Common issue: "Version check hangs with asdf"

- Cause: asdf shims require environment context
- Solution 1: Use `command = "~/.asdf/shims/ansible-navigator"`
- Solution 2: Set `skip_version_check = true`
- Solution 3: Increase `version_check_timeout = "120s"`

## Constraints & assumptions

- Assumption: 60 seconds is sufficient for most systems to complete the version check
- Assumption: Users with very slow systems can manually increase the timeout
- Constraint: Must maintain backward compatibility - existing configs without `version_check_timeout` should work unchanged
- Constraint: The timeout only applies to version check, not to actual provisioning
- Assumption: asdf is the primary tool requiring special handling (pyenv, rbenv similar patterns)

## Acceptance criteria

- [ ] `version_check_timeout` config field added to both provisioners (ansible-navigator and ansible-navigator-local)
- [ ] Field type is duration string (e.g., "60s", "2m") parsed with `time.ParseDuration`
- [ ] Default value is 60 seconds when not specified
- [ ] `getVersion()` uses `context.WithTimeout` and `exec.CommandContext` instead of plain `exec.Command`
- [ ] Timeout errors return distinct, helpful error message with suggestions
- [ ] Non-timeout errors maintain existing error message format
- [ ] HCL2 spec regenerated with `make generate` to include new field
- [ ] Documentation updated with asdf-specific examples in INSTALLATION.md or similar
- [ ] Documentation includes troubleshooting section for version check issues
- [ ] Verification: Plugin with default timeout completes version check or fails with clear timeout error within 60s
- [ ] Verification: Plugin with custom `version_check_timeout = "10s"` times out in approximately 10s
- [ ] Verification: Plugin with `skip_version_check = true` bypasses timeout entirely
