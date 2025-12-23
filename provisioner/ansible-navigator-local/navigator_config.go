// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansiblenavigatorlocal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func envVarIsSetOrPassed(env *EnvironmentVariablesConfig, key string) bool {
	if env == nil {
		return false
	}
	if env.Set != nil {
		if _, ok := env.Set[key]; ok {
			return true
		}
	}
	for _, p := range env.Pass {
		if strings.EqualFold(p, key) {
			return true
		}
	}
	return false
}

// applyAutomaticEEDefaults applies safe defaults for execution environments.
//
// This is primarily to avoid permission errors inside containers when Ansible
// tries to write under non-writable default paths like "/.ansible/tmp".
//
// It also configures collections path mounting and environment variables when
// collectionsPath is provided.
func applyAutomaticEEDefaults(config *NavigatorConfig, collectionsPath string) {
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

	// Safe EE defaults (per missing key): only set when user did not explicitly set or pass-through.
	env := config.ExecutionEnvironment.EnvironmentVariables
	if !envVarIsSetOrPassed(env, "ANSIBLE_REMOTE_TMP") {
		env.Set["ANSIBLE_REMOTE_TMP"] = "/tmp/.ansible/tmp"
	}
	if !envVarIsSetOrPassed(env, "ANSIBLE_LOCAL_TMP") {
		env.Set["ANSIBLE_LOCAL_TMP"] = "/tmp/.ansible-local"
	}
	if !envVarIsSetOrPassed(env, "HOME") {
		env.Set["HOME"] = "/tmp"
	}
	if !envVarIsSetOrPassed(env, "XDG_CACHE_HOME") {
		env.Set["XDG_CACHE_HOME"] = "/tmp/.cache"
	}
	if !envVarIsSetOrPassed(env, "XDG_CONFIG_HOME") {
		env.Set["XDG_CONFIG_HOME"] = "/tmp/.config"
	}

	// Configure collections path for execution environment
	if collectionsPath != "" {
		// For local provisioner, the collections path is on the target machine
		// So we don't need to expand ~ here (it will be on the target)
		// But we ensure absolute path format
		absCollectionsPath := collectionsPath
		if !filepath.IsAbs(absCollectionsPath) && !strings.HasPrefix(absCollectionsPath, "/") {
			// For on-target execution, assume paths are already target-relative
			// or will be resolved on the target
			absCollectionsPath = collectionsPath
		}

		// Add ANSIBLE_COLLECTIONS_PATH environment variable pointing to container path
		containerCollectionsPath := "/tmp/.packer_ansible/collections"
		if !envVarIsSetOrPassed(env, "ANSIBLE_COLLECTIONS_PATH") {
			env.Set["ANSIBLE_COLLECTIONS_PATH"] = containerCollectionsPath
		}

		// Add volume mount for collections (read-only)
		// Check if this mount already exists to avoid duplicates
		mountExists := false
		for _, mount := range config.ExecutionEnvironment.VolumeMounts {
			if mount.Src == absCollectionsPath || mount.Dest == containerCollectionsPath {
				mountExists = true
				break
			}
		}
		if !mountExists {
			config.ExecutionEnvironment.VolumeMounts = append(config.ExecutionEnvironment.VolumeMounts,
				VolumeMount{
					Src:     absCollectionsPath,
					Dest:    containerCollectionsPath,
					Options: "ro",
				})
		}
	}
}

// generateNavigatorConfigYAML converts the NavigatorConfig struct to YAML format
// and applies automatic defaults when execution-environment.enabled = true
func generateNavigatorConfigYAML(config *NavigatorConfig, collectionsPath string) (string, error) {
	if config == nil {
		return "", fmt.Errorf("navigator_config cannot be nil")
	}

	applyAutomaticEEDefaults(config, collectionsPath)

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

	if config.Format != "" {
		result["format"] = config.Format
	}

	if config.TimeZone != "" {
		result["time-zone"] = config.TimeZone
	}

	if len(config.InventoryColumns) > 0 {
		result["inventory-columns"] = config.InventoryColumns
	}

	if config.CollectionDocCachePath != "" {
		result["collection-doc-cache-path"] = config.CollectionDocCachePath
	}

	if config.Color != nil {
		colorMap := make(map[string]interface{})
		if config.Color.Enable {
			colorMap["enable"] = true
		}
		if config.Color.Osc4 {
			colorMap["osc4"] = true
		}
		if len(colorMap) > 0 {
			result["color"] = colorMap
		}
	}

	if config.Editor != nil {
		editorMap := make(map[string]interface{})
		if config.Editor.Command != "" {
			editorMap["command"] = config.Editor.Command
		}
		if config.Editor.Console {
			editorMap["console"] = true
		}
		if len(editorMap) > 0 {
			result["editor"] = editorMap
		}
	}

	if config.Images != nil {
		imagesMap := make(map[string]interface{})
		if len(config.Images.Details) > 0 {
			imagesMap["details"] = config.Images.Details
		}
		if len(imagesMap) > 0 {
			result["images"] = imagesMap
		}
	}

	if config.ExecutionEnvironment != nil {
		eeMap := make(map[string]interface{})
		eeMap["enabled"] = config.ExecutionEnvironment.Enabled
		if config.ExecutionEnvironment.Image != "" {
			eeMap["image"] = config.ExecutionEnvironment.Image
		}
		if config.ExecutionEnvironment.ContainerEngine != "" {
			eeMap["container-engine"] = config.ExecutionEnvironment.ContainerEngine
		}
		if len(config.ExecutionEnvironment.ContainerOptions) > 0 {
			eeMap["container-options"] = config.ExecutionEnvironment.ContainerOptions
		}

		// execution-environment.pull.* is a nested object in ansible-navigator.yml.
		// We must create it when either policy or arguments are provided.
		pullMap := make(map[string]interface{})
		if config.ExecutionEnvironment.PullPolicy != "" {
			pullMap["policy"] = config.ExecutionEnvironment.PullPolicy
		}
		if len(config.ExecutionEnvironment.PullArguments) > 0 {
			pullMap["arguments"] = config.ExecutionEnvironment.PullArguments
		}
		if len(pullMap) > 0 {
			eeMap["pull"] = pullMap
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
		if len(config.ExecutionEnvironment.VolumeMounts) > 0 {
			volumeMounts := make([]map[string]interface{}, 0, len(config.ExecutionEnvironment.VolumeMounts))
			for _, mount := range config.ExecutionEnvironment.VolumeMounts {
				mountMap := make(map[string]interface{})
				mountMap["src"] = mount.Src
				mountMap["dest"] = mount.Dest
				if mount.Options != "" {
					mountMap["options"] = mount.Options
				}
				volumeMounts = append(volumeMounts, mountMap)
			}
			eeMap["volume-mounts"] = volumeMounts
		}
		result["execution-environment"] = eeMap
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
