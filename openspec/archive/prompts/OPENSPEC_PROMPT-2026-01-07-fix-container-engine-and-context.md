# Refactor to Pure YAML Configuration and Fix Context Support

## Context
Users are experiencing critical compatibility issues with `ansible-navigator` v25.12.0+:
1. **CLI Flag Incompatibility**: The plugin generates `--execution-environment-container-engine`, which is rejected by newer versions. CLI flags in `ansible-navigator` are less stable than the YAML configuration schema.
2. **Docker Context Mismatch**: Users running Rootless Docker or non-default Docker Contexts fail to find local images because Packer does not inherit the implicit `DOCKER_HOST`.

## Goal
Refactor the plugin to use a **Pure YAML Configuration** strategy (generating a complete `ansible-navigator.yml` from HCL) instead of the current hybrid CLI/YAML approach. This ensures better compatibility across `ansible-navigator` versions. Additionally, implement automatic Docker Context resolution.

## Scope
### In Scope
- **Refactor Configuration Generation**:
  - Modify `provisioner.go` to **always** generate a temporary `ansible-navigator.yml` using `generateNavigatorConfigYAML` when `navigator_config` is present.
  - Remove the call to `buildNavigatorCLIFlags`.
  - Ensure `ANSIBLE_NAVIGATOR_CONFIG` environment variable is set to the generated file path.
  - Remove `buildNavigatorCLIFlags` function and associated CLI mapping logic from `navigator_config.go`.
  - Ensure `generateNavigatorConfigYAML` includes ALL settings (it currently does, but verify).

- **Implement Context Resolution**:
  - Implement a `resolveDockerHost()` function in `provisioner.go` (or helper).
  - If `DOCKER_HOST` is unset, run `docker context inspect --format '{{.Endpoints.docker.Host}}'` to find the active socket.
  - Set `DOCKER_HOST` in the `ansible-navigator` process environment if found.

### Out of Scope
- Modifying `ansible-navigator` installation logic.

## Desired Behavior
1. **Configuration**: The plugin converts the entire HCL `navigator_config` block into a valid `ansible-navigator.yml` file. It executes `ansible-navigator run` with **no configuration flags** (relying entirely on `ANSIBLE_NAVIGATOR_CONFIG`).
2. **Context**: The plugin automatically detects and uses the correct Docker socket for Rootless/Context users.

## Benefits
- **Stability**: YAML schema is the primary, stable interface for `ansible-navigator`.
- **Simplicity**: Removes the complex and fragile HCL-to-CLI mapping logic.
- **Compatibility**: Works with v2.x and v25.x without version-specific flag logic.

## Acceptance Criteria
- [ ] `provisioner.go` no longer generates CLI flags for configuration.
- [ ] `ansible-navigator` is invoked with `ANSIBLE_NAVIGATOR_CONFIG` pointing to a full config file.
- [ ] `container_engine` setting in HCL works correctly (passed via YAML).
- [ ] Docker Context resolution works for Rootless Docker.
- [ ] Existing tests pass (may need updates for removed CLI logic).
