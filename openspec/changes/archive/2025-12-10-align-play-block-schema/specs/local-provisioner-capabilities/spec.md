# local-provisioner-capabilities Spec Delta

## MODIFIED Requirements

### Requirement: Play-Based Execution

The local provisioner SHALL support both traditional playbook files and modern collection-based plays through mutually exclusive configuration options. The play configuration MUST use HCL2 block syntax with the singular block name `play` (repeated `play { }` blocks), following standard HCL idioms for repeatable blocks. Legacy plural or array-based configuration forms for plays MUST NOT be accepted.

#### Scenario: Playbook file execution

- **GIVEN** a configuration with `playbook_file = "site.yml"`
- **AND** no `play` blocks are specified
- **WHEN** the provisioner executes
- **THEN** it SHALL run the specified playbook file

#### Scenario: Collection plays execution with block syntax

- **GIVEN** a configuration with one or more `play` blocks using HCL2 block syntax
- **AND** `playbook_file` is not specified
- **WHEN** the provisioner executes
- **THEN** it SHALL execute each play in sequence
- **AND** for role FQCNs, it SHALL generate temporary playbooks

#### Scenario: Invalid play array assignment syntax

- **GIVEN** a configuration attempting `play = [...]` or `plays = [...]` array assignment syntax
- **WHEN** Packer parses the configuration
- **THEN** it SHALL return an error indicating block syntax is required
- **AND** the error message SHALL suggest using `play { }` block format

#### Scenario: Multiple plays using repeated blocks

- **GIVEN** a configuration with multiple `play` blocks
- **WHEN** the configuration is parsed
- **THEN** each `play { }` block SHALL be processed in declaration order
- **AND** each block SHALL support independent configuration (name, target, extra_vars, become, become_user, tags, skip_tags, etc.)

#### Scenario: Both playbook_file and play specified

- **GIVEN** a configuration with both `playbook_file` and `play` blocks
- **WHEN** the configuration is validated
- **THEN** it SHALL return an error: "you may specify only one of `playbook_file` or `play`"

#### Scenario: Neither playbook_file nor play specified

- **GIVEN** a configuration with neither `playbook_file` nor `play` blocks
- **WHEN** the configuration is validated
- **THEN** it SHALL return an error: "either `playbook_file` or `play` must be defined"

### Requirement: HCL Block Naming Convention

The local provisioner SHALL follow HCL idioms for block naming, using singular names for blocks that can be repeated.

#### Scenario: Block name follows HCL conventions

- **GIVEN** the provisioner HCL2 spec definition
- **WHEN** defining blocks that represent individual items in a collection
- **THEN** the block SHALL use singular naming (`play` not `plays`)
- **AND** multiple items SHALL be expressed as repeated singular blocks

#### Scenario: Deprecated plural block name (migration)

- **GIVEN** a configuration using the deprecated `plays { }` block name
- **WHEN** Packer parses the configuration
- **THEN** it SHALL return an error indicating the block is not recognized
- **AND** the error message SHALL indicate that `play { }` is the correct syntax
