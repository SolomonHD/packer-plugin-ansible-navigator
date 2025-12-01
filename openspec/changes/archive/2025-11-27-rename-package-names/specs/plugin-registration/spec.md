## ADDED Requirements

### Requirement: Package Naming Convention
Go package names SHALL consistently reflect the "ansible-navigator" branding to maintain code clarity and avoid confusion with the original HashiCorp ansible plugin.

The package naming convention SHALL be:
- `provisioner/ansible-navigator/` SHALL use `package ansiblenavigatorlocal`
- `provisioner/ansible-navigator-remote/` SHALL use `package ansiblenavigatorremote`

#### Scenario: Local provisioner package name
- **WHEN** examining the source code in `provisioner/ansible-navigator/`
- **THEN** all Go files SHALL declare `package ansiblenavigatorlocal`
- **AND** the package name SHALL clearly indicate both the tool (ansible-navigator) and execution mode (local)

#### Scenario: Remote provisioner package name
- **WHEN** examining the source code in `provisioner/ansible-navigator-remote/`
- **THEN** all Go files SHALL declare `package ansiblenavigatorremote`
- **AND** the package name SHALL clearly indicate both the tool (ansible-navigator) and execution mode (remote)

## MODIFIED Requirements

### Requirement: Plugin Registration Names
The plugin SHALL register two provisioners with Packer:
- `ansible-navigator` for local execution (default)
- `ansible-navigator-remote` for remote execution

The internal directory structure SHALL align with these registration names:
- `provisioner/ansible-navigator/` contains local execution code
- `provisioner/ansible-navigator-remote/` contains remote execution code

Import aliases in main.go SHALL match the package names:
- `ansiblenavigatorlocal` for the local provisioner import
- `ansiblenavigatorremote` for the remote provisioner import

#### Scenario: Local provisioner registration
- **WHEN** the plugin initializes
- **THEN** it SHALL register "ansible-navigator" using the Provisioner from `provisioner/ansible-navigator/`
- **AND** the directory name SHALL match the registration name for clarity
- **AND** the import alias SHALL be `ansiblenavigatorlocal`

#### Scenario: Remote provisioner registration  
- **WHEN** the plugin initializes
- **THEN** it SHALL register "ansible-navigator-remote" using the Provisioner from `provisioner/ansible-navigator-remote/`
- **AND** the directory name SHALL match the registration name for clarity
- **AND** the import alias SHALL be `ansiblenavigatorremote`

#### Scenario: Import path consistency
- **WHEN** main.go imports provisioners
- **THEN** import paths SHALL reflect actual directory names
- **AND** SHALL use: `provisioner/ansible-navigator` for local with alias `ansiblenavigatorlocal`
- **AND** SHALL use: `provisioner/ansible-navigator-remote` for remote with alias `ansiblenavigatorremote`