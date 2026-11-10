# Design

## Context

See `proposal.md` for motivation and scope. This change starts from `debug/connection-refused` at `4fd58cb`; the archived July change already supplies Docker bridge proxy defaults. Its real-container tests skipped, so that implementation is an input to qualify, not proof of a working build.

Source inspection identified specific gaps:

- `getVersion()` calls `exec.LookPath` and constructs the command before assigning the PATH produced by `buildEnvWithPath`. The child environment cannot repair lookup that has already occurred. Existing path tests mostly check string expansion and skip the version invocation.
- Shim resolution changes a local variable used for version checking. Play execution constructs a fresh command from `p.config.Command`, so it can launch a different executable or reenter the shim.
- `resolveDockerHost()` uses an unbounded subprocess and the parent environment. Debug preflight observes `/var/run/docker.sock` and parent state, not actual child daemon access. Existing specs require that debug diagnostic to stay fast and non-blocking.
- `provisioner_acc_test.go` and its fixture retain the inherited `ansible-playbook` setup, old provisioner/schema names, and recap-based verification. The July connectivity test launches BusyBox `nc`; neither demonstrates the complete Navigator-to-target path.
- Navigator reads settings on the host; generated inventory, private key, extra vars, and playbook must be visible across the host/EE boundary. The extra-vars file is created under host `/tmp`; successful command discovery alone says nothing about its visibility inside the EE.

There is pre-existing specification drift: `remote-provisioner-capabilities` still includes CLI-primary `Navigator Config File Generation` text, and `command-argument-construction` retains hybrid/minimal-YAML requirements. They contradict `configuration-generation`'s `Pure YAML Configuration`, the remote `Default Navigator Command`, and the actual remote execution path. This change follows the current remote pure-YAML configuration contract and preserves Ansible task arguments; it must not reintroduce obsolete EE configuration flags to satisfy historical hybrid text. Broader reconciliation of those specifications and local-provisioner parity is deferred and must remain visible in implementation review.

Read-only discovery on November 9, 2026 found a working WSL shell invocation of Navigator 26.11.0 and Packer 1.14.2, and Docker client/server 29.8.2 reachable through the default Unix endpoint. None of these observations counts as acceptance. No native Linux run has been observed.

## Goals / Non-Goals

**Goals:** isolate the first failing boundary, fix reproduced defects, and retain repeatable proof of target provisioning on both required host platforms. Keep the production configuration surface unchanged and use a single source revision and fixture for qualification.

**Non-Goals:** changes to the local provisioner, new runtime engines, remote/TLS Docker qualification, private collections, production infrastructure, or a broad dependency refresh. The installed baseline is the starting point; a dependency change must be justified by a reproduction and followed by requalification. Do not claim coverage beyond the exact recorded configurations.

## Decisions

### 1. Resolve once using the environment that will execute

Build an effective environment without mutating `os.Environ`. HOME-expand configured directories and prepend them in order. Resolve executable names against that PATH before calling `exec.Command`; preserve explicit absolute paths and resolve explicit relative paths against the preparation directory. Store the selected absolute executable in internal provisioner state and reuse it for the version check and every play. Preserve PATH for Python shebangs, Docker, and other child tools.

Use the existing supported shim detection for asdf, rbenv, and pyenv. Invoke the manager using the effective environment and the executable basename, validate the resulting executable, and reject unresolved shim chains. Bound resolution to 10 seconds; with version checking enabled, use the smaller remaining `version_check_timeout` budget and let resolution plus `--version` share that budget. With `skip_version_check=true`, retain the independent 10-second resolution bound while skipping only `--version`; keep existing version-timeout warning behavior. No new HCL option is needed.

Regression tests use temporary executable fixtures and a restricted parent PATH to show that the configured path actually controls both launch sites. Tests cover multiple plays, conflicting installations, explicit/relative paths, failed/hanging managers, and skipped version checks. Synthetic executables are unit-test fixtures only, never live acceptance substitutes.

Alternative: add more shell startup or global PATH adjustments. Rejected because the plugin must honor its own documented path input in a noninteractive process.

