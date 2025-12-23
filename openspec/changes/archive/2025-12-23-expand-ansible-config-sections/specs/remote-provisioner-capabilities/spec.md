## ADDED Requirements

### Requirement: REQ-ANSIBLE-CONFIG-SECTIONS-REMOTE-001 Additional ansible.cfg sections are supported via `navigator_config.ansible_config` (remote)

The SSH-based `ansible-navigator` provisioner SHALL accept additional ansible.cfg section blocks under `navigator_config.ansible_config` and preserve their values through to configuration artifact generation.

#### Scenario: Packer HCL decodes additional ansible_config section blocks

Given: a Packer template using `provisioner "ansible-navigator"` with `navigator_config.ansible_config` containing additional section blocks
When: Packer parses the configuration
Then: parsing MUST succeed without “Unsupported argument” errors for these documented blocks

#### Scenario: ansible.cfg contains additional sections derived from HCL blocks

Given: a configuration where `navigator_config.ansible_config` sets at least one of: `privilege_escalation`, `persistent_connection`, `inventory`, `paramiko_connection`, `colors`, `diff`, `galaxy`
When: the provisioner generates configuration artifacts
Then: it SHALL generate an ansible.cfg that includes a corresponding INI section for each configured block (e.g., `[privilege_escalation]`)
And: within each INI section, it SHALL emit keys that correspond to the configured HCL fields for that section
And: it SHALL serialize scalar values deterministically (string / number / boolean)

#### Scenario: ansible-navigator.yml remains schema-compliant while referencing ansible.cfg

Given: a configuration where `navigator_config.ansible_config` sets one or more additional section blocks
When: the provisioner generates the `ansible-navigator.yml`
Then: `ansible.config` in YAML MUST contain only allowed keys (`help`, `path`, and/or `cmdline`)
And: the generated YAML SHALL reference the generated ansible.cfg file path via `ansible.config.path`

### Requirement: REQ-ANSIBLE-CONFIG-GALAXY-SCOPE-REMOTE-001 ansible_config.galaxy affects Ansible runtime config only (remote)

The SSH-based provisioner SHALL treat `navigator_config.ansible_config.galaxy` as Ansible runtime configuration (ansible.cfg) and SHALL NOT conflate it with plugin-managed dependency installation behavior.

#### Scenario: galaxy configuration is emitted only to ansible.cfg

Given: a configuration that sets `navigator_config.ansible_config.galaxy` options
When: the provisioner generates configuration artifacts
Then: the configuration SHALL be rendered under the `[galaxy]` section of the generated ansible.cfg
And: it SHALL NOT change which `ansible-galaxy` subcommands or flags the plugin uses for dependency installation
