# Tasks: Update SSH Tunnel Documentation

## Documentation Updates

- [x] Add "SSH Tunnel Mode (Bastion/Jump Host)" section to [`docs/CONFIGURATION.md`](../../../docs/CONFIGURATION.md) after SSH proxy documentation
  - [x] Add "When to Use SSH Tunnel Mode" subsection with decision criteria
  - [x] Add configuration fields table with all `ssh_tunnel_mode` and `bastion_*` fields
  - [x] Add working example: AWS EC2 via bastion with key authentication
  - [x] Add working example: Lab environment with password authentication
  - [x] Add ASCII architecture comparison diagrams (proxy adapter vs SSH tunnel)
  - [x] Add "When SSH Tunnel Mode is Required" subsection with specific scenarios
  - [x] Document mutual exclusivity with `use_proxy`
  - [x] Note HOME expansion for file paths using `~`

- [x] Add "SSH Tunnel Connection Issues" section to [`docs/TROUBLESHOOTING.md`](../../../docs/TROUBLESHOOTING.md) after "Authentication failures"
  - [x] Add "SSH Tunnel Fails to Establish" subsection with manual verification commands
  - [x] Add common error messages table with causes and solutions
  - [x] Add "WSL2 Execution Environment Issues" subsection with resolution steps
  - [x] Add "Key File Permissions" subsection with permission commands

## Validation

- [x] Verify all HCL examples are syntactically correct
- [x] Verify architecture diagrams render correctly in markdown viewers
- [x] Verify all internal documentation links work (relative paths)
- [x] Verify configuration field table is complete and accurate
- [x] Verify security notes are present (password variables, key permissions)
- [x] Verify technical terms are used consistently (bastion vs jump host)

## OpenSpec Compliance

- [x] Validate proposal against requirements: `openspec validate update-ssh-tunnel-documentation --strict`
- [x] Resolve any validation errors by updating spec files only
