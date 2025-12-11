# Examples

> **⚠️ Deprecation Notice:** Several configuration options shown in older examples are deprecated. See [Migration Guide](MIGRATION.md) for migration paths to `navigator_config`.

## 1) Minimal playbook run

```hcl
provisioner "ansible-navigator" {
  play {
    target = "site.yml"
  }
}
```

## 2) Multiple ordered plays + requirements

```hcl
provisioner "ansible-navigator" {
  requirements_file = "./requirements.yml"

  play {
    name   = "Base"
    target = "site.yml"
  }

  play {
    name   = "Install Docker"
    target = "geerlingguy.docker"
    become = true
  }
}
```

## 3) Execution environment + ansible.cfg defaults

> **⚠️ Deprecated:** This example uses deprecated options. See example 5 for the recommended `navigator_config` approach.

```hcl
provisioner "ansible-navigator" {
  # DEPRECATED: Use navigator_config instead (see example 5)
  execution_environment = "quay.io/ansible/creator-ee:latest"
  ansible_cfg = {
    defaults = {
      remote_tmp = "/tmp/.ansible/tmp"
      local_tmp  = "/tmp/.ansible-local"
    }
  }

  play { target = "site.yml" }
}
```

## 4) JSON logging

> **⚠️ Deprecated:** `navigator_mode` is deprecated. Use `navigator_config.mode` instead (see example 6).

```hcl
provisioner "ansible-navigator" {
  # DEPRECATED: Use navigator_config.mode instead
  navigator_mode      = "json"
  structured_logging  = true
  log_output_path     = "./logs/ansible-summary.json"
  verbose_task_output = true

  play { target = "site.yml" }
}
```

## 5) Modern configuration with navigator_config (recommended for ansible-navigator v3+)

```hcl
provisioner "ansible-navigator" {
  navigator_config = {
    mode = "stdout"
    execution-environment = {
      enabled = true
      image = "quay.io/ansible/creator-ee:latest"
      pull-policy = "missing"
    }
    ansible = {
      config = {
        defaults = {
          remote_tmp = "/tmp/.ansible/tmp"
          local_tmp = "/tmp/.ansible-local"
        }
      }
    }
  }

  play { target = "site.yml" }
}
```

## 6) Advanced navigator_config with custom environment variables

```hcl
provisioner "ansible-navigator" {
  navigator_config = {
    mode = "json"
    execution-environment = {
      enabled = true
      image = "quay.io/ansible/creator-ee:latest"
      pull-policy = "always"
      environment-variables = {
        ANSIBLE_REMOTE_TMP = "/custom/tmp"
        ANSIBLE_LOCAL_TMP = "/custom/local"
        CUSTOM_VAR = "production"
      }
    }
    ansible = {
      config = {
        defaults = {
          host_key_checking = "False"
          gathering = "smart"
        }
        ssh_connection = {
          pipelining = "True"
          timeout = "30"
        }
      }
    }
    logging = {
      level = "debug"
      append = true
    }
  }

  structured_logging = true
  log_output_path = "./logs/deployment.json"

  play {
    name = "Deploy Application"
    target = "app.deploy"
    become = true
  }
}
```

## 7) Migration example: both legacy and navigator_config (navigator_config takes precedence)

```hcl
provisioner "ansible-navigator" {
  # Legacy options (will be overridden by navigator_config)
  execution_environment = "old-image:latest"
  navigator_mode = "json"
  
  # Modern configuration (takes precedence)
  navigator_config = {
    mode = "stdout"
    execution-environment = {
      enabled = true
      image = "quay.io/ansible/creator-ee:latest"
    }
  }
  
  play { target = "site.yml" }
}
```

In this example, the final execution will use:

- Mode: `stdout` (from navigator_config, not `json` from navigator_mode)
- Image: `quay.io/ansible/creator-ee:latest` (from navigator_config, not `old-image:latest`)

---

[← Back to docs index](README.md)
