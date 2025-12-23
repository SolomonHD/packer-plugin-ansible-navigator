## Context

`navigator_config { ... }` is a typed HCL2 surface that is converted into an `ansible-navigator.yml` file and passed to the runtime via `ANSIBLE_NAVIGATOR_CONFIG`.

The current implementation uses:

- Typed Go structs (RPC-safe; no `map[string]interface{}` + `cty.DynamicPseudoType`)
- Generated HCL2 specs (`packer-sdc mapstructure-to-hcl2`)
- A manual YAML conversion step that maps underscore field names to the ansible-navigator YAML schema (hyphenated keys and nested objects).

This proposal expands the remaining **top-level** ansible-navigator v3.x settings.

## Decisions

### Decision: Continue using typed structs (no dynamic maps)

All new nested configuration sections (e.g., `mode_settings`, `format`, `color`, `images`, `documentation`, `editor`, `replay`) will be represented as typed structs (and included in the `packer-sdc mapstructure-to-hcl2` type list), so the generated HCL2 spec includes them.

### Decision: Preserve underscore-in-HCL / hyphen-in-YAML mapping

- HCL keys and `mapstructure` tags remain underscore-based (e.g., `time_zone`, `mode_settings`).
- The YAML generator converts to the ansible-navigator schema names (e.g., `time-zone`, `mode-settings`).

### Decision: Keep the existing `ansible-navigator` YAML root wrapper

The YAML generator currently wraps all settings under:

```yaml
ansible-navigator:
  ...
```

This wrapper will be preserved while new fields are added underneath.

## Risks / Trade-offs

- The upstream schema uses several nested objects and hyphenated keys; missing conversion logic can cause silent configuration mismatches.
- The two provisioners currently have separate YAML conversion functions. This proposal does not mandate refactoring to share code; it only requires parity of behavior.
