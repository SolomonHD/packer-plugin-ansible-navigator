# OpenSpec change prompt

## Context

This project provides the `ansible-navigator` remote (SSH-based) and local (on-target) Packer provisioners. Today, the `command` configuration is loosely defined and the plugin does not support shell-style `~` expansion for user-supplied paths. Users who install `ansible-navigator` in custom locations or reference paths under `~` must work around this with environment setup rather than clear plugin config.

## Goal

Make the provisioner command and path handling predictable and ergonomic by:

- Normalizing what `command` means (binary only, not "binary plus arguments"),
- Adding explicit `~` (HOME) expansion for path-like config values, and
- Introducing an `ansible_navigator_path` option to influence PATH used when locating and running `ansible-navigator`.

## Scope

- In scope:
  - Remote provisioner implementation in `provisioner/ansible-navigator/`.
  - Local provisioner implementation in `provisioner/ansible-navigator-local/` where it shares the modern ansible-navigator flow.
  - Config schema changes and docs for new/clarified options (`command`, `ansible_navigator_path`, `work_dir`, file path fields).
- Out of scope:
  - Changing the high-level execution model (SSH proxy vs local execution).
  - Adding new ansible features beyond command/path behavior.

## Desired behavior

### Command semantics

- `command` in both provisioners represents only the ansible-navigator **executable** (name or absolute/relative path), not a full shell command.
- The default remote provisioner command is `ansible-navigator`, with the `run` subcommand passed as the first argument when building `exec.Command`.
- The local provisioner continues to invoke `ansible-navigator run` via its remote shell command string, but treats the configured `command` value as the executable name/path only.
- If `command` contains whitespace, `Prepare` fails fast with a clear validation error explaining that only the executable (no extra arguments) is allowed.

### Tilde (`~`) expansion

- Implement a small helper (e.g. `expandUserPath`) that expands:
  - `~` → `$HOME`
  - `~/foo` → `$HOME/foo`
  - Leaves `~user/...` unchanged (no multi-user home resolution).
- Apply this helper to path-like config fields on the **local** side before validation or use, including at minimum:
  - Remote provisioner: `command` (when it looks like a path), `work_dir`, `playbook_file`, `requirements_file`, `galaxy_file`, `ssh_host_key_file`, `ssh_authorized_key_file`, `collections_cache_dir`, `roles_cache_dir`.
  - Local provisioner: local-side `playbook_file`, `playbook_files`, `playbook_paths`, `role_paths`, `collection_paths`, `group_vars`, `host_vars`, `inventory_file`, `galaxy_file`, `requirements_file`, `collections_cache_dir`, `roles_cache_dir`, plus `command` and `work_dir` if they look like paths.
- After expansion, existing validation helpers (e.g. `validateFileConfig`, `validateDirConfig`) continue to use `os.Stat` on the expanded path.

### PATH control via `ansible_navigator_path`

- Add a new optional config field (name can be finalized in spec; assumed here as `ansible_navigator_path`) to both provisioners:
  - Type: list of strings.
  - Purpose: additional directories to prepend to PATH for resolving and running `ansible-navigator`.
- Apply `expandUserPath` to each entry in `ansible_navigator_path`.
- Remote provisioner:
  - In the version check helper, build an `exec.Command` for `ansible-navigator --version` that sets `cmd.Env` to the current environment, but with PATH replaced by:
    - The `ansible_navigator_path` entries, joined with `os.PathListSeparator`, followed by the original PATH.
  - In the actual `exec.Command` invocations for `executePlays` and `executeSinglePlaybook`, apply the same PATH construction to `cmd.Env`.
- Local provisioner:
  - When building the remote command string that runs ansible-navigator, incorporate `ansible_navigator_path` by prepending PATH at the beginning of the command (for example: `PATH="<extra>:$PATH" ...`) so the remote shell can resolve the binary in those directories.

### Backwards compatibility

- Existing configs that do not set `command` or `ansible_navigator_path` continue to work, using the current default behavior (`ansible-navigator run`) but implemented as `exec.Command("ansible-navigator", "run", ...)` on the remote provisioner.
- Existing absolute or relative paths for all file-related options remain valid; they simply gain optional `~` handling.
- If users previously relied on stuffing arguments into `command`, they get a clear error telling them to move arguments into supported fields (`extra_arguments`, play-level settings, etc.).

## Constraints & assumptions

- Assume we want minimal "magic": only `~` is expanded, not `$VARS` or other shell features.
- Assume the new `ansible_navigator_path` option should be documented alongside `command` and other core settings in the configuration docs.
- Assume tests already exist around version detection and ansible-navigator invocation; extend them to cover the new path behavior and tilde expansion.

## Acceptance criteria

- [ ] `command` is treated strictly as an executable name/path in both provisioners, and defaults to `ansible-navigator` (with `run` provided as an argument where needed).
- [ ] A helper for `~` expansion is implemented and used for all documented path-like config fields, and unit tests cover `~`, `~/subdir`, and non-`~` inputs.
- [ ] A new `ansible_navigator_path` (or agreed final name) list option exists, is HCL2-mapped, and is used to construct PATH for `ansible-navigator` version checks and runs on both remote and local provisioners.
- [ ] Version checks and ansible-navigator invocations succeed when the binary is installed only in a directory referenced via `ansible_navigator_path` (including a `~/...` entry).
- [ ] Misuse of `command` with embedded spaces is rejected at `Prepare` time with a clear, user-facing error, and this behavior is documented.
- [ ] Configuration docs are updated to describe the new `ansible_navigator_path` and the refined meaning of `command`, including examples with HOME-relative paths.
