# Change: Auto-configure Docker Host Mapping

## Problem
When running Ansible inside a container (Execution Environment), it often needs to communicate with services on the host machine (e.g., the Docker daemon itself, or other services). The standard way to do this in Docker Desktop is using the DNS name `gateway.docker.internal`. However, on standard Linux Docker installations, this DNS name is not resolved by default unless the container is run with `--add-host=gateway.docker.internal:host-gateway`.

Currently, users must manually add this flag to `navigator_config.execution_environment.container_options` if they want to use `gateway.docker.internal` as their `ansible_proxy_host`. This is error-prone and creates friction for users moving between Docker Desktop (Mac/Windows) and Linux environments.

## Proposed Solution
Modify the plugin to automatically detect when `ansible_proxy_host` is set to `gateway.docker.internal` and inject the necessary `--add-host` flag into the execution environment configuration.

## Implementation Details

### 1. Update `applyAutomaticEEDefaults`
- Update function signature to accept `ansibleProxyHost` string.
- Add logic to check if `ansibleProxyHost` equals `gateway.docker.internal`.
- If true, check if `--add-host=gateway.docker.internal:host-gateway` is already present in `config.ExecutionEnvironment.ContainerOptions`.
- If not present, append it.

### 2. Update `GenerateNavigatorConfigYAML`
- Update function signature to accept `ansibleProxyHost` string.
- Pass this value to `applyAutomaticEEDefaults`.

### 3. Update Call Sites
- Update `provisioner.go` to pass `p.config.AnsibleProxyHost` when calling `GenerateNavigatorConfigYAML`.

## Impact Analysis
- **Backward Compatibility:** Fully compatible. Existing configurations that manually specify the flag will not be duplicated. Configurations not using `gateway.docker.internal` will be unaffected.
- **Security:** No new security risks. The `host-gateway` mapping is a standard Docker feature.
- **Performance:** Negligible impact.
