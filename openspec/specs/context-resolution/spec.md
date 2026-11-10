# context-resolution Specification

## Purpose
TBD - created by archiving change fix-container-engine-and-context. Update Purpose after archive.
## Requirements
### Requirement: Docker Context Resolution

The provisioner MUST automatically detect and use the active Docker Context when `DOCKER_HOST` is not explicitly set, using the effective Navigator environment and bounded lookup, ensuring compatibility with Rootless Docker and non-default contexts.

#### Scenario: Detect Docker Host

- **GIVEN** `DOCKER_HOST` is not set and `docker` is available
- **WHEN** the provisioner prepares to execute Navigator
- **THEN** it MUST execute `docker context inspect --format '{{.Endpoints.docker.Host}}'` with the effective Navigator environment
- **AND** if the command succeeds and returns a non-empty value, it MUST set `DOCKER_HOST` in the Navigator process environment
- **AND** the inherited explicit `DOCKER_CONTEXT`, Docker configuration, and TLS-related settings MUST retain their meaning

#### Scenario: Existing Docker Host

- **GIVEN** `DOCKER_HOST` is explicitly set in the effective Navigator environment
- **WHEN** the provisioner prepares to execute Navigator
- **THEN** it MUST NOT replace that value with a discovered endpoint
- **AND** it MUST preserve explicit context selection without changing Docker's own selection precedence

#### Scenario: Docker Not Available

- **GIVEN** Docker is not available on the effective search path
- **WHEN** the provisioner attempts automatic Docker host discovery
- **THEN** discovery MUST proceed without setting `DOCKER_HOST`
- **AND** this optional-discovery fallback MUST NOT be reported as proof of Docker availability

#### Scenario: Context lookup fails or times out

- **GIVEN** automatic context discovery fails or does not complete within its documented timeout
- **WHEN** the provisioner prepares the Navigator environment
- **THEN** it MUST stop the discovery subprocess within that bound
- **AND** it MUST retain inherited explicit settings and issue a sanitized warning before proceeding without a synthesized endpoint
- **AND** it MUST NOT silently substitute a hardcoded socket or change Docker configuration

### Requirement: Docker Access Acceptance Stages

The explicitly invoked provisioning acceptance harness MUST distinguish executable availability, Docker daemon access, and real Navigator execution-environment startup using the effective environment of the tested invocation.

#### Scenario: Selected local endpoint is accessible

- **GIVEN** a supported local Docker endpoint selected by inherited settings or active context
- **WHEN** the acceptance harness checks Docker access
- **THEN** it MUST obtain a bounded successful daemon response using the tested runtime environment
- **AND** it MUST subsequently prove that the real Navigator executable launches the specified execution environment
- **AND** a successful Docker CLI probe alone MUST NOT count as successful Navigator startup

#### Scenario: Daemon or socket is inaccessible

- **GIVEN** the Docker client exists but the selected endpoint is missing, denied, unreachable, or times out
- **WHEN** explicit acceptance runs
- **THEN** it MUST fail at the Docker access stage with a corrective diagnostic
- **AND** it MUST NOT run sudo, change socket permissions, start a daemon, switch contexts, or retry a different endpoint automatically

#### Scenario: Image or execution-environment startup fails

- **GIVEN** daemon access succeeds but Navigator cannot start the selected image
- **WHEN** explicit acceptance runs
- **THEN** it MUST fail at the image/startup stage and preserve the subprocess result
- **AND** it MUST distinguish that failure from executable lookup and SSH/proxy failure

#### Scenario: Diagnostics preserve privacy

- **WHEN** acceptance records Docker selection and failure evidence
- **THEN** it MUST identify the selection source and a sanitized endpoint description
- **AND** it MUST NOT dump full environments, Docker inspect objects, registry credentials, or generated secret-bearing files
- **AND** ordinary debug preflight MUST remain a separate fast diagnostic rather than a blocking daemon health check
