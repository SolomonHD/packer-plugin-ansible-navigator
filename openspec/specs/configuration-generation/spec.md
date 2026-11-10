# configuration-generation Specification

## Purpose
TBD - created by archiving change fix-container-engine-and-context. Update Purpose after archive.
## Requirements
### Requirement: Pure YAML Configuration

The provisioner MUST generate a complete `ansible-navigator.yml` file for all configuration settings and use it as the sole configuration source, eliminating CLI configuration flags.

#### Scenario: Generate Configuration
Given: A valid `navigator_config` block in HCL
When: The provisioner prepares to execute `ansible-navigator`
Then: It MUST generate a complete `ansible-navigator.yml` file containing all settings
And: It MUST NOT generate any CLI configuration flags (e.g., `--execution-environment-image`)
And: It MUST set the `ANSIBLE_NAVIGATOR_CONFIG` environment variable to the generated file path

#### Scenario: Cleanup Configuration
Given: A temporary `ansible-navigator.yml` file was generated
When: The provisioner finishes execution (success or failure)
Then: It MUST remove the temporary file

### Requirement: Remove CLI Flag Logic

The provisioner MUST NOT contain logic for generating CLI configuration flags, as this is replaced by pure YAML configuration.

#### Scenario: CLI Flag Generation
Given: The `navigator_config.go` file
When: The code is compiled
Then: It MUST NOT contain `buildNavigatorCLIFlags` function
And: It MUST NOT contain `hasUnmappedSettings` function
And: It MUST NOT contain `generateMinimalYAML` function

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
- **AND** `navigator_config.execution_environment.container_options` already contains `--add-host=gateway.docker.internal:host-gateway`
- **WHEN** Navigator configuration is generated
- **THEN** the option SHALL NOT be duplicated
