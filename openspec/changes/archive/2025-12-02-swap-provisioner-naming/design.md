# Design: Swap Provisioner Naming

## Context

This change affects multiple interconnected systems:
- Go source files and package declarations
- Plugin registration in main.go
- Directory structure
- Documentation and user-facing references
- Build/release tooling

The change must be coordinated carefully to avoid broken imports, incorrect references, and user confusion during migration.

## Goals / Non-Goals

### Goals
- Align provisioner naming with Packer ecosystem conventions (SSH-based as default)
- Minimize confusion for users migrating from official `ansible` plugin
- Maintain all existing functionality during the rename
- Provide clear migration path for existing v2.x users

### Non-Goals
- Adding new features (this is purely a naming/structure change)
- Changing core provisioning logic
- Modifying configuration options (beyond name references in docs)

## Decisions

### Decision 1: Directory Naming Strategy

**Choice:** Swap directory names to match HCL provisioner names

**Rationale:**
- `provisioner/ansible-navigator/` should contain the default provisioner (remote/SSH)
- `provisioner/ansible-navigator-local/` should contain the local execution provisioner
- This follows the pattern established by the official ansible plugin

**Alternatives Considered:**
1. Keep directories as-is, only change registration names → Rejected: Creates confusion between directory names and HCL names
2. Use generic names like `provisioner/remote/` and `provisioner/local/` → Rejected: Less descriptive, harder to identify in codebase

### Decision 2: Package Naming Convention

**Choice:** Use distinct package names that reflect execution mode

- `provisioner/ansible-navigator/` → `package ansiblenavigator`
- `provisioner/ansible-navigator-local/` → `package ansiblenavigatorlocal`

**Rationale:**
- Go package names cannot contain hyphens
- Package names should indicate the execution mode clearly
- Primary provisioner uses shorter name (no suffix)

### Decision 3: Registration Name Pattern

**Choice:** Follow Packer SDK conventions

- `plugin.DEFAULT_NAME` → primary provisioner (SSH-based, HCL: `ansible-navigator`)
- `"local"` → secondary provisioner (HCL: `ansible-navigator-local`)

**Rationale:**
- Matches official ansible plugin pattern (`ansible` and `ansible-local`)
- `plugin.DEFAULT_NAME` ensures the primary provisioner doesn't get double-prefixed

### Decision 4: Rename Execution Strategy

**Choice:** Use temporary directory for safe swap

Since we're swapping two directories, we need to avoid conflicts:
1. Rename `provisioner/ansible-navigator/` → `provisioner/ansible-navigator-temp/`
2. Rename `provisioner/ansible-navigator-remote/` → `provisioner/ansible-navigator/`
3. Rename `provisioner/ansible-navigator-temp/` → `provisioner/ansible-navigator-local/`
4. Update all Go files with new package declarations
5. Update main.go imports and registrations
6. Run `go build ./...` to verify

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Breaking existing user configurations | CHANGELOG with clear migration guide, major version bump |
| Broken imports during rename | Atomic directory swaps, immediate `go build` verification |
| Documentation drift | Update all docs in same PR, search for old names |
| Test failures | Run full test suite after changes |

## Migration Plan

### For Users

1. Update Packer configuration:
   - `provisioner "ansible-navigator"` → `provisioner "ansible-navigator-local"` (if using local execution)
   - `provisioner "ansible-navigator-remote"` → `provisioner "ansible-navigator"` (if using SSH execution)

2. Update `required_plugins` block:
   ```hcl
   ansible-navigator = {
     version = ">= 3.0.0"
     source  = "github.com/solomonhd/ansible-navigator"
   }
   ```

3. Run `packer init` to update plugin

### Rollback

If issues discovered post-release:
1. Revert all changes in a hotfix release (v3.0.1)
2. Document issue and provide workaround
3. Re-attempt in subsequent release with fixes

## Open Questions

- None at this time. The scope and approach are well-defined.