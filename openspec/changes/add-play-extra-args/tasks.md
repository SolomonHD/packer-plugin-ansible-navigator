## 1. Implementation

- [ ] 1.1 Add `extra_args` to the play configuration struct(s) for both provisioners (schema: list(string)).
- [ ] 1.2 Ensure play-level `extra_args` are appended verbatim to the ansible-navigator command for that play.
- [ ] 1.3 Define and enforce deterministic argument ordering for both provisioners:
  - `ansible-navigator`, `run`
  - enforced `--mode` behavior (when configured)
  - play-level `extra_args`
  - plugin-generated inventory/extra-vars/etc
  - play target (playbook path or role FQDN)
- [ ] 1.4 Regenerate HCL2 spec files (`make generate`) and verify `play.extra_args` appears in both generated specs.
- [ ] 1.5 Add/adjust unit tests to cover:
  - `extra_args` wiring for both provisioners
  - ordering relative to `--mode` and provisioner-generated args
- [ ] 1.6 Update docs (`docs/CONFIGURATION.md`) with at least one example showing `play.extra_args` usage.

## 2. Validation

- [ ] 2.1 Run `go test ./...`.
- [ ] 2.2 Run `make plugin-check`.

## 3. OpenSpec Maintenance

- [ ] 3.1 After implementation, apply the deltas to the base specs under `openspec/specs/` and validate with `openspec validate --strict`.

