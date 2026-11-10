## ADDED Requirements

### Requirement: Container Engine Configuration (local provisioner)

The on-target provisioner SHALL accept `container_engine` in the `navigator_config.execution_environment` block and preserve explicitly selected Docker or Podman engines in the configuration supplied to Navigator.

#### Scenario: Explicit engine survives HCL decoding and configuration generation

- **GIVEN** an execution environment block with `container_engine = "podman"`
- **WHEN** the provisioner decodes HCL and generates Navigator settings
- **THEN** the selected engine SHALL remain `podman`

### Requirement: Container Options Configuration (local provisioner)

The on-target provisioner SHALL accept an ordered string list in `container_options` and preserve user-specified options in the settings supplied to Navigator.

#### Scenario: Ordered container options

- **GIVEN** `container_options = ["--security-opt", "label=disable", "--network=host"]`
- **WHEN** HCL is decoded and Navigator settings are generated
- **THEN** all three values SHALL retain their order and content

### Requirement: Pull Arguments Configuration (local provisioner)

The on-target provisioner SHALL accept ordered `pull_arguments` independently of `pull_policy` and preserve them for Navigator's image pull operation.

#### Scenario: Arguments without policy

- **GIVEN** `pull_arguments = ["--quiet"]` and no pull policy
- **WHEN** the provisioner generates Navigator YAML
- **THEN** `ansible-navigator.execution-environment.pull.arguments` SHALL contain `["--quiet"]`
- **AND** the YAML pull object SHALL omit `policy`

#### Scenario: Arguments and policy together

- **GIVEN** `pull_arguments = ["--quiet"]` and `pull_policy = "always"`
- **WHEN** the provisioner generates effective Navigator settings
- **THEN** both settings SHALL be preserved
- **AND** any generated YAML pull arguments SHALL be nested under `execution-environment.pull.arguments`

#### Scenario: Policy without arguments

- **GIVEN** `pull_policy = "missing"` and no pull arguments
- **WHEN** the provisioner generates effective Navigator settings
- **THEN** the policy SHALL remain `missing`
- **AND** the YAML pull object, if generated, SHALL omit `arguments`

#### Scenario: Neither pull setting configured

- **GIVEN** neither pull policy nor pull arguments
- **WHEN** the provisioner generates Navigator YAML
- **THEN** the `pull` key SHALL be omitted

### Requirement: Additional Volume Mounts (local provisioner)

The on-target provisioner SHALL accept repeatable `volume_mounts` blocks and preserve user mounts alongside applicable automatic collections mounts without duplicating a mount source or collections destination.

#### Scenario: Repeatable mount block syntax

- **GIVEN** an execution environment containing:
  ```hcl
  volume_mounts {
    src     = "/runtime/data"
    dest    = "/container/data"
    options = "ro"
  }
  volume_mounts {
    src     = "/runtime/config"
    dest    = "/container/config"
    options = "ro"
  }
  ```
- **WHEN** the provisioner decodes the HCL
- **THEN** both mounts SHALL retain their values and order
- **AND** list assignment syntax `volume_mounts = [...]` SHALL be rejected

#### Scenario: Explicit collections mount takes precedence

- **GIVEN** a user mount targeting `/tmp/.packer_ansible/collections` and an applicable automatic collections path
- **WHEN** the provisioner adds automatic EE defaults repeatedly
- **THEN** the user mount SHALL be preserved
- **AND** a duplicate collections mount SHALL NOT be added

#### Scenario: Mounts refer to target paths

- **GIVEN** custom mounts and an applicable collections path on the target
- **WHEN** the provisioner generates its remote invocation
- **THEN** custom mounts SHALL be supplied with their target-side paths
- **AND** the collections mount SHALL coexist with custom mounts without duplication

## MODIFIED Requirements

### Requirement: Navigator Config File Generation

The provisioner MUST generate ansible-navigator configuration using CLI flags as the primary method. YAML configuration files MUST be uploaded to the target only when the configuration contains settings not mapped to CLI flags by this provisioner (PlaybookArtifact, CollectionDocCache, and execution-environment pull arguments).

