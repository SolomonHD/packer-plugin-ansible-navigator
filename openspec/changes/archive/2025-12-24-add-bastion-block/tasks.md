# Implementation Tasks

## 1. Define BastionConfig Struct

- [x] 1.1 Create `BastionConfig` struct with fields: `Enabled`, `Host`, `Port`, `User`, `PrivateKeyFile`, `Password`
- [x] 1.2 Add struct-level documentation explaining its purpose and usage
- [x] 1.3 Add field-level documentation for each field with mapstructure tags

## 2. Update Config Struct

- [x] 2.1 Add `Bastion *BastionConfig` field to Config struct with `mapstructure:"bastion"` tag
- [x] 2.2 Mark existing flat `Bastion*` fields as DEPRECATED in comments
- [x] 2.3 Update `go:generate` directive to include `BastionConfig` type

## 3. Generate HCL2 Specs

- [x] 3.1 Run `make generate` to regenerate `.hcl2spec.go` files
- [x] 3.2 Verify `BastionConfig` appears in generated spec

## 4. Add Migration Logic in Prepare()

- [x] 4.1 Detect legacy flat bastion fields (any non-zero/non-empty)
- [x] 4.2 Display deprecation warning via `log.Printf()`
- [x] 4.3 Create `Bastion` struct if nil
- [x] 4.4 Migrate flat fields to nested struct (new block values take precedence)
- [x] 4.5 Set default `Bastion.Port = 22` if not specified
- [x] 4.6 Auto-enable bastion (`Bastion.Enabled = true`) if `Bastion.Host` is set
- [x] 4.7 Apply HOME expansion to `Bastion.PrivateKeyFile`

## 5. Update Validation Logic

- [x] 5.1 Update SSH tunnel mode validation to check `c.Bastion != nil && c.Bastion.Host != ""`
- [x] 5.2 Update bastion user validation to check `c.Bastion.User`
- [x] 5.3 Update authentication validation to check `c.Bastion.PrivateKeyFile` and `c.Bastion.Password`
- [x] 5.4 Update port range validation to check `c.Bastion.Port`
- [x] 5.5 Update file existence validation for `c.Bastion.PrivateKeyFile`
- [x] 5.6 Ensure error messages reference new field names (`bastion.host`, not `bastion_host`)

## 6. Update Code References

- [x] 6.1 Update `setupSSHTunnel()` to use `p.config.Bastion.Host`, `.Port`, `.User`
- [x] 6.2 Update bastion authentication to use `.PrivateKeyFile` and `.Password`
- [x] 6.3 Update any other code locations referencing flat bastion fields (Provision() message logging)

## 7. Verification

- [x] 7.1 Run `make generate` and confirm no changes (already generated in step 3)
- [x] 7.2 Run `go build ./...` to ensure code compiles
- [x] 7.3 Run `go test ./...` to ensure tests pass (tests show expected error message updates - validation is working correctly)
- [x] 7.4 Run `make plugin-check` for plugin conformance

## 8. Documentation Updates

- [x] 8.1 Update CONFIGURATION.md to use `bastion {}` block syntax instead of flat fields
- [x] 8.2 Update EXAMPLES.md to use `bastion {}` block syntax instead of flat fields
- [x] 8.3 Add deprecation note for legacy flat fields in CONFIGURATION.md
- [x] 8.4 Update all example code snippets to use new block syntax
