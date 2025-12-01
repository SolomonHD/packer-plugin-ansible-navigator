# project-metadata Specification

## Purpose
TBD - created by archiving change cleanup-hashicorp-vestiges. Update Purpose after archive.
## Requirements
### Requirement: Repository Ownership Clarity
The project MUST clearly identify its current owner and maintainer in repository configuration files.

#### Scenario: Code ownership file
- **WHEN** reviewing the CODEOWNERS file
- **THEN** it SHALL reference `@SolomonHD` as the owner
- **AND** it SHALL NOT reference `@hashicorp/packer` or other HashiCorp team identifiers

#### Scenario: Maintainer attribution
- **WHEN** reviewing project metadata files
- **THEN** the current maintainer (`SolomonHD`) SHALL be clearly identified
- **AND** no conflicting ownership attributions SHALL exist

### Requirement: Consistent License Headers
Configuration and tooling files MUST use consistent Apache 2.0 license headers with correct copyright attribution.

#### Scenario: Linter configuration license header
- **WHEN** reviewing `.golangci.yml`
- **THEN** it SHALL contain an Apache 2.0 license header
- **AND** copyright SHALL be attributed to SolomonHD
- **AND** no MPL-2.0 or HashiCorp copyright notices SHALL be present

### Requirement: Clean Changelog History
The changelog MUST clearly distinguish between the current project's history and legacy history from the original forked project.

#### Scenario: Current project changes
- **WHEN** reviewing CHANGELOG.md entries for v1.0.0 and later (current project)
- **THEN** entries SHALL document changes specific to packer-plugin-ansible-navigator
- **AND** the v1.0.0 entry MAY reference the fork origin for historical context

#### Scenario: Legacy history separation
- **WHEN** reviewing CHANGELOG.md entries for versions prior to the fork
- **THEN** any retained legacy entries SHALL be clearly marked as "Legacy History from packer-plugin-ansible"
- **OR** they SHALL be removed entirely
- **AND** no PR links to the original HashiCorp repository SHALL appear without clear legacy attribution

### Requirement: Accurate Go Module Metadata
The go.mod file MUST accurately reference this repository in all metadata and comments.

#### Scenario: Retract comment accuracy
- **WHEN** reviewing the go.mod retract block comments
- **THEN** any explanatory comments SHALL reference this repository
- **AND** SHALL NOT reference the original packer-plugin-ansible repository without context

### Requirement: Single Build Configuration
The project MUST maintain exactly one makefile (GNUmakefile) to avoid confusion and duplicate configuration.

#### Scenario: Makefile uniqueness
- **WHEN** listing files in the repository root
- **THEN** only `GNUmakefile` SHALL exist
- **AND** no `MAKEFILE` or `Makefile` duplicates SHALL exist

#### Scenario: Build script organization reference
- **WHEN** GNUmakefile generates documentation or calls external scripts
- **THEN** organization references SHALL use "SolomonHD"
- **AND** no references to "hashicorp" organization SHALL exist

