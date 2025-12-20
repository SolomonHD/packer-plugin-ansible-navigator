// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

//go:generate packer-sdc mapstructure-to-hcl2 -type Config,Play,PathEntry,NavigatorConfig,ExecutionEnvironment,EnvironmentVariablesConfig,VolumeMount,AnsibleConfig,AnsibleConfigDefaults,AnsibleConfigConnection,LoggingConfig,PlaybookArtifact,CollectionDocCache
//go:generate packer-sdc struct-markdown

package ansiblenavigatorlocal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer-plugin-sdk/tmp"
	"github.com/hashicorp/packer-plugin-sdk/uuid"
)

// Compile-time interface check
var _ packersdk.Provisioner = &Provisioner{}

const DefaultStagingDir = "/tmp/packer-provisioner-ansible-local"

// NavigatorConfig represents the ansible-navigator.yml configuration structure.
// This is the root configuration container for ansible-navigator settings.
type NavigatorConfig struct {
	// Ansible-navigator execution mode
	Mode string `mapstructure:"mode"`
	// Execution environment configuration
	ExecutionEnvironment *ExecutionEnvironment `mapstructure:"execution_environment"`
	// Ansible configuration settings
	AnsibleConfig *AnsibleConfig `mapstructure:"ansible_config"`
	// Logging configuration
	Logging *LoggingConfig `mapstructure:"logging"`
	// Playbook artifact settings
	PlaybookArtifact *PlaybookArtifact `mapstructure:"playbook_artifact"`
	// Collection documentation cache settings
	CollectionDocCache *CollectionDocCache `mapstructure:"collection_doc_cache"`
}

// isPluginDebugEnabled returns true iff plugin debug output should be enabled.
//
// Spec: plugin debug mode is enabled if and only if navigator_config.logging.level
// equals "debug" (case-insensitive).
func isPluginDebugEnabled(nc *NavigatorConfig) bool {
	if nc == nil || nc.Logging == nil {
		return false
	}
	return strings.EqualFold(nc.Logging.Level, "debug")
}

func debugf(ui packersdk.Ui, enabled bool, format string, args ...interface{}) {
	if !enabled {
		return
	}
	ui.Message(fmt.Sprintf("[DEBUG] "+format, args...))
}

// ExecutionEnvironment represents execution environment settings
type ExecutionEnvironment struct {
	// Enable execution environment
	Enabled bool `mapstructure:"enabled"`
	// Container image to use
	Image string `mapstructure:"image"`
	// Pull policy for the container image
	PullPolicy string `mapstructure:"pull_policy"`
	// Environment variables to pass to the execution environment
	EnvironmentVariables *EnvironmentVariablesConfig `mapstructure:"environment_variables"`
	// Volume mounts for the execution environment container
	VolumeMounts []VolumeMount `mapstructure:"volume_mounts"`
}

// VolumeMount represents a volume mount for the execution environment
type VolumeMount struct {
	// Source path on the host
	Src string `mapstructure:"src"`
	// Destination path in the container
	Dest string `mapstructure:"dest"`
	// Mount options (e.g., "ro" for read-only)
	Options string `mapstructure:"options"`
}

// EnvironmentVariablesConfig represents environment variable configuration
type EnvironmentVariablesConfig struct {
	// List of environment variables to pass from the host
	Pass []string `mapstructure:"pass"`
	// Explicit key-value pairs of environment variables to set
	Set map[string]string `mapstructure:"set"`
}

// AnsibleConfig represents ansible-specific configuration
type AnsibleConfig struct {
	// Path to ansible.cfg file
	Config string `mapstructure:"config"`
	// Defaults section
	Defaults *AnsibleConfigDefaults `mapstructure:"defaults"`
	// SSH connection section
	SSHConnection *AnsibleConfigConnection `mapstructure:"ssh_connection"`
}

// AnsibleConfigDefaults represents ansible defaults configuration
type AnsibleConfigDefaults struct {
	// Remote temp directory
	RemoteTmp string `mapstructure:"remote_tmp"`
	// Local temp directory
	LocalTmp string `mapstructure:"local_tmp"`
	// Host key checking
	HostKeyChecking bool `mapstructure:"host_key_checking"`
}

