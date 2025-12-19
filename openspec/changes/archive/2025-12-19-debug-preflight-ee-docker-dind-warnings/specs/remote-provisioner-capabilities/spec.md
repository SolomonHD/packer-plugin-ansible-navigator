## ADDED Requirements

### Requirement: REQ-EE-DEBUG-001 Debug-only EE container-runtime preflight diagnostics (remote provisioner)

When plugin debug mode is enabled and `navigator_config.execution_environment.enabled = true`, the SSH-based provisioner SHALL emit DEBUG-only preflight diagnostics for container-runtime wiring on the machine running Packer.

#### Scenario: Preflight diagnostics are emitted when debug mode and EE are enabled
Given: a configuration for `provisioner "ansible-navigator"` where plugin debug mode is enabled
Given: `navigator_config.execution_environment.enabled = true`
When: the provisioner is about to invoke `ansible-navigator run` on the machine running Packer
Then: the provisioner SHALL emit DEBUG-only preflight diagnostics via Packer UI messages
Then: the diagnostics SHALL include whether `DOCKER_HOST` is unset or, if set, a safe representation of the value (e.g., redacting embedded credentials)
Then: the diagnostics SHALL include whether `/var/run/docker.sock` exists and whether it is a socket
Then: the diagnostics SHALL include whether the `docker` client is available in PATH
Then: the diagnostics SHALL NOT include unrelated environment variables

#### Scenario: Preflight diagnostics are NOT emitted when debug mode is disabled
Given: a configuration for `provisioner "ansible-navigator"` where plugin debug mode is disabled
When: the provisioner executes
Then: the new EE/Docker/DinD preflight diagnostics from this change SHALL NOT be emitted

#### Scenario: Preflight diagnostics are NOT emitted when EE is disabled
Given: a configuration for `provisioner "ansible-navigator"` where plugin debug mode is enabled
Given: `navigator_config.execution_environment.enabled` is unset or `false`
When: the provisioner executes
Then: the new EE/Docker/DinD preflight diagnostics from this change SHALL NOT be emitted

### Requirement: REQ-EE-DEBUG-002 Debug-only “likely DinD” warning heuristic (remote provisioner)

When plugin debug mode is enabled and `navigator_config.execution_environment.enabled = true`, the SSH-based provisioner SHALL emit a warning-only debug advisory when a `dockerd` process is detected, indicating a likely Docker-in-Docker setup.

#### Scenario: Dockerd presence emits a warning-only advisory
Given: a configuration for `provisioner "ansible-navigator"` where plugin debug mode is enabled
Given: `navigator_config.execution_environment.enabled = true`
Given: a `dockerd` process is detected in the same execution environment
When: the provisioner runs the EE preflight checks
Then: the provisioner SHALL emit a warning-labeled debug message (e.g. prefixed with `[DEBUG][WARN]`)
Then: the warning SHALL be advisory only and SHALL NOT hard-fail the build
Then: the warning SHALL include an actionable next step (e.g., advise using a host/remote daemon rather than Docker-in-Docker)

### Requirement: REQ-EE-DEBUG-003 Preflight checks are safe and non-blocking (remote provisioner)

The SSH-based provisioner's debug-only EE preflight checks SHALL be fast, non-blocking, and SHALL NOT change execution behavior beyond emitting Packer UI debug messages.

#### Scenario: Checks avoid slow/hanging docker operations
Given: a configuration for `provisioner "ansible-navigator"` where plugin debug mode is enabled
Given: `navigator_config.execution_environment.enabled = true`
When: the provisioner runs the EE preflight checks
Then: the checks SHALL be fast and non-blocking
Then: the checks SHALL NOT run potentially slow/hanging docker commands such as `docker info`
Then: the checks SHALL NOT change execution behavior beyond emitting debug UI messages
