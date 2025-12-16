# Change: Simplify Galaxy dependency configuration surface

## Why

The plugin currently exposes **overlapping and inconsistently-named** configuration options for installing dependencies from `requirements_file`.

Examples of the current pain:

- Install destinations are named like caches (`roles_cache_dir`, `collections_cache_dir`) even though they behave like install paths.
- Multiple “force” knobs exist (`force_update`, `galaxy_force_install`) with unclear precedence.
- Local and remote provisioners differ in how Galaxy is invoked and how environment paths are exported.

This makes the configuration harder to learn and results in inconsistent behavior between the two provisioners.

## What Changes

This proposal defines a **minimal, coherent** dependency-install configuration surface (breaking; no backward compatibility requirements):

- Retain dependency installation driven only by `requirements_file`.
- Rename install destination options:
  - **`roles_cache_dir` → `roles_path`**
  - **`collections_cache_dir` → `collections_path`**
- Replace multiple force knobs with a single boolean:
  - **Add `galaxy_force`** (maps to `ansible-galaxy --force`)
  - **Remove `force_update`**
  - **Remove `galaxy_force_install`**
- Keep and implement consistently:
  - `galaxy_force_with_deps` (maps to `ansible-galaxy --force-with-deps`)
- Add Galaxy command escape hatches:
  - `galaxy_command` (string; defaults to `ansible-galaxy`)
  - `galaxy_args` (list(string); appended to Galaxy invocations)
- Ensure Ansible environment wiring is correct and consistent across both provisioners:
  - `roles_path` influences `ANSIBLE_ROLES_PATH`
  - `collections_path` influences `ANSIBLE_COLLECTIONS_PATHS`

## Impact

- **Affected specs:**
  - `openspec/specs/remote-provisioner-capabilities/spec.md`
  - `openspec/specs/local-provisioner-capabilities/spec.md`
- **Affected code (implementation phase; not in this workflow):**
  - `provisioner/ansible-navigator/provisioner.go`
  - `provisioner/ansible-navigator-local/provisioner.go`
  - `provisioner/ansible-navigator/galaxy.go`
  - `provisioner/ansible-navigator-local/galaxy.go`
- **Docs/examples (implementation phase):**
  - `docs/CONFIGURATION.md` and examples

## Non-Goals

- Galaxy authentication/token support.
- Dependency installation mechanisms other than `requirements_file`.
- Backward compatibility shims or deprecated aliases.
