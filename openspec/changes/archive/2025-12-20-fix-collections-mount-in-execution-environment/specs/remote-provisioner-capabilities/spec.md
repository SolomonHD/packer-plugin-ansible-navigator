# Spec Delta: Remote Provisioner Capabilities

## MODIFIED Requirements

### Requirement: Collections Path Environment Variable

The SSH-based provisioner SHALL export the collections path to Ansible using the `ANSIBLE_COLLECTIONS_PATH` environment variable (singular form, not deprecated plural `ANSIBLE_COLLECTIONS_PATHS`).

#### Scenario: collections_path exported via ANSIBLE_COLLECTIONS_PATH

- **GIVEN** a configuration with `collections_path` set
- **WHEN** the provisioner executes any ansible-galaxy operation and any ansible-navigator play execution
- **THEN** it SHALL set `ANSIBLE_COLLECTIONS_PATH` (singular) to the provided `collections_path` value
- **AND** it SHALL NOT set `ANSIBLE_COLLECTIONS_PATHS` (deprecated plural form)

#### Scenario: default collections path when not specified

- **GIVEN** a configuration without explicit `collections_path`
- **WHEN** the provisioner needs to export collections path
- **THEN** it SHALL use `~/.packer.d/ansible_collections_cache/ansible_collections` as the default path
- **AND** it SHALL set `ANSIBLE_COLLECTIONS_PATH` to this default value

## ADDED Requirements

### Requirement: Execution Environment Collections Volume Mount

When `navigator_config.execution_environment.enabled` is `true`, the SSH-based provisioner SHALL automatically mount the collections cache directory as a volume in the execution environment container.

#### Scenario: automatic volume mount for collections cache with EE enabled

- **GIVEN** a configuration with:
  ```hcl
  navigator_config {
    execution_environment {
      enabled = true
      image = "quay.io/ansible/creator-ee:latest"
    }
  }
  ```
- **AND** collections will be or have been installed to the collections cache directory
- **WHEN** the provisioner generates the ansible-navigator configuration
- **THEN** it SHALL add a volume mount mapping the host collections cache directory to a container path
- **AND** the volume mount SHALL be read-only (`:ro` suffix)
- **AND** the host path SHALL be the absolute path to `~/.packer.d/ansible_collections_cache/ansible_collections` (with tilde expanded)
- **AND** the container path SHALL be a writable location like `/tmp/.packer_ansible/collections`

#### Scenario: custom collections_path with EE enabled

- **GIVEN** a configuration with:
  ```hcl
  collections_path = "./.ansible/collections"
  navigator_config {
    execution_environment {
      enabled = true
      image = "quay.io/ansible/creator-ee:latest"
    }
  }
  ```
- **WHEN** the provisioner generates the ansible-navigator configuration
- **THEN** it SHALL mount the custom `collections_path` as a volume
- **AND** the host path SHALL be the absolute path to `./.ansible/collections`

#### Scenario: no volume mount when EE is disabled

- **GIVEN** a configuration without `navigator_config.execution_environment.enabled` or with it set to `false`
- **WHEN** the provisioner generates the ansible-navigator configuration
- **THEN** it SHALL NOT add any collections cache volume mount
- **AND** collections SHALL be accessed directly from the host filesystem

### Requirement: ANSIBLE_COLLECTIONS_PATH in Execution Environment

When `navigator_config.execution_environment.enabled` is `true`, the SSH-based provisioner SHALL set `ANSIBLE_COLLECTIONS_PATH` to point to the mounted collections path inside the container.

#### Scenario: ANSIBLE_COLLECTIONS_PATH set in EE environment variables

- **GIVEN** a configuration with:
  ```hcl
  navigator_config {
    execution_environment {
      enabled = true
      image = "quay.io/ansible/creator-ee:latest"
    }
  }
  ```
- **WHEN** the provisioner generates the ansible-navigator configuration
- **THEN** it SHALL add `ANSIBLE_COLLECTIONS_PATH` to the execution environment's environment variables
- **AND** the value SHALL be the container-side path of the mounted collections (e.g., `/tmp/.packer_ansible/collections`)
- **AND** this SHALL allow Ansible to discover collection roles inside the container

#### Scenario: ANSIBLE_COLLECTIONS_PATH not set outside EE context

- **GIVEN** a configuration without execution environment enabled
- **WHEN** the provisioner executes
- **THEN** it SHALL set `ANSIBLE_COLLECTIONS_PATH` as a host environment variable
- **AND** it SHALL NOT modify navigator_config environment variables

#### Scenario: user-provided navigator_config environment variables preserved

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
- **WHEN** the provisioner adds `ANSIBLE_COLLECTIONS_PATH` to environment variables
- **THEN** it SHALL preserve all user-provided environment variables
- **AND** it SHALL merge `ANSIBLE_COLLECTIONS_PATH` with existing environment variables
