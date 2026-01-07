# Spec: Automatic Docker Host Configuration

## Background
To support `gateway.docker.internal` on Linux Docker, the container must be started with `--add-host=gateway.docker.internal:host-gateway`.

## Requirements

### Requirement: Automatic Host Gateway Injection
The plugin MUST automatically configure the container runtime to resolve `gateway.docker.internal` when that hostname is used for the Ansible proxy.

#### Scenario: Proxy host is gateway.docker.internal
- **GIVEN** `ansible_proxy_host` is set to `"gateway.docker.internal"`
- **AND** `execution_environment.enabled` is `true`
- **WHEN** the navigator configuration is generated
- **THEN** `execution-environment.container-options` SHALL include `--add-host=gateway.docker.internal:host-gateway`

#### Scenario: Proxy host is NOT gateway.docker.internal
- **GIVEN** `ansible_proxy_host` is set to `"127.0.0.1"` (or any other value)
- **WHEN** the navigator configuration is generated
- **THEN** `execution-environment.container-options` SHALL NOT be modified to include the add-host flag

#### Scenario: User already specified the flag
- **GIVEN** `ansible_proxy_host` is set to `"gateway.docker.internal"`
- **AND** `navigator_config.execution_environment.container_options` already contains `--add-host=gateway.docker.internal:host-gateway`
- **WHEN** the navigator configuration is generated
- **THEN** the flag SHALL NOT be duplicated in `container-options`
