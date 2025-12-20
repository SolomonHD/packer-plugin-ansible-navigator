# OpenSpec Change Prompt

## Context

The packer-plugin-ansible-navigator provisioner constructs command-line arguments for `ansible-navigator run` using Go's `exec.Command`. While Go bypasses shell interpretation, ansible-navigator internally invokes `ansible-playbook` which may re-parse arguments. Using separate flag/value arguments (e.g., `--flag value`) causes parsing failures when values start with `-` or contain spaces, as they get interpreted as new flags.

## Goal

Standardize all command-line argument construction to use `--flag=value` syntax instead of `--flag value` (two separate arguments) to prevent argument parsing issues when ansible-navigator passes arguments to ansible-playbook inside execution environments.

## Scope

**In scope:**
- `provisioner/ansible-navigator/provisioner.go`: All `args = append(args, ...)` calls that construct ansible-navigator/ansible-playbook arguments
- `provisioner/ansible-navigator/galaxy.go`: All `args = append(args, ...)` calls for ansible-galaxy commands
- `provisioner/ansible-navigator-local/provisioner.go`: All argument construction
- `provisioner/ansible-navigator-local/galaxy.go`: All argument construction

**Out of scope:**
- Positional arguments (playbook path, inventory file) - these must remain separate
- Boolean flags without values (e.g., `--become`, `--force`, `--offline`)
- Environment variable handling
- Config file generation (ansible.cfg, ansible-navigator.yml)

## Desired Behavior

- All flag+value pairs use `--flag=value` syntax as a single argument
- Boolean flags remain standalone (no change needed)
- Positional arguments remain as separate array elements
- Command execution behavior is unchanged; only argument formatting differs

## Examples

**Before (vulnerable to parsing issues):**
```go
args = append(args, "--ssh-extra-args", "-o IdentitiesOnly=yes")
args = append(args, "--extra-vars", fmt.Sprintf("@%s", extraVarsFilePath))
args = append(args, "--mode", p.config.NavigatorConfig.Mode)
args = append(args, "--limit", p.config.Limit)
args = append(args, "-i", inventory)
args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
args = append(args, "-p", rolesPath)
```

**After (robust):**
```go
args = append(args, fmt.Sprintf("--ssh-extra-args=-o IdentitiesOnly=yes"))
args = append(args, fmt.Sprintf("--extra-vars=@%s", extraVarsFilePath))
args = append(args, fmt.Sprintf("--mode=%s", p.config.NavigatorConfig.Mode))
args = append(args, fmt.Sprintf("--limit=%s", p.config.Limit))
args = append(args, fmt.Sprintf("-i=%s", inventory))
args = append(args, fmt.Sprintf("-e=%s=%s", k, v))
args = append(args, fmt.Sprintf("-p=%s", rolesPath))
```

## Constraints & Assumptions

- Assumption: ansible-navigator and ansible-galaxy accept `--flag=value` syntax for all flags currently using `--flag value`
- Assumption: Short flags (`-e`, `-i`, `-p`) also accept `-f=value` syntax
- Constraint: Must maintain backward compatibility with existing HCL configurations
- Constraint: Do not change the order of arguments or the final command semantics
- Constraint: Run `make generate`, `go build ./...`, `go test ./...`, and `make plugin-check` after changes

## Acceptance Criteria

- [ ] All `--flag value` patterns in argument construction converted to `--flag=value`
- [ ] All `-f value` short flag patterns converted to `-f=value`
- [ ] Boolean flags (no value) remain unchanged
- [ ] Positional arguments remain as separate array elements
- [ ] Plugin builds successfully (`go build ./...`)
- [ ] All tests pass (`go test ./...`)
- [ ] Plugin check passes (`make plugin-check`)
- [ ] Manual test: Build runs successfully without `argument expected` errors
