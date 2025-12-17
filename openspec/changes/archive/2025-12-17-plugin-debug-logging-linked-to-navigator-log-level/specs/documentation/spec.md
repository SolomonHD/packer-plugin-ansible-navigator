## ADDED Requirements

### Requirement: Document linkage between navigator logging and plugin debug output

User-facing documentation SHALL state that `navigator_config.logging.level` controls both ansible-navigator logging and the plugin’s debug output (prefixed with `[DEBUG]`), and SHALL NOT introduce a separate plugin-specific debug/log-level configuration field.

#### Scenario: Linkage is documented prominently

- **GIVEN** a user consults the configuration documentation for `navigator_config.logging.level`
- **WHEN** the documentation describes what `navigator_config.logging.level` controls
- **THEN** it SHALL explicitly state that `navigator_config.logging.level` controls both:
  - ansible-navigator logging behavior
  - the plugin’s debug output (prefixed with `[DEBUG]`)

#### Scenario: Documentation does not introduce a separate plugin log-level option

- **GIVEN** a user searches documentation for how to enable plugin debug output
- **WHEN** the documentation provides guidance
- **THEN** it SHALL NOT introduce a new plugin-specific `debug` or `log_level` option
- **AND** it SHALL direct users to set `navigator_config.logging.level = "debug"` instead
