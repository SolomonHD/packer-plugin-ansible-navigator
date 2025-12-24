package ansiblenavigator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSSHTunnelValidation tests the validation logic for SSH tunnel mode configuration
func TestSSHTunnelValidation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a temporary key file for testing
	tmpKeyFile := filepath.Join(tmpDir, "test_key")
	if err := os.WriteFile(tmpKeyFile, []byte("fake key"), 0600); err != nil {
		t.Fatalf("failed to create test key file: %v", err)
	}

	// Create a temporary playbook file for testing
	tmpPlaybook := filepath.Join(tmpDir, "playbook.yml")
	if err := os.WriteFile(tmpPlaybook, []byte("---\n- hosts: all\n  tasks: []\n"), 0644); err != nil {
		t.Fatalf("failed to create test playbook file: %v", err)
	}

	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorText   string
	}{
		{
			name: "Valid SSH tunnel configuration with key file",
			config: Config{
				ConnectionMode:        "ssh_tunnel",
				BastionHost:           "bastion.example.com",
				BastionPort:           22,
				BastionUser:           "deploy",
				BastionPrivateKeyFile: tmpKeyFile,
				Plays: []Play{
					{Target: tmpPlaybook},
				},
			},
			expectError: false,
		},
		{
			name: "Valid SSH tunnel configuration with password",
			config: Config{
				ConnectionMode:  "ssh_tunnel",
				BastionHost:     "bastion.example.com",
				BastionPort:     2222,
				BastionUser:     "deploy",
				BastionPassword: "secret",
				Plays: []Play{
					{Target: tmpPlaybook},
				},
			},
			expectError: false,
		},
		{
			name: "Invalid: connection_mode=ssh_tunnel but missing bastion_host",
			config: Config{
				ConnectionMode:  "ssh_tunnel",
				BastionUser:     "deploy",
				BastionPassword: "secret",
				Plays: []Play{
					{Target: tmpPlaybook},
				},
			},
			expectError: true,
			errorText:   "bastion_host is required when connection_mode='ssh_tunnel'",
		},
		{
			name: "Invalid: connection_mode=ssh_tunnel but missing bastion_user",
			config: Config{
				ConnectionMode:  "ssh_tunnel",
				BastionHost:     "bastion.example.com",
				BastionPassword: "secret",
				Plays: []Play{
					{Target: tmpPlaybook},
				},
			},
			expectError: true,
			errorText:   "bastion_user is required when connection_mode='ssh_tunnel'",
		},
		{
			name: "Invalid: connection_mode=ssh_tunnel but missing both key and password",
			config: Config{
				ConnectionMode: "ssh_tunnel",
				BastionHost:    "bastion.example.com",
				BastionUser:    "deploy",
				Plays: []Play{
					{Target: tmpPlaybook},
				},
			},
			expectError: true,
			errorText:   "either bastion_private_key_file or bastion_password must be provided",
		},
		{
			name: "Invalid: bastion_port out of range (too high)",
			config: Config{
				ConnectionMode:  "ssh_tunnel",
				BastionHost:     "bastion.example.com",
				BastionPort:     99999,
				BastionUser:     "deploy",
				BastionPassword: "secret",
				Plays: []Play{
					{Target: tmpPlaybook},
				},
			},
			expectError: true,
			errorText:   "bastion_port must be between 1 and 65535",
		},
		{
			name: "Invalid: bastion_port out of range (zero)",
			config: Config{
				ConnectionMode:  "ssh_tunnel",
				BastionHost:     "bastion.example.com",
				BastionPort:     0,
				BastionUser:     "deploy",
				BastionPassword: "secret",
				Plays: []Play{
					{Target: tmpPlaybook},
				},
			},
			expectError: true,
			errorText:   "bastion_port must be between 1 and 65535",
		},
		{
			name: "Invalid: bastion_private_key_file does not exist",
			config: Config{
				ConnectionMode:        "ssh_tunnel",
				BastionHost:           "bastion.example.com",
				BastionPort:           22,
				BastionUser:           "deploy",
				BastionPrivateKeyFile: "/nonexistent/path/to/key",
				Plays: []Play{
					{Target: tmpPlaybook},
				},
			},
			expectError: true,
			errorText:   "bastion_private_key_file",
		},
		{
			name: "Valid: proxy mode (default behavior)",
			config: Config{
				ConnectionMode: "proxy",
				Plays: []Play{
					{Target: tmpPlaybook},
				},
			},
			expectError: false,
		},
		{
			name: "Valid: direct mode",
			config: Config{
				ConnectionMode: "direct",
				Plays: []Play{
					{Target: tmpPlaybook},
				},
			},
			expectError: false,
		},
		{
			name: "Invalid: invalid connection_mode",
			config: Config{
				ConnectionMode: "invalid_mode",
				Plays: []Play{
					{Target: tmpPlaybook},
				},
			},
			expectError: true,
			errorText:   "connection_mode must be one of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected validation error but got none")
				} else if tt.errorText != "" && !strings.Contains(err.Error(), tt.errorText) {
					t.Errorf("Expected error to contain %q, got: %v", tt.errorText, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}
