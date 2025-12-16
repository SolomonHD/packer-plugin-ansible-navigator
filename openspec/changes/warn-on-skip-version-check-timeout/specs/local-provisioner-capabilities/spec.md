## ADDED Requirements

### Requirement: Provide `skip_version_check` configuration field (parity)

The on-target provisioner SHALL include a `skip_version_check` configuration field for parity with the remote provisioner, even though local version checks are not currently performed.

#### Scenario: Configuration field present
Given: A configuration for `provisioner "ansible-navigator-local"` including `skip_version_check = true`
When: Packer parses the configuration
Then: Parsing succeeds and the field is accepted (non-fatal)

### Requirement: Warn when `version_check_timeout` is ineffective due to `skip_version_check`

When users explicitly configure `version_check_timeout` but also set `skip_version_check = true`, the plugin SHALL emit a user-visible warning indicating that the timeout is ignored.

#### Scenario: Warning when skip_version_check=true and version_check_timeout explicitly set
Given: A configuration for `provisioner "ansible-navigator-local"` with `skip_version_check = true` and an explicitly set `version_check_timeout`
When: The provisioner prepares for execution (configuration validation/prepare)
Then: The provisioner prints a non-fatal warning in Packer UI output stating that `version_check_timeout` is ignored when `skip_version_check=true`

#### Scenario: No warning when skip_version_check=false
Given: A configuration for `provisioner "ansible-navigator-local"` with `skip_version_check = false` and an explicitly set `version_check_timeout`
When: The provisioner prepares for execution (configuration validation/prepare)
Then: No warning about `version_check_timeout` being ignored is printed

#### Scenario: No warning when version_check_timeout not explicitly set
Given: A configuration for `provisioner "ansible-navigator-local"` with `skip_version_check = true` and without an explicitly set `version_check_timeout`
When: The provisioner prepares for execution (configuration validation/prepare)
Then: No warning about `version_check_timeout` being ignored is printed
