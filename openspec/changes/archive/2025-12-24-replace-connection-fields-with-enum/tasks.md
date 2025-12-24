# Tasks: Replace connection_mode with Enum (Breaking Change)

## Phase 1: Remove Legacy Fields

- [x] Remove `UseProxy config.Trilean` field from [`Config`](../../../provisioner/ansible-navigator/provisioner.go) struct (line ~495)
- [x] Remove `SSHTunnelMode bool` field from [`Config`](../../../provisioner/ansible-navigator/provisioner.go) struct (line ~535)

## Phase 2: Add New Field

- [x] Add `ConnectionMode string` field to [`Config`](../../../provisioner/ansible-navigator/provisioner.go) struct with:
  - `mapstructure:"connection_mode"` tag
  - Comprehensive documentation comment explaining valid values and use cases
  - Position after the `WinRMUseHTTP` field (line ~510)

## Phase 3: Update Validation

- [x] Replace SSH tunnel validation section in [`Validate()`](../../../provisioner/ansible-navigator/provisioner.go:560) (lines ~571-607) with new validation logic:
  - Apply default: if `ConnectionMode == ""`, set to `"proxy"`
  - Validate enum: check `ConnectionMode` is one of `["proxy", "ssh_tunnel", "direct"]`
  - Conditionally validate bastion fields when `ConnectionMode == "ssh_tunnel"`
- [x] Remove validation for mutual exclusivity between `use_proxy` and `ssh_tunnel_mode` (no longer exists)

## Phase 4: Update Prepare() Defaults

- [x] Add default `connection_mode` assignment in [`Prepare()`](../../../provisioner/ansible-navigator/provisioner.go:739):
  - After line ~795 (bastion port default), add: if `p.config.ConnectionMode == ""` { `p.config.ConnectionMode = "proxy"` }

## Phase 5: Update Provision() Connection Logic

- [x] Replace fragmented connection logic in [`Provision()`](../../../provisioner/ansible-navigator/provisioner.go:1351) (lines ~1361-1537) with a clean switch statement on `connection_mode`:
  - Case `"ssh_tunnel"`: Establish SSH tunnel (existing logic from lines ~1381-1475)
  - Case `"direct"`: Use direct connection (existing logic from lines ~1499-1537)
  - Case `"proxy"`: Set up proxy adapter (existing logic from lines ~1476-1497)
  - Default: Return error (should not reach due to validation)
- [x] Remove fallback logic that checks `UseProxy.False()` and sets `UseProxy = TriTrue` (lines ~1361-1376)

## Phase 6: Update createInventoryFile() Logic

- [x] Replace `UseProxy.False()` check in [`createInventoryFile()`](../../../provisioner/ansible-navigator/provisioner.go:1289) (line ~1304) with:
  - `p.config.ConnectionMode == "direct"`
- [x] Replace `!p.config.UseProxy.False()` check (line ~1313) with:
  - `p.config.ConnectionMode == "proxy" || p.config.ConnectionMode == "ssh_tunnel"`

## Phase 7: Update createCmdArgs() Logic

- [x] Replace `UseProxy.False()` check in [`createCmdArgs()`](../../../provisioner/ansible-navigator/provisioner.go:1622) (line ~1647) with:
  - `p.config.ConnectionMode == "direct"`

## Phase 8: Regenerate HCL2 Specs

- [x] Verify `ConnectionMode` is included in `go:generate` directive (line 6)
- [x] Run `make generate` to regenerate [`provisioner.hcl2spec.go`](../../../provisioner/ansible-navigator/provisioner.hcl2spec.go)
- [x] Verify `connection_mode` field appears in generated HCL2 spec
- [x] Verify `use_proxy` and `ssh_tunnel_mode` fields are removed from generated spec

## Phase 9: Verification

- [x] Run `go build ./...` to verify code compiles
- [x] Run `go test ./...` to verify tests pass (update tests if they reference removed fields)
- [x] Run `make plugin-check` to verify plugin SDK compliance
- [ ] Manually test all three connection modes:
  - `connection_mode = "proxy"` with Docker builder
  - `connection_mode = "ssh_tunnel"` with bastion configuration
  - `connection_mode = "direct"` with direct SSH access
- [ ] Verify validation errors for:
  - Invalid `connection_mode` value (e.g., `"invalid"`)
  - Missing bastion fields when `connection_mode = "ssh_tunnel"`

## Phase 10: Documentation Updates

- [x] Update [`docs/CONFIGURATION.md`](../../../docs/CONFIGURATION.md) to document `connection_mode` field
  - Replaced all references to `use_proxy` and `ssh_tunnel_mode` with `connection_mode`
  - Updated connection mode section with clear descriptions of `"proxy"`, `"ssh_tunnel"`, and `"direct"`
  - Updated all examples to use `connection_mode` syntax
  - Added new Example 4 showing direct connection mode
- [x] Update [`docs/EXAMPLES.md`](../../../docs/EXAMPLES.md) with examples for all three modes
  - Added Example 7: SSH tunnel through bastion
  - Added Example 8: Direct connection (no proxy)
  - Added Example 9: Proxy adapter (default)
- [x] Add migration guide showing old vs new configuration
  - Added comprehensive migration examples to CHANGELOG.md
  - Shows before/after for all three connection modes
- [x] Update [`CHANGELOG.md`](../../../CHANGELOG.md) with breaking change notice
  - Added "Unreleased" section with detailed breaking change documentation
  - Includes migration examples for all three modes
  - Documents rationale referencing project policy

## Notes

**Breaking Change**: This change removes `use_proxy` and `ssh_tunnel_mode` fields entirely. Existing configurations will fail to parse.

**Critical**: Run `make generate` after any changes to the `Config` struct to ensure HCL2 spec is synchronized.

**Testing**: Focus testing on the switch statement logic in [`Provision()`](../../../provisioner/ansible-navigator/provisioner.go:1351) to ensure all three connection modes work as expected.
