# remote-provisioner-capabilities Spec Delta

## ADDED Requirements

### Requirement: Play block naming and validation (remote)

The SSH-based provisioner SHALL expose multi-play configuration via a singular repeatable `play` block and SHALL reject legacy plural or array-based forms that conflict with this naming.

#### Scenario: Singular play block naming

- **GIVEN** a Packer configuration using `provisioner "ansible-navigator"`
- **AND** one or more `play { ... }` blocks are defined
- **WHEN** the configuration is parsed
- **THEN** each `play { ... }` block SHALL be accepted as a repeatable block
- **AND** the resulting configuration SHALL represent the plays as a collection in declaration order

#### Scenario: Legacy plays block rejected

- **GIVEN** a configuration that defines one or more `plays { ... }` blocks
- **WHEN** the configuration is validated
- **THEN** validation SHALL fail
- **AND** the error message SHALL explain that `plays` blocks are no longer supported
- **AND** the error message SHALL direct the user to use repeated `play { ... }` blocks instead

#### Scenario: Array syntax rejected in favor of blocks

- **GIVEN** a configuration that attempts `plays = [...]` or `play = [...]` array-style syntax
- **WHEN** the configuration is validated
- **THEN** validation SHALL fail
- **AND** the error message SHALL state that HCL2 block syntax is required
- **AND** the error message SHALL include an example using `play { ... }` blocks
