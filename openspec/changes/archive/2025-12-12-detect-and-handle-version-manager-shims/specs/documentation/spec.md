# Spec Delta: Documentation

## ADDED Requirements

### Requirement: Version Check Troubleshooting Documentation

The TROUBLESHOOTING.md documentation SHALL explain version check issues including automatic version manager shim detection and resolution, with clear guidance on when manual configuration is needed.

#### Scenario: Version manager shim documentation

- **WHEN** a user consults `docs/TROUBLESHOOTING.md` for version check hangs or timeout issues
- **THEN** the "Version check hangs / timeouts" section SHALL explain:
  - The plugin automatically detects and resolves shims for asdf, rbenv, and pyenv
  - Shim detection happens transparently during version check
  - Most shim configurations work automatically without manual intervention
  - Manual configuration is only needed when automatic resolution fails
- **AND** it SHALL include an example showing automatic shim resolution works by default

#### Scenario: Manual configuration still documented

- **WHEN** automatic shim resolution fails
- **THEN** the documentation SHALL provide manual configuration options:
  - Using `command` with full path to real binary
  - Using `ansible_navigator_path` to add directories containing the real binary
  - Using `skip_version_check = true` as last resort
- **AND** it SHALL explain these are fallback options when automatic detection fails

#### Scenario: Example configuration reflects shim support

- **WHEN** documentation shows example configurations for version managers
- **THEN** examples SHALL indicate shims now work automatically:

  ```hcl
  # Shims work automatically - no special configuration needed!
  # The plugin detects and resolves asdf/rbenv/pyenv shims automatically.
  
  provisioner "ansible-navigator" {
    play { target = "site.yml" }
  }
  
  # Only needed if automatic resolution fails:
  # provisioner "ansible-navigator" {
  #   ansible_navigator_path = ["~/.asdf/installs/ansible/X.Y.Z/bin"]
  #   play { target = "site.yml" }
  # }
  ```

- **AND** manual configuration SHALL be shown as optional/fallback

#### Scenario: No outdated workaround advice

- **WHEN** reviewing TROUBLESHOOTING.md
- **THEN** it SHALL NOT present manual shim configuration as the primary or required solution
- **AND** it SHALL NOT imply users must manually configure paths for version managers to work
- **AND** historical workaround advice SHALL be updated to reflect automatic detection
