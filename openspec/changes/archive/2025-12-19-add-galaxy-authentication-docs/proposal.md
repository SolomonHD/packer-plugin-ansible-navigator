# Change: Document Galaxy Authentication Scenarios

## Why

The plugin wraps [`ansible-galaxy`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/galaxy.go:1) for installing roles and collections via [`requirements_file`](packer/plugins/packer-plugin-ansible-navigator/openspec/specs/remote-provisioner-capabilities/spec.md:64) configuration option. Current documentation covers basic usage but lacks guidance for enterprise scenarios requiring authentication: private repositories (GitHub/GitLab), Private Automation Hub, custom Galaxy servers, and git credential management.

Users in enterprise environments commonly encounter authentication challenges when installing collections from private sources. The plugin delegates credential handling entirely to [`ansible-galaxy`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/galaxy.go:1) and external git configuration, but this architectural boundary is not documented, leading to confusion about what should be configured in the plugin versus external tools.

## What Changes

- **NEW**: [`docs/GALAXY_AUTHENTICATION.md`](packer/plugins/packer-plugin-ansible-navigator/docs/GALAXY_AUTHENTICATION.md:1) - Comprehensive guide covering:
  - Public Galaxy (default behavior)
  - Private GitHub repositories (SSH and HTTPS with tokens)
  - Private Automation Hub / Red Hat Automation Hub
  - Custom Galaxy servers with API key authentication
  - Git credential prerequisites and setup
  - CI/CD integration patterns (secrets injection, ssh-agent setup)
  - Boundary clarification: plugin config vs. external config
  
- **MODIFIED**: [`docs/README.md`](packer/plugins/packer-plugin-ansible-navigator/docs/README.md:1) - Add link to new authentication guide

- **MODIFIED**: [`docs/TROUBLESHOOTING.md`](packer/plugins/packer-plugin-ansible-navigator/docs/TROUBLESHOOTING.md:1) - Add new section on authentication failures with common error messages (401 Unauthorized, Permission denied, etc.)

## Impact

- **Affected specs**: [`documentation`](packer/plugins/packer-plugin-ansible-navigator/openspec/specs/documentation/spec.md:1)
- **Affected code**: None - documentation-only change
- **User experience**: Users will have clear guidance on configuring authentication for private collection sources
- **Security**: Documentation will emphasize secure practices (avoiding embedded tokens in version control)
