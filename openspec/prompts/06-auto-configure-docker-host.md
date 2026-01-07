# OpenSpec Change Prompt

## Context
The `ansible-navigator` plugin runs Ansible inside a container. To allow the containerized Ansible to communicate with the host (e.g., for Docker operations), `gateway.docker.internal` is often used as the proxy host. However, this DNS name is not automatically available in all environments (specifically standard Linux Docker) without explicit configuration.

## Goal
Implement automatic detection and configuration of Docker host mapping to ensure `gateway.docker.internal` resolves correctly in both Linux and Docker Desktop environments.

## Scope

**In scope:**
- Modify `GenerateNavigatorConfigYAML` or `applyAutomaticEEDefaults` in the plugin code.
- Logic to detect if `ansible_proxy_host` is set to `gateway.docker.internal`.
- Logic to automatically append `--add-host=gateway.docker.internal:host-gateway` to `execution-environment.container-options`.
- Logic to prevent duplication if the flag is already present.

**Out of scope:**
- Changes to other provisioners or builders.
- Modifying the default value of `ansible_proxy_host` (only reacting to it).

## Desired Behavior
- When `ansible_proxy_host` is set to `gateway.docker.internal`:
  - The plugin automatically adds `--add-host=gateway.docker.internal:host-gateway` to the container options in the generated `ansible-navigator.yml`.
- If the user has already manually specified this `--add-host` flag in their configuration, the plugin does **not** add a duplicate.
- The generated configuration is valid for `ansible-navigator`.

## Constraints & Assumptions
- Assumption: The container engine (Docker/Podman) supports the `--add-host` flag and the `host-gateway` special value.
- Constraint: Must not interfere with other user-defined container options.

## Acceptance Criteria
- [ ] `GenerateNavigatorConfigYAML` or equivalent function includes the detection logic.
- [ ] If `ansible_proxy_host` is `gateway.docker.internal`, the generated YAML contains `--add-host=gateway.docker.internal:host-gateway`.
- [ ] If the user manually adds `--add-host=gateway.docker.internal:host-gateway`, it appears only once in the generated YAML.
- [ ] If `ansible_proxy_host` is NOT `gateway.docker.internal`, the flag is NOT automatically added.
