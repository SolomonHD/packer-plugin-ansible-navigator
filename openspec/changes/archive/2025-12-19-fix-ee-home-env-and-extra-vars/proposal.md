# Change: Fix execution-environment HOME/XDG defaults and extra-vars construction

## Why

When `navigator_config.execution_environment.enabled = true`, execution environment (EE) containers may run with the host UID/GID but lack a matching `/etc/passwd` entry and/or a writable home directory. This can result in `HOME=/` and Ansible attempting to create `/.ansible` (permission denied).

Separately, the provisioner currently passes multiple `-e` / `--extra-vars` arguments to `ansible-navigator run` (including Packer-derived variables). In some circumstances this has resulted in malformed Ansible invocation (e.g., a standalone `-e`, argument shifting where an extra-vars value becomes the playbook path), and playbook artifacts showing `ansible.cmdline` starting with `-e -e ...`.

## What Changes

- Ensure EE runs have safe, writable home-related environment defaults when the user did not explicitly provide them.
  - Default `HOME` to a writable path (e.g. `/tmp`) **only when** the user has not set or passed through `HOME`.
  - Default `XDG_CACHE_HOME` and `XDG_CONFIG_HOME` under `/tmp` **only when** the user has not set or passed through them.
  - Do not override any user-provided environment variable values.

- Change extra-vars construction so the provisioner cannot produce malformed `-e` usage.
  - Encode provisioner-generated extra vars as a single JSON object passed via one `-e`/`--extra-vars` argument.
  - Ensure the play target (playbook path) cannot be displaced by extra-vars.

- Add/adjust unit tests to prevent regressions for both behaviors.

- Add a minimal documentation example showing:
  - `navigator_config { execution_environment { enabled = true ... } }`
  - collection installation via `requirements.yml`
  - execution of a collection role via `play { target = "<namespace>.<collection>.<role>" }`

## Out of Scope / Non-Goals

- Changing container engine configuration (e.g., Docker), host permissions, or requiring users to rebuild EE images.
- Adding new provisioner schema fields (behavior should be achievable via safe defaults and existing `environment_variables` override mechanisms).
- Unrelated refactors.

## Impact

- Affected specs:
  - `openspec/specs/remote-provisioner-capabilities/spec.md`
  - `openspec/specs/local-provisioner-capabilities/spec.md`
  - `openspec/specs/documentation/spec.md`

- Affected implementation areas (for follow-on implementation task):
  - navigator_config EE default injection
  - command argument construction for extra-vars
  - unit tests for both provisioners
