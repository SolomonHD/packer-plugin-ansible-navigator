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

If `ansible-navigator --version` is slow or hangs on your system:

- Prefer adding the directory to `ansible_navigator_path` (or setting `command` to an explicit path)
- Increase `version_check_timeout`
- Or set `skip_version_check = true` (not recommended unless you control the environment)

```hcl
provisioner "ansible-navigator" {
  ansible_navigator_path  = ["~/.asdf/shims"]
  version_check_timeout   = "120s"
  # skip_version_check = true

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
