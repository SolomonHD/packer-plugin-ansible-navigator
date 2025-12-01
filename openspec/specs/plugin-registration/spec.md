# plugin-registration Specification

## Purpose
TBD - created by archiving change standardize-provisioner-naming. Update Purpose after archive.
## Requirements
### Requirement: Plugin Registration Names

The plugin SHALL register two provisioners with Packer using Packer SDK naming conventions:

- `plugin.DEFAULT_NAME` (equals `"-packer-default-plugin-name-"`) for local execution, accessible in HCL as `ansible-navigator`
- `"remote"` for remote execution, accessible in HCL as `ansible-navigator-remote`

The internal directory structure SHALL remain unchanged:
- `provisioner/ansible-navigator/` contains local execution code
- `provisioner/ansible-navigator-remote/` contains remote execution code

Import aliases in main.go SHALL match the package names:
- `ansiblenavigatorlocal` for the local provisioner import
- `ansiblenavigatorremote` for the remote provisioner import

#### Scenario: Local provisioner registration with DEFAULT_NAME
- **WHEN** the plugin initializes
- **THEN** it SHALL register using `plugin.DEFAULT_NAME` constant
- **AND** the provisioner SHALL be accessible in HCL as `ansible-navigator`
- **AND** the import alias SHALL be `ansiblenavigatorlocal`

#### Scenario: Remote provisioner registration with short name
- **WHEN** the plugin initializes
- **THEN** it SHALL register `"remote"` using the Provisioner from `provisioner/ansible-navigator-remote/`
- **AND** the provisioner SHALL be accessible in HCL as `ansible-navigator-remote` (Packer prefixes with plugin alias)
- **AND** the import alias SHALL be `ansiblenavigatorremote`

#### Scenario: Describe command output
- **WHEN** running `./packer-plugin-ansible-navigator describe`
- **THEN** the JSON output SHALL list provisioners as `["-packer-default-plugin-name-", "remote"]`
- **AND** the output SHALL NOT contain the full names `ansible-navigator` or `ansible-navigator-remote`

#### Scenario: Import path consistency
- **WHEN** main.go imports provisioners
- **THEN** import paths SHALL reflect actual directory names
- **AND** SHALL use: `provisioner/ansible-navigator` for local with alias `ansiblenavigatorlocal`
- **AND** SHALL use: `provisioner/ansible-navigator-remote` for remote with alias `ansiblenavigatorremote`

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

