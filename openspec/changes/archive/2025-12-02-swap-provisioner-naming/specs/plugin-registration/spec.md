## MODIFIED Requirements

### Requirement: Plugin Registration Names

The plugin SHALL register two provisioners with Packer using Packer SDK naming conventions, following the pattern established by the official ansible plugin:

- `plugin.DEFAULT_NAME` (equals `"-packer-default-plugin-name-"`) for SSH-based remote execution, accessible in HCL as `ansible-navigator`
- `"local"` for on-target local execution, accessible in HCL as `ansible-navigator-local`

The directory structure SHALL align with these HCL names:
- `provisioner/ansible-navigator/` contains SSH-based remote execution code
- `provisioner/ansible-navigator-local/` contains on-target local execution code

Import aliases in main.go SHALL match the package names:
- `ansiblenavigator` for the SSH-based provisioner import (default)
- `ansiblenavigatorlocal` for the local provisioner import

#### Scenario: SSH-based provisioner registration with DEFAULT_NAME
- **WHEN** the plugin initializes
- **THEN** it SHALL register the SSH-based provisioner using `plugin.DEFAULT_NAME` constant
- **AND** the provisioner SHALL be accessible in HCL as `ansible-navigator`
- **AND** the import alias SHALL be `ansiblenavigator`
- **AND** this SHALL run from the local machine and connect to the target via SSH

#### Scenario: Local provisioner registration with short name
- **WHEN** the plugin initializes
- **THEN** it SHALL register `"local"` using the Provisioner from `provisioner/ansible-navigator-local/`
- **AND** the provisioner SHALL be accessible in HCL as `ansible-navigator-local` (Packer prefixes with plugin alias)
- **AND** the import alias SHALL be `ansiblenavigatorlocal`
- **AND** this SHALL run directly on the target machine

#### Scenario: Describe command output
- **WHEN** running `./packer-plugin-ansible-navigator describe`
- **THEN** the JSON output SHALL list provisioners as `["-packer-default-plugin-name-", "local"]`
- **AND** the output SHALL NOT contain full names like `ansible-navigator` or `ansible-navigator-local`

#### Scenario: Import path consistency
- **WHEN** main.go imports provisioners
- **THEN** import paths SHALL reflect actual directory names
- **AND** SHALL use: `provisioner/ansible-navigator` for SSH-based with alias `ansiblenavigator`
- **AND** SHALL use: `provisioner/ansible-navigator-local` for local with alias `ansiblenavigatorlocal`

### Requirement: Package Naming Convention
Go package names SHALL consistently reflect the "ansible-navigator" branding and execution mode to maintain code clarity and align with Packer ecosystem conventions.

The package naming convention SHALL be:
- `provisioner/ansible-navigator/` SHALL use `package ansiblenavigator` (primary/default)
- `provisioner/ansible-navigator-local/` SHALL use `package ansiblenavigatorlocal`

#### Scenario: SSH-based provisioner package name
- **WHEN** examining the source code in `provisioner/ansible-navigator/`
- **THEN** all Go files SHALL declare `package ansiblenavigator`
- **AND** the package name SHALL indicate this is the primary provisioner

#### Scenario: Local provisioner package name
- **WHEN** examining the source code in `provisioner/ansible-navigator-local/`
- **THEN** all Go files SHALL declare `package ansiblenavigatorlocal`
- **AND** the package name SHALL clearly indicate the local execution mode