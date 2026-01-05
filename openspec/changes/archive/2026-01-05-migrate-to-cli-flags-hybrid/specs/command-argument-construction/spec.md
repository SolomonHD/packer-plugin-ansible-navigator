# command-argument-construction Spec Delta

## ADDED Requirements

### Requirement: Navigator CLI Flag Generation

The plugin MUST generate ansible-navigator CLI flags from NavigatorConfig fields as the primary configuration method, minimizing YAML file generation.

#### Scenario: Mode flag construction
**Given:** NavigatorConfig specifies `Mode = "stdout"`  
**When:** Building ansible-navigator command arguments  
**Then:** The argument list includes `--mode=stdout`  
**And:** No YAML file is generated for this setting

#### Scenario: Execution environment image flag
**Given:** ExecutionEnvironment.Image is set to `"quay.io/ansible/creator-ee:latest"`  
**When:** Building ansible-navigator command arguments  
**Then:** The argument list includes `--execution-environment-image=quay.io/ansible/creator-ee:latest`

#### Scenario: Pull policy flag
**Given:** ExecutionEnvironment.PullPolicy is set to `"never"`  
**When:** Building ansible-navigator command arguments  
**Then:** The argument list includes `--pull-policy=never`  
**And:** Docker pull is effectively prevented during execution

#### Scenario: Container engine selection
**Given:** ExecutionEnvironment.ContainerEngine is set to `"docker"`  
**When:** Building ansible-navigator command arguments  
**Then:** The argument list includes `--execution-environment-container-engine=docker`

#### Scenario: Execution environment enabled flag
**Given:** ExecutionEnvironment.Enabled is set to `true`  
**When:** Building ansible-navigator command arguments  
**Then:** The argument list includes `--execution-environment=true`

#### Scenario: Repeatable environment variables
**Given:** ExecutionEnvironment.EnvironmentVariables.Set contains `{"ANSIBLE_REMOTE_TMP": "/tmp/.ansible/tmp", "HOME": "/tmp"}`  
**When:** Building ansible-navigator command arguments  
**Then:** The argument list includes `--eev=ANSIBLE_REMOTE_TMP=/tmp/.ansible/tmp`  
**And:** The argument list includes `--eev=HOME=/tmp`  
**And:** Each variable appears as a separate `--eev` flag

#### Scenario: Passthrough environment variables
**Given:** ExecutionEnvironment.EnvironmentVariables.Pass contains `["PATH", "USER"]`  
**When:** Building ansible-navigator command arguments  
**Then:** The argument list includes `--eev=PATH`  
**And:** The argument list includes `--eev=USER`  
**And:** Each variable is passed without a value

#### Scenario: Repeatable volume mounts
**Given:** ExecutionEnvironment.VolumeMounts contains `["/host/collections:/tmp/collections:ro", "/host/roles:/tmp/roles:ro"]`  
**When:** Building ansible-navigator command arguments  
**Then:** The argument list includes `--evm=/host/collections:/tmp/collections:ro`  
**And:** The argument list includes `--evm=/host/roles:/tmp/roles:ro`  
**And:** Each mount appears as a separate `--evm` flag

#### Scenario: Container options
**Given:** ExecutionEnvironment.ContainerOptions contains `"--network=host --privileged"`  
**When:** Building ansible-navigator command arguments  
**Then:** The argument list includes `--container-options=--network=host --privileged`

#### Scenario: Logging level flag
**Given:** Logging.Level is set to `"info"`  
**When:** Building ansible-navigator command arguments  
**Then:** The argument list includes `--log-level=info`

#### Scenario: Logging file flag
**Given:** Logging.File is set to `"/tmp/navigator.log"`  
**When:** Building ansible-navigator command arguments  
**Then:** The argument list includes `--log-file=/tmp/navigator.log`

#### Scenario: Logging append flag
**Given:** Logging.Append is set to `true`  
**When:** Building ansible-navigator command arguments  
**Then:** The argument list includes `--log-append=true`

#### Scenario: Ansible config path flag
**Given:** AnsibleConfig.Config is set to `"/tmp/ansible.cfg"`  
**When:** Building ansible-navigator command arguments  
**Then:** The argument list includes `--ansible-config=/tmp/ansible.cfg`

### Requirement: Minimal YAML Fallback Generation

The plugin MUST generate minimal YAML configuration files ONLY when NavigatorConfig contains settings that lack CLI flag equivalents.

#### Scenario: No unmapped settings (CLI-only path)
**Given:** NavigatorConfig contains only settings with CLI flag equivalents (mode, execution environment, logging)  
**When:** Preparing to execute ansible-navigator  
**Then:** No YAML file is generated  
**And:** The command uses only CLI flags  
**And:** No `--settings` flag is added to the command

#### Scenario: Playbook artifact configured (hybrid path)
**Given:** NavigatorConfig.PlaybookArtifact is configured with `{ Enable: true, SaveAs: "/path/artifact.json" }`  
**And:** Other settings have CLI flag equivalents  
**When:** Preparing to execute ansible-navigator  
**Then:** A minimal YAML file is generated containing ONLY playbook-artifact settings  
**And:** The command uses CLI flags for mapped settings  
**And:** The command includes `--settings=/path/to/minimal-config.yml`

#### Scenario: Collection doc cache configured (hybrid path)
**Given:** NavigatorConfig.CollectionDocCache is configured  
**And:** Other settings have CLI flag equivalents  
**When:** Preparing to execute ansible-navigator  
**Then:** A minimal YAML file is generated containing ONLY collection-doc-cache settings  
**And:** The command uses CLI flags for mapped settings  
**And:** The command includes `--settings=/path/to/minimal-config.yml`

#### Scenario: Multiple unmapped settings
**Given:** NavigatorConfig contains both PlaybookArtifact and CollectionDocCache settings  
**When:** Generating minimal YAML  
**Then:** The YAML includes both playbook-artifact and collection-doc-cache sections  
**And:** No other sections are included in the YAML

### Requirement: Hybrid Command Construction

The plugin MUST combine CLI flags and optional minimal YAML settings files correctly.

#### Scenario: CLI flags before settings file
**Given:** Both CLI flags and minimal YAML are required  
**When:** Constructing the complete ansible-navigator command  
**Then:** CLI flags appear before the `--settings` flag  
**And:** The positional playbook argument appears last

#### Scenario: Command debugging visibility
**Given:** A command is constructed with CLI flags  
**When:** The provisioner logs the command to be executed  
**Then:** All CLI flags are visible in the log output  
**And:** Users can copy-paste the command for debugging
