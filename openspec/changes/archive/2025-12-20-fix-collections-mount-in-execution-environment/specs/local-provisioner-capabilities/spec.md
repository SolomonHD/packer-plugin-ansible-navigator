# Spec Delta: Local Provisioner Capabilities

## MODIFIED Requirements

### Requirement: Collections Path Environment Variable

The on-target provisioner SHALL export the collections path to Ansible using the `ANSIBLE_COLLECTIONS_PATH` environment variable (singular form, not deprecated plural `ANSIBLE_COLLECTIONS_PATHS`).

#### Scenario: collections_path exported via ANSIBLE_COLLECTIONS_PATH

- **GIVEN** a configuration with `collections_path` set
- **WHEN** the provisioner executes any ansible-galaxy operation and any ansible-navigator play execution on the target
- **THEN** it SHALL ensure `ANSIBLE_COLLECTIONS_PATH` (singular) is set to the provided `collections_path` value for those operations
- **AND** it SHALL NOT set `ANSIBLE_COLLECTIONS_PATHS` (deprecated plural form)

#### Scenario: default collections path when not specified

- **GIVEN** a configuration without explicit `collections_path`
- **WHEN** the provisioner needs to export collections path on the target
- **THEN** it SHALL use a staging directory subdirectory for collections
- **AND** it SHALL set `ANSIBLE_COLLECTIONS_PATH` to this default path

## ADDED Requirements

### Requirement: Execution Environment Collections Volume Mount (On-Target)

When `navigator_config.execution_environment.enabled` is `true`, the on-target provisioner SHALL automatically mount the collections directory as a volume in the execution environment container running on the target.

#### Scenario: automatic volume mount for collections with EE enabled on target

- **GIVEN** a configuration with:
  ```hcl
  navigator_config {
    execution_environment {
      enabled = true
      image = "quay.io/ansible/creator-ee:latest"
    }
  }
  ```
- **AND** collections have been uploaded to the target staging directory
- **WHEN** the provisioner generates the ansible-navigator configuration on the target
- **THEN** it SHALL add a volume mount mapping the target collections directory to a container path
- **AND** the volume mount SHALL be read-only (`:ro` suffix)
- **AND** the host path SHALL be the absolute path to the collections directory in the staging area
- **AND** the container path SHALL be a writable location like `/tmp/.packer_ansible/collections`

#### Scenario: custom collections_path with EE enabled on target

- **GIVEN** a configuration with:
  ```hcl
  collections_path = "/opt/ansible/collections"
  navigator_config {
    execution_environment {
      enabled = true
      image = "quay.io/ansible/creator-ee:latest"
    }
  }
  ```
- **AND** the provisioner is running on the target
- **WHEN** the provisioner generates the ansible-navigator configuration
- **THEN** it SHALL mount the custom `collections_path` as a volume
- **AND** the host path SHALL be the absolute path to `/opt/ansible/collections` on the target

#### Scenario: no volume mount when EE is disabled on target

- **GIVEN** a configuration without `navigator_config.execution_environment.enabled` or with it set to `false`
- **WHEN** the provisioner executes on the target
- **THEN** it SHALL NOT add any collections volume mount
- **AND** collections SHALL be accessed directly from the target filesystem

### Requirement: ANSIBLE_COLLECTIONS_PATH in Execution Environment (On-Target)

When `navigator_config.execution_environment.enabled` is `true`, the on-target provisioner SHALL set `ANSIBLE_COLLECTIONS_PATH` to point to the mounted collections path inside the container.

#### Scenario: ANSIBLE_COLLECTIONS_PATH set in EE environment variables on target

- **GIVEN** a configuration with:
  ```hcl
  navigator_config {
    execution_environment {
      enabled = true
      image = "quay.io/ansible/creator-ee:latest"
    }
  }
  ```
- **AND** ansible-navigator is running on the target
- **WHEN** the provisioner generates the ansible-navigator configuration
- **THEN** it SHALL add `ANSIBLE_COLLECTIONS_PATH` to the execution environment's environment variables
- **AND** the value SHALL be the container-side path of the mounted collections (e.g., `/tmp/.packer_ansible/collections`)
- **AND** this SHALL allow Ansible to discover collection roles inside the container on the target

#### Scenario: ANSIBLE_COLLECTIONS_PATH not set outside EE context on target

- **GIVEN** a configuration without execution environment enabled
- **WHEN** the provisioner executes on the target
- **THEN** it SHALL set `ANSIBLE_COLLECTIONS_PATH` as a target environment variable
- **AND** it SHALL NOT modify navigator_config environment variables

#### Scenario: user-provided navigator_config environment variables preserved on target

- **GIVEN** a configuration with:
  ```hcl
  navigator_config {
    execution_environment {
      enabled = true
      image = "quay.io/ansible/creator-ee:latest"
      environment_variables {
        set = {
          CUSTOM_VAR = "value"
        }
      }
    }
  }
  ```
- **WHEN** the provisioner adds `ANSIBLE_COLLECTIONS_PATH` to environment variables on the target
- **THEN** it SHALL preserve all user-provided environment variables
- **AND** it SHALL merge `ANSIBLE_COLLECTIONS_PATH` with existing environment variables
