# documentation Specification

## Purpose
TBD - created by archiving change detect-and-handle-version-manager-shims. Update Purpose after archive.
## Requirements
### Requirement: Version Check Troubleshooting Documentation

The TROUBLESHOOTING.md documentation SHALL explain version check issues including automatic version manager shim detection and resolution, with clear guidance on when manual configuration is needed.

#### Scenario: Version manager shim documentation

- **WHEN** a user consults `docs/TROUBLESHOOTING.md` for version check hangs or timeout issues
- **THEN** the "Version check hangs / timeouts" section SHALL explain:
  - The plugin automatically detects and resolves shims for asdf, rbenv, and pyenv
  - Shim detection happens transparently during version check
  - Most shim configurations work automatically without manual intervention
  - Manual configuration is only needed when automatic resolution fails
- **AND** it SHALL include an example showing automatic shim resolution works by default

#### Scenario: Manual configuration still documented

- **WHEN** automatic shim resolution fails
- **THEN** the documentation SHALL provide manual configuration options:
  - Using `command` with full path to real binary
  - Using `ansible_navigator_path` to add directories containing the real binary
  - Using `skip_version_check = true` as last resort
- **AND** it SHALL explain these are fallback options when automatic detection fails

#### Scenario: Example configuration reflects shim support

- **WHEN** documentation shows example configurations for version managers
- **THEN** examples SHALL indicate shims now work automatically:

  ```hcl
  # Shims work automatically - no special configuration needed!
  # The plugin detects and resolves asdf/rbenv/pyenv shims automatically.
  
  provisioner "ansible-navigator" {
    play { target = "site.yml" }
  }
  
  # Only needed if automatic resolution fails:
  # provisioner "ansible-navigator" {
  #   ansible_navigator_path = ["~/.asdf/installs/ansible/X.Y.Z/bin"]
  #   play { target = "site.yml" }
  # }
  ```

- **AND** manual configuration SHALL be shown as optional/fallback

#### Scenario: No outdated workaround advice

- **WHEN** reviewing TROUBLESHOOTING.md
- **THEN** it SHALL NOT present manual shim configuration as the primary or required solution
- **AND** it SHALL NOT imply users must manually configure paths for version managers to work
- **AND** historical workaround advice SHALL be updated to reflect automatic detection

### Requirement: Document linkage between navigator logging and plugin debug output

User-facing documentation SHALL state that `navigator_config.logging.level` controls both ansible-navigator logging and the plugin's debug output (prefixed with `[DEBUG]`), and SHALL NOT introduce a separate plugin-specific debug/log-level configuration field.

#### Scenario: Linkage is documented prominently

- **GIVEN** a user consults the configuration documentation for `navigator_config.logging.level`
- **WHEN** the documentation describes what `navigator_config.logging.level` controls
- **THEN** it SHALL explicitly state that `navigator_config.logging.level` controls both:
  - ansible-navigator logging behavior
  - the plugin's debug output (prefixed with `[DEBUG]`)

#### Scenario: Documentation does not introduce a separate plugin log-level option

- **GIVEN** a user searches documentation for how to enable plugin debug output
- **WHEN** the documentation provides guidance
- **THEN** it SHALL NOT introduce a new plugin-specific `debug` or `log_level` option
- **AND** it SHALL direct users to set `navigator_config.logging.level = "debug"` instead

### Requirement: Documentation Index and Navigation

The [`docs/README.md`](packer/plugins/packer-plugin-ansible-navigator/docs/README.md:1) documentation index SHALL link to the Galaxy authentication guide and SHALL help users discover authentication guidance when needed.

#### Scenario: Galaxy authentication guide linked from index

- **WHEN** a user views [`docs/README.md`](packer/plugins/packer-plugin-ansible-navigator/docs/README.md:1)
- **THEN** it SHALL include a link to [`docs/GALAXY_AUTHENTICATION.md`](packer/plugins/packer-plugin-ansible-navigator/docs/GALAXY_AUTHENTICATION.md:1) in the "Detailed Guides" section
- **AND** the link description SHALL indicate it covers authentication for private collections, Private Automation Hub, and custom Galaxy servers

#### Scenario: Existing navigation preserved

- **WHEN** the documentation index is updated
- **THEN** all existing links SHALL remain functional
- **AND** the new link SHALL be logically ordered with other detailed guides

### Requirement: Authentication Failure Troubleshooting

The [`docs/TROUBLESHOOTING.md`](packer/plugins/packer-plugin-ansible-navigator/docs/TROUBLESHOOTING.md:1) guide SHALL include a dedicated section on authentication failures with common error messages and resolutions.

#### Scenario: Authentication failures section exists

