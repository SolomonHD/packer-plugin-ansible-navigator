# Change: Add option to display extra vars JSON in output log

## Why

When debugging playbook execution issues, users often need to verify what extra variables are being passed to `ansible-navigator run`. Currently, the provisioner constructs a JSON object with extra vars (including Packer-injected variables like `ansible_ssh_private_key_file`, `packer_build_name`, and user-defined extra vars) and passes them via `--extra-vars`. However, users cannot easily inspect this JSON content without adding verbose debugging.

This feature adds a configuration option to log/display the extra vars JSON content during provisioner execution, making it easier to troubleshoot and verify variable passing.

## What Changes

- Add a new boolean configuration option `show_extra_vars` (default: `false`) to both provisioners
- When enabled, the provisioner will log the extra vars JSON content to the Packer UI output before executing ansible-navigator
- **Security consideration**: The output will be sanitized to redact sensitive values (passwords, private key paths) following the existing sanitization approach used for command logging

## Impact

- Affected specs: `remote-provisioner-capabilities`, `local-provisioner-capabilities`
- Affected code: `provisioner/ansible-navigator/provisioner.go`, `provisioner/ansible-navigator-local/provisioner.go`
- HCL schema changes required: new boolean field in Config struct
- HCL2 spec regeneration required
