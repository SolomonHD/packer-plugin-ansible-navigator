# ssh-tunnel-runtime Spec Delta

## ADDED Requirements

### Requirement: Target Port Type Handling

The provisioner SHALL accept target port from `generatedData["Port"]` as either `int` or `string` type and convert both to a validated integer port number.

#### Scenario: Port provided as int type

- **GIVEN** `ssh_tunnel_mode = true`
- **AND** `generatedData["Port"]` is an `int` with value `22`
- **WHEN** the Provision() function extracts the target port
- **THEN** it SHALL successfully assign the port as `22`
- **AND** no type conversion SHALL be needed
- **AND** tunnel setup SHALL proceed with port `22`

#### Scenario: Port provided as string type

- **GIVEN** `ssh_tunnel_mode = true`
- **AND** `generatedData["Port"]` is a `string` with value `"22"`
- **WHEN** the Provision() function extracts the target port
- **THEN** it SHALL parse the string using `strconv.Atoi()`
- **AND** the parsed port SHALL be `22`
- **AND** tunnel setup SHALL proceed with port `22`

#### Scenario: Custom port as string

- **GIVEN** `ssh_tunnel_mode = true`
- **AND** `generatedData["Port"]` is a `string` with value `"2222"`
- **WHEN** the Provision() function extracts the target port
- **THEN** it SHALL parse the string successfully
- **AND** the parsed port SHALL be `2222`
- **AND** tunnel setup SHALL proceed with port `2222`

### Requirement: Port Value Validation

The provisioner SHALL validate that extracted port values are within the valid TCP port range.

#### Scenario: Valid port within range

- **GIVEN** extracted target port is `22`
- **WHEN** port validation is performed
- **THEN** validation SHALL succeed
- **AND** tunnel setup SHALL proceed

#### Scenario: Port below valid range

- **GIVEN** extracted target port is `0`
- **WHEN** port validation is performed
- **THEN** validation SHALL fail
- **AND** an error SHALL be returned containing "port must be between 1-65535, got 0"

#### Scenario: Port above valid range

- **GIVEN** extracted target port is `99999`
- **WHEN** port validation is performed
- **THEN** validation SHALL fail
- **AND** an error SHALL be returned containing "port must be between 1-65535, got 99999"

#### Scenario: Negative port value

- **GIVEN** extracted target port is `-1`
- **WHEN** port validation is performed
- **THEN** validation SHALL fail
- **AND** an error SHALL be returned containing "port must be between 1-65535"

### Requirement: Port Extraction Error Handling

The provisioner SHALL provide clear, actionable error messages for all port extraction failure modes.

#### Scenario: Port missing from generatedData

- **GIVEN** `ssh_tunnel_mode = true`
- **AND** `generatedData["Port"]` is not set (nil)
- **WHEN** the Provision() function attempts to extract the target port
- **THEN** it SHALL return an error
- **AND** the error message SHALL contain "Port must be int or string, got type <nil>"
- **AND** the error message SHALL include the actual type and value for debugging

#### Scenario: Port as invalid string format

- **GIVEN** `ssh_tunnel_mode = true`
- **AND** `generatedData["Port"]` is a `string` with value `"abc"`
- **WHEN** the Provision() function attempts to parse the port
- **THEN** parsing SHALL fail
- **AND** an error SHALL be returned containing "invalid port value \"abc\""
- **AND** the error SHALL wrap the underlying `strconv.Atoi()` error

#### Scenario: Port as unsupported type

- **GIVEN** `ssh_tunnel_mode = true`
- **AND** `generatedData["Port"]` is a `float64` with value `22.5`
- **WHEN** the Provision() function attempts to extract the target port
- **THEN** it SHALL return an error
- **AND** the error message SHALL contain "Port must be int or string, got type float64 with value 22.5"

#### Scenario: Port as empty string

- **GIVEN** `ssh_tunnel_mode = true`
- **AND** `generatedData["Port"]` is a `string` with value `""`
- **WHEN** the Provision() function attempts to parse the port
- **THEN** parsing SHALL fail
- **AND** an error SHALL be returned containing "invalid port value \"\""
- **AND** the error SHALL wrap the underlying parsing error

#### Scenario: Port string with whitespace

- **GIVEN** `ssh_tunnel_mode = true`
- **AND** `generatedData["Port"]` is a `string` with value `" 22 "`
- **WHEN** the Provision() function attempts to parse the port
- **THEN** parsing MAY fail due to whitespace (not trimmed)
- **OR** implementation MAY choose to trim whitespace before parsing
- **AND** if parsing fails, error message SHALL indicate invalid format

## MODIFIED Requirements

### Requirement: Integration with Provision Flow

The provisioner SHALL integrate SSH tunnel setup into the provisioning lifecycle with type-safe Port extraction.

#### Scenario: Tunnel replaces proxy adapter when enabled (UPDATED)

- **GIVEN** a configuration with `ssh_tunnel_mode = true`
- **WHEN** Provision() is called
- **THEN** it SHALL extract `generatedData["Port"]` using type-safe handling (int or string)
- **AND** it SHALL validate the extracted port is between 1-65535
- **AND** if extraction/validation succeeds, it SHALL NOT call setupAdapter()
- **AND** it SHALL call setupSSHTunnel() instead with the validated port
- **AND** `generatedData["Host"]` SHALL be overridden to "127.0.0.1"
- **AND** `generatedData["Port"]` SHALL be overridden to the tunnel's local port

## Implementation Notes

The fix SHALL use a Go type switch pattern:

```go
var targetPort int
switch v := generatedData["Port"].(type) {
case int:
    targetPort = v
case string:
    var err error
    targetPort, err = strconv.Atoi(v)
    if err != nil {
        return fmt.Errorf("SSH tunnel mode: invalid port value %q: %w", v, err)
    }
default:
    return fmt.Errorf("SSH tunnel mode: Port must be int or string, got type %T with value %v", v, v)
}

if targetPort < 1 || targetPort > 65535 {
    return fmt.Errorf("SSH tunnel mode: port must be between 1-65535, got %d", targetPort)
}
```

This pattern is idiomatic Go and provides clear error messages for debugging.
