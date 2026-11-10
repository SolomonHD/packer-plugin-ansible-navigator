## Why

The remote provisioner currently defaults both ends of its SSH proxy route to `127.0.0.1`. When Ansible Navigator runs inside an execution-environment container, that inventory address refers to the container itself rather than the Packer host, so the documented default configuration fails with connection refused on WSL2 and standard Linux hosts.

## What Changes

- Resolve proxy bind and advertised inventory addresses from the execution topology instead of applying unconditional loopback defaults.
- Keep loopback defaults for host-side execution, but choose a container-reachable host gateway and a non-loopback listener when an execution environment uses bridged networking.
- Respect explicit proxy addresses and explicit host-network container options instead of overwriting user intent.
- Make Docker and Podman gateway behavior deterministic, including the Docker host-gateway mapping already supported by generated Navigator configuration.
- Emit actionable diagnostics showing the listener endpoint, advertised inventory endpoint, selected container engine/topology, and unsafe or contradictory combinations.
- Add automated coverage that launches a real execution-environment container and proves it can reach the plugin SSH proxy on Linux-compatible Docker runtimes.
- Update examples and troubleshooting guidance for WSL2, native Linux bridge networking, host networking, and explicit overrides.

No HCL fields are removed or renamed. The effective defaults for `ansible_proxy_bind_address` and `ansible_proxy_host` change only when the remote provisioner uses an enabled execution environment; users who require the former loopback behavior can set both fields explicitly or select host networking.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `remote-provisioner-capabilities`: Make proxy endpoint defaults execution-topology-aware and require useful connectivity diagnostics and validation.
- `configuration-generation`: Generate engine-appropriate host-gateway configuration without duplicating or overriding explicit container options.
- `build-tooling`: Add an opt-in, real-container acceptance test for the execution-environment-to-proxy route.
- `documentation`: Document working WSL2 and native Linux execution-environment configurations and troubleshooting evidence.

## Impact

- Affected code paths include remote provisioner preparation/defaulting, SSH adapter setup, inventory generation, Navigator YAML generation, debug logging, and acceptance-test setup under `provisioner/ansible-navigator`.
- Generated `ansible-navigator.yml` may gain an engine-specific host mapping for bridged Docker execution environments.
- The remote provisioner's execution-environment defaults become network-reachable by construction; local provisioner behavior and remote runs without an execution environment remain unchanged.
- No new runtime dependency is required. Real-container acceptance tests require an available Docker-compatible runtime and execution-environment image and must skip cleanly when those prerequisites are absent.
