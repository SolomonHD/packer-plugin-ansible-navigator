// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansiblenavigator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandUserPath(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err, "Failed to get home directory")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Bare tilde expands to HOME",
			input:    "~",
			expected: home,
		},
		{
			name:     "Tilde with subdirectory",
			input:    "~/ansible/playbooks",
			expected: filepath.Join(home, "ansible/playbooks"),
		},
		{
			name:     "Tilde with single level",
			input:    "~/bin",
			expected: filepath.Join(home, "bin"),
		},
		{
			name:     "Multi-user home preserved",
			input:    "~otheruser/files",
			expected: "~otheruser/files",
		},
		{
			name:     "Absolute path unchanged",
			input:    "/usr/local/bin/ansible",
			expected: "/usr/local/bin/ansible",
		},
		{
			name:     "Relative path unchanged",
			input:    "./local/ansible",
			expected: "./local/ansible",
		},
		{
			name:     "Empty string unchanged",
			input:    "",
			expected: "",
		},
		{
			name:     "Just filename unchanged",
			input:    "ansible-navigator",
			expected: "ansible-navigator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandUserPath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildEnvWithPath(t *testing.T) {
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	tests := []struct {
		name                 string
		ansibleNavigatorPath []string
		wantPathPrefix       []string
		wantContainsOriginal bool
	}{
		{
			name:                 "Empty path list returns original environment",
			ansibleNavigatorPath: []string{},
			wantPathPrefix:       nil,
			wantContainsOriginal: true,
		},
		{
			name:                 "Single directory prepended",
			ansibleNavigatorPath: []string{"/opt/ansible/bin"},
			wantPathPrefix:       []string{"/opt/ansible/bin"},
			wantContainsOriginal: true,
		},
		{
			name:                 "Multiple directories prepended in order",
			ansibleNavigatorPath: []string{"/opt/ansible/bin", "/usr/local/ansible/bin"},
			wantPathPrefix:       []string{"/opt/ansible/bin", "/usr/local/ansible/bin"},
			wantContainsOriginal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := buildEnvWithPath(tt.ansibleNavigatorPath)

			// Find PATH in environment
			var pathValue string
			for _, envVar := range env {
				if strings.HasPrefix(envVar, "PATH=") {
					pathValue = strings.TrimPrefix(envVar, "PATH=")
					break
				}
			}

			if len(tt.wantPathPrefix) == 0 {
				// Should contain original PATH
				if tt.wantContainsOriginal {
					assert.Contains(t, pathValue, originalPath)
				}
			} else {
				// Check that new directories are at the beginning
				for _, dir := range tt.wantPathPrefix {
					assert.Contains(t, pathValue, dir)
				}

				// Check they appear before original PATH
				if tt.wantContainsOriginal && originalPath != "" {
					assert.Contains(t, pathValue, originalPath)
					// First entry should be from our list
					pathParts := strings.Split(pathValue, string(os.PathListSeparator))
					assert.GreaterOrEqual(t, len(pathParts), len(tt.wantPathPrefix))
					for i, expectedDir := range tt.wantPathPrefix {
						assert.Equal(t, expectedDir, pathParts[i])
					}
				}
			}
		})
	}
}

func TestConfigValidateCommandWithWhitespace(t *testing.T) {
	tmpDir := t.TempDir()
	validPlaybook := filepath.Join(tmpDir, "test.yml")
	if err := os.WriteFile(validPlaybook, []byte("---\n- hosts: all\n"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		command string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Valid command without spaces",
			command: "ansible-navigator",
			wantErr: false,
		},
		{
			name:    "Valid command as path",
			command: "/usr/local/bin/ansible-navigator",
			wantErr: false,
		},
		{
			name:    "Invalid command with space",
			command: "ansible-navigator run",
			wantErr: true,
			errMsg:  "command must be only the executable name or path (no arguments)",
		},
		{
			name:    "Invalid command with multiple spaces",
			command: "ansible-navigator --mode json",
			wantErr: true,
			errMsg:  "Found whitespace in",
		},
		{
			name:    "Invalid command with tab",
			command: "ansible-navigator\trun",
			wantErr: true,
			errMsg:  "command must be only the executable name or path",
		},
		{
			name:    "Invalid command with leading space",
			command: " ansible-navigator",
			wantErr: true,
			errMsg:  "Found whitespace in",
		},
		{
			name:    "Invalid command with trailing space",
			command: "ansible-navigator ",
			wantErr: true,
			errMsg:  "Found whitespace in",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				Command: tt.command,
				Plays:   []Play{{Target: validPlaybook}},
			}

			err := cfg.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPrepareWithAnsibleNavigatorPath(t *testing.T) {
	tmpDir := t.TempDir()
	validPlaybook := filepath.Join(tmpDir, "test.yml")
	if err := os.WriteFile(validPlaybook, []byte("---\n- hosts: all\n"), 0644); err != nil {
		t.Fatal(err)
	}

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		name                 string
		ansibleNavigatorPath []string
		wantExpandedPath     []string
	}{
		{
			name:                 "Absolute paths unchanged",
			ansibleNavigatorPath: []string{"/opt/ansible/bin", "/usr/local/bin"},
			wantExpandedPath:     []string{"/opt/ansible/bin", "/usr/local/bin"},
		},
		{
			name:                 "Tilde paths expanded",
			ansibleNavigatorPath: []string{"~/bin", "~/ansible/bin"},
			wantExpandedPath:     []string{filepath.Join(home, "bin"), filepath.Join(home, "ansible/bin")},
		},
		{
			name:                 "Mixed absolute and tilde paths",
			ansibleNavigatorPath: []string{"~/bin", "/opt/ansible/bin"},
			wantExpandedPath:     []string{filepath.Join(home, "bin"), "/opt/ansible/bin"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provisioner{}
			config := map[string]interface{}{
				"play":                   []map[string]interface{}{{"target": validPlaybook}},
				"skip_version_check":     true,
				"ansible_navigator_path": tt.ansibleNavigatorPath,
			}

			err := p.Prepare(config)
			require.NoError(t, err)

			assert.Equal(t, tt.wantExpandedPath, p.config.AnsibleNavigatorPath)
		})
	}
}

func TestPreparePathExpansion(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	// Create a playbook in HOME so we can reference it via ~
	homePlaybook, err := os.CreateTemp(home, "packer-ansible-navigator-home-playbook-*.yml")
	require.NoError(t, err)
	_, _ = homePlaybook.WriteString("---\n- hosts: all\n")
	require.NoError(t, homePlaybook.Close())
	defer os.Remove(homePlaybook.Name())

	tildePlaybook := filepath.ToSlash(filepath.Join("~", filepath.Base(homePlaybook.Name())))
	absPlaybook := filepath.Join(home, filepath.Base(homePlaybook.Name()))

	p := &Provisioner{}
	config := map[string]interface{}{
		"skip_version_check": true,
		"play":               []map[string]interface{}{{"target": tildePlaybook}},
		"command":            "~/bin/ansible-navigator",
		"work_dir":           "~/ansible-work",
	}

	require.NoError(t, p.Prepare(config))
	require.Equal(t, filepath.Join(home, "bin/ansible-navigator"), p.config.Command)
	require.Equal(t, filepath.Join(home, "ansible-work"), p.config.WorkDir)
	require.Len(t, p.config.Plays, 1)
	require.Equal(t, absPlaybook, p.config.Plays[0].Target)
}
