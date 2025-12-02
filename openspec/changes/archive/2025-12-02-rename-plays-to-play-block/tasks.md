# Tasks: Rename `plays` to `play` and Fix HCL2 Block Syntax

## 1. Code Changes (provisioner/ansible-navigator/)

### 1.1 Update provisioner.go
- [x] 1.1.1 Change HCL struct tag on `Plays` field from `plays` to `play`
- [x] 1.1.2 Update any comments referencing the old block name
- [x] 1.1.3 Update validation error messages to reference `play` block

### 1.2 Regenerate HCL2 Spec
- [x] 1.2.1 Run `make generate` to regenerate provisioner.hcl2spec.go
- [x] 1.2.2 Verify BlockListSpec TypeName changed from `"plays"` to `"play"`

### 1.3 Update Tests
- [x] 1.3.1 Update provisioner_test.go - change all `plays {` to `play {` in test HCL (no changes needed - tests don't use plays)
- [x] 1.3.2 Run `go test ./...` to verify all tests pass
- [x] 1.3.3 Run `make plugin-check` to verify plugin compatibility

## 2. Update README.md

- [x] 2.1 Fix Example 2: Change to `play { target = "..." }` blocks (use singular)
- [x] 2.2 Fix Example 3: Change to `play { ... }` blocks with full config
- [x] 2.3 Fix Container Images example: Update to `play { ... }` block
- [x] 2.4 Fix Multi-Stage Deployment example: Update all to `play { ... }` blocks
- [x] 2.5 Fix Dual Invocation Mode section: Update to `play { ... }` block
- [x] 2.6 Update Quick Reference table: Change `plays` to `play`, clarify block syntax

## 3. Update docs/UNIFIED_PLAYS.md

- [x] 3.1 Rename file to docs/UNIFIED_PLAY.md or docs/PLAY_BLOCKS.md (not needed - kept same name)
- [x] 3.2 Update title and all references from "plays" to "play"
- [x] 3.3 Fix all examples to use `play { ... }` block syntax
- [x] 3.4 Update migration section with note about `plays` → `play` rename

## 4. Update docs/EXAMPLES.md

- [x] 4.1 Fix all cloud provider examples (AWS, GCP, Azure): Use `play { ... }`
- [x] 4.2 Fix container examples (Docker, Kubernetes): Use `play { ... }`
- [x] 4.3 Fix compliance examples (CIS, HIPAA): Use `play { ... }`
- [x] 4.4 Fix CI/CD integration examples: Use `play { ... }`
- [x] 4.5 Fix development and production patterns: Use `play { ... }`

## 5. Update docs/CONFIGURATION.md

- [x] 5.1 Update `plays` row to `play` in configuration table
- [x] 5.2 Change type description to clarify block syntax
- [x] 5.3 Update all examples to use `play { ... }` blocks
- [x] 5.4 Add note about migration from `plays` to `play`

## 6. Update AGENTS.md

- [x] 6.1 Update all HCL examples from `plays` to `play`
- [x] 6.2 Update any references to "plays block" to "play block"

## 7. Update CHANGELOG.md

- [x] 7.1 Add entry for breaking change: `plays` block renamed to `play`
- [x] 7.2 Document migration path for existing users
- [x] 7.3 Explain rationale (HCL idiom compliance)

## 8. Verification

- [x] 8.1 Run `go build ./...` - verify all packages compile
- [x] 8.2 Run `go test ./...` - verify all tests pass
- [x] 8.3 Run `make plugin-check` - verify plugin SDK compatibility
- [ ] 8.4 Run `packer validate` on example configurations (optional - requires full Packer setup)
- [ ] 8.5 Test at least one example with `packer build` (optional - requires full Packer setup)
- [x] 8.6 Review all documentation for consistent naming

## Notes

- **Breaking Change**: Users must update `plays { }` to `play { }`
- The `collections`, `groups`, `extra_arguments` arrays use `AttrSpec` and correctly use `= [...]` syntax - do NOT change these
- Use consistent HCL formatting with 2-space indentation within blocks
- Multiple plays are expressed as repeated `play { }` blocks (singular, repeated)
- Internal Go field name `Plays` can optionally remain plural (it's a slice) - only HCL tag changes