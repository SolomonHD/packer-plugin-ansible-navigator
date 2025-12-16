# OpenSpec Orchestrator Workflow

This workspace includes a custom KiloCode mode, **OpenSpec Orchestrator**, defined in [`.kilocodemodes`](.kilocodemodes:1).

## What problem this solves

The built-in OpenSpec prompt workflow commonly assumes **one prompt file** at a time (typically [`OPENSPEC_PROMPT.md`](OPENSPEC_PROMPT.md:1)).

For larger refactors, it is better to:
- Decompose work into multiple small OpenSpec changes
- Produce one prompt per change
- Track ordering and dependencies

## Output convention

The orchestrator writes prompt files under:

- `openspec/prompts/INDEX.md`
- `openspec/prompts/NN-<change-id>.md` (NN = 2-digit order number)

Example:

- `openspec/prompts/01-simplify-galaxy-config-surface.md`
- `openspec/prompts/02-add-play-extra-args.md`

## Handoff to proposal generation

The proposal workflow supports **Prompt Directory Mode**, so it can read prompts directly from `openspec/prompts/`.

For each numbered prompt:

1. Run the `/openspec-proposal.md` workflow and point it at the prompts folder (it will auto-pick `NEXT.md` or the lowest `NN-*.md`)
2. Generate the OpenSpec proposal for that one change
3. Implement and merge that change
4. Repeat for the next prompt

Optional compatibility step:

- If you have tooling that still expects a single root prompt file, you may copy the current prompt into [`OPENSPEC_PROMPT.md`](OPENSPEC_PROMPT.md:1) temporarily.

## Acceptance criteria for the orchestrator output

- `INDEX.md` lists all changes, in order, with dependencies.
- Each prompt file is self-contained and scoped to one change.
- Each prompt file includes:
  - Context
  - Goal
  - Scope (In/Out)
  - Desired Behavior
  - Constraints & Assumptions
  - Acceptance Criteria
