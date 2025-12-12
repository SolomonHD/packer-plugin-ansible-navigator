# Change: Fix ansible.config Schema Validation Error

## Summary

Fix the plugin's `ansible-navigator.yml` generation to comply with the ansible-navigator settings schema by removing the unsupported `defaults` section from `ansible.config` and using only valid properties.

## Problem

The plugin currently generates invalid `ansible-navigator.yml` files that fail schema validation with the error:

```
In 'ansible-navigator.ansible.config': Additional properties are not allowed ('defaults' was unexpected).
```

The generated structure incorrectly nests Ansible configuration under `ansible.config.defaults`:

```yaml
ansible-navigator:
  ansible:
    config:
      defaults:              # ❌ NOT ALLOWED by schema
        remote_tmp: ...
        host_key_checking: ...
      ssh_connection:        # ❌ NOT ALLOWED by schema
        ssh_timeout: ...
        pipelining: ...
```

According to the [official ansible-navigator settings schema](https://docs.ansible.com/projects/navigator/settings/#the-ansible-navigator-settings-file), the `ansible.config` section only supports:

- `help` (boolean): Show help
- `path` (string): Path to ansible.cfg file
- `cmdline` (string): Additional ansible-config command-line arguments

## Solution

Modify the YAML generation logic in both provisioners to:

1. **Remove** the `Defaults` and `SSHConnection` fields from the `AnsibleConfig` struct (or stop using them in YAML generation)
2. **Add** the valid `Path`, `Help`, and `Cmdline` fields to support the actual ansible-navigator schema
3. **Generate** a separate `ansible.cfg` file when Ansible-specific configuration (like `remote_tmp`) is needed
4. **Reference** that `ansible.cfg` file via `ansible.config.path`

### Expected Structure

```yaml
ansible-navigator:
  ansible:
    config:
      help: false
      path: /tmp/packer-ansible-<uuid>.cfg    # Points to generated ansible.cfg
      cmdline: ""                              # Optional additional args
  execution-environment:
    # ... unchanged ...
```

### Ansible.cfg Generation

When execution environment is enabled and temp directory defaults are needed, generate an `ansible.cfg` file:

```ini
[defaults]
remote_tmp = /tmp/.ansible/tmp
host_key_checking = False

[ssh_connection]
ssh_timeout = 30
pipelining = True
```

Then reference this file via `ansible.config.path` in the ansible-navigator.yml.

## Impact

### Files Modified

- `provisioner/ansible-navigator/provisioner.go` - Update `AnsibleConfig` struct
- `provisioner/ansible-navigator/navigator_config.go` - Fix YAML generation logic, add ansible.cfg generation
- `provisioner/ansible-navigator-local/provisioner.go` - Update `AnsibleConfig` struct
- `provisioner/ansible-navigator-local/navigator_config.go` - Fix YAML generation logic, add ansible.cfg generation

### Breaking Change Assessment

**Minor breaking change for users who:**

- Manually specified `navigator_config.ansible_config.defaults` or `navigator_config.ansible_config.ssh_connection` in HCL

**Mitigation:**

- These fields were generating invalid YAML that caused ansible-navigator to fail validation
- Users experiencing the schema error will need to update their configuration, but the result will actually work
- Document the change in CHANGELOG.md with migration guidance

## Implementation Notes

1. The automatic EE defaults (lines 22-53 in navigator_config.go) currently set `Defaults.RemoteTmp` - this logic will need to generate ansible.cfg content instead
2. Both `ansible-navigator` and `ansible-navigator-local` provisioners have identical code that needs fixing
3. The generated ansible.cfg file should be cleaned up after execution (like the navigator config file)
4. Need to run `make generate` after modifying the Config structs to regenerate HCL2 specs

## References

- [Ansible Navigator Settings Documentation](https://docs.ansible.com/projects/navigator/settings/#the-ansible-navigator-settings-file)
- OPENSPEC_PROMPT.md (archived after proposal creation)
