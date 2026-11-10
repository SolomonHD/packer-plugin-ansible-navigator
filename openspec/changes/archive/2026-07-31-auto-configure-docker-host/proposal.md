## Why

When running Ansible inside a container (Execution Environment), it often needs to communicate with services on the host machine (e.g., the Docker daemon itself, or other services). The standard way to do this in Docker Desktop is using the DNS name `gateway.docker.internal`. However, on standard Linux Docker installations, this DNS name is not resolved by default unless the container is run with `--add-host=gateway.docker.internal:host-gateway`.

Currently, users must manually add this flag to `navigator_config.execution_environment.container_options` if they want to use `gateway.docker.internal` as their `ansible_proxy_host`. This is error-prone and creates friction for users moving between Docker Desktop (Mac/Windows) and Linux environments.

## What Changes

- Detect `gateway.docker.internal` when it is configured as `ansible_proxy_host`.
- Inject `--add-host=gateway.docker.internal:host-gateway` into execution-environment container options.
- Preserve explicitly configured options and avoid duplicating the host-gateway flag.

## Capabilities

### New Capabilities

- `configuration-generation`: Automatically configure Linux container host-gateway resolution when the Ansible proxy uses `gateway.docker.internal`.

### Modified Capabilities

- None.

## Impact

- `provisioner/ansible-navigator/navigator_config.go`: Applies the automatic container option.
- `provisioner/ansible-navigator/provisioner.go`: Supplies the configured proxy host during navigator configuration generation.
- Existing configurations remain compatible; unrelated proxy hosts are unaffected.
