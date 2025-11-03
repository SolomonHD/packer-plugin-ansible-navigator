// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansible

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
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

func TestResolveCollectionsCacheDir(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{
			name:      "empty path uses default",
			input:     "",
			wantError: false,
		},
		{
			name:      "absolute path",
			input:     "/tmp/collections",
			wantError: false,
		},
		{
			name:      "relative path",
			input:     "./collections",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolveCollectionsCacheDir(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("resolveCollectionsCacheDir() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && result == "" {
				t.Errorf("resolveCollectionsCacheDir() returned empty path")
			}
			if !tt.wantError && !filepath.IsAbs(result) {
				t.Errorf("resolveCollectionsCacheDir() = %v, want absolute path", result)
			}
		})
	}
}

func TestIsCollectionInstalled(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")

	// Create a mock collection structure
	collectionDir := filepath.Join(cacheDir, "ansible_collections", "community", "general")
	if err := os.MkdirAll(collectionDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create MANIFEST.json
	manifestPath := filepath.Join(collectionDir, "MANIFEST.json")
	if err := os.WriteFile(manifestPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create MANIFEST.json: %v", err)
	}

	tests := []struct {
		name       string
		collection string
		cacheDir   string
		want       bool
	}{
		{
			name:       "installed collection",
			collection: "community.general",
			cacheDir:   cacheDir,
			want:       true,
		},
		{
			name:       "installed collection with version",
			collection: "community.general:5.11.0",
			cacheDir:   cacheDir,
			want:       true,
		},
		{
			name:       "not installed collection",
			collection: "ansible.posix",
			cacheDir:   cacheDir,
			want:       false,
		},
		{
			name:       "invalid collection name",
			collection: "invalid",
			cacheDir:   cacheDir,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCollectionInstalled(tt.collection, tt.cacheDir)
			if got != tt.want {
				t.Errorf("isCollectionInstalled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnsureCollections_NoCollections(t *testing.T) {
	ui := newMockUi()
	config := &Config{}

	err := ensureCollections(ui, config)
	if err != nil {
		t.Errorf("ensureCollections() with no collections should not error, got: %v", err)
	}
}

func TestEnsureCollections_OfflineMissing(t *testing.T) {
	ui := newMockUi()
	tmpDir := t.TempDir()

	config := &Config{
		Collections:         []string{"myorg.missing"},
		CollectionsOffline:  true,
		CollectionsCacheDir: tmpDir,
	}

	err := ensureCollections(ui, config)
	if err == nil {
		t.Error("ensureCollections() should error for missing collection in offline mode")
	}
}

func TestEnsureCollections_RequirementsFileNotFound(t *testing.T) {
	ui := newMockUi()
	tmpDir := t.TempDir()

	config := &Config{
		CollectionsRequirements: "/nonexistent/requirements.yml",
		CollectionsCacheDir:     tmpDir,
	}

	err := ensureCollections(ui, config)
	if err == nil {
		t.Error("ensureCollections() should error for non-existent requirements file")
	}
}

func TestSetCollectionsPath(t *testing.T) {
	// Save original environment
	origPath := os.Getenv("ANSIBLE_COLLECTIONS_PATHS")
	defer func() {
		if origPath != "" {
			os.Setenv("ANSIBLE_COLLECTIONS_PATHS", origPath)
		} else {
			os.Unsetenv("ANSIBLE_COLLECTIONS_PATHS")
		}
	}()

	tests := []struct {
		name         string
		cacheDir     string
		existingPath string
		wantContains string
	}{
		{
			name:         "set new path",
			cacheDir:     "/tmp/collections",
			existingPath: "",
			wantContains: "/tmp/collections",
		},
		{
			name:         "append to existing path",
			cacheDir:     "/tmp/collections",
			existingPath: "/existing/path",
			wantContains: "/tmp/collections:/existing/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.existingPath != "" {
				os.Setenv("ANSIBLE_COLLECTIONS_PATHS", tt.existingPath)
			} else {
				os.Unsetenv("ANSIBLE_COLLECTIONS_PATHS")
			}

			p := &Provisioner{
				config: Config{
					CollectionsCacheDir: tt.cacheDir,
				},
			}

			err := p.setCollectionsPath()
			if err != nil {
				t.Errorf("setCollectionsPath() error = %v", err)
				return
			}

			got := os.Getenv("ANSIBLE_COLLECTIONS_PATHS")
			if got != tt.wantContains {
				t.Errorf("ANSIBLE_COLLECTIONS_PATHS = %v, want %v", got, tt.wantContains)
			}
		})
	}
}

func TestSetCollectionsPath_EmptyCacheDir(t *testing.T) {
	p := &Provisioner{
		config: Config{
			CollectionsCacheDir: "",
		},
	}

	err := p.setCollectionsPath()
	if err != nil {
		t.Errorf("setCollectionsPath() with empty cache dir should not error, got: %v", err)
	}
}
