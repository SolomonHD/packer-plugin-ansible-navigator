# Change: Restructure Bastion Configuration as Nested HCL Block

## Why

Currently, bastion configuration uses flat top-level fields (`bastion_host`, `bastion_port`, `bastion_user`, `bastion_private_key_file`, `bastion_password`). This flat structure doesn't clearly group related settings and doesn't match the semantic structure where bastion is a cohesive set of properties.

Restructuring to a nested `bastion {}` block will:

- Improve visual scanning and readability
- Match the semantic concept of "bastion configuration" as a unit
- Allow for explicit enable/disable flags
- Follow established HCL patterns for related configuration groups

## What Changes

- **Add** `BastionConfig` struct with fields: `Enabled`, `Host`, `Port`, `User`, `PrivateKeyFile`, `Password`
- **Add** `Bastion *BastionConfig` field to the `Config` struct
- **Mark as deprecated** existing flat `Bastion*` fields (keeping them for backward compatibility)
- **Add** migration logic that moves flat fields to nested block in `Prepare()`
- **Update** validation logic to check `Bastion` struct instead of flat fields
- **Update** all code references to use `p.config.Bastion.*` instead of `p.config.Bastion*`
- **Update** HCL2 spec generation to include `BastionConfig` type

## Impact

- **Affected specs:**
  - [`ssh-tunnel-configuration`](../../specs/ssh-tunnel-configuration/spec.md) - bastion field requirements
  - [`connection-mode-configuration`](../../specs/connection-mode-configuration/spec.md) - validation when `connection_mode = "ssh_tunnel"`
  
- **Affected code:**
  - [`provisioner/ansible-navigator/provisioner.go`](../../../provisioner/ansible-navigator/provisioner.go) - struct definition, validation, `Prepare()`, `setupSSHTunnel()`
  - Generated HCL2 spec files

- **Backward compatibility:** Fully maintained. Legacy flat fields are migrated automatically with deprecation warnings. New block syntax takes precedence when both are present.

## Dependencies

- **Depends on:** Change 02 (`replace-connection-fields-with-enum`) - validation checks `connection_mode == "ssh_tunnel"`
- **Related to:** Change 01 (`fix-port-type-coercion`) - independent but should be in same release
