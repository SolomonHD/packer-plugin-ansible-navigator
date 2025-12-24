# Spec: SSH Tunnel Runtime - Port Type Handling

## MODIFIED Requirements

### Requirement: Target Port Type Handling

The provisioner MUST correctly extract the target port from Packer's `generatedData["Port"]` field for all integer types that Packer builders may provide.

#### Scenario: Port provided as int

Given: `generatedData["Port"]` is set to `int(22)`  
When: SSH tunnel mode is enabled  
Then: `targetPort` is successfully set to `22`

#### Scenario: Port provided as int64

Given: `generatedData["Port"]` is set to `int64(22)`  
When: SSH tunnel mode is enabled  
Then: `targetPort` is successfully set to `22`

#### Scenario: Port provided as int32

Given: `generatedData["Port"]` is set to `int32(2222)`  
When: SSH tunnel mode is enabled  
Then: `targetPort` is successfully set to `2222`

#### Scenario: Port provided as int16

Given: `generatedData["Port"]` is set to `int16(3333)`  
When: SSH tunnel mode is enabled  
Then: `targetPort` is successfully set to `3333`

#### Scenario: Port provided as int8

Given: `generatedData["Port"]` is set to `int8(80)`  
When: SSH tunnel mode is enabled  
Then: `targetPort` is successfully set to `80`

#### Scenario: Port provided as uint

Given: `generatedData["Port"]` is set to `uint(8080)`  
When: SSH tunnel mode is enabled  
Then: `targetPort` is successfully set to `8080`

#### Scenario: Port provided as uint64

Given: `generatedData["Port"]` is set to `uint64(443)`  
When: SSH tunnel mode is enabled  
Then: `targetPort` is successfully set to `443`

#### Scenario: Port provided as uint32

Given: `generatedData["Port"]` is set to `uint32(8000)`  
When: SSH tunnel mode is enabled  
Then: `targetPort` is successfully set to `8000`

#### Scenario: Port provided as uint16

Given: `generatedData["Port"]` is set to `uint16(9090)`  
When: SSH tunnel mode is enabled  
Then: `targetPort` is successfully set to `9090`

#### Scenario: Port provided as uint8

Given: `generatedData["Port"]` is set to `uint8(80)`  
When: SSH tunnel mode is enabled  
Then: `targetPort` is successfully set to `80`

#### Scenario: Port provided as string

Given: `generatedData["Port"]` is set to `"2222"`  
When: SSH tunnel mode is enabled  
Then: `targetPort` is successfully parsed to `2222`

#### Scenario: Invalid string port value

Given: `generatedData["Port"]` is set to `"not-a-number"`  
When: SSH tunnel mode is enabled  
Then: The provisioner returns an error "SSH tunnel mode: invalid port value \"not-a-number\""

#### Scenario: Unsigned port exceeding maximum

Given: `generatedData["Port"]` is set to `uint64(70000)`  
When: SSH tunnel mode is enabled  
Then: The provisioner returns an error "SSH tunnel mode: port value 70000 exceeds maximum 65535"

#### Scenario: Unsupported type

Given: `generatedData["Port"]` is set to `float64(22.5)`
When: SSH tunnel mode is enabled
Then: The provisioner returns an error indicating unsupported type

#### Scenario: Existing int configuration still works

Given: An existing Packer configuration with Port provided as `int(22)`
When: The provisioner runs with the updated code
Then: The tunnel is established successfully without errors

#### Scenario: Existing string configuration still works

Given: An existing Packer configuration with Port provided as `"2222"`
When: The provisioner runs with the updated code
Then: The tunnel is established successfully without errors

### Requirement: Port Value Validation

The provisioner MUST validate that the extracted port value is within the valid TCP port range (1-65535) regardless of the source type.

#### Scenario: Port value within valid range

Given: `generatedData["Port"]` is `int64(443)`  
When: Port extraction completes  
Then: Validation passes and tunnel setup continues

#### Scenario: Port value below minimum

Given: `generatedData["Port"]` is `int64(0)`  
When: Port extraction completes  
Then: The provisioner returns an error "SSH tunnel mode: port must be between 1-65535, got 0"

#### Scenario: Port value above maximum

Given: `generatedData["Port"]` is `int64(70000)`  
When: Port extraction completes  
Then: The provisioner returns an error "SSH tunnel mode: port must be between 1-65535, got 70000"

#### Scenario: Negative port value

Given: `generatedData["Port"]` is `int64(-22)`
When: Port extraction completes
Then: The provisioner returns an error "SSH tunnel mode: port must be between 1-65535, got -22"

### Requirement: Port Extraction Error Handling

The provisioner MUST provide clear, actionable error messages that indicate the specific failure mode when port extraction fails.

#### Scenario: Error message indicates type mismatch

Given: `generatedData["Port"]` is an unsupported type `bool(true)`  
When: Port extraction fails  
Then: The error message includes "got type bool with value true"

#### Scenario: Error message indicates invalid string format

Given: `generatedData["Port"]` is `"abc"`  
When: Port extraction fails  
Then: The error message includes "invalid port value \"abc\""

#### Scenario: Error message indicates range violation

Given: `generatedData["Port"]` is `uint64(100000)`  
When: Port extraction fails  
Then: The error message includes "exceeds maximum 65535"
