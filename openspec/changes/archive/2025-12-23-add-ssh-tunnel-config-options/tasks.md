# Tasks: Add SSH Tunnel Configuration Options

## Implementation Tasks

- [x] Add new configuration fields to [`Config`](../../provisioner/ansible-navigator/provisioner.go) struct with proper mapstructure tags:
  - `SSHTunnelMode bool` (`mapstructure:"ssh_tunnel_mode"`)
  - `BastionHost string` (`mapstructure:"bastion_host"`)
  - `BastionPort int` (`mapstructure:"bastion_port"`)
  - `BastionUser string` (`mapstructure:"bastion_user"`)
  - `BastionPrivateKeyFile string` (`mapstructure:"bastion_private_key_file"`)
  - `BastionPassword string` (`mapstructure:"bastion_password"`)

- [x] Update [`go:generate`](../../provisioner/ansible-navigator/provisioner.go) directive to include new Config fields
  - Note: Config was already included in the directive

- [x] Add validation logic to [`Config.Validate()`](../../provisioner/ansible-navigator/provisioner.go):
  - Enforce mutual exclusivity between `ssh_tunnel_mode` and `use_proxy`
  - Require `bastion_host` when `ssh_tunnel_mode = true`
  - Require `bastion_user` when `ssh_tunnel_mode = true`
  - Require either `bastion_private_key_file` OR `bastion_password` when `ssh_tunnel_mode = true`
  - Validate `bastion_port` is between 1 and 65535
  - Validate `bastion_private_key_file` exists if specified (use existing `validateFileConfig()` pattern)

- [x] Add HOME expansion for `bastion_private_key_file` in [`Provisioner.Prepare()`](../../provisioner/ansible-navigator/provisioner.go) using `expandUserPath()`

- [x] Run `make generate` to update `.hcl2spec.go` files

## Verification Tasks

- [x] Run `go build ./...` to verify compilation

- [x] Run `go test ./...` to ensure existing tests pass

- [x] Run `make plugin-check` to verify plugin conformance

- [x] Test validation: Configuration with `ssh_tunnel_mode = true` AND all required fields validates successfully
  - Validated with both key file and password authentication

- [x] Test validation: Configuration with `ssh_tunnel_mode = true` AND `use_proxy = true` fails with clear error message

- [x] Test validation: Configuration with `ssh_tunnel_mode = true` but missing `bastion_host` fails validation

- [x] Test validation: Configuration with `ssh_tunnel_mode = true` but missing `bastion_user` fails validation

- [x] Test validation: Configuration with `ssh_tunnel_mode = true` but missing both key and password fails validation

- [x] Test validation: Configuration with `bastion_port = 99999` fails with port range error
  - Also tested with bastion_port = 0

- [x] Test validation: Configuration with `bastion_private_key_file = "/nonexistent"` fails with file not found error

## Dependencies

None - this is the first task in the SSH tunnel feature sequence.