// AnsibleConfigConnection represents ansible SSH connection settings
type AnsibleConfigConnection struct {
	// SSH timeout
	SSHTimeout int `mapstructure:"ssh_timeout"`
	// Pipelining
	Pipelining bool `mapstructure:"pipelining"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	// Log level
	Level string `mapstructure:"level"`
	// Log file path
	File string `mapstructure:"file"`
	// Append to log file
	Append bool `mapstructure:"append"`
}

// PlaybookArtifact represents playbook artifact settings
type PlaybookArtifact struct {
	// Enable playbook artifact
	Enable bool `mapstructure:"enable"`
	// Replay directory
	Replay string `mapstructure:"replay"`
	// Save directory
	SaveAs string `mapstructure:"save_as"`
}

// CollectionDocCache represents collection documentation cache settings
type CollectionDocCache struct {
	// Path to collection doc cache
	Path string `mapstructure:"path"`
	// Timeout for collection doc cache
	Timeout int `mapstructure:"timeout"`
}

// Play represents a single Ansible play execution with its configuration.
// It supports both traditional playbook files and Ansible Collection role FQDNs.
// Each play can have its own variables, tags, and privilege escalation settings.
type Play struct {
	// Name of the play for logging and identification
	Name string `mapstructure:"name"`
	// Target can be either a playbook path (.yml/.yaml) or a role FQDN
	Target string `mapstructure:"target"`
	// Extra variables to pass to this play
	ExtraVars map[string]string `mapstructure:"extra_vars"`
	// Tags to execute for this play
	Tags []string `mapstructure:"tags"`
	// Variables files to load for this play
	VarsFiles []string `mapstructure:"vars_files"`
	// Whether to use privilege escalation (become) for this play
	Become bool `mapstructure:"become"`
	// Extra arguments to pass verbatim to ansible-navigator for this play.
	// These are appended after `ansible-navigator run` (and enforced `--mode`),
	// and before plugin-generated inventory/extra-vars/etc.
	ExtraArgs []string `mapstructure:"extra_args"`
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	ctx                 interpolate.Context
	// The command to invoke ansible-navigator. Defaults to `ansible-navigator`.
	// This must be the executable name or path only, without additional arguments.
	// Use extra_arguments or play-level options for additional flags.
	// Examples: "ansible-navigator", "/usr/local/bin/ansible-navigator", "~/bin/ansible-navigator"
	Command string `mapstructure:"command"`
	// Additional directories to prepend to PATH when locating and running ansible-navigator on the target.
	// Entries are HOME-expanded locally and prepended to PATH in the remote shell command.
	// Example: ["~/bin", "/opt/ansible/bin"]
	AnsibleNavigatorPath []string `mapstructure:"ansible_navigator_path"`

	// Maximum time to wait for ansible-navigator version check to complete.
	// Defaults to "60s". This prevents indefinite hangs when ansible-navigator
	// is not properly configured or cannot be found.
	// Format: duration string (e.g., "30s", "1m", "90s").
	// Set to "0" to disable timeout (not recommended).
	VersionCheckTimeout *string `mapstructure:"version_check_timeout"`
	// Internal: tracks whether version_check_timeout was explicitly provided by the user.
	// This allows warning when skip_version_check makes it ineffective.
	versionCheckTimeoutWasSet bool `mapstructure:"-"`

	// Check if ansible-navigator is installed prior to running.
	// Set this to `true`, for example, if you're going to install it during the packer run.
	// Note: the local provisioner does not currently perform a version check, but this is
	// kept for configuration parity and for warning when it makes version_check_timeout ineffective.
	SkipVersionCheck bool `mapstructure:"skip_version_check"`

	// Modern Ansible Navigator fields (aligned with remote provisioner)
	// Continue executing remaining plays even if one fails.
	// When true, a play failure won't stop execution of subsequent plays.
	// Default: false
	KeepGoing bool `mapstructure:"keep_going"`
	// Enable structured JSON parsing and detailed task-level reporting.
	// When true, parses JSON events from ansible-navigator and provides enhanced error reporting.
	// Only effective when navigator_mode is set to "json".
	StructuredLogging bool `mapstructure:"structured_logging"`
	// Optional path to write a structured summary JSON file containing task results and failures.
	// Only used when structured_logging is enabled.
	LogOutputPath string `mapstructure:"log_output_path"`
	// Include detailed task output in logs when using structured logging.
	// Only effective when structured_logging is true.
	// Default: false
	VerboseTaskOutput bool `mapstructure:"verbose_task_output"`

	// Play-based execution (aligned with remote provisioner)

	// Array of play definitions supporting both playbooks and role FQDNs.
	// Mutually exclusive with playbook_file.
	Plays []Play `mapstructure:"play"`
	// Path to a unified requirements.yml file containing both roles and collections.
	// Alternative to galaxy_file with enhanced support for modern Ansible requirements format.
	RequirementsFile string `mapstructure:"requirements_file"`
	// Destination directory for installed roles.
	// This value is passed to ansible-galaxy as the roles install path and exported to Ansible via ANSIBLE_ROLES_PATH.
	// Defaults to ~/.packer.d/ansible_roles_cache if not specified.
	RolesPath string `mapstructure:"roles_path"`
	// Destination directory for installed collections.
	// This value is passed to ansible-galaxy as the collections install path and exported to Ansible via ANSIBLE_COLLECTIONS_PATHS.
	// Defaults to ~/.packer.d/ansible_collections_cache if not specified.
	CollectionsPath string `mapstructure:"collections_path"`
	// When true, skip network operations for both collections and roles.
	// This maps to ansible-galaxy --offline.
	OfflineMode bool `mapstructure:"offline_mode"`
	// The ansible-galaxy executable to invoke.
	// Defaults to "ansible-galaxy".
	GalaxyCommand string `mapstructure:"galaxy_command"`
	// Additional arguments appended to all ansible-galaxy invocations.
	GalaxyArgs []string `mapstructure:"galaxy_args"`
	// When true, pass --force to ansible-galaxy.
	// Ignored if galaxy_force_with_deps is true.
	GalaxyForce bool `mapstructure:"galaxy_force"`
	// When true, pass --force-with-deps to ansible-galaxy.
	GalaxyForceWithDeps bool `mapstructure:"galaxy_force_with_deps"`

	// Display extra vars JSON content in output for debugging.
	// When enabled, logs the extra vars JSON passed to ansible-navigator with sensitive values redacted.
	// Default: false
	ShowExtraVars bool `mapstructure:"show_extra_vars"`

	// Modern declarative ansible-navigator configuration via YAML file generation.
	// Maps directly to ansible-navigator.yml schema structure.
	// Supports full ansible-navigator.yml structure including:
	//   - ansible section (config overrides, playbook settings)
	//   - execution-environment object (enabled, image, pull-policy, environment-variables)
	//   - mode (stdout, json, yaml, interactive)
	//   - All other ansible-navigator.yml options
	// When provided:
	//   - Plugin generates temporary ansible-navigator.yml file on local side
	//   - Uploads it to the target alongside playbooks
	//   - Sets ANSIBLE_NAVIGATOR_CONFIG environment variable in remote shell
	//   - Automatically sets EE temp dir defaults when execution-environment.enabled = true
	//   - Cleans up config file after execution
	// Modern declarative ansible-navigator configuration via typed structs.
	// When provided, the plugin generates a temporary ansible-navigator.yml file.
	// This replaces the previous map[string]interface{} approach to ensure RPC serializability.
	// Use block syntax in HCL:
	//   navigator_config {
	//     mode = "stdout"
	//     execution_environment {
	//       enabled = true
	//       image = "quay.io/ansible/creator-ee:latest"
	//     }
	//   }
	NavigatorConfig *NavigatorConfig `mapstructure:"navigator_config"`
}

// Validate performs comprehensive validation of the Config
func (c *Config) Validate() error {
	var errs *packersdk.MultiError

	// Validate command contains no whitespace (must be executable only)
	if c.Command != "" && strings.ContainsAny(c.Command, " \t\n\r") {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
			"command must be only the executable name or path (no arguments). "+
				"Found whitespace in: %q. Use extra_arguments or play-level options for additional flags", c.Command))
	}

	// Validate play configuration
	if len(c.Plays) == 0 {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
			"at least one `play` block must be defined"))
	}

	// Validate plays
	for i, play := range c.Plays {
		if play.Target == "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
				"play %d: target must be specified", i))
			continue
		}

		// Validate playbook files referenced in plays
		if strings.HasSuffix(play.Target, ".yml") || strings.HasSuffix(play.Target, ".yaml") {
			if err := validateFileConfig(play.Target, fmt.Sprintf("play %d target", i), true); err != nil {
				errs = packersdk.MultiErrorAppend(errs, err)
			}
		}

		// Validate vars_files referenced in plays
		for j, varsFile := range play.VarsFiles {
			if err := validateFileConfig(varsFile, fmt.Sprintf("play %d vars_files[%d]", i, j), true); err != nil {
				errs = packersdk.MultiErrorAppend(errs, err)
			}
		}
	}

	// Validate galaxy_file
	// Validate requirements_file
	if c.RequirementsFile != "" {
		if err := validateFileConfig(c.RequirementsFile, "requirements_file", true); err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	// Validate version_check_timeout format
	if c.VersionCheckTimeout != nil {
		if _, err := time.ParseDuration(*c.VersionCheckTimeout); err != nil {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
				"invalid version_check_timeout: %q (must be a valid duration like '30s', '1m', '90s'): %w",
				*c.VersionCheckTimeout, err))
		}
	}

	// Validate navigator_config
	// Note: Empty struct pointer is valid (will be ignored), but if present, should have some configuration
	if c.NavigatorConfig != nil {
		// Basic validation - ensure at least one field is set
		isEmpty := c.NavigatorConfig.Mode == "" &&
			c.NavigatorConfig.ExecutionEnvironment == nil &&
			c.NavigatorConfig.AnsibleConfig == nil &&
			c.NavigatorConfig.Logging == nil &&
			c.NavigatorConfig.PlaybookArtifact == nil &&
			c.NavigatorConfig.CollectionDocCache == nil
		if isEmpty {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
				"navigator_config block cannot be empty. Either provide configuration or omit the block"))
		}

		// Schema compliance: ansible_config.config (path) is mutually exclusive with
		// the nested defaults/ssh_connection blocks (which map to a generated ansible.cfg).
		if c.NavigatorConfig.AnsibleConfig != nil {
			ac := c.NavigatorConfig.AnsibleConfig
			if ac.Config != "" && (ac.Defaults != nil || ac.SSHConnection != nil) {
				errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
					"navigator_config.ansible_config.config is mutually exclusive with navigator_config.ansible_config.defaults and navigator_config.ansible_config.ssh_connection"))
			}
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}
	return nil
}

type Provisioner struct {
	config Config

	stagingDir            string
	galaxyRolesPath       string
	galaxyCollectionsPath string
	generatedData         map[string]interface{}
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		PluginType:         "ansible-local",
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

	// Defaults
	if p.config.Command == "" {
		p.config.Command = "ansible-navigator"
	}

	// Apply HOME expansion to command if it looks like a path
	p.config.Command = expandUserPath(p.config.Command)

	// Apply HOME expansion to ansible_navigator_path entries
	for i, path := range p.config.AnsibleNavigatorPath {
		p.config.AnsibleNavigatorPath[i] = expandUserPath(path)
	}

	// Apply HOME expansion to path-like configuration fields on local side
	p.config.RequirementsFile = expandUserPath(p.config.RequirementsFile)
	p.config.CollectionsPath = expandUserPath(p.config.CollectionsPath)
	p.config.RolesPath = expandUserPath(p.config.RolesPath)
	p.config.GalaxyCommand = expandUserPath(p.config.GalaxyCommand)

	// Apply HOME expansion to plays
	for i := range p.config.Plays {
		p.config.Plays[i].Target = expandUserPath(p.config.Plays[i].Target)
		for j := range p.config.Plays[i].VarsFiles {
			p.config.Plays[i].VarsFiles[j] = expandUserPath(p.config.Plays[i].VarsFiles[j])
		}
	}
	// Detect explicit timeout setting before defaulting
	p.config.versionCheckTimeoutWasSet = p.config.VersionCheckTimeout != nil

	// Set default version check timeout
	if p.config.VersionCheckTimeout == nil {
		p.config.VersionCheckTimeout = stringPtr("60s")
	}

	p.stagingDir = filepath.ToSlash(filepath.Join(DefaultStagingDir, uuid.TimeOrderedUUID()))
	p.galaxyRolesPath = filepath.ToSlash(filepath.Join(p.stagingDir, "galaxy_roles"))
	p.galaxyCollectionsPath = filepath.ToSlash(filepath.Join(p.stagingDir, "galaxy_collections"))

	// Defaults for galaxy
	if p.config.GalaxyCommand == "" {
		p.config.GalaxyCommand = "ansible-galaxy"
	}

	// Set default install directories if not specified
	if p.config.CollectionsPath == "" {
		usr, err := os.UserHomeDir()
		if err == nil {
			p.config.CollectionsPath = filepath.Join(usr, ".packer.d", "ansible_collections_cache")
		}
	}

	if p.config.RolesPath == "" {
		usr, err := os.UserHomeDir()
		if err == nil {
			p.config.RolesPath = filepath.Join(usr, ".packer.d", "ansible_roles_cache")
		}
	}

	// Call comprehensive validation
	if err := p.config.Validate(); err != nil {
		return err
	}

	return nil
}

func (p *Provisioner) Provision(ctx context.Context, ui packersdk.Ui, comm packersdk.Communicator, generatedData map[string]interface{}) error {
	ui.Say("Provisioning with Ansible...")
	if p.config.SkipVersionCheck && p.config.versionCheckTimeoutWasSet {
		ui.Message("Warning: version_check_timeout is ignored when skip_version_check=true")
	}
	p.generatedData = generatedData

	ui.Message("Creating Ansible staging directory...")
	if err := p.createDir(ui, comm, p.stagingDir); err != nil {
		return fmt.Errorf("Error creating staging directory: %s", err)
	}
	defer func() {
		ui.Message("Removing staging directory...")
		_ = p.removeDir(ui, comm, p.stagingDir)
	}()

	// Upload requirements_file (if provided) into staging directory so GalaxyManager can install it
	if p.config.RequirementsFile != "" {
		ui.Message("Uploading requirements file...")
		remoteReq := filepath.ToSlash(filepath.Join(p.stagingDir, filepath.Base(p.config.RequirementsFile)))
		if err := p.uploadFile(ui, comm, remoteReq, p.config.RequirementsFile); err != nil {
			return fmt.Errorf("Error uploading requirements_file: %s", err)
		}
	}

	// Upload any vars_files referenced by plays (ansible uses -e @file; on-target execution needs remote files)
	for i := range p.config.Plays {
		for j, varsFile := range p.config.Plays[i].VarsFiles {
			ui.Message(fmt.Sprintf("Uploading vars file: %s", varsFile))
			remoteVars := filepath.ToSlash(filepath.Join(p.stagingDir, filepath.Base(varsFile)))
			if err := p.uploadFile(ui, comm, remoteVars, varsFile); err != nil {
				return fmt.Errorf("Error uploading vars_files: %s", err)
			}
			p.config.Plays[i].VarsFiles[j] = remoteVars
		}
	}

	// Create minimal inventory file (local connection) and upload
	tf, err := tmp.File("packer-provisioner-ansible-navigator-local-inventory")
	if err != nil {
		return fmt.Errorf("Error preparing inventory file: %s", err)
	}
	defer os.Remove(tf.Name())
	if _, err := tf.Write([]byte("127.0.0.1")); err != nil {
		tf.Close()
		return fmt.Errorf("Error preparing inventory file: %s", err)
	}
	if err := tf.Close(); err != nil {
		return fmt.Errorf("Error preparing inventory file: %s", err)
	}
	inventoryRemote := filepath.ToSlash(filepath.Join(p.stagingDir, "inventory.ini"))
	ui.Message("Uploading inventory file...")
	if err := p.uploadFile(ui, comm, inventoryRemote, tf.Name()); err != nil {
		return fmt.Errorf("Error uploading inventory file: %s", err)
	}

	// Generate and upload ansible-navigator.yml if configured
	var navigatorConfigRemotePath string
	if p.config.NavigatorConfig != nil {
		// For local provisioner, collections are in the staging directory on the target
		collectionsPath := p.galaxyCollectionsPath

		// If ansible_config.defaults / ansible_config.ssh_connection are provided
		// (explicitly or via EE defaults), generate an ansible.cfg file, upload it
		// to the staging directory, and reference it from the navigator config YAML
		// via ansible.config.path.
		if p.config.NavigatorConfig.AnsibleConfig != nil && needsGeneratedAnsibleCfg(p.config.NavigatorConfig.AnsibleConfig) {
			if p.config.NavigatorConfig.AnsibleConfig.Config != "" {
				return fmt.Errorf("invalid navigator_config.ansible_config: config is mutually exclusive with defaults/ssh_connection")
			}

			cfgContent, err := generateAnsibleCfgContent(p.config.NavigatorConfig.AnsibleConfig)
			if err != nil {
				return fmt.Errorf("Error generating ansible.cfg content: %s", err)
			}
			if cfgContent != "" {
				cfgTmpPath, err := createTempAnsibleCfgFile(cfgContent)
				if err != nil {
					return fmt.Errorf("Error creating temporary ansible.cfg: %s", err)
				}
				defer os.Remove(cfgTmpPath)

				cfgRemotePath := filepath.ToSlash(filepath.Join(p.stagingDir, "ansible.cfg"))
				ui.Message("Uploading generated ansible.cfg...")
				if err := p.uploadFile(ui, comm, cfgRemotePath, cfgTmpPath); err != nil {
					return fmt.Errorf("Error uploading ansible.cfg: %s", err)
				}

				p.config.NavigatorConfig.AnsibleConfig.Config = cfgRemotePath
			}
		}

		yamlContent, err := generateNavigatorConfigYAML(p.config.NavigatorConfig, collectionsPath)
		if err != nil {
			return fmt.Errorf("Error generating navigator_config YAML: %s", err)
		}

		// Create temporary local file
		tmpPath, err := createNavigatorConfigFile(yamlContent)
		if err != nil {
			return fmt.Errorf("Error creating temporary ansible-navigator.yml: %s", err)
		}
		defer os.Remove(tmpPath)

		ui.Message("Uploading generated ansible-navigator.yml...")
		// Upload to remote staging directory
		navigatorConfigRemotePath = filepath.ToSlash(filepath.Join(p.stagingDir, "ansible-navigator.yml"))
		if err := p.uploadFile(ui, comm, navigatorConfigRemotePath, tmpPath); err != nil {
			return fmt.Errorf("Error uploading ansible-navigator.yml: %s", err)
		}
	}

	// Install dependencies using GalaxyManager
	galaxyManager := NewGalaxyManager(&p.config, ui, comm, p.stagingDir, p.galaxyRolesPath, p.galaxyCollectionsPath)
	if err := galaxyManager.InstallRequirements(); err != nil {
		return fmt.Errorf("Error installing requirements: %s", err)
	}

	if err := p.executeAnsible(ui, comm, inventoryRemote, navigatorConfigRemotePath); err != nil {
		return fmt.Errorf("Error executing Ansible Navigator: %s", err)
	}
	return nil
}

func stringPtr(s string) *string { return &s }
func (p *Provisioner) executeAnsible(ui packersdk.Ui, comm packersdk.Communicator, inventoryRemotePath string, navigatorConfigRemotePath string) error {
	// Execute plays (required by validation)
	return p.executePlays(ui, comm, inventoryRemotePath, navigatorConfigRemotePath)
}

// executePlays executes multiple Ansible plays in sequence
func (p *Provisioner) executePlays(ui packersdk.Ui, comm packersdk.Communicator, inventory string, navigatorConfigRemotePath string) error {
	debugEnabled := isPluginDebugEnabled(p.config.NavigatorConfig)
	debugf(ui, debugEnabled, "Plugin debug mode enabled (gated by navigator_config.logging.level=debug)")
	debugf(ui, debugEnabled, "ansible-navigator command=%q", p.config.Command)
	if len(p.config.AnsibleNavigatorPath) > 0 {
		debugf(ui, debugEnabled, "ansible_navigator_path prefixes=%v", p.config.AnsibleNavigatorPath)
	} else {
		debugf(ui, debugEnabled, "ansible_navigator_path not set; using existing PATH")
	}
	if navigatorConfigRemotePath != "" {
		debugf(ui, debugEnabled, "ANSIBLE_NAVIGATOR_CONFIG=%s", navigatorConfigRemotePath)
	}

	for i, play := range p.config.Plays {
		playName := play.Name
		if playName == "" {
			playName = fmt.Sprintf("Play %d", i+1)
		}

		ui.Say(fmt.Sprintf("Executing %s: %s", playName, play.Target))

		var playbookPath string
		var cleanupFunc func()

		// Determine if target is a playbook file or a role FQDN
		if strings.HasSuffix(play.Target, ".yml") || strings.HasSuffix(play.Target, ".yaml") {
			// It's a playbook file - upload to remote
			if absPath, err := filepath.Abs(play.Target); err == nil {
				debugf(ui, debugEnabled, "Resolved playbook path: %s -> %s", play.Target, absPath)
			}
			ui.Message(fmt.Sprintf("Uploading playbook: %s", play.Target))
			remotePath := filepath.ToSlash(filepath.Join(p.stagingDir, filepath.Base(play.Target)))
			debugf(ui, debugEnabled, "Remote playbook path=%s", remotePath)
			if err := p.uploadFile(ui, comm, remotePath, play.Target); err != nil {
				return fmt.Errorf("Play '%s': failed to upload playbook: %s", playName, err)
			}
			playbookPath = remotePath
		} else {
			// It's a role FQDN - generate temporary playbook locally then upload
			debugf(ui, debugEnabled, "Play target treated as role; generating temporary playbook for role=%s", play.Target)
			ui.Message(fmt.Sprintf("Generating temporary playbook for role: %s", play.Target))
			tmpPlaybook, err := p.createRolePlaybook(play.Target, play)
			if err != nil {
				return fmt.Errorf("play %q: failed to generate role playbook: %w", playName, err)
			}
			debugf(ui, debugEnabled, "Generated temporary playbook path=%s", tmpPlaybook)

			// Upload generated playbook to remote
			remotePath := filepath.ToSlash(filepath.Join(p.stagingDir, filepath.Base(tmpPlaybook)))
			debugf(ui, debugEnabled, "Remote generated playbook path=%s", remotePath)
			if err := p.uploadFile(ui, comm, remotePath, tmpPlaybook); err != nil {
				os.Remove(tmpPlaybook)
				return fmt.Errorf("Play '%s': failed to upload generated playbook: %s", playName, err)
			}
			playbookPath = remotePath
			cleanupFunc = func() {
				os.Remove(tmpPlaybook)
			}
		}

		// Execute the play
		err := p.executeAnsiblePlaybook(ui, comm, playbookPath, play, inventory, navigatorConfigRemotePath)

		// Cleanup temporary playbook if it was generated
		if cleanupFunc != nil {
			cleanupFunc()
		}

		if err != nil {
			ui.Error(fmt.Sprintf("Play '%s' failed: %v", playName, err))
			// If keep_going is false, return immediately on error
			if !p.config.KeepGoing {
				return fmt.Errorf("Play '%s' failed with exit code 2", playName)
			}
			// Otherwise, log but continue to next play
			ui.Message(fmt.Sprintf("Continuing to next play despite failure (keep_going=true)"))
		}

		if i < len(p.config.Plays)-1 {
			ui.Message(fmt.Sprintf("Completed %s", playName))
		}
	}

	ui.Say("All plays completed successfully!")
	return nil
}

// logExtraVarsJSON logs the extra vars JSON with sensitive values redacted
func logExtraVarsJSON(ui packersdk.Ui, extraVars map[string]interface{}) {
	// Create a copy for sanitization
	sanitized := make(map[string]interface{})
	for k, v := range extraVars {
		// Redact sensitive keys
		if k == "ansible_password" || strings.Contains(strings.ToLower(k), "password") {
			sanitized[k] = "*****"
		} else {
			// Note: ansible_ssh_private_key_file path is shown (path is not secret, content is)
			sanitized[k] = v
		}
	}

	// Marshal to formatted JSON
	jsonBytes, err := json.MarshalIndent(sanitized, "", "  ")
	if err != nil {
		ui.Message(fmt.Sprintf("[Extra Vars] Failed to format JSON: %v", err))
		return
	}

	ui.Message(fmt.Sprintf("[Extra Vars] JSON content:\n%s", string(jsonBytes)))
}

// createExtraVarsFile writes provisioner-generated extra vars to a temporary JSON file
// and returns the file path. For local provisioner, this file will be uploaded to the
// staging directory on the target. The caller is responsible for cleanup (which happens
// automatically when the staging directory is removed).
//
// This file-based approach prevents shell interpretation errors when ansible-navigator
// invokes ansible-playbook inside execution environment containers. Inline JSON passed
// via --extra-vars "{...}" gets shell-interpreted inside the container, causing brace
// expansion to split the argument.
func createExtraVarsFile(extraVars map[string]interface{}) (string, error) {
	// Marshal extra vars to JSON
	extraVarsJSON, err := json.MarshalIndent(extraVars, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal extra vars to JSON: %w", err)
	}

	// Create temporary file with unique name
	tmpFile, err := os.CreateTemp("", "packer-extravars-*.json")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary extra vars file: %w", err)
	}

	// Write JSON content
	if _, err := tmpFile.Write(extraVarsJSON); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write extra vars to file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to close extra vars file: %w", err)
	}

	return tmpFile.Name(), nil
}

// buildPluginArgsForPlay constructs ansible-navigator arguments and returns
// the local path to the temporary extra vars file (to be uploaded to staging directory).
// The file should be cleaned up by the caller (or automatically via staging directory cleanup).
func (p *Provisioner) buildPluginArgsForPlay(ui packersdk.Ui, play Play, inventory string) (args []string, extraVarsLocalPath string, err error) {
	args = make([]string, 0)

	// Provisioner-generated extra vars MUST be conveyed via a single JSON object
	// passed through exactly one -e/--extra-vars argument pair to avoid malformed
	// argument construction and positional argument shifting.
	//
	// To prevent shell interpretation errors in execution environments, we write
	// the JSON to a temporary file and pass it via @filepath syntax.
	extraVars := make(map[string]interface{})
	if p.config.PackerBuildName != "" {
		extraVars["packer_build_name"] = p.config.PackerBuildName
	}
	extraVars["packer_builder_type"] = p.config.PackerBuilderType
	if httpAddr, ok := p.generatedData["PackerHTTPAddr"]; ok {
		extraVars["packer_http_addr"] = fmt.Sprint(httpAddr)
	}

	// Create temp file with extra vars
	extraVarsLocalPath, err = createExtraVarsFile(extraVars)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create extra vars file: %w", err)
	}

	// NOTE: The actual --extra-vars argument with @filepath will be added by the caller
	// after uploading the file to the staging directory on the remote target.

	// Log extra vars if ShowExtraVars is enabled
	if p.config.ShowExtraVars {
		logExtraVarsJSON(ui, extraVars)
	}

	if play.Become {
		args = append(args, "--become")
	}

	for _, tag := range play.Tags {
		args = append(args, fmt.Sprintf("--tags=%s", tag))
	}

	if len(play.ExtraVars) > 0 {
		keys := make([]string, 0, len(play.ExtraVars))
		for k := range play.ExtraVars {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			args = append(args, fmt.Sprintf("-e=%s=%s", k, play.ExtraVars[k]))
		}
	}

	for _, varsFile := range play.VarsFiles {
		args = append(args, fmt.Sprintf("-e=@%s", varsFile))
	}

	args = append(args, "-c=local", fmt.Sprintf("-i=%s", inventory))

	return args, extraVarsLocalPath, nil
}

// shellEscapePOSIX returns a POSIX-shell-safe representation of s.
// It uses single-quote escaping, suitable for remote shell commands.
func shellEscapePOSIX(s string) string {
	if s == "" {
		return "''"
	}
	// Fast path: no characters that require quoting.
	if !strings.ContainsAny(s, " \t\n\r\"'\\$&;|<>*?[]{}()!`") {
		return s
	}
	// Single-quote escape: close, escape single quote, reopen.
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

