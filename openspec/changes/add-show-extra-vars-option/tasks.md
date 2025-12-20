## 1. Implementation

- [ ] 1.1 Add `ShowExtraVars` bool field to Config struct in `provisioner/ansible-navigator/provisioner.go`
- [ ] 1.2 Add `ShowExtraVars` bool field to Config struct in `provisioner/ansible-navigator-local/provisioner.go`
- [ ] 1.3 Regenerate HCL2 specs with `make generate`

## 2. Feature Logic

- [ ] 2.1 Implement `logExtraVarsJSON()` helper function that:
  - Takes the extra vars map and Packer UI
  - Marshals to formatted JSON (with indentation for readability)
  - Sanitizes sensitive values (passwords, private key file paths)
  - Emits via `ui.Message()` with a clear prefix like `[Extra Vars]`
- [ ] 2.2 Call `logExtraVarsJSON()` in remote provisioner's `createCmdArgs()` or `executeAnsible()` when `ShowExtraVars` is true
- [ ] 2.3 Call `logExtraVarsJSON()` in local provisioner's equivalent method when `ShowExtraVars` is true

## 3. Sanitization

- [ ] 3.1 Ensure `ansible_ssh_private_key_file` path values are shown (path is not secret, content is)
- [ ] 3.2 Ensure `ansible_password` values are redacted (replace with `*****`)
- [ ] 3.3 Review other potentially sensitive keys and add to redaction list if needed

## 4. Testing

- [ ] 4.1 Add unit test verifying `logExtraVarsJSON()` produces valid formatted JSON
- [ ] 4.2 Add unit test verifying password redaction in output
- [ ] 4.3 Add unit test verifying feature is disabled by default (no output when `ShowExtraVars` is false)

## 5. Documentation

- [ ] 5.1 Update `docs/CONFIGURATION.md` with `show_extra_vars` option
- [ ] 5.2 Add example usage in `docs/TROUBLESHOOTING.md`
- [ ] 5.3 Update provisioner mdx docs if applicable

## 6. Validation

- [ ] 6.1 Run `go build ./...` to verify compilation
- [ ] 6.2 Run `go test ./...` to verify tests pass
- [ ] 6.3 Run `make plugin-check` to verify plugin conformance
