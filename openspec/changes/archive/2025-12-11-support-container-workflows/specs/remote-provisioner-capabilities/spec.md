# remote-provisioner-capabilities Specification

## ADDED Requirements

### Requirement: Configurable SSH Proxy Bind Address

The SSH-based provisioner SHALL support configuring the bind address for the SSH proxy to allow connections from external containers (e.g., WSL2).

#### Scenario: Default bind address

- **GIVEN** a configuration without `ansible_proxy_bind_address` specified
- **WHEN** the provisioner sets up the SSH proxy
- **THEN** it SHALL bind to `127.0.0.1`
- **AND** this SHALL ensure security by restricting access to localhost

#### Scenario: Custom bind address

- **GIVEN** a configuration with `ansible_proxy_bind_address = "0.0.0.0"`
- **WHEN** the provisioner sets up the SSH proxy
- **THEN** it SHALL bind to `0.0.0.0`
- **AND** this SHALL allow connections from other interfaces (e.g., container bridge)

### Requirement: Configurable Inventory Host Address

The SSH-based provisioner SHALL support configuring the host address used in the generated inventory file to ensure the execution environment can reach the host.

#### Scenario: Default inventory host

- **GIVEN** a configuration without `ansible_proxy_host` specified
- **WHEN** the provisioner generates the inventory file
- **THEN** it SHALL use `127.0.0.1` as the `ansible_host`
- **AND** this SHALL work for local execution environments

#### Scenario: Custom inventory host

- **GIVEN** a configuration with `ansible_proxy_host = "host.containers.internal"`
- **WHEN** the provisioner generates the inventory file
- **THEN** it SHALL use `host.containers.internal` as the `ansible_host`
- **AND** this SHALL allow the containerized execution environment to connect back to the host

### Requirement: Unbuffered Python Output

The SSH-based provisioner SHALL force unbuffered Python output to ensure logs are streamed immediately to Packer, preventing apparent hangs during connection timeouts.

#### Scenario: PYTHONUNBUFFERED injection

- **WHEN** the provisioner executes `ansible-navigator` (for plays or playbooks)
- **THEN** it SHALL inject `PYTHONUNBUFFERED=1` into the environment variables
- **AND** this SHALL ensure that Python output is flushed immediately
