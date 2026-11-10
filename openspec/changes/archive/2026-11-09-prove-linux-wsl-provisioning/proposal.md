# Proposal: Prove provisioning on WSL and native Linux

## Why

The project has no recorded successful end-to-end provisioning run on either WSL2 or native Linux. Binary discovery, Docker access, execution-environment startup, and proxy connectivity must be proved together before additional features or a broad dependency refresh can establish a useful baseline.

## What Changes

- Establish four ordered milestones: resolve and execute Navigator reliably; prove its Docker daemon and execution-environment access; complete a real Packer build that provisions a disposable target; repeat that build on WSL2 with Docker Desktop integration and native Linux with Docker Engine.
- Resolve `command` against the effective `ansible_navigator_path` plus inherited PATH before constructing subprocesses. Reuse the resolved executable for version checking and every play, including supported version-manager shim resolution with bounded execution.
- Exercise Docker context/endpoint selection and access with the same effective environment used for Navigator. Distinguish executable, daemon, image/startup, SSH/proxy, and target-provisioning failures with actionable, sanitized evidence.
- Add a minimal, isolated acceptance fixture using the locally built plugin, explicit Docker and stdout mode, an enabled execution environment, and the branch's bridge proxy defaults. Ansible writes a unique run marker on the target; a subsequent Packer shell provisioner verifies its exact value independently.
- After the explicit-path baseline passes, repeat with Navigator discoverable only through `ansible_navigator_path`, then the normal PATH/shim route where installed. Record the exact tested revisions, versions, immutable image references, commands, stage results, and cleanup status for each platform.
- Make the explicitly requested acceptance command fail when required prerequisites are missing or any stage is unproved. Ordinary unit tests stay independent of Docker; a skip cannot count as platform acceptance.
- Provide one repeatable runbook and a two-platform verification record. Keep the existing quick TCP test as a narrower diagnostic, with its original opt-in behavior.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `remote-provisioner-capabilities`: Correct executable lookup and reuse across validation and play execution while preserving the executable-only command contract.
- `context-resolution`: Bound Docker context lookup and verify that the selected local Docker endpoint is usable in Navigator's effective environment.
- `build-tooling`: Add isolated end-to-end provisioning acceptance and require successful WSL2 and native Linux evidence.
- `documentation`: Document staged setup, reproducible acceptance, failure diagnosis, and the distinction between shell checks and plugin execution.

## Impact

- Builds on `debug/connection-refused` at `4fd58cb`, including the archived July proxy topology implementation; it does not reopen or rewrite that historical change.
- Expected implementation paths: `provisioner/ansible-navigator/provisioner.go`, executable/path/shim tests, Docker context tests, acceptance harness and fixtures, `GNUmakefile`, and acceptance/troubleshooting documentation.
- No HCL fields are added, removed, or renamed. Explicit commands and proxy settings remain authoritative. A resolved absolute executable replaces the previous literal-name launch; missing executables and timed-out resolution produce earlier, more specific failures.
- The first baseline uses existing repository dependencies and installed tools; discovery observed Packer 1.14.2 and Navigator 26.11.0 in the current WSL shell. Those checks are not end-to-end acceptance or a claim of native Linux support. Change dependencies only when a reproduced blocker requires it, then rerun both platforms on the same resulting revision.
- Scope is the SSH-based `ansible-navigator` provisioner with a local Docker endpoint. Local provisioner parity, Podman/Apple Container qualification, remote Docker daemons, private registries/collections, broad dependency modernization, and infrastructure deployment are deferred.
- Implementation creates only disposable test resources under explicit acceptance execution. Creating this proposal does not authorize running containers or changing workstation tools, socket permissions, Docker services, or production resources.
