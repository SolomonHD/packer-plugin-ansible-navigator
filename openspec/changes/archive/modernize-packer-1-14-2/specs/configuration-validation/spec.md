## MODIFIED Requirements

### Requirement: Centralized Validation
The plugin SHALL perform all configuration validation in a single centralized Config.Validate() method.

#### Scenario: All validation is centralized
- **WHEN** configuration is prepared
- **THEN** Config.Validate() is called exactly once
- **AND** all validation logic is in Validate() method
- **AND** no validation occurs scattered in other methods

#### Scenario: Validation errors are aggregated
- **WHEN** multiple validation errors exist
- **THEN** all errors are collected and returned together
- **AND** user sees all issues at once
- **AND** errors include field names and constraints

### Requirement: HCL2 Schema Generation
The plugin SHALL generate accurate HCL2 schemas for all configuration parameters.

#### Scenario: Schema reflects actual types
- **WHEN** HCL2 reads configuration
- **THEN** schema matches struct field types exactly
- **AND** optional vs required is correctly specified
- **AND** type constraints are enforced

## ADDED Requirements

### Requirement: Validation Helper Functions
The plugin SHALL provide reusable validation helper functions for common patterns.

#### Scenario: File validation is reusable
- **WHEN** validating file paths
- **THEN** validateFileConfig() helper is used
- **AND** helper checks file existence
- **AND** helper checks file vs directory
- **AND** helper provides consistent error format

#### Scenario: Directory validation is reusable
- **WHEN** validating directory paths
- **THEN** validateInventoryDirectoryConfig() helper is used
- **AND** helper checks directory existence
- **AND** helper provides consistent error format

### Requirement: Cross-Field Validation
The plugin SHALL validate relationships between configuration fields.

#### Scenario: Mutual exclusivity is enforced
- **WHEN** playbook_file and plays are both set
- **THEN** validation fails with clear error
- **AND** error explains mutual exclusivity
- **AND** error is reported during Prepare()

#### Scenario: Conditional requirements are checked
- **WHEN** structured_logging is enabled
- **THEN** navigator_mode must be "json"
- **AND** validation enforces this constraint
- **AND** helpful error message is provided

### Requirement: Default Value Application
The plugin SHALL apply sensible defaults before validation.

#### Scenario: Defaults are set in Prepare()
- **WHEN** configuration is prepared
- **THEN** empty required fields get defaults
- **AND** defaults align with best practices
- **AND** defaults are documented in godoc