func shellEscapeAllPOSIX(args []string) []string {
	out := make([]string, 0, len(args))
	for _, a := range args {
		out = append(out, shellEscapePOSIX(a))
	}
	return out
}

// createRolePlaybook generates a temporary playbook file for executing an Ansible role
func (p *Provisioner) createRolePlaybook(role string, play Play) (string, error) {
	// Create a temporary file for the playbook
	tmpFile, err := tmp.File("packer-role-playbook-*.yml")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary playbook file: %w", err)
	}

	// Build the playbook content
	playbookContent := "---\n- hosts: all\n"
	playbookContent += "  connection: local\n"

	if play.Become {
		playbookContent += "  become: yes\n"
	}

	if len(play.VarsFiles) > 0 {
		playbookContent += "  vars_files:\n"
		for _, varsFile := range play.VarsFiles {
			playbookContent += fmt.Sprintf("    - %s\n", varsFile)
		}
	}

	playbookContent += "  roles:\n"
	playbookContent += fmt.Sprintf("    - role: %s\n", role)

	if len(play.ExtraVars) > 0 {
		playbookContent += "      vars:\n"
		for k, v := range play.ExtraVars {
			playbookContent += fmt.Sprintf("        %s: %s\n", k, v)
		}
	}

	// Write the content
	if _, err := tmpFile.WriteString(playbookContent); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write temporary playbook: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to close temporary playbook: %w", err)
	}

	return tmpFile.Name(), nil
}

