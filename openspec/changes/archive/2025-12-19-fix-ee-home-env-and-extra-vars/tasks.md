## 1. Implementation

- [x] 1.1 Update EE default injection to set `HOME`, `XDG_CACHE_HOME`, and `XDG_CONFIG_HOME` when `navigator_config.execution_environment.enabled=true` and the user did not provide or pass-through those values.
- [x] 1.2 Ensure EE defaults do not override any user-provided env var values (including cases where the user uses `environment_variables.pass`).
- [x] 1.3 Update extra-vars construction so provisioner-generated vars are encoded as a single JSON object passed via one `-e`/`--extra-vars` argument.
- [x] 1.4 Ensure argument ordering guarantees the play target remains last and cannot be shifted by extra-vars.
- [x] 1.5 Add/adjust unit tests:
  - EE home-related env defaults (HOME/XDG) injected only when unset/unpassed
  - Extra-vars JSON construction is stable and cannot produce a standalone `-e`
  - Validation: `go test ./...`
- [x] 1.6 Update documentation with a minimal EE + requirements.yml + role FQDN example.
- [x] 1.7 Run validation checks (tests + plugin-check) in the implementation task (not in this proposal workflow).
  - Validation: `go test ./...`
  - Validation: `make plugin-check`
