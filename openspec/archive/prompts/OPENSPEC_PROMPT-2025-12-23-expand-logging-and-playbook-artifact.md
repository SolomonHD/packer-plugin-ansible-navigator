# OpenSpec Change Prompt: Expand Logging and Playbook Artifact Configuration

## Context

The plugin currently provides basic logging and playbook-artifact support in [`navigator_config`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:1), but only covers a small subset of the options documented in ansible-navigator v3.x. For logging, only basic level/file/append are supported. For playbook-artifact, only enable/save-as are available. Users cannot configure additional ansible-navigator logging options (like log-level, log-append) or playbook-artifact options (like replay options).

## Goal

Expand the [`LoggingConfig`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:1) and [`PlaybookArtifact`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:1) structs to support all documented ansible-navigator v3.x logging and playbook-artifact configuration options.

## Scope

**In scope:**
- Expand `LoggingConfig` struct with all ansible-navigator v3.x logging options:
  - Current: `level`, `file`, `append`
  - Add missing: any additional logging options from ansible-navigator v3.x schema
- Expand `PlaybookArtifact` struct with all ansible-navigator v3.x playbook-artifact options:
  - Current: `enable`, `save_as`
  - Add missing: `replay` options and any other playbook-artifact fields from v3.x schema
- Update `//go:generate packer-sdc mapstructure-to-hcl2` type list if new nested types are added
- Regenerate HCL2 spec files
- Update YAML generation to emit all logging and playbook-artifact fields correctly
- Add tests for HCL decoding and YAML generation

**Out of scope:**
- Expanding execution-environment (change #01)
- Expanding ansible-config (change #02)
- Expanding other top-level navigator_config sections like mode-settings, format, etc. (change #04)
- Creating/managing artifact files (ansible-navigator does that)

## Desired Behavior

A user can configure all documented ansible-navigator v3.x logging and playbook-artifact options:

```hcl
provisioner "ansible-navigator" {
  navigator_config {
    logging {
      level  = "debug"
      file   = "/tmp/navigator.log"
      append = true
    }
    
    playbook_artifact {
      enable  = true
      save_as = "/tmp/artifact-{playbook_name}.json"
      replay  = "previous-artifact.json"
    }
  }
}
```

The generated [`ansible-navigator.yml`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:1) correctly includes:

```yaml
logging:
  level: debug
  file: /tmp/navigator.log
  append: true

playbook-artifact:
  enable: true
  save-as: /tmp/artifact-{playbook_name}.json
  replay: previous-artifact.json
```

## Constraints & Assumptions

- **Assumption:** ansible-navigator v3.x documentation is the authoritative source for available options.
- **Constraint:** HCL field names use underscores (`save_as`), YAML keys use hyphens (`save-as`).
- **Constraint:** Preserve backward compatibility - existing logging and playbook_artifact configs must continue to work.
- **Constraint:** Use typed structs (no dynamic maps).
- **Constraint:** The logging configuration in ansible-navigator may have multiple fields for controlling log behavior - ensure all are captured.
- **Constraint:** Playbook artifact replay functionality may have complex options - research ansible-navigator v3.x schema to understand correct structure.

## Acceptance Criteria

- [ ] [`LoggingConfig`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:1) struct includes all ansible-navigator v3.x logging options.
- [ ] [`PlaybookArtifact`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:1) struct includes all ansible-navigator v3.x playbook-artifact options.
- [ ] HCL configuration with complete logging and playbook-artifact blocks decodes without error.
- [ ] Generated [`ansible-navigator.yml`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:1) includes correct key names (hyphenated where needed).
- [ ] A minimal Packer template using expanded logging and playbook-artifact options passes `packer validate`.
- [ ] Existing logging and playbook_artifact configurations continue to work unchanged.
- [ ] [`make generate`](packer/plugins/packer-plugin-ansible-navigator/GNUmakefile:1) regenerates HCL2 specs successfully.
- [ ] [`go build ./...`](packer/plugins/packer-plugin-ansible-navigator/GNUmakefile:1) compiles without errors.
- [ ] Unit tests validate HCL decoding and YAML structure for logging and playbook-artifact.

## Files Expected to Change

- [`provisioner/ansible-navigator/navigator_config.go`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:1) (expand structs, YAML generation)
- [`provisioner/ansible-navigator-local/navigator_config.go`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator-local/navigator_config.go:1) (same structs)
- [`provisioner/ansible-navigator/*.hcl2spec.go`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/*.hcl2spec.go:1) (generated)
- [`provisioner/ansible-navigator-local/*.hcl2spec.go`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator-local/*.hcl2spec.go:1) (generated)
- Test files (new tests)
- Possibly [`example/`](packer/plugins/packer-plugin-ansible-navigator/example/) templates

## Dependencies

- **None required** (can be done independently)
- **Suggested order:** After #01 and #02 for logical progression