func (p *Provisioner) executeAnsiblePlaybook(
	ui packersdk.Ui,
	comm packersdk.Communicator,
	playbookFile string,
	play Play,
	inventory string,
	navigatorConfigRemotePath string,
) error {
	ctx := context.TODO()
	env_vars := ""

	debugEnabled := isPluginDebugEnabled(p.config.NavigatorConfig)

	// Build environment variables for collections and roles paths
	galaxyManager := NewGalaxyManager(&p.config, ui, comm, p.stagingDir, p.galaxyRolesPath, p.galaxyCollectionsPath)
	envPaths := galaxyManager.SetupEnvironmentPaths()
	if len(envPaths) > 0 {
		env_vars = strings.Join(envPaths, " ") + " "
	}

	// Add ANSIBLE_NAVIGATOR_CONFIG if provided
	if navigatorConfigRemotePath != "" {
		env_vars += fmt.Sprintf("ANSIBLE_NAVIGATOR_CONFIG=%s ", navigatorConfigRemotePath)
		debugf(ui, debugEnabled, "Setting ANSIBLE_NAVIGATOR_CONFIG=%s", navigatorConfigRemotePath)
	}

	// Add standard Ansible environment variables
	env_vars += "ANSIBLE_FORCE_COLOR=1 PYTHONUNBUFFERED=1"

	// Build PATH override if ansible_navigator_path is set
	pathPrefixAssignment := ""
	if len(p.config.AnsibleNavigatorPath) > 0 {
		pathPrefixAssignment = buildPathPrefixForRemoteShell(p.config.AnsibleNavigatorPath)
		debugf(ui, debugEnabled, "Remote shell will prefix PATH with %v", p.config.AnsibleNavigatorPath)
	}
	pathPrefix := ""
	if pathPrefixAssignment != "" {
		pathPrefix = pathPrefixAssignment + " "
	}

	// Build plugin args and get local extra vars file path
	pluginArgs, extraVarsLocalPath, err := p.buildPluginArgsForPlay(ui, play, inventory)
	if err != nil {
		return fmt.Errorf("failed to build plugin args: %w", err)
	}

	// Upload extra vars file to staging directory
	extraVarsRemotePath := ""
	if extraVarsLocalPath != "" {
		defer os.Remove(extraVarsLocalPath) // Clean up local temp file

		extraVarsRemotePath = filepath.ToSlash(filepath.Join(p.stagingDir, filepath.Base(extraVarsLocalPath)))
		debugf(ui, debugEnabled, "Uploading extra vars file: %s -> %s", extraVarsLocalPath, extraVarsRemotePath)
		if err := p.uploadFile(ui, comm, extraVarsRemotePath, extraVarsLocalPath); err != nil {
			return fmt.Errorf("failed to upload extra vars file: %w", err)
		}
		debugf(ui, debugEnabled, "Extra vars file uploaded to staging directory")
	}

	// Deterministic ordering:
	//   1) ansible-navigator run (+ enforced --mode)
	//   2) play.extra_args (verbatim)
	//   3) plugin-generated inventory/extra-vars/etc (including play-level flags)
	//   4) play target (playbook path)
	runArgs := []string{"run"}
	if p.config.NavigatorConfig != nil && p.config.NavigatorConfig.Mode != "" {
		runArgs = append(runArgs, fmt.Sprintf("--mode=%s", p.config.NavigatorConfig.Mode))
	}
	runArgs = append(runArgs, play.ExtraArgs...)

	// Add the extra-vars file reference BEFORE other plugin args
	if extraVarsRemotePath != "" {
		runArgs = append(runArgs, fmt.Sprintf("--extra-vars=@%s", extraVarsRemotePath))
	}

	runArgs = append(runArgs, pluginArgs...)
	runArgs = append(runArgs, playbookFile)

	command := ""
	if debugEnabled && isExecutionEnvironmentEnabled(p.config.NavigatorConfig) {
		preflight := buildEEDockerPreflightShell(pathPrefixAssignment)
		command = fmt.Sprintf(
			"cd %s && %s; %s%s %s %s",
			shellEscapePOSIX(p.stagingDir),
			preflight,
			pathPrefix,
			env_vars,
			shellEscapePOSIX(p.config.Command),
			strings.Join(shellEscapeAllPOSIX(runArgs), " "),
		)
	} else {
		command = fmt.Sprintf(
			"cd %s && %s%s %s %s",
			shellEscapePOSIX(p.stagingDir),
			pathPrefix,
			env_vars,
			shellEscapePOSIX(p.config.Command),
			strings.Join(shellEscapeAllPOSIX(runArgs), " "),
		)
	}
	ui.Message(fmt.Sprintf("Executing Ansible Navigator: %s", command))
	cmd := &packersdk.RemoteCmd{
		Command: command,
	}
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return err
	}
	if cmd.ExitStatus() != 0 {
		if cmd.ExitStatus() == 127 {
			return fmt.Errorf("%s could not be found. Verify that it is available on the\n"+
				"PATH after connecting to the machine.",
				p.config.Command)
		}

		return fmt.Errorf("Non-zero exit status: %d", cmd.ExitStatus())
	}
	return nil
}

