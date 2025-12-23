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
			PrivilegeEscalation: &AnsibleConfigPrivilegeEscalation{
				Become:       true,
				BecomeMethod: "sudo",
				BecomeUser:   "root",
			},
			PersistentConnection: &AnsibleConfigPersistentConnection{ConnectTimeout: 30},
			Inventory:            &AnsibleConfigInventory{EnablePlugins: []string{"ini", "yaml"}},
			Colors:               &AnsibleConfigColors{ForceColor: true},
		},
	}

	yamlStr, err := generateNavigatorConfigYAML(config, "")
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	if !strings.Contains(yamlStr, "path: /tmp/ansible.cfg") {
		t.Fatalf("expected ansible.config.path in YAML, got: %s", yamlStr)
	}

	for _, forbidden := range []string{
		"defaults",
		"ssh_connection",
		"privilege_escalation",
		"persistent_connection",
		"inventory",
		"paramiko_connection",
		"colors",
		"diff",
		"galaxy",
	} {
		if strings.Contains(yamlStr, forbidden) {
			t.Fatalf("did not expect %s under ansible.config in YAML, got: %s", forbidden, yamlStr)
		}
	}
}

func TestGenerateAnsibleCfgContent_NewSections(t *testing.T) {
	content, err := generateAnsibleCfgContent(&AnsibleConfig{
		PrivilegeEscalation: &AnsibleConfigPrivilegeEscalation{
			Become:       true,
			BecomeMethod: "sudo",
			BecomeUser:   "root",
		},
		PersistentConnection: &AnsibleConfigPersistentConnection{
			ConnectTimeout:      30,
			ConnectRetryTimeout: 15,
			CommandTimeout:      60,
		},
		Inventory: &AnsibleConfigInventory{EnablePlugins: []string{"ini", "yaml"}},
		ParamikoConnection: &AnsibleConfigParamikoConnection{
			ProxyCommand: "ssh -W %h:%p jumphost",
		},
		Colors: &AnsibleConfigColors{ForceColor: true},
		Diff:   &AnsibleConfigDiff{Always: true, Context: 3},
		Galaxy: &AnsibleConfigGalaxy{ServerList: []string{"automation_hub"}, IgnoreCerts: true},
	})
	if err != nil {
		t.Fatalf("generateAnsibleCfgContent failed: %v", err)
	}
	for _, expected := range []string{
		"[privilege_escalation]",
		"become = True",
		"become_method = sudo",
		"become_user = root",
		"[persistent_connection]",
		"connect_timeout = 30",
		"connect_retry_timeout = 15",
		"command_timeout = 60",
		"[inventory]",
		"enable_plugins = ini,yaml",
		"[paramiko_connection]",
		"proxy_command = ssh -W %h:%p jumphost",
		"[colors]",
		"force_color = True",
		"[diff]",
		"always = True",
		"context = 3",
		"[galaxy]",
		"server_list = automation_hub",
		"ignore_certs = True",
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("expected %q in generated ansible.cfg, got: %q", expected, content)
		}
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

func TestGenerateNavigatorConfigYAML_ExecutionEnvironment_ContainerEngine(t *testing.T) {
	config := &NavigatorConfig{
		ExecutionEnvironment: &ExecutionEnvironment{
			Enabled:         true,
			ContainerEngine: "podman",
		},
	}

	yamlStr, err := generateNavigatorConfigYAML(config, "")
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	var parsed map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &parsed); err != nil {
		t.Fatalf("Generated YAML is not valid: %v\nYAML:\n%s", err, yamlStr)
	}

	root, ok := parsed["ansible-navigator"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected ansible-navigator root key in parsed YAML, got: %v", parsed)
	}
	ee, ok := root["execution-environment"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected execution-environment map in parsed YAML, got: %v", root)
	}
	if ee["container-engine"] != "podman" {
		t.Fatalf("Expected container-engine=podman, got: %v", ee["container-engine"])
	}
}

