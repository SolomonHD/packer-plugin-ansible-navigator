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

## 3) Minimal execution environment + requirements.yml + collection role target

`requirements.yml`:

```yaml
collections:
  - name: devsec.hardening
```

Packer HCL:

```hcl
provisioner "ansible-navigator" {
  requirements_file = "./requirements.yml"

  navigator_config {
    mode = "stdout"

    execution_environment {
      enabled = true
      image   = "quay.io/ansible/creator-ee:latest"
    }
  }

  # Use a collection role FQDN target (not a playbook path)
  play {
    target = "devsec.hardening.os_hardening"
    become = true
  }
}
```

## 4) Modern configuration with navigator_config (recommended for ansible-navigator v3+)

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

## 5) Advanced navigator_config with custom environment variables

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

## 6) Expanded ansible.cfg sections via `navigator_config.ansible_config`

The plugin can generate an `ansible.cfg` containing additional sections like `[privilege_escalation]`, `[persistent_connection]`, and `[inventory]`.

See full working example: [`example/ansible-config-sections.pkr.hcl`](packer/plugins/packer-plugin-ansible-navigator/example/ansible-config-sections.pkr.hcl:1)

```hcl
provisioner "ansible-navigator" {
  navigator_config {
    mode = "stdout"

    ansible_config {
      privilege_escalation {
        become        = true
        become_method = "sudo"
        become_user   = "root"
      }

      persistent_connection {
        connect_timeout       = 30
        connect_retry_timeout = 15
        command_timeout       = 60
      }

      inventory {
        enable_plugins = ["ini", "yaml"]
      }
    }
  }

  play { target = "site.yml" }
}
```

## 7) Connection modes: SSH tunnel through bastion

Use `connection_mode = "ssh_tunnel"` when targets are only accessible through a bastion/jump host:

```hcl
provisioner "ansible-navigator" {
  # Use SSH tunnel mode instead of default proxy adapter
  connection_mode = "ssh_tunnel"
  
  # Bastion configuration
  bastion_host             = "bastion.example.com"
  bastion_user             = "ec2-user"
  bastion_private_key_file = "~/.ssh/bastion.pem"
  
  navigator_config {
    mode = "stdout"
    
    execution_environment {
      enabled     = true
      image       = "quay.io/ansible/creator-ee:latest"
      pull_policy = "missing"
    }
  }
  
  play { target = "site.yml" }
}
```

## 8) Connection modes: Direct connection (no proxy)

Use `connection_mode = "direct"` when the target is directly accessible and you want to bypass the proxy adapter:

```hcl
provisioner "ansible-navigator" {
  # Direct connection - Ansible handles SSH natively
  connection_mode = "direct"
  
  play { target = "site.yml" }
}
```

## 9) Connection modes: Proxy adapter (default)

The default `connection_mode = "proxy"` uses Packer's SSH proxy adapter. This is the default and doesn't need to be explicitly set:

```hcl
provisioner "ansible-navigator" {
  # connection_mode = "proxy"  ← Default, can be omitted
  
  play { target = "site.yml" }
}
```

---

[← Back to docs index](README.md)
