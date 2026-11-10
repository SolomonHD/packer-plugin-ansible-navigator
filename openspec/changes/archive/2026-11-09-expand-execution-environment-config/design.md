# Design

## Context

See proposal.md for the reconciliation rationale. Both provisioners already have advanced EE structs and generated schemas. Their execution paths differ: the SSH provisioner supplies full YAML; the on-target provisioner combines CLI flags with a minimal settings file for fields not mapped by its flag builder. The original proposal had no design artifact. This document records the reconciled approach during the separately requested apply work.

## Goals / Non-Goals

**Goals:** Exercise the public HCL schema and the configuration actually handed to each invocation. Extend the existing on-target minimal settings path to carry ordered pull arguments.

**Non-Goals:** Replace the hybrid invocation architecture, add public fields or dependencies, change SSH defaults, or claim live on-target Docker acceptance from mock-communicator tests.

## Decisions

- Decode real HCL through the generated schema and SDK cty/config conversion. Direct struct or map construction alone would miss schema drift and incorrect mount syntax.
- Keep SSH full YAML generation unchanged and test all pull combinations plus automatic-default interactions. The fields already exist, so reintroducing them would add no behavior.
- Treat nonempty local pull arguments as a reason to upload minimal YAML. Put only the ordered arguments at `ansible-navigator.execution-environment.pull.arguments`; retain pull policy and other mapped settings on their established CLI path. Adding a CLI mapping or switching to full YAML would change invocation behavior beyond the scoped forwarding gap.
- Observe uploaded content, settings-file reference, mapped flags, and temporary-file cleanup together using the local communicator fixture, for success and failure. Fixture checks establish the handoff contract without requiring privileged target runtime setup.
- Use a companion ping playbook and separate host/target example builds. Validate both with an isolated source-built plugin installation; do not run the example's infrastructure builds as part of this change.

## Risks / Trade-offs

- Existing local engine/option CLI mappings have not been certified against Navigator 26.11 by these fixtures → Preserve them, identify the limitation in the verification record, and require separate live on-target compatibility qualification before treating that runtime combination as accepted.
- Settings-file and CLI precedence can hide dropped fields → Assert exact minimal YAML and command together, with policy-only/arguments-only/combined/empty coverage and artifact/cache coexistence.
- Remote cache/default behavior differs from target paths → Keep SSH default-cache omission rules scoped to SSH and document where mount sources must exist.
- Main OpenSpec files contain pre-existing legacy contracts and formatting warnings → Limit this delta to advanced EE behavior and its local hybrid extension; record repository-wide validation findings separately.

## Migration Plan

No configuration migration, dependency update, or schema change is required. Deploy the rebuilt plugin through the normal plugin installation process. Reverting the local runtime change restores the prior behavior of silently ignoring local pull arguments. No deployment or release is performed during apply.

## Closeout

Review the implementation and verification record before archive or merge. Spec sync, archive, commits/pushes, merging, and release work are separate lifecycle actions.
