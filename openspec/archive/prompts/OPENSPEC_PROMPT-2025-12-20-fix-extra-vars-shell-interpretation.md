# OpenSpec Change Prompt

## Context
The `packer-plugin-ansible-navigator` provisioner passes `--extra-vars` with a JSON object to `ansible-navigator run`. When execution environments are enabled, `ansible-navigator` launches a container and invokes `ansible-playbook` inside that container. During this handoff, the JSON argument gets shell-interpreted inside the container, causing brace expansion to split the argument. This results in `ansible-playbook: error: argument -e/--extra-vars: expected one argument`.

Example broken command (from debug logs):
```
ansible-navigator run --mode stdout --extra-vars {"ansible_ssh_private_key_file":"/tmp/key","packer_build_name":"foo"} -i inventory playbook.yml
```

When `ansible-navigator` passes this to `ansible-playbook` inside the EE container, the shell interprets `{...}` and splits it, leaving `--extra-vars` without an argument.

## Goal
Fix the provisioner to pass extra vars using Ansible's file-based method (`--extra-vars @/path/to/file.json`) instead of inline JSON, preventing shell interpretation inside execution environment containers.

## Scope

**In scope:**
- Write provisioner-generated extra vars to a temporary JSON file
- Pass `--extra-vars @/path/to/extravars.json` instead of inline JSON
- Clean up temp files after execution (success or failure)
- Fix both `provisioner/ansible-navigator/provisioner.go` and `provisioner/ansible-navigator-local/provisioner.go`
- Update unit tests to verify file-based approach

**Out of scope:**
- Changes to HCL configuration schema
- User-specified play.extra_vars (those remain as separate `-e key=value` pairs)
- Container networking or EE image issues
- Unrelated refactors

## Desired Behavior
- Provisioner-generated extra vars (packer_build_name, packer_builder_type, ansible_ssh_private_key_file) are written to a temp JSON file
- The file path is prefixed with `@` and passed as `--extra-vars @/path/to/packer-extravars-*.json`
- The temp file must be accessible from inside EE containers (use same directory as other temp files)
- Cleanup occurs in defer blocks to ensure deletion even on errors
- The fix works with both EE-enabled and EE-disabled modes

## Constraints & Assumptions
- Confirmed: The plugin correctly uses Go's `exec.Command`, but `ansible-navigator` introduces shell interpretation when invoking `ansible-playbook` inside EE containers
- Constraint: Must not change how user-defined play.extra_vars are passed (those work correctly as `-e key=value`)
- Constraint: Temp files must be cleaned up reliably (use defer)
- Assumption: The temp extra-vars file can use the same temporary directory strategy as existing packer temp files
- Assumption: The `@` prefix is supported by both `ansible-navigator` and `ansible-playbook`

## Acceptance Criteria
- [ ] Running with EE enabled no longer produces `ansible-playbook: error: argument -e/--extra-vars: expected one argument`
- [ ] Extra vars containing braces, quotes, colons, and nested structures are passed correctly
- [ ] Temp extra-vars JSON file is created with proper permissions
- [ ] Temp file is cleaned up after play execution (both success and failure paths)
- [ ] Unit tests verify:
  - [ ] Temp file creation with correct JSON content
  - [ ] File path is formatted with `@` prefix
  - [ ] Arguments array contains `["--extra-vars", "@/tmp/packer-extravars-*.json"]`
  - [ ] Cleanup occurs in defer blocks
- [ ] Debug logs show `--extra-vars @/tmp/packer-extravars-*.json` instead of inline JSON
- [ ] Both remote (`provisioner/ansible-navigator/`) and local (`provisioner/ansible-navigator-local/`) provisioners are fixed
- [ ] Verification commands pass: `make generate && go build ./... && go test ./... && make plugin-check`
