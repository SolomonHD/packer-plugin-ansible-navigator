# Tasks

All tasks are initially unfinished. The July topology implementation is baseline code, not evidence that these acceptance criteria have passed. Required platform runs remain unfinished when unavailable, skipped, or failed.

## 1. Prove and repair Navigator executable discovery

- [x] 1.1 Add a reproduction in `path_command_test.go` with Navigator available only in a configured search directory and version checking enabled; verify it fails on the original lookup behavior and covers both version and play launch sites.
- [x] 1.2 Resolve the executable against the effective HOME-expanded search path without changing parent environment state; verify ordered configured-path precedence, inherited-PATH fallback, explicit absolute/relative paths, and executable-only command validation with focused tests.
- [x] 1.3 Retain the resolved executable for every play and the version invocation; verify subprocess fixtures observe the same executable across multiple plays and conflicting installed versions.
- [x] 1.4 Bound supported asdf/rbenv/pyenv shim resolution and reuse its result; verify hanging/failing managers, invalid results, unresolved shim chains, shared version-check budget, and independent resolution when version checking is skipped with regression tests.
- [x] 1.5 Document executable selection and resolution deadlines in `docs/TROUBLESHOOTING.md`; verify examples match the tested configuration and preserve existing skip-version-timeout warning semantics.

## 2. Prove the effective Docker endpoint and EE startup

- [x] 2.1 Use the effective Navigator environment and Docker executable in bounded context discovery; verify explicit `DOCKER_HOST`, inherited `DOCKER_CONTEXT` and related settings, named-context discovery, missing Docker, lookup failure/timeout, sanitized warnings, and fail-open optional discovery with focused tests.
- [x] 2.2 Add bounded daemon and real Navigator EE startup stages to the dedicated acceptance runner; verify stage outcomes distinguish a missing binary, denied/missing endpoint, missing image, startup error, and timeout using runner-level regression fixtures.
- [x] 2.3 Correlate real EE launch evidence with a run-specific label and harmless startup nonce; verify the runner cannot accept a Docker version response or an unrelated container event as successful Navigator startup.
- [x] 2.4 Keep debug-only preflight observational and fast; verify existing preflight tests pass and ordinary debug output launches no daemon probes or containers.
- [x] 2.5 Add Docker selection/startup troubleshooting and expected evidence to the acceptance runbook; verify diagnostics refer to the effective endpoint, redact sensitive values, and require no socket mounts, sudo, or permission changes.

## 3. Build the disposable end-to-end provisioning fixture

- [x] 3.1 Add an explicit preparation profile for a public Python/shell target image, a public EE with required Ansible/Runner/SSH tools, and an exact Docker-builder version; verify real available references, record immutable image digests and architectures, and keep preparation separate from runtime acceptance with pulls disabled there.
- [x] 3.2 Add an isolated source-plugin build/install path to the runner and expose it with a dedicated `GNUmakefile` target; verify the source binary checksum and Packer-selected path agree and user plugin installations are unchanged.
- [x] 3.3 Add current-schema HCL and a minimal builtin-module playbook using explicit Docker/stdout/EE/proxy configuration and default bridge routing; verify `packer validate` accepts the fixture with the source plugin and version checking enabled.
- [x] 3.4 Add pre-play absence and post-play exact-value checks through independent Packer shell provisioners, using a fresh nonce supplied through `play.extra_vars`; verify a host-only marker, stale marker, wrong value, or play recap alone cannot satisfy the verifier.
- [x] 3.5 Exercise the actual generated settings, inventory, private key, extra vars, and playbook through the EE; if a visibility or route defect reproduces, add its focused regression, fix only the failing path, and verify the real fixture works without broad `/tmp`, Docker socket, or host-network workarounds.
- [x] 3.6 Add an intentional Ansible task-failure mode and independent negative-control check; verify Packer returns nonzero for that task and the acceptance verifier records no successful provisioning result.
- [x] 3.7 Add bounded per-stage execution, owned-resource cleanup, and sanitized evidence capture; verify missing prerequisites fail explicit acceptance, interruption/failure cleans only owned resources, shared images remain, and incomplete cleanup prevents a full pass.
- [x] 3.8 Add the fixture/runbook instructions and evidence record format under `docs/acceptance/`; verify all required versions, identities, checksums, stages, commands, and cleanup fields exist and no full environments, private keys, or credential-bearing generated files are retained.

## 4. Demonstrate WSL2 provisioning, then repeat on native Linux

- [x] 4.1 Run the real explicit-path baseline on WSL2 with Docker Desktop integration; verify Packer exit zero, labeled EE startup, independent target marker verification, expected task-failure propagation, and successful owned cleanup, then retain the sanitized record.
- [x] 4.2 Repeat WSL2 using Navigator discoverable only via `ansible_navigator_path`, then normal PATH and a live supported shim where installed; verify each run's independent marker and record the actual executable, environment selection sources, and results.
- [x] 4.3 Select an available authorized native Linux host with Docker Engine and prepare the same recorded fixture/image inputs; verify actual host identity and kernel, local endpoint access, matching source/fixture checksums, and exact tool versions without changing production services or host permissions.
- [x] 4.4 Run explicit-path, configured-PATH-only, and normal PATH/shim variants on that native Linux host with the negative control; verify actual EE startup, independent target markers, failure propagation, and owned cleanup, retaining a separate sanitized record.
- [x] 4.5 Compare both platform records against the final source and fixture identity; rerun both after any plugin code or fixture fix, and verify every required variant is passed or the platform task remains unfinished. Document reproduced platform differences without weakening the bridge acceptance criteria.

## 5. Final integration verification

- [x] 5.1 Run `go test ./...`, `go vet ./...`, and `go build ./...`; verify they pass without requiring live acceptance resources, and run `make generate` only if mapstructure-tagged configuration structs changed, otherwise record why regeneration was unnecessary.
- [x] 5.2 Run plugin conformance for the built source binary; verify `make plugin-check` passes or document the existing bare-basename invocation needed by packer-sdc without bypassing its checks.
- [x] 5.3 Run `openspec validate prove-linux-wsl-provisioning --strict`, inspect the final diff and both real platform records, and verify the four milestones are complete with no required skip counted as success or unsupported platform claim.

## Workflow follow-up

- Present implementation and platform acceptance evidence for review before routine closeout.
- PR creation, merge, release/tagging, spec sync, and archive are separate lifecycle work; follow the user's requested workflow when those actions are authorized.
