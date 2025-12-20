# OpenSpec Change Prompt

## Context
The `packer-plugin-ansible-navigator` provisioners install Ansible collections via `ansible-galaxy` into a configurable `collections_path` (defaulting under `~/.packer.d/ansible_collections_cache`). When execution environments are enabled, the plugin also generates an `ansible-navigator.yml` and adds a container volume mount + env vars so Ansible inside the container can discover installed collections.

Today, the plugin incorrectly appends an extra `ansible_collections` path segment when generating the execution-environment mount source, which can produce an invalid directory layout inside the container (e.g., the container sees `<mount>/<namespace>/<collection>` instead of `<mount>/ansible_collections/<namespace>/<collection>`). This can cause role FQDN lookups like `integration.common_tasks.etc_profile_d` to fail even though the collection was installed successfully.

## Goal
Make collection installation paths and execution-environment mounts consistent so that collections installed by `ansible-galaxy` are reliably discoverable by Ansible/ansible-navigator inside execution environments.

## Scope

**In scope:**
- Remove any automatic appending of `ansible_collections` to the configured `collections_path` when generating the execution-environment mount and related defaults.
- Ensure the execution-environment mount source/destination and env vars point to a directory layout that matches Ansible’s expectations for installed collections.
- Apply the fix consistently to both provisioners:
  - `provisioner/ansible-navigator`
  - `provisioner/ansible-navigator-local`
- Update/extend unit tests to cover the corrected path behavior.

**Out of scope:**
- Changing how roles are installed or discovered.
- Changing default cache directory locations.
- Adding new user-facing configuration options unless strictly required.

## Desired Behavior
- When `collections_path` is set (or defaulted) and execution environments are enabled, the plugin mounts the host directory such that the container can discover collections at:
  - `<container_mount_root>/ansible_collections/<namespace>/<collection>`
- The plugin must not introduce an extra `ansible_collections` suffix that breaks the directory layout.
- The plugin sets collection discovery environment variables in a way that works with current Ansible behavior:
  - Prefer `ANSIBLE_COLLECTIONS_PATH` (singular) and do not require `ANSIBLE_COLLECTIONS_PATHS`.
  - Do not override user-provided env vars if they already set/pass `ANSIBLE_COLLECTIONS_PATH`.
- The generated `ansible-navigator.yml` contains the expected `volume-mounts` and env var values when EE is enabled.

## Constraints & Assumptions
- Assumption: `ansible-galaxy collection install -p <collections_path>` installs collections under `<collections_path>/ansible_collections/...`.
- Constraint: Do not change the meaning of the user-facing `collections_path` option; treat it as the root install directory that contains (or will contain) `ansible_collections/`.
- Constraint: Keep behavior deterministic and avoid breaking existing configurations that already work.
- Assumption: The repo already has test coverage for navigator config generation and can be extended to assert correct mount source/destination.

## Acceptance Criteria
- [ ] With execution environments enabled, a collection installed to `collections_path` is discoverable inside the container using a role FQDN (e.g., `namespace.collection.role`) without requiring users to add manual mounts.
- [ ] The plugin no longer appends `ansible_collections` to `collections_path` when generating execution-environment mounts.
- [ ] The generated `ansible-navigator.yml` includes a `volume-mounts` entry that results in `<container_mount_root>/ansible_collections/...` existing in the container.
- [ ] The plugin uses `ANSIBLE_COLLECTIONS_PATH` (singular) for discovery and does not clobber a user-provided `ANSIBLE_COLLECTIONS_PATH`.
- [ ] Unit tests are updated/added to prevent regressions in collections path handling for both the remote and local provisioners.
