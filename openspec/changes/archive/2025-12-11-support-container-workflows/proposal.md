# Change: Support Container Workflows (WSL2)

## Summary

Enhance the `ansible-navigator` plugin to support containerized environments like WSL2 by making the SSH proxy networking configurable and forcing unbuffered output.

## Motivation

Users running `ansible-navigator` in containerized environments (specifically WSL2) experience "hangs" because:

1. The plugin's SSH proxy binds only to `127.0.0.1`, making it unreachable from the execution environment container.
2. The generated inventory hardcodes `ansible_host=127.0.0.1`, which resolves to the container itself, not the host.
3. Python output buffering hides the connection timeout logs, making it appear as a silent hang.

## Proposed Solution

1. **New Config Option: `ansible_proxy_bind_address`**
    * Defaults to `127.0.0.1` (secure default).
    * Users can set to `0.0.0.0` to allow connections from containers.

2. **New Config Option: `ansible_proxy_host`**
    * Defaults to `127.0.0.1`.
    * Users can set to `host.containers.internal`, `host.docker.internal`, or a specific IP to tell the container how to reach the host.
    * This value is used in the generated inventory file (`ansible_host=<value>`).

3. **Unbuffered Output**
    * The plugin automatically injects `PYTHONUNBUFFERED=1` into the `ansible-navigator` environment to ensure logs stream immediately to Packer.

4. **Documentation**
    * Add a "WSL2 and Containers" section to `docs/TROUBLESHOOTING.md`.
