## ADDED Requirements

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
