# Prove Linux and WSL provisioning

Packer and Navigator run on the host. Navigator uses that host's selected Docker
endpoint to launch an execution environment (EE). Ansible in the EE connects to
Packer's SSH proxy, which forwards operations through the Docker builder's
communicator into a separate, disposable Python target. Neither container is the
host. The target does not need an SSH daemon. The fixture uses normal Docker
bridge routing and the plugin's automatic proxy defaults.

The command is deliberately separate from `make test` and the optional BusyBox
TCP diagnostic. A passing shell invocation, a readable socket, a Docker version
response, or an Ansible recap alone is insufficient. Acceptance requires a
source-built plugin, real labeled EE startup, a fresh target, exact target marker
verification, expected task-failure propagation, and confirmed owned cleanup.

## Inputs and preparation

Run from the repository root on WSL2 with Docker Desktop integration or native
Linux with Docker Engine. Required host tools are Go (the version in `go.mod`),
Packer 1.14.2 or later, Python 3, Docker, and a real Navigator installation. The
initial observed WSL baseline is Packer 1.14.2 and Navigator 26.11.0; an observation
is not a qualification result. The runner records the versions actually used.

The checked-in [profile](../../provisioner/ansible-navigator/test-fixtures/linux-wsl/profile.json)
selects public Python and Ansible development images by immutable index digest,
records their observed amd64 manifest digests, and pins Docker builder 1.1.2.
The initial profile is `linux/amd64`; do not silently substitute images for a
different architecture. Runtime startup records Python, ansible-core, Runner,
and SSH from the actual EE and fails if required tools are absent.

Preparation is a separate, explicit action. It can download the pinned public
images and the exact builder into a dedicated directory. It does not install the
source plugin into the user's Packer plugin directory.

```sh
python3 scripts/linux_wsl_acceptance.py prepare --prepared-dir /tmp/navigator-prepared
```

Runtime acceptance never pulls images or downloads plugins. Keep prepared images
and the prepared builder between runs. No registry login, private collections,
cloud credentials, Docker socket mounts, socket permission changes, sudo,
daemon restarts, or Docker context switches are part of this procedure.

Public reference checks used during profile selection:

