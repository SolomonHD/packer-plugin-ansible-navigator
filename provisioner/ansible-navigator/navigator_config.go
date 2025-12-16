// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansiblenavigator

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// applyAutomaticEEDefaults applies safe defaults for execution environments.
//
// This is primarily to avoid permission errors inside containers when Ansible
// tries to write under non-writable default paths like "/.ansible/tmp".
func applyAutomaticEEDefaults(config *NavigatorConfig) {
	if config == nil {
		return
	}

	if config.ExecutionEnvironment == nil || !config.ExecutionEnvironment.Enabled {
		return
	}

	// Initialize AnsibleConfig if not present
	if config.AnsibleConfig == nil {
		config.AnsibleConfig = &AnsibleConfig{}
	}
	if config.AnsibleConfig.Defaults == nil {
		config.AnsibleConfig.Defaults = &AnsibleConfigDefaults{}
	}

	// Set temp directory defaults to prevent permission errors
	if config.AnsibleConfig.Defaults.RemoteTmp == "" {
		config.AnsibleConfig.Defaults.RemoteTmp = "/tmp/.ansible/tmp"
	}
	if config.AnsibleConfig.Defaults.LocalTmp == "" {
		config.AnsibleConfig.Defaults.LocalTmp = "/tmp/.ansible-local"
	}

	// Set environment variables for EE container
	if config.ExecutionEnvironment.EnvironmentVariables == nil {
		config.ExecutionEnvironment.EnvironmentVariables = &EnvironmentVariablesConfig{
			Set: make(map[string]string),
		}
	}
	if config.ExecutionEnvironment.EnvironmentVariables.Set == nil {
		config.ExecutionEnvironment.EnvironmentVariables.Set = make(map[string]string)
	}

	if _, hasRemoteTmp := config.ExecutionEnvironment.EnvironmentVariables.Set["ANSIBLE_REMOTE_TMP"]; !hasRemoteTmp {
		config.ExecutionEnvironment.EnvironmentVariables.Set["ANSIBLE_REMOTE_TMP"] = "/tmp/.ansible/tmp"
	}
	if _, hasLocalTmp := config.ExecutionEnvironment.EnvironmentVariables.Set["ANSIBLE_LOCAL_TMP"]; !hasLocalTmp {
		config.ExecutionEnvironment.EnvironmentVariables.Set["ANSIBLE_LOCAL_TMP"] = "/tmp/.ansible-local"
	}
}

// generateNavigatorConfigYAML converts the NavigatorConfig struct to YAML format
// and applies automatic defaults when execution-environment.enabled = true
func generateNavigatorConfigYAML(config *NavigatorConfig) (string, error) {
	if config == nil {
		return "", fmt.Errorf("navigator_config cannot be nil")
	}

	applyAutomaticEEDefaults(config)

	// Convert to YAML-friendly structure with proper field names
	yamlConfig := convertToYAMLStructure(config)

	// Marshal to YAML
	yamlData, err := yaml.Marshal(yamlConfig)
	if err != nil {
		return "", fmt.Errorf("failed to marshal navigator_config to YAML: %w", err)
	}

	return string(yamlData), nil
}

// convertToYAMLStructure converts NavigatorConfig to a map with YAML-friendly field names
// (using hyphens instead of underscores for ansible-navigator.yml compatibility)
func convertToYAMLStructure(config *NavigatorConfig) map[string]interface{} {
	// Create nested structure with ansible-navigator root key
	ansibleNavigator := make(map[string]interface{})

	if config.Mode != "" {
		ansibleNavigator["mode"] = config.Mode
	}

	if config.ExecutionEnvironment != nil {
		eeMap := make(map[string]interface{})
		eeMap["enabled"] = config.ExecutionEnvironment.Enabled
		if config.ExecutionEnvironment.Image != "" {
			eeMap["image"] = config.ExecutionEnvironment.Image
		}
		if config.ExecutionEnvironment.PullPolicy != "" {
			eeMap["pull"] = map[string]interface{}{
				"policy": config.ExecutionEnvironment.PullPolicy,
			}
		}
		if config.ExecutionEnvironment.EnvironmentVariables != nil {
			envVarsMap := make(map[string]interface{})
			if len(config.ExecutionEnvironment.EnvironmentVariables.Pass) > 0 {
				envVarsMap["pass"] = config.ExecutionEnvironment.EnvironmentVariables.Pass
			}
			if len(config.ExecutionEnvironment.EnvironmentVariables.Set) > 0 {
				envVarsMap["set"] = config.ExecutionEnvironment.EnvironmentVariables.Set
			}
			if len(envVarsMap) > 0 {
				eeMap["environment-variables"] = envVarsMap
			}
		}
		ansibleNavigator["execution-environment"] = eeMap
	}

	if config.AnsibleConfig != nil {
		ansibleMap := make(map[string]interface{})
		// Schema compliance: ansible.config may contain only help/path/cmdline.
		// We represent defaults/ssh_connection via a generated ansible.cfg file and
		// reference it here.
		if config.AnsibleConfig.Config != "" {
			ansibleMap["config"] = map[string]interface{}{
				"path": config.AnsibleConfig.Config,
			}
		}
		if len(ansibleMap) > 0 {
			ansibleNavigator["ansible"] = ansibleMap
		}
	}

	if config.Logging != nil {
		loggingMap := make(map[string]interface{})
		if config.Logging.Level != "" {
			loggingMap["level"] = config.Logging.Level
		}
		if config.Logging.File != "" {
			loggingMap["file"] = config.Logging.File
		}
		loggingMap["append"] = config.Logging.Append
		if len(loggingMap) > 0 {
			ansibleNavigator["logging"] = loggingMap
		}
	}

	if config.PlaybookArtifact != nil {
		artifactMap := make(map[string]interface{})
		artifactMap["enable"] = config.PlaybookArtifact.Enable
		if config.PlaybookArtifact.Replay != "" {
			artifactMap["replay"] = config.PlaybookArtifact.Replay
		}
		if config.PlaybookArtifact.SaveAs != "" {
			artifactMap["save-as"] = config.PlaybookArtifact.SaveAs
		}
		if len(artifactMap) > 0 {
			ansibleNavigator["playbook-artifact"] = artifactMap
		}
	}

	if config.CollectionDocCache != nil {
		cacheMap := make(map[string]interface{})
		if config.CollectionDocCache.Path != "" {
			cacheMap["path"] = config.CollectionDocCache.Path
		}
		if config.CollectionDocCache.Timeout > 0 {
			cacheMap["timeout"] = config.CollectionDocCache.Timeout
		}
		if len(cacheMap) > 0 {
			ansibleNavigator["collection-doc-cache"] = cacheMap
		}
	}

	// Wrap everything under the ansible-navigator root key
	return map[string]interface{}{
		"ansible-navigator": ansibleNavigator,
	}
}

// createNavigatorConfigFile creates a temporary ansible-navigator.yml file
// Returns the absolute path to the file
func createNavigatorConfigFile(content string) (string, error) {
	tmpFile, err := os.CreateTemp("", "packer-ansible-navigator-*.yml")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary ansible-navigator.yml file: %w", err)
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write ansible-navigator.yml content: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to close ansible-navigator.yml temp file: %w", err)
	}

	return tmpFile.Name(), nil
}