- **WHEN** a user consults [`docs/TROUBLESHOOTING.md`](packer/plugins/packer-plugin-ansible-navigator/docs/TROUBLESHOOTING.md:1) for authentication issues
- **THEN** there SHALL be a section titled "Authentication failures" or similar
- **AND** the section SHALL be logically placed after the "Dependencies not installed" section

#### Scenario: HTTP 401 Unauthorized documented

- **WHEN** a user encounters "HTTP Error 401: Unauthorized" from ansible-galaxy
- **THEN** [`docs/TROUBLESHOOTING.md`](packer/plugins/packer-plugin-ansible-navigator/docs/TROUBLESHOOTING.md:1) SHALL document:
  - This error occurs when accessing Private Automation Hub or custom Galaxy servers without valid credentials
  - Resolution steps: verify API token, check [`ansible.cfg`](packer/plugins/packer-plugin-ansible-navigator/openspec/specs/remote-provisioner-capabilities/spec.md:100) server configuration
  - Cross-reference to [`docs/GALAXY_AUTHENTICATION.md`](packer/plugins/packer-plugin-ansible-navigator/docs/GALAXY_AUTHENTICATION.md:1) Private Automation Hub section

#### Scenario: Git SSH Permission denied documented

- **WHEN** a user encounters "Permission denied (publickey)" from git
- **THEN** [`docs/TROUBLESHOOTING.md`](packer/plugins/packer-plugin-ansible-navigator/docs/TROUBLESHOOTING.md:1) SHALL document:
  - This error occurs when SSH keys are not configured or ssh-agent is not running
  - Resolution steps: verify SSH key is added to GitHub/GitLab, start ssh-agent, add key to agent
  - Cross-reference to [`docs/GALAXY_AUTHENTICATION.md`](packer/plugins/packer-plugin-ansible-navigator/docs/GALAXY_AUTHENTICATION.md:1) Private GitHub SSH section

#### Scenario: Git HTTPS Authentication failed documented

- **WHEN** a user encounters "fatal: Authentication failed" from git
- **THEN** [`docs/TROUBLESHOOTING.md`](packer/plugins/packer-plugin-ansible-navigator/docs/TROUBLESHOOTING.md:1) SHALL document:
  - This error occurs when git credential helper is not configured or token is invalid
  - Resolution steps: verify token, check git credential helper setup, check environment variables
  - Cross-reference to [`docs/GALAXY_AUTHENTICATION.md`](packer/plugins/packer-plugin-ansible-navigator/docs/GALAXY_AUTHENTICATION.md:1) Private GitHub HTTPS section

#### Scenario: Generic installation failure documented

- **WHEN** a user encounters "Could not find/install packages" errors
- **THEN** [`docs/TROUBLESHOOTING.md`](packer/plugins/packer-plugin-ansible-navigator/docs/TROUBLESHOOTING.md:1) SHALL document:
  - This may be caused by authentication issues if collections are from private sources
  - Resolution: check if collections require authentication, consult [`docs/GALAXY_AUTHENTICATION.md`](packer/plugins/packer-plugin-ansible-navigator/docs/GALAXY_AUTHENTICATION.md:1)
  - Alternative causes: network issues, wrong collection name, offline mode enabled

### Requirement: Galaxy Authentication Documentation

The documentation SHALL provide comprehensive guidance on authenticating to private Ansible Galaxy sources, including git repositories, Private Automation Hub, and custom Galaxy servers, with clear explanation of the architectural boundary between plugin configuration and external authentication setup.

#### Scenario: Public Galaxy documented as default

- **WHEN** a user consults [`docs/GALAXY_AUTHENTICATION.md`](packer/plugins/packer-plugin-ansible-navigator/docs/GALAXY_AUTHENTICATION.md:1)
- **THEN** the documentation SHALL explain that public Galaxy (galaxy.ansible.com) works by default with no additional configuration
- **AND** it SHALL include an example using [`requirements_file`](packer/plugins/packer-plugin-ansible-navigator/openspec/specs/remote-provisioner-capabilities/spec.md:64) with public collections

#### Scenario: Private GitHub SSH authentication documented

