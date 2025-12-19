# Change: DEBUG-only EE / Docker / DinD Preflight Diagnostics

## Why

Users running the provisioners in containerized CI environments (including Dagger) frequently hit failures or confusing behavior when using ansible-navigator Execution Environments (EE) without the required container-runtime wiring (docker client, docker socket/`DOCKER_HOST`, etc.).

The plugin already supports a DEBUG mode that is gated by `navigator_config.logging.level`. This change adds **debug-only** preflight diagnostics to help users quickly confirm whether their environment is plausibly set up correctly, without adding noise for normal users.

## What Changes

- When **plugin debug mode is enabled** and **execution environments are enabled**, emit additional preflight diagnostics via Packer UI.
- Diagnostics include:
  - `DOCKER_HOST` visibility (value or unset)
  - `/var/run/docker.sock` visibility (exists and is a socket)
  - docker client availability in PATH
  - a **warning-only** “likely DinD” heuristic when a `dockerd` process is detected
- No new user-facing configuration fields are introduced.
- No hard-fail is added solely due to “likely DinD”.

## Scope

### In scope

- Apply consistently to both provisioners:
  - `ansible-navigator` (SSH-based, runs ansible-navigator on the machine running Packer)
  - `ansible-navigator-local` (on-target, runs ansible-navigator on the target machine)

### Out of scope

- Running potentially slow/hanging docker commands (e.g., `docker info`).
- Adding new configuration knobs.
- Changing execution behavior beyond additional debug UI messages.

## Impact

- Affected specs:
  - `remote-provisioner-capabilities`
  - `local-provisioner-capabilities`
- Affected code areas (implementation task, not part of this proposal):
  - Remote provisioner execution path prior to `ansible-navigator run`
  - Local provisioner remote-shell execution path prior to `ansible-navigator run`

## Dependencies

- Relies on the existing rule: plugin debug mode is enabled if and only if `navigator_config.logging.level` equals `"debug"` (case-insensitive).
