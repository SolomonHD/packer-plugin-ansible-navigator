## 1. Configuration surface changes

- [x] Add `skip_version_check` configuration field to `ansible-navigator-local` provisioner config (parity with remote).
- [x] Regenerate HCL2 spec files so the new field is accepted in HCL (`make generate`).
  - Validation: `make generate`

## 2. Warning behavior

- [x] Detect when `skip_version_check = true` AND `version_check_timeout` is explicitly set by the user (distinct from defaulting).
- [x] Emit a non-fatal warning to Packer UI output for the remote provisioner when both are set.
- [x] Emit a non-fatal warning to Packer UI output for the local provisioner when both are set.
- [x] Ensure no warning when `skip_version_check = false`.
- [x] Ensure no warning when `skip_version_check = true` but `version_check_timeout` was not explicitly configured.

## 3. Tests

- [x] Add/extend unit tests to cover the warning emission behavior for both provisioners.
- [x] Add/extend unit tests to ensure the warning does not fail validation/prepare.
  - Validation: `go test ./...`

## 4. Validation

- [x] Run `openspec validate warn-on-skip-version-check-timeout --strict` and fix any OpenSpec issues.
  - Validation: `openspec validate warn-on-skip-version-check-timeout --strict`
