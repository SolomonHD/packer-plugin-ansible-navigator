# OpenSpec Change Prompt

## Context

The packer-plugin-ansible-navigator generates incorrect CLI arguments for ansible-navigator's execution environment feature. When a user configures `execution_environment = "some-image:tag"`, the plugin generates:

```bash
--execution-environment some-image:tag
```

However, **modern ansible-navigator (v3+)** expects:

```bash
--ee true --eei some-image:tag
```

Where:

- `--ee` (or `--execution-environment`) is a boolean flag (`true`/`false`)
- `--eei` (or `--execution-environment-image`) specifies the container image

This causes ansible-navigator to fail with:

```
Error: The setting 'execution-environment' must be one of 'true' or 'false'
```

When run non-interactively (as in Packer), this error may cause a hang instead of a clear failure.

## Goal

Fix the CLI argument generation to use the correct modern ansible-navigator syntax so execution environments work properly.

## Scope

**In scope:**

- Fix `provisioner.go` to generate `--ee true --eei <image>` instead of `--execution-environment <image>`
- Update both `executePlays()` and `executeSinglePlaybook()` methods
- Update documentation in `docs/CONFIGURATION.md` to reflect the correct CLI behavior
- Update any inline comments or struct field docs that mention the old syntax

**Out of scope:**

- Adding new configuration options
- Changing the HCL config syntax (users still use `execution_environment = "image"`)
- Supporting legacy ansible-navigator versions with the old syntax
- Changes to the local provisioner (if not affected)

## Desired Behavior

After the fix:

1. When `execution_environment` is set, the plugin generates:

   ```
   ansible-navigator run --ee true --eei <image> --mode stdout ...
   ```

2. When `execution_environment` is empty/unset, no EE flags are added (ansible-navigator uses its default)

3. Documentation accurately describes the underlying CLI behavior

## Constraints & Assumptions

- Assume ansible-navigator v3.0+ is the minimum supported version
- The HCL config attribute name (`execution_environment`) stays the same for backward compatibility
- The `provisioner/ansible-navigator-local/provisioner.go` may also need the same fix if it has similar code

## Acceptance Criteria

- [ ] `provisioner.go` generates `--ee true --eei <image>` when `execution_environment` is set
- [ ] Both `executePlays()` and `executeSinglePlaybook()` are updated
- [ ] `go build ./...` passes
- [ ] `go test ./...` passes
- [ ] Documentation in `docs/CONFIGURATION.md` is updated to explain the actual CLI flags used
- [ ] Any unit tests covering execution environment args are updated or added
