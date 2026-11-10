# build-tooling Specification

## Purpose
TBD - created by archiving change update-go-version. Update Purpose after archive.
## Requirements
### Requirement: Go Version Compatibility
The project MUST build and run correctly with Go version 1.23.4.

#### Scenario: Successful build
- **WHEN** the project is built with Go 1.23.4
- **THEN** all tests pass and the plugin functions correctly

#### Scenario: Dependency compatibility
- **WHEN** go mod tidy is run with Go 1.23.4
- **THEN** all dependencies are resolved without conflicts

#### Scenario: Plugin verification
- **WHEN** make plugin-check is executed with Go 1.23.4
- **THEN** the plugin passes all Packer SDK compatibility checks

### Requirement: HCL2 Spec Regeneration

After ANY modification to a Config struct (adding, removing, or changing fields with `mapstructure` tags), the developer MUST regenerate HCL2 spec files by running `make generate`.

#### Scenario: Config field added

- **WHEN** a new field is added to a Config struct with a `mapstructure` tag
- **THEN** `make generate` is run before committing
- **AND** the new field appears in the corresponding `FlatConfig` struct in `*.hcl2spec.go`
- **AND** the new field appears in the `HCL2Spec()` return map

#### Scenario: Build verification after regeneration

- **WHEN** `make generate` completes successfully
- **THEN** `go build ./...` passes without errors
- **AND** `make plugin-check` passes Packer SDK validation

### Requirement: Execution-Environment Proxy Connectivity Acceptance Test

The project MUST provide an opt-in acceptance test that proves a real execution-environment container can connect to the remote provisioner's host-side proxy route.

#### Scenario: Docker bridge connectivity succeeds

- **GIVEN** acceptance tests are enabled and a supported local Docker runtime and test image are available
- **WHEN** the execution-environment proxy connectivity test runs with default bridge topology
- **THEN** it MUST start a listener on the resolved bind address
- **AND** it MUST launch a disposable container with the generated host-gateway options
- **AND** the container MUST connect to the advertised host and allocated listener port

#### Scenario: Host-network connectivity succeeds where supported

- **GIVEN** acceptance tests are enabled and the local runtime supports host networking
- **WHEN** the connectivity test runs with an explicit host-network option
- **THEN** the container MUST connect to the listener through `127.0.0.1`

#### Scenario: Runtime prerequisites are unavailable

- **GIVEN** the acceptance test runtime, daemon, required network feature, or configured test image is unavailable
- **WHEN** the acceptance test starts
- **THEN** it MUST skip with a specific reason rather than fail an ordinary unit-test run
- **AND** it MUST clean up any listener or container it created

### Requirement: Isolated Real Navigator Provisioning Acceptance

The project MUST provide an explicitly invoked acceptance command that builds the source plugin, runs a real Packer build through Navigator and a Docker execution environment, and independently verifies provisioning on a disposable target.

#### Scenario: Source plugin and real execution environment are used

- **GIVEN** the required tools and fixed test images are available
- **WHEN** acceptance is explicitly invoked
- **THEN** it MUST load the plugin built from the recorded source revision using an isolated plugin installation
- **AND** it MUST record the installed binary checksum and Packer's selected plugin path
- **AND** it MUST use `provisioner "ansible-navigator"`, an ordered `play` block, explicit Docker, stdout mode, an enabled execution environment, the SSH proxy, and normal bridge networking
- **AND** it MUST prove real Navigator execution-environment startup rather than substitute ansible-playbook, a mock executable, or host-local Ansible execution

#### Scenario: Marker proves target provisioning

- **GIVEN** a fresh disposable target and a unique run value
- **WHEN** the real playbook provisions the target through generated inventory and the SSH proxy
- **THEN** the marker MUST be absent before the play
- **AND** Ansible MUST receive the unique value through the plugin's extra-vars path and write it on the target
- **AND** an independent subsequent Packer shell provisioner MUST verify the marker's exact value on that target
- **AND** acceptance MUST require successful Packer completion and independent verification, rather than a play recap or TCP probe alone

