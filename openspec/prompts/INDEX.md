# OpenSpec Prompt Index (Generated)

This directory contains **numbered OpenSpec prompt files**. Run them in order.

## How to use

For each prompt file:

1. Run your existing `/openspec-proposal.md` workflow (it can read directly from `openspec/prompts/`)
2. Generate the proposal for that one prompt
3. Implement/review/merge that change
4. Proceed to the next prompt

Optional compatibility step:

- If you have tooling that still expects a single root prompt file, copy the current prompt to [`OPENSPEC_PROMPT.md`](OPENSPEC_PROMPT.md:1) before running the proposal workflow.

## Planned changes (execution order)

| # | Change ID | Title | Depends on |
|---:|----------|-------|------------|
| 01 | simplify-galaxy-config-surface | Minimal Galaxy config surface + naming | — |
| 02 | add-play-extra-args | Add `play.extra_args` escape hatch | 01 |
| 03 | navigator-config-add-local-tmp | Add `navigator_config.ansible_config.defaults.local_tmp` | — |
| 04 | remove-work-dir | Remove `work_dir` from both provisioners | — |
| 05 | warn-on-skip-version-check-timeout | Warn when `skip_version_check` ignores timeout | — |
| 06 | plugin-debug-logging-linked-to-navigator-log-level | Gate plugin debug output on `navigator_config.logging.level` | — |
| 07 | debug-preflight-ee-docker-dind-warnings | DEBUG-only EE/Docker/DinD preflight warnings | 06 |

## Notes

- Prompts are intentionally scoped so each OpenSpec change is reviewable.
- Numbering is the stable dependency order.
