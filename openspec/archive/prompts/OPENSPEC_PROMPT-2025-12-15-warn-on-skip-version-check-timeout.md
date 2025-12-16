# OpenSpec Change Prompt

## Context
The plugin supports `skip_version_check` and `version_check_timeout`. When version checks are skipped, the timeout value is ignored, but the user receives no feedback.
This is a configuration footgun.

## Goal
Add a user-visible warning when `skip_version_check=true` and `version_check_timeout` is also set.

## Scope

**In scope:**
- Emit a warning during configuration validation/prepare for both provisioners.
- Ensure the warning is surfaced in Packer UI output (not only logs).

**Out of scope:**
- Changing default timeout behavior.
- Removing either option.

## Desired Behavior
If the user configures:

```hcl
provisioner "ansible-navigator" {
  skip_version_check    = true
  version_check_timeout = "60s"
  play { target = "site.yml" }
}
```

The plugin prints a warning similar to:

> `Warning: version_check_timeout is ignored when skip_version_check=true`

## Constraints & Assumptions
- Warning should not fail the build.

## Acceptance Criteria
- [ ] Warning appears for both provisioners when both settings are set.
- [ ] No warning when `skip_version_check=false`.

## Expected areas/files touched
- `provisioner/ansible-navigator/provisioner.go`
- `provisioner/ansible-navigator-local/provisioner.go`

