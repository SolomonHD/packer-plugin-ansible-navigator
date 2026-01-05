# local-provisioner-capabilities Spec Delta

## MODIFIED Requirements

### Requirement: Navigator Config File Generation

The provisioner MUST generate ansible-navigator configuration using CLI flags as the primary method. YAML configuration files MUST be uploaded to the target only when the configuration contains settings without CLI flag equivalents (PlaybookArtifact, CollectionDocCache).

#### Scenario: CLI-only configuration (common path)
**Given:** NavigatorConfig contains mode, execution environment, and logging settings  
**And:** No PlaybookArtifact or CollectionDocCache settings are configured  
**When:** Executing Provision()  
**Then:** CLI flags are generated for all settings  
**And:** No temporary YAML file is created or uploaded  
**And:** The remote ansible-navigator command includes flags like `--mode=stdout --execution-environment-image=... --pull-policy=never`  
**And:** No `--settings` flag is present in the remote command

#### Scenario: Hybrid configuration with playbook artifact
**Given:** NavigatorConfig contains CLI-mappable settings (mode, execution environment)  
**And:** PlaybookArtifact is configured with `{ Enable: true, SaveAs: "/tmp/artifact.json" }`  
**When:** Executing Provision()  
**Then:** CLI flags are generated for mode, execution environment, and logging  
**And:** A minimal YAML file is created locally containing ONLY playbook-artifact settings  
**And:** The minimal YAML file is uploaded to the target staging directory  
**And:** The remote ansible-navigator command includes both CLI flags and `--settings=/staging/ansible-navigator.yml`

#### Scenario: Empty navigator configuration
**Given:** NavigatorConfig is nil or unspecified  
**When:** Executing Provision()  
**Then:** No CLI flags are added for navigator configuration  
**And:** No YAML file is created or uploaded  
**And:** Default ansible-navigator behavior is used on the target

#### Scenario: Pull policy enforcement via CLI flag
**Given:** NavigatorConfig.ExecutionEnvironment.PullPolicy is `"never"`  
**When:** Executing Provision() on the target  
**Then:** The remote command includes `--pull-policy=never`  
**And:** The target's container engine does NOT attempt to pull the image  
**And:** Execution uses existing local images on the target only

#### Scenario: Multiple environment variables via CLI flags
**Given:** NavigatorConfig.ExecutionEnvironment.EnvironmentVariables.Set contains multiple variables  
**When:** Executing Provision()  
**Then:** Each variable is passed as a separate `--eev KEY=VALUE` flag in the remote command  
**And:** No YAML file is needed for environment variables

#### Scenario: Volume mounts via CLI flags
**Given:** NavigatorConfig.ExecutionEnvironment.VolumeMounts contains multiple mount specifications  
**When:** Executing Provision()  
**Then:** Each mount is passed as a separate `--evm src:dest:options` flag in the remote command  
**And:** No YAML file is needed for volume mounts

#### Scenario: Minimal YAML upload only when needed
**Given:** NavigatorConfig contains only CLI-mappable settings  
**When:** Preparing files to upload to target  
**Then:** The upload file list does NOT include ansible-navigator.yml  
**And:** Only playbook/role files and ansible.cfg are uploaded

#### Scenario: Minimal YAML cleanup after hybrid execution
**Given:** NavigatorConfig contained unmapped settings requiring minimal YAML  
**And:** The minimal YAML file was created locally and uploaded  
**When:** Provision() completes successfully or with error  
**Then:** The local temporary minimal YAML file is removed  
**And:** The remote minimal YAML file remains in staging directory (cleaned by Packer's cleanup)
