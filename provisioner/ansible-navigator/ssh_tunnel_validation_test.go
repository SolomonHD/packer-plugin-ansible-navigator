package ansiblenavigator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/template/config"
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
				SSHTunnelMode:         true,
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
				SSHTunnelMode:   true,
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
			name: "Invalid: ssh_tunnel_mode=true and use_proxy=true",
			config: Config{
				SSHTunnelMode:   true,
				UseProxy:        config.TriTrue,
				BastionHost:     "bastion.example.com",
				BastionUser:     "deploy",
				BastionPassword: "secret",
				Plays: []Play{
					{Target: tmpPlaybook},
				},
			},
			expectError: true,
			errorText:   "ssh_tunnel_mode and use_proxy cannot both be true",
		},
		{
			name: "Invalid: ssh_tunnel_mode=true but missing bastion_host",
			config: Config{
				SSHTunnelMode:   true,
				BastionUser:     "deploy",
				BastionPassword: "secret",
				Plays: []Play{
					{Target: tmpPlaybook},
				},
			},
			expectError: true,
			errorText:   "bastion_host is required when ssh_tunnel_mode is true",
		},
		{
			name: "Invalid: ssh_tunnel_mode=true but missing bastion_user",
			config: Config{
				SSHTunnelMode:   true,
				BastionHost:     "bastion.example.com",
				BastionPassword: "secret",
				Plays: []Play{
					{Target: tmpPlaybook},
				},
			},
			expectError: true,
			errorText:   "bastion_user is required when ssh_tunnel_mode is true",
		},
		{
			name: "Invalid: ssh_tunnel_mode=true but missing both key and password",
			config: Config{
				SSHTunnelMode: true,
				BastionHost:   "bastion.example.com",
				BastionUser:   "deploy",
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
				SSHTunnelMode:   true,
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
				SSHTunnelMode:   true,
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
				SSHTunnelMode:         true,
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
			name: "Valid: SSH tunnel disabled (default behavior)",
			config: Config{
				SSHTunnelMode: false,
				Plays: []Play{
					{Target: tmpPlaybook},
				},
			},
			expectError: false,
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
