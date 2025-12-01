# directory-structure Specification

## Purpose
TBD - created by archiving change standardize-provisioner-naming. Update Purpose after archive.
## Requirements
### Requirement: Consistent Directory Naming
The provisioner directory structure SHALL align with plugin registration names to eliminate confusion and follow the principle of least surprise.

#### Scenario: Default provisioner directory
- **WHEN** locating the default "ansible-navigator" provisioner code
- **THEN** it SHALL be found in `provisioner/ansible-navigator/` directory
- **AND** this SHALL contain local execution logic
- **AND** no "-local" suffix SHALL be used in the directory name

#### Scenario: Remote provisioner directory
- **WHEN** locating the "ansible-navigator-remote" provisioner code
- **THEN** it SHALL be found in `provisioner/ansible-navigator-remote/` directory
- **AND** this SHALL contain remote execution logic
- **AND** the directory name SHALL match the plugin registration name

#### Scenario: No legacy naming
- **WHEN** searching the codebase for provisioner references
- **THEN** there SHALL be no references to "ansible-navigator-local"
- **AND** all references SHALL use either "ansible-navigator" or "ansible-navigator-remote"

### Requirement: Package Import Alignment
Go package import paths SHALL match the physical directory structure without legacy naming artifacts.

#### Scenario: Main package imports
- **WHEN** `main.go` imports provisioner packages
- **THEN** it SHALL import from `github.com/SolomonHD/packer-plugin-ansible-navigator/provisioner/ansible-navigator`
- **AND** it SHALL import from `github.com/SolomonHD/packer-plugin-ansible-navigator/provisioner/ansible-navigator-remote`
- **AND** no import paths SHALL reference "ansible-navigator-local"

#### Scenario: Cross-package imports
- **WHEN** code within one provisioner package needs to reference another
- **THEN** import paths SHALL use the standardized directory names
- **AND** SHALL maintain consistency across all Go files