func TestGenerateNavigatorConfigYAML_ExecutionEnvironment_ContainerOptions(t *testing.T) {
	config := &NavigatorConfig{
		ExecutionEnvironment: &ExecutionEnvironment{
			Enabled:          true,
			ContainerOptions: []string{"--net=host", "--security-opt=label=disable"},
		},
	}

	yamlStr, err := generateNavigatorConfigYAML(config, "")
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	var parsed map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &parsed); err != nil {
		t.Fatalf("Generated YAML is not valid: %v\nYAML:\n%s", err, yamlStr)
	}

	root := parsed["ansible-navigator"].(map[string]interface{})
	ee := root["execution-environment"].(map[string]interface{})

	opts, ok := ee["container-options"].([]interface{})
	if !ok {
		t.Fatalf("Expected container-options list, got: %T (%v)", ee["container-options"], ee["container-options"])
	}
	if len(opts) != 2 {
		t.Fatalf("Expected 2 container-options, got %d: %v", len(opts), opts)
	}
	if opts[0] != "--net=host" || opts[1] != "--security-opt=label=disable" {
		t.Fatalf("Unexpected container-options ordering/content: %v", opts)
	}
}

func TestGenerateNavigatorConfigYAML_ExecutionEnvironment_PullArgumentsOnly_CreatesPullObject(t *testing.T) {
	config := &NavigatorConfig{
		ExecutionEnvironment: &ExecutionEnvironment{
			Enabled:              true,
			PullArguments:        []string{"--tls-verify=false"},
			PullPolicy:           "", // explicit for clarity
			Image:                "quay.io/ansible/creator-ee:latest",
			EnvironmentVariables: &EnvironmentVariablesConfig{Set: map[string]string{"CUSTOM": "x"}},
		},
	}

	yamlStr, err := generateNavigatorConfigYAML(config, "")
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	var parsed map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &parsed); err != nil {
		t.Fatalf("Generated YAML is not valid: %v\nYAML:\n%s", err, yamlStr)
	}

	root := parsed["ansible-navigator"].(map[string]interface{})
	ee := root["execution-environment"].(map[string]interface{})
	pull, ok := ee["pull"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected pull map to be present when pull_arguments set, got: %T (%v)", ee["pull"], ee["pull"])
	}
	if _, ok := pull["policy"]; ok {
		t.Fatalf("Did not expect pull.policy when PullPolicy is empty, got: %v", pull["policy"])
	}
	args, ok := pull["arguments"].([]interface{})
	if !ok || len(args) != 1 || args[0] != "--tls-verify=false" {
		t.Fatalf("Unexpected pull.arguments: %T (%v)", pull["arguments"], pull["arguments"])
	}
}

func TestGenerateNavigatorConfigYAML_ExecutionEnvironment_PullPolicyAndArguments(t *testing.T) {
	config := &NavigatorConfig{
		ExecutionEnvironment: &ExecutionEnvironment{
			Enabled:       true,
			PullPolicy:    "missing",
			PullArguments: []string{"--tls-verify=false"},
		},
	}

	yamlStr, err := generateNavigatorConfigYAML(config, "")
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	var parsed map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &parsed); err != nil {
		t.Fatalf("Generated YAML is not valid: %v\nYAML:\n%s", err, yamlStr)
	}

	root := parsed["ansible-navigator"].(map[string]interface{})
	ee := root["execution-environment"].(map[string]interface{})
	pull := ee["pull"].(map[string]interface{})
	if pull["policy"] != "missing" {
		t.Fatalf("Expected pull.policy=missing, got: %v", pull["policy"])
	}
	args := pull["arguments"].([]interface{})
	if len(args) != 1 || args[0] != "--tls-verify=false" {
		t.Fatalf("Unexpected pull.arguments: %v", args)
	}
}

func TestGenerateNavigatorConfigYAML_ExecutionEnvironment_AllNewFieldsConfigured(t *testing.T) {
	config := &NavigatorConfig{
		ExecutionEnvironment: &ExecutionEnvironment{
			Enabled:          true,
			ContainerEngine:  "podman",
			ContainerOptions: []string{"--net=host"},
			PullPolicy:       "missing",
			PullArguments:    []string{"--tls-verify=false"},
		},
	}

	yamlStr, err := generateNavigatorConfigYAML(config, "")
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	var parsed map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &parsed); err != nil {
		t.Fatalf("Generated YAML is not valid: %v\nYAML:\n%s", err, yamlStr)
	}

	root := parsed["ansible-navigator"].(map[string]interface{})
	ee := root["execution-environment"].(map[string]interface{})
	if ee["container-engine"] != "podman" {
		t.Fatalf("Expected container-engine=podman, got: %v", ee["container-engine"])
	}
	opts := ee["container-options"].([]interface{})
	if len(opts) != 1 || opts[0] != "--net=host" {
		t.Fatalf("Unexpected container-options: %v", opts)
	}
	pull := ee["pull"].(map[string]interface{})
	if pull["policy"] != "missing" {
		t.Fatalf("Expected pull.policy=missing, got: %v", pull["policy"])
	}
	args := pull["arguments"].([]interface{})
	if len(args) != 1 || args[0] != "--tls-verify=false" {
		t.Fatalf("Unexpected pull.arguments: %v", args)
	}
}

