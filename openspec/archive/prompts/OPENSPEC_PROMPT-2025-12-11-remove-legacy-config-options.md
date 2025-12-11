# OpenSpec change prompt

## Context

The Packer `ansible-navigator` provisioner frequently runs Ansible inside an **execution environment (EE) container**. In this mode, users commonly hit confusing failures/hangs because host-side configuration (e.g. `ansible.cfg` + `ANSIBLE_CONFIG`) does not reliably influence Ansible running inside the container.

This repo should make **ansible-navigator’s own config** (`ansible-navigator.yml`) the primary configuration surface so container-specific settings (temp dirs, pull policy, container engine, env propagation) are explicit and reliable.

## Goal

Add a first-class provisioner option that **generates and uses an `ansible-navigator.yml` config file** for each run, and make this the **primary** way to configure container-related behavior.

Backward compatibility is **not required** for this change (treat the plugin as still in active development regardless of the published version).

## Scope

- In scope:
  - Add a new HCL configuration option to the provisioner(s) that represents an ansible-navigator config and causes the plugin to write a temporary `ansible-navigator.yml`.
  - Ensure ansible-navigator actually uses that generated config for all executions (including EE runs).
  - Use the generated config to set EE-safe temp directories to prevent `/.ansible/tmp` permission failures.
  - Remove redundant/overlapping configuration paths that cause confusion.

- Out of scope:
  - Adding compatibility shims for older config fields.
  - Expanding features unrelated to ansible-navigator config (new play semantics, new galaxy behavior, etc.).

## Desired behavior

- New provisioner option (name is up to you, but must be explicit and discoverable), e.g. `navigator_config` or `ansible_navigator_config`:
  - Accepts a structured representation of an `ansible-navigator.yml` file (YAML-equivalent), suitable for HCL.
  - The plugin writes this config to a temporary file for the run.
  - The plugin ensures ansible-navigator loads this file for the run.

- The generated config must make EE runs non-hanging by default:
  - Configure execution-environment env vars so Ansible uses writable temp paths inside the container, at minimum:
    - `ANSIBLE_REMOTE_TMP=/tmp/.ansible/tmp`
    - `ANSIBLE_LOCAL_TMP=/tmp/.ansible-local`

- Remove redundant/overlapping configuration:
  - Remove the `ansible_cfg` mechanism and any automatic `ansible.cfg` generation logic.
  - Remove any other config surfaces that duplicate the same intent (EE temp dirs, container engine/pull policy, env propagation) so there is a single “source of truth”: the ansible-navigator config.

## Constraints & assumptions

- Assume ansible-navigator supports being directed to a config file via either:
  - an environment variable (e.g. `ANSIBLE_NAVIGATOR_CONFIG=...`), or
  - a CLI flag (e.g. `--config ...`).

Implement using whichever is supported by the ansible-navigator versions targeted by this plugin.

- The plugin must clean up any temporary config file it generates.

## Acceptance criteria

- [ ] Provisioner supports a new option that generates an `ansible-navigator.yml` file from HCL and uses it for the run.
- [ ] With `execution_environment` enabled, Ansible no longer attempts to use `/.ansible/tmp` inside the container (no permission error/hang symptom).
- [ ] `ansible_cfg` support and ansible.cfg generation are removed.
- [ ] Documentation and examples are updated to use the new ansible-navigator config option as the primary configuration mechanism.
