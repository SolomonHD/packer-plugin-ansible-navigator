package ansiblenavigator

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDetectShim(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		expectedMgr    string
		expectedIsShim bool
	}{
		{
			name: "asdf shim",
			content: `#!/usr/bin/env bash
# asdf-plugin: ansible-navigator
exec asdf exec ansible-navigator "$@"`,
			expectedMgr:    "asdf",
			expectedIsShim: true,
		},
		{
			name: "rbenv shim",
			content: `#!/usr/bin/env bash
set -e
[ -n "$RBENV_DEBUG" ] && set -x
program="${0##*/}"
export RBENV_ROOT="/home/user/.rbenv"
exec rbenv exec "$program" "$@"`,
			expectedMgr:    "rbenv",
			expectedIsShim: true,
		},
		{
			name: "pyenv shim",
			content: `#!/usr/bin/env bash
set -e
[ -n "$PYENV_DEBUG" ] && set -x
program="${0##*/}"
export PYENV_ROOT="/home/user/.pyenv"
exec pyenv exec "$program" "$@"`,
			expectedMgr:    "pyenv",
			expectedIsShim: true,
		},
		{
			name: "regular binary",
			content: `#!/usr/bin/env bash
# Regular script
echo "Hello World"`,
			expectedMgr:    "",
			expectedIsShim: false,
		},
		{
			name: "real python executable",
			content: `#!/usr/bin/python3
import sys
print("Not a shim")`,
			expectedMgr:    "",
			expectedIsShim: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file with test content
			tmpFile, err := os.CreateTemp("", "shim-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.WriteString(tt.content); err != nil {
				t.Fatalf("Failed to write temp file: %v", err)
			}
			tmpFile.Close()

			// Test detectShim
			manager, isShim := detectShim(tmpFile.Name())

			if isShim != tt.expectedIsShim {
				t.Errorf("detectShim() isShim = %v, want %v", isShim, tt.expectedIsShim)
			}

			if manager != tt.expectedMgr {
				t.Errorf("detectShim() manager = %v, want %v", manager, tt.expectedMgr)
			}
		})
	}
}

func TestDetectShimNonExistentFile(t *testing.T) {
	manager, isShim := detectShim("/nonexistent/file")
	if isShim {
		t.Errorf("detectShim() on nonexistent file should return false, got true")
	}
	if manager != "" {
		t.Errorf("detectShim() on nonexistent file should return empty manager, got %s", manager)
	}
}

func TestDetectShimPerformance(t *testing.T) {
	// Create a test shim file
	tmpFile, err := os.CreateTemp("", "shim-perf-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	shimContent := `#!/usr/bin/env bash
# asdf-plugin: ansible-navigator
exec asdf exec ansible-navigator "$@"`
	if _, err := tmpFile.WriteString(shimContent); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Measure detection time
	start := time.Now()
	detectShim(tmpFile.Name())
	duration := time.Since(start)

	// Should complete in less than 100ms
	if duration > 100*time.Millisecond {
		t.Errorf("detectShim() took %v, should be < 100ms", duration)
	}
}

func TestResolveShim(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		manager     string
		shouldSkip  func() bool
		expectError bool
	}{
		{
			name:    "asdf resolution",
			command: "ansible-navigator",
			manager: "asdf",
			shouldSkip: func() bool {
				_, err := exec.LookPath("asdf")
				return err != nil
			},
			expectError: false,
		},
		{
			name:    "rbenv resolution",
			command: "ansible-navigator",
			manager: "rbenv",
			shouldSkip: func() bool {
				_, err := exec.LookPath("rbenv")
				return err != nil
			},
			expectError: false,
		},
		{
			name:    "pyenv resolution",
			command: "ansible-navigator",
			manager: "pyenv",
			shouldSkip: func() bool {
				_, err := exec.LookPath("pyenv")
				return err != nil
			},
			expectError: false,
		},
		{
			name:        "unsupported manager",
			command:     "ansible-navigator",
			manager:     "unknown",
			shouldSkip:  func() bool { return false },
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldSkip != nil && tt.shouldSkip() {
				t.Skipf("Skipping test: %s not available", tt.manager)
				return
			}

			resolvedPath, err := resolveShim(tt.command, tt.manager)

			if tt.expectError {
				if err == nil {
					t.Errorf("resolveShim() expected error, got nil")
				}
			} else {
				if err != nil {
					// Only fail if the command is actually supposed to be installed
					// For CI/test environments, this might legitimately fail
					t.Logf("resolveShim() returned error (may be expected if command not installed): %v", err)
					return
				}
				if resolvedPath == "" {
					t.Errorf("resolveShim() returned empty path without error")
				}
				if !filepath.IsAbs(resolvedPath) {
					t.Errorf("resolveShim() returned non-absolute path: %s", resolvedPath)
				}
			}
		})
	}
}

func TestResolveShimMocked(t *testing.T) {
	// Create a mock "asdf" command that outputs a test path
	mockDir, err := os.MkdirTemp("", "mock-bin-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(mockDir)

	mockAsdfPath := filepath.Join(mockDir, "asdf")
	mockContent := `#!/bin/bash
if [ "$1" = "which" ] && [ "$2" = "ansible-navigator" ]; then
  echo "/usr/local/bin/ansible-navigator"
  exit 0
fi
exit 1`

	if err := os.WriteFile(mockAsdfPath, []byte(mockContent), 0755); err != nil {
		t.Fatalf("Failed to write mock asdf: %v", err)
	}

	// Prepend mock directory to PATH
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", mockDir+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	// Test resolution
	resolvedPath, err := resolveShim("ansible-navigator", "asdf")
	if err != nil {
		t.Fatalf("resolveShim() unexpected error: %v", err)
	}

	expected := "/usr/local/bin/ansible-navigator"
	if resolvedPath != expected {
		t.Errorf("resolveShim() = %v, want %v", resolvedPath, expected)
	}
}

func TestResolveShimEmptyOutput(t *testing.T) {
	// Create a mock version manager that returns empty output
	mockDir, err := os.MkdirTemp("", "mock-bin-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(mockDir)

	mockAsdfPath := filepath.Join(mockDir, "asdf")
	mockContent := `#!/bin/bash
echo ""
exit 0`

	if err := os.WriteFile(mockAsdfPath, []byte(mockContent), 0755); err != nil {
		t.Fatalf("Failed to write mock asdf: %v", err)
	}

	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", mockDir+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	_, err = resolveShim("ansible-navigator", "asdf")
	if err == nil {
		t.Error("resolveShim() should error on empty output")
	}
	if !strings.Contains(err.Error(), "empty path") {
		t.Errorf("resolveShim() error should mention empty path, got: %v", err)
	}
}
