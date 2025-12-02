# Change: Swap Provisioner Naming to Align with Packer Conventions

## Why

The plugin's current provisioner naming is inverted from standard Packer conventions. The official HashiCorp `ansible` plugin uses:
- `ansible` (default) – runs **from** the local machine, SSHs **to** the target
- `ansible-local` – runs **on** the target machine with local connection

This plugin currently has it backwards:
- `ansible-navigator` (DEFAULT) – runs **on** target (should be "-local")
- `ansible-navigator-remote` – runs **from** local machine (should be default)

This causes confusion for users familiar with the official Ansible plugin and violates the principle of least surprise.

## What Changes

### **BREAKING** – Major Version Bump (v3.0.0)

1. **Directory Renames**
   - `provisioner/ansible-navigator/` → `provisioner/ansible-navigator-local/`
   - `provisioner/ansible-navigator-remote/` → `provisioner/ansible-navigator/`

2. **Go Package Name Changes**
   - Current `ansible-navigator/` package (`ansiblenavigatorlocal`) → stays `ansiblenavigatorlocal` in new location
   - Current `ansible-navigator-remote/` package (`ansiblenavigatorremote`) → becomes `ansiblenavigator` (primary)

3. **Plugin Registration in main.go**
   - `plugin.DEFAULT_NAME` → remote provisioner (new `ansible-navigator`)
   - `"local"` → local provisioner (new `ansible-navigator-local`)

4. **HCL Usage After Change**
   - `provisioner "ansible-navigator"` – runs from local machine, connects to target via SSH
   - `provisioner "ansible-navigator-local"` – runs directly on the target machine

5. **Documentation Updates**
   - All docs updated to reflect new naming
   - Migration guide added for v2.x → v3.x users
   - CHANGELOG entry documenting breaking change

## Impact

- **Affected specs:**
  - `directory-structure` – directory names change
  - `plugin-registration` – registration names and semantics swap
  - `local-provisioner-capabilities` – rename to clarify scope (or create `remote-provisioner-capabilities`)

- **Affected code:**
  - `main.go` – import paths and registrations
  - All files in `provisioner/ansible-navigator/` – move and update package name
  - All files in `provisioner/ansible-navigator-remote/` – move and update package name
  - `GNUmakefile` – if any paths are referenced
  - `.goreleaser.yaml` – if any paths are referenced
  - `README.md`, `docs/` – documentation updates
  - `CHANGELOG.md` – breaking change documentation

- **User Migration Required:**
  - Users currently using `provisioner "ansible-navigator"` must switch to `provisioner "ansible-navigator-local"`
  - Users currently using `provisioner "ansible-navigator-remote"` can switch to `provisioner "ansible-navigator"` (or keep using the old name if aliased)