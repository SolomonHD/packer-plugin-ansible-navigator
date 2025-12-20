# extra-vars-file-based Specification

## Purpose
TBD - created by archiving change fix-extra-vars-shell-interpretation. Update Purpose after archive.
## Requirements
### Requirement: REQ-EXTRA-VARS-FILE-001 Provisioner-generated extra-vars are passed via temporary JSON file (remote)

The SSH-based provisioner SHALL pass provisioner-generated Ansible extra vars (including Packer-derived variables and automatically added variables like `ansible_ssh_private_key_file`) using Ansible's file-based extra vars mechanism to prevent shell interpretation errors in execution environments.

#### Scenario: Extra vars written to temporary JSON file

- **GIVEN** the provisioner needs to pass provisioner-generated extra vars to ansible-navigator
- **WHEN** constructing the `ansible-navigator run` argument list
- **THEN** it SHALL write the extra vars as a JSON object to a temporary file
- **AND** the temporary file SHALL be named `packer-extravars-<random>.json` where `<random>` is a unique identifier
- **AND** the file SHALL be created in an accessible location for execution environments (same strategy as other Packer temp files)
- **AND** the file SHALL contain valid JSON with all provisioner-generated extra vars

#### Scenario: File path passed with @ prefix

- **GIVEN** a temporary extra vars JSON file has been created at `/tmp/packer-extravars-abc123.json`
- **WHEN** the provisioner constructs the `ansible-navigator run` argument list
- **THEN** it SHALL pass the file path via `--extra-vars @/tmp/packer-extravars-abc123.json`
- **AND** the `@` prefix SHALL be included to signal file-based vars to Ansible
- **AND** the argument list SHALL contain exactly one `--extra-vars` flag with the `@filepath` value

#### Scenario: Temp file cleaned up after execution

- **GIVEN** a temporary extra vars JSON file was created for play execution
- **WHEN** the play execution completes (either success or failure)
- **THEN** the temporary file SHALL be deleted
- **AND** cleanup SHALL occur via defer block to ensure deletion even on errors
- **AND** no orphaned temp files SHALL remain after provisioning

#### Scenario: File accessible from inside execution environment containers

- **GIVEN** execution environments are enabled
- **AND** ansible-navigator will run ansible-playbook inside a container
- **WHEN** the temp extra vars file is created
- **THEN** it SHALL be created in a location that is accessible from inside the container
- **AND** the file path SHALL use the same directory strategy as existing Packer temp files
- **AND** the file SHALL be readable by the container user

#### Scenario: Shell-safe argument passing prevents brace expansion

- **GIVEN** provisioner-generated extra vars contain shell metacharacters (braces, quotes, colons)
- **WHEN** the vars are written to the temp file and passed as `@filepath`
- **THEN** ansible-navigator SHALL receive the file path as a single argument
- **AND** the container shell SHALL NOT interpret braces or other metacharacters
- **AND** ansible-playbook inside the container SHALL receive the correct `--extra-vars` argument
- **AND** no "expected one argument" errors SHALL occur

#### Scenario: JSON marshaling errors handled gracefully

- **GIVEN** an unexpected error occurs when marshaling extra vars to JSON
- **WHEN** the provisioner attempts to write the temp file
- **THEN** it SHALL handle the error gracefully
- **AND** it SHALL return a clear error message indicating the JSON marshaling failure
- **AND** it SHALL NOT create an empty or invalid temp file

### Requirement: REQ-EXTRA-VARS-FILE-002 Provisioner-generated extra-vars are passed via temporary JSON file (local)

The on-target provisioner SHALL pass provisioner-generated Ansible extra vars using Ansible's file-based extra vars mechanism to prevent shell interpretation errors in execution environments.

#### Scenario: Extra vars written to temporary JSON file on target

- **GIVEN** the on-target provisioner needs to pass provisioner-generated extra vars
- **WHEN** constructing the remote shell command for `ansible-navigator run`
- **THEN** it SHALL write the extra vars as a JSON object to a temporary file in the staging directory
- **AND** the temporary file SHALL be named `packer-extravars-<random>.json`
- **AND** the file SHALL contain valid JSON with all provisioner-generated extra vars