### 2. Separate Docker discovery from acceptance of Docker access

Pass the effective environment and effective Docker executable into bounded context discovery. Preserve explicit `DOCKER_HOST`, `DOCKER_CONTEXT`, Docker configuration, and TLS-related environment settings; use Docker's own precedence rather than inventing an override order. Retain optional discovery's fail-open behavior, with a sanitized warning, and add a 10-second context lookup deadline. A missing or failed optional discovery is never a successful access result.

The explicit acceptance command separately requires a bounded successful daemon probe and a real Navigator `exec` probe using the same EE settings and effective environment as the subsequent play. The tiny startup probe emits a harmless run nonce and records launch evidence for the selected image. Each stage has its own failure classification. Keep ordinary debug preflight fast and observational; do not make it launch containers or block on a health check.

The supported baseline is a local Unix endpoint: WSL2 Docker Desktop integration first, native Linux Docker Engine second. Unit coverage includes an explicit endpoint, an active named context, lookup timeout, absent Docker, and preserved inherited settings. A real alternate/rootless endpoint is recorded only if actually tested; it is not a prerequisite for the two primary baseline records.

Alternative: stat `/var/run/docker.sock` or mount it into the EE. Rejected as proof: the selected endpoint can differ, and the host Navigator process is responsible for launching Docker.

### 3. Use a minimal disposable target and an independent verifier

Add a separate acceptance fixture under `provisioner/ansible-navigator/test-fixtures/` and a dedicated runner under `scripts/`, exposed by a new explicit make target rather than replacing ordinary `make test` or silently running stale inherited acceptance cases. The runner builds this source plugin into a temporary isolated `PACKER_PLUGIN_PATH`, records its checksum, installs the Docker builder at an exact recorded version there, and checks Packer's selected plugin path. Do not write to the normal user plugin installation or permit a downloaded Navigator plugin to shadow the build.

Use a Docker-builder target with `discard=true`, a shell, and Python 3. No SSH daemon is needed on the target: the existing Packer communicator/SSH adapter route is the behavior being proved. Use a public EE containing ansible-core, Runner, SSH tooling, and Python. Select exact available images during explicit preparation and check in immutable digest references before acceptance; do not invent digest values or rely on `latest`. Record the Docker-builder version and image/platform architecture. Use the same inputs on both platforms, or document architecture-specific digests from the same image index.

The template uses current `provisioner "ansible-navigator"` and one `play` block, explicit `container_engine="docker"`, `mode="stdout"`, `enabled=true`, `use_proxy=true`, and `skip_version_check=false`. Use the existing bridge proxy defaults without manual proxy addresses, host networking, extra capabilities, or socket mounts. No Galaxy requirements, private registry login, cloud credentials, or role collections participate in this first baseline.

Before the play, an independent shell provisioner proves the run-specific marker is absent on the fresh target. Pass a fresh harmless nonce through `play.extra_vars`; Ansible writes it using a builtin module to a target-only path. A following shell provisioner reads and compares the exact value before the builder discards the target. The play targets generated inventory, never `localhost` or `connection: local`. A success requires Packer exit zero, real labeled EE launch evidence, and independent target verification. A recap or host-side marker cannot satisfy it.

Use per-run container labels and bounded metadata/event collection to attribute EE launches without broad Docker inspection. Exercise generated file visibility through the real play and fix any reproduced staging/mount defect with narrowly scoped paths. Never bind all host `/tmp` or the Docker socket to make the test pass. Add an intentional Ansible failure mode; its separate negative-control check succeeds only when Packer fails for that task and records no successful provisioning result.

Alternative: reuse only the BusyBox TCP probe or mock Navigator. Both can isolate a component but omit the integration that has never worked.

### 4. Qualify explicit configuration first, then discovery, on both hosts

Run the same fixture in three variants: an explicit real Navigator path; a command name discoverable only via `ansible_navigator_path`; normal inherited PATH, including a live supported shim where used. Restrict PATH in the second variant while retaining required Docker/Python/Packer utilities; do not remove or alter the user's installations. Give every build its own target and nonce.

