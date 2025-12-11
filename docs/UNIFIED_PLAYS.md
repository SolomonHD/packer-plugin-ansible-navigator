# Play Blocks and Ordered Execution

This plugin executes provisioning using one or more ordered `play { ... }` blocks.

## Targets

Each `play` has a required `target`:

- Playbook path: `site.yml`, `./playbooks/deploy.yaml`
- Role FQDN: `namespace.collection.role` (or a simple role name)

When the target is a role FQDN, the provisioner generates a temporary playbook that runs the role.

## Requirements file

Use `requirements_file` to install roles + collections before any plays run.

```hcl
provisioner "ansible-navigator" {
  requirements_file = "./requirements.yml"

  play { target = "site.yml" }
  play { target = "geerlingguy.docker" }
}
```

---

[← Back to docs index](README.md)