#### Scenario: File path passed with @ prefix in remote command

- **GIVEN** a temporary extra vars JSON file was uploaded to the target staging directory
- **WHEN** the provisioner constructs the remote shell command
- **THEN** it SHALL pass the file path via `--extra-vars @<staging_dir>/packer-extravars-<random>.json`
- **AND** the `@` prefix SHALL be included to signal file-based vars
- **AND** the argument SHALL be properly quoted for the remote shell

#### Scenario: Temp file cleaned up with staging directory

- **GIVEN** a temporary extra vars JSON file was created in the staging directory
- **WHEN** provisioning completes and staging directory cleanup occurs
- **THEN** the extra vars file SHALL be removed with the staging directory
- **AND** no manual cleanup SHALL be required beyond standard staging directory cleanup

#### Scenario: File accessible to ansible-navigator on target

- **GIVEN** the temp extra vars file is created in the staging directory
- **WHEN** ansible-navigator runs on the target
- **THEN** it SHALL be able to read the file from the staging directory path
- **AND** the file SHALL be readable by the user running ansible-navigator

#### Scenario: Shell-safe argument passing in remote command

- **GIVEN** provisioner-generated extra vars contain shell metacharacters
- **WHEN** the vars are written to temp file and passed as `@filepath` in remote command
- **THEN** the remote shell SHALL NOT interpret braces or metacharacters in the file path
- **AND** ansible-navigator on the target SHALL receive the correct file path
- **AND** no "expected one argument" errors SHALL occur

### Requirement: REQ-EXTRA-VARS-FILE-003 Temporary extra vars file creation and cleanup (remote)

The SSH-based provisioner SHALL provide a helper function to create and manage temporary extra vars JSON files.

#### Scenario: Helper function creates temp file with unique name

- **GIVEN** a map of extra vars to be written
- **WHEN** the helper function is called to create a temp extra vars file
- **THEN** it SHALL generate a unique filename in the form `packer-extravars-<uuid>.json`
- **AND** it SHALL create the file in the system temporary directory
- **AND** it SHALL write the JSON-marshaled extra vars to the file
- **AND** it SHALL return the absolute file path

#### Scenario: File permissions are appropriate

- **GIVEN** a temp extra vars file is created
- **WHEN** examining file permissions
- **THEN** the file SHALL have permissions that allow reading by the current user
- **AND** the file SHALL be accessible to ansible-navigator when running locally or in containers

#### Scenario: Cleanup tracking via file path return

- **GIVEN** the helper function successfully creates a temp file
- **WHEN** it returns the file path
- **THEN** the caller SHALL be responsible for cleanup
- **AND** the caller SHALL use defer blocks to ensure cleanup occurs

### Requirement: REQ-EXTRA-VARS-FILE-004 Temporary extra vars file creation and cleanup (local)

The on-target provisioner SHALL create temporary extra vars JSON files in the staging directory on the target.

#### Scenario: Temp file created in staging directory

- **GIVEN** provisioner-generated extra vars need to be passed to ansible-navigator on target
- **WHEN** the provisioner prepares to upload files to the staging directory
- **THEN** it SHALL create a temp extra vars JSON file locally
- **AND** it SHALL upload the file to the staging directory on the target
- **AND** the filename SHALL be unique: `packer-extravars-<uuid>.json`

#### Scenario: Staging directory path used in command

- **GIVEN** the temp extra vars file is uploaded to staging directory
- **WHEN** constructing the remote ansible-navigator command
- **THEN** the file path SHALL be `<staging_dir>/packer-extravars-<uuid>.json`
- **AND** the path SHALL be prefixed with `@` in the `--extra-vars` argument

#### Scenario: Cleanup via staging directory removal

- **GIVEN** the temp extra vars file is in the staging directory
- **WHEN** the staging directory is cleaned up after provisioning
- **THEN** the extra vars file SHALL be removed automatically
- **AND** no separate cleanup SHALL be needed

