# Change: Fix ansible-navigator.yml `ansible.config` schema compliance

## Why

The plugin currently generates an `ansible-navigator.yml` that fails ansible-navigator schema validation with:

> `In 'ansible-navigator.ansible.config': Additional properties are not allowed ('defaults' was unexpected).`

This blocks or breaks provisioning runs when ansible-navigator rejects the generated config.

See the archived change prompt in [`packer/plugins/packer-plugin-ansible-navigator/openspec/archive/prompts/OPENSPEC_PROMPT-2025-12-15-fix-navigator-ansible-config-schema.md:5`](packer/plugins/packer-plugin-ansible-navigator/openspec/archive/prompts/OPENSPEC_PROMPT-2025-12-15-fix-navigator-ansible-config-schema.md:5).

## What Changes

- The generated `ansible-navigator.yml` MUST conform to the ansible-navigator settings schema for `ansible.config`:
  - `ansible.config` may contain only `help`, `path`, and/or `cmdline`.
  - It MUST NOT contain `defaults`, `ssh_connection`, or any other nested Ansible configuration.
- The existing `navigator_config { ansible_config { ... } }` HCL surface remains available.
  - If users set `ansible_config.defaults` and/or `ansible_config.ssh_connection`, the plugin will generate an `ansible.cfg` file (INI) and point ansible-navigator to it using `ansible.config.path`.
  - If users set `ansible_config.config` (path), it is mapped to `ansible.config.path`.
  - If users set BOTH `ansible_config.config` AND nested blocks (`defaults`/`ssh_connection`), configuration validation MUST fail (mutually exclusive).
- Automatic EE defaults (temp directories) continue to be applied, but via:
  - execution-environment `environment-variables.set` in `ansible-navigator.yml`, and
  - the generated `ansible.cfg` (not by emitting `defaults` under `ansible.config`).

## Impact

- **Affected specs**:
  - [`packer/plugins/packer-plugin-ansible-navigator/openspec/specs/remote-provisioner-capabilities/spec.md:239`](packer/plugins/packer-plugin-ansible-navigator/openspec/specs/remote-provisioner-capabilities/spec.md:239)
  - [`packer/plugins/packer-plugin-ansible-navigator/openspec/specs/local-provisioner-capabilities/spec.md:283`](packer/plugins/packer-plugin-ansible-navigator/openspec/specs/local-provisioner-capabilities/spec.md:283)
- **Affected code (implementation later)**:
  - YAML generation in [`packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:15`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:15)
  - YAML generation in [`packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator-local/navigator_config.go:15`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator-local/navigator_config.go:15)
