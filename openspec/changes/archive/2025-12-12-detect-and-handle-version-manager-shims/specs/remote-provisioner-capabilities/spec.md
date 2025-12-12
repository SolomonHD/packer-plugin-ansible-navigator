# Spec Delta: Remote Provisioner Capabilities

## ADDED Requirements

### Requirement: Version Manager Shim Detection

The SSH-based provisioner SHALL detect when `ansible-navigator` resolves to a version manager shim (asdf, rbenv, pyenv) and attempt automatic resolution to prevent silent hangs caused by shim recursion loops.

#### Scenario: asdf shim detected and resolved successfully

- **GIVEN** ansible-navigator is installed via asdf
- **AND** the command resolves to `~/.asdf/shims/ansible-navigator`
- **WHEN** the provisioner performs version check in `getVersion()`
- **THEN** it SHALL detect the file is an asdf shim by reading the file header
- **AND** it SHALL execute `asdf which ansible-navigator` to resolve the real binary path
- **AND** it SHALL use the resolved path for the version check
- **AND** the version check SHALL succeed without hanging

#### Scenario: rbenv shim detected and resolved

- **GIVEN** ansible-navigator is installed via rbenv
- **AND** the command resolves to a rbenv shim file
- **WHEN** the provisioner performs version check
- **THEN** it SHALL detect the rbenv shim pattern in the file header
- **AND** it SHALL execute `rbenv which ansible-navigator` to resolve the real binary
- **AND** it SHALL use the resolved binary path for version check
- **AND** the version check SHALL complete successfully

#### Scenario: pyenv shim detected and resolved

- **GIVEN** ansible-navigator is installed via pyenv
- **AND** the command resolves to a pyenv shim file
- **WHEN** the provisioner performs version check
- **THEN** it SHALL detect the pyenv shim pattern in the file header
- **AND** it SHALL execute `pyenv which ansible-navigator` to resolve the real binary
- **AND** it SHALL use the resolved binary path for version check
- **AND** the version check SHALL complete successfully

#### Scenario: Shim detected but resolution fails

- **GIVEN** ansible-navigator command resolves to an asdf shim
- **AND** `asdf which ansible-navigator` returns an error or empty result
- **WHEN** the provisioner attempts version check
- **THEN** it SHALL fail immediately with error message:

  ```
  Error: ansible-navigator appears to be an asdf shim, but the real binary could not be resolved.
  
  SOLUTIONS:
  1. Verify ansible-navigator is installed:
     $ asdf list ansible
  
  2. Find the actual binary path:
     $ asdf which ansible-navigator
  
  3. Configure the plugin to use the actual binary:
     provisioner "ansible-navigator" {
       command = "/path/to/actual/ansible-navigator"
       # OR
       ansible_navigator_path = ["/path/to/bin/directory"]
     }
  
  4. Alternatively, skip the version check (not recommended):
     provisioner "ansible-navigator" {
       skip_version_check = true
     }
  ```

- **AND** the error message SHALL include the detected version manager name (asdf/rbenv/pyenv)

#### Scenario: Non-shim file detected correctly

- **GIVEN** ansible-navigator is a regular binary (not a shim)
- **WHEN** the provisioner performs shim detection
- **THEN** it SHALL correctly identify the file as NOT a shim
- **AND** it SHALL proceed with normal version check using the binary directly
- **AND** no resolution attempt SHALL be made

#### Scenario: Shim detection is fast

- **GIVEN** any ansible-navigator executable
- **WHEN** the provisioner performs shim detection
- **THEN** the detection overhead SHALL be less than 100 milliseconds
- **AND** it SHALL not significantly impact plugin initialization time

### Requirement: Enhanced Timeout Error Messages

The SSH-based provisioner SHALL provide enhanced error messages when version check times out, including guidance for troubleshooting version manager shim issues.

#### Scenario: Timeout error includes shim troubleshooting

- **WHEN** ansible-navigator version check times out after the configured duration
- **THEN** it SHALL return an error message that includes:

  ```
  ansible-navigator version check timed out after 60s.
  
  COMMON CAUSES:
  1. Version manager shim (asdf/rbenv/pyenv) causing infinite recursion
  2. ansible-navigator not properly installed or not in PATH
  3. ansible-navigator requires additional configuration
  
  TROUBLESHOOTING:
  1. Check if you're using a version manager:
     $ which ansible-navigator
     $ head -1 $(which ansible-navigator)  # Should show shebang
  
  2. If using asdf, find the real binary:
     $ asdf which ansible-navigator
  
  3. Configure the plugin with the actual binary:
     provisioner "ansible-navigator" {
       command = "/home/user/.asdf/installs/ansible/2.9.0/bin/ansible-navigator"
       # OR
       ansible_navigator_path = ["/home/user/.asdf/installs/ansible/2.9.0/bin"]
     }
  
  4. Verify ansible-navigator works independently:
     $ ansible-navigator --version
  
  5. Skip version check if needed (not recommended):
     provisioner "ansible-navigator" {
       skip_version_check = true
     }
  ```

- **AND** it SHALL mention version manager shims as the first common cause

## MODIFIED Requirements

### Requirement: Version Check Timeout Configuration

The SSH-based provisioner SHALL support a configurable timeout for the version check operation to prevent indefinite hangs when ansible-navigator cannot be located or is slow to respond, with enhanced error messages for version manager shim troubleshooting.

#### Scenario: Default timeout with shim detection

- **GIVEN** a configuration for `provisioner "ansible-navigator"` without `version_check_timeout` specified
- **AND** `skip_version_check` is not set or is `false`
- **WHEN** the provisioner performs version check
- **THEN** it SHALL first attempt shim detection and resolution
- **AND** if a shim is detected, it SHALL use the resolved path
- **AND** it SHALL apply a default timeout of 60 seconds to the version check
- **AND** if the timeout expires, the error message SHALL include shim troubleshooting guidance

#### Scenario: Shim resolution bypasses timeout issues

- **GIVEN** ansible-navigator is installed via asdf and would normally cause timeout
- **WHEN** the provisioner detects the asdf shim and successfully resolves to the real binary
- **THEN** the version check SHALL use the real binary path
- **AND** the version check SHALL complete within normal timeframes (< 5 seconds)
- **AND** no timeout error SHALL occur

#### Scenario: Manual command configuration bypasses shim detection

- **GIVEN** a configuration with `command = "/absolute/path/to/ansible-navigator"`
- **WHEN** the provisioner performs version check
- **THEN** shim detection MAY be skipped if the path is absolute and points to a real binary
- **OR** shim detection MAY still run but resolution SHALL use the configured absolute path
- **AND** existing manual configurations SHALL continue to work without changes
