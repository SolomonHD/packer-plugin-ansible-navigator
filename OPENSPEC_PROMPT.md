# OpenSpec Change Prompt

## Context
The repo contains a Go-based Packer provisioner plugin at `packer/plugins/packer-plugin-ansible-navigator/` that generates an `ansible-navigator.yml` file from a typed `navigator_config { ... }` HCL block. Today, the plugin’s typed schema only covers a small subset of the `ansible-navigator.yml` configuration surface (mode, a limited execution-environment block, logging, playbook-artifact, collection-doc-cache, and a minimal ansible.config.path mapping). Users attempting to set valid ansible-navigator config fields (e.g., `container-options`) receive “Unsupported argument” errors because the fields are not modeled in the plugin’s HCL2 spec.

## Goal
Provide full configuration parity with **ansible-navigator v3.x** `ansible-navigator.yml` (current stable) so that users can express *any documented ansible-navigator YAML option* via `navigator_config` in Packer HCL, and the plugin will generate a correct `ansible-navigator.yml` reflecting that configuration.

## Scope

**In scope:**
- Expand the plugin’s typed `NavigatorConfig` model to cover the full ansible-navigator v3.x `ansible-navigator.yml` schema (all documented keys, nested objects/blocks, and lists).
- Ensure HCL naming follows the existing convention: underscores in HCL / Go (`execution_environment`, `container_options`, etc.) with conversion to the correct hyphenated YAML keys when generating `ansible-navigator.yml`.
- Update YAML generation to emit all supported keys correctly (including any nested key renames like `pull.policy`, and any list/map shapes required by the v3.x schema).
- Update the `//go:generate packer-sdc mapstructure-to-hcl2 ...` type list so HCL2 spec generation includes all newly-added structs.
- Regenerate all `*.hcl2spec.go` files and ensure `packer validate` accepts the expanded configuration.
- Add/extend tests that validate:
  - HCL decoding into the new structs works.
  - YAML output contains the expected keys/structure for representative configurations.
  - Backward compatibility: existing configs that worked before still parse and generate equivalent YAML.

**Out of scope:**
- Changing Packer’s plugin registration behavior or provisioner names.
- Changing runtime behavior unrelated to configuration modeling (except where necessary to interpret new config fields).
- Supporting ansible-navigator schema versions outside v3.x.

## Desired Behavior
- A user can configure any ansible-navigator v3.x `ansible-navigator.yml` option through `navigator_config { ... }` in Packer HCL.
- The plugin generates a valid `ansible-navigator.yml` that matches the configured values (including correct YAML key names and nesting).
- Unsupported-argument errors for documented v3.x config keys are eliminated.
- Existing automatic defaults (e.g., safe temp directories and collections mounting when execution environments are enabled) continue to work, and do not override user-specified values.

## Constraints & Assumptions
- Assumption: “Parity” means parity with **documented** ansible-navigator v3.x `ansible-navigator.yml` options (not undocumented/experimental keys).
- Constraint: Avoid non-serializable dynamic config patterns (no `map[string]interface{}` that relies on `cty.DynamicPseudoType`); use typed structs and generated HCL2 specs.
- Constraint: Preserve backward compatibility for existing `navigator_config` usage and for non-EE scenarios.
- Constraint: YAML key conversion must match ansible-navigator’s expected schema (hyphenated keys where required).
- Assumption: The canonical list of v3.x config keys should be derived from ansible-navigator’s official documentation/schema, then encoded into Go structs.

## Acceptance Criteria
- [ ] For every documented ansible-navigator v3.x `ansible-navigator.yml` configuration key, there is a corresponding HCL attribute/block under `navigator_config` that decodes without error.
- [ ] The plugin’s generated YAML includes correct key names, nesting, and value types for representative configs that touch each major section of the schema.
- [ ] Existing `navigator_config` examples continue to work unchanged and generate equivalent YAML.
- [ ] `make generate` (or equivalent repo target) regenerates HCL2 specs successfully and the repo builds.
- [ ] A minimal Packer template using newly-added options passes `packer validate`.
