## Context

The plugin supports `navigator_config { ... }`, which generates an `ansible-navigator.yml` file and points ansible-navigator to it via `ANSIBLE_NAVIGATOR_CONFIG`.

For Ansible configuration, the project MUST remain compliant with ansible-navigator’s schema for `ansible.config`:

- `ansible.config` in YAML may only contain `help`, `path`, and/or `cmdline`.
- Any Ansible settings like `[defaults]`, `[ssh_connection]`, etc. MUST be expressed via an ansible.cfg (INI) file, referenced from YAML via `ansible.config.path`.

Existing specs already require:

- mutual exclusivity between `ansible_config.config` (path override) and nested blocks like `ansible_config.defaults` / `ansible_config.ssh_connection`
- generation of ansible.cfg from nested blocks

This change extends the *set of supported nested blocks*.

## Goals

- Add HCL blocks for additional ansible.cfg sections under `navigator_config.ansible_config`.
- Generate an ansible.cfg that includes the configured sections.
- Preserve existing schema compliance and mutual exclusivity rules.
- Keep the configuration surface typed and gRPC-serializable (no `cty.DynamicPseudoType`).

## Non-Goals

- Perfect 1:1 modeling of every single ansible.cfg option from Ansible core.
- Semantic validation of option values.

## Decisions

### Decision: Model “major sections” as typed nested structs

Add explicit struct types representing major ansible.cfg sections, exposed as nested HCL blocks:

- `defaults`
- `ssh_connection`
- `privilege_escalation`
- `persistent_connection`
- `inventory`
- `paramiko_connection`
- `colors`
- `diff`
- `galaxy`

Rationale:

- Provides a discoverable HCL surface.
- Ensures generated HCL2 spec is concrete and RPC-serializable.
- Prevents the “unsupported argument” failure mode for these common sections.

### Decision: Limit fields to commonly used options initially

Within each section, include only a practical subset of commonly used fields (booleans, ints, strings, lists) needed for typical execution environment and SSH tuning.

Rationale:

- Full enumeration is large and hard to maintain.
- The project can iterate by adding fields as real use cases arise.

### Decision: INI serialization is the canonical output for these sections

The nested sections are rendered to ansible.cfg INI under their respective section names (e.g. `[privilege_escalation]`). Values are serialized deterministically.

Rationale:

- Aligns with Ansible’s expectations.
- Avoids schema-invalid YAML under `ansible.config`.

## Risks / Trade-offs

- Users may still need an external ansible.cfg for rare options that are not yet modeled.
- Adding many fields increases maintenance burden; this change intentionally focuses on common options first.

## Migration / Compatibility Notes

- Existing configurations using `ansible_config.defaults` and/or `ansible_config.ssh_connection` continue to work.
- This change is additive: new blocks become available; existing HCL does not need updates.
