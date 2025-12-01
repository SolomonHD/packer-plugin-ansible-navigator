# Change: Fix Provisioner Registration Names

## Why

The plugin currently registers provisioners with full names (`ansible-navigator`, `ansible-navigator-remote`) in main.go. This violates Packer SDK naming conventions and causes awkward HCL usage where users get duplicated names like `ansible-navigator-ansible-navigator` when referencing provisioners.

The Packer SDK provides `plugin.DEFAULT_NAME` (equals `"-packer-default-plugin-name-"`) specifically to handle the primary provisioner, and expects short names for secondary provisioners that Packer will prefix with the plugin alias.

## What Changes

- Update `main.go` to use `plugin.DEFAULT_NAME` for the local/primary provisioner
- Update `main.go` to use short name `"remote"` for the remote provisioner
- Update documentation to show correct HCL usage patterns
- **BREAKING**: The remote provisioner changes from `ansible-navigator-remote` to `ansible-navigator-remote` (no actual user-facing change for remote, but registration name changes internally)

## Impact

- Affected specs: `plugin-registration`
- Affected code: `main.go`, `README.md`
- Affected tools: `describe` output format changes
- Version bump: Minor version (API-compatible but changes user-facing naming)

## Desired Behavior After Change

| Registration | HCL Usage |
|--------------|-----------|
| `plugin.DEFAULT_NAME` | `provisioner "ansible-navigator" { }` |
| `"remote"` | `provisioner "ansible-navigator-remote" { }` |

Running `./packer-plugin-ansible-navigator describe` should output provisioners as:
```json
{
  "provisioners": ["-packer-default-plugin-name-", "remote"]
}
```

## Reference

This follows the pattern used by `hashicorp/packer-plugin-ansible` which registers:
- `-packer-default-plugin-name-` → accessible as `ansible` in HCL
- `local` → accessible as `ansible-local` in HCL