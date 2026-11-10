## Context

The remote provisioner starts an SSH adapter on the Packer host and writes that adapter's address into generated inventory. Today, `Prepare` independently defaults both `ansible_proxy_bind_address` and `ansible_proxy_host` to `127.0.0.1`. That is correct when Ansible runs on the host, but an enabled execution environment moves Ansible into a container where loopback is a different network namespace.

The repository already has a partial Docker solution: generated Navigator configuration adds `--add-host=gateway.docker.internal:host-gateway` when the advertised proxy host is exactly `gateway.docker.internal`. Users must nevertheless discover and set both proxy fields, and every documented EE example currently omits them. Podman conventionally exposes `host.containers.internal`, while explicit `--network=host` makes host loopback usable again. The implementation must select one coherent topology rather than defaulting each address in isolation.

## Goals / Non-Goals

**Goals:**

- Make a default remote-provisioner EE configuration able to connect from its container to the host SSH adapter on WSL2 and native Linux.
- Preserve secure loopback binding when Ansible runs outside an EE or when an EE explicitly uses host networking.
- Resolve Docker, Podman, and automatic engine selection deterministically and expose that decision in diagnostics.
- Preserve valid explicit proxy addresses and container options while rejecting combinations known to be unreachable.
- Cover both pure resolution logic and the actual container-to-host connection path.

**Non-Goals:**

- Add a new HCL networking abstraction or manage arbitrary custom container networks.
- Modify the local provisioner, whose Ansible process runs on the target rather than through the host SSH adapter.
- Guarantee host-gateway behavior for unsupported or remote container daemons.
- Replace Ansible Navigator or Ansible Runner's container lifecycle management.

## Decisions

### Resolve one proxy topology before applying address defaults

Add a pure resolver used by remote `Prepare` before empty proxy strings are defaulted. Its inputs are EE enabled state, requested container engine, container options, and the two proxy address values including whether each was supplied. Its output records the resolved engine, network mode, bind address, advertised host, and which values were inferred.

The resolution matrix is:

| Execution topology | Default bind address | Default inventory host |
| --- | --- | --- |
| EE disabled | `127.0.0.1` | `127.0.0.1` |
| EE using host networking | `127.0.0.1` | `127.0.0.1` |
| Docker EE using bridge networking | `0.0.0.0` | `gateway.docker.internal` |
| Podman EE using bridge networking | `0.0.0.0` | `host.containers.internal` |

An explicit non-empty proxy field is never replaced. An omitted companion field is inferred from the selected topology. Known-impossible combinations, such as a bridge-network EE advertising `127.0.0.1` or a bridge route using a loopback-only listener, fail preparation with corrective guidance.

Alternative considered: always bind to `0.0.0.0` and advertise a gateway. This makes non-EE operation unnecessarily exposed and breaks the useful host-network topology. Rejected in favor of topology-aware defaults.

### Recognize host networking from existing container options

Do not add a new HCL field. Normalize the existing `container_options` forms `--network=host`, `--network host`, `--net=host`, and `--net host`. If present, select the host-network row and do not inject a host-gateway mapping.

Alternative considered: require the two proxy fields in all EE configurations. This retains the current failure-prone experience and leaves the documented minimal examples invalid.

### Resolve `auto` to the engine Navigator will use

Treat an empty engine and `auto` as automatic. Probe supported executables in Navigator's documented preference order, select the first available engine, and write that resolved value to generated Navigator configuration so proxy resolution and Navigator cannot choose different engines. An explicitly configured engine remains authoritative. If automatic selection cannot find a supported engine, return an actionable preparation error.

For Docker bridge networking, retain the existing `gateway.docker.internal` alias and inject its `host-gateway` mapping. For Podman bridge networking, use `host.containers.internal` and rely on Podman's host-gateway convention; do not pass Docker's `host-gateway` syntax.

Alternative considered: use one hostname for all runtimes. Docker Engine on Linux and Podman do not provide the same alias semantics, so a single implicit hostname is not portable.

### Separate resolution, validation, and observability

Keep topology resolution as a side-effect-free helper with table-driven unit tests. Validate the resolved result during `Prepare`. At provisioning time, log the actual listener address after port allocation and the advertised inventory endpoint. Normal output gets a concise route message; debug output includes the selected engine, network mode, inferred/explicit origin of each address, and generated host mapping. Errors must report the actual subprocess exit code rather than replacing it with a constant.

This separation makes the networking rules testable without Docker and leaves the runtime acceptance test focused on the network boundary.

### Add an opt-in connectivity acceptance test

Add a Linux-compatible acceptance test that starts the same TCP/SSH adapter listener shape used by the provisioner, launches a disposable container with the generated runtime options, and confirms a connection to the advertised host and allocated port. Gate it behind the project's acceptance-test environment convention and skip with a reason when the selected runtime or test image is unavailable. The test must exercise bridge defaults; a second case covers explicit host networking where supported.

The existing acceptance test that substitutes `ansible-playbook` does not cover Navigator or container networking and remains useful for its original scope.

## Risks / Trade-offs

- [Binding `0.0.0.0` exposes the ephemeral adapter port on host interfaces during a bridged EE run] → Keep loopback everywhere else, use one-time SSH keys, close the listener promptly, and state the exposure in debug output and documentation.
- [Executable probing could diverge from a future Navigator engine preference] → Centralize the preference order, test it, and pin the resolved engine into generated YAML.
- [Remote Docker daemons cannot reach the local Packer listener through their own `host-gateway`] → Detecting daemon locality is outside this change; document the limitation and require explicit addresses for remote daemons.
- [User-supplied custom network options may not fit bridge or host classification] → Preserve explicit addresses, classify only recognized host-network spellings, and provide actionable validation rather than guessing for known contradictions.
- [Docker Desktop host networking availability varies by platform and settings] → Keep bridge mode as the default and make host-network acceptance coverage conditional.

## Migration Plan

1. Introduce the pure topology resolver and unit-test the matrix, explicit overrides, option parsing, and contradictions.
2. Integrate resolution into remote `Prepare`, then feed the same result to adapter setup, inventory generation, Navigator YAML generation, and diagnostics.
3. Update generated configuration tests and add the opt-in real-container acceptance test.
4. Update examples and troubleshooting docs before release.

Rollback consists of reverting the resolver integration and documentation. Users can retain the former behavior without rollback by explicitly setting both proxy fields to `127.0.0.1` together with host networking or a non-EE execution path.

## Open Questions

- Confirm Navigator's current automatic engine preference against the supported Navigator version during implementation and encode it in one tested helper.
- Decide which small, pinned container image the connectivity acceptance test will use, favoring an image already present in project CI or configurable through an environment variable.

