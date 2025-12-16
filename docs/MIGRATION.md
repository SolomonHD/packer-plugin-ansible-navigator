# Migration Guide: Legacy Options to navigator_config

> **📢 Deprecation Notice:** Several legacy configuration options are deprecated and will be removed in a future version (1-2 releases). Please migrate to the recommended `navigator_config` approach.

## Overview

This guide provides step-by-step instructions for migrating from deprecated options to the modern `navigator_config` field, which aligns with ansible-navigator v3+ best practices.

## Deprecation Timeline

1. **Current release**: Warnings added, deprecated options still work
2. **1-2 releases later**: Deprecated options will be removed entirely

**Action required:** Update your configurations during this grace period to avoid breakage.

## Deprecated Options Summary

| Deprecated Option | Replacement | Status |
|-------------------|-------------|--------|
| `execution_environment` | `navigator_config.execution-environment` | ❌ Removed |
| `navigator_mode` | `navigator_config.mode` | ❌ Removed |
| `ansible_cfg` | `navigator_config.ansible.config` | ❌ Removed |
| `ansible_env_vars` | `navigator_config.execution-environment.environment-variables` | ❌ Removed |
| `ansible_ssh_extra_args` | Play-level options or `navigator_config` | ❌ Removed |
| `extra_arguments` | `navigator_config` or play-level options | ❌ Removed |
| `roles_cache_dir` | `roles_path` | ❌ Renamed (breaking) |
| `collections_cache_dir` | `collections_path` | ❌ Renamed (breaking) |
| `force_update` | `galaxy_force` / `galaxy_force_with_deps` | ❌ Removed |
| `galaxy_force_install` | `galaxy_force` | ❌ Removed |

## Migration Examples

### 1. Migrating `execution_environment`

**Before (deprecated):**

```hcl
provisioner "ansible-navigator" {
  execution_environment = "quay.io/ansible/creator-ee:latest"
  
  play {
    target = "site.yml"
  }
}
```

**After (recommended):**

```hcl
provisioner "ansible-navigator" {
  navigator_config = {
    execution-environment = {
      enabled = true
      image = "quay.io/ansible/creator-ee:latest"
      pull-policy = "missing"
    }
  }
  
  play {
    target = "site.yml"
  }
}
```

**Benefits:**

- More explicit control over execution environment behavior
- Automatic safe defaults for temp directories
- Better alignment with ansible-navigator configuration schema

### 2. Migrating `navigator_mode`

**Before (deprecated):**

```hcl
provisioner "ansible-navigator" {
  navigator_mode = "json"
  structured_logging = true
  
  play {
    target = "site.yml"
  }
}
```

**After (recommended):**

```hcl
provisioner "ansible-navigator" {
  navigator_config = {
    mode = "json"
  }
  
  structured_logging = true
  
  play {
    target = "site.yml"
  }
}
```

### 3. Migrating `ansible_cfg`

**Before (deprecated):**

```hcl
provisioner "ansible-navigator" {
  ansible_cfg = {
    defaults = {
      remote_tmp = "/tmp/.ansible/tmp"
      local_tmp = "/tmp/.ansible-local"
      host_key_checking = "False"
    }
  }
  
  play {
    target = "site.yml"
  }
}
```

**After (recommended):**

```hcl
provisioner "ansible-navigator" {
  navigator_config = {
    ansible = {
      config = {
        defaults = {
          remote_tmp = "/tmp/.ansible/tmp"
          local_tmp = "/tmp/.ansible-local"
          host_key_checking = "False"
        }
      }
    }
  }
  
  play {
    target = "site.yml"
  }
}
```

### 4. Migrating `ansible_env_vars`

**Before (deprecated):**

```hcl
provisioner "ansible-navigator" {
  execution_environment = "quay.io/ansible/creator-ee:latest"
  ansible_env_vars = {
    ANSIBLE_REMOTE_TMP = "/custom/tmp"
    CUSTOM_VAR = "production"
  }
  
  play {
    target = "site.yml"
  }
}
```

**After (recommended):**

```hcl
provisioner "ansible-navigator" {
  navigator_config = {
    execution-environment = {
      enabled = true
      image = "quay.io/ansible/creator-ee:latest"
      environment-variables = {
        ANSIBLE_REMOTE_TMP = "/custom/tmp"
        CUSTOM_VAR = "production"
      }
    }
  }
  
  play {
    target = "site.yml"
  }
}
```

### 5. Migrating `ansible_ssh_extra_args`

**Before (deprecated):**

```hcl
provisioner "ansible-navigator" {
  ansible_ssh_extra_args = ["-o ControlMaster=auto", "-o ControlPersist=60s"]
  
  play {
    target = "site.yml"
  }
}
```

**After (recommended - Option A: navigator_config):**

```hcl
provisioner "ansible-navigator" {
  navigator_config = {
    ansible = {
      config = {
        ssh_connection = {
          ssh_args = "-o ControlMaster=auto -o ControlPersist=60s"
        }
      }
    }
  }
  
  play {
    target = "site.yml"
  }
}
```

**After (recommended - Option B: ansible.cfg file):**

