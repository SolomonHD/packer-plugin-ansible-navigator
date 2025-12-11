// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansiblenavigator

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/stretchr/testify/require"
)

// mockUi implements a simple UI for testing
type mockUi struct {
	sayMessages     []string
	messageMessages []string
	errorMessages   []string
}

func (m *mockUi) Say(message string) {
	m.sayMessages = append(m.sayMessages, message)
}

func (m *mockUi) Message(message string) {
	m.messageMessages = append(m.messageMessages, message)
}

func (m *mockUi) Error(message string) {
	m.errorMessages = append(m.errorMessages, message)
}

func (m *mockUi) Machine(t string, args ...string) {}

func (m *mockUi) Ask(string) (string, error) {
	return "", nil
}

func (m *mockUi) Askf(query string, args ...interface{}) (string, error) {
	return "", nil
}

func (m *mockUi) Sayf(message string, args ...interface{}) {
	m.sayMessages = append(m.sayMessages, fmt.Sprintf(message, args...))
}

func (m *mockUi) Messagef(message string, args ...interface{}) {
	m.messageMessages = append(m.messageMessages, fmt.Sprintf(message, args...))
}

func (m *mockUi) Errorf(message string, args ...interface{}) {
	m.errorMessages = append(m.errorMessages, fmt.Sprintf(message, args...))
}

func (m *mockUi) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) (body io.ReadCloser) {
	return stream
}

func newMockUi() packersdk.Ui {
	return &mockUi{
		sayMessages:     make([]string, 0),
		messageMessages: make([]string, 0),
		errorMessages:   make([]string, 0),
	}
}

func TestGalaxyManager_InstallRequirements_NoRequirementsFile(t *testing.T) {
	ui := newMockUi()
	config := &Config{}
	gm := NewGalaxyManager(config, ui)

	if err := gm.InstallRequirements(); err != nil {
		t.Fatalf("InstallRequirements() unexpected error: %v", err)
	}
}

func TestGalaxyManager_InstallRequirements_FromRequirementsFile_CallsGalaxy(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "galaxy_calls.txt")
	stubPath := filepath.Join(tmpDir, "ansible-galaxy-stub.sh")
	requirementsFile := filepath.Join(tmpDir, "requirements.yml")

	stub := `#!/usr/bin/env bash
set -euo pipefail

echo "$@" >> "${OUTPUT_FILE}"
exit 0
`
	require.NoError(t, os.WriteFile(stubPath, []byte(stub), 0o755))

	// Ensure both roles and collections sections exist so both installs are attempted.
	require.NoError(t, os.WriteFile(requirementsFile, []byte("roles:\n  - src: test.role\ncollections:\n  - name: community.general\n"), 0o644))

	os.Setenv("OUTPUT_FILE", outputFile)
	t.Cleanup(func() { _ = os.Unsetenv("OUTPUT_FILE") })

	ui := newMockUi()
	cfg := &Config{
		RequirementsFile: requirementsFile,
	}
	// Override the default ansible-galaxy command with our stub
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", tmpDir+":"+oldPath)
	t.Cleanup(func() { os.Setenv("PATH", oldPath) })
	// Rename stub to ansible-galaxy so it's found on PATH
	os.Rename(stubPath, filepath.Join(tmpDir, "ansible-galaxy"))
	gm := NewGalaxyManager(cfg, ui)

	require.NoError(t, gm.InstallRequirements())

	data, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	got := string(data)
	require.Contains(t, got, "install -r")
	require.Contains(t, got, "collection install -r")
}

func TestGalaxyManager_SetupEnvironmentPaths(t *testing.T) {
	// Save original environment
	origCollPath := os.Getenv("ANSIBLE_COLLECTIONS_PATHS")
	origRolePath := os.Getenv("ANSIBLE_ROLES_PATH")
	defer func() {
		if origCollPath != "" {
			os.Setenv("ANSIBLE_COLLECTIONS_PATHS", origCollPath)
		} else {
			os.Unsetenv("ANSIBLE_COLLECTIONS_PATHS")
		}
		if origRolePath != "" {
			os.Setenv("ANSIBLE_ROLES_PATH", origRolePath)
		} else {
			os.Unsetenv("ANSIBLE_ROLES_PATH")
		}
	}()

	tests := []struct {
		name               string
		collectionCacheDir string
		rolesCacheDir      string
		existingCollPath   string
		existingRolePath   string
		wantCollPath       string
		wantRolePath       string
	}{
		{
			name:               "set new paths",
			collectionCacheDir: "/tmp/collections",
			rolesCacheDir:      "/tmp/roles",
			existingCollPath:   "",
			existingRolePath:   "",
			wantCollPath:       "/tmp/collections",
			wantRolePath:       "/tmp/roles",
		},
		{
			name:               "append to existing paths",
			collectionCacheDir: "/tmp/collections",
			rolesCacheDir:      "/tmp/roles",
			existingCollPath:   "/existing/collections",
			existingRolePath:   "/existing/roles",
			wantCollPath:       "/tmp/collections:/existing/collections",
			wantRolePath:       "/tmp/roles:/existing/roles",
		},
		{
			name:               "empty cache dirs",
			collectionCacheDir: "",
			rolesCacheDir:      "",
			existingCollPath:   "/existing/collections",
			existingRolePath:   "/existing/roles",
			wantCollPath:       "/existing/collections",
			wantRolePath:       "/existing/roles",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.existingCollPath != "" {
				os.Setenv("ANSIBLE_COLLECTIONS_PATHS", tt.existingCollPath)
			} else {
				os.Unsetenv("ANSIBLE_COLLECTIONS_PATHS")
			}
			if tt.existingRolePath != "" {
				os.Setenv("ANSIBLE_ROLES_PATH", tt.existingRolePath)
			} else {
				os.Unsetenv("ANSIBLE_ROLES_PATH")
			}

			ui := newMockUi()
			config := &Config{
				CollectionsCacheDir: tt.collectionCacheDir,
				RolesCacheDir:       tt.rolesCacheDir,
			}
			gm := NewGalaxyManager(config, ui)

			err := gm.SetupEnvironmentPaths()
			if err != nil {
				t.Errorf("SetupEnvironmentPaths() error = %v", err)
				return
			}

			gotCollPath := os.Getenv("ANSIBLE_COLLECTIONS_PATHS")
			if gotCollPath != tt.wantCollPath {
				t.Errorf("ANSIBLE_COLLECTIONS_PATHS = %v, want %v", gotCollPath, tt.wantCollPath)
			}

			gotRolePath := os.Getenv("ANSIBLE_ROLES_PATH")
			if gotRolePath != tt.wantRolePath {
				t.Errorf("ANSIBLE_ROLES_PATH = %v, want %v", gotRolePath, tt.wantRolePath)
			}
		})
	}
}
