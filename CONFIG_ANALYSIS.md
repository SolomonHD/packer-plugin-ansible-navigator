# Config Options Analysis: Potential Redundancies After ansible_cfg

## Summary

After introducing `ansible_cfg` as a unified way to configure Ansible settings, **we can DEPRECATE but NOT remove** most scattered config options that map directly to ansible.cfg settings. This maintains backward compatibility while encouraging migration to the cleaner approach.

## Configuration Options Analysis

### ✅ KEEP (Not Redundant - Plugin-Specific)

These options control the plugin's behavior, not Ansible's:

| Option | Reason to Keep |
|--------|----------------|
| `command` | Specifies ansible-navigator executable path |
| `ansible_navigator_path` | PATH manipulation for finding navigator |
| `navigator_mode` | Plugin-level control of navigator output mode |
| `execution_environment` | EE image name (though could also be in ansible.cfg) |
| `keep_going` | Plugin control for multi-play execution |
| `structured_logging` | Plugin feature for parsing JSON output |
| `log_output_path` | Plugin feature for structured logs |
| `verbose_task_output` | Plugin feature for log verbosity |
| `skip_version_check` | Plugin startup behavior |
| `version_check_timeout` | Plugin startup behavior |
| `playbook_file` | (Deprecated) Play target specification |
| `plays` | Play target specification |
|  `requirements_file` | Galaxy dependencies management |
| `roles_cache_dir` | Plugin-managed cache location |
| `offline_mode` | Plugin control of galaxy operations |
| `force_update` | Plugin control of galaxy operations |
| `collections` | Plugin list of collections to install |
| `collections_cache_dir` | Plugin-managed cache location |
| `collections_offline` | Plugin control |
| `collections_force_update` | Plugin control |
| `collections_requirements` | Galaxy dependencies |
| `galaxy_file` | Galaxy dependencies (deprecated) |
| `galaxy_command` | Galaxy executable path |
| `galaxy_force_install` | Galaxy operation control |
| `galaxy_force_with_deps` | Galaxy operation control |
| `roles_path` | Galaxy install location |
| `collections_path` | Galaxy install location |
| `groups` | Inventory generation control |
| `empty_groups` | Inventory generation control |
| `host_alias` | Inventory generation control |
| `user` | Inventory generation control (though could be ansible.cfg)|
| `local_port` | SSH proxy configuration |
| `ssh_host_key_file` | SSH proxy configuration |
| `ssh_authorized_key_file` | SSH proxy configuration |
| `ansible_proxy_key_type` | SSH proxy configuration |
| `ansible_proxy_bind_address` | SSH proxy configuration |
| `ansible_proxy_host` | SSH proxy configuration |
| `sftp_command` | SFTP server command (though could be ansible.cfg) |
| `use_sftp` | Plugin control |
| `inventory_directory` | Inventory generation control |
| `inventory_file_template` | Inventory generation control |
| `inventory_file` | Inventory file location |
| `limit` | Playbook execution limit |
| `keep_inventory_file` | Plugin cleanup behavior |
| `use_proxy` | Plugin architecture decision |
| `winrm_use_http` | Ansible WinRM config (convenience) |

### ⚠️ DEPRECATE (Map to ansible.cfg, but keep for backward compat)

These options are essentially wrappers around Ansible configuration:

| Option | Maps to ansible.cfg | Reasoning |
|--------|---------------------|-----------|
| `ansible_ssh_extra_args` | `[ssh_connection] ssh_args` | Direct ansible.cfg setting |
| `ansible_env_vars` | Environment variables | **KEEP** - These are broader than ansible.cfg (e.g., PYTHONUNBUFFERED) |
| `extra_arguments` | Various CLI flags | **KEEP** - Flexible catch-all for any ansible-playbook args |

### 🔄 Current State

**None of the current options are truly redundant** because:

1. **Most control plugin behavior**, not Ansible behavior
2. **Many are convenience wrappers** that set Ansible settings in a user-friendly way
3. **Some combine multiple Ansible settings** into a single option
4. **Backward compatibility** is critical  

## What `ansible_cfg` Actually Replaces

The `ansible_cfg` option primarily replaces the **current workaround pattern**:

### Before `ansible_cfg`

