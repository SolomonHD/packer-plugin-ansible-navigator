# OpenSpec Change Prompt: Ansible.cfg Generation Feature

## Context

The Packer Ansible Navigator plugin currently has scattered configuration options that often need to be translated into Ansible settings. A specific issue arises when using execution environments: ansible-runner defaults to creating temporary directories at `/.ansible/tmp`, which fails when the container runs as a non-root user.

Current workarounds require users to:

- Manually create an `ansible.cfg` file in their working directory
- Use `extra_arguments` to pass `-e ansible_remote_tmp=/tmp/.ansible/tmp`
- Set environment variables that may not be properly inherited by the container

This is error-prone and requires understanding of both Packer plugin configuration AND Ansible configuration.

## Goal

Add a new `ansible_cfg` configuration option to the plugin that:

1. Accepts a map/object of Ansible configuration settings
2. Automatically generates a temporary `ansible.cfg` file before provisioning
3. Passes the path to ansible-navigator via `ANSIBLE_CONFIG` environment variable
4. Cleans up the temporary file after provisioning
5. Provides sensible defaults for execution environment use cases

## Scope

### In scope

- New `ansible_cfg` configuration option (map structure)
- Temporary ansible.cfg file generation logic
- Default settings for execution environment compatibility
- Cleanup of generated ansible.cfg file
- Documentation of the new option
- Backward compatibility with existing configurations

### Out of scope

- Modifying user-provided ansible.cfg files
- Complex ansible.cfg validation beyond basic INI formatting
- Support for ansible.cfg includes or advanced features
- Changes to how environment variables are currently handled (separate concern)

## Desired Behavior

### Configuration Example

```hcl
provisioner "ansible-navigator" {
  play {
    name = "Configure system"
    target = "integration.common_tasks.setup"
  }
  
  execution_environment = "containers.github.service.emory.edu/myorg/ee:latest"
  
  # NEW: ansible_cfg option
  ansible_cfg = {
    defaults = {
      remote_tmp      = "/tmp/.ansible/tmp"
      local_tmp       = "/tmp/.ansible-local"
      host_key_checking = "False"
      timeout         = "30"
    }
    ssh_connection = {
      pipelining      = "True"
      ssh_args        = "-o ControlMaster=auto -o ControlPersist=60s"
    }
  }
}
```

### Generated File Example

The plugin should generate `/tmp/packer-ansible-cfg-RANDOM.ini`:

```ini
[defaults]
remote_tmp = /tmp/.ansible/tmp
local_tmp = /tmp/.ansible-local
host_key_checking = False
timeout = 30

[ssh_connection]
pipelining = True
ssh_args = -o ControlMaster=auto -o ControlPersist=60s
```

### Execution Behavior

1. **Before ansible-navigator runs:**
   - Generate temporary ansible.cfg from `ansible_cfg` map
   - Set `ANSIBLE_CONFIG=/tmp/packer-ansible-cfg-RANDOM.ini`
   - Add to cleanup list

2. **ansible-navigator execution:**
   - Honors the generated ansible.cfg via ANSIBLE_CONFIG
   - Container inherits ANSIBLE_CONFIG environment variable
   - Ansible inside container uses the specified settings

3. **After provisioning:**
   - Delete temporary ansible.cfg file
   - Unset ANSIBLE_CONFIG (if it was set by plugin)

## Constraints & Assumptions

### Assumptions

- Users may already have an `ansible.cfg` in their working directory (should take precedence)
- The generated file should be temporary and not committed to source control
- INI file structure for ansible.cfg is sufficient (no YAML format needed)
- The plugin has write access to `/tmp` or system temporary directory

### Constraints

- Must NOT overwrite user-provided ansible.cfg files
- Must work with both local and containerized ansible-navigator
- Changes must be backward compatible (existing configs still work)
- Generated file must be cleaned up even if provisioning fails

## Default Settings for Execution Environments

When `execution_environment` is set, the plugin should apply these defaults (unless overridden by user):

```hcl
ansible_cfg = {
  defaults = {
    remote_tmp = "/tmp/.ansible/tmp"      # Writable for non-root container users
    local_tmp  = "/tmp/.ansible-local"    # Writable for non-root container users
  }
}
```

These defaults solve the "Permission denied: /.ansible" error that occurs when containers run as non-root.

## Acceptance Criteria

- [ ] New `ansible_cfg` configuration option accepts nested map structure
- [ ] Plugin generates valid ansible.cfg INI file from the map
- [ ] Generated file is created in temporary directory with random name
- [ ] ANSIBLE_CONFIG environment variable points to generated file
- [ ] Temporary file is cleaned up after provisioning (success or failure)
- [ ] Default settings are applied when `execution_environment` is used
- [ ] User-provided ansible.cfg in working directory takes precedence
- [ ] Backward compatible: existing configurations work without changes
- [ ] Documentation updated with examples and explanation
- [ ] Works with both ansible-navigator modes (stdout, json, interactive)

## Related Issues

- Fixes: "Permission denied: /.ansible" error when using execution environments
- Consolidates: Scattered ansible configuration options
- Improves: User experience by reducing boilerplate configuration
