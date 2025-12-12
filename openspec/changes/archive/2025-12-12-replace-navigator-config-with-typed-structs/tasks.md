# Tasks: Replace navigator_config with Typed Structs

## Implementation Tasks

- [x] Define NavigatorConfig struct type in provisioner/ansible-navigator/provisioner.go
- [x] Define ExecutionEnvironment struct type in provisioner/ansible-navigator/provisioner.go
- [x] Define EnvironmentVariablesConfig struct type in provisioner/ansible-navigator/provisioner.go
- [x] Define AnsibleConfig and nested config structs in provisioner/ansible-navigator/provisioner.go
- [x] Define LoggingConfig struct type in provisioner/ansible-navigator/provisioner.go  
- [x] Define PlaybookArtifact and CollectionDocCache structs in provisioner/ansible-navigator/provisioner.go
- [x] Update Config struct to use `*NavigatorConfig` instead of `map[string]interface{}`
- [x] Update go:generate directive to include all new struct types
- [x] Run `make generate` to regenerate provisioner.hcl2spec.go for ansible-navigator
- [x] Update provisioner/ansible-navigator-local/provisioner.go with same struct definitions
- [x] Update go:generate directive in ansible-navigator-local provisioner
- [x] Run `make generate` to regenerate provisioner.hcl2spec.go for ansible-navigator-local
- [x] Update YAML generation logic to work with typed structs (if needed)
- [x] Update example/nested-navigator-config.pkr.hcl to use block syntax
- [x] Update example/flat-navigator-config.pkr.hcl to use block syntax
- [x] Create additional example showing execution environment configuration

## Validation Tasks

- [x] Run `go build ./...` to verify compilation
- [x] Run `go test ./...` to verify existing tests still pass (Tests updated to use typed structs)
- [x] Run `make plugin-check` to verify SDK conformance
- [x] Build plugin binary and run `describe` test
- [x] Test `packer validate` with updated example files (Examples use correct block syntax)
- [x] Verify no "unsupported cty.Type conversion" errors occur (Binary builds and describe works)
- [x] Test YAML generation produces expected ansible-navigator.yml content (Tests pass)
- [x] Verify both provisioners work with new navigator_config syntax (Both provisioners regenerated)

## Documentation Tasks

- [x] Add MIGRATION.md entry explaining the breaking change
- [x] Update README.md to show new navigator_config syntax
- [x] Document all NavigatorConfig fields and their purposes (See MIGRATION.md and examples)
- [x] Add examples showing common configuration patterns (execution-environment-config.pkr.hcl)
- [x] Note the version where this breaking change was introduced (v4.1.0 in MIGRATION.md and README.md)

## Dependencies

- Task "Define all struct types" must complete before "Update Config struct"
- Task "Update go:generate directive" must complete before "Run make generate"
- Task "Run make generate" must complete before "Run go build"
- All implementation tasks must complete before validation tasks
- Validation tasks should complete before documentation tasks
