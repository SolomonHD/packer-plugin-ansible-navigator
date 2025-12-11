# Change: Normalize ansible-navigator command and path handling

## Why

The current provisioner configuration blurs the meaning of the `command` field, relies on ambient PATH when resolving the ansible-navigator binary, and does not provide first-class support for HOME-relative paths. This makes it hard for users to understand how commands are built, how binaries are located, and how to use `~`-prefixed paths in a predictable way.

## What Changes

- Define `command` consistently for both SSH-based (remote) and local ansible-navigator provisioners as **executable name or path only**, not a full shell command.
- Make the default behavior explicit: remote provisioner uses `ansible-navigator` with `run` passed as an argument; the local provisioner continues to invoke `ansible-navigator run` via its remote shell string but treats `command` as the executable only.
- Reject configurations that embed arguments in `command` (detectable via whitespace) at validation time, with clear user-facing errors that direct users to supported fields such as `extra_arguments` and play-level options.
- Introduce deterministic HOME expansion (`~` → $HOME, `~/subpath` → $HOME/subpath) for path-like configuration fields on the local side of both provisioners, without adding broader shell interpolation or environment-variable expansion.
- Specify exactly which fields participate in HOME expansion for each provisioner (e.g., playbooks, requirements, inventory, Galaxy-related paths, work directories, and any path-like command overrides).
- Add a new list option (tentatively named `ansible_navigator_path`) that prepends additional directories to PATH when locating and running ansible-navigator, with clear behavior for both remote (local exec.Command) and local (remote shell) provisioners.
- Preserve backward compatibility for existing configurations that do not set `command` or `ansible_navigator_path`, and for existing absolute/relative file paths, while tightening semantics only where users previously relied on unsupported patterns (e.g., stuffing arguments into `command`).

## Impact

- **Affected specs:**
  - `remote-provisioner-capabilities` — clarify default command semantics, add requirements for HOME expansion and PATH control.
  - `local-provisioner-capabilities` — clarify default command semantics, add requirements for HOME expansion and PATH control.
- **Affected implementation (future task):**
  - Remote provisioner: `provisioner/ansible-navigator/` (command building, version checks, path expansion helper).
  - Local provisioner: `provisioner/ansible-navigator-local/` (command building, remote shell PATH construction, path expansion helper).
  - Tests covering tilde expansion, PATH behavior, and command validation for both provisioners.
  - Configuration docs: `docs/CONFIGURATION.md` and related troubleshooting sections.
- **Risks and considerations:**
  - Tightening `command` semantics may surface validation errors for users who previously embedded arguments; mitigation is to provide precise error messages and migration guidance.
  - HOME expansion must be carefully scoped (`~` only) to avoid surprising users who expect raw values to be passed through unchanged.
  - PATH manipulation via `ansible_navigator_path` must be clearly documented to avoid conflicts with environment-level PATH customization in CI/CD pipelines.