- **WHEN** a user needs to install collections from private GitHub repositories using SSH
- **THEN** [`docs/GALAXY_AUTHENTICATION.md`](packer/plugins/packer-plugin-ansible-navigator/docs/GALAXY_AUTHENTICATION.md:1) SHALL document:
  - SSH key prerequisites (key generation, adding to GitHub)
  - SSH agent configuration for local development
  - SSH agent setup for CI/CD environments
  - Example requirements.yml with SSH URLs (git+ssh://)
  - Troubleshooting for "Permission denied (publickey)" errors

#### Scenario: Private GitHub HTTPS token authentication documented

- **WHEN** a user needs to install collections from private GitHub repositories using HTTPS with tokens
- **THEN** [`docs/GALAXY_AUTHENTICATION.md`](packer/plugins/packer-plugin-ansible-navigator/docs/GALAXY_AUTHENTICATION.md:1) SHALL document:
  - Git credential helper setup (store, cache, osxkeychain)
  - Environment variable patterns (`GIT_ASKPASS`, `GIT_USERNAME`, `GIT_PASSWORD`)
  - Example requirements.yml with HTTPS URLs
  - Secure token handling (emphasize NOT embedding in version control)
  - CI/CD secrets injection patterns (GitHub Actions, GitLab CI, Jenkins)

#### Scenario: Private Automation Hub documented

- **WHEN** a user needs to install collections from Private Automation Hub or Red Hat Automation Hub
- **THEN** [`docs/GALAXY_AUTHENTICATION.md`](packer/plugins/packer-plugin-ansible-navigator/docs/GALAXY_AUTHENTICATION.md:1) SHALL document:
  - Configuration via [`ansible.cfg`](packer/plugins/packer-plugin-ansible-navigator/openspec/specs/remote-provisioner-capabilities/spec.md:100) `[galaxy_server.*]` sections
  - Example using HCL [`ansible_cfg`](packer/plugins/packer-plugin-ansible-navigator/openspec/specs/remote-provisioner-capabilities/spec.md:100) block to define server configuration
  - Token/API key configuration patterns
  - Handling self-signed certificates (`verify_ssl = false` or adding CA certificates)

#### Scenario: Custom Galaxy server documented

- **WHEN** a user needs to install collections from a custom Galaxy server
- **THEN** [`docs/GALAXY_AUTHENTICATION.md`](packer/plugins/packer-plugin-ansible-navigator/docs/GALAXY_AUTHENTICATION.md:1) SHALL document:
  - Using [`galaxy_args`](packer/plugins/packer-plugin-ansible-navigator/openspec/specs/remote-provisioner-capabilities/spec.md:77) to pass `--server`, `--api-key`, `--ignore-certs`
  - Example HCL configuration with [`galaxy_args`](packer/plugins/packer-plugin-ansible-navigator/openspec/specs/remote-provisioner-capabilities/spec.md:77)
  - Combining with [`ansible.cfg`](packer/plugins/packer-plugin-ansible-navigator/openspec/specs/remote-provisioner-capabilities/spec.md:100) for multiple server configuration

#### Scenario: Configuration boundary clearly explained

- **WHEN** a user reads [`docs/GALAXY_AUTHENTICATION.md`](packer/plugins/packer-plugin-ansible-navigator/docs/GALAXY_AUTHENTICATION.md:1)
- **THEN** the documentation SHALL clearly explain:
  - What the plugin handles: invoking [`ansible-galaxy`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/galaxy.go:1) with [`requirements_file`](packer/plugins/packer-plugin-ansible-navigator/openspec/specs/remote-provisioner-capabilities/spec.md:64), passing [`galaxy_args`](packer/plugins/packer-plugin-ansible-navigator/openspec/specs/remote-provisioner-capabilities/spec.md:77), and managing install paths
  - What must be configured externally: git credentials, SSH keys, [`ansible.cfg`](packer/plugins/packer-plugin-ansible-navigator/openspec/specs/remote-provisioner-capabilities/spec.md:100) Galaxy server definitions
  - That the plugin delegates all authentication to [`ansible-galaxy`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/galaxy.go:1) and git

#### Scenario: CI/CD integration patterns documented

- **WHEN** a user consults [`docs/GALAXY_AUTHENTICATION.md`](packer/plugins/packer-plugin-ansible-navigator/docs/GALAXY_AUTHENTICATION.md:1) for CI/CD guidance
- **THEN** the documentation SHALL include examples for:
  - GitHub Actions (using secrets, SSH key setup)
  - GitLab CI (CI/CD variables, SSH_PRIVATE_KEY)
  - Jenkins (credentials binding)
- **AND** examples SHALL demonstrate ssh-agent setup and environment variable injection
- **AND** examples SHALL emphasize secure practices (secrets as CI variables, not in templates)

#### Scenario: Secure practices emphasized

- **WHEN** documentation shows authentication examples
- **THEN** it SHALL explicitly warn against:
  - Embedding tokens or passwords in Packer templates committed to version control
  - Hardcoding credentials in requirements.yml files
- **AND** it SHALL recommend secure alternatives:
  - Using CI/CD secret variables
  - Using git credential helpers
  - Using SSH keys with ssh-agent

### Requirement: REQ-DOC-EE-ROLE-001 Minimal example: EE + requirements.yml + collection role target

Documentation SHALL include a minimal, copy-pasteable Packer HCL example demonstrating:

- `navigator_config { execution_environment { enabled = true ... } }`
- collection installation via `requirements.yml`
- role execution via `play { target = "<namespace>.<collection>.<role>" }`

#### Scenario: Minimal example exists and uses role FQDN target

- **GIVEN** a user is reading the documentation examples
- **WHEN** they look for a minimal execution-environment example
- **THEN** the docs SHALL include an example that:
  - enables execution environment via `navigator_config.execution_environment.enabled = true`
  - installs at least one collection via `requirements_file = "requirements.yml"`
  - uses `play { target = "<namespace>.<collection>.<role>" }` (role FQDN), not only playbook-path targets

### Requirement: Execution-Environment Proxy Connectivity Documentation

The documentation SHALL explain the remote provisioner's execution-environment proxy topology and provide working guidance for WSL2 and native Linux.

#### Scenario: Minimal bridged execution-environment example

- **WHEN** a user follows the minimal remote execution-environment example
- **THEN** the example SHALL enable Navigator stdout mode and the execution environment
- **AND** it SHALL rely on the topology-aware proxy defaults without requiring manual proxy fields
- **AND** it SHALL identify the effective Docker or Podman gateway behavior

#### Scenario: Explicit host-network example

- **WHEN** a user chooses host networking
- **THEN** the documentation SHALL show the supported `container_options` syntax
- **AND** it SHALL explain why loopback proxy endpoints are correct for that topology
- **AND** it SHALL note platform and Docker Desktop support constraints

#### Scenario: Connection-refused troubleshooting

- **WHEN** a user encounters an execution-environment SSH connection refusal
- **THEN** troubleshooting guidance SHALL explain the distinction between host and container loopback
- **AND** it SHALL show how to inspect the logged listener endpoint, advertised inventory endpoint, engine, network mode, and generated host mapping
- **AND** it SHALL cover WSL2, native Linux, remote container-daemon limitations, and explicit proxy overrides

#### Scenario: Security implications are documented

- **WHEN** documentation describes bridge-network defaults
- **THEN** it SHALL state that the ephemeral SSH adapter binds to non-loopback host interfaces for the duration of the bridged run
- **AND** it SHALL explain the one-time-key and listener-lifetime mitigations

### Requirement: First Successful Provisioning Runbook

The project SHALL provide a repeatable runbook for proving Navigator executable discovery, Docker access, execution-environment startup, and real target provisioning on WSL2 and native Linux.

#### Scenario: Architecture and prerequisites are clear

- **WHEN** a user follows the runbook
- **THEN** it SHALL explain that Packer and Navigator run on the host, Navigator uses the host's Docker endpoint to launch the execution environment, and Ansible reaches the disposable target through the Packer SSH proxy
- **AND** it SHALL distinguish the Docker-builder target from the execution-environment container
- **AND** it SHALL list the required tools, versions, fixed images, isolated plugin installation, and explicitly invoked image preparation steps
- **AND** it SHALL NOT require mounting the host Docker socket into the execution environment

#### Scenario: Explicit-path success precedes automatic discovery

- **WHEN** the runbook establishes the first baseline
- **THEN** it SHALL show explicit executable selection, Docker selection, stdout mode, enabled execution environment, version checking enabled, and normal bridge proxy defaults
- **AND** it SHALL then show the same fixture using `ansible_navigator_path` and normal PATH/shim discovery
- **AND** it SHALL explain independent target marker verification and the intentional-failure control

#### Scenario: Failure is diagnosed by stage

- **WHEN** a required stage fails
- **THEN** the runbook SHALL distinguish missing/non-executable Navigator, shim resolution, Docker client discovery, socket/daemon access, image/startup failure, execution-environment file visibility, SSH/proxy reachability, and target task failure
- **AND** each stage SHALL include bounded diagnostic commands and expected success evidence
- **AND** it SHALL explain why a successful interactive-shell command or debug socket observation alone does not prove plugin execution
- **AND** it SHALL NOT recommend automatic sudo, socket permission changes, global PATH changes, or skipped version checks as the baseline fix

#### Scenario: Platform status reflects actual evidence

- **WHEN** the verification record is read
- **THEN** it SHALL list WSL2 with Docker Desktop and native Linux with Docker Engine separately
- **AND** it SHALL link sanitized reproduction evidence for each completed platform
- **AND** it SHALL mark absent, failed, or skipped required runs as pending rather than supported
- **AND** implementation progress SHALL be reported separately from merge, release, and archive status