```hcl
provisioner "ansible-navigator" {
  # Users have to know which Ansible settings to pass as extra_arguments
  extra_arguments = [
    "-e", "ansible_remote_tmp=/tmp/.ansible/tmp",
    "-e", "ansible_local_tmp=/tmp/.ansible-local",
    "-e", "timeout=30"
  ]
  
  # Or use environment variables which may not work in containers
  ansible_env_vars = [
    "ANSIBLE_TIMEOUT=30"
  ]
}
```

### After `ansible_cfg`

```hcl
provisioner "ansible-navigator" {
  # Clean, type-safe ansible.cfg configuration
  ansible_cfg = {
    defaults = {
      remote_tmp = "/tmp/.ansible/tmp"
      local_tmp  = "/tmp/.ansible-local"
      timeout    = "30"
    }
  }
}
```

## Proposed Deprecation Strategy

### Phase 1: Introduce `ansible_cfg` (v4.0.0)

- Add `ansible_cfg` option
- Apply EE defaults automatically
- Keep all existing options working
- Document migration path

### Phase 2: Soft Deprecation (v4.1.0)

- Log warnings when using patterns that could be replaced with `ansible_cfg`
- Example: "ansible_ssh_extra_args can be replaced with ansible_cfg.ssh_connection.ssh_args"
- Keep everything functional

### Phase 3: Hard Deprecation (v5.0.0)

- Move documentation focus entirely to `ansible_cfg`
- Label old options as "legacy" in docs
- Still functional, but clearly discouraged

### Phase 4: Removal (v6.0.0 or later)

- **ONLY if** usage metrics show minimal adoption of deprecated options
- **Never remove** if users rely on them heavily

## Special Cases

### `ansible_ssh_extra_args`

This could be deprecated in favor of:

```hcl
ansible_cfg = {
  ssh_connection = {
    ssh_args = "-o ControlMaster=auto -o ControlPersist=60s"
  }
}
```

But it's still useful as a convenience, so: **KEEP with soft deprecation**

### `ansible_env_vars`

This sets environment variables for the navigator process itself, not just Ansible config.
Example: `PYTHONUNBUFFERED=1` affects Python behavior, not Ansible config.
**KEEP - Not redundant**

### `extra_arguments`

This is a flexible catch-all for any ansible-playbook arguments. Can pass things like:

- `--vault-id`
- `--start-at-task`
- `--step`

These don't all map cleanly to ansible.cfg. **KEEP - Not redundant**

## Recommendation

**Do NOT remove any existing options.** Instead:

1. ✅ Add `ansible_cfg` as the **preferred** way to configure Ansible
2. ✅ Apply smart defaults when `execution_environment` is set
3. ✅ Document `ansible_cfg` prominently
4. ✅ Show migration examples in changelog
5. ⚠️ Add soft deprecation warnings (v4.1+) for patterns replaced by `ansible_cfg`:
   - Using `ansible_ssh_extra_args` for settings that could be in `ansible_cfg`
   - Using `extra_arguments` with `-e ansible_*` variables
6. ⚠️ Update documentation to show `ansible_cfg` as the primary method
7. ❌ Do NOT remove any options until v6.0+ and only with strong usage data supporting it

## Benefits of Not Removing Options

1. **Backward Compatibility** - Existing configurations continue to work
2. **Migration Flexibility** - Users can migrate at their own pace
3. **Convenience** - Some simple cases don't need full ansible.cfg
4. **Emergency Overrides** - `extra_arguments` as escape hatch is valuable
5. **Reduced Breaking Changes** - Fewer frustrated users

## The Real Value of `ansible_cfg`

The value is NOT in removing redundant options. The value is:

1. **Cleaner new configurations** - Modern HCL-native syntax
2. **Automatic EE defaults** - Solves the `/.ansible` permission problem automatically
3. **Type safety** - Map structure vs string arrays
4. **Discoverability** - IDE autocomplete can help
5. **Consistency** - One place for Ansible configuration instead of scattered options

## Conclusion

**Answer to your question**: No, we should NOT remove any config options after adding `ansible_cfg`.

The new option complements existing ones and provides a better path forward, but existing options remain useful for:

- Backward compatibility
- Simple use cases
- Override/escape hatch scenarios
- Non-Ansible settings (env vars, plugin behavior)

Focus the implementation on making `ansible_cfg` the **preferred** option, not on deprecating/removing existing ones.
