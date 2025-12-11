# remote-provisioner-capabilities Specification Deltas

## MODIFIED Requirements

### Requirement: Play-Based Execution

The SSH-based provisioner SHALL execute provisioning via one or more ordered `play { ... }` blocks. Each play SHALL specify a `target` and MAY specify play-level settings. The provisioner SHALL process plays in declaration order.

#### Scenario: At least one play is required

- **GIVEN** a configuration using `provisioner "ansible-navigator"`
- **AND** no `play { ... }` blocks are defined
- **WHEN** the configuration is validated
- **THEN** validation SHALL fail
- **AND** the error message SHALL state that at least one `play` block must be defined

#### Scenario: Playbook target execution

- **GIVEN** a configuration with a `play` block whose `target` ends in `.yml` or `.yaml`
- **WHEN** the provisioner executes
- **THEN** it SHALL run that playbook against the target via SSH

#### Scenario: Role FQDN target execution

- **GIVEN** a configuration with a `play` block whose `target` does not end in `.yml` or `.yaml`
- **WHEN** the provisioner executes
- **THEN** it SHALL treat the target as a role FQDN
- **AND** it SHALL generate a temporary playbook and execute it via SSH

#### Scenario: Ordered execution

- **GIVEN** a configuration with multiple `play { ... }` blocks
- **WHEN** the provisioner executes
- **THEN** it SHALL execute each play in declaration order

### Requirement: Remote HOME Expansion for Path-Like Configuration

The SSH-based provisioner SHALL expand HOME-relative (`~`) paths for supported path-like configuration fields on the local side before validation and execution.

#### Scenario: Expand tilde for supported fields

- **GIVEN** a configuration for `provisioner "ansible-navigator"`
- **AND** one or more supported path-like fields are set using `~` or `~/subdir` (for example: `command`, `ansible_navigator_path` entries, `requirements_file`, `work_dir`, play `target` when it is a playbook path, and play `vars_files` entries)
- **WHEN** the configuration is prepared or validated
- **THEN** each value SHALL be expanded by replacing the leading `~` with the user's HOME directory

## RENAMED Requirements

- FROM: `### Requirement: Collections Management`
- TO: `### Requirement: Dependency Management (requirements_file)`

## MODIFIED Requirements

### Requirement: Dependency Management (requirements_file)

The SSH-based provisioner SHALL support dependency installation via an optional `requirements_file` that can define both roles and collections.

#### Scenario: requirements_file installs roles and collections

- **GIVEN** a configuration with `requirements_file = "requirements.yml"`
- **WHEN** the provisioner executes
- **THEN** it SHALL install roles and collections from that file before executing any plays

#### Scenario: requirements_file omitted

- **GIVEN** a configuration with no `requirements_file`
- **WHEN** the provisioner executes
- **THEN** it SHALL proceed without performing dependency installation

## REMOVED Requirements

### Requirement: Play block naming and validation (remote)

**Reason**: The supported contract is `play` blocks only; active specs do not enumerate deprecated terminology.
