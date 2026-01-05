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

	"gopkg.in/yaml.v3"
)

// Test successful YAML generation from navigator_config
func TestGenerateNavigatorConfigYAML_Basic(t *testing.T) {
	config := &NavigatorConfig{
		Mode: "stdout",
		AnsibleConfig: &AnsibleConfig{
			Defaults: &AnsibleConfigDefaults{
				HostKeyChecking: false,
			},
		},
	}

	yaml, err := generateNavigatorConfigYAML(config, "")
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	if !strings.Contains(yaml, "mode: stdout") {
		t.Errorf("Expected mode: stdout in YAML, got: %s", yaml)
	}

	// Schema compliance: ansible defaults live in a generated ansible.cfg and are
	// NOT embedded in ansible-navigator.yml.
	if strings.Contains(yaml, "host_key_checking") {
		t.Errorf("Did not expect host_key_checking in YAML, got: %s", yaml)
	}
}

// Test automatic EE defaults when execution-environment.enabled = true
func TestGenerateNavigatorConfigYAML_AutomaticEEDefaults(t *testing.T) {
	config := &NavigatorConfig{
		ExecutionEnvironment: &ExecutionEnvironment{
			Enabled: true,
			Image:   "quay.io/ansible/creator-ee:latest",
		},
	}

	yaml, err := generateNavigatorConfigYAML(config, "")
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	// Note: remote_tmp/local_tmp are configured via ansible.cfg (referenced by path)
	// and MUST NOT appear under ansible.config in this YAML.
	if strings.Contains(yaml, "remote_tmp") {
		t.Errorf("Did not expect remote_tmp in YAML, got: %s", yaml)
	}
	if strings.Contains(yaml, "local_tmp") {
		t.Errorf("Did not expect local_tmp in YAML, got: %s", yaml)
	}

	if !strings.Contains(yaml, "ANSIBLE_REMOTE_TMP: /tmp/.ansible/tmp") {
		t.Errorf("Expected automatic ANSIBLE_REMOTE_TMP env var in YAML, got: %s", yaml)
	}

	if !strings.Contains(yaml, "ANSIBLE_LOCAL_TMP: /tmp/.ansible-local") {
		t.Errorf("Expected automatic ANSIBLE_LOCAL_TMP env var in YAML, got: %s", yaml)
	}
}

func TestGenerateNavigatorConfigYAML_AutomaticEEHomeXDGDefaults_WhenNotSetOrPassed(t *testing.T) {
	config := &NavigatorConfig{
		ExecutionEnvironment: &ExecutionEnvironment{
			Enabled: true,
			Image:   "quay.io/ansible/creator-ee:latest",
		},
	}

	yamlStr, err := generateNavigatorConfigYAML(config, "")
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	// Defaults only when user did not set or pass-through.
	for _, expected := range []string{
		"HOME: /tmp",
		"XDG_CACHE_HOME: /tmp/.cache",
		"XDG_CONFIG_HOME: /tmp/.config",
	} {
		if !strings.Contains(yamlStr, expected) {
			t.Fatalf("expected %q in YAML, got: %s", expected, yamlStr)
		}
	}
}

func TestGenerateNavigatorConfigYAML_DoesNotSetHomeXDGDefaults_WhenPassedThrough(t *testing.T) {
	config := &NavigatorConfig{
		ExecutionEnvironment: &ExecutionEnvironment{
			Enabled: true,
			Image:   "quay.io/ansible/creator-ee:latest",
			EnvironmentVariables: &EnvironmentVariablesConfig{
				Pass: []string{"HOME", "XDG_CACHE_HOME", "XDG_CONFIG_HOME"},
			},
		},
	}

	yamlStr, err := generateNavigatorConfigYAML(config, "")
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	// Ensure we did not inject defaults for values the user intends to pass-through.
	for _, forbidden := range []string{
		"HOME: /tmp",
		"XDG_CACHE_HOME: /tmp/.cache",
		"XDG_CONFIG_HOME: /tmp/.config",
	} {
		if strings.Contains(yamlStr, forbidden) {
			t.Fatalf("did not expect %q in YAML when passed-through, got: %s", forbidden, yamlStr)
		}
	}
}

