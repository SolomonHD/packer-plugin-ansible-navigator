# OpenSpec Change Prompt

## Context
The Packer `ansible-navigator` provisioner runs Ansible inside an execution environment (EE) container. Some EE images are run with the host UID/GID (e.g. `1000:1000`) but do not provide a matching `/etc/passwd` entry or writable home directory, causing `HOME=/` and Ansible to try to create `/.ansible` (permission denied). Separately, a recent run produced an `ansible-playbook: error: argument -e/--extra-vars: expected one argument`, and the `ansible-navigator` playbook artifact showed a malformed Ansible command line beginning with `-e -e ...`.

## Goal
Make the Packer `ansible-navigator` provisioner robust when running in EEs without a valid home directory, and ensure extra-vars are passed to Ansible without producing malformed `-e` arguments.

## Scope

**In scope:**
- When `navigator_config.execution_environment.enabled = true`, ensure the EE container has a writable home-related environment so Ansible does not attempt to write to `/.ansible`.
- Change how the provisioner passes Packer-derived variables (and any automatically added Ansible vars like `ansible_ssh_private_key_file`) so the invocation cannot produce a standalone `-e` and cannot shift positional arguments (e.g. playbook path).
- Add/adjust unit tests to prevent regressions for both behaviors.
- Add a minimal documentation example (simpler than the current `packer-navigator.pkr.hcl`) that:
  - uses `navigator_config { execution_environment { enabled = true ... } }`
  - installs a collection providing a role via `requirements.yml`
  - executes that collection role via `play { target = "<collection_namespace>.<collection_name>.<role_name>" }`

**Out of scope:**
- Changing container engine configuration (e.g. Docker), host permissions, or requiring users to rebuild their EE images.
- Adding new configuration fields to the provisioner schema (prefer safe defaults that can be overridden via existing `environment_variables.set`).
- Any unrelated refactors.

## Desired Behavior
- If `navigator_config.execution_environment.enabled` is true and `navigator_config.execution_environment.environment_variables.set.HOME` is not provided, set it to a universally writable path (e.g. `/tmp`) for the EE.
- Also default `XDG_CACHE_HOME` and `XDG_CONFIG_HOME` under `/tmp` when not explicitly set, to avoid cache/config writes to non-writable locations.
- User-provided env var values (including `HOME`, `XDG_*`, `ANSIBLE_*`) must not be overwritten.
- Packer-derived variables passed to Ansible must be conveyed in a way that cannot yield malformed `-e` usage (e.g. encode as a single JSON object passed via `-e`/`--extra-vars`), and must not cause argument shifting where an extra-var value becomes the playbook path.

## Constraints & Assumptions
- Assumption: The provisioner often runs the EE as the host user (UID/GID) and cannot rely on a valid passwd entry inside the image.
- Constraint: Keep behavior compatible with `ansible-navigator run` and Ansible’s `-e/--extra-vars` parsing (no shell-dependent quoting).
- Constraint: Prefer changes that are deterministic and easy to validate via unit tests.

## Acceptance Criteria
- [ ] With EE enabled and no explicit `HOME` set in `environment_variables.set`, Ansible no longer attempts to create `/.ansible` (no `Permission denied: '/.ansible'` warning).
- [ ] The provisioner never executes Ansible with a standalone `-e` / `--extra-vars` flag lacking an argument.
- [ ] The `ansible-navigator` playbook artifact no longer shows `ansible.cmdline` (command line) starting with `-e -e ...`, and the playbook path is the actual generated/uploaded playbook (not an extra-vars value).
- [ ] Unit tests cover:
  - [ ] EE default env var injection behavior (HOME/XDG defaults only when unset)
  - [ ] extra-vars argument construction cannot produce malformed `-e` usage
- [ ] Documentation includes a minimal Packer example using a collection role installed via `requirements.yml`.
