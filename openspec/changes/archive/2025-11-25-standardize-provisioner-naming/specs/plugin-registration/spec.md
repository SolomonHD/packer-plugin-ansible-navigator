## ADDED Requirements

### Requirement: Plugin Registration Names
The plugin SHALL register two provisioners with Packer:
- `ansible-navigator` for local execution (default)
- `ansible-navigator-remote` for remote execution

The internal directory structure SHALL align with these registration names:
- `provisioner/ansible-navigator/` contains local execution code
- `provisioner/ansible-navigator-remote/` contains remote execution code

#### Scenario: Local provisioner registration
- **WHEN** the plugin initializes
- **THEN** it SHALL register "ansible-navigator" using the Provisioner from `provisioner/ansible-navigator/`
- **AND** the directory name SHALL match the registration name for clarity

#### Scenario: Remote provisioner registration  
- **WHEN** the plugin initializes
- **THEN** it SHALL register "ansible-navigator-remote" using the Provisioner from `provisioner/ansible-navigator-remote/`
- **AND** the directory name SHALL match the registration name for clarity

#### Scenario: Import path consistency
- **WHEN** main.go imports provisioners
- **THEN** import paths SHALL reflect actual directory names
- **AND** SHALL use: `provisioner/ansible-navigator` for local
- **AND** SHALL use: `provisioner/ansible-navigator-remote` for remote