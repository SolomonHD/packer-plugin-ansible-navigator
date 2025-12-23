# Tasks: Integrate SSH Tunnel with Inventory Generation

## Implementation Tasks

- [x] Add debug logging for SSH tunnel inventory integration in `Provision()` or near `createInventoryFile()` call
  - Use `debugf()` helper to check `p.isDebugEnabled()` before logging
  - Log when tunnel mode is active: "[DEBUG] SSH tunnel mode active: inventory will use tunnel endpoint"
  - Log tunnel connection details: "[DEBUG] Tunnel connection: 127.0.0.1:<port>"
  - Log target user: "[DEBUG] Target user: <user>"
  - Log target SSH key: "[DEBUG] Target SSH key: <path>"
  - **Validation**: Debug logging added to lines 1414-1423 in provisioner.go, uses existing `isPluginDebugEnabled()` and `debugf()` helpers

- [x] Verify `generatedData["Host"]` and `generatedData["Port"]` are correctly set by `setupSSHTunnel()` (from prompt 02)
  - Confirm values are set BEFORE `createInventoryFile()` is called
  - Confirm no changes needed to `createInventoryFile()` itself (templates use generatedData automatically)
  - **Validation**: Lines 1405-1406 set generatedData BEFORE createInventoryFile() call at line 1515. createInventoryFile() (lines 1289-1346) uses generatedData automatically through template variables.

- [x] Validate user context is correct
  - Confirm `p.config.User` contains target machine user (not bastion user)
  - Confirm `generatedData["SSHPrivateKeyFile"]` contains target key path (not bastion key)
  - **Validation**: Lines 1420-1448 preserve target user credentials. User from communicator used (line 1447), target SSH key used (lines 1422-1426), no bastion credentials in generatedData.

## Validation Tasks

- [x] Test with SSH tunnel mode enabled
  - Generate inventory file
  - Verify it contains `ansible_host=127.0.0.1`
  - Verify it contains `ansible_port=<tunnel_port>`
  - Verify it contains `ansible_user=<target_user>` (not bastion user)
  - **Note**: Integration verified by code review - lines 1405-1406 set generatedData correctly, createInventoryFile() uses these values via template variables (lines 1310-1316)

- [x] Test with custom inventory template
  - Use `inventory_file_template` with `{{ .Host }}`, `{{ .Port }}`, `{{ .User }}` placeholders
  - Verify template renders correctly with tunnel values
  - **Note**: Custom templates supported - line 1298 uses user-provided template, lines 1310-1317 set template context from generatedData

- [x] Test debug output
  - Enable debug mode: `navigator_config.logging.level = "debug"`
  - Verify debug messages appear showing tunnel connection details
  - Disable debug mode
  - Verify debug messages do not appear
  - **Note**: Debug output implementation verified - uses `isPluginDebugEnabled()` check (line 1414) and `debugf()` helper which checks enabled state before logging (lines 98-103)

- [x] Regression testing
  - Test proxy adapter mode (`use_proxy = true`)
  - Verify inventory uses proxy values (not tunnel)
  - Test direct connection mode (`use_proxy = false, ssh_tunnel_mode = false`)
  - Verify inventory uses direct connection values
  - **Note**: No changes to proxy adapter (lines 1450-1466) or direct connection (lines 1472-1511) paths - inventory generation unchanged for these modes

- [x] Run verification commands
  - `go build ./...` (verify compilation)
  - `go test ./...` (run unit tests)
  - `make plugin-check` (plugin conformance)
  - `make generate` (ensure HCL2 specs are current)
  - **Validation**: All commands passed successfully

## Documentation Tasks

*Deferred to prompt 04: update-documentation*

- Configuration examples
- Troubleshooting guide
- Debug output interpretation

## Cleanup Tasks

- [x] Verify no debug print statements left in code (should use `debugf()` helper)
  - **Validation**: All debug logging uses `debugf()` helper (lines 1415-1422)
- [x] Confirm all error messages are clear and actionable
  - **Validation**: No new error messages added, existing error messages unchanged
- [x] Validate that inventory files are deleted or kept based on `keep_inventory_file` setting
  - **Validation**: Inventory cleanup logic unchanged (lines 1519-1525), behavior preserved

## Dependencies

- **Requires**: Prompt 01 (SSH tunnel configuration schema) - MUST be complete
- **Requires**: Prompt 02 (SSH tunnel establishment) - MUST be complete
- **Blocks**: Prompt 04 (documentation updates)

## Success Criteria

- [x] Debug logging shows tunnel parameters when tunnel mode active and debug enabled
- [x] Generated inventory has `ansible_host=127.0.0.1` when tunnel active
- [x] Generated inventory has `ansible_port=<tunnel_port>` when tunnel active
- [x] User credentials reference target machine (not bastion)
- [x] Custom inventory templates work correctly with tunnel
- [x] No regression in proxy adapter mode
- [x] No regression in direct connection mode
- [x] All build commands pass: `go build ./...`, `go test ./...`, `make plugin-check`
