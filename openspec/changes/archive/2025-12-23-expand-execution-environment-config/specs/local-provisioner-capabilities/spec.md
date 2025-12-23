## ADDED Requirements

### Requirement: Container Engine Configuration (local provisioner)

The on-target (`ansible-navigator-local`) provisioner SHALL support specifying the container engine used by ansible-navigator for execution environments.

#### Scenario: container_engine field accepted in HCL

- **GIVEN** a configuration with `provisioner "ansible-navigator-local"`
- **AND** `navigator_config.execution_environment.container_engine = "docker"`
- **WHEN** Packer parses the configuration
- **THEN** parsing SHALL succeed
- **AND** the configuration SHALL preserve the `container_engine` value

#### Scenario: container_engine generated in YAML

- **GIVEN** a configuration with `navigator_config.execution_environment.container_engine = "podman"`
- **WHEN** the provisioner generates `ansible-navigator.yml` on the target
- **THEN** the YAML SHALL include under `ansible-navigator.execution-environment`:
  ```yaml
  container-engine: podman
  ```
- **AND** ansible-navigator SHALL use the specified container runtime on the target

### Requirement: Container Options Configuration (local provisioner)

The on-target provisioner SHALL support passing arbitrary container runtime flags via `container_options`.

#### Scenario: container_options accepted as list of strings

- **GIVEN** a configuration with:
  ```hcl
  navigator_config {
    execution_environment {
      container_options = [
        "--net=host",
        "--security-opt", "label=disable"
      ]
    }
  }
  ```
- **WHEN** Packer parses the configuration
- **THEN** parsing SHALL succeed
- **AND** all container option values SHALL be preserved

#### Scenario: container_options generated in YAML on target

- **GIVEN** a configuration with `container_options = ["--privileged"]`
- **WHEN** the provisioner generates `ansible-navigator.yml` on the target
- **THEN** the YAML SHALL include the container-options list
- **AND** ansible-navigator SHALL pass these options to the container runtime on the target

### Requirement: Pull Arguments Configuration (local provisioner)

The on-target provisioner SHALL support passing arguments to the container image pull command via `pull_arguments`.

#### Scenario: pull_arguments accepted in HCL

- **GIVEN** a configuration with:
  ```hcl
  navigator_config {
    execution_environment {
      pull_arguments = ["--tls-verify=false"]
    }
  }
  ```
- **WHEN** Packer parses the configuration
- **THEN** parsing SHALL succeed
- **AND** all pull argument values SHALL be preserved

#### Scenario: pull_arguments nested under pull in YAML on target

- **GIVEN** a configuration with `pull_arguments = ["--tls-verify=false"]` and `pull_policy = "always"`
- **WHEN** the provisioner generates `ansible-navigator.yml` on the target
- **THEN** the YAML SHALL include:
  ```yaml
  ansible-navigator:
    execution-environment:
      pull:
        policy: always
        arguments:
          - --tls-verify=false
  ```
- **AND** both pull.policy and pull.arguments SHALL be nested correctly

### Requirement: Additional Volume Mounts (local provisioner)

The on-target provisioner SHALL allow users to specify additional volume mounts (if applicable for local execution).

#### Scenario: User-specified volume mounts accepted

- **GIVEN** a configuration with:
  ```hcl
  navigator_config {
    execution_environment {
      volume_mounts = [
        {
          src     = "/target/host/path"
          dest    = "/container/path"
          options = "ro"
        }
      ]
    }
  }
  ```
- **WHEN** Packer parses the configuration
- **THEN** parsing SHALL succeed
- **AND** volume mount values SHALL be preserved

#### Scenario: Volume mounts generated in YAML on target

- **GIVEN** a configuration with user-specified volume mounts
- **WHEN** the provisioner generates `ansible-navigator.yml` on the target
- **THEN** the YAML SHALL include the volume-mounts list
- **AND** ansible-navigator on the target SHALL apply these mounts to the execution environment container
