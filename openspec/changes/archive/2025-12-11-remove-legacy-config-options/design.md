## Context

This change intentionally prioritizes a smaller, more consistent configuration surface over compatibility with earlier plugin states.

The current codebase and specs include configuration options that exist primarily to support legacy workflows (single `playbook_file`, multi-file playbook upload, inventory group compatibility fields, and alternative dependency specification forms).

## Goals / Non-Goals

### Goals

- Remove legacy configuration options from the supported contract.
- Ensure the only provisioning unit is an ordered, repeatable `play { ... }` block.
- Ensure dependency installation uses a single mechanism: `requirements_file`.
- Keep `ansible_cfg` as the preferred interface for Ansible/runner settings.

### Non-Goals

- No attempt to preserve migration paths.
- No attempt to provide compatibility shims.
- No implementation changes in this workflow.

## Decisions

### Decision: Play execution uses ordered repeated `play` blocks only

- The user must configure one or more `play { ... }` blocks.
- Plays are executed in **declaration order**.
- A play’s `target` determines execution:
  - If `target` ends with `.yml` or `.yaml`, it is treated as a playbook path.
  - Otherwise, it is treated as a role FQDN and a temporary playbook is generated.

### Decision: Dependencies are expressed via `requirements_file`

- The plugin accepts a single `requirements_file` that can contain roles and collections.
- Inline list-style collection specifications are removed from the supported contract.

### Decision: No deprecated terminology in active specs

- Active specs and proposals MUST avoid mentioning deprecated block/field names.
- Historical details may remain only under `openspec/archive/`.

## Risks / Trade-offs

- Breaking change for anyone using legacy playbook-only config or legacy dependency mechanisms.
- Short-term friction while examples and docs are updated.

## Migration Plan (conceptual)

- Rewrite legacy playbook-only usage into a single `play { target = "..." }` block.
- Rewrite any inline dependency specifications into a `requirements.yml` and point `requirements_file` at it.

## Open Questions

- None for this proposal; scope is intentionally strict and breaking.
