## MODIFIED Requirements

### Requirement: Default Navigator Command

The SSH-based provisioner SHALL use ansible-navigator as its default executable, pass `run` as the first argument, and treat the `command` field strictly as the ansible-navigator executable name or path (without embedded arguments).

#### Scenario: Default command construction

- **GIVEN** a Packer configuration using `provisioner "ansible-navigator"`
- **AND** the `command` field is not specified
- **WHEN** the provisioner constructs the command to run ansible-navigator
- **THEN** it SHALL resolve `ansible-navigator` using its effective search path and invoke the resolved executable with `run` as the first argument on the local machine
- **AND** configuration SHALL be controlled ONLY via ANSIBLE_NAVIGATOR_CONFIG environment variable
- **AND** NO mode, EE, or other CLI configuration flags SHALL be passed (all settings come from the config file)
- **AND** legacy CLI configuration flags (`--mode`, `--ee`, `--eei`) SHALL NOT be generated

### Requirement: Remote PATH Control for ansible-navigator

The SSH-based ansible-navigator provisioner SHALL support an `ansible_navigator_path` configuration option that prepends additional directories to PATH when resolving and running ansible-navigator for version checks and play execution.

#### Scenario: PATH prepended for version check

- **GIVEN** a configuration with `ansible_navigator_path = ["~/bin", "/opt/ansible/bin"]`
- **AND** ansible-navigator is installed only under one of these directories
- **WHEN** the provisioner performs `ansible-navigator --version`
- **THEN** each configured directory SHALL be HOME-expanded
- **AND** executable lookup SHALL search those directories in configured order before the inherited process PATH
- **AND** the child PATH SHALL contain that same ordered prefix followed by the inherited PATH
- **AND** the version check SHALL succeed without changing the parent process environment

#### Scenario: PATH prepended for play execution

- **GIVEN** a configuration with one or more `ansible_navigator_path` directories
- **AND** ansible-navigator is installed only under one of those directories
- **WHEN** the provisioner executes one or more plays
- **THEN** it SHALL use the same effective search path and resolved executable as the version check
- **AND** every play SHALL receive that same PATH-prepending behavior

#### Scenario: No change when ansible_navigator_path is unset

- **GIVEN** a configuration without `ansible_navigator_path`
- **WHEN** the provisioner resolves or executes ansible-navigator
- **THEN** it SHALL use the inherited process PATH unchanged
- **AND** existing executable search precedence SHALL be preserved

#### Scenario: Explicit executable path remains authoritative

- **GIVEN** `command` specifies an executable path
- **WHEN** the provisioner prepares the command
- **THEN** it SHALL expand a supported HOME-relative path and resolve a relative path against the preparation working directory
- **AND** it SHALL NOT substitute another installation merely because it appears earlier on PATH
- **AND** the effective child PATH SHALL remain available for Navigator's own subprocesses

#### Scenario: Configured path takes precedence over inherited installation

- **GIVEN** both a configured search directory and inherited PATH contain Navigator installations
- **WHEN** the provisioner resolves the executable name
- **THEN** it SHALL select the first executable in the configured directory order
- **AND** validation and all plays SHALL use that installation

## ADDED Requirements

### Requirement: Resolved Navigator Executable Reuse

The remote provisioner MUST use one resolved executable for its version check and all play invocations, with actionable lookup and shim-resolution failures.

#### Scenario: Supported shim resolves successfully

- **GIVEN** executable lookup selects an asdf, rbenv, or pyenv shim
- **WHEN** the supported manager resolves it to an executable
- **THEN** the version check and all plays MUST use the resolved executable
- **AND** resolution MUST use the same effective environment as execution

#### Scenario: Shim manager hangs or fails

- **GIVEN** a supported shim manager hangs, fails, returns another unresolved shim, or returns an invalid executable
- **WHEN** Navigator is resolved
- **THEN** the provisioner MUST stop within a documented resolution timeout
- **AND** when version checking is enabled, resolution and the version invocation together MUST NOT exceed the configured version-check timeout budget
- **AND** it MUST report the failed resolution stage and a corrective explicit-path configuration
- **AND** it MUST NOT fall back silently to a different executable

#### Scenario: Version check is disabled

- **GIVEN** `skip_version_check = true`
- **WHEN** the provisioner prepares to launch its first play
- **THEN** executable lookup and supported shim resolution MUST still occur before execution
- **AND** only the version invocation MUST be omitted

#### Scenario: Executable is missing or cannot execute

- **GIVEN** the configured executable cannot be found or executed
- **WHEN** the provisioner resolves or launches it
- **THEN** the error MUST distinguish lookup failure from a subprocess failure
- **AND** it MUST identify the configured command and effective search-directory sources without dumping environment values
- **AND** missing Docker or proxy connectivity MUST NOT be reported as the cause of a Navigator lookup failure