Create an `ansible.cfg` file in your project and reference it via `navigator_config`:

```ini
# ansible.cfg
[ssh_connection]
ssh_args = -o ControlMaster=auto -o ControlPersist=60s
```

```hcl
provisioner "ansible-navigator" {
  navigator_config = {
    ansible = {
      config = "./ansible.cfg"  # Reference external file
    }
  }
  
  play {
    target = "site.yml"
  }
}
```

### 6. Migrating `extra_arguments`

**Before (deprecated):**

```hcl
provisioner "ansible-navigator" {
  extra_arguments = ["--verbose", "--timeout=60"]
  
  play {
    target = "site.yml"
  }
}
```

**After (recommended):**

Most `extra_arguments` use cases should be replaced with `navigator_config` settings or play-level options. For verbosity:

```hcl
provisioner "ansible-navigator" {
  navigator_config = {
    mode = "stdout"
    logging = {
      level = "debug"
    }
  }
  
  verbose_task_output = true
  
  play {
    target = "site.yml"
  }
}
```

### 7. Migrating `roles_cache_dir` and `collections_cache_dir`

**Before (removed):**

```hcl
provisioner "ansible-navigator" {
  roles_cache_dir       = "./.ansible/roles"
  collections_cache_dir = "./.ansible/collections"
  
  play {
    target = "myapp.deploy"
  }
}
```

**After (recommended):**

Use the renamed install destination fields:

```hcl
provisioner "ansible-navigator" {
  requirements_file = "./requirements.yml"
  roles_path        = "./.ansible/roles"
  collections_path  = "./.ansible/collections"
  
  play {
    target = "myapp.deploy"
  }
}
```

Notes:

- `roles_path` and `collections_path` are used as **Galaxy install destinations** and are exported to Ansible via `ANSIBLE_ROLES_PATH` and `ANSIBLE_COLLECTIONS_PATHS`.
- `requirements_file` remains the only supported dependency declaration mechanism.

### 8. Migrating `force_update` and `galaxy_force_install`

**Before (removed):**

```hcl
provisioner "ansible-navigator" {
  requirements_file = "./requirements.yml"
  force_update      = true
  galaxy_force_install = true
  
  play {
    target = "site.yml"
  }
}
```

**After (recommended):**

```hcl
provisioner "ansible-navigator" {
  requirements_file = "./requirements.yml"

  galaxy_force           = true
  galaxy_force_with_deps = true
  
  play {
    target = "site.yml"
  }
}
```

Notes:

- `galaxy_force_with_deps` takes precedence over `galaxy_force` (the provisioner passes `--force-with-deps` and omits `--force`).

## Complete Migration Example

**Before (using multiple deprecated options):**

```hcl
provisioner "ansible-navigator" {
  execution_environment = "quay.io/ansible/creator-ee:latest"
  navigator_mode = "json"
  ansible_cfg = {
    defaults = {
      remote_tmp = "/tmp/.ansible/tmp"
      local_tmp = "/tmp/.ansible-local"
    }
  }
  ansible_env_vars = {
    CUSTOM_VAR = "production"
  }
  ansible_ssh_extra_args = ["-o ControlMaster=auto"]
  extra_arguments = ["--verbose"]
  roles_path = "./roles"
  collections_path = "./collections"
  
  play {
    target = "site.yml"
  }
}
```

**After (modern navigator_config approach):**

```hcl
provisioner "ansible-navigator" {
  navigator_config = {
    mode = "json"
    execution-environment = {
      enabled = true
      image = "quay.io/ansible/creator-ee:latest"
      pull-policy = "missing"
      environment-variables = {
        CUSTOM_VAR = "production"
      }
    }
    ansible = {
      config = {
        defaults = {
          remote_tmp = "/tmp/.ansible/tmp"
          local_tmp = "/tmp/.ansible-local"
          roles_path = "./roles"
          collections_path = "./collections"
        }
        ssh_connection = {
          ssh_args = "-o ControlMaster=auto"
        }
      }
    }
    logging = {
      level = "debug"
    }
  }
  
  structured_logging = true
  verbose_task_output = true
  
  play {
    target = "site.yml"
  }
}
```

## Benefits of Migrating

1. **Future-proof**: Aligns with ansible-navigator v3+ and future versions
2. **Single source of truth**: All navigator settings in one place
3. **Better defaults**: Automatic safe EE temp directory configuration
4. **More explicit**: Clearer what each setting controls
5. **Easier debugging**: Configuration matches ansible-navigator.yml schema

## Getting Help

If you encounter issues during migration:

1. Check deprecation warnings in Packer output
2. Review [Configuration Reference](CONFIGURATION.md)
3. See [Examples](EXAMPLES.md) for more patterns
4. Open an issue: [GitHub Issues](https://github.com/solomonhd/packer-plugin-ansible-navigator/issues)

## Timeline Reminder

- **Now**: Update your configurations to use `navigator_config`
- **1-2 releases**: Deprecated options will be removed
- **Future**: Only `navigator_config` and modern options will be supported

---

[← Back to docs index](README.md)
