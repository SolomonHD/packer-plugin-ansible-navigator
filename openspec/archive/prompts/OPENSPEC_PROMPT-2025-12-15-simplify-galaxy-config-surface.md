# OpenSpec Change Prompt

## Context
The `packer-plugin-ansible-navigator` repo currently has overlapping and inconsistently-named Galaxy configuration for dependency installation.
For example, the code uses install destinations named `roles_cache_dir`/`collections_cache_dir` even though they behave as install paths, and it has multiple "force" knobs (e.g., `force_update`, `galaxy_force_install`).
The local and remote provisioners also differ in how Galaxy is invoked and how environment paths are exported.

## Goal
Define and implement a minimal, coherent Galaxy configuration surface (no backward compatibility requirements) that is consistent across both provisioners.

## Scope

**In scope:**
- Keep dependency installation driven only by `requirements_file`.
- Rename install destination options:
  - `roles_cache_dir` -> `roles_path`
  - `collections_cache_dir` -> `collections_path`
- Replace multiple force knobs with a single Boolean:
  - Add `galaxy_force` (Boolean)
  - Remove `force_update`
  - Remove or rename `galaxy_force_install` to `galaxy_force` (no aliases)
- Keep and implement consistently for both provisioners:
  - `galaxy_force_with_deps`
- Add two escape hatches:
  - `galaxy_command` (string; defaults to `ansible-galaxy`)
  - `galaxy_args` (list(string); appended to Galaxy commands)
- Ensure environment wiring is correct and consistent:
  - `roles_path` influences `ANSIBLE_ROLES_PATH`
  - `collections_path` influences `ANSIBLE_COLLECTIONS_PATHS` (or the correct Ansible env var naming)
- Regenerate HCL2 specs after Config changes.
- Update docs and examples to use the new names.

**Out of scope:**
- Adding Galaxy authentication/token support.
- Supporting dependency installation mechanisms other than `requirements_file`.
- Backward compatibility shims or deprecated aliases.

## Desired Behavior
- Users configure dependency installation like:

```hcl
provisioner "ansible-navigator" {
  requirements_file = "./requirements.yml"

  roles_path       = "./.ansible/roles"
  collections_path = "./.ansible/collections"

  galaxy_force           = true
  galaxy_force_with_deps = true

  galaxy_command = "/opt/venv/bin/ansible-galaxy"
  galaxy_args    = ["--ignore-certs"]

  play { target = "site.yml" }
}
```

- Galaxy installation uses `galaxy_command` and `galaxy_args` consistently for:
  - roles install
  - collections install
- `galaxy_force` results in `--force`.
- `galaxy_force_with_deps` results in `--force-with-deps`.
- `roles_path`/`collections_path` are used as install destinations and are exported to Ansible runtime via environment.

## Constraints & Assumptions
- Assumption: breaking changes are acceptable (no migration path).
- Constraint: both provisioners must expose the same Galaxy options.
- Constraint: `requirements_file` remains the only supported dependency declaration.

## Acceptance Criteria
- [ ] Remote provisioner Config includes: `requirements_file`, `offline_mode`, `roles_path`, `collections_path`, `galaxy_force`, `galaxy_force_with_deps`, `galaxy_command`, `galaxy_args`.
- [ ] Local provisioner Config includes the same options.
- [ ] `force_update` is removed everywhere.
- [ ] Galaxy invocations use `galaxy_command` and include `galaxy_args`.
- [ ] Roles/collections install destinations use `roles_path`/`collections_path`.
- [ ] Environment variables for roles/collections are set consistently for both provisioners.
- [ ] HCL2 specs regenerated and docs updated.

## Expected areas/files touched
- `provisioner/ansible-navigator/provisioner.go`
- `provisioner/ansible-navigator-local/provisioner.go`
- `provisioner/ansible-navigator/galaxy.go`
- `provisioner/ansible-navigator-local/galaxy.go`
- `docs/CONFIGURATION.md` and any examples

