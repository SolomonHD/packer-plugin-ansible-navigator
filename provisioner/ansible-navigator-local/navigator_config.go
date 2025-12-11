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

// generateNavigatorConfigYAML converts the NavigatorConfig map to YAML format
// and applies automatic defaults when execution-environment.enabled = true
func generateNavigatorConfigYAML(config map[string]interface{}) (string, error) {
	if config == nil || len(config) == 0 {
		return "", fmt.Errorf("navigator_config cannot be empty")
	}

	// Apply automatic EE defaults if execution-environment.enabled = true
	if eeConfig, ok := config["execution-environment"].(map[string]interface{}); ok {
		if enabled, ok := eeConfig["enabled"].(bool); ok && enabled {
			// Set ansible config defaults if not already present
			if _, hasAnsible := config["ansible"]; !hasAnsible {
				config["ansible"] = make(map[string]interface{})
			}

			ansibleConfig := config["ansible"].(map[string]interface{})
			if _, hasConfig := ansibleConfig["config"]; !hasConfig {
				ansibleConfig["config"] = make(map[string]interface{})
			}

			ansibleCfg := ansibleConfig["config"].(map[string]interface{})
			if _, hasDefaults := ansibleCfg["defaults"]; !hasDefaults {
				ansibleCfg["defaults"] = make(map[string]interface{})
			}

			defaults := ansibleCfg["defaults"].(map[string]interface{})

			// Set temp directory defaults to prevent permission errors
			if _, hasRemoteTmp := defaults["remote_tmp"]; !hasRemoteTmp {
				defaults["remote_tmp"] = "/tmp/.ansible/tmp"
			}
			if _, hasLocalTmp := defaults["local_tmp"]; !hasLocalTmp {
				defaults["local_tmp"] = "/tmp/.ansible-local"
			}

			// Set environment variables for EE container
			if _, hasEnvVars := eeConfig["environment-variables"]; !hasEnvVars {
				eeConfig["environment-variables"] = make(map[string]interface{})
			}

			envVars := eeConfig["environment-variables"].(map[string]interface{})
			if _, hasRemoteTmp := envVars["ANSIBLE_REMOTE_TMP"]; !hasRemoteTmp {
				envVars["ANSIBLE_REMOTE_TMP"] = "/tmp/.ansible/tmp"
			}
			if _, hasLocalTmp := envVars["ANSIBLE_LOCAL_TMP"]; !hasLocalTmp {
				envVars["ANSIBLE_LOCAL_TMP"] = "/tmp/.ansible-local"
			}
		}
	}

	// Marshal to YAML
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal navigator_config to YAML: %w", err)
	}

	return string(yamlData), nil
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
