# Change: Verify and Complete Execution Environment Configuration

## Why

The original proposal predates the restoration of `container_engine`, `container_options`, `pull_arguments`, and `volume_mounts` in both provisioners and their generated HCL schemas. Its 38-task expansion checklist no longer describes the repository. Dedicated decoding and pull-output regressions are missing, examples use an incorrect volume-mount assignment, and the on-target provisioner's active hybrid path silently drops pull arguments despite accepting them in HCL.

## What Changes

- Reconcile the existing EE contracts with the implemented HCL block schema and the SSH provisioner's automatic engine, bridge, stdout TTY, and collections-cache behavior.
- Add regressions covering actual HCL decoding for both provisioners and generated configuration for engine, ordered options, pull policy/arguments, and custom mounts.
- Preserve on-target pull arguments in the existing minimal-YAML upload path while retaining CLI flags for mapped settings. Verify the uploaded contents and command together using the communicator fixture.
- Document runtime-specific pull arguments, repeatable `volume_mounts` blocks, and the difference between host-based SSH and on-target execution. Correct and validate the example against the source-built plugin.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `remote-provisioner-capabilities`: specify and verify the already-implemented advanced EE settings without contradicting current automatic defaults.
- `local-provisioner-capabilities`: specify advanced EE settings and complete pull-argument forwarding through the active hybrid configuration path.

## Impact

- Runtime fix is limited to the on-target provisioner's existing minimal-YAML detection and generation. The public structs, HCL schema, dependencies, and SSH runtime behavior remain unchanged.
- Both provisioner test suites, README, configuration reference, EE example, and these delta specs are affected. Regenerate docs/schemas to check for drift.
- No breaking HCL/schema changes or migration are required; existing option and mount order is preserved.
- Acceptance uses HCL decoding, generated YAML, on-target communicator integration, example validation, and repository checks. Prior WSL/native live evidence qualifies the SSH provisioner only. Live on-target container-engine provisioning remains separately unqualified and is not claimed by these fixture tests.
