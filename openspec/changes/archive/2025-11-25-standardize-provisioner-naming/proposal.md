# Change: Standardize Provisioner Naming Convention

## Why

The current codebase has inconsistent naming between directory structures, import paths, and plugin registrations:

- Directory `provisioner/ansible-navigator/` contains **remote** execution code
- Directory `provisioner/ansible-navigator-local/` contains **local** execution code  
- Plugin registration uses "ansible-navigator" for local and "ansible-navigator-remote" for remote
- This creates confusion where the default name "ansible-navigator" maps to local execution in registration but the directory with that name contains remote execution code

This misalignment makes the codebase difficult to navigate and violates the principle of least surprise.

## What Changes

- **Rename** `provisioner/ansible-navigator-local/` → `provisioner/ansible-navigator/`
- **Rename** `provisioner/ansible-navigator/` → `provisioner/ansible-navigator-remote/`
- **Update** all import paths to reflect new directory structure
- **Update** package declarations in Go files
- **Update** all references in tests, documentation, and configuration files
- Maintain backward compatibility in plugin registration names (no breaking changes)

**Post-change state:**
- `ansible-navigator` (no suffix) = local execution (default) - directory and registration aligned
- `ansible-navigator-remote` = remote execution - directory and registration aligned
- Eliminates all references to `ansible-navigator-local`

## Impact

**Affected specs:**
- plugin-registration
- local-provisioning  
- remote-provisioning

**Affected code:**
- `main.go` - import paths
- `provisioner/ansible-navigator/` (entire directory renamed to `ansible-navigator-remote/`)
- `provisioner/ansible-navigator-local/` (entire directory renamed to `ansible-navigator/`)
- All test files in both directories
- Documentation in `docs/` referencing provisioner names

**Breaking changes:** None - plugin registration names remain the same, only internal structure changes