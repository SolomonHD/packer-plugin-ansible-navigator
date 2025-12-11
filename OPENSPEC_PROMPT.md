# OpenSpec change prompt

## Context

The `ansible_navigator_path` configuration field was added to the Config struct in `provisioner/ansible-navigator/provisioner.go` but the HCL2 spec files were not regenerated. Packer silently ignores HCL attributes that don't exist in the generated `.hcl2spec.go` files.

## Goal

Regenerate all HCL2 spec files so that every Config struct field (including `ansible_navigator_path`) is recognized by Packer.

## Scope

- In scope:
  - Running `make generate` to regenerate `.hcl2spec.go` files
  - Verifying the regenerated files include all Config fields
- Out of scope:
  - Adding new configuration fields
  - Modifying existing Go code
  - Changing plugin behavior

## Desired behavior

- The `FlatConfig` struct in `provisioner.hcl2spec.go` includes all fields from the `Config` struct
- The `HCL2Spec()` function returns specs for all fields including `ansible_navigator_path`
- Users can set `ansible_navigator_path` in Packer HCL without it being silently ignored

## Constraints & assumptions

- Assume the `Config` struct is correctly defined with proper `mapstructure` tags
- The `//go:generate` directive already exists in provisioner.go
- Both provisioners (ansible-navigator and ansible-navigator-local) may need regeneration

## Acceptance criteria

- [ ] `grep "AnsibleNavigatorPath" provisioner/ansible-navigator/provisioner.hcl2spec.go` returns matches
- [ ] `go build ./...` succeeds without errors
- [ ] `make plugin-check` passes
- [ ] All fields in Config struct appear in corresponding FlatConfig struct