func TestGenerateNavigatorConfigYAML_DoesNotOverrideHomeXDG_WhenUserSetsValues(t *testing.T) {
	config := &NavigatorConfig{
		ExecutionEnvironment: &ExecutionEnvironment{
			Enabled: true,
			Image:   "quay.io/ansible/creator-ee:latest",
			EnvironmentVariables: &EnvironmentVariablesConfig{
				Set: map[string]string{
					"HOME":            "/custom/home",
					"XDG_CACHE_HOME":  "/custom/cache",
					"XDG_CONFIG_HOME": "/custom/config",
				},
			},
		},
	}

	yamlStr, err := generateNavigatorConfigYAML(config, "")
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	for _, expected := range []string{
		"HOME: /custom/home",
		"XDG_CACHE_HOME: /custom/cache",
		"XDG_CONFIG_HOME: /custom/config",
	} {
		if !strings.Contains(yamlStr, expected) {
			t.Fatalf("expected %q in YAML, got: %s", expected, yamlStr)
		}
	}
}

// Test that user-provided values are not overridden
func TestGenerateNavigatorConfigYAML_UserValuesPreserved(t *testing.T) {
	config := &NavigatorConfig{
		ExecutionEnvironment: &ExecutionEnvironment{
			Enabled: true,
			Image:   "quay.io/ansible/creator-ee:latest",
			EnvironmentVariables: &EnvironmentVariablesConfig{
				Set: map[string]string{
					"ANSIBLE_REMOTE_TMP": "/custom/path",
				},
			},
		},
		AnsibleConfig: &AnsibleConfig{
			Defaults: &AnsibleConfigDefaults{
				RemoteTmp: "/another/custom/path",
			},
		},
	}

	yaml, err := generateNavigatorConfigYAML(config, "")
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	// Verify user values are preserved
	if !strings.Contains(yaml, "/custom/path") {
		t.Errorf("Expected user-provided ANSIBLE_REMOTE_TMP to be preserved, got: %s", yaml)
	}

	// Verify defaults are not added when user values exist
	if strings.Count(yaml, "ANSIBLE_REMOTE_TMP") > 1 {
		t.Errorf("ANSIBLE_REMOTE_TMP should only appear once (user value), got: %s", yaml)
	}
}

func TestGenerateNavigatorConfigYAML_AnsibleConfigPathSchemaCompliant(t *testing.T) {
	config := &NavigatorConfig{
		Mode: "stdout",
		AnsibleConfig: &AnsibleConfig{
			Config: "/tmp/ansible.cfg",
			Defaults: &AnsibleConfigDefaults{
				RemoteTmp: "/tmp/.ansible/tmp",
				LocalTmp:  "/tmp/.ansible-local",
			},
			SSHConnection: &AnsibleConfigConnection{Pipelining: true},
		},
	}

	yamlStr, err := generateNavigatorConfigYAML(config, "")
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	if !strings.Contains(yamlStr, "path: /tmp/ansible.cfg") {
		t.Fatalf("expected ansible.config.path in YAML, got: %s", yamlStr)
	}

	if strings.Contains(yamlStr, "defaults") {
		t.Fatalf("did not expect defaults under ansible.config in YAML, got: %s", yamlStr)
	}
	if strings.Contains(yamlStr, "ssh_connection") {
		t.Fatalf("did not expect ssh_connection under ansible.config in YAML, got: %s", yamlStr)
	}
}

func TestGenerateAnsibleCfgContent_LocalTmpIncludedWhenSet(t *testing.T) {
	content, err := generateAnsibleCfgContent(&AnsibleConfig{
		Defaults: &AnsibleConfigDefaults{
			RemoteTmp:       "/tmp/.ansible/tmp",
			LocalTmp:        "/tmp/.ansible-local",
			HostKeyChecking: false,
		},
	})
	if err != nil {
		t.Fatalf("generateAnsibleCfgContent failed: %v", err)
	}
	if !strings.Contains(content, "local_tmp = /tmp/.ansible-local") {
		t.Fatalf("expected local_tmp in generated ansible.cfg, got: %q", content)
	}
}

