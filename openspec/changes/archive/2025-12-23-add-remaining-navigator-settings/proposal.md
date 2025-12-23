# Change: Add remaining `navigator_config` top-level ansible-navigator v3.x settings

## Why

The plugin supports `navigator_config { ... }` to generate an `ansible-navigator.yml`, but today the typed schema only models a subset of the ansible-navigator v3.x settings surface. Users attempting to configure documented v3.x keys such as `mode-settings`, `format`, `color`, `images`, `time-zone`, `documentation`, `editor`, `inventory-columns`, or richer `collection-doc-cache` options can hit “Unsupported argument” errors because these fields are not present in the generated HCL2 spec.

## What Changes

- Expand the typed `NavigatorConfig` model for both provisioners to include the remaining documented ansible-navigator v3.x **top-level** configuration sections:
  - `mode_settings` (per-mode configuration)
  - `format` (output format configuration)
  - `color` (color scheme settings)
  - `images` (container image preferences)
  - `time_zone` (time zone)
  - `documentation` (documentation source URLs)
  - `editor` (editor settings)
  - `inventory_columns` (inventory display columns)
  - `replay` (replay options; if top-level in the v3.x schema)
  - Expand `collection_doc_cache` beyond its current minimal fields if required by the v3.x schema
- Ensure YAML generation emits these keys with correct ansible-navigator YAML names:
  - HCL/Go: underscore names (e.g., `time_zone`, `mode_settings`)
  - YAML: hyphenated names where required (e.g., `time-zone`, `mode-settings`)
- Keep the existing `navigator_config` behavior stable for already-supported sections (`execution-environment`, `ansible`, `logging`, `playbook-artifact`, existing `collection-doc-cache`).

## Non-Goals

- No changes to provisioner registration or provisioner names.
- No changes to execution-environment modeling (handled by earlier prompt items).
- No changes to ansible-config modeling (handled by earlier prompt items).
- No validation of ansible-navigator option values beyond schema typing (ansible-navigator is the source of truth for runtime validation).

## Impact

- **Affected specs (deltas in this proposal):**
  - `openspec/specs/remote-provisioner-capabilities/spec.md`
  - `openspec/specs/local-provisioner-capabilities/spec.md`
  - `openspec/specs/documentation/spec.md`
- **Expected implementation areas (not in this proposal workflow):**
  - Typed config structs and `//go:generate packer-sdc mapstructure-to-hcl2 ...` type lists in both provisioners
  - YAML conversion logic for `ansible-navigator.yml` generation
  - Regenerated HCL2 spec files (`*.hcl2spec.go`)
  - Unit tests covering HCL decoding and YAML generation

## Notes

- This change is intentionally **additive**: it expands supported keys under `navigator_config` and SHOULD NOT break existing configurations.
- The authoritative list of v3.x options MUST be sourced from upstream ansible-navigator documentation (e.g., the settings reference), and the implementation should enumerate “remaining top-level keys” explicitly before coding.
