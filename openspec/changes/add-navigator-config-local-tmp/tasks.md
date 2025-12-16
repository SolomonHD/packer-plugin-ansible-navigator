## 1. Implementation

- [ ] Add `local_tmp` to the typed navigator config structs under `navigator_config.ansible_config.defaults` (remote + local provisioners).
- [ ] Ensure HCL2 schema supports `navigator_config.ansible_config.defaults.local_tmp` (update `go:generate` type list if required, then run `make generate`).
- [ ] Update ansible.cfg generation logic to write `local_tmp` into `[defaults]` when configured.
- [ ] Ensure the generated `ansible-navigator.yml` references the generated ansible.cfg via `ansible.config.path` and remains schema-compliant (no `defaults` under `ansible.config`).
- [ ] Add unit tests covering:
  - `local_tmp` appears in generated ansible.cfg when set
  - `local_tmp` is omitted when unset
  - generated YAML remains schema-compliant
- [ ] Update docs/examples to demonstrate configuring both `remote_tmp` and `local_tmp`.

## 2. Validation

- [ ] Run `make generate`.
- [ ] Run `go test ./...`.
- [ ] Run `make plugin-check`.

## 3. OpenSpec Maintenance

- [ ] Ensure `openspec validate add-navigator-config-local-tmp --strict` passes.

