## MODIFIED Requirements

### Requirement: Configurable SSH Proxy Bind Address

The SSH-based provisioner SHALL support configuring the bind address for the SSH proxy and SHALL choose a safe, reachable default from the resolved execution topology when it is omitted.

#### Scenario: Default bind address without a bridged execution environment

- **GIVEN** a configuration without `ansible_proxy_bind_address` specified
- **AND** Ansible runs on the host or the execution environment explicitly uses host networking
- **WHEN** the provisioner sets up the SSH proxy
- **THEN** it SHALL bind to `127.0.0.1`
- **AND** it SHALL restrict the listener to host loopback

#### Scenario: Default bind address for a bridged execution environment

- **GIVEN** a remote provisioner with an enabled execution environment using bridge networking
- **AND** `ansible_proxy_bind_address` is not specified
- **WHEN** the provisioner resolves the proxy topology
- **THEN** it SHALL bind the SSH proxy to `0.0.0.0`
- **AND** the execution-environment container SHALL be able to reach the allocated listener port through the selected host gateway

#### Scenario: Custom bind address

- **GIVEN** a configuration with an explicit `ansible_proxy_bind_address`
- **WHEN** the provisioner resolves the proxy topology
- **THEN** it SHALL preserve the explicit bind address
- **AND** it SHALL validate that address against the selected network topology

### Requirement: Configurable Inventory Host Address

The SSH-based provisioner SHALL support configuring the host address used in generated inventory and SHALL choose a container-engine-appropriate default when it is omitted.

#### Scenario: Default inventory host without a bridged execution environment

- **GIVEN** a configuration without `ansible_proxy_host` specified
- **AND** Ansible runs on the host or the execution environment explicitly uses host networking
- **WHEN** the provisioner generates inventory
- **THEN** it SHALL use `127.0.0.1` as `ansible_host`

#### Scenario: Default inventory host for Docker bridge networking

- **GIVEN** a remote provisioner with an enabled Docker execution environment using bridge networking
- **AND** `ansible_proxy_host` is not specified
- **WHEN** the provisioner resolves the proxy topology
- **THEN** it SHALL use `gateway.docker.internal` as `ansible_host`

#### Scenario: Default inventory host for Podman bridge networking

- **GIVEN** a remote provisioner with an enabled Podman execution environment using bridge networking
- **AND** `ansible_proxy_host` is not specified
- **WHEN** the provisioner resolves the proxy topology
- **THEN** it SHALL use `host.containers.internal` as `ansible_host`

#### Scenario: Custom inventory host

- **GIVEN** a configuration with an explicit `ansible_proxy_host`
- **WHEN** the provisioner resolves the proxy topology
- **THEN** it SHALL preserve the explicit host
- **AND** it SHALL use that value as `ansible_host` in generated inventory

## ADDED Requirements

### Requirement: Deterministic Execution-Environment Proxy Topology

The remote provisioner MUST resolve the container engine, network mode, listener address, and advertised inventory host as one coherent proxy topology before provisioning begins.

#### Scenario: Automatic container engine selection

- **GIVEN** an enabled execution environment with `container_engine` omitted or set to `auto`
- **WHEN** the provisioner prepares its proxy topology
- **THEN** it MUST select an available supported engine using the same preference order as Ansible Navigator
- **AND** it MUST write the selected engine into generated Navigator configuration
- **AND** it MUST return an actionable preparation error when no supported engine is available

#### Scenario: Existing host-network option is recognized

- **GIVEN** container options contain `--network=host`, `--network host`, `--net=host`, or `--net host`
- **WHEN** the provisioner resolves the proxy topology
- **THEN** it MUST classify the execution environment as host-networked
- **AND** it MUST retain loopback defaults for both proxy endpoints when neither is explicit

#### Scenario: Explicit values take precedence

- **GIVEN** either proxy address is explicitly configured
- **WHEN** the provisioner resolves the omitted companion value
- **THEN** it MUST preserve every explicit value
- **AND** it MUST infer only omitted values from the selected topology

### Requirement: Proxy Topology Validation and Diagnostics

The remote provisioner MUST reject known-unreachable proxy topologies and SHALL report the effective route needed to diagnose container connectivity.

#### Scenario: Bridged execution environment advertises loopback

- **GIVEN** an enabled bridge-network execution environment
- **AND** the resolved inventory host is a loopback address
- **WHEN** configuration is prepared
- **THEN** preparation MUST fail with an error explaining that container loopback cannot reach the host proxy
- **AND** the error MUST identify the proxy fields or host-network option that can correct the configuration

#### Scenario: Bridged execution environment uses a loopback-only listener

- **GIVEN** an enabled bridge-network execution environment
- **AND** the resolved bind address is loopback while the inventory host is a host gateway
- **WHEN** configuration is prepared
- **THEN** preparation MUST fail with an error explaining that the listener is not reachable through the container bridge

#### Scenario: Effective route is logged

- **GIVEN** the SSH adapter has allocated its listener port
- **WHEN** the provisioner starts Ansible Navigator
- **THEN** output SHALL identify the actual listener endpoint and advertised inventory endpoint
- **AND** debug output SHALL identify the resolved container engine, network mode, inferred or explicit address origins, and automatic host mapping

#### Scenario: Navigator process fails

- **GIVEN** Ansible Navigator exits unsuccessfully
- **WHEN** the provisioner reports the failure
- **THEN** it MUST report the actual subprocess exit code
- **AND** it MUST retain the effective proxy-route diagnostics in the build log

