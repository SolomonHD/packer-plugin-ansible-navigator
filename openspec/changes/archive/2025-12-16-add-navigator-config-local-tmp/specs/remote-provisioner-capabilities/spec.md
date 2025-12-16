## ADDED Requirements

### Requirement: `ansible_config.defaults.local_tmp` support (remote provisioner)

The SSH-based provisioner SHALL support configuring Ansible's local temp directory via `navigator_config.ansible_config.defaults.local_tmp`.

#### Scenario: HCL schema accepts local_tmp under ansible_config.defaults

- **GIVEN** a configuration using `provisioner "ansible-navigator"` with:

  ```hcl
  navigator_config {
    ansible_config {
      defaults {
        remote_tmp = "/tmp/.ansible/tmp"
        local_tmp  = "/tmp/.ansible-local"
      }
    }
  }
  ```

- **WHEN** Packer parses the configuration
- **THEN** parsing SHALL succeed
- **AND** the parsed configuration SHALL preserve the `local_tmp` value

#### Scenario: local_tmp is written to generated ansible.cfg and referenced from YAML

- **GIVEN** a configuration using `provisioner "ansible-navigator"` that sets `navigator_config.ansible_config.defaults.local_tmp`
- **WHEN** the provisioner prepares for execution and generates configuration artifacts
- **THEN** it SHALL generate an ansible.cfg that includes `local_tmp` under `[defaults]`
- **AND** the generated `ansible-navigator.yml` SHALL reference the generated ansible.cfg via `ansible.config.path`
- **AND** the generated `ansible-navigator.yml` MUST NOT emit `defaults` under `ansible.config` (schema-compliance requirement)

#### Scenario: local_tmp omitted when unset

- **GIVEN** a configuration that does not set `navigator_config.ansible_config.defaults.local_tmp`
- **WHEN** the provisioner generates ansible.cfg
- **THEN** the generated ansible.cfg SHALL NOT include a `local_tmp` entry

