# Change: Add execution_environment Variable

## Why

Ansible Navigator's primary advantage is its ability to run playbooks inside containerized execution environments, providing consistent, reproducible, and portable Ansible execution. Currently, the plugin does not expose the `execution_environment` configuration option, which means:

1. Users cannot specify which container image should be used for playbook execution
2. The plugin relies on ansible-navigator's default execution environment selection
3. Users cannot leverage custom execution environments with specific Ansible versions, collections, or dependencies pre-installed
4. Organizations cannot enforce standardized execution environments across their infrastructure

## What Changes

- Add `execution_environment` configuration field to the provisioner
- Allow users to specify container image for ansible-navigator to use (e.g., `quay.io/ansible/creator-ee:latest`)
- Pass the execution environment to ansible-navigator via the `--execution-environment` flag
- Add validation to ensure the value is properly formatted
- Update documentation with examples of execution environment usage

## Impact

- **Affected specs**: execution-environment (new capability)
- **Affected code**: 
  - `provisioner/ansible-navigator/provisioner.go` - Add Config field and command argument
  - `provisioner/ansible-navigator/provisioner.hcl2spec.go` - Auto-generated from Config
  - Documentation files in `docs/`
- **Breaking changes**: None - this is a purely additive change
- **Backward compatibility**: Fully maintained - if not specified, ansible-navigator will use its default behavior