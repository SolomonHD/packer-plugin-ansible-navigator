## 1. Implementation

- [x] Remove `work_dir` from the remote provisioner config struct and any usage.
- [x] Remove `work_dir` from the local provisioner config struct and any usage.
- [x] Ensure HCL parsing rejects `work_dir` for both provisioners (schema removal, not runtime validation).
- [x] Regenerate HCL2 specs (`make generate`) and confirm `work_dir` is absent from generated specs.
- [x] Remove `work_dir` from documentation and examples (e.g., configuration reference).

## 2. Validation

- [x] Run unit tests (`go test ./...`).
- [x] Run plugin conformance checks (`make plugin-check`).
- [x] Run `openspec validate remove-work-dir --strict` and ensure it passes.
