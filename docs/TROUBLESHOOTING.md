# Troubleshooting

## Configuration errors

### Error: "at least one `play` block must be defined"

Add one or more `play { ... }` blocks:

```hcl
provisioner "ansible-navigator" {
  play {
    target = "site.yml"
  }
}
```

### Error: command must be only the executable name or path (no arguments)

`command` must be an executable only (no embedded args like `run`).

✅ Correct:

```hcl
command = "~/.asdf/shims/ansible-navigator"
```

❌ Wrong:

```hcl
command = "~/.asdf/shims/ansible-navigator run"
```

If you need to pass flags, use provisioner fields or the `navigator_config` block.

## Version check hangs / timeouts

### Automatic shim resolution (v0.4.0+)

The plugin automatically detects and resolves version manager shims (asdf, rbenv, pyenv). In most cases, no manual configuration is needed.

**When automatic resolution works:**

```bash
# Your version manager is configured
which ansible-navigator
# → ~/.asdf/shims/ansible-navigator (a shim)

asdf which ansible-navigator  # or rbenv/pyenv which
# → /home/user/.asdf/installs/ansible-navigator/2.3.0/bin/ansible-navigator
```

The plugin detects the shim and automatically uses the real binary path.

**When manual configuration is needed:**

If automatic resolution fails (version manager not in PATH, or `which` command fails), you'll see a clear error message with solutions:

```hcl
provisioner "ansible-navigator" {
  # Option 1: Specify the full path directly
  command = "/home/user/.asdf/installs/ansible-navigator/2.3.0/bin/ansible-navigator"
  # Find your path with: asdf which ansible-navigator

  play { target = "site.yml" }
}
```

**Or add the directory to PATH:**

```hcl
provisioner "ansible-navigator" {
  # Option 2: Add directories to be prepended to PATH
  ansible_navigator_path = ["/home/user/.asdf/installs/ansible-navigator/2.3.0/bin"]

  play { target = "site.yml" }
}
```

### Other timeout causes

If timeouts occur for reasons other than shims (network delays, container image pulls):

```hcl
provisioner "ansible-navigator" {
  version_check_timeout = "120s"  # Increase timeout
  # skip_version_check = true      # Last resort only

  play { target = "site.yml" }
}
```

## Execution environment permissions (`/.ansible/tmp`)

If Ansible in the execution environment fails due to non-root temp dirs, set `ansible_cfg` (or rely on the provisioner defaults when `execution_environment` is set):

```hcl
provisioner "ansible-navigator" {
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

## Dependencies not installed

If roles/collections are missing, ensure `requirements_file` exists and includes the needed items:

```hcl
provisioner "ansible-navigator" {
  requirements_file = "./requirements.yml"
  play { target = "site.yml" }
}
```

---

[← Back to docs index](README.md)
