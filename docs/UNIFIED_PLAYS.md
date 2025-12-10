# Unified Play Execution Model

This document describes the new unified play execution model introduced in packer-plugin-ansible-navigator, which allows you to execute multiple plays (both playbooks and role FQDNs) with per-play customization and unified dependency management.

## Overview

The unified play execution model provides:

- **Multiple Play Support**: Execute multiple playbooks and/or roles in sequence
- **Role FQDN Support**: Direct execution of Ansible roles by their fully-qualified names
- **Per-Play Customization**: Configure variables, tags, privilege escalation per play
- **Unified Requirements**: Single `requirements_file` for both collections and roles
- **Enhanced Caching**: Separate cache directories for collections and roles with offline mode support
- **Backward Compatibility**: Existing `playbook_file` configurations continue to work

## Configuration

### Basic Configuration

```hcl
provisioner "ansible-navigator" {
  requirements_file = "./requirements.yml"
  
  play {
    name   = "Setup Docker"
    target = "geerlingguy.docker"
  }
}
```

### Complete Configuration Options

```hcl
provisioner "ansible-navigator" {
  # Unified requirements file (replaces separate collections/roles requirements)
  requirements_file = "./requirements.yml"
  
  # Cache directories
  collections_cache_dir = "~/.packer.d/ansible_collections_cache"
  roles_cache_dir       = "~/.packer.d/ansible_roles_cache"
  
  # Dependency management
  offline_mode = false  # Use only cached dependencies
  force_update = false  # Always reinstall dependencies
  
  # Play definitions
  play {
    name       = "Play Name"
    target     = "role.fqdn or playbook.yml"
    extra_vars = { key = "value" }
    tags       = ["tag1", "tag2"]
    vars_files = ["vars/file.yml"]
    become     = true
  }
}
```

## Play Configuration

### Play Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | No | Descriptive name for logging (auto-generated if omitted) |
| `target` | string | Yes | Playbook path (`.yml`/`.yaml`) or role FQDN |
| `extra_vars` | map | No | Variables to pass to the play |
| `tags` | list | No | Ansible tags to execute |
| `vars_files` | list | No | Variable files to load |
| `become` | bool | No | Enable privilege escalation |

### Target Types

The `target` field accepts two types:

**1. Playbook Files**

```hcl
target = "site.yml"
target = "./playbooks/deploy.yaml"
```

**2. Role FQDNs (Fully-Qualified Domain Names)**

```hcl
target = "geerlingguy.docker"
target = "namespace.collection.role"
target = "simple_role_name"
```

When using a role FQDN, the provisioner automatically generates a temporary playbook that executes the role with the specified configuration.

## Requirements File

The `requirements_file` supports both collections and roles in a single file:

```yaml
---
# Collections
collections:
  - name: community.general
    version: ">=5.0.0"
  - name: ansible.posix

# Roles
roles:
  - name: geerlingguy.docker
    version: "6.1.0"
  - src: https://github.com/user/ansible-role-custom
    name: custom.role
```

## Examples

### Example 1: Multiple Roles

```hcl
provisioner "ansible-navigator" {
  requirements_file = "./requirements.yml"
  
  play {
    name   = "Install Docker"
    target = "geerlingguy.docker"
    extra_vars = {
      docker_install_compose = "true"
    }
  }
  
  play {
    name   = "Install Node.js"
    target = "geerlingguy.nodejs"
    become = true
  }
}
```

### Example 2: Mix of Playbooks and Roles

```hcl
provisioner "ansible-navigator" {
  requirements_file = "./requirements.yml"
  
  play {
    name   = "Base configuration"
    target = "site.yml"
  }
  
  play {
    name   = "Install Docker"
    target = "geerlingguy.docker"
    extra_vars = {
      docker_edition = "ce"
    }
  }
  
  play {
    name       = "Deploy application"
    target     = "deploy.yml"
    become     = true
    tags       = ["deploy"]
    vars_files = ["vars/production.yml"]
  }
}
```

### Example 3: Per-Play Customization

