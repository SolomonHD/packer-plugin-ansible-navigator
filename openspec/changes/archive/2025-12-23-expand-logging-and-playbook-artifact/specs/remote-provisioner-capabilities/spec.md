## ADDED Requirements

### Requirement: REQ-NAVCFG-LOG-REMOTE-001 Expanded `navigator_config.logging` schema support (remote)

The SSH-based `ansible-navigator` provisioner SHALL support all documented ansible-navigator configuration options under `navigator_config.logging` for the supported schema baseline.

#### Scenario: Packer HCL decodes documented logging options

Given: a Packer template using `provisioner "ansible-navigator"`
And: the template includes a `navigator_config { logging { ... } }` block containing any documented ansible-navigator logging options for the supported schema baseline
When: Packer parses the configuration
Then: parsing MUST succeed without “Unsupported argument” errors for those logging options

#### Scenario: Generated YAML contains logging configuration with correct key names

Given: a configuration using `provisioner "ansible-navigator"` with `navigator_config.logging` configured
When: the provisioner generates `ansible-navigator.yml`
Then: the YAML MUST include the `logging:` section under the `ansible-navigator:` root
And: any keys that are hyphenated in ansible-navigator.yml MUST be emitted with hyphens in YAML
And: the YAML MUST reflect the configured logging values without renaming user intent

### Requirement: REQ-NAVCFG-ARTIFACT-REMOTE-001 Expanded `navigator_config.playbook_artifact` schema support (remote)

The SSH-based `ansible-navigator` provisioner SHALL support all documented ansible-navigator configuration options under `navigator_config.playbook_artifact` for the supported schema baseline.

#### Scenario: Packer HCL decodes documented playbook-artifact options

Given: a Packer template using `provisioner "ansible-navigator"`
And: the template includes a `navigator_config { playbook_artifact { ... } }` block containing any documented ansible-navigator playbook-artifact options for the supported schema baseline
When: Packer parses the configuration
Then: parsing MUST succeed without “Unsupported argument” errors for those playbook-artifact options

#### Scenario: Generated YAML contains playbook-artifact configuration with correct key names

Given: a configuration using `provisioner "ansible-navigator"` with `navigator_config.playbook_artifact` configured
When: the provisioner generates `ansible-navigator.yml`
Then: the YAML MUST include the `playbook-artifact:` section under the `ansible-navigator:` root
And: the YAML MUST emit `save-as` (and any other hyphenated keys) using hyphenated YAML key names
And: the YAML MUST reflect the configured playbook-artifact values
