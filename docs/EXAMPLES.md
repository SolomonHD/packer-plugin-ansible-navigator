# Examples

> **⚠️ BREAKING CHANGE:** Legacy options have been REMOVED: `execution_environment`, `ansible_cfg`, `ansible_env_vars`, `ansible_ssh_extra_args`, `extra_arguments`, `navigator_mode`.
>
> **⚠️ BREAKING CHANGE (dependencies):** `roles_cache_dir`/`collections_cache_dir` were renamed to `roles_path`/`collections_path`, and `force_update`/`galaxy_force_install` were replaced by `galaxy_force`/`galaxy_force_with_deps`.
>
> All examples below use the current supported configuration. See [Migration Guide](MIGRATION.md) for upgrade instructions.

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

## 3) Modern configuration with navigator_config (recommended for ansible-navigator v3+)

```hcl
provisioner "ansible-navigator" {
  navigator_config {
    mode = "stdout"

    execution_environment {
      enabled     = true
      image       = "quay.io/ansible/creator-ee:latest"
      pull_policy = "missing"
    }

    ansible_config {
      defaults {
        remote_tmp = "/tmp/.ansible/tmp"
        local_tmp  = "/tmp/.ansible-local"
      }
    }
  }

  play { target = "site.yml" }
}
```

## 4) Advanced navigator_config with custom environment variables

```hcl
provisioner "ansible-navigator" {
  navigator_config {
    mode = "json"

    execution_environment {
      enabled     = true
      image       = "quay.io/ansible/creator-ee:latest"
      pull_policy = "always"

      environment_variables {
        set = {
          ANSIBLE_REMOTE_TMP = "/custom/tmp"
          ANSIBLE_LOCAL_TMP  = "/custom/local"
          CUSTOM_VAR         = "production"
        }
      }
    }

    # Any ansible_config.defaults/ssh_connection settings are written into a
    # generated ansible.cfg file and referenced from ansible-navigator.yml via
    # ansible.config.path.
    ansible_config {
      defaults {
        remote_tmp        = "/custom/tmp"
        local_tmp         = "/custom/local"
        host_key_checking = false
      }

      ssh_connection {
        pipelining  = true
        ssh_timeout = 30
      }
    }

    logging {
      level  = "debug"
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

---

[← Back to docs index](README.md)
