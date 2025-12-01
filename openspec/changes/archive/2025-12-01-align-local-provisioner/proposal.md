# Proposal: Align Local Provisioner with Remote Provisioner

## Why

The local `ansible-navigator` provisioner is a vestige from the original `packer-plugin-ansible` fork. It still uses `ansible-playbook` as its default command and lacks the modern features documented in the README. This creates confusion because:

1. The README shows features like `plays`, `collections`, `execution_environment`, `navigator_mode`, and `structured_logging` for `provisioner "ansible-navigator"`, but these only exist in the remote variant
2. Users expect `ansible-navigator` to use `ansible-navigator run`, not `ansible-playbook`
3. The naming suggests modern Ansible Navigator support, but the implementation is legacy

## What Changes

This change refactors the local provisioner to use `ansible-navigator run` as its default command and adds all modern configuration options currently available only in the remote provisioner.

**Configuration Changes:**
- Default command changes from `ansible-playbook` to `ansible-navigator run`
- Add Config fields: `plays`, `collections`, `execution_environment`, `navigator_mode`, `structured_logging`, `log_output_path`, `keep_going`, `work_dir`, `requirements_file`, `groups`
- Add Play struct for collection-based execution
- Port GalaxyManager integration for collections management

**Behavior Changes:**
- Support both `playbook_file` and `plays` (mutually exclusive)
- Enforce configuration validation via Config.Validate()
- Support execution environments for containerized Ansible
- Enable structured JSON logging when configured

**Backward Compatibility:**
- Existing `playbook_file` configurations continue to work
- Legacy `ansible-playbook` command can still be specified explicitly
- All existing configuration options remain supported

## Proposed Solution

Align the local provisioner with the remote provisioner by:

1. Changing the default command from `ansible-playbook` to `ansible-navigator run`
2. Adding missing Config fields to match the remote provisioner
3. Porting execution logic from remote to local where applicable
4. Updating validation to enforce `playbook_file` XOR `plays` mutual exclusivity

## Scope

### In Scope

- Refactor `provisioner/ansible-navigator/provisioner.go` to use `ansible-navigator run`
- Add missing Config fields: `plays`, `collections`, `execution_environment`, `navigator_mode`, `structured_logging`, `log_output_path`, `keep_going`, `work_dir`, `requirements_file`, `groups`
- Port execution logic from remote to local (Play struct, validation, GalaxyManager integration)
- Update HCL2 spec generation for new fields
- Update tests for local provisioner
- Ensure backward compatibility for existing `playbook_file` functionality

### Out of Scope

- Changing the remote provisioner behavior
- Breaking changes to existing `playbook_file` functionality
- Changes to plugin registration or naming
- Documentation updates (handled separately after implementation)

## Acceptance Criteria

- [ ] Local provisioner default command is `ansible-navigator run` (not `ansible-playbook`)
- [ ] Local provisioner Config struct includes: `plays`, `collections`, `execution_environment`, `navigator_mode`, `structured_logging`, `log_output_path`, `keep_going`, `work_dir`, `requirements_file`, `groups`
- [ ] HCL2 spec generated correctly for all new fields
- [ ] Validation enforces `playbook_file` XOR `plays` mutual exclusivity
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] `make plugin-check` passes
- [ ] README examples work with the local provisioner

## Related Specifications

- `plugin-registration` - Defines provisioner registration names and package conventions
- `build-tooling` - Go version and build compatibility requirements

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Breaking existing configurations using `ansible-playbook` features | Medium | High | Keep legacy fields, add deprecation warnings |
| Test failures due to execution context differences | Medium | Medium | Port test patterns from remote provisioner |
| HCL2 spec generation issues | Low | Medium | Run `make generate` and validate output |