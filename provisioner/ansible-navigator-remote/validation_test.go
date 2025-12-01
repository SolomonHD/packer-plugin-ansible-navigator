// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansiblenavigatorremote

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/template/config"
)

func TestConfigValidate(t *testing.T) {
	// Create temporary test files
	tmpDir := t.TempDir()
	validPlaybook := filepath.Join(tmpDir, "test.yml")
	validGalaxyFile := filepath.Join(tmpDir, "requirements.yml")

	if err := os.WriteFile(validPlaybook, []byte("---\n- hosts: all\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(validGalaxyFile, []byte("roles:\n  - role: test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid playbook file",
			config: Config{
				PlaybookFile: validPlaybook,
			},
			wantErr: false,
		},
		{
			name: "valid plays array",
			config: Config{
				Plays: []Play{
					{Target: validPlaybook},
				},
			},
			wantErr: false,
		},
		{
			name: "valid role FQDN",
			config: Config{
				Plays: []Play{
					{Target: "namespace.collection.role_name"},
				},
			},
			wantErr: false,
		},
		{
			name: "both playbook_file and plays specified",
			config: Config{
				PlaybookFile: validPlaybook,
				Plays: []Play{
					{Target: "test.role"},
				},
			},
			wantErr: true,
			errMsg:  "you may specify only one of",
		},
		{
			name:    "neither playbook_file nor plays specified",
			config:  Config{},
			wantErr: true,
			errMsg:  "must be defined",
		},
		{
			name: "play without target",
			config: Config{
				Plays: []Play{
					{Name: "test"},
				},
			},
			wantErr: true,
			errMsg:  "target must be specified",
		},
		{
			name: "invalid navigator mode",
			config: Config{
				PlaybookFile:  validPlaybook,
				NavigatorMode: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid navigator_mode",
		},
		{
			name: "valid navigator mode stdout",
			config: Config{
				PlaybookFile:  validPlaybook,
				NavigatorMode: "stdout",
			},
			wantErr: false,
		},
		{
			name: "valid navigator mode json",
			config: Config{
				PlaybookFile:  validPlaybook,
				NavigatorMode: "json",
			},
			wantErr: false,
		},
		{
			name: "valid navigator mode yaml",
			config: Config{
				PlaybookFile:  validPlaybook,
				NavigatorMode: "yaml",
			},
			wantErr: false,
		},
		{
			name: "valid navigator mode interactive",
			config: Config{
				PlaybookFile:  validPlaybook,
				NavigatorMode: "interactive",
			},
			wantErr: false,
		},
		{
			name: "invalid port too high",
			config: Config{
				PlaybookFile: validPlaybook,
				LocalPort:    70000,
			},
			wantErr: true,
			errMsg:  "must be a valid port",
		},
		{
			name: "valid port",
			config: Config{
				PlaybookFile: validPlaybook,
				LocalPort:    8080,
			},
			wantErr: false,
		},
		{
			name: "invalid adapter key type",
			config: Config{
				PlaybookFile:   validPlaybook,
				AdapterKeyType: "INVALID",
			},
			wantErr: true,
			errMsg:  "invalid value for ansible_proxy_key_type",
		},
		{
			name: "valid adapter key type RSA",
			config: Config{
				PlaybookFile:   validPlaybook,
				AdapterKeyType: "RSA",
			},
			wantErr: false,
		},
		{
			name: "valid adapter key type ECDSA",
			config: Config{
				PlaybookFile:   validPlaybook,
				AdapterKeyType: "ECDSA",
			},
			wantErr: false,
		},
		{
			name: "missing playbook file",
			config: Config{
				PlaybookFile: "/nonexistent/file.yml",
			},
			wantErr: true,
			errMsg:  "is invalid",
		},
		{
			name: "missing galaxy file",
			config: Config{
				PlaybookFile: validPlaybook,
				GalaxyFile:   "/nonexistent/galaxy.yml",
			},
			wantErr: true,
			errMsg:  "is invalid",
		},
		{
			name: "valid galaxy file",
			config: Config{
				PlaybookFile: validPlaybook,
				GalaxyFile:   validGalaxyFile,
			},
			wantErr: false,
		},
		{
			name: "multiple plays with different types",
			config: Config{
				Plays: []Play{
					{Target: validPlaybook},
					{Target: "namespace.collection.role"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Config.Validate() expected error but got nil")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Config.Validate() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Config.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestConfigValidate_InventoryDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	validPlaybook := filepath.Join(tmpDir, "test.yml")
	if err := os.WriteFile(validPlaybook, []byte("---\n- hosts: all\n"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		setupFunc func() string
		wantErr   bool
		errMsg    string
	}{
		{
			name: "valid inventory directory",
			setupFunc: func() string {
				dir := filepath.Join(tmpDir, "inventory")
				os.Mkdir(dir, 0755)
				return dir
			},
			wantErr: false,
		},
		{
			name: "inventory directory is file",
			setupFunc: func() string {
				return validPlaybook
			},
			wantErr: true,
			errMsg:  "must point to a directory",
		},
		{
			name: "nonexistent inventory directory",
			setupFunc: func() string {
				return "/nonexistent/directory"
			},
			wantErr: true,
			errMsg:  "is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				PlaybookFile:       validPlaybook,
				InventoryDirectory: tt.setupFunc(),
			}

			err := cfg.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Config.Validate() expected error but got nil")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Config.Validate() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Config.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestConfigValidate_SSHKeyFiles(t *testing.T) {
	tmpDir := t.TempDir()
	validPlaybook := filepath.Join(tmpDir, "test.yml")
	validKeyFile := filepath.Join(tmpDir, "key.pem")

	if err := os.WriteFile(validPlaybook, []byte("---\n- hosts: all\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(validKeyFile, []byte("fake key"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid ssh authorized key file",
			config: Config{
				PlaybookFile:         validPlaybook,
				SSHAuthorizedKeyFile: validKeyFile,
			},
			wantErr: false,
		},
		{
			name: "valid ssh host key file",
			config: Config{
				PlaybookFile:   validPlaybook,
				SSHHostKeyFile: validKeyFile,
			},
			wantErr: false,
		},
		{
			name: "missing ssh authorized key file",
			config: Config{
				PlaybookFile:         validPlaybook,
				SSHAuthorizedKeyFile: "/nonexistent/key.pem",
			},
			wantErr: true,
			errMsg:  "is invalid",
		},
		{
			name: "missing ssh host key file",
			config: Config{
				PlaybookFile:   validPlaybook,
				SSHHostKeyFile: "/nonexistent/host_key.pem",
			},
			wantErr: true,
			errMsg:  "is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Config.Validate() expected error but got nil")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Config.Validate() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Config.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestPrepare_Defaults(t *testing.T) {
	tmpDir := t.TempDir()
	validPlaybook := filepath.Join(tmpDir, "test.yml")
	if err := os.WriteFile(validPlaybook, []byte("---\n- hosts: all\n"), 0644); err != nil {
		t.Fatal(err)
	}

	p := &Provisioner{}
	config := map[string]interface{}{
		"playbook_file":      validPlaybook,
		"skip_version_check": true,
	}

	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("Prepare() unexpected error: %v", err)
	}

	// Check defaults
	if p.config.Command != "ansible-navigator run" {
		t.Errorf("Command = %v, want %v", p.config.Command, "ansible-navigator run")
	}

	if p.config.GalaxyCommand != "ansible-galaxy" {
		t.Errorf("GalaxyCommand = %v, want %v", p.config.GalaxyCommand, "ansible-galaxy")
	}

	if p.config.HostAlias != "default" {
		t.Errorf("HostAlias = %v, want %v", p.config.HostAlias, "default")
	}

	if p.config.NavigatorMode != "stdout" {
		t.Errorf("NavigatorMode = %v, want %v", p.config.NavigatorMode, "stdout")
	}

	if p.config.AdapterKeyType != "ECDSA" {
		t.Errorf("AdapterKeyType = %v, want %v", p.config.AdapterKeyType, "ECDSA")
	}
}

func TestPrepare_UseProxy(t *testing.T) {
	tmpDir := t.TempDir()
	validPlaybook := filepath.Join(tmpDir, "test.yml")
	if err := os.WriteFile(validPlaybook, []byte("---\n- hosts: all\n"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		useProxy  interface{}
		wantProxy config.Trilean
	}{
		{
			name:      "use_proxy unset defaults to true",
			useProxy:  nil,
			wantProxy: config.TriUnset,
		},
		{
			name:      "use_proxy explicitly true",
			useProxy:  true,
			wantProxy: config.TriTrue,
		},
		{
			name:      "use_proxy explicitly false",
			useProxy:  false,
			wantProxy: config.TriFalse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provisioner{}
			config := map[string]interface{}{
				"playbook_file":      validPlaybook,
				"skip_version_check": true,
			}
			if tt.useProxy != nil {
				config["use_proxy"] = tt.useProxy
			}

			err := p.Prepare(config)
			if err != nil {
				t.Fatalf("Prepare() unexpected error: %v", err)
			}

			if p.config.UseProxy != tt.wantProxy {
				t.Errorf("UseProxy = %v, want %v", p.config.UseProxy, tt.wantProxy)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
