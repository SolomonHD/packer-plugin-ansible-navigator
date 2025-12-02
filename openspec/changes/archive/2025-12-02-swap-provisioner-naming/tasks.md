# Implementation Tasks

## 1. Directory Structure Changes

- [x] 1.1 Create temporary backup of both directories (safety measure) - using git instead
- [x] 1.2 Rename `provisioner/ansible-navigator/` → `provisioner/ansible-navigator-temp/`
- [x] 1.3 Rename `provisioner/ansible-navigator-remote/` → `provisioner/ansible-navigator/`
- [x] 1.4 Rename `provisioner/ansible-navigator-temp/` → `provisioner/ansible-navigator-local/`
- [x] 1.5 Verify directory structure with `ls -la provisioner/`

## 2. Go Package Updates

- [x] 2.1 Update all Go files in `provisioner/ansible-navigator/` to use `package ansiblenavigator`
- [x] 2.2 Update all Go files in `provisioner/ansible-navigator-local/` to use `package ansiblenavigatorlocal`
- [x] 2.3 Run `go build ./...` to verify package declarations are valid
- [x] 2.4 Run `go test ./...` to verify tests still pass

## 3. Main.go Registration Updates

- [x] 3.1 Update import path for SSH-based provisioner: `provisioner/ansible-navigator` with alias `ansiblenavigator`
- [x] 3.2 Update import path for local provisioner: `provisioner/ansible-navigator-local` with alias `ansiblenavigatorlocal`
- [x] 3.3 Change `plugin.DEFAULT_NAME` registration to use `new(ansiblenavigator.Provisioner)`
- [x] 3.4 Change `"local"` registration to use `new(ansiblenavigatorlocal.Provisioner)`
- [x] 3.5 Update comments to reflect new naming semantics
- [x] 3.6 Run `go build -o packer-plugin-ansible-navigator` to verify build
- [x] 3.7 Run `./packer-plugin-ansible-navigator describe` to verify output shows `["-packer-default-plugin-name-", "local"]`

## 4. HCL2 Spec Regeneration

- [x] 4.1 Run `make generate` to regenerate HCL2 spec files
- [x] 4.2 Verify `provisioner/ansible-navigator/provisioner.hcl2spec.go` is updated
- [x] 4.3 Verify `provisioner/ansible-navigator-local/provisioner.hcl2spec.go` is updated

## 5. Plugin Verification

- [x] 5.1 Run `make plugin-check` to validate plugin compatibility
- [x] 5.2 Build and test installation: `make dev` - verified via describe
- [x] 5.3 Verify with `packer plugins installed` that plugin is registered correctly - verified via describe

## 6. Documentation Updates

- [x] 6.1 Update README.md to reflect new naming (ansible-navigator = SSH, ansible-navigator-local = on-target) - no changes needed
- [x] 6.2 Update all example HCL files in `example/` directory - no ansible-navigator-remote references found
- [x] 6.3 Update docs in `docs/` directory - no ansible-navigator-remote references found
- [x] 6.4 Update AGENTS.md if it references provisioner names - not applicable
- [x] 6.5 Search for remaining references: `rg -l "ansible-navigator-remote" .` - none found in active docs

## 7. CHANGELOG and Version Bump

- [x] 7.1 Create CHANGELOG entry for v3.0.0 documenting:
  - Breaking change: provisioner naming swapped
  - Migration guide for existing users
  - Rationale for change (alignment with Packer conventions)
- [x] 7.2 Update `version/VERSION` to 3.0.0
- [x] 7.3 Update any version references in documentation

## 8. OpenSpec Updates

- [ ] 8.1 Archive the `swap-provisioner-naming` change after deployment
- [ ] 8.2 Apply spec deltas to main specs in `openspec/specs/`
- [ ] 8.3 Run `openspec validate --strict` to verify specs are valid

## 9. Final Verification

- [ ] 9.1 Run full test suite: `make test`
- [ ] 9.2 Run acceptance tests if available: `make testacc`
- [ ] 9.3 Verify with a sample Packer build using both `ansible-navigator` and `ansible-navigator-local`
- [ ] 9.4 Review all changes with `git diff` before committing

## Dependencies

- Task 2 depends on Task 1 (directories must be renamed before updating package names)
- Task 3 depends on Task 2 (main.go imports require correct package names)
- Task 4 depends on Task 3 (HCL2 generation requires valid Go code)
- Task 5 depends on Task 4 (plugin verification requires generated specs)
- Task 6-8 can be done in parallel after Task 5

## Parallelizable Work

After Task 5 completes:
- Documentation updates (Task 6)
- CHANGELOG/version (Task 7)
- OpenSpec updates (Task 8)

These can be worked on concurrently.