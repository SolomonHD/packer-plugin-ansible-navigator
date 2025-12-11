// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansiblenavigator

import (
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// Test successful YAML generation from navigator_config
func TestGenerateNavigatorConfigYAML_Basic(t *testing.T) {
	config := map[string]interface{}{
		"mode": "stdout",
		"ansible": map[string]interface{}{
			"config": map[string]interface{}{
				"defaults": map[string]interface{}{
					"host_key_checking": "False",
				},
			},
		},
	}

	yaml, err := generateNavigatorConfigYAML(config)
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	if !strings.Contains(yaml, "mode: stdout") {
		t.Errorf("Expected mode: stdout in YAML, got: %s", yaml)
	}

	if !strings.Contains(yaml, "host_key_checking") {
		t.Errorf("Expected host_key_checking in YAML, got: %s", yaml)
	}
}

// Test automatic EE defaults when execution-environment.enabled = true
func TestGenerateNavigatorConfigYAML_AutomaticEEDefaults(t *testing.T) {
	config := map[string]interface{}{
		"execution-environment": map[string]interface{}{
			"enabled": true,
			"image":   "quay.io/ansible/creator-ee:latest",
		},
	}

	yaml, err := generateNavigatorConfigYAML(config)
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	// Verify automatic defaults were applied
	if !strings.Contains(yaml, "remote_tmp: /tmp/.ansible/tmp") {
		t.Errorf("Expected automatic remote_tmp default in YAML, got: %s", yaml)
	}

	if !strings.Contains(yaml, "local_tmp: /tmp/.ansible-local") {
		t.Errorf("Expected automatic local_tmp default in YAML, got: %s", yaml)
	}

	if !strings.Contains(yaml, "ANSIBLE_REMOTE_TMP: /tmp/.ansible/tmp") {
		t.Errorf("Expected automatic ANSIBLE_REMOTE_TMP env var in YAML, got: %s", yaml)
	}

	if !strings.Contains(yaml, "ANSIBLE_LOCAL_TMP: /tmp/.ansible-local") {
		t.Errorf("Expected automatic ANSIBLE_LOCAL_TMP env var in YAML, got: %s", yaml)
	}
}

// Test that user-provided values are not overridden
func TestGenerateNavigatorConfigYAML_UserValuesPreserved(t *testing.T) {
	config := map[string]interface{}{
		"execution-environment": map[string]interface{}{
			"enabled": true,
			"image":   "quay.io/ansible/creator-ee:latest",
			"environment-variables": map[string]interface{}{
				"ANSIBLE_REMOTE_TMP": "/custom/path",
			},
		},
		"ansible": map[string]interface{}{
			"config": map[string]interface{}{
				"defaults": map[string]interface{}{
					"remote_tmp": "/another/custom/path",
				},
			},
		},
	}

	yaml, err := generateNavigatorConfigYAML(config)
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	// Verify user values are preserved
	if !strings.Contains(yaml, "/custom/path") {
		t.Errorf("Expected user-provided ANSIBLE_REMOTE_TMP to be preserved, got: %s", yaml)
	}

	if !strings.Contains(yaml, "/another/custom/path") {
		t.Errorf("Expected user-provided remote_tmp to be preserved, got: %s", yaml)
	}

	// Verify defaults are not added when user values exist
	if strings.Count(yaml, "ANSIBLE_REMOTE_TMP") > 1 {
		t.Errorf("ANSIBLE_REMOTE_TMP should only appear once (user value), got: %s", yaml)
	}
}

// Test error when config is empty
func TestGenerateNavigatorConfigYAML_EmptyConfig(t *testing.T) {
	config := map[string]interface{}{}

	_, err := generateNavigatorConfigYAML(config)
	if err == nil {
		t.Fatal("Expected error for empty config, got nil")
	}

	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("Expected 'cannot be empty' error message, got: %v", err)
	}
}

// Test error when config is nil
func TestGenerateNavigatorConfigYAML_NilConfig(t *testing.T) {
	_, err := generateNavigatorConfigYAML(nil)
	if err == nil {
		t.Fatal("Expected error for nil config, got nil")
	}

	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("Expected 'cannot be empty' error message, got: %v", err)
	}
}

