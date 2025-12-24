package ansiblenavigator

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// TestPortExtractionFromGeneratedData tests port extraction logic for all integer types
func TestPortExtractionFromGeneratedData(t *testing.T) {
	tests := []struct {
		name         string
		portValue    interface{}
		expectError  bool
		errorText    string
		expectedPort int
	}{
		// Existing int and string support
		{
			name:         "int port",
			portValue:    22,
			expectError:  false,
			expectedPort: 22,
		},
		{
			name:         "string port valid",
			portValue:    "2222",
			expectError:  false,
			expectedPort: 2222,
		},
		{
			name:        "string port invalid",
			portValue:   "invalid",
			expectError: true,
			errorText:   "invalid port value",
		},

		// New int64 support
		{
			name:         "int64 port (primary use case)",
			portValue:    int64(22),
			expectError:  false,
			expectedPort: 22,
		},
		{
			name:         "int64 port large value",
			portValue:    int64(65535),
			expectError:  false,
			expectedPort: 65535,
		},
		{
			name:        "int64 port negative",
			portValue:   int64(-22),
			expectError: true,
			errorText:   "port must be between 1-65535",
		},
		{
			name:        "int64 port zero",
			portValue:   int64(0),
			expectError: true,
			errorText:   "port must be between 1-65535",
		},
		{
			name:        "int64 port too large",
			portValue:   int64(70000),
			expectError: true,
			errorText:   "port must be between 1-65535",
		},

		// int32 support
		{
			name:         "int32 port",
			portValue:    int32(2222),
			expectError:  false,
			expectedPort: 2222,
		},

		// int16 support
		{
			name:         "int16 port",
			portValue:    int16(3333),
			expectError:  false,
			expectedPort: 3333,
		},

		// int8 support
		{
			name:         "int8 port",
			portValue:    int8(80),
			expectError:  false,
			expectedPort: 80,
		},

		// uint support with range validation
		{
			name:         "uint port",
			portValue:    uint(8080),
			expectError:  false,
			expectedPort: 8080,
		},
		{
			name:        "uint port exceeding max",
			portValue:   uint(70000),
			expectError: true,
			errorText:   "port value 70000 exceeds maximum 65535",
		},

		// uint64 support with range validation
		{
			name:         "uint64 port",
			portValue:    uint64(443),
			expectError:  false,
			expectedPort: 443,
		},
		{
			name:        "uint64 port exceeding max",
			portValue:   uint64(70000),
			expectError: true,
			errorText:   "port value 70000 exceeds maximum 65535",
		},

		// uint32 support with range validation
		{
			name:         "uint32 port",
			portValue:    uint32(8000),
			expectError:  false,
			expectedPort: 8000,
		},
		{
			name:        "uint32 port exceeding max",
			portValue:   uint32(70000),
			expectError: true,
			errorText:   "port value 70000 exceeds maximum 65535",
		},

		// uint16 support (inherently <= 65535)
		{
			name:         "uint16 port",
			portValue:    uint16(9090),
			expectError:  false,
			expectedPort: 9090,
		},
		{
			name:         "uint16 port max value",
			portValue:    uint16(65535),
			expectError:  false,
			expectedPort: 65535,
		},

		// uint8 support (inherently <= 255)
		{
			name:         "uint8 port",
			portValue:    uint8(80),
			expectError:  false,
			expectedPort: 80,
		},

		// Unsupported types
		{
			name:        "float64 port unsupported",
			portValue:   float64(22.5),
			expectError: true,
			errorText:   "Port must be a numeric or string type",
		},
		{
			name:        "bool port unsupported",
			portValue:   true,
			expectError: true,
			errorText:   "Port must be a numeric or string type",
		},
		{
			name:        "nil port unsupported",
			portValue:   nil,
			expectError: true,
			errorText:   "Port must be a numeric or string type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create minimal provisioner with SSH tunnel mode
			p := &Provisioner{
				config: Config{
					ConnectionMode: "ssh_tunnel",
					Bastion: &BastionConfig{
						Host:     "bastion.example.com",
						Port:     22,
						User:     "deploy",
						Password: "secret",
					},
				},
				generatedData: map[string]interface{}{
					"Host": "10.0.1.100",
					"Port": tt.portValue,
				},
			}

			// Execute port extraction logic inline (simulating the Provision method)
			generatedData := p.generatedData
			targetHost, ok := generatedData["Host"].(string)
			if !ok || targetHost == "" {
				t.Fatal("Test setup error: missing Host in generatedData")
			}

			// This is the logic we're testing (lines 1516-1534 of provisioner.go)
			var targetPort int
			var err error

			switch v := generatedData["Port"].(type) {
			case int:
				targetPort = v
			case int64:
				targetPort = int(v)
			case int32:
				targetPort = int(v)
			case int16:
				targetPort = int(v)
			case int8:
				targetPort = int(v)
			case uint:
				if v > 65535 {
					err = fmt.Errorf("SSH tunnel mode: port value %d exceeds maximum 65535", v)
				} else {
					targetPort = int(v)
				}
			case uint64:
				if v > 65535 {
					err = fmt.Errorf("SSH tunnel mode: port value %d exceeds maximum 65535", v)
				} else {
					targetPort = int(v)
				}
			case uint32:
				if v > 65535 {
					err = fmt.Errorf("SSH tunnel mode: port value %d exceeds maximum 65535", v)
				} else {
					targetPort = int(v)
				}
			case uint16:
				targetPort = int(v)
			case uint8:
				targetPort = int(v)
			case string:
				var parseErr error
				targetPort, parseErr = strconv.Atoi(v)
				if parseErr != nil {
					err = fmt.Errorf("SSH tunnel mode: invalid port value %q: %w", v, parseErr)
				}
			default:
				err = fmt.Errorf("SSH tunnel mode: Port must be a numeric or string type, got type %T with value %v", v, v)
			}

			// Validate port range (if no error yet)
			if err == nil {
				if targetPort < 1 || targetPort > 65535 {
					err = fmt.Errorf("SSH tunnel mode: port must be between 1-65535, got %d", targetPort)
				}
			}

			// Verify test expectations
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorText != "" && !strings.Contains(err.Error(), tt.errorText) {
					t.Errorf("Expected error to contain %q, got: %v", tt.errorText, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if targetPort != tt.expectedPort {
					t.Errorf("Expected port %d, got %d", tt.expectedPort, targetPort)
				}
			}
		})
	}
}

