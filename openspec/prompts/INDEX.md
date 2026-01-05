# OpenSpec Prompt Index

This directory contains **numbered OpenSpec prompt files** for expanding navigator_config to support the full ansible-navigator v3.x schema. Run them in order.

## How to use

For each prompt file:

1. Run your existing `/openspec-proposal.md` workflow (it can read directly from `openspec/prompts/`)
2. Generate the proposal for that one prompt
3. Implement/review/merge that change
4. Proceed to the next prompt

Optional compatibility step:

- If you have tooling that still expects a single root prompt file, copy the current prompt to [`OPENSPEC_PROMPT.md`](../../OPENSPEC_PROMPT.md:1) before running the proposal workflow.

## Planned changes (execution order)

| # | Change ID | Title | Depends on |
|---:|----------|-------|------------|
| 01 | expand-execution-environment-config | Expand execution-environment configuration | — |
| 02 | expand-ansible-config-sections | Expand ansible.config sections | — (suggested after 01) |
| 03 | expand-logging-and-playbook-artifact | Expand logging and playbook-artifact | — (suggested after 01-02) |
| 04 | add-remaining-navigator-settings | Add remaining navigator settings | 01-03 recommended |

## Overview

This series decomposes the complex task of achieving full ansible-navigator v3.x configuration parity into four reviewable changes:

### Change 01: Execution Environment

Expands [`navigator_config.execution_environment`](../../provisioner/ansible-navigator/navigator_config.go:1) to include all v3.x EE options:
- `container_engine`, `container_options`, `volume_mounts`, `pull_arguments`
- Complete environment variable configuration (pass/set)

**Why separate:** Execution environments are complex container configurations with multiple nested structures. This change focuses exclusively on the EE block to keep it reviewable.

### Change 02: Ansible Config Sections

Expands [`navigator_config.ansible_config`](../../provisioner/ansible-navigator/navigator_config.go:1) to include all ansible.cfg sections:
- `privilege_escalation`, `persistent_connection`, `inventory`, `colors`, `diff`, `galaxy`, `paramiko_connection`
- All ansible.cfg sections supported by ansible-navigator v3.x

**Why separate:** Ansible configuration has many sections with different option types. This change adds comprehensive ansible.cfg support without mixing concerns with EE or other navigator settings.

### Change 03: Logging and Playbook Artifact

Completes [`navigator_config.logging`](../../provisioner/ansible-navigator/navigator_config.go:1) and [`navigator_config.playbook_artifact`](../../provisioner/ansible-navigator/navigator_config.go:1):
- All v3.x logging options
- Complete playbook-artifact options including replay

**Why separate:** These two related configuration areas are grouped together but kept separate from the larger EE and ansible-config changes. They're simpler and can be reviewed quickly.

### Change 04: Remaining Navigator Settings

Adds all other top-level ansible-navigator v3.x options:
- `mode_settings`, `format`, `color`, `images`, `time_zone`, `editor`
- Expanded `collection_doc_cache`, `inventory_columns`
- Any other v3.x top-level options not covered above

**Why last:** This is the "catch-all" for miscellaneous navigator settings. It depends on the previous structural changes being complete and benefits from the patterns established in changes 01-03.

## Testing Strategy

Each change includes:
- HCL decoding tests for new configuration options
- YAML generation tests verifying correct structure and key names
- Backward compatibility tests ensuring existing configs still work
- At least one minimal Packer template demonstrating new options

## Success Criteria

After completing all four changes:
- ✅ Users can configure any documented ansible-navigator v3.x option through `navigator_config` HCL blocks
- ✅ Generated ansible-navigator.yml files match the v3.x schema correctly
- ✅ No "Unsupported argument" errors for valid v3.x configuration options
- ✅ All existing navigator_config usage continues to work (backward compatibility)
- ✅ Plugin builds and passes all tests with `make generate`, `go build ./...`, and `go test ./...`

## Notes

- Each prompt is self-contained and follows the standard OpenSpec prompt structure
- Changes can be implemented independently, though the suggested order provides logical progression
- Changes 02-04 can potentially be parallelized if multiple reviewers are available
- The automatic EE defaults (safe temp directories, collections mounting) are preserved throughout - these changes only expand user configuration options, they don't change automatic behavior
