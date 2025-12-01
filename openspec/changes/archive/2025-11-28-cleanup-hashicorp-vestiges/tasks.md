# Tasks: Clean Up HashiCorp Vestiges

## 1. Update Repository Ownership Files
- [x] 1.1 Update CODEOWNERS to replace `@hashicorp/packer` with `@SolomonHD`

## 2. Update License Headers
- [x] 2.1 Update .golangci.yml to remove HashiCorp MPL-2.0 copyright comment and add Apache 2.0 header with SolomonHD copyright

## 3. Clean Up Changelog
- [x] 3.1 Remove or clearly mark v1.1.4 (July 2025) HashiCorp PR references as legacy history from original project
- [x] 3.2 Preserve v1.0.0 evolution note as intentional historical attribution

## 4. Update Go Module Metadata
- [x] 4.1 Update go.mod retract comment to reference this repository (`github.com/SolomonHD/packer-plugin-ansible-navigator`) instead of original `packer-plugin-ansible`

## 5. Fix Build Tooling
- [x] 5.1 Update GNUmakefile web-docs script to use "SolomonHD" organization instead of "hashicorp"
- [x] 5.2 Delete duplicate MAKEFILE (keep only GNUmakefile)

## 6. Validation
- [x] 6.1 Run `go build ./...` to verify build succeeds after changes
- [x] 6.2 Run `go test ./...` to verify tests pass (if applicable)