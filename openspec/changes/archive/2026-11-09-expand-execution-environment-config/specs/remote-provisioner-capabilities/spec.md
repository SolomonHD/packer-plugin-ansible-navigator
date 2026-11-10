## ADDED Requirements

### Requirement: Container Engine Configuration (remote provisioner)

The SSH-based provisioner SHALL accept `container_engine` in the `navigator_config.execution_environment` block and preserve explicitly selected Docker or Podman engines in the configuration supplied to Navigator.

#### Scenario: Explicit engine survives HCL decoding and configuration generation

- **GIVEN** an execution environment block with `container_engine = "podman"`
- **WHEN** the provisioner decodes HCL and generates Navigator settings
- **THEN** the selected engine SHALL remain `podman`

#### Scenario: Automatic engine resolution

- **GIVEN** an omitted or `auto` engine with an enabled EE
- **WHEN** the provisioner prepares its proxy topology
- **THEN** it SHALL retain the established Podman-before-Docker resolution and supply the resolved engine to Navigator

### Requirement: Container Options Configuration (remote provisioner)

The SSH-based provisioner SHALL accept an ordered string list in `container_options` and preserve user-specified options in the settings supplied to Navigator.

#### Scenario: Ordered container options

- **GIVEN** `container_options = ["--security-opt", "label=disable", "--network=host"]`
- **WHEN** HCL is decoded and Navigator settings are generated
- **THEN** all three values SHALL retain their order and content

#### Scenario: Automatic Docker options coexist with user options

- **GIVEN** an enabled Docker stdout EE using bridge networking
- **WHEN** the provisioner applies its automatic defaults
- **THEN** the established host-gateway mapping SHALL be supplied once
- **AND** `--tty=false` SHALL be supplied only when no explicit TTY choice exists
- **AND** user options SHALL be preserved

### Requirement: Pull Arguments Configuration (remote provisioner)

The SSH-based provisioner SHALL accept ordered `pull_arguments` independently of `pull_policy` and preserve them for Navigator's image pull operation.

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

### Requirement: Additional Volume Mounts (remote provisioner)

The SSH-based provisioner SHALL accept repeatable `volume_mounts` blocks and preserve user mounts alongside applicable automatic collections mounts without duplicating a mount source or collections destination.

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

#### Scenario: Missing unused default cache

- **GIVEN** a missing plugin-default collections path, no requirements file, and user volume mounts
- **WHEN** the provisioner generates launch settings
- **THEN** the user mounts SHALL remain present
- **AND** the optional default collections mount SHALL be omitted
- **AND** explicit paths, existing caches, and requirements-backed caches SHALL retain their established mount behavior
