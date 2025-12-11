# Change: Remove legacy config options and deprecated terminology

## Why

The plugin is still in development and does not need to preserve legacy configuration surfaces inherited from the fork lineage. Keeping deprecated options and migration-era semantics increases cognitive load, expands the schema unnecessarily, and encourages new users to copy patterns we do not want to support long-term.

## What Changes

This change proposal defines a **simplified, HCL2-first** configuration contract that:

- Uses **only** repeated `play { ... }` blocks for provisioning (ordered; no playbook-only legacy path).
- Uses **only** `requirements_file` for dependency installation (roles + collections).
- Keeps `ansible_cfg` as the primary Ansible configuration mechanism.
- Removes **legacy / deprecated** configuration options and avoids mentioning deprecated terms (e.g., plural `plays`) in active specs.

### Configuration surface after this change (high-level)

Shared (both provisioners):

- `play { ... }` (repeatable, ordered)
- `requirements_file` (optional)
- `ansible_cfg` (optional; may be auto-defaulted when `execution_environment` is set)
- Runtime options (minimal): `command`, `ansible_navigator_path`, `navigator_mode`, `execution_environment`, `work_dir`, `keep_going`, `structured_logging`, `log_output_path`, `verbose_task_output`, `skip_version_check` (remote only), `version_check_timeout` (remote; local reserved)

## Impact

- **Specs affected**:
  - `remote-provisioner-capabilities`
  - `local-provisioner-capabilities`

- **Breaking change**: **Yes**.
  - Removes legacy playbook-only configuration and other legacy configuration knobs.

- **Code and docs likely affected (non-exhaustive; not implemented here)**:
  - `provisioner/ansible-navigator/provisioner.go`
  - `provisioner/ansible-navigator-local/provisioner.go`
  - Generated `*.hcl2spec.go`
  - `README.md` and `docs/*` configuration docs
