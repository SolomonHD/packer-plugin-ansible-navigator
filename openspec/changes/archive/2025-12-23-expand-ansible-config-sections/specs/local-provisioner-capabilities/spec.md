## ADDED Requirements

### Requirement: REQ-ANSIBLE-CONFIG-SECTIONS-LOCAL-001 Additional ansible.cfg sections are supported via `navigator_config.ansible_config` (local)

The on-target `ansible-navigator-local` provisioner SHALL accept additional ansible.cfg section blocks under `navigator_config.ansible_config` and preserve their values through to configuration artifact generation.

#### Scenario: Packer HCL decodes additional ansible_config section blocks

Given: a Packer template using `provisioner "ansible-navigator-local"` with `navigator_config.ansible_config` containing additional section blocks
When: Packer parses the configuration
Then: parsing MUST succeed without “Unsupported argument” errors for these documented blocks

#### Scenario: ansible.cfg contains additional sections derived from HCL blocks

Given: a configuration where `navigator_config.ansible_config` sets at least one of: `privilege_escalation`, `persistent_connection`, `inventory`, `paramiko_connection`, `colors`, `diff`, `galaxy`
When: the provisioner generates configuration artifacts
Then: it SHALL generate an ansible.cfg that includes a corresponding INI section for each configured block
And: it SHALL ensure ansible-navigator on the target uses that file via `ansible.config.path` in the generated `ansible-navigator.yml`

### Requirement: REQ-ANSIBLE-CONFIG-GALAXY-SCOPE-LOCAL-001 ansible_config.galaxy affects Ansible runtime config only (local)

The on-target provisioner SHALL treat `navigator_config.ansible_config.galaxy` as Ansible runtime configuration (ansible.cfg) and SHALL NOT conflate it with plugin-managed dependency installation behavior.

#### Scenario: galaxy configuration is emitted only to ansible.cfg

Given: a configuration that sets `navigator_config.ansible_config.galaxy` options
When: the provisioner generates configuration artifacts
Then: the configuration SHALL be rendered under the `[galaxy]` section of the generated ansible.cfg
And: it SHALL NOT change which `ansible-galaxy` subcommands or flags the plugin uses for dependency installation