#### Scenario: Required execution-environment files cross the boundary

- **WHEN** the acceptance play runs from its execution environment
- **THEN** host-side Navigator MUST successfully consume its generated settings
- **AND** Ansible inside the EE MUST successfully consume the generated inventory, temporary SSH key, extra-vars file, and playbook from their actual execution locations
- **AND** the test MUST NOT mask a missing mount by binding the Docker socket or the entire host temporary directory into the execution environment

#### Scenario: Intentional failure cannot pass acceptance

- **GIVEN** the acceptance fixture is configured to run an intentionally failing Ansible task
- **WHEN** Packer executes that task
- **THEN** the build MUST fail and the verifier MUST reject the result
- **AND** a separate negative-control check MUST confirm that expected failure propagation
- **AND** the failed build MUST NOT be recorded as successful platform provisioning

### Requirement: Staged Acceptance Results and Owned Cleanup

Explicit provisioning acceptance MUST report bounded stage outcomes and clean only its own disposable resources. Missing prerequisites or skipped required stages MUST prevent a passing result.

#### Scenario: Required prerequisites are missing

- **GIVEN** explicit acceptance was requested but a required binary, daemon, plugin, or image is unavailable
- **WHEN** the harness checks prerequisites
- **THEN** it MUST return a nonzero result identifying the blocked stage
- **AND** it MUST NOT convert the condition into a successful skip or fabricate downstream results
- **AND** image provisioning MUST require a separate explicit preparation action

#### Scenario: Ordinary unit tests run without Docker

- **WHEN** ordinary unit tests run without acceptance enabled
- **THEN** they MUST remain independent of Docker and external tools
- **AND** existing opt-in component-test skips MUST NOT count toward the end-to-end acceptance record

#### Scenario: Run fails or is interrupted

- **WHEN** an acceptance stage fails or is interrupted
- **THEN** bounded cleanup MUST remove only containers, images, temporary files, and plugin installation state created by that run
- **AND** it MUST retain sanitized failure evidence
- **AND** unknown or incomplete cleanup MUST be reported separately and MUST prevent a fully passing result
- **AND** it MUST NOT prune shared resources or alter existing Docker services, contexts, socket permissions, or user plugin installations

### Requirement: WSL2 and Native Linux Provisioning Evidence

Completion MUST include successful real provisioning on WSL2 with Docker Desktop integration and on an actual native Linux host with Docker Engine, using the same source revision and acceptance fixture.

#### Scenario: Each platform completes the baseline and discovery variants

- **WHEN** qualification is performed on either required platform
- **THEN** a build with an explicit real Navigator path MUST pass first
- **AND** a build with Navigator available only through `ansible_navigator_path` MUST subsequently pass
- **AND** normal PATH discovery MUST pass, with a live supported-shim variant where that platform uses a shim
- **AND** the variants MUST use the same fixed image inputs and independently verify their own unique target markers

#### Scenario: Reproducible evidence is retained

- **WHEN** a platform run is recorded
- **THEN** evidence MUST identify host platform and kernel, source SHA and dirty state, plugin checksum, tool and Docker-builder versions, Navigator executable path, immutable image references, Docker selection source, exact reproduction commands, stage outcomes, marker verification, negative-control outcome, and cleanup status
- **AND** the record MUST distinguish live results from mocked/unit coverage and untested variants

#### Scenario: Only one platform is available

- **GIVEN** the other required platform has not been exercised successfully
- **WHEN** completion is assessed
- **THEN** its acceptance task MUST remain unfinished
- **AND** a Linux container or VM used only as a target under WSL MUST NOT count as native-Linux-host execution
- **AND** any subsequent plugin code or fixture change MUST require rerunning both platform acceptance records against the final revision
