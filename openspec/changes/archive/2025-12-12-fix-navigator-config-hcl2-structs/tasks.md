## 1. Update EnvironmentVariablesConfig Struct

- [x] 1.1 Replace `Variables map[string]string` with `Pass []string` and `Set map[string]string` in remote provisioner
- [x] 1.2 Replace `Variables map[string]string` with `Pass []string` and `Set map[string]string` in local provisioner
- [x] 1.3 Update mapstructure tags: `mapstructure:"pass"` and `mapstructure:"set"` (remove `,remain`)
- [x] 1.4 Add docstrings explaining the pass/set structure

## 2. Fix AnsibleConfig Struct

- [x] 2.1 Remove `AnsibleConfigInner` struct from remote provisioner
- [x] 2.2 Remove `AnsibleConfigInner` struct from local provisioner
- [x] 2.3 Move `Defaults` and `SSHConnection` fields directly into `AnsibleConfig` struct
- [x] 2.4 Remove `mapstructure:",squash"` tag from both provisioners
- [x] 2.5 Update mapstructure tags: `mapstructure:"defaults"` and `mapstructure:"ssh_connection"`

## 3. Update go:generate Directive

- [x] 3.1 Update directive in remote provisioner to remove `AnsibleConfigInner`
- [x] 3.2 Update directive in local provisioner to remove `AnsibleConfigInner`
- [x] 3.3 Verify all struct types are listed in the directive

## 4. Regenerate HCL2 Specs

- [x] 4.1 Run `make generate` to regenerate all `.hcl2spec.go` files
- [x] 4.2 Verify `FlatEnvironmentVariablesConfig` has `pass` and `set` fields in HCL2Spec
- [x] 4.3 Verify `FlatAnsibleConfig` has `defaults` and `ssh_connection` as BlockSpecs
- [x] 4.4 Verify no `cty.DynamicPseudoType` or `mapstructure:",remain"` artifacts in generated code

## 5. Update YAML Generation Logic

- [x] 5.1 Update `generateNavigatorConfigYAML()` to handle new `EnvironmentVariablesConfig` structure
- [x] 5.2 Ensure YAML output uses hyphenated keys (`environment-variables`, `execution-environment`)
- [x] 5.3 Generate correct `pass` and `set` sections under `environment-variables`
- [x] 5.4 Update handling of `AnsibleConfig` to use direct fields instead of `Inner`

## 6. Update Example Files

- [x] 6.1 Update `example/nested-navigator-config.pkr.hcl` with new `environment_variables` syntax
- [x] 6.2 Update any other example files using `environment_variables` block
- [x] 6.3 Update example files using `ansible_config` to show `defaults` and `ssh_connection` blocks

## 7. Update MIGRATION.md

- [x] 7.1 Document the `environment_variables` syntax change (from inline vars to `pass`/`set`)
- [x] 7.2 Document the `ansible_config` syntax change (explicit `defaults`/`ssh_connection` blocks)
- [x] 7.3 Add migration examples showing old vs new syntax

## 8. Validation

- [x] 8.1 Run `make generate` after struct changes (generates HCL2 specs)
- [x] 8.2 Run `go build ./...` to verify compilation
- [x] 8.3 Run `go test ./...` to verify existing tests pass
- [x] 8.4 Run `make plugin-check` to verify plugin conformance
- [x] 8.5 Test `packer validate` on updated example files
- [x] 8.6 Verify YAML generation produces correct ansible-navigator.yml content

## Notes

- Both provisioner files (`ansible-navigator/provisioner.go` and `ansible-navigator-local/provisioner.go`) require identical struct changes
- The YAML generation logic must convert underscore field names (HCL) to hyphenated keys (ansible-navigator.yml)
- This is a **breaking change** - users will need to update their HCL configurations
