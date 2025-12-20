# Migration Guide

## v6.1.0: Collections Mount and Environment Variable Updates

### Overview

Version 6.1.0 fixes collections support in execution environments and updates the collections path environment variable to use the non-deprecated singular form.

### Changes

#### 1. Automatic Collections Mounting (No Action Required)

Starting with v6.1.0, collections are automatically mounted into execution environment containers when `navigator_config.execution_environment.enabled = true`. No manual volume mount configuration is needed.

**What this means:**

- Collections installed from `requirements_file` to `~/.packer.d/ansible_collections_cache/ansible_collections` are automatically mounted as read-only volumes in the EE container
- `ANSIBLE_COLLECTIONS_PATH` is automatically set inside the container to point to the mounted collections
- Collection roles now work correctly inside execution environments without additional configuration

**Before (v6.0.0 and earlier - broken):**

```hcl
provisioner "ansible-navigator" {
  requirements_file = "./requirements.yml"
  
  navigator_config {
    execution_environment {
      enabled = true
      image = "quay.io/ansible/creator-ee:latest"
    }
  }
  
  play {
    target = "namespace.collection.role_name"
    # This would fail with "unable to find role"
  }
}
```

**After (v6.1.0+ - works):**

```hcl
provisioner "ansible-navigator" {
  requirements_file = "./requirements.yml"
  
  navigator_config {
    execution_environment {
      enabled = true
      image = "quay.io/ansible/creator-ee:latest"
    }
  }
  
  play {
    target = "namespace.collection.role_name"
    # Collections are automatically mounted - works correctly
  }
}
```

#### 2. Environment Variable Name Change (Internal Only)

The plugin now uses `ANSIBLE_COLLECTIONS_PATH` (singular) instead of the deprecated `ANSIBLE_COLLECTIONS_PATHS` (plural) to eliminate Ansible deprecation warnings.

**User Impact**: None - this is an internal change. Your HCL configurations do not need any updates.

### Migration Steps

1. **Upgrade to v6.1.0+**
2. **Remove any manual volume mount workarounds** if you were using them
3. **Test your collection-based playbooks with execution environments**

No HCL configuration changes required - collections mounting is now automatic.

---

## v4.2.0: Navigator Config Structure Updates

### Overview

Version 4.2.0 introduces breaking changes to the `environment_variables` and `ansible_config` structures within `navigator_config` to correctly model the ansible-navigator.yml schema and fix HCL2 spec generation issues.

### Breaking Changes

#### 1. Environment Variables: Inline Keys → Pass/Set Structure

**Before (v4.1.x):**

```hcl
navigator_config {
  execution_environment {
    environment_variables {
      ANSIBLE_REMOTE_TMP = "/tmp/.ansible/tmp"
      CUSTOM_VAR = "value"
    }
  }
}
```

**After (v4.2.0+):**

```hcl
navigator_config {
  execution_environment {
    environment_variables {
      pass = ["SSH_AUTH_SOCK", "AWS_ACCESS_KEY_ID"]
      set = {
        ANSIBLE_REMOTE_TMP = "/tmp/.ansible/tmp"
        CUSTOM_VAR        = "value"
      }
    }
  }
}
```

**Why?**: This matches the actual ansible-navigator.yml schema which has `pass` (list of variables to pass from host) and `set` (explicit key-value pairs) sections.

#### 2. Ansible Config: Removed Inner Struct

**Before (v4.1.x):**

The `ansible_config` structure had an implicit `Inner` field with `mapstructure:",squash"` that didn't properly generate HCL2 specs.

**After (v4.2.0+):**

```hcl
navigator_config {
  ansible_config {
    config = "/etc/ansible/ansible.cfg"
    
    defaults {
      remote_tmp        = "/tmp/.ansible/tmp"
      host_key_checking = false
    }
    
    ssh_connection {
      ssh_timeout = 30
      pipelining  = true
    }
  }
}
```

The `defaults` and `ssh_connection` blocks are now direct children of `ansible_config` instead of being nested within an implicit `Inner` struct.

### Migration Steps

1. **Update `environment_variables` blocks**:
   - Wrap inline variable assignments in a `set` map
   - Optionally add a `pass` list for pass-through variables

2. **Update `ansible_config` blocks**:
   - Ensure `defaults` and `ssh_connection` are direct children
   - Verify the `config` field uses a simple string assignment

3. **Test with `packer validate`**

### Complete Migration Example

**Before (v4.1.x):**

```hcl
navigator_config {
  execution_environment {
    environment_variables {
      ANSIBLE_REMOTE_TMP = "/tmp/.ansible/tmp"
      ANSIBLE_LOCAL_TMP  = "/tmp/.ansible-local"
    }
  }
}
```

**After (v4.2.0+):**

```hcl
navigator_config {
  execution_environment {
    environment_variables {
      pass = ["SSH_AUTH_SOCK"]
      set = {
        ANSIBLE_REMOTE_TMP = "/tmp/.ansible/tmp"
        ANSIBLE_LOCAL_TMP  = "/tmp/.ansible-local"
      }
    }
  }
}
```