func TestGenerateAnsibleCfgContent_LocalTmpOmittedWhenUnset(t *testing.T) {
	content, err := generateAnsibleCfgContent(&AnsibleConfig{
		Defaults: &AnsibleConfigDefaults{
			RemoteTmp:       "/tmp/.ansible/tmp",
			HostKeyChecking: false,
		},
	})
	if err != nil {
		t.Fatalf("generateAnsibleCfgContent failed: %v", err)
	}
	if strings.Contains(content, "local_tmp") {
		t.Fatalf("did not expect local_tmp in generated ansible.cfg when unset, got: %q", content)
	}
}

// Test that empty config is allowed (validation happens in Packer Config.Validate())
func TestGeneratorNavigatorConfigYAML_EmptyConfig(t *testing.T) {
	config := &NavigatorConfig{}

	// Empty config is technically allowed by generateNavigatorConfigYAML
	// The validation that it must have at least one field happens in Config.Validate()
	yaml, err := generateNavigatorConfigYAML(config, "")
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML should not error on empty config: %v", err)
	}

	// Should produce minimal YAML (likely just empty map or minimal structure)
	if yaml == "" {
		t.Errorf("Expected some YAML output even for empty config")
	}
}

// Test error when config is nil
func TestGenerateNavigatorConfigYAML_NilConfig(t *testing.T) {
	_, err := generateNavigatorConfigYAML(nil, "")
	if err == nil {
		t.Fatal("Expected error for nil config, got nil")
	}

	if !strings.Contains(err.Error(), "cannot be nil") {
		t.Errorf("Expected 'cannot be nil' error message, got: %v", err)
	}
}

