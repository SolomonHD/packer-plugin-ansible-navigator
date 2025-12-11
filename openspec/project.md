# Project Context

## Purpose

This repository contains a **Packer x5 plugin** that provides two Ansible Navigator-based provisioners:

- `ansible-navigator` (SSH-based): runs `ansible-navigator` from the machine running Packer and connects to the target via SSH.
- `ansible-navigator-local` (on-target): uploads any required artifacts to the target and runs `ansible-navigator` on the target.

The project goal is to provide a **modern, opinionated, HCL2-first** configuration surface for running Ansible content (playbooks and role FQDNs) via `ansible-navigator run`.

## Tech Stack

- Go
- `github.com/hashicorp/packer-plugin-sdk` (x5)
- `packer-sdc` for HCL2 schema generation

## Project Conventions

### Compatibility / Deprecations

- The plugin is under active development.
- **Backward compatibility is not a goal**: breaking changes and removal of legacy configuration options are acceptable.
- Documentation and OpenSpec artifacts should avoid normalizing deprecated concepts. Historical references may remain **only** in `openspec/archive/`.

### Code Style

- Keep configuration fields **explicit and minimal**.
- Prefer **object maps** over lists for configuration that represents named things.
- For repeatable, ordered actions (e.g., provisioning runs), prefer **repeatable HCL blocks**.

### Architecture Patterns

- Both provisioners invoke `ansible-navigator run`.
- Use generated HCL2 specs (`packer-sdc`) as the contract between HCL and Go config structs.

### Testing Strategy

- Unit tests (`go test ./...`).
- Plugin conformance (`make plugin-check`).
- Spec workflow validation (`openspec validate --strict`).

### Git Workflow

- Proposals live under `openspec/changes/<change-id>/`.
- Implementation work must be performed in separate tasks after proposal approval.

## Domain Context

- `ansible-navigator run` can execute either a playbook file or a generated playbook that runs a role FQDN.
- Execution environments (containers) often require `ansible.cfg` configuration (e.g., tmp dirs) to avoid permission errors.

## Important Constraints

- HCL2-first. No legacy JSON template support considerations.
- Remove and avoid reintroducing deprecated configuration patterns.

## External Dependencies

- `ansible-navigator` (v3+) and its execution environment tooling.
- `ansible-galaxy` (for roles/collections when using `requirements_file`).
