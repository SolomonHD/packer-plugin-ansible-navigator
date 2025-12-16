## 1. Implementation

- [ ] 1.1 Add/adjust config validation: `ansible_config.config` MUST be mutually exclusive with `ansible_config.defaults` and `ansible_config.ssh_connection`.
- [ ] 1.2 Update YAML generation for remote provisioner to emit `ansible.config` with only `help`, `path`, and/or `cmdline`.
- [ ] 1.3 Update YAML generation for local provisioner to emit `ansible.config` with only `help`, `path`, and/or `cmdline`.
- [ ] 1.4 Implement `ansible.cfg` (INI) generation from `ansible_config.defaults` and `ansible_config.ssh_connection`.
- [ ] 1.5 Ensure the generated `ansible.cfg` is referenced via `ansible.config.path` in the generated `ansible-navigator.yml`.
- [ ] 1.6 Ensure temp file cleanup works for both generated files (ansible-navigator.yml and ansible.cfg) in both provisioners.

## 2. Tests / Validation

- [ ] 2.1 Add unit tests asserting the generated YAML does NOT contain `ansible.config.defaults` or `ansible.config.ssh_connection`.
- [ ] 2.2 Add unit tests asserting `ansible.config` contains only `help`, `path`, and/or `cmdline`.
- [ ] 2.3 Add unit tests for mutual-exclusion validation errors.

## 3. Documentation

- [ ] 3.1 Update configuration docs to clarify that `ansible_config.defaults` and `ansible_config.ssh_connection` are rendered to `ansible.cfg` and referenced via `ansible.config.path`.
- [ ] 3.2 Add a troubleshooting note for the original schema error and its resolution.

