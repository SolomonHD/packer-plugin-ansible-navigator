// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansiblenavigator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlayStructure(t *testing.T) {
	play := Play{
		Name:   "Test Play",
		Target: "geerlingguy.docker",
		ExtraVars: map[string]string{
			"docker_install_compose": "true",
		},
		Tags:      []string{"setup", "config"},
		VarsFiles: []string{"vars/main.yml"},
		Become:    true,
	}

	assert.Equal(t, "Test Play", play.Name)
	assert.Equal(t, "geerlingguy.docker", play.Target)
	assert.True(t, play.Become)
	assert.Len(t, play.ExtraVars, 1)
	assert.Len(t, play.Tags, 2)
	assert.Len(t, play.VarsFiles, 1)
}

func TestGenerateRolePlaybook(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		play     Play
		expected string
	}{
		{
			name: "Simple role without extras",
			role: "geerlingguy.docker",
			play: Play{
				Target: "geerlingguy.docker",
			},
			expected: `---
- hosts: all
  roles:
    - role: geerlingguy.docker
`,
		},
		{
			name: "Role with become",
			role: "geerlingguy.docker",
			play: Play{
				Target: "geerlingguy.docker",
				Become: true,
			},
			expected: `---
- hosts: all
  become: yes
  roles:
    - role: geerlingguy.docker
`,
		},
		{
			name: "Role with extra vars",
			role: "geerlingguy.docker",
			play: Play{
				Target: "geerlingguy.docker",
				ExtraVars: map[string]string{
					"docker_install_compose": "true",
					"docker_edition":         "ce",
				},
			},
			expected: "---\n- hosts: all\n  roles:\n    - role: geerlingguy.docker\n      vars:\n",
		},
		{
			name: "Role with vars files",
			role: "geerlingguy.docker",
			play: Play{
				Target:    "geerlingguy.docker",
				VarsFiles: []string{"vars/docker.yml", "vars/main.yml"},
			},
			expected: `---
- hosts: all
  vars_files:
    - vars/docker.yml
    - vars/main.yml
  roles:
    - role: geerlingguy.docker
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := createRolePlaybook(tt.role, tt.play)
			assert.NoError(t, err)
			assert.NotEmpty(t, tmpFile)

			// Clean up
			defer os.Remove(tmpFile)

			// Read the generated file
			content, err := os.ReadFile(tmpFile)
			assert.NoError(t, err)

			// Check that content contains expected elements
			contentStr := string(content)
			assert.Contains(t, contentStr, "hosts: all")
			assert.Contains(t, contentStr, "role: "+tt.role)

			if tt.play.Become {
				assert.Contains(t, contentStr, "become: yes")
			}
		})
	}
}

func TestConfigValidationWithPlays(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid play configuration",
			config: Config{
				Plays: []Play{
					{
						Name:   "Setup",
						Target: "geerlingguy.docker",
					},
				},
			},
			expectError: false,
		},
		{
			name: "Play without target",
			config: Config{
				Plays: []Play{
					{
						Name: "Setup",
					},
				},
			},
			expectError: true,
			errorMsg:    "target must be specified",
		},
		{
			name: "Both playbook_file and plays specified",
			config: Config{
				PlaybookFile: "site.yml",
				Plays: []Play{
					{
						Target: "geerlingguy.docker",
					},
				},
			},
			expectError: true,
			errorMsg:    "you may specify only one of `playbook_file` or `play` blocks",
		},
		{
			name:        "Neither playbook_file nor plays specified",
			config:      Config{},
			expectError: true,
			errorMsg:    "either `playbook_file` or `play` blocks must be defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provisioner{
				config: tt.config,
			}
			p.config.Command = "ansible-navigator run"
			p.config.GalaxyCommand = "ansible-galaxy"
			p.config.HostAlias = "default"
			p.config.User = "testuser"
			p.config.SkipVersionCheck = true

			err := p.Prepare(&tt.config)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPlaybookVsRoleDetection(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		isPlaybook bool
	}{
		{
			name:       "YML playbook",
			target:     "site.yml",
			isPlaybook: true,
		},
		{
			name:       "YAML playbook",
			target:     "deploy.yaml",
			isPlaybook: true,
		},
		{
			name:       "Role FQDN",
			target:     "geerlingguy.docker",
			isPlaybook: false,
		},
		{
			name:       "Role with namespace",
			target:     "myorg.webserver.deploy",
			isPlaybook: false,
		},
		{
			name:       "Simple role name",
			target:     "docker",
			isPlaybook: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isPlaybook := filepath.Ext(tt.target) == ".yml" || filepath.Ext(tt.target) == ".yaml"
			assert.Equal(t, tt.isPlaybook, isPlaybook)
		})
	}
}

func TestDefaultCacheDirectories(t *testing.T) {
	p := &Provisioner{}
	p.config.Command = "ansible-navigator run"
	p.config.GalaxyCommand = "ansible-galaxy"
	p.config.HostAlias = "default"
	p.config.User = "testuser"
	p.config.SkipVersionCheck = true
	p.config.Plays = []Play{
		{
			Target: "geerlingguy.docker",
		},
	}

	err := p.Prepare(&p.config)
	assert.NoError(t, err)

	// Check that default cache directories were set
	assert.Contains(t, p.config.CollectionsCacheDir, ".packer.d/ansible_collections_cache")
	assert.Contains(t, p.config.RolesCacheDir, ".packer.d/ansible_roles_cache")
}

func TestMultiplePlayConfiguration(t *testing.T) {
	config := Config{
		Plays: []Play{
			{
				Name:   "Setup base system",
				Target: "geerlingguy.docker",
				ExtraVars: map[string]string{
					"docker_install_compose": "true",
				},
			},
			{
				Name:   "Deploy web stack",
				Target: "deploy.yml",
				Become: true,
				Tags:   []string{"deploy", "web"},
			},
			{
				Name:      "Custom role",
				Target:    "myorg.custom",
				VarsFiles: []string{"vars/custom.yml"},
			},
		},
	}

	assert.Len(t, config.Plays, 3)
	assert.Equal(t, "Setup base system", config.Plays[0].Name)
	assert.Equal(t, "geerlingguy.docker", config.Plays[0].Target)
	assert.True(t, config.Plays[1].Become)
	assert.Len(t, config.Plays[1].Tags, 2)
}

func TestUnifiedRequirementsFieldPresent(t *testing.T) {
	config := Config{
		RequirementsFile: "./requirements.yml",
		RolesCacheDir:    "~/.packer.d/ansible_roles_cache",
		OfflineMode:      false,
		ForceUpdate:      true,
		Plays: []Play{
			{
				Target: "geerlingguy.docker",
			},
		},
	}

	assert.Equal(t, "./requirements.yml", config.RequirementsFile)
	assert.False(t, config.OfflineMode)
	assert.True(t, config.ForceUpdate)
}
