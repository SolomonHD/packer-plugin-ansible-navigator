# Proposal: Fix Collections Mount and Path in Execution Environment

## Why

When using execution environments (EE), Ansible collections installed to `~/.packer.d/ansible_collections_cache` are not accessible inside the container because:
1. The host collections directory is not mounted as a volume in the EE container
2. The environment variable uses the deprecated plural form `ANSIBLE_COLLECTIONS_PATHS` instead of singular `ANSIBLE_COLLECTIONS_PATH`
3. The collections path is not configured to point to the mounted location inside the container

This causes collection role execution to fail with errors like "unable to find role" even though the collections were successfully installed on the host.

## What Changes

- Mount the host collections cache directory as a read-only volume in the EE container
- Use `ANSIBLE_COLLECTIONS_PATH` (singular) instead of `ANSIBLE_COLLECTIONS_PATHS` (plural) to eliminate deprecation warnings
- Set `ANSIBLE_COLLECTIONS_PATH` inside the EE container to point to the mounted collections
- Handle both EE-enabled and EE-disabled scenarios correctly
- Ensure no regressions when execution environment is not used

## Impact

**Affected Specs:**
- `remote-provisioner-capabilities` - Collections path configuration and EE volume mounts
- `local-provisioner-capabilities` - Collections path configuration and EE volume mounts (on-target execution)

**Affected Code:**
- `provisioner/ansible-navigator/galaxy.go` - Environment variable name change
- `provisioner/ansible-navigator/navigator_config.go` - EE volume mount configuration
- `provisioner/ansible-navigator-local/galaxy.go` - Environment variable name change
- `provisioner/ansible-navigator-local/navigator_config.go` - EE volume mount configuration
- Tests for both provisioners

**User-Facing:**
- **Non-breaking**: Existing configurations continue to work
- Fixes broken functionality when using collections with execution environments
- Eliminates deprecation warning from Ansible
