## Context

This change standardizes the dependency installation surface for both provisioners.

The current state exposes multiple overlapping fields and inconsistent semantics (install paths vs “cache dirs”, multiple force flags, and inconsistent invocation patterns).

## Goals / Non-Goals

- Goals:
  - Provide a minimal set of dependency-install options that is identical across both provisioners.
  - Make option names reflect semantics (`roles_path`/`collections_path` as install destinations).
  - Ensure Ansible runtime receives consistent environment variables so that installed content is discoverable.
  - Provide escape hatches for custom `ansible-galaxy` locations and additional flags.

- Non-Goals:
  - Adding new dependency declaration formats beyond `requirements_file`.
  - Backward compatibility for removed/renamed fields.

## Decisions

### Decision: `roles_path`/`collections_path` are treated as install destinations and exported to Ansible

- `roles_path` SHALL be used as the roles install destination AND exported via `ANSIBLE_ROLES_PATH`.
- `collections_path` SHALL be used as the collections install destination AND exported via `ANSIBLE_COLLECTIONS_PATHS`.

This keeps “where Galaxy installs” and “where Ansible searches” aligned.

### Decision: `galaxy_force_with_deps` takes precedence over `galaxy_force`

`ansible-galaxy`’s `--force-with-deps` implies forcing dependency installs as well.

- If `galaxy_force_with_deps=true`, the plugin SHOULD pass `--force-with-deps`.
- If `galaxy_force=true` and `galaxy_force_with_deps=false`, the plugin SHOULD pass `--force`.
- If both are set, `--force-with-deps` SHOULD be used and `--force` omitted.

### Decision: `galaxy_args` is appended last

`galaxy_args` is designed as an escape hatch for flags like `--ignore-certs`.

- The plugin SHOULD construct `ansible-galaxy ...` arguments from structured config first.
- Then append `galaxy_args` so users can add additional flags without breaking required structure.

## Risks / Trade-offs

- **Breaking changes**: removing/renaming options will break existing templates.
  - Mitigation: update docs and provide a migration section describing renames.

- **Environment variable semantics**: users may expect colon-separated lists.
  - Mitigation: treat `roles_path` and `collections_path` as opaque strings passed directly to the environment variables.

## Open Questions

- Should `offline_mode` map to `ansible-galaxy --offline` for both roles and collections installs, or should it disable installation entirely?
  - This proposal specifies `--offline` mapping.