func validateDirConfig(path string, config string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s: %s is invalid: %s", config, path, err)
	} else if !info.IsDir() {
		return fmt.Errorf("%s: %s must point to a directory", config, path)
	}
	return nil
}

func validateFileConfig(name string, config string, req bool) error {
	if req {
		if name == "" {
			return fmt.Errorf("%s must be specified.", config)
		}
	}
	info, err := os.Stat(name)
	if err != nil {
		return fmt.Errorf("%s: %s is invalid: %s", config, name, err)
	} else if info.IsDir() {
		return fmt.Errorf("%s: %s must point to a file", config, name)
	}
	return nil
}

func (p *Provisioner) uploadFile(ui packersdk.Ui, comm packersdk.Communicator, dst, src string) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("Error opening: %s", err)
	}
	defer f.Close()

	if err = comm.Upload(dst, f, nil); err != nil {
		return fmt.Errorf("Error uploading %s: %s", src, err)
	}
	return nil
}

func (p *Provisioner) createDir(ui packersdk.Ui, comm packersdk.Communicator, dir string) error {
	ctx := context.TODO()
	cmd := &packersdk.RemoteCmd{
		Command: fmt.Sprintf("mkdir -p '%s'", dir),
	}

	ui.Message(fmt.Sprintf("Creating directory: %s", dir))
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return err
	}

	if cmd.ExitStatus() != 0 {
		return fmt.Errorf("Non-zero exit status. See output above for more information.")
	}
	return nil
}

