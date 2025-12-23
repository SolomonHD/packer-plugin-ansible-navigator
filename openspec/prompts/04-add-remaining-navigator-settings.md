# OpenSpec Change Prompt: Add Remaining Navigator Settings

## Context

The plugin currently supports only a subset of top-level ansible-navigator v3.x configuration options in [`navigator_config`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:1). While execution-environment, ansible-config, logging, and playbook-artifact have been expanded in previous changes (#01-#03), many other documented v3.x top-level settings remain unsupported, including `mode-settings`, `format`, `color`, `images`, `time-zone`, `documentation`, `editor`, `collection-doc-cache` (currently has minimal support), and others.

## Goal

Add all remaining documented ansible-navigator v3.x top-level configuration options to [`NavigatorConfig`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:1), completing full configuration parity with the ansible-navigator.yml schema.

## Scope

**In scope:**
- Add missing top-level navigator_config fields and nested structs:
  - `mode_settings` (configuration per mode: stdout, interactive)
  - `format` (output format configuration)
  - `color` (color scheme settings)
  - `images` (container image preferences)
  - `time_zone` (timezone for timestamps)
  - `documentation` (documentation source URLs)
  - `editor` (preferred editor settings)
  - Expand `collection_doc_cache` (currently minimal, may need additional fields)
  - `inventory_columns` (inventory display columns)
  - `replay` (replay file options at top level if different from playbook-artifact)
  - Any other documented v3.x top-level options not covered by previous changes
- Create typed structs for any nested configuration sections (e.g., ModeSettings, ColorConfig, etc.)
- Update `//go:generate packer-sdc mapstructure-to-hcl2` type list with all new struct types
- Regenerate HCL2 spec files
- Update YAML generation to emit all new fields correctly
- Add tests for HCL decoding and YAML generation
- Update README/docs with examples of new options

**Out of scope:**
- Execution-environment expansion (change #01)
- Ansible-config expansion (change #02)
- Logging and playbook-artifact expansion (change #03)
- Validating option values (ansible-navigator does that)
- Implementing ansible-navigator behavior (we only generate config)

## Desired Behavior

A user can configure any remaining documented ansible-navigator v3.x top-level options:

```hcl
provisioner "ansible-navigator" {
  navigator_config {
    mode = "stdout"
    
    mode_settings {
      stdout {
        show_task_path = true
      }
      interactive {
        command_timeout = 30
      }
    }
    
    format {
      json_indent = 2
    }
    
    color {
      enable = true
      osc4   = true
    }
    
    images {
      default = "quay.io/ansible/creator-ee:latest"
    }
    
    time_zone = "UTC"
    
    editor {
      command = "vim"
      console = false
    }
    
    collection_doc_cache {
      path    = "/tmp/.cache/ansible-navigator"
      timeout = 3600
    }
    
    inventory_columns = ["name", "ansible_host", "groups"]
  }
}
```

The generated [`ansible-navigator.yml`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:1) includes:

```yaml
mode: stdout

mode-settings:
  stdout:
    show-task-path: true
  interactive:
    command-timeout: 30

format:
  json-indent: 2

color:
  enable: true
  osc4: true

images:
  default: quay.io/ansible/creator-ee:latest

time-zone: UTC

editor:
  command: vim
  console: false

collection-doc-cache:
  path: /tmp/.cache/ansible-navigator
  timeout: 3600

inventory-columns:
  - name
  - ansible_host
  - groups
```

## Constraints & Assumptions

- **Assumption:** ansible-navigator v3.x documentation is the authoritative source.
- **Constraint:** HCL field names use underscores, YAML keys use hyphens where documented.
- **Constraint:** Some settings may be complex nested structures (e.g., mode_settings has stdout/interactive subsections) - use typed structs.
- **Constraint:** The `mode` field already exists at top level - ensure new mode_settings doesn't conflict.
- **Constraint:** Preserve backward compatibility - existing minimal navigator_config must continue to work.
- **Constraint:** Document which v3.x options are supported clearly (it's acceptable to support most common options initially, then expand later if needed).
- **Constraint:** Use typed structs for all nested configuration (no dynamic maps).

## Acceptance Criteria

- [ ] [`NavigatorConfig`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:1) struct includes all remaining ansible-navigator v3.x top-level options (mode-settings, format, color, images, time-zone, editor, collection-doc-cache expansion, inventory-columns).
- [ ] HCL configuration with new top-level settings decodes without error.
- [ ] Generated [`ansible-navigator.yml`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:1) includes correct key names (hyphenated where needed) and nesting.
- [ ] A comprehensive Packer template using representative new options passes `packer validate`.
- [ ] Existing minimal navigator_config configurations continue to work unchanged.
- [ ] [`make generate`](packer/plugins/packer-plugin-ansible-navigator/GNUmakefile:1) regenerates HCL2 specs successfully.
- [ ] [`go build ./...`](packer/plugins/packer-plugin-ansible-navigator/GNUmakefile:1) compiles without errors.
- [ ] Unit tests validate HCL decoding and YAML structure for new settings.
- [ ] [`README.md`](packer/plugins/packer-plugin-ansible-navigator/README.md:1) or docs updated with examples of new configuration options.

## Files Expected to Change

- [`provisioner/ansible-navigator/navigator_config.go`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/navigator_config.go:1) (add structs, YAML generation)
- [`provisioner/ansible-navigator-local/navigator_config.go`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator-local/navigator_config.go:1) (same structs)
- [`provisioner/ansible-navigator/*.hcl2spec.go`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator/*.hcl2spec.go:1) (generated)
- [`provisioner/ansible-navigator-local/*.hcl2spec.go`](packer/plugins/packer-plugin-ansible-navigator/provisioner/ansible-navigator-local/*.hcl2spec.go:1) (generated)
- Test files (comprehensive new tests)
- [`README.md`](packer/plugins/packer-plugin-ansible-navigator/README.md:1) or [`docs/`](packer/plugins/packer-plugin-ansible-navigator/docs/) (examples, configuration reference)
- [`example/`](packer/plugins/packer-plugin-ansible-navigator/example/) templates (showcase new options)

## Dependencies

- **Recommended order:** After changes #01, #02, and #03 (execution-environment, ansible-config, logging/playbook-artifact)
- This change completes the full navigator_config v3.x parity effort
