## 1. Configuration surface changes

- [ ] Add `skip_version_check` configuration field to `ansible-navigator-local` provisioner config (parity with remote).
- [ ] Regenerate HCL2 spec files so the new field is accepted in HCL (`make generate`).

## 2. Warning behavior

- [ ] Detect when `skip_version_check = true` AND `version_check_timeout` is explicitly set by the user (distinct from defaulting).
- [ ] Emit a non-fatal warning to Packer UI output for the remote provisioner when both are set.
- [ ] Emit a non-fatal warning to Packer UI output for the local provisioner when both are set.
- [ ] Ensure no warning when `skip_version_check = false`.
- [ ] Ensure no warning when `skip_version_check = true` but `version_check_timeout` was not explicitly configured.

## 3. Tests

- [ ] Add/extend unit tests to cover the warning emission behavior for both provisioners.
- [ ] Add/extend unit tests to ensure the warning does not fail validation/prepare.

## 4. Validation

- [ ] Run `openspec validate warn-on-skip-version-check-timeout --strict` and fix any OpenSpec issues.
