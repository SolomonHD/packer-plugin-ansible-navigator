# OpenSpec Change Prompt: Expand Ansible Config Sections

## Context

The plugin currently supports minimal ansible.cfg configuration through [`navigator_config.ansible_config`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:1), providing only a `config` path field and basic `defaults` and `ssh_connection` sections. Users cannot configure other important ansible.cfg sections like `privilege_escalation`, `persistent_connection`, `inventory`, etc., which are documented and supported in ansible-navigator v3.x.

## Goal

Expand the [`AnsibleConfig`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:1) struct to support all documented ansible-navigator v3.x ansible configuration sections, matching the full `ansible.config` surface in ansible-navigator.yml.

## Scope

**In scope:**
- Add missing ansible.cfg section structs:
  - `privilege_escalation` (become settings)
  - `persistent_connection` (connection persistence)
  - `inventory` (inventory behavior)
  - `paramiko_connection` (paramiko-specific)
  - `colors` (output coloring)
  - `diff` (diff display settings)
  - `galaxy` (galaxy URL/settings - note: this is different from the plugin's galaxy install options)
  - Any other ansible.cfg sections supported by ansible-navigator v3.x `ansible.config`
- Update `//go:generate packer-sdc mapstructure-to-hcl2` type list with all new section struct types
- Regenerate HCL2 spec files
- Update YAML generation to emit all ansible config sections with correct ansible.cfg section names
- Add tests for HCL decoding and YAML structure

**Out of scope:**
- Expanding execution-environment (previous change #01)
- Expanding logging, playbook-artifact, or other top-level navigator_config sections (later changes)
- Validating ansible.cfg option values (ansible-navigator itself handles that)
- Documenting every possible ansible.cfg option (link to ansible docs in comments/examples)

## Desired Behavior

A user can configure any documented ansible.cfg section through [`navigator_config { ansible_config { ... } }`](packer/plugins/packer-plugin-ansible-navigator/README.md:1):

```hcl
provisioner "ansible-navigator" {
  navigator_config {
    ansible_config {
      config = "/path/to/ansible.cfg"  # optional path override
      
      defaults {
        remote_tmp        = "/tmp/.ansible/tmp"
        host_key_checking = false
        timeout           = 30
      }
      
      privilege_escalation {
        become        = true
        become_method = "sudo"
        become_user   = "root"
      }
      
      ssh_connection {
        ssh_timeout = 30
        pipelining  = true
      }
      
      persistent_connection {
        command_timeout          = 30
        connect_timeout          = 30
        connect_retry_timeout    = 15
      }
      
      inventory {
        enable_plugins = ["ini", "yaml"]
      }
    }
  }
}
```

The generated [`ansible-navigator.yml`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:1) correctly includes:

```yaml
ansible:
  config:
    path: /path/to/ansible.cfg
  config-override:
    defaults:
      remote_tmp: /tmp/.ansible/tmp
      host_key_checking: false
      timeout: 30
    privilege_escalation:
      become: true
      become_method: sudo
      become_user: root
    ssh_connection:
      ssh_timeout: 30
      pipelining: true
    persistent_connection:
      command_timeout: 30
      connect_timeout: 30
      connect_retry_timeout: 15
    inventory:
      enable_plugins:
        - ini
        - yaml
```

## Constraints & Assumptions

- **Assumption:** ansible-navigator v3.x `ansible.config` schema is the authoritative source for which sections/fields to support.
- **Constraint:** HCL field names use underscores, YAML uses hyphens/underscores per ansible.cfg conventions.
- **Constraint:** Some ansible.cfg options are boolean, some string, some list, some integer - struct field types must match.
- **Constraint:** The YAML structure may nest ansible config sections under `ansible.config-override` or similar - verify correct nesting from ansible-navigator v3.x schema.
- **Constraint:** Preserve backward compatibility - existing `defaults` and `ssh_connection` configs must continue to work.
- **Constraint:** Use typed structs for each section (no dynamic maps).
- **Constraint:** It's acceptable to support a subset of ansible.cfg options per section (most commonly used) rather than every possible option - document which are supported.

## Acceptance Criteria

- [ ] [`AnsibleConfig`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:1) struct includes all major ansible.cfg section structs (privilege_escalation, persistent_connection, inventory, colors, diff, galaxy, paramiko_connection).
- [ ] HCL configuration with multiple ansible config sections decodes without error.
- [ ] Generated [`ansible-navigator.yml`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:1) includes correct ansible config section names and nesting.
- [ ] A minimal Packer template using new ansible config sections passes `packer validate`.
- [ ] Existing ansible_config configurations (defaults, ssh_connection) continue to work unchanged.
- [ ] [`make generate`](packer/plugins/packer-plugin-ansible-navigator/GNUmakefile:1) regenerates HCL2 specs successfully.
- [ ] [`go build ./...`](packer/plugins/packer-plugin-ansible-navigator/GNUmakefile:1) compiles without errors.
- [ ] Unit tests validate HCL decoding and YAML structure for representative ansible config sections.

## Files Expected to Change

- [`provisioner/ansible-navigator/navigator_config.go`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:1) (new section structs, YAML generation)
- [`provisioner/ansible-navigator-local/navigator_config.go`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator-local/navigator_config.go:1) (same structs)
- [`provisioner/ansible-navigator/*.hcl2spec.go`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/*.hcl2spec.go:1) (generated)
- [`provisioner/ansible-navigator-local/*.hcl2spec.go`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator-local/*.hcl2spec.go:1) (generated)
- Test files (new tests for ansible config sections)
- Possibly [`example/`](packer/plugins/packer-plugin-ansible-navigator/example/) templates

## Dependencies

- **None required** (can be done independently of #01, though typically done after)
- **Suggested order:** After #01 (expand-execution-environment-config) for logical progression
