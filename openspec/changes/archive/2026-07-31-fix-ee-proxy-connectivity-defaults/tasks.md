## 1. Proxy Topology Resolution

- [x] 1.1 Confirm the supported Ansible Navigator version's automatic container-engine preference and encode it in one documented resolver constant or helper.
- [x] 1.2 Add a side-effect-free proxy topology model and resolver covering EE disabled, Docker bridge, Podman bridge, and host networking.
- [x] 1.3 Parse all supported `--network` and `--net` host-network option spellings without modifying the user's option order.
- [x] 1.4 Preserve explicit proxy values, infer only omitted companions, and return actionable errors for known-unreachable bridge combinations.
- [x] 1.5 Add table-driven unit tests for the complete topology matrix, engine auto-resolution, explicit overrides, option parsing, and validation errors.

## 2. Remote Provisioner Integration

- [x] 2.1 Resolve and store the proxy topology during remote `Prepare` before applying proxy address defaults.
- [x] 2.2 Pin an automatically selected container engine in the Navigator configuration so execution uses the same engine selected for proxy routing.
- [x] 2.3 Feed the resolved bind address to SSH adapter setup and the resolved advertised host to generated inventory.
- [x] 2.4 Log the allocated listener endpoint and advertised inventory endpoint, and add debug details for engine, network mode, value origins, and automatic mappings.
- [x] 2.5 Preserve and report the actual Ansible Navigator subprocess exit code on failure.
- [x] 2.6 Extend remote provisioner preparation and execution tests for defaults, explicit values, diagnostics, and failure reporting.

## 3. Navigator Configuration Generation

- [x] 3.1 Refactor automatic EE defaults to consume the resolved topology instead of inferring behavior from the proxy hostname alone.
- [x] 3.2 Generate the Docker `gateway.docker.internal:host-gateway` option only for the matching Docker bridge topology and avoid duplicates.
- [x] 3.3 Preserve Podman, host-network, custom-host, and user-supplied container options without Docker-specific injection.
- [x] 3.4 Extend YAML generation unit tests to cover each engine/network topology and option-preservation case.
- [x] 3.5 If implementation changes any configuration struct or mapstructure tag, run `make generate` and verify the corresponding `*.hcl2spec.go` output; otherwise record that regeneration was not required.

  Regeneration was not required because no configuration struct or mapstructure tag changed.

## 4. Real-Container Acceptance Coverage

- [x] 4.1 Add an opt-in Linux-compatible acceptance-test harness with configurable runtime and pinned test image prerequisites.
- [x] 4.2 Prove that a disposable Docker bridge container can connect through the generated gateway option to the allocated host listener.
- [x] 4.3 Add conditional host-network connectivity coverage and precise skip reasons for unsupported runtimes, daemon features, or missing images.
- [x] 4.4 Ensure every acceptance-test path closes listeners and removes disposable containers.

## 5. Documentation

- [x] 5.1 Update the minimal remote EE example to use `mode = "stdout"` and the new topology-aware defaults.
- [x] 5.2 Document Docker bridge, Podman bridge, and explicit host-network routes for WSL2 and native Linux.
- [x] 5.3 Add connection-refused troubleshooting based on the logged listener, advertised host, engine, network mode, and host mapping.
- [x] 5.4 Document non-loopback listener exposure, one-time-key mitigation, platform constraints, remote-daemon limitations, and explicit override behavior.

## 6. Verification

- [x] 6.1 Run `go test ./...`, including the new unit coverage and the opt-in container acceptance test when prerequisites are available.

  The opt-in acceptance cases skipped because the pinned test image was not locally available; the harness does not pull images implicitly.
- [x] 6.2 Run `go vet ./...`, `go build ./...`, and `make plugin-check`.

  `packer-sdc plugin-check` passed from `/tmp` using the bare plugin basename required by packer-sdc's name validator; the Make target cannot accept an absolute `BINARY` path for this check.
- [x] 6.3 Run `openspec validate --all --strict` and resolve all errors and warnings introduced by this change.
