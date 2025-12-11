## ADDED Requirements

### Requirement: HCL2 Spec Regeneration

After ANY modification to a Config struct (adding, removing, or changing fields with `mapstructure` tags), the developer MUST regenerate HCL2 spec files by running `make generate`.

#### Scenario: Config field added

- **WHEN** a new field is added to a Config struct with a `mapstructure` tag
- **THEN** `make generate` is run before committing
- **AND** the new field appears in the corresponding `FlatConfig` struct in `*.hcl2spec.go`
- **AND** the new field appears in the `HCL2Spec()` return map

#### Scenario: Build verification after regeneration

- **WHEN** `make generate` completes successfully
- **THEN** `go build ./...` passes without errors
- **AND** `make plugin-check` passes Packer SDK validation
