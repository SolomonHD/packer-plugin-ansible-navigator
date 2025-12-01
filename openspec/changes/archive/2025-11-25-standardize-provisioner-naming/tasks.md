# Implementation Tasks

## 1. Preparation
- [x] 1.1 Document current state and all affected files
- [x] 1.2 Create comprehensive proposal
- [x] 1.3 Get approval for the naming change

## 2. Directory Restructuring
- [x] 2.1 Rename `provisioner/ansible-navigator/` to `provisioner/ansible-navigator-remote/`
- [x] 2.2 Rename `provisioner/ansible-navigator-local/` to `provisioner/ansible-navigator/`
- [x] 2.3 Verify all files moved correctly (no files left behind)

## 3. Update Import Paths
- [x] 3.1 Update `main.go` import statements
  - Change `provisioner/ansible-navigator` → `provisioner/ansible-navigator-remote`
  - Change `provisioner/ansible-navigator-local` → `provisioner/ansible-navigator`
- [x] 3.2 Update cross-package imports if any exist
- [x] 3.3 Update `go.mod` if package paths changed

## 4. Update Package Declarations
- [x] 4.1 Update package name in all files under `provisioner/ansible-navigator-remote/`
  - Verify package declaration is `package ansible` (for remote)
- [x] 4.2 Update package name in all files under `provisioner/ansible-navigator/`
  - Kept as `package ansiblelocal` for differentiation

## 5. Update Plugin Registration (main.go)
- [x] 5.1 Update variable names for clarity
  - Updated imports to use `ansiblelocal` and `ansible`
- [x] 5.2 Verify registration calls map correctly:
  - `RegisterProvisioner("ansible-navigator", ...)` uses local provisioner
  - `RegisterProvisioner("ansible-navigator-remote", ...)` uses remote provisioner

## 6. Update Test Files
- [x] 6.1 Update import paths in all test files under `provisioner/ansible-navigator/`
- [x] 6.2 Update import paths in all test files under `provisioner/ansible-navigator-remote/`
- [x] 6.3 Run all tests to verify no breakage

## 7. Update Documentation
- [x] 7.1 Search and update all documentation references
  - `docs/provisioners/ansible-navigator-local.mdx` → content updates
  - `docs/provisioners/ansible-navigator.mdx` → verify refers to correct variant
- [x] 7.2 Update README.md if it references directory structure
- [x] 7.3 Update AGENTS.md if it references provisioner structure
- [x] 7.4 Update any example files in `example/` directory

## 8. Additional File Updates
- [x] 8.1 Update `.gitignore` if it has directory-specific entries (N/A)
- [x] 8.2 Update any build scripts or Makefiles (N/A)
- [x] 8.3 Update CI/CD configuration if it references specific directories (N/A)

## 9. Validation
- [x] 9.1 Run `go build ./...` to ensure all packages compile
- [x] 9.2 Run `go test ./...` to ensure all tests pass
- [x] 9.3 Run `make test` if Makefile has test targets (used go test)
- [ ] 9.4 Verify plugin loads correctly with `packer init` (deferred to user)
- [ ] 9.5 Test actual provisioning with both variants (deferred to user)

## 10. Final Checks
- [x] 10.1 Search codebase for any remaining "ansible-navigator-local" references
- [x] 10.2 Verify no hardcoded paths in tests
- [ ] 10.3 Review CHANGELOG.md - add entry if public-facing (for maintainer)
- [ ] 10.4 Update version/documentation as needed (for maintainer)