// Test that YAML can be parsed back
func TestGenerateNavigatorConfigYAML_ValidYAML(t *testing.T) {
	config := map[string]interface{}{
		"mode": "json",
		"execution-environment": map[string]interface{}{
			"enabled": true,
			"image":   "quay.io/ansible/creator-ee:latest",
		},
	}

	yamlStr, err := generateNavigatorConfigYAML(config)
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	// Try to parse the generated YAML
	var parsed map[string]interface{}
	err = yaml.Unmarshal([]byte(yamlStr), &parsed)
	if err != nil {
		t.Fatalf("Generated YAML is not valid: %v\nYAML:\n%s", err, yamlStr)
	}

	// Verify key fields exist in parsed output
	if parsed["mode"] != "json" {
		t.Errorf("Expected mode=json in parsed YAML, got: %v", parsed["mode"])
	}
}

// Test creating navigator config file
func TestCreateNavigatorConfigFile(t *testing.T) {
	content := `---
mode: stdout
execution-environment:
  enabled: true
  image: quay.io/ansible/creator-ee:latest
`

	path, err := createNavigatorConfigFile(content)
	if err != nil {
		t.Fatalf("createNavigatorConfigFile failed: %v", err)
	}

	// Ensure cleanup
	defer os.Remove(path)

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Config file was not created at: %s", path)
	}

	// Verify file content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	if string(data) != content {
		t.Errorf("File content mismatch.\nExpected:\n%s\nGot:\n%s", content, string(data))
	}
}

// Test createNavigatorConfigFile cleanup on error
func TestCreateNavigatorConfigFile_Cleanup(t *testing.T) {
	content := `---
mode: stdout
`

	path, err := createNavigatorConfigFile(content)
	if err != nil {
		t.Fatalf("createNavigatorConfigFile failed: %v", err)
	}

	// File should exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Config file was not created at: %s", path)
	}

	// Cleanup
	err = os.Remove(path)
	if err != nil {
		t.Fatalf("Failed to cleanup file: %v", err)
	}

	// File should no longer exist
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("Config file still exists after cleanup: %s", path)
	}
}

// Test EE enabled=false doesn't add defaults
func TestGenerateNavigatorConfigYAML_EEDisabled(t *testing.T) {
	config := map[string]interface{}{
		"execution-environment": map[string]interface{}{
			"enabled": false,
			"image":   "quay.io/ansible/creator-ee:latest",
		},
	}

	yaml, err := generateNavigatorConfigYAML(config)
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	// Verify automatic defaults were NOT applied when enabled=false
	if strings.Contains(yaml, "remote_tmp") {
		t.Errorf("Should not add automatic remote_tmp when EE is disabled, got: %s", yaml)
	}

	if strings.Contains(yaml, "ANSIBLE_REMOTE_TMP") {
		t.Errorf("Should not add automatic ANSIBLE_REMOTE_TMP when EE is disabled, got: %s", yaml)
	}
}

// Test complex nested configuration
func TestGenerateNavigatorConfigYAML_ComplexConfig(t *testing.T) {
	config := map[string]interface{}{
		"mode": "json",
		"ansible": map[string]interface{}{
			"config": map[string]interface{}{
				"defaults": map[string]interface{}{
					"host_key_checking": "False",
					"gathering":         "smart",
				},
				"ssh_connection": map[string]interface{}{
					"pipelining": "True",
				},
			},
			"playbook": map[string]interface{}{
				"timeout": 30,
			},
		},
		"execution-environment": map[string]interface{}{
			"enabled":     true,
			"image":       "quay.io/ansible/creator-ee:latest",
			"pull-policy": "missing",
			"environment-variables": map[string]interface{}{
				"CUSTOM_VAR": "custom_value",
			},
		},
		"logging": map[string]interface{}{
			"level":  "debug",
			"append": true,
		},
	}

	yamlStr, err := generateNavigatorConfigYAML(config)
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	// Verify all sections are present
	expectedStrings := []string{
		"mode: json",
		"host_key_checking",
		"gathering",
		"pipelining",
		"timeout",
		"pull-policy",
		"CUSTOM_VAR",
		"level: debug",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(yamlStr, expected) {
			t.Errorf("Expected '%s' in YAML, got: %s", expected, yamlStr)
		}
	}

	// Verify it's valid YAML
	var parsed map[string]interface{}
	err = yaml.Unmarshal([]byte(yamlStr), &parsed)
	if err != nil {
		t.Fatalf("Generated YAML is not valid: %v", err)
	}
}