#### Scenario: CLI-only configuration (common path)
- **GIVEN** NavigatorConfig contains mode, execution environment, and logging settings
- **AND** No PlaybookArtifact, CollectionDocCache, or pull arguments are configured
- **WHEN** Executing Provision()
- **THEN** CLI flags are generated for all settings
- **AND** No temporary YAML file is created or uploaded
- **AND** The remote ansible-navigator command includes flags like `--mode=stdout --execution-environment-image=... --pull-policy=never`
- **AND** No `--settings` flag is present in the remote command

#### Scenario: Hybrid configuration with playbook artifact
- **GIVEN** NavigatorConfig contains CLI-mappable settings (mode, execution environment)
- **AND** PlaybookArtifact is configured with `{ Enable: true, SaveAs: "/tmp/artifact.json" }`
- **WHEN** Executing Provision()
- **THEN** CLI flags are generated for mode, execution environment, and logging
- **AND** A minimal YAML file is created locally containing ONLY playbook-artifact settings
- **AND** The minimal YAML file is uploaded to the target staging directory
- **AND** The remote ansible-navigator command includes both CLI flags and `--settings=/staging/ansible-navigator.yml`

#### Scenario: Empty navigator configuration
- **GIVEN** NavigatorConfig is nil or unspecified
- **WHEN** Executing Provision()
- **THEN** No CLI flags are added for navigator configuration
- **AND** No YAML file is created or uploaded
- **AND** Default ansible-navigator behavior is used on the target

#### Scenario: Pull policy enforcement via CLI flag
- **GIVEN** NavigatorConfig.ExecutionEnvironment.PullPolicy is `"never"`
- **WHEN** Executing Provision() on the target
- **THEN** The remote command includes `--pull-policy=never`
- **AND** The target's container engine does NOT attempt to pull the image
- **AND** Execution uses existing local images on the target only

#### Scenario: Multiple environment variables via CLI flags
- **GIVEN** NavigatorConfig.ExecutionEnvironment.EnvironmentVariables.Set contains multiple variables
- **WHEN** Executing Provision()
- **THEN** Each variable is passed as a separate `--eev KEY=VALUE` flag in the remote command
- **AND** No YAML file is needed for environment variables

#### Scenario: Volume mounts via CLI flags
- **GIVEN** NavigatorConfig.ExecutionEnvironment.VolumeMounts contains multiple mount specifications
- **WHEN** Executing Provision()
- **THEN** Each mount is passed as a separate `--evm src:dest:options` flag in the remote command
- **AND** No YAML file is needed for volume mounts

#### Scenario: Minimal YAML upload only when needed
- **GIVEN** NavigatorConfig contains only CLI-mappable settings
- **WHEN** Preparing files to upload to target
- **THEN** The upload file list does NOT include ansible-navigator.yml
- **AND** Only playbook/role files and ansible.cfg are uploaded

#### Scenario: Minimal YAML cleanup after hybrid execution
- **GIVEN** NavigatorConfig contained unmapped settings requiring minimal YAML
- **AND** The minimal YAML file was created locally and uploaded
- **WHEN** Provision() completes successfully or with error
- **THEN** The local temporary minimal YAML file is removed
- **AND** The remote minimal YAML file remains in staging directory (cleaned by Packer's cleanup)


#### Scenario: Pull arguments trigger minimal YAML upload

- **GIVEN** an enabled execution environment with pull arguments and optionally a pull policy
- **WHEN** the provisioner uploads settings and invokes Navigator on the target
- **THEN** minimal YAML SHALL contain `ansible-navigator.execution-environment.pull.arguments` with the configured list
- **AND** mapped mode, engine, container options, mounts, and pull policy SHALL continue to use their existing CLI flags
- **AND** the remote invocation SHALL reference the uploaded settings file
- **AND** the local temporary YAML SHALL be removed after provisioning
