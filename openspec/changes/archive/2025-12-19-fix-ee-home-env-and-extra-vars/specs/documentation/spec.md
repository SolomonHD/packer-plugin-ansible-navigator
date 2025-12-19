## ADDED Requirements

### Requirement: REQ-DOC-EE-ROLE-001 Minimal example: EE + requirements.yml + collection role target

Documentation SHALL include a minimal, copy-pasteable Packer HCL example demonstrating:

- `navigator_config { execution_environment { enabled = true ... } }`
- collection installation via `requirements.yml`
- role execution via `play { target = "<namespace>.<collection>.<role>" }`

#### Scenario: Minimal example exists and uses role FQDN target

- **GIVEN** a user is reading the documentation examples
- **WHEN** they look for a minimal execution-environment example
- **THEN** the docs SHALL include an example that:
  - enables execution environment via `navigator_config.execution_environment.enabled = true`
  - installs at least one collection via `requirements_file = "requirements.yml"`
  - uses `play { target = "<namespace>.<collection>.<role>" }` (role FQDN), not only playbook-path targets

