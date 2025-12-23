# Change: Expand Execution Environment Configuration

## Why

Users attempting to configure valid ansible-navigator v3.x execution environment options like `container-options`, additional `volume-mounts`, or `pull-arguments` receive "Unsupported argument" errors because these fields are not modeled in the plugin's HCL2 spec. The current [`ExecutionEnvironment`](../../provisioner/ansible-navigator/provisioner.go:92) struct only covers basic fields (enabled, image, pull_policy) and basic environment variable support.

This limits users' ability to:
- Configure custom container network modes (e.g., `--net=host`)
- Apply container security options (e.g., `--security-opt`)
- Mount custom volumes beyond the automatic collections mount
- Pass credentials or other flags to image pull operations
- Select specific container engines (docker vs podman)

Without these capabilities, users working with advanced execution environment scenarios (WSL2, custom networks, air-gapped registries) cannot use the plugin effectively and must resort to workarounds or fall back to non-EE execution.

## What Changes

- Expand [`ExecutionEnvironment`](../../provisioner/ansible-navigator/provisioner.go:92) struct with three new fields:
  - `container_engine` (string) - select "auto", "podman", or "docker"
  - `container_options` ([]string) - additional container runtime flags
  - `pull_arguments` ([]string) - arguments passed to image pull command
- Update [`convertToYAMLStructure()`](../../provisioner/ansible-navigator/navigator_config.go:160) to emit new fields with correct YAML key naming:
  - `container-engine` (string)
  - `container-options` (list)
  - `pull.arguments` (list, nested alongside pull.policy under `pull` parent key)
- Regenerate HCL2 spec files via `make generate`
- Add unit tests for HCL decoding and YAML generation of new fields
- Preserve backward compatibility - new fields are optional, existing configs unchanged

## Impact

- **Affected specs**: `remote-provisioner-capabilities`, `local-provisioner-capabilities`
- **Affected code**:
  - [`provisioner/ansible-navigator/provisioner.go`](../../provisioner/ansible-navigator/provisioner.go:1) - struct definitions
  - [`provisioner/ansible-navigator/navigator_config.go`](../../provisioner/ansible-navigator/navigator_config.go:1) - YAML generation
  - [`provisioner/ansible-navigator-local/provisioner.go`](../../provisioner/ansible-navigator-local/provisioner.go:1) - struct definitions (shared)
  - [`provisioner/ansible-navigator-local/navigator_config.go`](../../provisioner/ansible-navigator-local/navigator_config.go:1) - YAML generation (shared)
  - Generated [`*.hcl2spec.go`](../../provisioner/ansible-navigator/*.hcl2spec.go:1) files
- **Schema changes**: Three new optional fields in `ExecutionEnvironment` struct
- **HCL2 spec regeneration required**: Yes, run `make generate`
- **Breaking changes**: None (purely additive, all new fields optional)
- **Migration required**: No
