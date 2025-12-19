# OpenSpec Change Prompt

## Context

The packer-plugin-ansible-navigator wraps `ansible-galaxy` for installing roles and collections via a `requirements_file` configuration option. The current documentation covers basic usage but lacks guidance for enterprise scenarios: private repositories, Private Automation Hub, git credential management, and custom Galaxy server authentication.

Users in enterprise environments commonly need to install collections from:
- Private GitHub/GitLab repositories (SSH or HTTPS with tokens)
- Red Hat Automation Hub or self-hosted Private Automation Hub
- Custom Galaxy servers with API key authentication

The plugin delegates credential handling entirely to `ansible-galaxy` and external git configuration, but this is not documented.

## Goal

Create comprehensive documentation explaining how to install Ansible collections from authenticated/private sources when using this plugin. The documentation should clarify what the plugin handles vs. what must be configured externally.

## Scope

**In scope:**
- New documentation file `docs/GALAXY_AUTHENTICATION.md`
- Update to `docs/README.md` to add the new doc to the index
- Update to `docs/TROUBLESHOOTING.md` to add authentication failure scenarios
- Examples for each authentication scenario
- Cross-references to ansible-galaxy and ansible.cfg documentation

**Out of scope:**
- Code changes to the plugin itself
- New configuration options
- Changes to the GalaxyManager implementation
- Support for additional authentication methods beyond what ansible-galaxy supports

## Desired Behavior

- Users can find documentation on how to authenticate when installing from private sources
- Documentation explains the boundary: plugin config vs. external config (ansible.cfg, git credentials)
- Each authentication scenario has a working example
- Troubleshooting includes common auth failure messages and resolutions

## Constraints & Assumptions

- Assumption: The plugin will continue to delegate to `ansible-galaxy` without managing credentials directly
- Assumption: Users may use this plugin in CI/CD environments where credentials must be injected at runtime
- Constraint: Must not recommend insecure practices (e.g., embedding tokens in requirements.yml in version control)
- Constraint: Documentation must work for both provisioner types (remote and local)

## Acceptance Criteria

- [ ] New file `docs/GALAXY_AUTHENTICATION.md` exists
- [ ] Doc covers: Public Galaxy (default), Private GitHub (SSH), Private GitHub (HTTPS), Private Automation Hub, Custom Galaxy servers
- [ ] Doc explains `galaxy_args` usage for passing `--server`, `--api-key`, `--ignore-certs`
- [ ] Doc explains ansible.cfg `[galaxy_server.*]` configuration
- [ ] Doc explains git credential prerequisites (SSH keys, credential helpers, environment variables)
- [ ] Doc includes CI/CD guidance (secrets injection, ssh-agent setup)
- [ ] `docs/README.md` updated with link to new doc
- [ ] `docs/TROUBLESHOOTING.md` includes section on authentication failures (401, Permission denied)
- [ ] All examples are copy-pasteable and accurate
