## Context

ansible-navigator validates its settings file (`ansible-navigator.yml`) against a JSON schema. The plugin currently generates invalid YAML by placing ansible.cfg-like settings (e.g. `defaults`) under `ansible-navigator.ansible.config`, which fails validation.

The YAML generation logic currently emits nested structures like `ansible: { config: { defaults: ... } }` in [`packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:103`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:103) and [`packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator-local/navigator_config.go:102`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator-local/navigator_config.go:102).

## Goals / Non-Goals

- Goals:
  - Generated `ansible-navigator.yml` is schema-valid, specifically for `ansible.config`.
  - Preserve the existing HCL surface (`navigator_config { ansible_config { ... } }`).
  - Keep execution-environment temp-dir defaults working in EE runs.

- Non-goals:
  - Reworking play execution, inventory generation, or galaxy behavior.
  - Introducing a new user-facing configuration surface (beyond what already exists).

## Decisions

### Decision 1: Use `ansible.config.path` for ansible.cfg

**Choice**: represent Ansible configuration via an `ansible.cfg` file and point ansible-navigator to it with `ansible.config.path`.

**Rationale**:

- The schema allows `ansible.config.path` but does not allow nested `defaults`/`ssh_connection` under `ansible.config`.
- The plugin already manages temporary files for `ansible-navigator.yml`; generating an additional temporary file fits existing patterns.

### Decision 2: Keep existing HCL fields but change their mapping

**Choice**: keep `ansible_config.defaults` and `ansible_config.ssh_connection` as inputs that are rendered into `ansible.cfg` (INI), not into `ansible-navigator.yml`.

**Rationale**:

- Avoids breaking the HCL interface while still achieving schema compliance.
- Keeps EE-safe defaults (`remote_tmp`) expressible without relying on invalid YAML.

### Decision 3: Mutual exclusivity for `ansible_config.config` vs nested config blocks

**Choice**: if `ansible_config.config` is set AND either `defaults` or `ssh_connection` is set, the plugin MUST fail validation.

**Rationale**:

- Avoids ambiguous behavior (merge vs override).
- Keeps configuration outcomes predictable.

## Migration / Compatibility Notes

- Existing configurations that relied on `ansible_config.defaults` will continue to work, but via generated `ansible.cfg` rather than YAML nesting.
- Any configuration that currently sets both `ansible_config.config` and nested blocks will become invalid and must be adjusted.

## Open Questions

- None (mutual-exclusion behavior chosen for mixed `config` + nested blocks).

