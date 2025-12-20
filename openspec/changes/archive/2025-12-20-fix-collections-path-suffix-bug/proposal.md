# Proposal: Fix Collections Path Suffix Bug in Execution Environment Mounts

## Why

The plugin incorrectly appends `ansible_collections` to the configured `collections_path` when generating execution environment mounts (lines 1307-1312 in `provisioner/ansible-navigator/provisioner.go`). This breaks the directory layout inside the container because:

1. `ansible-galaxy` installs collections to `<collections_path>/ansible_collections/<namespace>/<collection>`
2. The plugin then tries to mount `<collections_path>/ansible_collections` inside the container
3. This results in a double `ansible_collections` path segment that breaks role FQDN discovery

Example failure case:
- Collections installed to: `~/.packer.d/ansible_collections_cache/ansible_collections/integration/common_tasks`
- Mount source becomes: `~/.packer.d/ansible_collections_cache/ansible_collections` (WRONG - adds extra suffix)
- Container sees: `<mount>/integration/common_tasks` instead of `<mount>/ansible_collections/integration/common_tasks`
- Ansible cannot find collection role: `integration.common_tasks.etc_profile_d`

Additionally, the plugin currently uses the deprecated plural form `ANSIBLE_COLLECTIONS_PATHS` which generates deprecation warnings from Ansible.

## What Changes

**Remove the buggy path suffix logic:**
- Delete the code at lines 1307-1312 in `provisioner/ansible-navigator/provisioner.go` that appends `ansible_collections` to `collections_path`
- Pass `collections_path` directly to `generateNavigatorConfigYAML()` without modification
- Apply the same fix to `provisioner/ansible-navigator-local/provisioner.go`

**Fix environment variable naming:**
- Update specs to use `ANSIBLE_COLLECTIONS_PATH` (singular) instead of `ANSIBLE_COLLECTIONS_PATHS` (plural)
- Note: The code in `galaxy.go` may already use the correct singular form; verify and document

**Update tests:**
- Modify tests that assert on the collections path to expect the unmodified path
- Add test cases that verify collections mounted under EE are discoverable via role FQDN

## Impact

**Affected Specs:**
- `remote-provisioner-capabilities` - MODIFIED requirement for collections_path environment variable
- `local-provisioner-capabilities` - MODIFIED requirement for collections_path environment variable

**Affected Code:**
- `provisioner/ansible-navigator/provisioner.go` - Remove suffix-appending logic (lines 1307-1312)
- `provisioner/ansible-navigator-local/provisioner.go` - Same fix for local provisioner
- `provisioner/ansible-navigator/galaxy.go` - Verify env var name (may already be correct)
- `provisioner/ansible-navigator-local/galaxy.go` - Verify env var name (may already be correct)
- Tests for both provisioners

**User-Facing:**
- **Bugfix**: Collections installed via `requirements_file` now work correctly with execution environments
- **Non-breaking**: The change fixes broken functionality; users don't need to modify their HCL configs
- **Deprecation fix**: Eliminates Ansible deprecation warning about `ANSIBLE_COLLECTIONS_PATHS`

## Risks

**Low risk:**
- The current code is clearly buggy (appending when the path already contains the suffix)
- The fix aligns `collections_path` semantics with how `ansible-galaxy` actually uses it
- No user configuration changes needed