func (p *Provisioner) removeDir(ui packersdk.Ui, comm packersdk.Communicator, dir string) error {
	ctx := context.TODO()
	cmd := &packersdk.RemoteCmd{
		Command: fmt.Sprintf("rm -rf '%s'", dir),
	}

	ui.Message(fmt.Sprintf("Removing directory: %s", dir))
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return err
	}

	if cmd.ExitStatus() != 0 {
		return fmt.Errorf("Non-zero exit status. See output above for more information.")
	}
	return nil
}

func (p *Provisioner) uploadDir(ui packersdk.Ui, comm packersdk.Communicator, dst, src string) error {
	if err := p.createDir(ui, comm, dst); err != nil {
		return err
	}

	// Make sure there is a trailing "/" so that the directory isn't
	// created on the other side.
	if src[len(src)-1] != '/' {
		src = src + "/"
	}
	return comm.UploadDir(dst, src, nil)
}

// expandUserPath expands HOME-relative paths on the local side.
// It handles:
// - "~" -> $HOME
// - "~/subdir" -> $HOME/subdir
// - "~user/..." -> unchanged (no multi-user home resolution)
// - Other paths -> unchanged
func expandUserPath(path string) string {
	if path == "" {
		return path
	}

	// Only expand if it starts with ~
	if !strings.HasPrefix(path, "~") {
		return path
	}

	// Don't expand ~user/ patterns (multi-user home directories)
	if len(path) > 1 && path[1] != '/' && path[1] != filepath.Separator {
		return path
	}

	// Get HOME directory
	home, err := os.UserHomeDir()
	if err != nil {
		// If we can't get HOME, return the path unchanged
		return path
	}

	// Handle bare "~"
	if path == "~" {
		return home
	}

	// Handle "~/..." pattern
	if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~"+string(filepath.Separator)) {
		return filepath.Join(home, path[2:])
	}

	// Shouldn't reach here, but return unchanged if we do
	return path
}

// buildPathPrefixForRemoteShell constructs a PATH override prefix for remote shell commands
// Returns string in format: PATH="dir1:dir2:$PATH"
// Returns empty string if no entries provided
func buildPathPrefixForRemoteShell(ansibleNavigatorPath []string) string {
	if len(ansibleNavigatorPath) == 0 {
		return ""
	}

	// Join expanded paths with colon (standard Unix path separator)
	// Entries are already HOME-expanded in Prepare()
	pathEntries := strings.Join(ansibleNavigatorPath, ":")

	return fmt.Sprintf(`PATH="%s:$PATH"`, pathEntries)
}
