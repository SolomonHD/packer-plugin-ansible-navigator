## ADDED Requirements

### Requirement: Pure YAML Configuration

The provisioner MUST generate a complete `ansible-navigator.yml` file for all configuration settings and use it as the sole configuration source, eliminating CLI configuration flags.

#### Scenario: Generate Configuration
Given: A valid `navigator_config` block in HCL
When: The provisioner prepares to execute `ansible-navigator`
Then: It MUST generate a complete `ansible-navigator.yml` file containing all settings
And: It MUST NOT generate any CLI configuration flags (e.g., `--execution-environment-image`)
And: It MUST set the `ANSIBLE_NAVIGATOR_CONFIG` environment variable to the generated file path

#### Scenario: Cleanup Configuration
Given: A temporary `ansible-navigator.yml` file was generated
When: The provisioner finishes execution (success or failure)
Then: It MUST remove the temporary file

### Requirement: Remove CLI Flag Logic

The provisioner MUST NOT contain logic for generating CLI configuration flags, as this is replaced by pure YAML configuration.

#### Scenario: CLI Flag Generation
Given: The `navigator_config.go` file
When: The code is compiled
Then: It MUST NOT contain `buildNavigatorCLIFlags` function
And: It MUST NOT contain `hasUnmappedSettings` function
And: It MUST NOT contain `generateMinimalYAML` function
