// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansiblenavigator

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
// collectionsPath is provided. The collectionsPath parameter should be the parent
// directory containing the ansible_collections/ subdirectory (e.g., ~/.packer.d/ansible_collections_cache).
// ansible-galaxy installs collections to <collectionsPath>/ansible_collections/<namespace>/<collection>,
// and this function mounts the entire collectionsPath directory into the container.
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
		// Expand tilde to get absolute path for volume mount
		absCollectionsPath := expandUserPath(collectionsPath)
		// Ensure we have an absolute path
		if !filepath.IsAbs(absCollectionsPath) {
			var err error
			absCollectionsPath, err = filepath.Abs(absCollectionsPath)
			if err != nil {
				// Fallback to unexpanded path if Abs() fails
				absCollectionsPath = collectionsPath
			}
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
// and applies automatic defaults when execution-environment.enabled = true.
//
// The generated YAML conforms to ansible-navigator Version 2 format, which wraps
// all configuration under the "ansible-navigator:" top-level key. This format is
// required by ansible-navigator 25.x+ to avoid version migration prompts.
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
// (using hyphens instead of underscores for ansible-navigator.yml compatibility).
//
// This function generates ansible-navigator Version 2 format by wrapping all settings
// under the "ansible-navigator:" top-level key. Version 2 format is required by
// ansible-navigator 25.x+ and ensures settings like pull-policy are correctly recognized.
func convertToYAMLStructure(config *NavigatorConfig) map[string]interface{} {
	// Create nested structure with ansible-navigator root key (Version 2 format requirement)
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
		if config.ExecutionEnvironment.ContainerEngine != "" {
			eeMap["container-engine"] = config.ExecutionEnvironment.ContainerEngine
		}
		if len(config.ExecutionEnvironment.ContainerOptions) > 0 {
			eeMap["container-options"] = config.ExecutionEnvironment.ContainerOptions
		}

		// execution-environment.pull.* is a nested object in ansible-navigator.yml Version 2 format.
		// The pull policy MUST be nested as "pull: { policy: <value> }" rather than flat "pull-policy: <value>".
		// This nested structure is required for ansible-navigator 25.x+ to correctly recognize and honor
		// the pull policy setting (e.g., "never" to prevent Docker registry pulls).
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

// buildNavigatorCLIFlags converts NavigatorConfig into ansible-navigator CLI flags.
// This is the primary configuration method, avoiding YAML generation in most cases.
//
// CLI Flag Mapping Table:
//
//	HCL Field                                 -> CLI Flag
//	-----------------------------------------------------------------------
//	navigator_config.mode                     -> --mode <value>
//	execution_environment.enabled (false)     -> --execution-environment false
//	execution_environment.image               -> --execution-environment-image <value>
//	execution_environment.container_engine    -> --execution-environment-container-engine <value>
//	execution_environment.pull_policy         -> --pull-policy <value>
//	execution_environment.container_options   -> --container-options '<json-array>'
//	execution_environment.environment_variables.set  -> --eev KEY=VALUE (repeatable)
//	execution_environment.volume_mounts       -> --evm src:dest[:options] (repeatable)
//	logging.level                             -> --log-level <value>
//	logging.file                              -> --log-file <value>
//	logging.append (true)                     -> --log-append true
//	ansible_config.config                     -> --ansible-config <path>
//
// Returns a slice of CLI flags ready to append to ansible-navigator command.
func buildNavigatorCLIFlags(config *NavigatorConfig) []string {
	if config == nil {
		return nil
	}

	var flags []string

	// Mode flag
	if config.Mode != "" {
		flags = append(flags, fmt.Sprintf("--mode=%s", config.Mode))
	}

	// Execution environment flags
	if config.ExecutionEnvironment != nil {
		// Explicit disable takes precedence
		if !config.ExecutionEnvironment.Enabled {
			flags = append(flags, "--execution-environment=false")
		} else {
			// When enabled, specify other EE flags
			if config.ExecutionEnvironment.Image != "" {
				flags = append(flags, fmt.Sprintf("--execution-environment-image=%s", config.ExecutionEnvironment.Image))
			}
			if config.ExecutionEnvironment.ContainerEngine != "" {
				flags = append(flags, fmt.Sprintf("--execution-environment-container-engine=%s", config.ExecutionEnvironment.ContainerEngine))
			}
			if config.ExecutionEnvironment.PullPolicy != "" {
				flags = append(flags, fmt.Sprintf("--pull-policy=%s", config.ExecutionEnvironment.PullPolicy))
			}

			// Container options (passed as JSON array string)
			if len(config.ExecutionEnvironment.ContainerOptions) > 0 {
				// ansible-navigator expects --container-options as a JSON array string
				// Format: --container-options '["--network=host","--cap-add=NET_ADMIN"]'
				optionsJSON := "["
				for i, opt := range config.ExecutionEnvironment.ContainerOptions {
					if i > 0 {
						optionsJSON += ","
					}
					// Escape quotes in the option value
					escapedOpt := strings.ReplaceAll(opt, "\"", "\\\"")
					optionsJSON += fmt.Sprintf("\"%s\"", escapedOpt)
				}
				optionsJSON += "]"
				flags = append(flags, fmt.Sprintf("--container-options=%s", optionsJSON))
			}

			// Environment variables (repeatable --eev flag)
			if config.ExecutionEnvironment.EnvironmentVariables != nil {
				// Set variables as KEY=VALUE
				for key, value := range config.ExecutionEnvironment.EnvironmentVariables.Set {
					flags = append(flags, fmt.Sprintf("--eev=%s=%s", key, value))
				}
				// Pass-through variables (just the key name, ansible-navigator will read from host env)
				for _, key := range config.ExecutionEnvironment.EnvironmentVariables.Pass {
					flags = append(flags, fmt.Sprintf("--eev=%s", key))
				}
			}

			// Volume mounts (repeatable --evm flag)
			// Format: --evm src:dest[:options]
			for _, mount := range config.ExecutionEnvironment.VolumeMounts {
				evmValue := fmt.Sprintf("%s:%s", mount.Src, mount.Dest)
				if mount.Options != "" {
					evmValue += ":" + mount.Options
				}
				flags = append(flags, fmt.Sprintf("--evm=%s", evmValue))
			}
		}
	}

	// Logging flags
	if config.Logging != nil {
		if config.Logging.Level != "" {
			flags = append(flags, fmt.Sprintf("--log-level=%s", config.Logging.Level))
		}
		if config.Logging.File != "" {
			flags = append(flags, fmt.Sprintf("--log-file=%s", config.Logging.File))
		}
		if config.Logging.Append {
			flags = append(flags, "--log-append=true")
		}
	}

	// Ansible config path flag
	if config.AnsibleConfig != nil && config.AnsibleConfig.Config != "" {
		flags = append(flags, fmt.Sprintf("--ansible-config=%s", config.AnsibleConfig.Config))
	}

	return flags
}

// hasUnmappedSettings checks if NavigatorConfig contains settings that cannot be
// expressed via CLI flags and require YAML configuration fallback.
//
// Currently unmapped settings (require YAML):
//   - playbook_artifact.*  (enable, replay, save-as)
//   - collection_doc_cache.* (path, timeout)
//
// Returns true if minimal YAML generation is needed.
func hasUnmappedSettings(config *NavigatorConfig) bool {
	if config == nil {
		return false
	}

	// Check for playbook artifact settings
	if config.PlaybookArtifact != nil {
		if config.PlaybookArtifact.Enable || config.PlaybookArtifact.Replay != "" || config.PlaybookArtifact.SaveAs != "" {
			return true
		}
	}

	// Check for collection doc cache settings
	if config.CollectionDocCache != nil {
		if config.CollectionDocCache.Path != "" || config.CollectionDocCache.Timeout > 0 {
			return true
		}
	}

	return false
}

// generateMinimalYAML generates a minimal ansible-navigator.yml containing ONLY
// settings that cannot be expressed via CLI flags (playbook-artifact, collection-doc-cache).
//
// This function is only called when hasUnmappedSettings() returns true.
// Returns empty string if no unmapped settings exist.
func generateMinimalYAML(config *NavigatorConfig) (string, error) {
	if config == nil || !hasUnmappedSettings(config) {
		return "", nil
	}

	// Build minimal YAML structure with ONLY unmapped settings
	ansibleNavigator := make(map[string]interface{})

	// Include playbook-artifact if configured
	if config.PlaybookArtifact != nil {
		if config.PlaybookArtifact.Enable || config.PlaybookArtifact.Replay != "" || config.PlaybookArtifact.SaveAs != "" {
			artifactMap := make(map[string]interface{})
			artifactMap["enable"] = config.PlaybookArtifact.Enable
			if config.PlaybookArtifact.Replay != "" {
				artifactMap["replay"] = config.PlaybookArtifact.Replay
			}
			if config.PlaybookArtifact.SaveAs != "" {
				artifactMap["save-as"] = config.PlaybookArtifact.SaveAs
			}
			ansibleNavigator["playbook-artifact"] = artifactMap
		}
	}

	// Include collection-doc-cache if configured
	if config.CollectionDocCache != nil {
		if config.CollectionDocCache.Path != "" || config.CollectionDocCache.Timeout > 0 {
			cacheMap := make(map[string]interface{})
			if config.CollectionDocCache.Path != "" {
				cacheMap["path"] = config.CollectionDocCache.Path
			}
			if config.CollectionDocCache.Timeout > 0 {
				cacheMap["timeout"] = config.CollectionDocCache.Timeout
			}
			ansibleNavigator["collection-doc-cache"] = cacheMap
		}
	}

	// If no unmapped settings, return empty
	if len(ansibleNavigator) == 0 {
		return "", nil
	}

	// Wrap under ansible-navigator root key (Version 2 format)
	yamlConfig := map[string]interface{}{
		"ansible-navigator": ansibleNavigator,
	}

	// Marshal to YAML
	yamlData, err := yaml.Marshal(yamlConfig)
	if err != nil {
		return "", fmt.Errorf("failed to marshal minimal YAML: %w", err)
	}

	return string(yamlData), nil
}
