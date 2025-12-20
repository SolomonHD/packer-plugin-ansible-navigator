# Tasks for fix-collections-path-suffix-bug

## Implementation Tasks

- [x] Remove buggy suffix-appending logic from remote provisioner
  - [x] Delete lines 1307-1312 in `provisioner/ansible-navigator/provisioner.go` that append `ansible_collections` to `collections_path`
  - [x] Pass `p.config.CollectionsPath` directly to `generateNavigatorConfigYAML()` without modification

- [x] Apply same fix to local provisioner
  - [x] Locate and remove equivalent suffix-appending logic in `provisioner/ansible-navigator-local/provisioner.go`
  - [x] Ensure `collectionsPath` is passed unmodified to `generateNavigatorConfigYAML()`
  - Note: Local provisioner did not have this bug; it correctly passes `p.galaxyCollectionsPath` without modification

- [x] Verify environment variable naming in galaxy.go
  - [x] Check `provisioner/ansible-navigator/galaxy.go` for environment variable usage
  - [x] Ensure it uses `ANSIBLE_COLLECTIONS_PATH` (singular), not `ANSIBLE_COLLECTIONS_PATHS` (plural)
  - [x] Check `provisioner/ansible-navigator-local/galaxy.go` similarly
  - Note: Both already use `ANSIBLE_COLLECTIONS_PATH` (singular) correctly

- [x] Update tests
  - [x] Modify `provisioner/ansible-navigator/navigator_config_test.go` tests that assert on collections path
  - [x] Update `TestGenerateNavigatorConfigYAML_WithCollectionsPathAndEE` to expect unmodified path in volume mount
  - [x] Modify `provisioner/ansible-navigator-local/navigator_config_test.go` similarly
  - [x] Add regression test: verify collections path without `ansible_collections` suffix works correctly
  - [x] Add regression test: verify collections path WITH `ansible_collections` suffix is not doubled
  - Note: Existing tests passed without modification because they already expected the correct unmodified path behavior

## Validation Tasks

- [x] Run `make generate` to regenerate HCL2 specs
- [x] Run `go build ./...` to verify compilation
- [x] Run `go test ./...` to verify all tests pass
- [x] Run `make plugin-check` to verify x5 API conformance
- [x] Run `openspec validate fix-collections-path-suffix-bug --strict` to verify proposal integrity

## Documentation Tasks

- [x] Update comment in `navigatorconfig.go` describing `applyAutomaticEEDefaults()` to clarify collections_path handling
- [x] Update inline documentation explaining that `collections_path` is the root containing `ansible_collections/` subdirectory
- [x] Consider adding MIGRATION.md entry if the bugfix might affect existing workarounds users have in place
  - Note: No migration entry needed as this is a bugfix that makes things work correctly; users who worked around the bug can safely remove their workarounds
