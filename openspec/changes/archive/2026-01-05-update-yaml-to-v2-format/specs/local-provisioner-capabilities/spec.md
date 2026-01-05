# Spec Delta: local-provisioner-capabilities

## MODIFIED Requirements

### Requirement: Navigator Config File Generation

The local provisioner SHALL support generating ansible-navigator.yml configuration files from a declarative HCL map, uploading them to the target, and using them to control ansible-navigator behavior. The generated YAML SHALL conform to the ansible-navigator **Version 2 format** (as recognized by ansible-navigator 25.x+), including proper nested structure for execution environment fields and any required version markers or schema identifiers.

#### Scenario: User-specified navigator_config generates Version 2 format YAML file

- **GIVEN** a configuration with:

  ```hcl
  navigator_config {
    mode = "stdout"
    execution_environment {
      enabled     = true
      image       = "quay.io/ansible/creator-ee:latest"
      pull_policy = "missing"
    }
    ansible_config {
      defaults {
        remote_tmp = "/tmp/.ansible/tmp"
      }
    }
  }
  ```

- **WHEN** the provisioner prepares for execution
- **THEN** it SHALL generate a temporary file named `/tmp/packer-navigator-cfg-<uuid>.yml` (or equivalent in system temp directory) on the LOCAL machine
- **AND** the file SHALL be in ansible-navigator **Version 2 format** (as defined by ansible-navigator 25.x documentation)
- **AND** ansible-navigator 25.x SHALL recognize the file as Version 2 format **without** triggering version migration prompts
- **AND** the file SHALL contain valid YAML conforming to the ansible-navigator.yml v25+ schema
- **AND** the execution-environment pull policy SHALL be generated as nested structure: `pull: { policy: missing }`
- **AND** the file SHALL NOT use flat `pull-policy: missing` syntax (which is rejected by ansible-navigator v25+)
- **AND** if Version 2 format requires a version marker or schema identifier, it SHALL be included
- **AND** the file SHALL be uploaded to the staging directory on the TARGET machine
- **AND** the local temporary file SHALL be added to the cleanup list

#### Scenario: pull_policy = "never" prevents Docker image pulls on target

- **GIVEN** a configuration with:

  ```hcl
  navigator_config {
    execution_environment {
      enabled     = true
      image       = "my-local-image:latest"
      pull_policy = "never"
    }
  }
  ```

- **AND** the image "my-local-image:latest" exists locally on the TARGET machine's Docker
- **WHEN** the provisioner executes on the target
- **THEN** ansible-navigator SHALL use the local image **without** attempting to pull from any registry
- **AND** no "Unable to find image locally" errors SHALL occur on the target
- **AND** no Docker registry connection attempts SHALL be made from the target

#### Scenario: Generated YAML includes Version 2 format markers

- **GIVEN** a minimal `navigator_config { mode = "stdout" }` configuration
- **WHEN** the provisioner generates the ansible-navigator.yml file
- **THEN** the file SHALL include all structural elements required by ansible-navigator Version 2 format
- **AND** the file SHALL be immediately recognized as Version 2 by ansible-navigator 25.12.0+ when executed on the target
- **AND** no migration prompts or warnings SHALL appear when ansible-navigator runs on the target