The runner reports separate stages for executable resolution, daemon access, image/startup, SSH/proxy and file visibility, target marker verification, negative control, and cleanup. Default bounds are 10 seconds for discovery, 30 seconds for daemon probing, 120 seconds for a startup probe, and 10 minutes per Packer build. Image downloads/builds belong to a separately invoked preparation step; runtime acceptance uses already prepared images with pull disabled.

Ordinary unit tests require no containers. The dedicated acceptance command is strict: missing prerequisites return nonzero with a blocked-stage reason. A failure after prerequisites, including host networking or image permission failure, cannot turn into a skip. Keep the existing optional TCP test unchanged as a component diagnostic.

Record sanitized evidence under `docs/acceptance/` for WSL2/Docker Desktop and native Linux/Docker Engine separately. Include reproduction commands, code SHA and dirty state, fixture/profile checksums, plugin checksum, executable path, Packer/Navigator/Python/Runner/ansible-core/Docker/builder versions, image digests, selection sources, stage outcomes, target verification and cleanup. Capture dependency versions from the actual host/EE used, not an assumed package set. Private raw logs, keys, full environments, and generated credential-bearing files are not committed. Source/plugin or fixture changes invalidate both prior platform results; documentation-only updates retain the recorded tested code identity.

Select the actual native Linux host during implementation using the user's authorized access path; a Linux target/container under WSL is not a second host platform. If it is unavailable, keep that task pending and report partial acceptance. Do not deploy or reconfigure a host to manufacture evidence.

### 5. Keep resources and completion evidence attributable

Use a private per-run directory and record resource identities before creating disposable containers or an isolated plugin installation. Cleanup handles success, failure, timeout, and interruption, removing only owned state. Never prune shared resources or remove prepared/shared base images. Report the build outcome and cleanup outcome independently; unknown cleanup is not a fully passing run. Do not retain sensitive temporary content for debugging.

The runbook mirrors these stages and explains Packer host, Navigator host process, EE, proxy, and target as separate boundaries. Existing specifications require PATH control and proxy connectivity, but current tests do not prove them; the new deltas make executable reuse and full acceptance explicit. The earlier optional-discovery and component-test skip behavior remains valid outside the strict acceptance command.

## Risks / Trade-offs

- [Multiple changing dependencies obscure the failure] -> Start with the installed baseline and current module pins. Record a narrowly justified dependency adjustment and rerun the matrix if required.
- [Docker Desktop cannot expose the Packer listener through its gateway] -> Let the bridge build fail visibly, inspect the recorded route, and fix or explicitly document the reproduced blocker. Host networking cannot substitute for bridge acceptance.
- [EE cannot read generated host files] -> Exercise those files in the real play and use explicit minimal staging/mounts with regression tests when needed.
- [Different executables pass validation and run provisioning] -> Record and reuse the resolved executable and test both launch sites.
- [Native Linux access is unavailable] -> Finish independent implementation work but leave native Linux acceptance and change completion unfinished.
- [A stale installed plugin produces a misleading pass] -> Isolate installation, record binary/fixture checksums, and verify Packer's selected path.

## Migration Plan

1. Reproduce executable discovery defects with local fixture tests, then implement shared resolution and executable reuse.
2. Bound context discovery and add stage-specific access probes to the dedicated acceptance runner.
3. Add the current-schema target fixture and independent verification; qualify WSL explicit-path baseline before discovery variants.
4. Run the same final source and fixture on native Linux. Repeat both platforms after any code/fixture fix, then review sanitized evidence.

No public configuration migration is planned. Rollback reverts resolver/context/harness changes; it does not alter Docker services, user installations, or the archived July history.

## Closeout

- Implementation is ready for review only after the required tests and both platform records pass; artifact existence or skipped tests do not satisfy acceptance.
- PR creation, merge, release/tagging, spec sync, and archive are later lifecycle actions, outside the tracked implementation checklist and subject to their own user requests.
