# OpenSpec change prompt

## Context
This repository was forked from HashiCorp's `packer-plugin-ansible`. While the project has been rebranded to `packer-plugin-ansible-navigator` under SolomonHD's ownership with Apache 2.0 licensing, several files still contain vestiges from the original project (HashiCorp copyright notices, team references, old PR links) that should be cleaned up.

## Goal
Remove or update all vestiges from the original HashiCorp packer-plugin-ansible repository that don't serve as proper historical attribution. Only keep references that clearly acknowledge the fork origin for transparency.

## Scope
- In scope:
  - Update CODEOWNERS to reference the new owner
  - Update .golangci.yml copyright and license headers
  - Clean up old HashiCorp PR references in CHANGELOG.md
  - Update go.mod retract comment
  - Fix GNUmakefile organization parameter
  - Remove duplicate MAKEFILE (keep only GNUmakefile)
- Out of scope:
  - AGENTS.md fork acknowledgment (intentional, keep)
  - CHANGELOG.md v1.0.0 evolution note (intentional, keep)
  - Any other intentional historical attribution

## Desired behavior
- CODEOWNERS should reference `@SolomonHD` or appropriate owner
- .golangci.yml should have Apache 2.0 license header with SolomonHD copyright
- CHANGELOG.md old entries (v1.1.4 July 2025 and earlier from original repo) should be removed or clearly separated as "Legacy History from packer-plugin-ansible"
- go.mod retract comment should reference this repo, not the original
- GNUmakefile web-docs script should use correct organization name
- Only one makefile should exist (GNUmakefile)

## Constraints & assumptions
- Assume Apache 2.0 is the correct license (already in LICENSE file)
- Assume SolomonHD is the sole owner/maintainer
- Keep the project buildable after changes (don't break makefile targets)

## Acceptance criteria
- [ ] CODEOWNERS references SolomonHD, not @hashicorp/packer
- [ ] .golangci.yml has Apache 2.0 header with SolomonHD copyright
- [ ] CHANGELOG.md old HashiCorp PR links are removed or clearly marked as legacy
- [ ] go.mod retract comment updated to reference new repo
- [ ] GNUmakefile uses "SolomonHD" or appropriate org in web-docs script
- [ ] MAKEFILE file is deleted (only GNUmakefile remains)
- [ ] `go build ./...` succeeds after changes