// Test that YAML can be parsed back
func TestGenerateNavigatorConfigYAML_ValidYAML(t *testing.T) {
	config := &NavigatorConfig{
		Mode: "json",
		ExecutionEnvironment: &ExecutionEnvironment{
			Enabled: true,
			Image:   "quay.io/ansible/creator-ee:latest",
		},
	}

	yamlStr, err := generateNavigatorConfigYAML(config, "")
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	// Try to parse the generated YAML
	var parsed map[string]interface{}
	err = yaml.Unmarshal([]byte(yamlStr), &parsed)
	if err != nil {
		t.Fatalf("Generated YAML is not valid: %v\nYAML:\n%s", err, yamlStr)
	}

	// Verify key fields exist in parsed output under ansible-navigator root key
	ansibleNavigator, ok := parsed["ansible-navigator"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected ansible-navigator root key in parsed YAML, got: %v", parsed)
	}
	if ansibleNavigator["mode"] != "json" {
		t.Errorf("Expected mode=json in parsed YAML under ansible-navigator key, got: %v", ansibleNavigator["mode"])
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
	config := &NavigatorConfig{
		ExecutionEnvironment: &ExecutionEnvironment{
			Enabled: false,
			Image:   "quay.io/ansible/creator-ee:latest",
		},
	}

	yaml, err := generateNavigatorConfigYAML(config, "")
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
	config := &NavigatorConfig{
		Mode: "json",
		AnsibleConfig: &AnsibleConfig{
			Defaults: &AnsibleConfigDefaults{
				HostKeyChecking: false,
				RemoteTmp:       "/tmp/.ansible/tmp",
			},
			SSHConnection: &AnsibleConfigConnection{
				Pipelining: true,
			},
		},
		ExecutionEnvironment: &ExecutionEnvironment{
			Enabled:    true,
			Image:      "quay.io/ansible/creator-ee:latest",
			PullPolicy: "missing",
			EnvironmentVariables: &EnvironmentVariablesConfig{
				Set: map[string]string{
					"CUSTOM_VAR": "custom_value",
				},
			},
		},
		Logging: &LoggingConfig{
			Level:  "debug",
			Append: true,
		},
	}

	yamlStr, err := generateNavigatorConfigYAML(config, "")
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	// Verify all YAML sections are present (note: ansible defaults / ssh_connection
	// are represented via ansible.cfg and are NOT embedded in YAML).
	expectedStrings := []string{
		"mode: json",
		"pull:",
		"policy: missing",
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

// Test volume mount added when EE is enabled and collections path provided
func TestGenerateNavigatorConfigYAML_VolumeMountWithCollections(t *testing.T) {
	config := &NavigatorConfig{
		ExecutionEnvironment: &ExecutionEnvironment{
			Enabled: true,
			Image:   "quay.io/ansible/creator-ee:latest",
		},
	}

	// Use a real home directory path
	testCollectionsPath := filepath.Join(os.TempDir(), "test_collections")

	yamlStr, err := generateNavigatorConfigYAML(config, testCollectionsPath)
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	// Verify volume mount is present
	if !strings.Contains(yamlStr, "volume-mounts:") {
		t.Errorf("Expected volume-mounts in YAML when collections path provided with EE, got: %s", yamlStr)
	}

	if !strings.Contains(yamlStr, "src:") {
		t.Errorf("Expected src in volume mount, got: %s", yamlStr)
	}

	if !strings.Contains(yamlStr, "dest: /tmp/.packer_ansible/collections") {
		t.Errorf("Expected dest: /tmp/.packer_ansible/collections in volume mount, got: %s", yamlStr)
	}

	if !strings.Contains(yamlStr, "options: ro") {
		t.Errorf("Expected options: ro in volume mount, got: %s", yamlStr)
	}

	// Verify ANSIBLE_COLLECTIONS_PATH environment variable
	if !strings.Contains(yamlStr, "ANSIBLE_COLLECTIONS_PATH: /tmp/.packer_ansible/collections") {
		t.Errorf("Expected ANSIBLE_COLLECTIONS_PATH env var in YAML, got: %s", yamlStr)
	}
}

// Test no volume mount when EE is disabled
func TestGenerateNavigatorConfigYAML_NoVolumeMountWhenEEDisabled(t *testing.T) {
	config := &NavigatorConfig{
		ExecutionEnvironment: &ExecutionEnvironment{
			Enabled: false,
		},
	}

	testCollectionsPath := filepath.Join(os.TempDir(), "test_collections")

	yamlStr, err := generateNavigatorConfigYAML(config, testCollectionsPath)
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	// Verify no volume mount is added when EE is disabled
	if strings.Contains(yamlStr, "volume-mounts") {
		t.Errorf("Did not expect volume-mounts when EE is disabled, got: %s", yamlStr)
	}

	// Verify ANSIBLE_COLLECTIONS_PATH is not added when EE is disabled
	if strings.Contains(yamlStr, "ANSIBLE_COLLECTIONS_PATH") {
		t.Errorf("Did not expect ANSIBLE_COLLECTIONS_PATH when EE is disabled, got: %s", yamlStr)
	}
}

// Test no volume mount when collections path is empty
func TestGenerateNavigatorConfigYAML_NoVolumeMountWhenNoCollections(t *testing.T) {
	config := &NavigatorConfig{
		ExecutionEnvironment: &ExecutionEnvironment{
			Enabled: true,
			Image:   "quay.io/ansible/creator-ee:latest",
		},
	}

	yamlStr, err := generateNavigatorConfigYAML(config, "")
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	// Verify no volume mount is added when collections path is empty
	if strings.Contains(yamlStr, "volume-mounts") {
		t.Errorf("Did not expect volume-mounts when collections path is empty, got: %s", yamlStr)
	}

	// Verify ANSIBLE_COLLECTIONS_PATH is not added when collections path is empty
	if strings.Contains(yamlStr, "ANSIBLE_COLLECTIONS_PATH") {
		t.Errorf("Did not expect ANSIBLE_COLLECTIONS_PATH when collections path is empty, got: %s", yamlStr)
	}
}

// Test that user-provided volume mounts are preserved
func TestGenerateNavigatorConfigYAML_UserVolumeMountsPreserved(t *testing.T) {
	config := &NavigatorConfig{
		ExecutionEnvironment: &ExecutionEnvironment{
			Enabled: true,
			Image:   "quay.io/ansible/creator-ee:latest",
			VolumeMounts: []VolumeMount{
				{
					Src:     "/host/custom",
					Dest:    "/container/custom",
					Options: "rw",
				},
			},
		},
	}

	testCollectionsPath := filepath.Join(os.TempDir(), "test_collections")

	yamlStr, err := generateNavigatorConfigYAML(config, testCollectionsPath)
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	// Verify user mount is preserved
	if !strings.Contains(yamlStr, "/host/custom") {
		t.Errorf("Expected user-provided mount /host/custom to be preserved, got: %s", yamlStr)
	}

	// Verify collections mount is added
	if !strings.Contains(yamlStr, "/tmp/.packer_ansible/collections") {
		t.Errorf("Expected automatic collections mount to be added, got: %s", yamlStr)
	}
}

// Test buildNavigatorCLIFlags
func TestBuildNavigatorCLIFlags_Basic(t *testing.T) {
	config := &NavigatorConfig{
		Mode: "stdout",
		ExecutionEnvironment: &ExecutionEnvironment{
			Enabled:    true,
			Image:      "quay.io/ansible/creator-ee:latest",
			PullPolicy: "missing",
		},
		Logging: &LoggingConfig{
			Level: "debug",
		},
	}

	flags := buildNavigatorCLIFlags(config)

	expectedFlags := []string{
		"--mode=stdout",
		"--execution-environment-image=quay.io/ansible/creator-ee:latest",
		"--pull-policy=missing",
		"--log-level=debug",
	}

	for _, expected := range expectedFlags {
		found := false
		for _, flag := range flags {
			if flag == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected flag %s not found in %v", expected, flags)
		}
	}
}

func TestBuildNavigatorCLIFlags_Complex(t *testing.T) {
	config := &NavigatorConfig{
		ExecutionEnvironment: &ExecutionEnvironment{
			Enabled: true,
			EnvironmentVariables: &EnvironmentVariablesConfig{
				Set:  map[string]string{"KEY": "VALUE"},
				Pass: []string{"HOME"},
			},
			VolumeMounts: []VolumeMount{
				{Src: "/src", Dest: "/dest", Options: "ro"},
			},
			ContainerOptions: []string{"--net=host"},
		},
	}

	flags := buildNavigatorCLIFlags(config)

	// Check for repeatable flags
	foundEEV := false
	foundEVM := false
	foundContainerOptions := false

	for _, flag := range flags {
		if strings.HasPrefix(flag, "--eev=") {
			if strings.Contains(flag, "KEY=VALUE") || strings.Contains(flag, "HOME") {
				foundEEV = true
			}
		}
		if strings.HasPrefix(flag, "--evm=") && strings.Contains(flag, "/src:/dest:ro") {
			foundEVM = true
		}
		if strings.HasPrefix(flag, "--container-options=") && strings.Contains(flag, "--net=host") {
			foundContainerOptions = true
		}
	}

	if !foundEEV {
		t.Error("Expected --eev flags not found")
	}
	if !foundEVM {
		t.Error("Expected --evm flag not found")
	}
	if !foundContainerOptions {
		t.Error("Expected --container-options flag not found")
	}
}

// Test hasUnmappedSettings
func TestHasUnmappedSettings(t *testing.T) {
	// Case 1: No unmapped settings
	config1 := &NavigatorConfig{
		Mode: "stdout",
	}
	if hasUnmappedSettings(config1) {
		t.Error("Expected hasUnmappedSettings to be false for basic config")
	}

	// Case 2: PlaybookArtifact enabled
	config2 := &NavigatorConfig{
		PlaybookArtifact: &PlaybookArtifact{
			Enable: true,
		},
	}
	if !hasUnmappedSettings(config2) {
		t.Error("Expected hasUnmappedSettings to be true when PlaybookArtifact is enabled")
	}

	// Case 3: CollectionDocCache path set
	config3 := &NavigatorConfig{
		CollectionDocCache: &CollectionDocCache{
			Path: "/tmp/cache",
		},
	}
	if !hasUnmappedSettings(config3) {
		t.Error("Expected hasUnmappedSettings to be true when CollectionDocCache path is set")
	}
}

// Test generateMinimalYAML
func TestGenerateMinimalYAML(t *testing.T) {
	config := &NavigatorConfig{
		Mode: "stdout", // Mapped
		PlaybookArtifact: &PlaybookArtifact{ // Unmapped
			Enable: true,
			SaveAs: "/tmp/artifact.json",
		},
	}

	yamlStr, err := generateMinimalYAML(config)
	if err != nil {
		t.Fatalf("generateMinimalYAML failed: %v", err)
	}

	if strings.Contains(yamlStr, "mode: stdout") {
		t.Error("Minimal YAML should NOT contain mapped settings like mode")
	}

	if !strings.Contains(yamlStr, "playbook-artifact") {
		t.Error("Minimal YAML SHOULD contain unmapped settings like playbook-artifact")
	}

	if !strings.Contains(yamlStr, "save-as: /tmp/artifact.json") {
		t.Error("Minimal YAML should contain save-as value")
	}
}
