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

For the remote provisioner, no proxy fields are needed in this minimal EE
example. The plugin selects Podman before Docker when `container_engine` is
omitted, binds the temporary SSH adapter to `0.0.0.0` for bridge networking,
and advertises `host.containers.internal` for Podman or
`gateway.docker.internal` for Docker. On Linux and WSL2, Docker also receives
the required `gateway.docker.internal:host-gateway` mapping automatically.

To use host networking explicitly (where the runtime and platform support it):

```hcl
navigator_config {
  mode = "stdout"

  execution_environment {
    enabled           = true
    image             = "quay.io/ansible/creator-ee:latest"
    container_engine  = "docker"
    container_options = ["--network=host"]
  }
}
```

With host networking, the container shares host loopback, so both proxy
endpoints safely default to `127.0.0.1`. Docker Desktop host networking depends
on the Docker Desktop version and settings; bridged networking is the portable
default.

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

---

[← Back to docs index](README.md)
