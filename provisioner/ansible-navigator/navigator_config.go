// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansiblenavigator

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// generateAnsibleCfg generates INI-formatted ansible.cfg content
func generateAnsibleCfg() string {
	var buf strings.Builder
	buf.WriteString("[defaults]\n")
	buf.WriteString("remote_tmp = /tmp/.ansible/tmp\n")
	buf.WriteString("host_key_checking = False\n")
	buf.WriteString("\n")
	buf.WriteString("[ssh_connection]\n")
	buf.WriteString("ssh_timeout = 30\n")
	buf.WriteString("pipelining = True\n")
	return buf.String()
}

// createAnsibleCfgFile creates a temporary ansible.cfg file and returns its path
func createAnsibleCfgFile(content string) (string, error) {
	tmpFile, err := os.CreateTemp("", "packer-ansible-cfg-*.cfg")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary ansible.cfg file: %w", err)
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write ansible.cfg content: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to close ansible.cfg temp file: %w", err)
	}

	return tmpFile.Name(), nil
}

// generateNavigatorConfigYAML converts the NavigatorConfig struct to YAML format
// and generates ansible.cfg when execution-environment is enabled.
// Returns (yamlContent, ansibleCfgPath, error)
func generateNavigatorConfigYAML(config *NavigatorConfig) (string, string, error) {
	if config == nil {
		return "", "", fmt.Errorf("navigator_config cannot be nil")
	}

	var ansibleCfgPath string

	// Generate ansible.cfg if execution-environment is enabled
	if config.ExecutionEnvironment != nil && config.ExecutionEnvironment.Enabled {
		// Generate ansible.cfg content
		ansibleCfgContent := generateAnsibleCfg()

		// Write to temp file
		path, err := createAnsibleCfgFile(ansibleCfgContent)
		if err != nil {
			return "", "", fmt.Errorf("failed to create ansible.cfg: %w", err)
		}
		ansibleCfgPath = path

		// Set the path in config for YAML generation
		if config.AnsibleConfig == nil {
			config.AnsibleConfig = &AnsibleConfig{}
		}
		config.AnsibleConfig.Path = ansibleCfgPath

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
		return "", ansibleCfgPath, fmt.Errorf("failed to marshal navigator_config to YAML: %w", err)
	}

	return string(yamlData), ansibleCfgPath, nil
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
		configMap := make(map[string]interface{})

		// Use Config field for user-provided ansible.cfg path (deprecated, use Path instead)
		if config.AnsibleConfig.Config != "" {
			configMap["path"] = config.AnsibleConfig.Config
		}
		// Path field is the canonical way to specify ansible.cfg location
		if config.AnsibleConfig.Path != "" {
			configMap["path"] = config.AnsibleConfig.Path
		}
		if config.AnsibleConfig.Help {
			configMap["help"] = config.AnsibleConfig.Help
		}
		if config.AnsibleConfig.Cmdline != "" {
			configMap["cmdline"] = config.AnsibleConfig.Cmdline
		}

		if len(configMap) > 0 {
			ansibleMap["config"] = configMap
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
