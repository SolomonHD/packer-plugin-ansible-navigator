## ADDED Requirements

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
