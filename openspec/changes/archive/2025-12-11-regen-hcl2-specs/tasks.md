## 1. Regenerate HCL2 Spec Files

- [x] 1.1 Run `make generate` to regenerate all `.hcl2spec.go` files
- [x] 1.2 Verify `AnsibleNavigatorPath` appears in `provisioner/ansible-navigator/provisioner.hcl2spec.go`

## 2. Verify Build

- [x] 2.1 Run `go build ./...` to ensure compilation succeeds
- [x] 2.2 Run `go test ./...` to ensure tests pass
- [x] 2.3 Run `make plugin-check` to verify Packer SDK compatibility

## 3. Verification

- [x] 3.1 Confirm FlatConfig struct includes all Config fields
- [x] 3.2 Confirm HCL2Spec() returns spec entries for all fields
- [x] 3.3 Test plugin binary with a Packer template using `ansible_navigator_path`
