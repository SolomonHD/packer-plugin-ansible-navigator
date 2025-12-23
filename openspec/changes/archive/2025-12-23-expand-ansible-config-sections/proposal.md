# Change: Expand `navigator_config.ansible_config` sections

## Why

The plugin currently supports only a small subset of Ansible configuration via `navigator_config.ansible_config` (primarily `defaults` and `ssh_connection`). Users need additional ansible.cfg sections (for example `privilege_escalation`, `persistent_connection`, `inventory`, etc.) to tune execution environments and remote connection behavior.

Without these sections, users either cannot express their desired configuration in HCL (parse-time “Unsupported argument” errors) or must rely on external ansible.cfg files and manual wiring.

## What Changes

- Add new nested blocks under `navigator_config.ansible_config { ... }` to represent additional ansible.cfg sections supported by ansible-navigator v3.x.
  - Target sections include at minimum:
    - `privilege_escalation`
    - `persistent_connection`
    - `inventory`
    - `paramiko_connection`
    - `colors`
    - `diff`
    - `galaxy`
  - Additional ansible.cfg sections MAY be added if they are documented and needed for parity.
- Extend the ansible.cfg (INI) generation so these new blocks are rendered as sectioned INI content (e.g., `[privilege_escalation]`) and referenced via `ansible.config.path` inside the generated `ansible-navigator.yml`.
- Keep ansible-navigator schema compliance:
  - `ansible.config` in YAML remains limited to allowed keys (`help`, `path`, `cmdline`).
  - All ansible.cfg tuning continues to be expressed via the generated ansible.cfg, not nested YAML keys under `ansible.config`.

## Impact

- Affected specs:
  - `remote-provisioner-capabilities`
  - `local-provisioner-capabilities`
  - `documentation` (examples + reference links)
- Affected code (implementation follow-up, not in this proposal):
  - [`provisioner/ansible-navigator/navigator_config.go`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:1)
  - [`provisioner/ansible-navigator-local/navigator_config.go`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator-local/navigator_config.go:1)
  - Generated `*.hcl2spec.go` files for both provisioners

## Non-Goals

- No changes to provisioner registration / names.
- No attempt to validate that user-provided ansible.cfg option values are semantically correct (Ansible/ansible-navigator remains responsible for interpreting option values).
- No support for ansible-navigator schema versions outside the agreed target version.
