## ADDED Requirements

### Requirement: REQ-DOC-ANSIBLE-CONFIG-SECTIONS-001 Document additional `ansible_config` section blocks

User-facing documentation SHALL include an example showing how to configure at least one additional ansible.cfg section through `navigator_config.ansible_config`.

#### Scenario: Documentation includes examples for additional sections

Given: a user is reading configuration documentation for `navigator_config.ansible_config`
When: they look for non-default Ansible configuration examples
Then: the docs SHALL include an example that uses at least one additional section block (e.g., `privilege_escalation` or `inventory`)
And: the docs SHALL link to the upstream Ansible ansible.cfg reference for comprehensive option meanings
