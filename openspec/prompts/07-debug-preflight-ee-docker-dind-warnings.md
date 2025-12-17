# Context

When users run the SSH-based provisioner in containerized CI environments (including Dagger), ansible-navigator Execution Environments (EE) frequently fail due to missing container engine wiring (docker client, docker socket, etc.). Users also confuse “Docker client in container talking to host daemon” (not DinD) with true “Docker-in-Docker” (dockerd running in the container).

We want **debug-only** diagnostics that help users confirm whether they are set up correctly, without adding noise for normal users.

This change depends on the repo’s decision to **link plugin debug output to** `navigator_config.logging.level` (see [`LoggingConfig.Level`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/provisioner.go:123)).

# Goal

Add DEBUG-only preflight diagnostics/warnings for EE container runtime setup and likely DinD conditions.

# Scope

## In scope

- When debug mode is enabled (per change 06), and when EE is enabled (`navigator_config.execution_environment.enabled = true`), emit diagnostic output that helps users troubleshoot:
  - docker client availability
  - docker socket availability / DOCKER_HOST
  - “likely DinD” heuristic (dockerd running in the same environment)
- Emit these diagnostics for both provisioners:
  - SSH-based: [`provisioner/ansible-navigator/`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/provisioner.go:9)
  - On-target: `provisioner/ansible-navigator-local/`

## Out of scope

- Hard-failing the build solely due to “likely DinD”. This MUST remain a warning.
- Running potentially slow/hanging docker commands such as `docker info`.
- Adding new user configuration fields.

# Desired Behavior

## Trigger conditions

The preflight checks run when all are true:

- Debug mode is enabled by `navigator_config.logging.level = "debug"` (case-insensitive)
- `navigator_config.execution_environment.enabled = true`

## Diagnostics emitted (DEBUG-only)

When triggered, the provisioner SHALL emit `[DEBUG]` messages that include:

1) **Docker endpoint visibility**

- `DOCKER_HOST` value (or that it is unset)
- whether `/var/run/docker.sock` exists (and that it is a socket)

2) **Docker client availability**

- whether `docker` is found in PATH

3) **Likely DinD heuristic (warning only)**

- If a `dockerd` process is detected in the current environment, emit a debug warning that this is likely DinD and advise using a host/remote daemon instead.

All warnings in this change MUST:

- be printed only in debug mode
- be clearly labeled as warnings (e.g. `[DEBUG][WARN] ...`)
- include an actionable “what to do next” sentence

## Performance and safety constraints

- Checks must be fast and non-blocking.
- Checks must not leak secrets.
- Checks must not change execution behavior; they only add diagnostic output.

# Constraints & Assumptions

- The plugin cannot reliably prove that a docker socket is host-mounted vs container-local.
- The plugin *can* reliably detect a strong DinD smell: `dockerd` running in the same PID namespace.
- The SSH-based provisioner runs ansible-navigator on the machine where Packer runs (so these checks belong near the local exec path, e.g. before executing `ansible-navigator run`).

# Acceptance Criteria

- [ ] When debug mode is enabled and EE is enabled, the provisioner emits the diagnostic messages described above.
- [ ] When debug mode is disabled, none of these new diagnostics appear.
- [ ] The “likely DinD” message is warning-only and does not fail the build.
- [ ] Unit tests exist demonstrating:
  - debug-only gating
  - behavior when docker client is missing
  - behavior when dockerd is detected

## Expected files/areas touched

- [`provisioner/ansible-navigator/provisioner.go`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/provisioner.go:1173) (preflight location should be near the exec path)
- `provisioner/ansible-navigator-local/provisioner.go`
- Tests under:
  - `provisioner/ansible-navigator/*_test.go`
  - `provisioner/ansible-navigator-local/*_test.go`

## Dependencies

- Depends on prompt 06: debug gating linked to `navigator_config.logging.level`.
