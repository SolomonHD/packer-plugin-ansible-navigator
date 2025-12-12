# OpenSpec Change Prompt

## Context

The packer-plugin-ansible-navigator currently hangs silently when `ansible-navigator` is an asdf shim (or other version manager shim like rbenv, pyenv). The issue occurs because:

1. The plugin executes the command found in PATH via `exec.CommandContext`
2. asdf shims are scripts that call asdf to locate the real binary
3. The shim re-executes with the same PATH, finding itself again
4. This creates an infinite recursion loop that hangs until the 60s timeout
5. The timeout error message doesn't identify shims as the likely cause

Users currently must manually work around this by:

- Finding the real binary path with `asdf which ansible-navigator`
- Configuring `command` to point directly to the real binary
- Using `ansible_navigator_path` to prepend the real install directory
- Skipping version check entirely (not recommended)

## Goal

Detect when `ansible-navigator` is a version manager shim (asdf, rbenv, pyenv, etc.) and either:

1. Automatically resolve the shim to the actual binary and use that
2. Fail fast with a clear, actionable error message explaining the shim issue and solutions

This should prevent silent hangs and provide users with clear guidance on how to fix the configuration.

## Scope

**In scope:**

- Detect common version manager shims (asdf, rbenv, pyenv) in `getVersion()` before execution
- Attempt automatic resolution using version manager commands (e.g., `asdf which ansible-navigator`)
- Provide enhanced error messages that specifically mention shim detection
- Update timeout error messages to include shim troubleshooting guidance
- Test with asdf, rbenv, and pyenv shims

**Out of scope:**

- Changes to the actual ansible-navigator execution flow (only affects version check)
- Support for obscure or custom shim implementations beyond common version managers
- Changes to other provisioner lifecycle methods beyond `getVersion()`
- Modifications to PATH handling (already supported via `ansible_navigator_path`)

## Desired Behavior

### Shim Detection and Resolution

**When the command resolves to a shim:**

1. Read the file header to detect shim patterns (#!/usr/bin/env bash with asdf/rbenv/pyenv keywords)
2. Identify which version manager is in use
3. Attempt to resolve using the appropriate command:
   - asdf: `asdf which ansible-navigator`
   - rbenv: `rbenv which ansible-navigator`
   - pyenv: `pyenv which ansible-navigator`
4. If resolution succeeds, use the real binary path for the version check
5. If resolution fails, provide a clear error message with troubleshooting steps

**When shim is detected but cannot be resolved:**

```
Error: ansible-navigator appears to be an asdf shim, but the real binary could not be resolved.

SOLUTIONS:
1. Verify ansible-navigator is installed:
   $ asdf list ansible

2. Find the actual binary path:
   $ asdf which ansible-navigator

3. Configure the plugin to use the actual binary:
   provisioner "ansible-navigator" {
     command = "/path/to/actual/ansible-navigator"
     # OR
     ansible_navigator_path = ["/path/to/bin/directory"]
   }

4. Alternatively, skip the version check (not recommended):
   provisioner "ansible-navigator" {
     skip_version_check = true
   }
```

### Enhanced Timeout Error

**When version check times out:**

```
ansible-navigator version check timed out after 60s.

COMMON CAUSES:
1. Version manager shim (asdf/rbenv/pyenv) causing infinite recursion
2. ansible-navigator not properly installed or not in PATH
3. ansible-navigator requires additional configuration

TROUBLESHOOTING:
1. Check if you're using a version manager:
   $ which ansible-navigator
   $ head -1 $(which ansible-navigator)  # Should show shebang

2. If using asdf, find the real binary:
   $ asdf which ansible-navigator

3. Configure the plugin with the actual binary:
   provisioner "ansible-navigator" {
     command = "/home/user/.asdf/installs/ansible/2.9.0/bin/ansible-navigator"
     # OR
     ansible_navigator_path = ["/home/user/.asdf/installs/ansible/2.9.0/bin"]
   }

4. Verify ansible-navigator works independently:
   $ ansible-navigator --version

5. Skip version check if needed (not recommended):
   provisioner "ansible-navigator" {
     skip_version_check = true
   }
```

## Constraints & Assumptions

- Assumption: Most users using version managers will have asdf, rbenv, or pyenv
- Assumption: Shims can be detected by reading the file header and looking for specific keywords
- Assumption: Version manager commands (`asdf which`, etc.) are available if the shim is detected
- Constraint: Detection should be fast (< 100ms) to not impact startup time
- Constraint: Must remain backwards compatible with existing configurations
- Constraint: Should work on Linux, macOS, and WSL2
- Assumption: Users who manually specify `command` with a full path don't need shim detection

## Acceptance Criteria

- [ ] Shim detection function correctly identifies asdf, rbenv, and pyenv shims
- [ ] Shim detection returns false for real binaries and non-shim scripts
- [ ] Automatic resolution succeeds when version manager commands are available
- [ ] Version check uses resolved real binary path when shim is detected
- [ ] Clear error message provided when shim detected but resolution fails
- [ ] Timeout error message enhanced to mention shim troubleshooting
- [ ] Shim detection does not significantly impact startup time (< 100ms overhead)
- [ ] Existing configurations without shims continue to work unchanged
- [ ] Unit tests cover shim detection for all supported version managers
- [ ] Integration test confirms asdf shim no longer causes hang
- [ ] Documentation updated to explain shim handling behavior
