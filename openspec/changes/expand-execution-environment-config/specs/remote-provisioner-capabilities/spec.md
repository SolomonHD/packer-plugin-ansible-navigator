## ADDED Requirements

### Requirement: Container Engine Configuration (remote provisioner)

The SSH-based provisioner SHALL support specifying the container engine used by ansible-navigator for execution environments.

#### Scenario: container_engine field accepted in HCL

- **GIVEN** a configuration with `provisioner "ansible-navigator"`
- **AND** `navigator_config.execution_environment.container_engine = "docker"`
- **WHEN** Packer parses the configuration
- **THEN** parsing SHALL succeed
- **AND** the configuration SHALL preserve the `container_engine` value

#### Scenario: container_engine generated in YAML

- **GIVEN** a configuration with `navigator_config.execution_environment.container_engine = "podman"`
- **WHEN** the provisioner generates `ansible-navigator.yml`
- **THEN** the YAML SHALL include under `ansible-navigator.execution-environment`:
  ```yaml
  container-engine: podman
  ```
- **AND** ansible-navigator SHALL use the specified container runtime

#### Scenario: container_engine omitted when not configured

- **GIVEN** a configuration that does not set `container_engine`
- **WHEN** the provisioner generates `ansible-navigator.yml`
- **THEN** the `container-engine` key SHALL NOT appear in the YAML
- **AND** ansible-navigator SHALL use its default container engine selection

### Requirement: Container Options Configuration (remote provisioner)

The SSH-based provisioner SHALL support passing arbitrary container runtime flags via `container_options`.

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

#### Scenario: container_options generated in YAML

- **GIVEN** a configuration with `container_options = ["--net=host", "--cap-add=SYS_ADMIN"]`
- **WHEN** the provisioner generates `ansible-navigator.yml`
- **THEN** the YAML SHALL include under `ansible-navigator.execution-environment`:
  ```yaml
  container-options:
    - --net=host
    - --cap-add=SYS_ADMIN
  ```
- **AND** ansible-navigator SHALL pass these options verbatim to the container runtime

#### Scenario: container_options omitted when not configured

- **GIVEN** a configuration that does not set `container_options`
- **WHEN** the provisioner generates `ansible-navigator.yml`
- **THEN** the `container-options` key SHALL NOT appear in the YAML
- **AND** no additional container options SHALL be applied

### Requirement: Pull Arguments Configuration (remote provisioner)

The SSH-based provisioner SHALL support passing arguments to the container image pull command via `pull_arguments`.

#### Scenario: pull_arguments accepted in HCL

- **GIVEN** a configuration with:
  ```hcl
  navigator_config {
    execution_environment {
      pull_arguments = [
        "--creds", "username:password",
        "--tls-verify=false"
      ]
    }
  }
  ```
- **WHEN** Packer parses the configuration
- **THEN** parsing SHALL succeed
- **AND** all pull argument values SHALL be preserved

#### Scenario: pull_arguments nested under pull in YAML

- **GIVEN** a configuration with `pull_arguments = ["--tls-verify=false"]`
- **AND** `pull_policy = "missing"`
- **WHEN** the provisioner generates `ansible-navigator.yml`
- **THEN** the YAML SHALL include under `ansible-navigator.execution-environment`:
  ```yaml
  pull:
    policy: missing
    arguments:
      - --tls-verify=false
  ```
- **AND** both pull.policy and pull.arguments SHALL be nested under the same `pull` parent key

#### Scenario: pull struct created when only pull_arguments provided

- **GIVEN** a configuration with `pull_arguments` set but `pull_policy` NOT set
- **WHEN** the provisioner generates `ansible-navigator.yml`
- **THEN** the YAML SHALL include:
  ```yaml
  pull:
    arguments:
      - <configured arguments>
  ```
- **AND** the `policy` field SHALL be omitted (ansible-navigator uses default)

#### Scenario: pull_arguments omitted when not configured

- **GIVEN** a configuration that does not set `pull_arguments`
- **WHEN** the provisioner generates `ansible-navigator.yml`
- **THEN** if `pull_policy` is set, YAML SHALL include `pull: { policy: <value> }`
- **AND** if `pull_policy` is NOT set, the `pull` key SHALL be omitted entirely

### Requirement: Additional Volume Mounts (remote provisioner)

The SSH-based provisioner SHALL allow users to specify additional volume mounts beyond the automatic collections mount.

#### Scenario: User-specified volume mounts preserved alongside automatic mounts

- **GIVEN** a configuration with:
  ```hcl
  requirements_file = "./requirements.yml"
  navigator_config {
    execution_environment {
      enabled = true
      volume_mounts = [
        {
          src     = "/host/custom"
          dest    = "/container/custom"
          options = "ro"
        }
      ]
    }
  }
  ```
- **WHEN** the provisioner applies automatic EE defaults (collections mount)
- **THEN** the final volume mounts list SHALL include BOTH:
  - User-specified mount: `/host/custom:/container/custom:ro`
  - Automatic collections mount: `<collections_path>:/tmp/.packer_ansible/collections:ro`
- **AND** the generated YAML SHALL include both mounts under `volume-mounts`

#### Scenario: Duplicate mount detection prevents conflicts

- **GIVEN** a configuration where user manually specifies a collections mount with the same destination as the automatic mount
- **WHEN** the provisioner applies automatic EE defaults
- **THEN** it SHALL detect the duplicate destination path
- **AND** it SHALL NOT add a second mount to the same destination
- **AND** the user's explicit mount SHALL take precedence

#### Scenario: Volume mounts work without collections

- **GIVEN** a configuration with `volume_mounts` but NO `requirements_file`
- **WHEN** the provisioner generates `ansible-navigator.yml`
- **THEN** the YAML SHALL include only the user-specified volume mounts
- **AND** no automatic collections mount SHALL be added

### Requirement: Complete Pull Configuration Structure (remote provisioner)

The YAML generation SHALL correctly nest both `pull.policy` and `pull.arguments` under a single `pull` parent key to match ansible-navigator v3.x schema.

#### Scenario: Pull policy and arguments combined in YAML

- **GIVEN** a configuration with both `pull_policy = "always"` AND `pull_arguments = ["--tls-verify=false"]`
- **WHEN** the provisioner generates `ansible-navigator.yml`
- **THEN** the YAML SHALL be structured as:
  ```yaml
  ansible-navigator:
    execution-environment:
      pull:
        policy: always
        arguments:
          - --tls-verify=false
  ```
- **AND** ansible-navigator SHALL accept this structure without errors

#### Scenario: Backward compatibility for pull_policy only

- **GIVEN** a configuration with only `pull_policy = "missing"` (no `pull_arguments`)
- **WHEN** the provisioner generates `ansible-navigator.yml`
- **THEN** the YAML SHALL include:
  ```yaml
  ansible-navigator:
    execution-environment:
      pull:
        policy: missing
  ```
- **AND** existing configurations SHALL continue to work without modification
