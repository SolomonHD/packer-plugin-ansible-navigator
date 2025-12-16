## 1. Implementation

- [x] Add `local_tmp` to the typed navigator config structs under `navigator_config.ansible_config.defaults` (remote + local provisioners).
  - Validation: added `LocalTmp` field with `mapstructure:"local_tmp"` to both `AnsibleConfigDefaults` structs.
- [x] Ensure HCL2 schema supports `navigator_config.ansible_config.defaults.local_tmp` (update `go:generate` type list if required, then run `make generate`).
  - Validation: `make generate` + confirmed `local_tmp` appears in both generated `provisioner.hcl2spec.go` files.
- [x] Update ansible.cfg generation logic to write `local_tmp` into `[defaults]` when configured.
  - Validation: unit tests assert `local_tmp = ...` is present/absent as expected.
- [x] Ensure the generated `ansible-navigator.yml` references the generated ansible.cfg via `ansible.config.path` and remains schema-compliant (no `defaults` under `ansible.config`).
  - Validation: unit tests assert YAML contains `ansible.config.path` and does not include `defaults`/`ssh_connection` under `ansible.config`.
- [x] Add unit tests covering:
  - `local_tmp` appears in generated ansible.cfg when set
  - `local_tmp` is omitted when unset
  - generated YAML remains schema-compliant
  - Validation: `go test ./...`
- [x] Update docs/examples to demonstrate configuring both `remote_tmp` and `local_tmp`.
  - Validation: updated docs examples to use `navigator_config { ansible_config { defaults { remote_tmp/local_tmp } } }`.

## 2. Validation

- [x] Run `make generate`.
- [x] Run `go test ./...`.
- [x] Run `make plugin-check`.

## 3. OpenSpec Maintenance

- [x] Ensure `openspec validate add-navigator-config-local-tmp --strict` passes.
