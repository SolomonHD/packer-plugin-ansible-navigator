## ADDED Requirements

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

