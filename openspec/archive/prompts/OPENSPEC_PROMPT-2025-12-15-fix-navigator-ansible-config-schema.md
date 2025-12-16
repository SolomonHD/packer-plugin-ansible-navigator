# OpenSpec Change Prompt

## Context

The `packer-plugin-ansible-navigator` provisioner generates an `ansible-navigator.yml` settings file at runtime to configure ansible-navigator behavior. The generated configuration currently has schema validation errors according to the official ansible-navigator settings schema.

The error message indicates:

```
In 'ansible-navigator.ansible.config': Additional properties are not allowed ('defaults' was unexpected).
```

The plugin is incorrectly placing an Ansible configuration section called `defaults` inside the `ansible-navigator.ansible.config` structure, which violates the ansible-navigator settings schema.

## Goal

Fix the plugin's YAML configuration generation to comply with the current ansible-navigator settings schema (<https://docs.ansible.com/projects/navigator/settings/#the-ansible-navigator-settings-file>).

## Scope

**In scope:**

- Identify where the plugin generates the `ansible-navigator.yml` configuration file
- Review the current structure being generated, particularly the `ansible-navigator.ansible.config` section
- Correct the YAML structure to match the official ansible-navigator schema
- Remove or relocate the `defaults` section that's causing the validation error
- Ensure the `ansible.config` section uses only valid properties: `help`, `path`, and `cmdline`

**Out of scope:**

- Changing the plugin's provisioner logic (playbook execution, inventory management, etc.)
- Modifying the HCL2 configuration interface for users
- Updating plugin installation or build processes

## Desired Behavior

**Current (incorrect) structure:**

```yaml
ansible-navigator:
  ansible:
    config:
      defaults:    # ❌ This causes schema validation error
        # ... settings ...
```

**Expected (correct) structure:**

```yaml
ansible-navigator:
  ansible:
    config:
      help: False
      path: /path/to/ansible.cfg    # Optional: path to ansible.cfg
      cmdline: "--forks 15"          # Optional: additional ansible-config cmdline args
```

The plugin should:

1. Generate a valid `ansible-navigator.yml` that passes schema validation
2. Use the correct property names according to ansible-navigator documentation
3. Only include properties that are supported by the schema
4. If Ansible configuration needs to be passed, use either:
   - A separate `ansible.cfg` file referenced via `config.path`
   - Command-line arguments via `config.cmdline`
   - NOT nested configuration inside `config.defaults`

## Constraints & Assumptions

- Assumption: The plugin is written in Go and generates YAML configuration dynamically
- Assumption: The configuration generation likely happens in a provisioner prepare or execution phase
- Constraint: Must maintain backward compatibility with existing HCL2 configuration options where possible
- Constraint: The generated YAML must conform to the ansible-navigator settings schema
- Assumption: The plugin may need to write a separate `ansible.cfg` file if complex Ansible settings are required

## Acceptance Criteria

- [ ] Plugin generates `ansible-navigator.yml` without schema validation errors
- [ ] The `ansible-navigator.ansible.config` section contains only valid properties: `help`, `path`, and/or `cmdline`
- [ ] No `defaults` or other unsupported properties appear under `ansible.config`
- [ ] Running `packer build` with the provisioner completes without ansible-navigator configuration errors
- [ ] Test playbook executes successfully using the generated configuration
- [ ] If Ansible-specific configuration was previously in `defaults`, it's now handled via `ansible.cfg` or command-line args
