## 1. Implementation

- [ ] Remove `work_dir` from the remote provisioner config struct and any usage.
- [ ] Remove `work_dir` from the local provisioner config struct and any usage.
- [ ] Ensure HCL parsing rejects `work_dir` for both provisioners (schema removal, not runtime validation).
- [ ] Regenerate HCL2 specs (`make generate`) and confirm `work_dir` is absent from generated specs.
- [ ] Remove `work_dir` from documentation and examples (e.g., configuration reference).

## 2. Validation

- [ ] Run unit tests (`go test ./...`).
- [ ] Run plugin conformance checks (`make plugin-check`).
- [ ] Run `openspec validate remove-work-dir --strict` and ensure it passes.

