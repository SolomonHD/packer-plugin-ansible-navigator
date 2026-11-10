# Implementation review evidence

The final qualification records identify revision
`7a9ea9e053cd4b8a1932622570d7914f8655c635` and the actual dirty state of each
invocation. The deterministic source/fixture identity, rather than the Git SHA
alone, identifies the tested code across the later squash workflow:
`8eff2ad665bf2a164719a780899f681f5ad30f04b5bf2d0e542044493b0be267`.
The aggregate covers Go source/tests, module inputs, version files, the runtime
runner and all fixture/profile files. Documentation-only edits do not change it.

The original executable regression failed at both launch sites when Navigator
was available only through the configured search directory. The implementation
now resolves against the effective child PATH, retains the selected executable,
and bounds supported shim resolution and version checking. Docker discovery uses
that effective environment and a 10-second deadline. Debug output redacts
discovered endpoints and remains observational.

Real WSL execution reproduced a separate Navigator 26.11.0/Runner noninteractive
stdin failure before the first Ansible task. The operator explicitly approved
adding `--tty=false` to enabled Docker stdout EEs unless existing user options
select TTY behavior. Its regression failed before the fix and passes afterward,
including idempotence and preservation of explicit short/long TTY options.
No new HCL fields or dependency updates were required.

Native execution then reproduced an absent default collections-cache mount:
Navigator rejected the automatically generated mount before Ansible started,
although the builtin-only fixture had no collection requirements. The operator
approved omitting that optional automatic mount only when the path was defaulted,
the directory is absent, and no requirements file is configured. Explicit user
paths/mounts and dependency-installation behavior remain authoritative. Both
final platform records were rerun after this source change.

The [WSL record](wsl2.json) proves all three variants, each with a successful
fresh target marker and the separate intentional-task-failure control. Generated
settings, inventory, key, extra-vars file and playbook were consumed by the real
EE. No generated-file visibility or bridge route repair was needed. The fixture
adds no socket mount, whole-host `/tmp` mount,
host-network mode or manual proxy address. Runner `TMPDIR` is private and outside
its working directory, so cwd mounting does not mask the temporary-file boundary.

Local final verification passed on Go 1.26.3:

- `go test ./...`, `go vet ./...`, `go build ./...`.
- `make plugin-check`, using private Go caches/GOPATH/GOBIN under `/tmp` and
  placing that GOBIN on the command's PATH. No conformance check was bypassed.
- 30 offline runner regressions, including failed prerequisites, spoofed/stale
  evidence, source-plugin selection/checksums, subprocess timeout, interruption,
  owned cleanup and cleanup failures.
- `openspec validate prove-linux-wsl-provisioning --strict`, `git diff --check`,
  and fixture `packer fmt -check`.

PR checks also verified Go 1.23.4 on Linux, macOS and Windows, plus module tidy,
formatting, import lint and documentation generation. The closeout corrections
were test-only: native POSIX shebang fixtures are guarded on Windows, cache-path
assertions use platform separators, and shell preflight fixtures avoid login
shell PATH replacement. Generated documentation was refreshed. Runtime Go code
and the real provisioning fixture were unchanged; both live platform matrices
were refreshed because the aggregate includes Go tests.

No mapstructure-tagged configuration struct changed, so HCL regeneration was
unnecessary. Ordinary Go/runner tests do not require Docker. Real acceptance is
an explicitly invoked command, and its passing records are separate evidence.

Repository-wide `openspec validate --all --strict` also ran during implementation. The three active
changes pass, but 11 unchanged main specifications fail strict lint on existing
placeholder purposes or legacy scenario formatting. Main specs were unchanged at
that verification point; the later archive sync updates the four affected
capabilities while preserving existing Purpose sections and unrelated scenarios.
The pre-existing pure-YAML versus hybrid-CLI specification drift described in
the change design remains a separate review item; this implementation follows
the current remote pure-YAML path.

The [native Linux record](native-linux.json) passes the same six builds on
`docker-infra`, Ubuntu 24.04.5 LTS with Docker Engine 29.9.3 and kernel
6.8.0-147-generic. WSL uses Docker Desktop 29.8.2 and host Python 3.10.15; native
uses host Python 3.12.3. Both use Go 1.26.3, Packer 1.14.2, Navigator 26.11.0 and
the same immutable `linux/amd64` images. Actual EE tools match: Python 3.14.10,
ansible-core 2.22.0, Runner 2.4.3 and OpenSSH 10.2p1. WSL's installed asdf shim
was exercised; native had no supported shim installed and used normal PATH.
The absent native default cache explains the second reproduced platform
difference; neither platform needed a bridge-routing workaround.

Native [preparation and cleanup](native-linux-preparation.json) retain the exact
tool downloads/checksums, real base checkout and final source overlay. Runner
cleanup and separate temporary-toolchain cleanup both passed. No owned labeled
containers remained, raw diagnostics and handoff files were deleted, and shared
prepared images were retained.

See [platform status](status.md). On November 9, 2026, the completed change was
synced to the four main capabilities and archived as
`openspec/changes/archive/2026-11-09-prove-linux-wsl-provisioning`. Merge is the
separate requested follow-up; no release or version bump is requested.
