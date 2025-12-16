## MODIFIED Requirements

### Requirement: Dependency Management (requirements_file)

The SSH-based provisioner SHALL support dependency installation via an optional `requirements_file` that can define both roles and collections.

The SSH-based provisioner SHALL expose a consistent set of dependency-install configuration options:

- `requirements_file` (string, optional)
- `offline_mode` (bool, optional)
- `roles_path` (string, optional)
- `collections_path` (string, optional)
- `galaxy_force` (bool, optional)
- `galaxy_force_with_deps` (bool, optional)
- `galaxy_command` (string, optional; defaults to `ansible-galaxy`)
- `galaxy_args` (list(string), optional)

#### Scenario: requirements_file installs roles and collections

- **GIVEN** a configuration with `requirements_file = "requirements.yml"`
- **WHEN** the provisioner executes
- **THEN** it SHALL install roles and collections from that file before executing any plays

#### Scenario: requirements_file omitted

- **GIVEN** a configuration with no `requirements_file`
- **WHEN** the provisioner executes
- **THEN** it SHALL proceed without performing dependency installation

#### Scenario: roles_path exported via ANSIBLE_ROLES_PATH

- **GIVEN** a configuration with `roles_path` set
- **WHEN** the provisioner executes any ansible-galaxy operation and any ansible-navigator play execution
- **THEN** it SHALL set `ANSIBLE_ROLES_PATH` to the provided `roles_path` value

#### Scenario: collections_path exported via ANSIBLE_COLLECTIONS_PATHS

- **GIVEN** a configuration with `collections_path` set
- **WHEN** the provisioner executes any ansible-galaxy operation and any ansible-navigator play execution
- **THEN** it SHALL set `ANSIBLE_COLLECTIONS_PATHS` to the provided `collections_path` value

#### Scenario: Galaxy command override and extra args

- **GIVEN** a configuration with `requirements_file` set
- **AND** `galaxy_command` set to a custom value
- **AND** `galaxy_args` set to one or more arguments
- **WHEN** the provisioner installs roles and collections
- **THEN** it SHALL invoke Galaxy using the configured `galaxy_command`
- **AND** it SHALL append `galaxy_args` to the constructed Galaxy argument list
- **AND** this behavior SHALL be consistent for both roles install and collections install

#### Scenario: galaxy_force maps to --force

- **GIVEN** a configuration with `galaxy_force = true`
- **AND** `galaxy_force_with_deps` is unset or `false`
- **WHEN** the provisioner invokes ansible-galaxy
- **THEN** it SHALL include `--force`

#### Scenario: galaxy_force_with_deps maps to --force-with-deps and takes precedence

- **GIVEN** a configuration with `galaxy_force_with_deps = true`
- **WHEN** the provisioner invokes ansible-galaxy
- **THEN** it SHALL include `--force-with-deps`
- **AND** it SHALL NOT additionally include `--force`

#### Scenario: offline_mode maps to --offline

- **GIVEN** a configuration with `offline_mode = true`
- **WHEN** the provisioner invokes ansible-galaxy to install from `requirements_file`
- **THEN** it SHALL include `--offline`
