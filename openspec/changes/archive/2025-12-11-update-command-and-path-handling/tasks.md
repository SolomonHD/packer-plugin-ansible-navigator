## 1. Implementation

- [x] 1.1 Review existing specs in `openspec/specs/remote-provisioner-capabilities/spec.md` and `openspec/specs/local-provisioner-capabilities/spec.md` to confirm current assumptions about `command`, path handling, and ansible-navigator invocation.

- [x] 1.2 Implement a shared helper function (suggested name: `expandUserPath`) to expand HOME-relative paths on the local side:
  - Expand `~` → `$HOME`
  - Expand `~/foo` → `$HOME/foo`
  - Leave `~user/...` unchanged (no multi-user home resolution)
  - Do NOT support environment variable expansion or other shell features

- [x] 1.3 Update the remote ansible-navigator provisioner to:
  - Treat `command` strictly as the executable name or path.
  - Default `command` to `ansible-navigator` and pass `run` as the first argument when building `exec.Command`.
  - Apply HOME expansion to specific local-side path-like fields before validation:
    - `playbook_file`
    - `playbook_files`
    - `playbook_paths`
    - `role_paths`
    - `collection_paths`
    - `group_vars`
    - `host_vars`
    - `inventory_file`
    - `galaxy_file`
    - `requirements_file`
    - `collections_cache_dir`
    - `roles_cache_dir`
    - `command` (when it looks like a path)
    - `work_dir` (when it looks like a path)
  - Ensure existing validation helpers (`validateFileConfig`, `validateDirConfig`) operate on expanded paths

- [x] 1.4 Update the local ansible-navigator provisioner to:
  - Treat `command` strictly as the executable name or path in its remote shell command string.
  - Continue invoking `ansible-navigator run` by default.
  - Apply HOME expansion to specific path-like fields before validation:
    - `command` (only when it looks like a path, e.g., starts with `/`, `./`, or `~`)
    - `work_dir`
    - `playbook_file`
    - `requirements_file`
    - `galaxy_file`
    - `ssh_host_key_file`
    - `ssh_authorized_key_file`
    - `collections_cache_dir`
    - `roles_cache_dir`
  - Ensure existing validation helpers (`validateFileConfig`, `validateDirConfig`) operate on expanded paths

- [x] 1.5 Add `ansible_navigator_path` (or the finalized name from the spec) to both provisioner Config structs and HCL2 specs as a repeatable list of strings.

- [x] 1.6 For the remote provisioner, update ansible-navigator invocations to use modified PATH:
  - In the version check helper (for `ansible-navigator --version`)
  - In `executePlays` method (when running multiple plays)
  - In `executeSinglePlaybook` method (when running a single playbook)
  - For each invocation, construct `cmd.Env` with PATH set to: `ansible_navigator_path` entries (HOME-expanded, joined with `os.PathListSeparator`) followed by the original PATH

- [x] 1.7 For the local provisioner, update remote shell command construction:
  - Prepend PATH override at the beginning of the remote shell command
  - Format: `PATH="expanded_entry1:expanded_entry2:$PATH" ansible-navigator run ...`
  - HOME-expand each `ansible_navigator_path` entry on the local side before constructing the remote command
  - Use colon (`:`) as the path separator in the shell command (standard Unix/Linux path separator)

- [x] 1.8 Add validation logic in both provisioners that rejects `command` values containing whitespace and returns a clear user-facing error explaining that only the executable (no extra arguments) is allowed and that additional arguments must use supported fields like `extra_arguments`.

- [x] 1.9 Ensure existing default behavior remains unchanged for configurations that do not set `command` or `ansible_navigator_path`, apart from the new HOME expansion support for path-like fields.

## 2. Testing and validation

- [x] 2.1 Add or extend unit tests for both provisioners to cover:
  - HOME expansion:
    - `~` → `$HOME`
    - `~/subdir` → `$HOME/subdir`
    - `~user/...` preserved unchanged
    - Non-`~` paths preserved unchanged
  - PATH construction using `ansible_navigator_path`:
    - ansible-navigator binary only present in configured directories
    - Multiple entries in `ansible_navigator_path`
    - HOME-relative entries in `ansible_navigator_path` (e.g., `~/bin`)
  - Validation failures:
    - `command` with single embedded space
    - `command` with multiple embedded spaces
    - `command` with leading/trailing spaces
  - Regression coverage:
    - Existing dual invocation modes (playbook vs. play blocks) still work
    - Default behavior unchanged when `command` and `ansible_navigator_path` are not set
    - Existing absolute/relative paths still work as before

- [x] 2.2 Run `go test ./...` and ensure tests pass.

- [x] 2.3 Run `go build ./...` and `make plugin-check` to confirm the plugin still passes build and plugin validation.

- [x] 2.4 Add or update documentation examples in `docs/CONFIGURATION.md` and any relevant troubleshooting guides to describe:
  - The refined meaning of `command` (executable name/path only, no arguments)
  - The new `ansible_navigator_path` option with examples:
    - `ansible_navigator_path = ["~/bin", "/opt/ansible/bin"]`
    - How it affects PATH resolution for ansible-navigator
  - Usage of HOME-relative paths in configuration:
    - `command = "~/bin/ansible-navigator"`
    - `playbook_file = "~/ansible/site.yml"`
    - `inventory_file = "~/inventory/hosts"`
    - `work_dir = "~/ansible-work"`
  - Migration guidance for users who previously embedded arguments in `command`
  - Clear examples of what is now invalid vs. the correct alternative

- [x] 2.5 Re-run `openspec validate update-command-and-path-handling --strict` and ensure the change passes validation.
