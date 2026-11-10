# Tasks

The original expansion is already implemented in both public structs and generated schemas. This reconciled checklist tracks the remaining work; it does not claim the former 38 tasks were performed in this session.

## 1. Host-based SSH configuration coverage

- [x] 1.1 Add actual HCL decoding regressions for engine, ordered options, pull arguments, and repeated volume mount blocks; verify invalid mount-list assignment is rejected by the generated schema.
- [x] 1.2 Add remote YAML pull-combination and all-fields regressions, and verify existing engine/TTY/bridge/mount-default tests pass without changing SSH runtime behavior.

## 2. On-target configuration forwarding

- [x] 2.1 Add equivalent local HCL and generated-YAML regressions, and forward pull arguments through the existing minimal-YAML path; verify arguments-only, combined policy, policy-only, empty, and unrelated unmapped-settings cases.
- [x] 2.2 Verify with a communicator integration test that Provision uploads the exact pull arguments, retains mapped CLI settings, references the uploaded file, and cleans the temporary local file.

## 3. Configuration documentation and example

- [x] 3.1 Update README and configuration reference with engine/options/pull arguments, mount blocks, runtime-specific options, existing automatic defaults, and target-path semantics; verify examples match the HCL decoding tests.
- [x] 3.2 Correct the EE example to demonstrate advanced fields for both provisioners and verify Packer formatting and validation against the source-built plugin without building infrastructure.

## 4. Integration verification

- [x] 4.1 Run generation, Go tests/vet/build, plugin-check, and strict validation of this change; record outcomes and any pre-existing repository-wide validation findings in a verification record.

## Workflow follow-up

- Review the implementation and verification evidence before archiving or merging.
- Keep existing SSH WSL/native qualification distinct from fixture-based on-target coverage; live on-target EE qualification is a separate acceptance effort.
- No release or version change is included.
