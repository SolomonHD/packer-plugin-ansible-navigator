# Change: Rename Package Names to Reflect Ansible Navigator

## Why
The current package names (`ansiblelocal`, `ansible`) are vestiges from the original HashiCorp `packer-plugin-ansible` codebase. Since this is now `packer-plugin-ansible-navigator`, the package names should consistently reflect the "ansible-navigator" branding to avoid confusion and improve code clarity.

## What Changes
- **provisioner/ansible-navigator/**: Rename package from `ansiblelocal` to `ansiblenavigatorlocal`
- **provisioner/ansible-navigator-remote/**: Rename package from `ansible` to `ansiblenavigatorremote`
- **main.go**: Update import aliases to match new package names
- **All test files**: Update any package declarations and imports

## Impact
- Affected specs: `plugin-registration`
- Affected code:
  - `provisioner/ansible-navigator/*.go`
  - `provisioner/ansible-navigator-remote/*.go`
  - `main.go`

## Notes
- Version loading from `version/VERSION` file is already correctly implemented using `//go:embed`
- No functional changes - this is purely a naming/consistency refactor
- Backward compatible - no external API changes