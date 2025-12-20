# OpenSpec Change Prompt

## Context
The `packer-plugin-ansible-navigator` provisioner constructs an `ansible-navigator run` command that includes `--extra-vars` with a JSON object. The JSON is passed to `exec.Command` without proper shell quoting, causing the shell to interpret curly braces and split the argument. This results in `ansible-playbook: error: argument -e/--extra-vars: expected one argument`.

Example broken command (from debug logs):
```
ansible-navigator run --mode stdout --extra-vars {"ansible_ssh_private_key_file":"/tmp/key","packer_build_name":"foo"} -i inventory playbook.yml
```

The JSON object is split by the shell, leaving `--extra-vars` with no argument.

## Goal
Fix the provisioner so that `--extra-vars` receives the complete JSON string as a single argument, preventing shell interpretation of the JSON content.

## Scope

**In scope:**
- Change how extra vars are passed to `ansible-navigator run` so the JSON cannot be split by shell expansion
- Ensure all Packer-injected vars (`ansible_ssh_private_key_file`, `packer_build_name`, `packer_builder_type`) and user-defined extra vars reach Ansible correctly
- Add unit tests verifying extra-vars argument construction

**Out of scope:**
- Changes to HCL configuration schema
- Container networking or EE image issues
- Unrelated refactors

## Desired Behavior
- Extra vars must be passed as a single, indivisible argument to `ansible-navigator run`
- Two implementation options (pick one):
  1. **File-based approach (preferred):** Write extra vars to a temp JSON file, pass `--extra-vars @/path/to/extravars.json`
  2. **Quoting approach:** Ensure the JSON string is passed as a single argument to `exec.Command` (Go's `exec.Command` with separate args should handle this, but verify the actual execution path)
- The fix must work regardless of JSON content (special characters, nested objects)

## Constraints & Assumptions
- Assumption: The provisioner uses Go's `exec.Command` or similar to run `ansible-navigator`
- Constraint: Must work with both EE-enabled and EE-disabled modes
- Constraint: Temp files (if used) must be cleaned up after execution

## Acceptance Criteria
- [ ] Running the provisioner no longer produces `ansible-playbook: error: argument -e/--extra-vars: expected one argument`
- [ ] Extra vars containing special characters (quotes, braces, colons) are passed correctly
- [ ] Unit tests verify:
  - [ ] Extra vars with Packer-injected variables are constructed correctly
  - [ ] Extra vars with user-defined variables are constructed correctly
  - [ ] Combined Packer + user vars work together
- [ ] Debug logs show the complete, properly-formed command
