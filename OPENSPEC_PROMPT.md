# OpenSpec Prompt: See Decomposed Prompts

## Notice

This single-file prompt has been **decomposed into multiple numbered OpenSpec prompt files** for better reviewability and incremental implementation.

## Where to Find the Prompts

All prompts are in the [`openspec/prompts/`](openspec/prompts/INDEX.md:1) directory:

- **[`INDEX.md`](openspec/prompts/INDEX.md:1)** - Overview and execution order
- **[`00-update-yaml-to-v2-format.md`](openspec/prompts/00-update-yaml-to-v2-format.md:1)** - **CRITICAL PREREQUISITE**: Fix YAML V1→V2 format issue
- **[`02-expand-ansible-config-sections.md`](openspec/prompts/02-expand-ansible-config-sections.md:1)** - Expand ansible.config sections
- **[`03-expand-logging-and-playbook-artifact.md`](openspec/prompts/03-expand-logging-and-playbook-artifact.md:1)** - Expand logging and playbook-artifact
- **[`04-add-remaining-navigator-settings.md`](openspec/prompts/04-add-remaining-navigator-settings.md:1)** - Add remaining navigator settings

## Start Here

**Read [`openspec/prompts/INDEX.md`](openspec/prompts/INDEX.md:1) first** for the full picture.

## Critical First Step

**Prompt 00 must be completed before any others.** It fixes a critical bug where:
- ansible-navigator 25.x shows migration prompts during Packer builds
- `pull_policy = "never"` doesn't work, causing Docker to attempt registry pulls

The original single-prompt content describing this issue is now in [`00-update-yaml-to-v2-format.md`](openspec/prompts/00-update-yaml-to-v2-format.md:1).

## Workflow

For each numbered prompt:

1. Run your existing `/openspec-proposal.md` workflow (it reads from `openspec/prompts/`)
2. Generate the proposal for that one prompt
3. Implement/review/merge that change
4. Proceed to the next prompt

Optional compatibility: If you have tooling that expects a single root prompt file, copy the current prompt to this file before running the proposal workflow.

---

**Original single-prompt content has been split for reviewability. See [`INDEX.md`](openspec/prompts/INDEX.md:1) for details.**
