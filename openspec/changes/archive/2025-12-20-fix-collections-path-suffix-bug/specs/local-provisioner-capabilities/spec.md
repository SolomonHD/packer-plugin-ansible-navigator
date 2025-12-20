# local-provisioner-capabilities Spec Delta

## MODIFIED Requirements

### Requirement: Dependency Management (requirements_file)

The local provisioner SHALL support dependency installation via an optional `requirements_file` that can define both roles and collections.

#### Scenario: collections_path exported via ANSIBLE_COLLECTIONS_PATHS

- **GIVEN** a configuration with `collections_path` set
- **WHEN** the provisioner executes any ansible-galaxy operation and any ansible-navigator play execution on the target
- **THEN** it SHALL ensure `ANSIBLE_COLLECTIONS_PATH` (singular) is set to the provided `collections_path` value for those operations
- **AND** it SHALL NOT set `ANSIBLE_COLLECTIONS_PATHS` (deprecated plural form)

## ADDED Requirements

### Requirement: Collections path MUST be passed unmodified for collection discovery (on-target)

The on-target provisioner SHALL treat `collections_path` as the exact path value to export to Ansible via `ANSIBLE_COLLECTIONS_PATH` and SHALL NOT append or remove path suffixes.

#### Scenario: collections_path passed unmodified to environment variable

- **GIVEN** a configuration with `collections_path = "/home/user/.packer.d/ansible_collections_cache"`
- **WHEN** the provisioner executes ansible-galaxy or ansible-navigator operations on the target
- **THEN** it SHALL set `ANSIBLE_COLLECTIONS_PATH=/home/user/.packer.d/ansible_collections_cache`
- **AND** it SHALL NOT append `/ansible_collections` or any other suffix to the path
- **AND** it SHALL use the singular form `ANSIBLE_COLLECTIONS_PATH` (not `ANSIBLE_COLLECTIONS_PATHS`)

#### Scenario: collections_path with trailing ansible_collections not doubled

- **GIVEN** a configuration with `collections_path = "/custom/path/ansible_collections"`
- **WHEN** the provisioner passes the path to Ansible via environment variables on the target
- **THEN** it SHALL set `ANSIBLE_COLLECTIONS_PATH=/custom/path/ansible_collections`
- **AND** it SHALL NOT check for or remove the suffix
- **AND** it SHALL NOT produce `ANSIBLE_COLLECTIONS_PATH=/custom/path/ansible_collections/ansible_collections`

#### Scenario: Execution environment mount uses unmodified collections_path

- **GIVEN** a configuration with:
  - `collections_path = "/home/user/.packer.d/ansible_collections_cache"`
  - `navigator_config.execution_environment.enabled = true`
- **WHEN** the provisioner generates the ansible-navigator.yml with automatic EE defaults
- **THEN** the volume mount source SHALL be `/home/user/.packer.d/ansible_collections_cache` (the exact configured value)
- **AND** it SHALL NOT append `ansible_collections` to create a mount source of `/home/user/.packer.d/ansible_collections_cache/ansible_collections`

#### Scenario: Collections installed by ansible-galaxy are accessible inside EE on target

- **GIVEN** a configuration with:
  - `requirements_file = "./requirements.yml"` containing a collection reference
  - `collections_path = "/home/user/.packer.d/ansible_collections_cache"`
  - `navigator_config.execution_environment.enabled = true`
- **WHEN** ansible-galaxy installs the collection on the target to `<collections_path>/ansible_collections/<namespace>/<collection>`
- **AND** the execution environment container is started on the target with the automatic volume mount
- **THEN** Ansible inside the container SHALL discover the collection at the mounted path
- **AND** role FQDNs like `<namespace>.<collection>.<role_name>` SHALL resolve successfully
- **AND** `unable to find role` errors SHALL NOT occur

#### Scenario: Deprecation warning eliminated on target

- **GIVEN** any configuration that uses `collections_path`
- **WHEN** the provisioner executes Ansible operations on the target
- **THEN** Ansible SHALL NOT emit deprecation warnings about `ANSIBLE_COLLECTIONS_PATHS` (plural)
- **AND** the plugin SHALL use only the modern singular form `ANSIBLE_COLLECTIONS_PATH`
