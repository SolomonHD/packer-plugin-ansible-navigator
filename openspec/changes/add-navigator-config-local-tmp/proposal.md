# Change: Add `navigator_config.ansible_config.defaults.local_tmp`

## Why
Execution environments and locked-down systems commonly require a writable *local* Ansible temp directory (`local_tmp`) in addition to the existing `remote_tmp` setting. Without `local_tmp`, runs can fail due to non-writable default paths.

## What Changes
- Add HCL support for `navigator_config.ansible_config.defaults.local_tmp` for both provisioners:
  - `provisioner "ansible-navigator"` (remote/SSH)
  - `provisioner "ansible-navigator-local"` (on-target)
- Ensure the generated configuration artifacts include `local_tmp` when set:
  - `local_tmp` is written into the generated **ansible.cfg** (`[defaults] local_tmp = ...`).
  - The generated `ansible-navigator.yml` references the ansible.cfg via `ansible.config.path` and remains schema-compliant (no `defaults` under `ansible.config`).
- Regenerate HCL2 specs.
- Update documentation/examples to include `local_tmp` usage.

## Non-Goals
- Reintroducing any legacy, top-level temp-dir settings outside `navigator_config`.
- Emitting `defaults` under `ansible.config` in the generated YAML (this remains forbidden by the schema-compliance requirements).

## Impact
- Affected specs:
  - `openspec/specs/remote-provisioner-capabilities/spec.md`
  - `openspec/specs/local-provisioner-capabilities/spec.md`
- Affected implementation areas (for follow-on work): navigator config typed structs, HCL2 schema generation, ansible.cfg generation, and docs/examples.

