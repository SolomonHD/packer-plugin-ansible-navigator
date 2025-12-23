## ADDED Requirements

### Requirement: REQ-DOCS-NAVCFG-TOPLEVEL-001 Document newly supported `navigator_config` top-level settings

The project documentation SHALL include examples for configuring the newly-supported ansible-navigator v3.x top-level settings via `navigator_config { ... }`, and SHALL clarify underscore (HCL) vs hyphen (YAML) naming.

#### Scenario: Documentation includes examples of new settings

Given: a user reading the configuration documentation for the plugin
When: the user searches for `navigator_config` examples
Then: the documentation MUST show at least one example that uses `mode_settings`, `format`, `color`, `images`, `time_zone`, `documentation`, `editor`, and `inventory_columns`

#### Scenario: Documentation clarifies naming conversion

Given: a user is configuring `navigator_config` fields that are hyphenated in ansible-navigator.yml
When: the user reads the documentation for those settings
Then: the documentation MUST state that HCL uses underscores (e.g., `time_zone`) and the generated YAML uses hyphens where required (e.g., `time-zone`)