---

## v4.1.0: Navigator Config Breaking Change

### Overview

Version 4.1.0 introduces a breaking change to the `navigator_config` field to fix a critical RPC serialization bug that prevented the plugin from working when `navigator_config` was specified.

### Problem Fixed

Previous versions used `map[string]interface{}` with `cty.DynamicPseudoType` for the `navigator_config` field, which caused the plugin to crash with:

```
unsupported cty.Type conversion from cty.pseudoTypeDynamic
```

### Breaking Changes

#### 1. Syntax Change: Map Assignment → Block Syntax

**Before (v4.0.x and earlier):**

```hcl
provisioner "ansible-navigator" {
  navigator_config = {
    mode = "stdout"
    execution-environment = {
      enabled = true
    }
  }
}
```

**After (v4.1.0+):**

```hcl
provisioner "ansible-navigator" {
  navigator_config {
    mode = "stdout"
    execution_environment {
      enabled = true
    }
  }
}
```

Notice:

- No `=` between `navigator_config` and the opening `{`
- This is now a **block**, not amap assignment

#### 2. Field Names: Hyphens → Underscores

Field names in HCL configuration now use underscores instead of hyphens:

| Old Field Name (Hyphens) | New Field Name (Underscores) |
|--------------------------|------------------------------|
| `execution-environment` | `execution_environment` |
| `environment-variables` | `environment_variables` |
| `pull-policy` | `pull_policy` |
| `playbook-artifact` | `playbook_artifact` |
| `save-as` | `save_as` |
| `collection-doc-cache` | `collection_doc_cache` |

**Why?**: Go struct field names cannot contain hyphens, and HCL block field names must match the struct tags.

#### 3. Nested Blocks Use Block Syntax

All nested configuration sections use block syntax (no `=`):

**Before:**

```hcl
navigator_config = {
  execution-environment = {
    environment-variables = {
      pass = ["SSH_AUTH_SOCK"]
    }
  }
}
```

**After:**

```hcl
navigator_config {
  execution_environment {
    environment_variables {
      pass = ["SSH_AUTH_SOCK"]
    }
  }
}
```

### Migration Steps

1. **Replace `navigator_config =` with `navigator_config`** (remove the `=`)
2. **Replace all hyphenated field names with underscored equivalents**
3. **Ensure all nested structures use block syntax, not map assignment**
4. **Test configuration with `packer validate`**

### Complete Migration Example

**Before (v4.0.x):**

```hcl
provisioner "ansible-navigator" {
  play {
    name   = "Configure system"
    target = "./playbook.yml"
  }

  navigator_config = {
    mode = "stdout"
    
    execution-environment = {
      enabled      = true
      image        = "quay.io/ansible/creator-ee:latest"
      pull-policy  = "missing"
      
      environment-variables = {
        pass = ["SSH_AUTH_SOCK"]
        ANSIBLE_REMOTE_TMP = "/tmp/.ansible/tmp"
      }
    }
    
    logging = {
      level = "debug"
      file  = "/tmp/navigator.log"
    }
    
    playbook-artifact = {
      enable  = true
      save-as = "/tmp/artifact.json"
    }
  }
}
```

**After (v4.1.0+):**

```hcl
provisioner "ansible-navigator" {
  play {
    name   = "Configure system"
    target = "./playbook.yml"
  }

  navigator_config {
    mode = "stdout"
    
    execution_environment {
      enabled     = true
      image       = "quay.io/ansible/creator-ee:latest"
      pull_policy = "missing"
      
      environment_variables {
        pass               = ["SSH_AUTH_SOCK"]
        ANSIBLE_REMOTE_TMP = "/tmp/.ansible/tmp"
      }
    }
    
    logging {
      level = "debug"
      file  = "/tmp/navigator.log"
    }
    
    playbook_artifact {
      enable  = true
      save_as = "/tmp/artifact.json"
    }
  }
}
```

### Benefits of This Change

1. **Plugin now works** - No more RPC serialization crashes
2. **Type safety** - Full compile-time type checking
3. **IDE support** - Autocomplete and validation in HCL-aware editors
4. **Better error messages** - Clear validation errors instead of runtime crashes
5. **Future-proof** - Aligns with Packer Plugin SDK best practices

### Troubleshooting

#### Error: "unsupported cty.Type conversion"

This means you're using an old version of the plugin with the old syntax. Upgrade to v4.1.0+ and follow this migration guide.

#### Error: "An argument named 'navigator_config' is not expected here"

You're using the old map assignment syntax (`navigator_config = {`). Change to block syntax (`navigator_config {`).

#### Error: "Unknown attribute 'execution-environment'"

You're using the old hyphenated field names. Change to underscored names (`execution_environment`).

### Need Help?

- Check the updated examples in the `example/` directory
- See [README.md](README.md) for complete configuration reference
- Review the [proposal document](openspec/changes/replace-navigator-config-with-typed-structs/proposal.md) for technical details
