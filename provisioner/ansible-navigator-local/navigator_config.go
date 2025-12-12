// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansiblenavigatorlocal

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// generateNavigatorConfigYAML converts the NavigatorConfig struct to YAML format
// and applies automatic defaults when execution-environment.enabled = true
func generateNavigatorConfigYAML(config *NavigatorConfig) (string, error) {
	if config == nil {
		return "", fmt.Errorf("navigator_config cannot be nil")
	}

	// Apply automatic EE defaults if execution-environment.enabled = true
	if config.ExecutionEnvironment != nil && config.ExecutionEnvironment.Enabled {
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
	result := make(map[string]interface{})

	if config.Mode != "" {
		result["mode"] = config.Mode
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
		result["execution-environment"] = eeMap
	}

	if config.AnsibleConfig != nil {
		ansibleMap := make(map[string]interface{})
		if config.AnsibleConfig.Config != "" {
			ansibleMap["config"] = config.AnsibleConfig.Config
		}
		configMap := make(map[string]interface{})
		if config.AnsibleConfig.Defaults != nil {
			defaultsMap := make(map[string]interface{})
			if config.AnsibleConfig.Defaults.RemoteTmp != "" {
				defaultsMap["remote_tmp"] = config.AnsibleConfig.Defaults.RemoteTmp
			}
			defaultsMap["host_key_checking"] = config.AnsibleConfig.Defaults.HostKeyChecking
			if len(defaultsMap) > 0 {
				configMap["defaults"] = defaultsMap
			}
		}
		if config.AnsibleConfig.SSHConnection != nil {
			sshMap := make(map[string]interface{})
			if config.AnsibleConfig.SSHConnection.SSHTimeout > 0 {
				sshMap["ssh_timeout"] = config.AnsibleConfig.SSHConnection.SSHTimeout
			}
			sshMap["pipelining"] = config.AnsibleConfig.SSHConnection.Pipelining
			if len(sshMap) > 0 {
				configMap["ssh_connection"] = sshMap
			}
		}
		if len(configMap) > 0 {
			ansibleMap["config"] = configMap
		}
		if len(ansibleMap) > 0 {
			result["ansible"] = ansibleMap
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
			result["logging"] = loggingMap
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
			result["playbook-artifact"] = artifactMap
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
			result["collection-doc-cache"] = cacheMap
		}
	}

	// Wrap everything under the ansible-navigator root key
	return map[string]interface{}{
		"ansible-navigator": result,
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
