# Proposal: Fix Extra Vars Shell Interpretation in Execution Environments

## Why

When `ansible-navigator` runs with execution environments enabled, it launches a container and invokes `ansible-playbook` inside that container. During this handoff, inline JSON passed via `--extra-vars {"key":"value"}` gets shell-interpreted inside the container, causing brace expansion to split the argument. This results in:

```
ansible-playbook: error: argument -e/--extra-vars: expected one argument
```

**Examples from debug logs:**
```bash
ansible-navigator run --mode stdout --extra-vars {"ansible_ssh_private_key_file":"/tmp/key","packer_build_name":"foo"} -i inventory playbook.yml
```

The shell interprets `{...}` and splits it, leaving `--extra-vars` without a proper argument.

## Root Cause

- The plugin uses Go's `exec.Command` correctly with separate arguments
- However, `ansible-navigator` introduces shell interpretation when invoking `ansible-playbook` inside EE containers
- The inline JSON syntax is vulnerable to shell metacharacter expansion

## What Changes

- Write provisioner-generated extra vars to a temporary JSON file
- Pass file-based extra vars via `--extra-vars @/path/to/packer-extravars-*.json` instead of inline JSON
- Ensure the temp file is accessible from inside EE containers (use the same temp directory strategy as existing Packer temp files)
- Clean up temp files reliably using `defer` blocks (success and failure)

## Proposed Solution (Detail)

Use Ansible's file-based extra vars method instead of inline JSON:

1. Write provisioner-generated extra vars (packer_build_name, packer_builder_type, ansible_ssh_private_key_file) to a temporary JSON file
2. Pass `--extra-vars @/path/to/packer-extravars-*.json` instead of inline JSON
3. Ensure the temp file is accessible from inside EE containers (use same directory strategy as existing temp files)
4. Clean up temp files reliably using defer blocks

## Scope

### In Scope
- Write provisioner-generated extra vars to temporary JSON file  
- Pass `--extra-vars @/path/to/extravars.json` instead of inline JSON
- Clean up temp files after execution (success or failure)
- Fix both [`provisioner/ansible-navigator/provisioner.go`](../../provisioner/ansible-navigator/provisioner.go) and [`provisioner/ansible-navigator-local/provisioner.go`](../../provisioner/ansible-navigator-local/provisioner.go)
- Update unit tests to verify file-based approach

### Out of Scope
- Changes to HCL configuration schema  
- User-specified `play.extra_vars` (those remain as separate `-e key=value` pairs)
- Container networking or EE image issues
- Unrelated refactors

## Benefits

1. **Fixes EE compatibility** - No more shell interpretation errors
2. **Future-proof** - Works regardless of EE container shell configuration  
3. **Cleaner command line** - File reference instead of potentially long JSON string
4. **Standard Ansible pattern** - The `@file` syntax is the recommended approach

## Implementation Notes

- Temp file naming: `packer-extravars-<random>.json`
- Use same temp directory strategy as other Packer temp files for EE accessibility
- Cleanup must occur in defer blocks to ensure deletion even on errors
- The fix works with both EE-enabled and EE-disabled modes

## Alternatives Considered

1. **Shell escaping** - Too fragile, depends on which shell is used inside the container
2. **Environment variables** - Not all EE configurations pass env vars through
3. **Keep inline JSON** - Doesn't solve the fundamental shell interpretation issue

## Success Criteria

- ✅ Running with EE enabled no longer produces `ansible-playbook: error: argument -e/--extra-vars: expected one argument`
- ✅ Extra vars containing braces, quotes, colons, and nested structures pass correctly
- ✅ Temp extra-vars JSON file is created with proper permissions
- ✅ Temp file is cleaned up after play execution (both success and failure paths)
- ✅ Debug logs show `--extra-vars @/tmp/packer-extravars-*.json` instead of inline JSON
- ✅ Verification commands pass: `make generate && go build ./... && go test ./... && make plugin-check`
