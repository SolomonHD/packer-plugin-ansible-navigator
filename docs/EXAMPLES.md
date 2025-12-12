# Examples

> **⚠️ BREAKING CHANGE (v4.0.0):** The following options have been REMOVED: `execution_environment`, `ansible_cfg`, `ansible_env_vars`, `ansible_ssh_extra_args`, `extra_arguments`, `navigator_mode`, `roles_path`, `collections_path`, `galaxy_command`. All examples below use the current supported configuration. See [Migration Guide](MIGRATION.md) for upgrade instructions.

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

## 4) Advanced navigator_config with custom environment variables

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

---

[← Back to docs index](README.md)