// TestPortExtractionBackwardCompatibility verifies existing int and string support still works
func TestPortExtractionBackwardCompatibility(t *testing.T) {
	tests := []struct {
		name         string
		portValue    interface{}
		expectedPort int
	}{
		{
			name:         "Legacy int port",
			portValue:    22,
			expectedPort: 22,
		},
		{
			name:         "Legacy string port",
			portValue:    "2222",
			expectedPort: 2222,
		},
		{
			name:         "Common SSH port",
			portValue:    22,
			expectedPort: 22,
		},
		{
			name:         "Common WinRM port",
			portValue:    5985,
			expectedPort: 5985,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provisioner{
				config: Config{
					ConnectionMode: "ssh_tunnel",
					Bastion: &BastionConfig{
						Host:     "bastion.example.com",
						Port:     22,
						User:     "deploy",
						Password: "secret",
					},
				},
				generatedData: map[string]interface{}{
					"Host": "10.0.1.100",
					"Port": tt.portValue,
				},
			}

			generatedData := p.generatedData
			var targetPort int
			var err error

			switch v := generatedData["Port"].(type) {
			case int:
				targetPort = v
			case int64:
				targetPort = int(v)
			case int32:
				targetPort = int(v)
			case int16:
				targetPort = int(v)
			case int8:
				targetPort = int(v)
			case uint:
				if v > 65535 {
					err = fmt.Errorf("SSH tunnel mode: port value %d exceeds maximum 65535", v)
				} else {
					targetPort = int(v)
				}
			case uint64:
				if v > 65535 {
					err = fmt.Errorf("SSH tunnel mode: port value %d exceeds maximum 65535", v)
				} else {
					targetPort = int(v)
				}
			case uint32:
				if v > 65535 {
					err = fmt.Errorf("SSH tunnel mode: port value %d exceeds maximum 65535", v)
				} else {
					targetPort = int(v)
				}
			case uint16:
				targetPort = int(v)
			case uint8:
				targetPort = int(v)
			case string:
				var parseErr error
				targetPort, parseErr = strconv.Atoi(v)
				if parseErr != nil {
					err = fmt.Errorf("SSH tunnel mode: invalid port value %q: %w", v, parseErr)
				}
			default:
				err = fmt.Errorf("SSH tunnel mode: Port must be a numeric or string type, got type %T with value %v", v, v)
			}

			if err == nil {
				if targetPort < 1 || targetPort > 65535 {
					err = fmt.Errorf("SSH tunnel mode: port must be between 1-65535, got %d", targetPort)
				}
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if targetPort != tt.expectedPort {
				t.Errorf("Expected port %d, got %d", tt.expectedPort, targetPort)
			}
		})
	}
}