func TestGenerateNavigatorConfigYAML_ExecutionEnvironment_DoesNotDuplicateCollectionsMount_WhenUserAlreadyHasDest(t *testing.T) {
	config := &NavigatorConfig{
		ExecutionEnvironment: &ExecutionEnvironment{
			Enabled: true,
			Image:   "quay.io/ansible/creator-ee:latest",
			VolumeMounts: []VolumeMount{
				{Src: "/already/mounted", Dest: "/tmp/.packer_ansible/collections", Options: "ro"},
			},
		},
	}

	collectionsPath := filepath.Join(os.TempDir(), "test_collections")
	yamlStr, err := generateNavigatorConfigYAML(config, collectionsPath)
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	var parsed map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &parsed); err != nil {
		t.Fatalf("Generated YAML is not valid: %v\nYAML:\n%s", err, yamlStr)
	}
	root := parsed["ansible-navigator"].(map[string]interface{})
	ee := root["execution-environment"].(map[string]interface{})
	vm, ok := ee["volume-mounts"].([]interface{})
	if !ok {
		t.Fatalf("Expected volume-mounts list, got: %T (%v)", ee["volume-mounts"], ee["volume-mounts"])
	}

	destCount := 0
	for _, item := range vm {
		m := item.(map[string]interface{})
		if m["dest"] == "/tmp/.packer_ansible/collections" {
			destCount++
		}
	}
	if destCount != 1 {
		t.Fatalf("Expected exactly 1 mount with dest /tmp/.packer_ansible/collections, got %d (volume-mounts=%v)", destCount, vm)
	}
}

func TestGenerateNavigatorConfigYAML_LoggingAndPlaybookArtifact_KeyNamesAndTypes(t *testing.T) {
	config := &NavigatorConfig{
		Mode: "stdout",
		Logging: &LoggingConfig{
			Level:  "debug",
			File:   "/tmp/ansible-navigator.log",
			Append: true,
		},
		PlaybookArtifact: &PlaybookArtifact{
			Enable: true,
			Replay: "/tmp/replay.json",
			SaveAs: "/tmp/save.json",
		},
	}

	yamlStr, err := generateNavigatorConfigYAML(config, "")
	if err != nil {
		t.Fatalf("generateNavigatorConfigYAML failed: %v", err)
	}

	var parsed map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &parsed); err != nil {
		t.Fatalf("Generated YAML is not valid: %v\nYAML:\n%s", err, yamlStr)
	}

	root, ok := parsed["ansible-navigator"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected ansible-navigator root key in parsed YAML, got: %T (%v)", parsed["ansible-navigator"], parsed["ansible-navigator"])
	}

	logging, ok := root["logging"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected logging map in YAML, got: %T (%v)", root["logging"], root["logging"])
	}
	if logging["level"] != "debug" {
		t.Fatalf("Expected logging.level=debug, got: %v", logging["level"])
	}
	if logging["file"] != "/tmp/ansible-navigator.log" {
		t.Fatalf("Expected logging.file=/tmp/ansible-navigator.log, got: %v", logging["file"])
	}
	if logging["append"] != true {
		t.Fatalf("Expected logging.append=true, got: %v", logging["append"])
	}

	artifact, ok := root["playbook-artifact"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected playbook-artifact map in YAML, got: %T (%v)", root["playbook-artifact"], root["playbook-artifact"])
	}
	if artifact["enable"] != true {
		t.Fatalf("Expected playbook-artifact.enable=true, got: %v", artifact["enable"])
	}
	if artifact["replay"] != "/tmp/replay.json" {
		t.Fatalf("Expected playbook-artifact.replay=/tmp/replay.json, got: %v", artifact["replay"])
	}
	if artifact["save-as"] != "/tmp/save.json" {
		t.Fatalf("Expected playbook-artifact.save-as=/tmp/save.json, got: %v", artifact["save-as"])
	}
	if _, hasUnderscore := artifact["save_as"]; hasUnderscore {
		t.Fatalf("Did not expect playbook-artifact.save_as in YAML (must be save-as), got: %v", artifact)
	}
}
