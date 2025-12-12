# Migration Guide

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
