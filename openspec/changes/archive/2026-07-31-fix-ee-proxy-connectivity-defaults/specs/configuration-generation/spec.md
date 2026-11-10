## MODIFIED Requirements

### Requirement: Automatic Host Gateway Injection

The plugin MUST generate container-engine-appropriate host-gateway configuration for the resolved proxy topology without duplicating or overriding explicit container options.

#### Scenario: Docker bridge topology uses gateway.docker.internal

- **GIVEN** the resolved topology uses Docker bridge networking
- **AND** `ansible_proxy_host` resolves to `gateway.docker.internal`
- **WHEN** Navigator configuration is generated
- **THEN** `execution-environment.container-options` SHALL include `--add-host=gateway.docker.internal:host-gateway`

#### Scenario: Podman bridge topology uses host.containers.internal

- **GIVEN** the resolved topology uses Podman bridge networking
- **AND** `ansible_proxy_host` resolves to `host.containers.internal`
- **WHEN** Navigator configuration is generated
- **THEN** it SHALL NOT add Docker's `host-gateway` option
- **AND** it SHALL preserve all user-supplied container options

#### Scenario: Host networking does not receive a host mapping

- **GIVEN** the execution environment explicitly uses host networking
- **WHEN** Navigator configuration is generated
- **THEN** it SHALL NOT add an automatic host-gateway mapping

#### Scenario: Custom proxy host does not receive an automatic mapping

- **GIVEN** `ansible_proxy_host` is explicitly set to a value other than the engine's automatic gateway alias
- **WHEN** Navigator configuration is generated
- **THEN** it SHALL NOT add an automatic host-gateway mapping

#### Scenario: User already specified the Docker gateway flag

- **GIVEN** the resolved topology requires `--add-host=gateway.docker.internal:host-gateway`
- **AND** `navigator_config.execution_environment.container_options` already contains that option
- **WHEN** Navigator configuration is generated
- **THEN** the option SHALL NOT be duplicated

