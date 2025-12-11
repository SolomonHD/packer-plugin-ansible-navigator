## 1. Schema and config surface

- [x] 1.1 Remove legacy playbook-only configuration fields from both provisioners (e.g., `playbook_file`, `playbook_files`, multi-path upload helpers). (validated via `go test ./...`)
- [x] 1.2 Remove legacy dependency configuration forms (e.g., inline `collections`, legacy galaxy file aliases) and standardize on `requirements_file`. (validated via `go test ./...`)
- [x] 1.3 Remove legacy inventory compatibility fields (e.g., `inventory_groups`) and any remaining legacy upload-only knobs not required for `play` execution. (validated via `go test ./...`)

## 2. Validation and error semantics

- [x] 2.1 Update validation so **at least one** `play { ... }` block is required. (validated via `go test ./...`)
- [x] 2.2 Ensure validation and error messages do not reference deprecated terminology. (validated via `go test ./...`)

## 3. HCL2 spec generation

- [x] 3.1 Run `make generate` to regenerate all `*.hcl2spec.go` after config struct changes. (validated via `make generate` + `go test ./...`)
- [x] 3.2 Verify generated specs expose only the supported fields and the `play` block. (validated via grep on generated `*hcl2spec.go`)

## 4. Tests

- [x] 4.1 Update unit tests for both provisioners to reflect the new schema (no playbook-only path). (validated via `go test ./...`)
- [x] 4.2 Add coverage for: required `play` blocks, ordered execution, and `requirements_file` dependency installation. (validated via `go test ./...`)

## 5. Documentation

- [x] 5.1 Update README and `docs/*` to remove legacy options and deprecated terminology. (validated via grep for removed fields)
- [x] 5.2 Provide a single canonical example showing:
  - multiple `play { ... }` blocks
  - optional `requirements_file`
  - optional `ansible_cfg`

## 6. OpenSpec maintenance

- [x] 6.1 Apply and archive this change after implementation, updating `openspec/specs/*` accordingly. (validated via `openspec validate --strict` and `openspec list`)