```hcl
provisioner "ansible-navigator" {
  requirements_file = "./requirements.yml"
  
  play {
    name       = "Security hardening"
    target     = "security.hardening"
    become     = true
    tags       = ["security", "hardening"]
    extra_vars = {
      security_level = "high"
    }
  }
  
  play {
    name       = "Application deployment"
    target     = "myorg.webapp"
    vars_files = ["vars/app.yml", "secrets/prod.yml"]
    tags       = ["deploy"]
  }
}
```

### Example 4: Offline Mode

```hcl
provisioner "ansible-navigator" {
  requirements_file     = "./requirements.yml"
  collections_cache_dir = "./cache/collections"
  roles_cache_dir       = "./cache/roles"
  offline_mode          = true
  
  play {
    target = "geerlingguy.docker"
  }
}
```

## Caching and Offline Mode

### Cache Directories

By default, dependencies are cached in:

- Collections: `~/.packer.d/ansible_collections_cache`
- Roles: `~/.packer.d/ansible_roles_cache`

### Offline Mode

When `offline_mode = true`:

- No network requests are made
- All dependencies must be present in cache
- Build fails if required dependencies are missing

### Force Update

When `force_update = true`:

- All dependencies are reinstalled from source
- Existing cache is overwritten
- Useful for ensuring latest versions

## Migration from playbook_file

### Before (Deprecated)

```hcl
provisioner "ansible-navigator" {
  playbook_file = "site.yml"
  galaxy_file   = "requirements.yml"
}
```

### After (Recommended)

```hcl
provisioner "ansible-navigator" {
  requirements_file = "requirements.yml"
  
  play {
    target = "site.yml"
  }
}
```

**Note**: The `playbook_file` field is still supported but deprecated, and a warning will be displayed if used. Legacy `plays { }` blocks and `plays = [...]` array syntax are no longer supported and must be rewritten as repeated `play { ... }` blocks.

## Role Playbook Generation

When a role FQDN is specified as a target, the provisioner automatically generates a temporary playbook:

```yaml
---
- hosts: all
  become: yes  # if become: true
  vars_files:  # if vars_files specified
    - vars/file.yml
  roles:
    - role: geerlingguy.docker
      vars:  # if extra_vars specified
        docker_install_compose: true
```

The temporary playbook is automatically cleaned up after execution.

## Error Handling

The provisioner provides detailed error messages:

| Scenario | Error Message |
|----------|---------------|
| Missing dependency | `Error: Unable to locate required dependency from requirements.yml` |
| Invalid play target | `Invalid play target 'foo.txt' — must be .yml/.yaml or role FQDN` |
| Play failure | `Play 'Deploy web stack' failed with exit code 2` |
| Offline missing | `Role 'geerlingguy.docker' not found and offline mode enabled` |

## Best Practices

1. **Use Descriptive Names**: Always provide meaningful `name` fields for plays
2. **Order Matters**: Plays execute sequentially; order them logically
3. **Cache Dependencies**: Use cache directories for faster builds
4. **Test Offline**: Verify offline mode works before production use
5. **Version Lock**: Specify exact versions in requirements.yml for reproducibility
6. **Separate Concerns**: Use different plays for distinct configuration phases

## Environment Variables

The provisioner automatically sets:

```bash
ANSIBLE_COLLECTIONS_PATHS=/path/to/collections_cache_dir
ANSIBLE_ROLES_PATH=/path/to/roles_cache_dir
```

These paths are prepended to existing environment variables if present.

## Compatibility

- **Go Version**: 1.25.3
- **Ansible Navigator**: All versions supporting `ansible-navigator run`
- **Packer**: Compatible with Packer Plugin SDK v0.6.4+

## Troubleshooting

### Plays not found

Ensure dependencies are installed:

```bash
ansible-galaxy install -r requirements.yml
```

### Role not found

Check the role name and verify it exists in:

- Galaxy: <https://galaxy.ansible.com>
- Requirements file
- Local cache directory

### Permission denied

Enable privilege escalation:

```hcl
play {
  target = "role.name"
  become = true
}
```

## See Also

- [Ansible Galaxy Requirements](https://docs.ansible.com/ansible/latest/collections_guide/collections_installing.html)
- [Ansible Navigator Documentation](https://ansible.readthedocs.io/projects/navigator/)
- [Packer Plugin SDK](https://github.com/hashicorp/packer-plugin-sdk)
