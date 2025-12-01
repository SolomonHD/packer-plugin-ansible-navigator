# OpenSpec change prompt

## Context

The `ansible-navigator` provisioner (local/default) is a vestige from the original `packer-plugin-ansible` fork. It still uses `ansible-playbook` as its default command and lacks the modern features documented in the README. Meanwhile, the `ansible-navigator-remote` provisioner correctly uses `ansible-navigator run` and has all the documented features.

This creates confusion: the README shows features like `plays`, `collections`, `execution_environment`, `navigator_mode`, and `structured_logging` for `provisioner "ansible-navigator"`, but these only exist in the remote variant.

## Goal

Align the local provisioner with the remote provisioner so both use `ansible-navigator run` and support the same configuration options. The local provisioner should be a first-class citizen, not a legacy artifact.

## Scope

In scope:
- Refactor `provisioner/ansible-navigator/provisioner.go` to use `ansible-navigator run`
- Add missing Config fields to match remote: `plays`, `collections`, `execution_environment`, `navigator_mode`, `structured_logging`, `log_output_path`, `keep_going`, `work_dir`, `requirements_file`, `groups`
- Port execution logic from remote to local where applicable
- Update HCL2 spec generation for new fields
- Update tests for local provisioner

Out of scope:
- Changing the remote provisioner behavior
- Breaking changes to existing `playbook_file` functionality
- Changes to plugin registration or naming

## Desired behavior

- Default command changes from `ansible-playbook` to `ansible-navigator run`
- Local provisioner supports dual invocation: `playbook_file` OR `plays` (mutually exclusive)
- Local provisioner supports `collections` for dependency management
- Local provisioner supports `execution_environment` for containerized runs
- Local provisioner supports `navigator_mode` (stdout/json)
- Local provisioner supports `structured_logging` and `log_output_path`
- Local provisioner supports `keep_going` for multi-play error handling
- All README examples work with `provisioner "ansible-navigator"` without modification

## Constraints & assumptions

- Assume backward compatibility for `playbook_file` usage is required
- Assume the remote provisioner implementation is the reference for correct behavior
- Assume tests should pass after changes

## Acceptance criteria

- [ ] Local provisioner default command is `ansible-navigator run` (not `ansible-playbook`)
- [ ] Local provisioner Config struct includes: `plays`, `collections`, `execution_environment`, `navigator_mode`, `structured_logging`, `log_output_path`, `keep_going`, `work_dir`, `requirements_file`, `groups`
- [ ] HCL2 spec generated correctly for all new fields
- [ ] Validation enforces `playbook_file` XOR `plays` mutual exclusivity
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] `make plugin-check` passes
- [ ] README examples work with the local provisioner