# Change: Fix execution environment CLI flag generation

## Why

The plugin generates incorrect CLI arguments for ansible-navigator's execution environment feature. When a user configures `execution_environment = "some-image:tag"`, the plugin generates `--execution-environment some-image:tag`. However, modern ansible-navigator (v3+) expects `--ee true --eei some-image:tag`, where `--ee` is a boolean flag and `--eei` specifies the container image.

This causes ansible-navigator to fail with:

```
Error: The setting 'execution-environment' must be one of 'true' or 'false'
```

## What Changes

- **Remote provisioner** (`provisioner/ansible-navigator/provisioner.go`): Change CLI flag generation from `--execution-environment <image>` to `--ee true --eei <image>`
- **Local provisioner** (`provisioner/ansible-navigator-local/provisioner.go`): Apply the same fix
- **Specs**: Update both `local-provisioner-capabilities` and `remote-provisioner-capabilities` specs to reflect correct CLI behavior
- **Documentation**: Update `docs/CONFIGURATION.md` to describe the actual CLI flags used

## Impact

- Affected specs: `local-provisioner-capabilities`, `remote-provisioner-capabilities`
- Affected code:
  - `provisioner/ansible-navigator/provisioner.go` (executePlays and executeSinglePlaybook methods)
  - `provisioner/ansible-navigator-local/provisioner.go` (command construction)
  - `docs/CONFIGURATION.md`
- No breaking changes to HCL configuration (users still use `execution_environment = "image"`)
- Minimum supported ansible-navigator version: v3.0+
