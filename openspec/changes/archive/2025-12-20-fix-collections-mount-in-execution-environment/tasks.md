# Implementation Tasks

## 1. Update Environment Variable Name

- [x] 1.1 Change `ANSIBLE_COLLECTIONS_PATHS` (plural) to `ANSIBLE_COLLECTIONS_PATH` (singular) in `provisioner/ansible-navigator/galaxy.go`
- [x] 1.2 Change `ANSIBLE_COLLECTIONS_PATHS` (plural) to `ANSIBLE_COLLECTIONS_PATH` (singular) in `provisioner/ansible-navigator-local/galaxy.go`
- [x] 1.3 Update corresponding test files to use singular form
- [x] 1.4 Update test assertions in `galaxy_test.go` for both provisioners

## 2. Add EE Volume Mount Configuration

- [x] 2.1 Detect when `navigator_config.execution_environment.enabled` is true in remote provisioner
- [x] 2.2 Automatically add volume mount for collections cache directory when EE is enabled (remote)
- [x] 2.3 Detect when `navigator_config.execution_environment.enabled` is true in local provisioner
- [x] 2.4 Automatically add volume mount for collections cache directory when EE is enabled (local)
- [x] 2.5 Mount collections as read-only (`ro`) for security
- [x] 2.6 Handle custom `collections_path` configuration correctly

## 3. Set ANSIBLE_COLLECTIONS_PATH in EE Container

- [x] 3.1 Add `ANSIBLE_COLLECTIONS_PATH` to execution environment variables when EE is enabled (remote)
- [x] 3.2 Add `ANSIBLE_COLLECTIONS_PATH` to execution environment variables when EE is enabled (local)
- [x] 3.3 Point to the mounted path inside the container (e.g., `/tmp/.packer_ansible/collections`)
- [x] 3.4 Ensure this only happens when EE is enabled, not for local Ansible execution

## 4. Handle Default Collections Path

- [x] 4.1 Resolve default collections cache path (`~/.packer.d/ansible_collections_cache/ansible_collections`)
- [x] 4.2 Use the default path for volume mount if `collections_path` is not explicitly set
- [x] 4.3 Expand tilde (`~`) to actual home directory path for volume mount source
- [x] 4.4 Ensure absolute paths are used for Docker/Podman volume mounts

## 5. Testing

- [x] 5.1 Add test for volume mount when EE is enabled
- [x] 5.2 Add test for ANSIBLE_COLLECTIONS_PATH environment variable in EE
- [x] 5.3 Add test that verifies no volume mount when EE is disabled
- [x] 5.4 Add test for custom collections_path with EE
- [x] 5.5 Verify existing tests pass with singular environment variable name
- [ ] 5.6 Integration test with actual collection role execution in EE

## 6. Documentation

- [x] 6.1 Update examples to show collections working with execution environments
- [x] 6.2 Document that collections are automatically mounted when using EE
- [x] 6.3 Add troubleshooting section for collections not found in EE
- [x] 6.4 Update MIGRATION.md if needed (environment variable name change)

## 7. Validation

- [x] 7.1 Run `go build ./...` successfully
- [x] 7.2 Run `go test ./...` with all tests passing
- [x] 7.3 Run `make plugin-check` successfully
- [x] 7.4 Run `openspec validate fix-collections-mount-in-execution-environment --strict` successfully
- [ ] 7.5 Manual test: collections + EE scenario from OPENSPEC_PROMPT.md (deferred to user testing)
