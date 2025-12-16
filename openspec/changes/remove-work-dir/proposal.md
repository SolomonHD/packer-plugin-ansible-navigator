# Change: Remove `work_dir` from both provisioners

## Why

`work_dir` is exposed in the configuration surface, but it is either inconsistently applied (remote provisioner) or effectively ignored (local provisioner runs within its staging directory). This creates user confusion and an unnecessary configuration knob.

## What Changes

- **BREAKING:** Remove `work_dir` from both provisioners' HCL configuration schema.
- Ensure any existing code paths and documentation references are removed so the field is no longer visible or usable.

## Impact

- Affected capabilities/specs:
  - `remote-provisioner-capabilities`
  - `local-provisioner-capabilities`
- Affected documentation:
  - Configuration reference and examples MUST no longer mention `work_dir`.

