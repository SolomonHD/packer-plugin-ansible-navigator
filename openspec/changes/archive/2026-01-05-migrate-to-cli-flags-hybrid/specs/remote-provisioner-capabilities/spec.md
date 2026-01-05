# remote-provisioner-capabilities Spec Delta

## MODIFIED Requirements

### Requirement: Navigator Config File Generation

The provisioner MUST generate ansible-navigator configuration using CLI flags as the primary method. YAML configuration files MUST be generated only when the configuration contains settings without CLI flag equivalents (PlaybookArtifact, CollectionDocCache).

#### Scenario: CLI-only configuration (common path)
**Given:** NavigatorConfig contains mode, execution environment, and logging settings  
**And:** No PlaybookArtifact or CollectionDocCache settings are configured  
**When:** Executing ProvisionRemote()  
**Then:** CLI flags are generated for all settings  
**And:** No temporary YAML file is created  
**And:** The ansible-navigator command includes flags like `--mode=stdout --execution-environment-image=... --pull-policy=never`  
**And:** No `--settings` flag is present in the command

#### Scenario: Hybrid configuration with playbook artifact
**Given:** NavigatorConfig contains CLI-mappable settings (mode, execution environment)  
**And:** PlaybookArtifact is configured with `{ Enable: true, SaveAs: "/tmp/artifact.json" }`  
**When:** Executing ProvisionRemote()  
**Then:** CLI flags are generated for mode, execution environment, and logging  
**And:** A minimal YAML file is created containing ONLY playbook-artifact settings  
**And:** The ansible-navigator command includes both CLI flags and `--settings=/tmp/packer-navigator-cfg-minimal-XXX.yml`  
**And:** The minimal YAML file is cleaned up after execution

#### Scenario: Empty navigator configuration
**Given:** NavigatorConfig is nil or unspecified  
**When:** Executing ProvisionRemote()  
**Then:** No CLI flags are added for navigator configuration  
**And:** No YAML file is generated  
**And:** Default ansible-navigator behavior is used

#### Scenario: Pull policy enforcement via CLI flag
**Given:** NavigatorConfig.ExecutionEnvironment.PullPolicy is `"never"`  
**When:** Executing ProvisionRemote()  
**Then:** The command includes `--pull-policy=never`  
**And:** Docker/Podman does NOT attempt to pull the image  
**And:** Execution uses existing local images only

#### Scenario: Multiple environment variables via CLI flags
**Given:** NavigatorConfig.ExecutionEnvironment.EnvironmentVariables.Set contains multiple variables  
**When:** Executing ProvisionRemote()  
**Then:** Each variable is passed as a separate `--eev KEY=VALUE` flag  
**And:** No YAML file is needed for environment variables

#### Scenario: Volume mounts via CLI flags
**Given:** NavigatorConfig.ExecutionEnvironment.VolumeMounts contains multiple mount specifications  
**When:** Executing ProvisionRemote()  
**Then:** Each mount is passed as a separate `--evm src:dest:options` flag  
**And:** No YAML file is needed for volume mounts

#### Scenario: Temporary YAML cleanup on error
**Given:** NavigatorConfig contains unmapped settings requiring minimal YAML  
**And:** An error occurs after YAML file creation but before execution completes  
**When:** The provisioner deferred cleanup executes  
**Then:** The minimal YAML file is removed from the system  
**And:** No temporary files are leaked
