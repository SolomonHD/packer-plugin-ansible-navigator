## ADDED Requirements

### Requirement: REQ-NAVCFG-TOPLEVEL-LOCAL-001 Expand `navigator_config` top-level schema support (local)

The on-target `ansible-navigator-local` provisioner SHALL support the remaining documented ansible-navigator v3.x **top-level** configuration options in `navigator_config`, including at minimum:

- `mode_settings` (YAML: `mode-settings`)
- `format` (YAML: `format`)
- `color` (YAML: `color`)
- `images` (YAML: `images`)
- `time_zone` (YAML: `time-zone`)
- `documentation` (YAML: `documentation`)
- `editor` (YAML: `editor`)
- `inventory_columns` (YAML: `inventory-columns`)
- `replay` (YAML: `replay`) when present as a top-level key in the v3.x schema
- Expanded `collection_doc_cache` (YAML: `collection-doc-cache`) beyond `path`/`timeout` as required by the v3.x schema

#### Scenario: Packer HCL decodes remaining top-level settings

Given: a Packer template using `provisioner "ansible-navigator-local"`
And: the template includes a `navigator_config { ... }` block containing supported v3.x top-level settings (e.g., `mode_settings`, `format`, `color`, `images`, `time_zone`, `documentation`, `editor`, `inventory_columns`, `replay`)
When: Packer parses the configuration
Then: parsing MUST succeed without “Unsupported argument” errors for those settings

#### Scenario: Generated YAML contains remaining top-level settings with correct key names

Given: a configuration using `provisioner "ansible-navigator-local"` with the supported top-level settings configured under `navigator_config`
When: the provisioner generates `ansible-navigator.yml`
Then: the YAML MUST include the configured settings under the `ansible-navigator:` root
And: any keys that are hyphenated in ansible-navigator.yml MUST be emitted with hyphens in YAML (e.g., `mode-settings`, `time-zone`, `inventory-columns`, `collection-doc-cache`)
And: the YAML MUST reflect the configured values without renaming user intent

#### Scenario: Existing navigator_config sections remain supported

Given: a configuration using `provisioner "ansible-navigator-local"` with only previously-supported `navigator_config` sections (e.g., `execution_environment`, `ansible_config`, `logging`, `playbook_artifact`, minimal `collection_doc_cache`)
When: Packer parses the configuration and the provisioner generates `ansible-navigator.yml`
Then: parsing and YAML generation MUST continue to succeed
