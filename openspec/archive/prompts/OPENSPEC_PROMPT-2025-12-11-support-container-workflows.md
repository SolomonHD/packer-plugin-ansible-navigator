# OpenSpec change prompt

## Context

Users running `ansible-navigator` in containerized environments (specifically WSL2) experience "hangs" because:

1. The plugin's SSH proxy binds only to `127.0.0.1`, making it unreachable from the execution environment container.
2. The generated inventory hardcodes `ansible_host=127.0.0.1`, which resolves to the container itself, not the host.
3. Python output buffering hides the connection timeout logs, making it appear as a silent hang.

## Goal

Update the plugin to support container-based workflows (like WSL2) by making the SSH proxy networking configurable and forcing unbuffered output.

## Scope

- **In scope:**
  - Add configuration to control the SSH proxy bind address.
  - Add configuration to control the host address used in the generated Ansible inventory.
  - Automatically force unbuffered Python output.
  - Update documentation with a specific section for WSL2/Container troubleshooting.

## Desired behavior

1. **New Config Option: `ansible_proxy_bind_address`**
   - Defaults to `127.0.0.1` (secure default).
   - Users can set to `0.0.0.0` to allow connections from containers.

2. **New Config Option: `ansible_proxy_host`**
   - Defaults to `127.0.0.1`.
   - Users can set to `host.containers.internal`, `host.docker.internal`, or a specific IP to tell the container how to reach the host.
   - This value is used in the generated inventory file (`ansible_host=<value>`).

3. **Unbuffered Output**
   - The plugin automatically injects `PYTHONUNBUFFERED=1` into the `ansible-navigator` environment to ensure logs stream immediately to Packer.

4. **Documentation**
   - Add a "WSL2 and Containers" section to `docs/TROUBLESHOOTING.md`.
   - Explain the need for `ansible_proxy_bind_address = "0.0.0.0"` and `ansible_proxy_host = "host.containers.internal"` (plus `--add-host` arg) in WSL2.

## Constraints & assumptions

- Assumption: The user is responsible for adding necessary container arguments (like `--add-host`) via `extra_arguments` if their container runtime requires it.
- Constraint: Default behavior must remain secure (bind to localhost) and backward compatible.

## Acceptance criteria

- [ ] `ansible_proxy_bind_address` is configurable and affects `net.Listen`.
- [ ] `ansible_proxy_host` is configurable and appears in the generated inventory file.
- [ ] `PYTHONUNBUFFERED=1` is present in the executed command environment.
- [ ] `docs/TROUBLESHOOTING.md` contains clear instructions for WSL2 users.
