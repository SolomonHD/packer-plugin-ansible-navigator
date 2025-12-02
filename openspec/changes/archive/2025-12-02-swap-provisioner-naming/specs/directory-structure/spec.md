## MODIFIED Requirements

### Requirement: Consistent Directory Naming
The provisioner directory structure SHALL align with plugin registration names and Packer ecosystem conventions, following the pattern established by the official ansible plugin.

#### Scenario: Default provisioner directory (SSH-based remote execution)
- **WHEN** locating the default "ansible-navigator" provisioner code
- **THEN** it SHALL be found in `provisioner/ansible-navigator/` directory
- **AND** this SHALL contain SSH-based remote execution logic (runs from local machine, connects to target)
- **AND** the package name SHALL be `ansiblenavigator`

#### Scenario: Local provisioner directory (on-target execution)
- **WHEN** locating the "ansible-navigator-local" provisioner code
- **THEN** it SHALL be found in `provisioner/ansible-navigator-local/` directory
- **AND** this SHALL contain local execution logic (runs directly on target machine)
- **AND** the directory name SHALL include "-local" suffix to indicate execution mode
- **AND** the package name SHALL be `ansiblenavigatorlocal`

#### Scenario: No legacy naming
- **WHEN** searching the codebase for provisioner references
- **THEN** there SHALL be no references to "ansible-navigator-remote"
- **AND** all references SHALL use either "ansible-navigator" (default/SSH) or "ansible-navigator-local" (on-target)

### Requirement: Package Import Alignment
Go package import paths SHALL match the physical directory structure and follow Packer ecosystem naming conventions.

#### Scenario: Main package imports
- **WHEN** `main.go` imports provisioner packages
- **THEN** it SHALL import from `github.com/solomonhd/packer-plugin-ansible-navigator/provisioner/ansible-navigator`
- **AND** it SHALL import from `github.com/solomonhd/packer-plugin-ansible-navigator/provisioner/ansible-navigator-local`
- **AND** no import paths SHALL reference "ansible-navigator-remote"

#### Scenario: Import alias conventions
- **WHEN** main.go declares import aliases
- **THEN** the SSH-based provisioner SHALL use alias `ansiblenavigator`
- **AND** the local provisioner SHALL use alias `ansiblenavigatorlocal`
- **AND** the aliases SHALL match the Go package names

#### Scenario: Cross-package imports
- **WHEN** code within one provisioner package needs to reference another
- **THEN** import paths SHALL use the standardized directory names
- **AND** SHALL maintain consistency across all Go files