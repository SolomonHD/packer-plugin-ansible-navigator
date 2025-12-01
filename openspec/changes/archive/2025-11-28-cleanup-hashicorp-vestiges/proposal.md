# Change: Clean Up HashiCorp Vestiges

## Why
This repository was forked from HashiCorp's `packer-plugin-ansible` and has been rebranded to `packer-plugin-ansible-navigator` under SolomonHD's ownership with Apache 2.0 licensing. However, several files still contain vestiges from the original project (HashiCorp copyright notices, team references, old PR links) that create confusion about the project's current ownership and maintenance.

## What Changes
- **CODEOWNERS**: Update from `@hashicorp/packer` to `@SolomonHD`
- **.golangci.yml**: Replace HashiCorp MPL-2.0 copyright header with Apache 2.0 header and SolomonHD copyright
- **CHANGELOG.md**: Remove or clearly mark old HashiCorp PR references (v1.1.4 July 2025 and earlier) as legacy history
- **go.mod**: Update retract comment to reference this repository instead of the original
- **GNUmakefile**: Update web-docs script to use "SolomonHD" organization
- **MAKEFILE**: Delete this duplicate file (only GNUmakefile should exist)

## Impact
- Affected specs: New `project-metadata` capability
- Affected code: Repository configuration and metadata files only (no functional code changes)
- Breaking changes: None
- Build verification: `go build ./...` must succeed after changes

## Notes
Intentionally kept for historical attribution:
- AGENTS.md fork acknowledgment
- CHANGELOG.md v1.0.0 evolution note (describes the transition from packer-plugin-ansible)