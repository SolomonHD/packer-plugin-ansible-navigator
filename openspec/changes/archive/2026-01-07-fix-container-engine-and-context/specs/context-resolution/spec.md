## ADDED Requirements

### Requirement: Docker Context Resolution

The provisioner MUST automatically detect and use the active Docker Context when running in environments where `DOCKER_HOST` is not explicitly set, ensuring compatibility with Rootless Docker and non-default contexts.

#### Scenario: Detect Docker Host
Given: `DOCKER_HOST` environment variable is NOT set
And: `docker` command is available
When: The provisioner prepares to execute `ansible-navigator`
Then: It MUST execute `docker context inspect --format '{{.Endpoints.docker.Host}}'`
And: If the command succeeds and returns a non-empty value, it MUST set `DOCKER_HOST` in the `ansible-navigator` process environment

#### Scenario: Existing Docker Host
Given: `DOCKER_HOST` environment variable IS set
When: The provisioner prepares to execute `ansible-navigator`
Then: It MUST NOT modify `DOCKER_HOST` (respect existing value)

#### Scenario: Docker Not Available
Given: `docker` command is NOT available
When: The provisioner attempts to resolve Docker host
Then: It MUST proceed without setting `DOCKER_HOST` (fail open)