- [Docker builder configuration](https://developer.hashicorp.com/packer/integrations/hashicorp/docker/latest/components/builder/docker)
- [Exact Docker builder source, v1.1.2](https://github.com/hashicorp/packer-plugin-docker/tree/v1.1.2)
- Registry `docker manifest inspect` for each checked-in digest and bounded local
  image metadata (`RepoDigests`, OS and architecture only). Tags in the profile
  are selection history, never runtime inputs.

## Run the baseline, then discovery

```sh
python3 scripts/linux_wsl_acceptance.py run \
  --prepared-dir /tmp/navigator-prepared \
  --record /tmp/navigator-wsl-result.json
```

The default sequence establishes the explicit real executable path first, then
Navigator available only through `ansible_navigator_path`, then inherited PATH.
An installed supported shim is exercised through the inherited-PATH variant;
its manager and resolved executable must be recorded. Each variant uses a new
nonce and disposable target, version checking enabled, Docker/stdout/EE enabled,
and `use_proxy=true`. Use `--navigator /absolute/path/to/ansible-navigator` to
select an installation. `--variant` supports diagnosing an individual variant;
an individual passing run does not replace the full platform sequence.

When Python itself uses an asdf shim, that shim can prepend its real installation
directory to the child PATH. To preserve the original shell PATH and exercise
Navigator's installed shim explicitly, invoke the runner with the real Python
path, for example:

```sh
"$(asdf which python3)" scripts/linux_wsl_acceptance.py run \
  --prepared-dir /tmp/navigator-prepared \
  --record /tmp/navigator-wsl-result.json
```

For Docker stdout EEs, the plugin supplies `--tty=false` unless user container
options explicitly select TTY behavior. This prevents Runner from requesting an
interactive Docker terminal in Packer's noninteractive subprocess. Startup probes
use that same option; neither probe nor build needs an interactive terminal.

The runner builds and installs the source plugin into private per-run
`PACKER_PLUGIN_PATH`, copies the prepared builder there, verifies Packer's selected
plugin path against the source binary checksum, and runs `packer validate`
before builds. It does not run `packer init` during acceptance. Its temporary
generated settings, inventory, keys, and logs must be removed even on failure.

## Required stage evidence and diagnosis

| Stage | Bound | Passing evidence / failure interpretation |
| --- | --- | --- |
| Executable | 10 seconds for resolution | Effective search path selects a real executable; supported shim resolution succeeds. A missing command or invalid manager result is a lookup failure. See [troubleshooting](../TROUBLESHOOTING.md). |
| Docker daemon | 30 seconds | Bounded Docker server response under the invocation's environment. Missing Docker, denied/missing Unix endpoint, and timeout are distinct failures. |
| Image and EE startup | 120 seconds | Prepared immutable image present; real Navigator `exec` emits the startup nonce and Docker records a start event matching this run's label and image. Unrelated events and Docker version output cannot satisfy it. |
| Template and plugin | Bounded by runner | Selected plugin path and checksum match the source build; current-schema HCL validates. Missing builder preparation fails here. |
| SSH/proxy and files | 10 minutes per build | Real play consumes generated inventory, temporary key, extra-vars file and playbook. File-not-found and SSH connection failures remain failures. Inspect the failing boundary before proposing a narrow fix. |
| Target marker | Within build | Pre-play shell proves absence; builtin Ansible copy writes the fresh nonce; subsequent independent target shell compares exact bytes and emits `ACCEPTANCE_VERIFIED:<nonce>`. Host-only, stale or wrong markers cannot pass. |
| Negative control | 10 minutes per build | An intentional builtin `fail` emits `ACCEPTANCE_INTENTIONAL_FAILURE:<nonce>`, Packer exits nonzero, and no successful target-verification result exists. A generic startup error is not the expected task failure. |
| Cleanup | Bounded separately | Only this run's labeled containers and private temporary state are removed; prepared/shared images remain. Unknown or failed cleanup prevents a full pass. |

Docker receives inherited `DOCKER_HOST`, `DOCKER_CONTEXT`, `DOCKER_CONFIG` and TLS
settings with its normal precedence. Configured Navigator search directories
also apply when locating Docker. Optional plugin context discovery uses a
10-second deadline and fails open with a sanitized warning; this does not prove
daemon access. Debug preflight is observational and does not launch daemon probes
or containers. The strict runner separately tests access and startup.

For a bounded local diagnostic, use `timeout 30 docker version --format
'{{.Server.Version}}'` under the same environment. A successful result proves
only daemon access. Describe selection as inherited context, inherited endpoint,
or active context; do not paste endpoint credentials, query strings, full
environments, broad Docker inspection objects, or generated SSH/extra-vars files
into a report. A remote/TLS endpoint is outside this baseline. Do not switch to
host networking, mount all `/tmp`, or mount the Docker socket to obtain a pass.

## Evidence records and platform status

Retain only the runner's sanitized JSON after reviewing it. Use a separate record
for each platform and link it from [status](status.md). Records must include:

- Host identity, actual kernel and platform; Docker client/server and builder versions.
- Source SHA and dirty state; deterministic source/fixture/profile identity and
  the exact built and selected plugin checksums/paths.
- Reproduction command, Navigator invocation and resolved executable, selection
  sources, host Python and actual EE Python/Runner/ansible-core/SSH versions.
- Immutable images and architecture; per-variant nonces, labeled startup events,
  Packer result, independent marker verification, and negative-control outcome.
- Each stage result and cleanup result separately; absent or failed stages stay
  pending, never passing skips.

Do not commit private raw logs, keys, full environments, generated credential-bearing
files, or registry authentication data. A source/plugin/fixture change invalidates
both earlier platform records; rerun both on the final identical inputs. A Linux
container under WSL is a target, not native-Linux-host evidence. If an authorized
native host lacks prerequisites, record that limitation. Use an existing
toolchain or obtain explicit authorization for an isolated temporary test
toolchain. Do not reconfigure production services or alter host permissions to
manufacture acceptance.

Implementation review requires passing WSL2/Docker Desktop and native
Linux/Docker Engine records, including every required discovery variant. Merge,
release, spec sync and archive are separate lifecycle actions.
