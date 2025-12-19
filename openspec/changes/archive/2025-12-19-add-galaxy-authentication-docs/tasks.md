# Implementation Tasks

## 1. Create Core Authentication Documentation

- [x] 1.1 Create [`docs/GALAXY_AUTHENTICATION.md`](packer/plugins/packer-plugin-ansible-navigator/docs/GALAXY_AUTHENTICATION.md:1) with structure:
  - Introduction explaining delegation to [`ansible-galaxy`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/galaxy.go:1)
  - Configuration boundary section (plugin vs. external)
  - Authentication scenarios (organized by source type)

- [x] 1.2 Document Public Galaxy (default) scenario:
  - Explain default behavior with no additional configuration
  - Example using [`requirements_file`](packer/plugins/packer-plugin-ansible-navigator/openspec/specs/remote-provisioner-capabilities/spec.md:64) with public collections

- [x] 1.3 Document Private GitHub (SSH) scenario:
  - SSH key prerequisites and setup
  - Example requirements.yml with SSH URLs
  - SSH agent configuration for CI/CD
  - Troubleshooting "Permission denied (publickey)" errors

- [x] 1.4 Document Private GitHub (HTTPS with tokens) scenario:
  - Git credential helper setup
  - Environment variable patterns (`GIT_ASKPASS`, credential helpers)
  - Example requirements.yml with HTTPS URLs
  - Secure token handling (avoid embedding in version control)
  - CI/CD secrets injection patterns

- [x] 1.5 Document Private Automation Hub scenario:
  - Explain [`ansible.cfg`](packer/plugins/packer-plugin-ansible-navigator/openspec/specs/remote-provisioner-capabilities/spec.md:100) `[galaxy_server.*]` configuration
  - Example using HCL [`ansible_cfg`](packer/plugins/packer-plugin-ansible-navigator/openspec/specs/remote-provisioner-capabilities/spec.md:100) block
  - Token/API key configuration
  - Self-signed certificates handling

- [x] 1.6 Document Custom Galaxy Servers scenario:
  - Using [`galaxy_args`](packer/plugins/packer-plugin-ansible-navigator/openspec/specs/remote-provisioner-capabilities/spec.md:77) to pass `--server`, `--api-key`, `--ignore-certs`
  - Example HCL configuration with [`galaxy_args`](packer/plugins/packer-plugin-ansible-navigator/openspec/specs/remote-provisioner-capabilities/spec.md:77)
  - Multiple server configuration patterns

- [x] 1.7 Document CI/CD integration patterns:
  - GitHub Actions secret injection examples
  - GitLab CI variables and SSH key setup
  - Jenkins credentials binding patterns
  - SSH agent setup for automated builds

## 2. Update Documentation Index

- [x] 2.1 Add link to [`GALAXY_AUTHENTICATION.md`](packer/plugins/packer-plugin-ansible-navigator/docs/GALAXY_AUTHENTICATION.md:1) in [`docs/README.md`](packer/plugins/packer-plugin-ansible-navigator/docs/README.md:1) under "Detailed Guides" section

- [x] 2.2 Add brief description explaining when users should consult the authentication guide

## 3. Enhance Troubleshooting Guide

- [x] 3.1 Add "Authentication Failures" section to [`docs/TROUBLESHOOTING.md`](packer/plugins/packer-plugin-ansible-navigator/docs/TROUBLESHOOTING.md:1)

- [x] 3.2 Document common authentication error patterns:
  - "HTTP Error 401: Unauthorized" (Private Automation Hub)
  - "Permission denied (publickey)" (Git SSH)
  - "fatal: Authentication failed" (Git HTTPS)
  - "Could not find/install packages" (offline/auth issues)

- [x] 3.3 For each error pattern, provide:
  - Symptom description
  - Root cause explanation
  - Resolution steps with cross-references to [`GALAXY_AUTHENTICATION.md`](packer/plugins/packer-plugin-ansible-navigator/docs/GALAXY_AUTHENTICATION.md:1)

## 4. Validation

- [x] 4.1 Ensure all examples are copy-pasteable and syntactically correct

- [x] 4.2 Verify HCL examples match current schema (use [`galaxy_args`](packer/plugins/packer-plugin-ansible-navigator/openspec/specs/remote-provisioner-capabilities/spec.md:77), [`requirements_file`](packer/plugins/packer-plugin-ansible-navigator/openspec/specs/remote-provisioner-capabilities/spec.md:64), etc.)

- [x] 4.3 Check that documentation emphasizes secure practices (no hardcoded tokens in templates)

- [x] 4.4 Verify cross-references between docs are accurate

- [x] 4.5 Confirm both provisioner types ([`ansible-navigator`](packer/plugins/packer-plugin-ansible-navigator/openspec/specs/remote-provisioner-capabilities/spec.md:1) and [`ansible-navigator-local`](packer/plugins/packer-plugin-ansible-navigator/openspec/specs/local-provisioner-capabilities/spec.md:1)) are covered where relevant
