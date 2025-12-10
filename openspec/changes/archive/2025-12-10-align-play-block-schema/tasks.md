## 1. HCL2 schema and config alignment

- [x] 1.1 Update the `Config` / `FlatConfig` types for the SSH-based provisioner to wire the multi-play field to a singular `play` block name while remaining a slice in Go.
- [x] 1.2 Regenerate `provisioner.hcl2spec.go` for the SSH-based provisioner and verify the `BlockListSpec` key and `TypeName` are `play`, with no `plays` entry for multi-play configuration.
- [x] 1.3 Repeat steps 1.1–1.2 for the local `ansible-navigator-local` provisioner so both provisioners share consistent `play` block semantics.
- [x] 1.4 Ensure the singular `play` block configuration is the only schema surface for multi-play configuration (no secondary alias or hidden `plays` block).

## 2. Validation and error behavior

- [x] 2.1 Update configuration validation for both provisioners so mutual exclusivity errors reference `playbook_file` and `play` blocks (singular) and no longer mention `plays`.
- [ ] 2.2 Introduce explicit guards that detect use of legacy `plays { ... }` blocks and fail fast with a clear migration error pointing users to repeated `play { ... }` blocks.
- [ ] 2.3 Ensure array-style syntax such as `plays = [...]` is rejected with an error that clearly states that HCL2 block syntax (`play { ... }`) is required.
- [x] 2.4 Propagate all validation failures through both the Packer UI and returned Go `error` values so CI/CD systems can detect misconfiguration.

## 3. Tests

- [ ] 3.1 Add or update unit tests for the SSH-based provisioner covering:
  - Multiple `play { ... }` blocks executing successfully in declaration order.
  - Mutual exclusivity errors for `playbook_file` plus `play` blocks.
  - Hard-error behavior when `plays { ... }` blocks are present.
  - Rejection of `plays = [...]` syntax with a migration hint.
- [ ] 3.2 Add or update corresponding tests for the local provisioner so behavior and error messages are consistent across both.

## 4. Documentation

- [x] 4.1 Sweep `README.md` examples and replace any remaining `plays`, `plays { ... }`, or `plays = [...]` usages with repeated `play { ... }` blocks. (Updated top-level docs README and JSON logging examples.)
- [x] 4.2 Update configuration and examples under `docs/` (including configuration, unified plays, examples, and troubleshooting guides) so they exclusively document singular `play` blocks for multi-play configuration. (Confirmed `CONFIGURATION.md`, `UNIFIED_PLAYS.md`, `EXAMPLES.md`, and `TROUBLESHOOTING.md` use only `play` blocks.)
- [x] 4.3 If a migration section exists (for v2 → v3 or similar), ensure it clearly states that `plays` blocks are no longer supported and must be rewritten as `play` blocks. (Migration note in unified plays docs now calls out `plays { }` / `plays = [...]` as unsupported.)
- [x] 4.4 Verify that AGENT- and contributor-facing docs (such as `AGENTS.md`) reflect the singular `play` block semantics. (Updated `AGENTS.md` to describe `play` blocks and remove `plays[]` from the config schema.)

## 5. Tooling and validation

- [x] 5.1 Run `go build ./...` and `go test ./...` to confirm the codebase compiles and tests pass after schema and validation changes.
- [x] 5.2 Run any existing plugin checks (for example, `make plugin-check`) to validate compatibility with the Packer Plugin SDK.
- [ ] 5.3 Manually verify with `packer validate` and `packer build` using templates that exercise:
  - Only `playbook_file`.
  - Only multiple `play { ... }` blocks.
  - Legacy `plays { ... }` and `plays = [...]` syntax (to confirm they now fail with clear migration errors).
  - _Note: Not yet executed in this session; manual Packer templates and runs still required._
- [x] 5.4 Run `openspec validate align-play-block-schema --strict` and resolve any spec issues before finalizing the change.
