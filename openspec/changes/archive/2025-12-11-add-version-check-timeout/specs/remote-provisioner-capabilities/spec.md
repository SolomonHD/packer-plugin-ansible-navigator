## ADDED Requirements

### Requirement: Version Check Timeout Configuration

The SSH-based ansible-navigator provisioner SHALL support a configurable timeout for the version check operation to prevent indefinite hangs when ansible-navigator cannot be located or is slow to respond.

#### Scenario: Default timeout value

- **GIVEN** a configuration for `provisioner "ansible-navigator"` that does not specify `version_check_timeout`
- **WHEN** the provisioner prepares for execution
- **THEN** it SHALL use a default timeout of 60 seconds for the version check
- **AND** the version check SHALL fail with a timeout error if it exceeds 60 seconds

#### Scenario: Custom timeout value

- **GIVEN** a configuration with `version_check_timeout = "30s"`
- **WHEN** the version check is performed
- **THEN** it SHALL use a 30-second timeout
- **AND** the timeout SHALL be enforced using `context.WithTimeout` and `exec.CommandContext`

#### Scenario: Timeout uses duration format

- **GIVEN** a configuration with `version_check_timeout` set to a value
- **WHEN** the configuration is prepared
- **THEN** the value SHALL be parsed using `time.ParseDuration`
- **AND** valid duration formats SHALL include "60s", "2m", "1m30s", etc.
- **AND** invalid duration formats SHALL cause configuration validation to fail

#### Scenario: Timeout error message

- **GIVEN** the version check command times out after the configured duration
- **WHEN** the timeout occurs
- **THEN** the error message SHALL indicate:
  - The version check timed out
  - The configured timeout value (e.g., "60s")
  - Suggestions including: use `ansible_navigator_path`, use `command` to specify full path, set `skip_version_check = true`, or increase `version_check_timeout`

#### Scenario: Not-found error remains distinct

- **GIVEN** ansible-navigator cannot be found in PATH before the timeout expires
- **WHEN** the command execution fails with a not-found error
- **THEN** the error message SHALL indicate ansible-navigator was not found
- **AND** the error SHALL not be reported as a timeout
- **AND** the error message SHALL mention both PATH and `ansible_navigator_path` as resolution mechanisms

#### Scenario: Skip version check bypasses timeout

- **GIVEN** a configuration with `skip_version_check = true`
- **WHEN** the provisioner prepares for execution
- **THEN** no version check SHALL be performed
- **AND** the `version_check_timeout` configuration SHALL be ignored

#### Scenario: Version check respects modified PATH

- **GIVEN** a configuration with both `ansible_navigator_path` and `version_check_timeout`
- **WHEN** the version check executes
- **THEN** it SHALL use the modified PATH from `ansible_navigator_path`
- **AND** it SHALL enforce the configured timeout
- **AND** both mechanisms SHALL work together correctly

#### Scenario: asdf installation compatibility

- **GIVEN** ansible-navigator is installed via asdf
- **AND** the configuration uses `command = "~/.asdf/shims/ansible-navigator"` or `ansible_navigator_path = ["~/.asdf/shims"]`
- **WHEN** the version check executes
- **THEN** it SHALL succeed within the timeout period if the shim responds correctly
- **OR** it SHALL timeout with a clear error message if the shim hangs or doesn't respond
