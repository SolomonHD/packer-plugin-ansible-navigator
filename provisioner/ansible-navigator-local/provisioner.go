// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

//go:generate packer-sdc mapstructure-to-hcl2 -type Config,Play,PathEntry
//go:generate packer-sdc struct-markdown

package ansiblenavigatorlocal

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
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
	// Extra arguments to pass to Ansible.
	// These arguments _will not_ be passed through a shell and arguments should
	// not be quoted. Usage example:
	//
	// ```json
	//    "extra_arguments": [ "--extra-vars", "Region={{user `Region`}} Stage={{user `Stage`}}" ]
	// ```
	//
	// In certain scenarios where you want to pass ansible command line arguments
	// that include parameter and value (for example `--vault-password-file pwfile`),
	// from ansible documentation this is correct format but that is NOT accepted here.
	// Instead you need to do it like `--vault-password-file=pwfile`.
	//
	// If you are running a Windows build on AWS, Azure, Google Compute, or OpenStack
	// and would like to access the auto-generated password that Packer uses to
	// connect to a Windows instance via WinRM, you can use the template variable
	//
	// ```build.Password``` in HCL templates or ```{{ build `Password`}}``` in
	// legacy JSON templates. For example:
	//
	// in JSON templates:
	//
	// ```json
	// "extra_arguments": [
	//    "--extra-vars", "winrm_password={{ build `Password`}}"
	// ]
	// ```
	//
	// in HCL templates:
	// ```hcl
	// extra_arguments = [
	//    "--extra-vars", "winrm_password=${build.Password}"
	// ]
	// ```
	ExtraArguments []string `mapstructure:"extra_arguments"`
	// A path to the directory containing ansible group
	// variables on your local system to be copied to the remote machine. By
	// default, this is empty.
	GroupVars string `mapstructure:"group_vars"`
	// A path to the directory containing ansible host variables on your local
	// system to be copied to the remote machine. By default, this is empty.
	HostVars string `mapstructure:"host_vars"`
	// A path to the complete ansible directory structure on your local system
	// to be copied to the remote machine as the `staging_directory` before all
	// other files and directories.
	PlaybookDir string `mapstructure:"playbook_dir"`
	// The playbook file to be executed by ansible. This file must exist on your
	// local system and will be uploaded to the remote machine. This option is
	// exclusive with `playbook_files`.
	PlaybookFile string `mapstructure:"playbook_file"`
	// The playbook files to be executed by ansible. These files must exist on
	// your local system. If the files don't exist in the `playbook_dir` or you
	// don't set `playbook_dir` they will be uploaded to the remote machine. This
	// option is exclusive with `playbook_file`.
	PlaybookFiles []string `mapstructure:"playbook_files"`
	// An array of directories of playbook files on your local system. These
	// will be uploaded to the remote machine under `staging_directory`/playbooks.
	// By default, this is empty.
	PlaybookPaths []string `mapstructure:"playbook_paths"`
	// An array of paths to role directories on your local system. These will be
	// uploaded to the remote machine under `staging_directory`/roles. By default,
	// this is empty.
	RolePaths []string `mapstructure:"role_paths"`

	// An array of local paths of collections to upload.
	CollectionPaths []string `mapstructure:"collection_paths"`

	// The directory where files will be uploaded. Packer requires write
	// permissions in this directory.
	StagingDir string `mapstructure:"staging_directory"`
	// If set to `true`, the content of the `staging_directory` will be removed after
	// executing ansible. By default this is set to `false`.
	CleanStagingDir bool `mapstructure:"clean_staging_directory"`
	// The inventory file to be used by ansible. This
	// file must exist on your local system and will be uploaded to the remote
	// machine.
	//
	// When using an inventory file, it's also required to `--limit` the hosts to the
	// specified host you're building. The `--limit` argument can be provided in the
	// `extra_arguments` option.
	//
	// An example inventory file may look like:
	//
	// ```text
	// [chi-dbservers]
	// db-01 ansible_connection=local
	// db-02 ansible_connection=local
	//
	// [chi-appservers]
	// app-01 ansible_connection=local
	// app-02 ansible_connection=local
	//
	// [chi:children]
	// chi-dbservers
	// chi-appservers
	//
	// [dbservers:children]
	// chi-dbservers
	//
	// [appservers:children]
	// chi-appservers
	// ```
	InventoryFile string `mapstructure:"inventory_file"`
	// `inventory_groups` (string) - A comma-separated list of groups to which
	// packer will assign the host `127.0.0.1`. A value of `my_group_1,my_group_2`
	// will generate an Ansible inventory like:
	//
	// ```text
	// [my_group_1]
	// 127.0.0.1
	// [my_group_2]
	// 127.0.0.1
	// ```
	InventoryGroups []string `mapstructure:"inventory_groups"`
	// A requirements file which provides a way to
	//  install roles or collections with the [ansible-galaxy
	//  cli](https://docs.ansible.com/ansible/latest/galaxy/user_guide.html#the-ansible-galaxy-command-line-tool)
	//  on the local machine before executing `ansible-playbook`. By default, this is empty.
	GalaxyFile string `mapstructure:"galaxy_file"`
	// The command to invoke ansible-galaxy. By default, this is
	// `ansible-galaxy`.
	GalaxyCommand string `mapstructure:"galaxy_command"`
	// Maximum time to wait for ansible-navigator version check to complete.
	// Defaults to "60s". This prevents indefinite hangs when ansible-navigator
	// is not properly configured or cannot be found.
	// Format: duration string (e.g., "30s", "1m", "90s").
	// Set to "0" to disable timeout (not recommended).
	VersionCheckTimeout string `mapstructure:"version_check_timeout"`

	// Force overwriting an existing role.
	//  Adds `--force` option to `ansible-galaxy` command. By default, this is
	//  `false`.
	GalaxyForceInstall bool `mapstructure:"galaxy_force_install"`

	// The path to the directory on the remote system in which to
	//   install the roles. Adds `--roles-path /path/to/your/roles` to
	//   `ansible-galaxy` command. By default, this will install to a 'galaxy_roles' subfolder in the
	//   staging/roles directory.
	GalaxyRolesPath string `mapstructure:"galaxy_roles_path"`

	// The path to the directory on the remote system in which to
	//   install the collections. Adds `--collections-path /path/to/your/collections` to
	//   `ansible-galaxy` command. By default, this will install to a 'galaxy_collections' subfolder in the
	//   staging/collections directory.
	GalaxyCollectionsPath string `mapstructure:"galaxy_collections_path"`

	// Modern Ansible Navigator fields (aligned with remote provisioner)

	// Execution mode for ansible-navigator. Valid values: stdout, json, yaml, interactive.
	// Defaults to "stdout" for non-interactive environments (Packer-safe).
	// When set to "interactive" without a TTY, it automatically switches to "stdout".
	NavigatorMode string `mapstructure:"navigator_mode"`
	// The container image to use as the execution environment for ansible-navigator.
	// Specifies which containerized environment runs the Ansible playbooks.
	// When unset, ansible-navigator uses its default execution environment.
	// Examples: "quay.io/ansible/creator-ee:latest", "my-registry.io/custom-ee:v1.0"
	ExecutionEnvironment string `mapstructure:"execution_environment"`
	// Working directory for ansible-navigator execution.
	// When specified, ansible-navigator will be executed from this directory.
	// Defaults to the current working directory if not set.
	WorkDir string `mapstructure:"work_dir"`
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

	// Collections support (aligned with remote provisioner)

	// List of Ansible collections to install automatically.
	// Each entry can be a collection name with optional version (e.g., "community.general:5.11.0")
	// or a local path (e.g., "myorg.mycollection@/path/to/collection").
	Collections []string `mapstructure:"collections"`
	// Directory to cache downloaded collections.
	// Defaults to ~/.packer.d/ansible_collections_cache if not specified.
	CollectionsCacheDir string `mapstructure:"collections_cache_dir"`
	// When true, skip network operations and only use locally cached collections.
	// Fails if a required collection is not present in the cache.
	CollectionsOffline bool `mapstructure:"collections_offline"`
	// When true, always reinstall collections even if they are already cached.
	CollectionsForceUpdate bool `mapstructure:"collections_force_update"`

	// Group management (aligned with remote provisioner)

	// The groups into which the Ansible host should be placed.
	// When unspecified, the host is not associated with any groups.
	// This extends inventory_groups functionality.
	Groups []string `mapstructure:"groups"`
	// Directory to cache downloaded roles. Similar to collections_cache_dir but for roles.
	// Defaults to ~/.packer.d/ansible_roles_cache if not specified.
	RolesCacheDir string `mapstructure:"roles_cache_dir"`
	// When true, skip network operations for both collections and roles.
	// Uses only locally cached dependencies.
	OfflineMode bool `mapstructure:"offline_mode"`
	// When true, always reinstall both collections and roles even if cached.
	ForceUpdate bool `mapstructure:"force_update"`
	//  A map of Ansible configuration settings organized by INI sections.
	// This automatically generates a temporary ansible.cfg file before provisioning begins.
	// The file is uploaded to the remote machine and its path is passed via ANSIBLE_CONFIG.
	//
	// When execution_environment is set but ansible_cfg is not explicitly configured,
	// the plugin automatically applies defaults to fix "Permission denied: /.ansible" errors:
	//   ansible_cfg = {
	//     defaults = {
	//       remote_tmp = "/tmp/.ansible/tmp"
	//       local_tmp  = "/tmp/.ansible-local"
	//     }
	//   }
	//
	// Example:
	//   ansible_cfg = {
	//     defaults = {
	//       remote_tmp      = "/tmp/.ansible/tmp"
	//       host_key_checking = "False"
	//     }
	//     ssh_connection = {
	//       pipelining = "True"
	//     }
	//   }
	AnsibleCfg map[string]map[string]string `mapstructure:"ansible_cfg"`
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

	// Validate navigator mode
	validModes := map[string]bool{
		"stdout":      true,
		"json":        true,
		"yaml":        true,
		"interactive": true,
	}
	if c.NavigatorMode != "" && !validModes[c.NavigatorMode] {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
			"invalid navigator_mode: %s (must be one of stdout, json, yaml, interactive)",
			c.NavigatorMode))
	}

	// Validate playbook_file vs plays mutual exclusivity
	hasPlaybookFile := c.PlaybookFile != "" || len(c.PlaybookFiles) > 0
	hasPlays := len(c.Plays) > 0

	if hasPlaybookFile && hasPlays {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
			"you may specify only one of `playbook_file`/`playbook_files` or `play` blocks"))
	}

	if !hasPlaybookFile && !hasPlays {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
			"either `playbook_file`/`playbook_files` or `play` blocks must be defined"))
	}

	// Validate playbook files if specified
	if c.PlaybookFile != "" {
		if err := validateFileConfig(c.PlaybookFile, "playbook_file", true); err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	for i, playbookFile := range c.PlaybookFiles {
		if err := validateFileConfig(playbookFile, fmt.Sprintf("playbook_files[%d]", i), true); err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
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
	if c.GalaxyFile != "" {
		if err := validateFileConfig(c.GalaxyFile, "galaxy_file", true); err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	// Validate requirements_file
	if c.RequirementsFile != "" {
		if err := validateFileConfig(c.RequirementsFile, "requirements_file", true); err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	// Validate inventory_file
	if c.InventoryFile != "" {
		if err := validateFileConfig(c.InventoryFile, "inventory_file", true); err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	// Validate directories
	if c.PlaybookDir != "" {
		if err := validateDirConfig(c.PlaybookDir, "playbook_dir"); err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	if c.GroupVars != "" {
		if err := validateDirConfig(c.GroupVars, "group_vars"); err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	if c.HostVars != "" {
		if err := validateDirConfig(c.HostVars, "host_vars"); err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	for i, path := range c.PlaybookPaths {
		if err := validateDirConfig(path, fmt.Sprintf("playbook_paths[%d]", i)); err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	for i, path := range c.RolePaths {
		if err := validateDirConfig(path, fmt.Sprintf("role_paths[%d]", i)); err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	for i, path := range c.CollectionPaths {
		if err := validateDirConfig(path, fmt.Sprintf("collection_paths[%d]", i)); err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	// Validate version_check_timeout format
	if c.VersionCheckTimeout != "" {
		if _, err := time.ParseDuration(c.VersionCheckTimeout); err != nil {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
				"invalid version_check_timeout: %q (must be a valid duration like '30s', '1m', '90s'): %w",
				c.VersionCheckTimeout, err))
		}
	}

	// Validate ansible_cfg
	if c.AnsibleCfg != nil && len(c.AnsibleCfg) == 0 {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
			"ansible_cfg cannot be an empty map. Either provide configuration sections or omit the field"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}
	return nil
}

type Provisioner struct {
	config Config

	playbookFiles []string
	generatedData map[string]interface{}
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

	// Reset the state.
	p.playbookFiles = make([]string, 0, len(p.config.PlaybookFiles))

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
	p.config.PlaybookFile = expandUserPath(p.config.PlaybookFile)
	p.config.InventoryFile = expandUserPath(p.config.InventoryFile)
	p.config.GalaxyFile = expandUserPath(p.config.GalaxyFile)
	p.config.RequirementsFile = expandUserPath(p.config.RequirementsFile)
	p.config.PlaybookDir = expandUserPath(p.config.PlaybookDir)
	p.config.GroupVars = expandUserPath(p.config.GroupVars)
	p.config.HostVars = expandUserPath(p.config.HostVars)
	p.config.CollectionsCacheDir = expandUserPath(p.config.CollectionsCacheDir)
	p.config.RolesCacheDir = expandUserPath(p.config.RolesCacheDir)
	p.config.WorkDir = expandUserPath(p.config.WorkDir)

	// Apply HOME expansion to multi-path fields
	for i := range p.config.PlaybookFiles {
		p.config.PlaybookFiles[i] = expandUserPath(p.config.PlaybookFiles[i])
	}
	for i := range p.config.PlaybookPaths {
		p.config.PlaybookPaths[i] = expandUserPath(p.config.PlaybookPaths[i])
	}
	for i := range p.config.RolePaths {
		p.config.RolePaths[i] = expandUserPath(p.config.RolePaths[i])
	}
	for i := range p.config.CollectionPaths {
		p.config.CollectionPaths[i] = expandUserPath(p.config.CollectionPaths[i])
	}

	// Apply HOME expansion to plays
	for i := range p.config.Plays {
		p.config.Plays[i].Target = expandUserPath(p.config.Plays[i].Target)
		for j := range p.config.Plays[i].VarsFiles {
			p.config.Plays[i].VarsFiles[j] = expandUserPath(p.config.Plays[i].VarsFiles[j])
		}
	}
	if p.config.GalaxyCommand == "" {
		p.config.GalaxyCommand = "ansible-galaxy"
	}

	// Set default navigator_mode to stdout for non-interactive environments
	if p.config.NavigatorMode == "" {
		p.config.NavigatorMode = "stdout"
	}

	// Apply default ansible.cfg when execution_environment is set but ansible_cfg is not
	if p.config.ExecutionEnvironment != "" && p.config.AnsibleCfg == nil {
		log.Println("[INFO] ExecutionEnvironment is set but ansible_cfg is not. Applying defaults for container compatibility.")
		p.config.AnsibleCfg = map[string]map[string]string{
			"defaults": {
				"remote_tmp": "/tmp/.ansible/tmp",
				"local_tmp":  "/tmp/.ansible-local",
			},
		}
	}

	// Set default version check timeout
	if p.config.VersionCheckTimeout == "" {
		p.config.VersionCheckTimeout = "60s"
	}

	if p.config.StagingDir == "" {
		p.config.StagingDir = filepath.ToSlash(filepath.Join(DefaultStagingDir, uuid.TimeOrderedUUID()))
	}

	if p.config.GalaxyRolesPath == "" {
		p.config.GalaxyRolesPath = filepath.ToSlash(filepath.Join(p.config.StagingDir, "galaxy_roles"))
	}

	if p.config.GalaxyCollectionsPath == "" {
		p.config.GalaxyCollectionsPath = filepath.ToSlash(filepath.Join(p.config.StagingDir, "galaxy_collections"))
	}

	// Set default cache directories if not specified
	if p.config.CollectionsCacheDir == "" {
		usr, err := os.UserHomeDir()
		if err == nil {
			p.config.CollectionsCacheDir = filepath.Join(usr, ".packer.d", "ansible_collections_cache")
		}
	}

	if p.config.RolesCacheDir == "" {
		usr, err := os.UserHomeDir()
		if err == nil {
			p.config.RolesCacheDir = filepath.Join(usr, ".packer.d", "ansible_roles_cache")
		}
	}

	// Build absolute paths for playbook_files
	for _, playbookFile := range p.config.PlaybookFiles {
		absPath, err := filepath.Abs(playbookFile)
		if err != nil {
			return fmt.Errorf("failed to resolve playbook_files path: %w", err)
		}
		p.playbookFiles = append(p.playbookFiles, absPath)
	}

	// Call comprehensive validation
	if err := p.config.Validate(); err != nil {
		return err
	}

	return nil
}

func (p *Provisioner) Provision(ctx context.Context, ui packersdk.Ui, comm packersdk.Communicator, generatedData map[string]interface{}) error {
	ui.Say("Provisioning with Ansible...")
	p.generatedData = generatedData

	if len(p.config.PlaybookDir) > 0 {
		ui.Message("Uploading Playbook directory to Ansible staging directory...")
		if err := p.uploadDir(ui, comm, p.config.StagingDir, p.config.PlaybookDir); err != nil {
			return fmt.Errorf("Error uploading playbook_dir directory: %s", err)
		}
	} else {
		ui.Message("Creating Ansible staging directory...")
		if err := p.createDir(ui, comm, p.config.StagingDir); err != nil {
			return fmt.Errorf("Error creating staging directory: %s", err)
		}
	}

	if p.config.PlaybookFile != "" {
		ui.Message("Uploading main Playbook file...")
		src := p.config.PlaybookFile
		dst := filepath.ToSlash(filepath.Join(p.config.StagingDir, filepath.Base(src)))
		if err := p.uploadFile(ui, comm, dst, src); err != nil {
			return fmt.Errorf("Error uploading main playbook: %s", err)
		}
	} else if err := p.provisionPlaybookFiles(ui, comm); err != nil {
		return err
	}

	if len(p.config.InventoryFile) == 0 {
		tf, err := tmp.File("packer-provisioner-ansible-local")
		if err != nil {
			return fmt.Errorf("Error preparing inventory file: %s", err)
		}
		defer os.Remove(tf.Name())

		// Support both legacy inventory_groups and modern groups
		groups := p.config.Groups
		if len(groups) == 0 {
			groups = p.config.InventoryGroups
		}

		if len(groups) != 0 {
			content := ""
			for _, group := range groups {
				content += fmt.Sprintf("[%s]\n127.0.0.1\n", group)
			}
			_, err = tf.Write([]byte(content))
		} else {
			_, err = tf.Write([]byte("127.0.0.1"))
		}
		if err != nil {
			tf.Close()
			return fmt.Errorf("Error preparing inventory file: %s", err)
		}
		tf.Close()
		p.config.InventoryFile = tf.Name()
		defer func() {
			p.config.InventoryFile = ""
		}()
	}

	if len(p.config.GalaxyFile) > 0 {
		ui.Message("Uploading galaxy file...")
		src := p.config.GalaxyFile
		dst := filepath.ToSlash(filepath.Join(p.config.StagingDir, filepath.Base(src)))
		if err := p.uploadFile(ui, comm, dst, src); err != nil {
			return fmt.Errorf("Error uploading galaxy file: %s", err)
		}
	}

	ui.Message("Uploading inventory file...")
	src := p.config.InventoryFile
	dst := filepath.ToSlash(filepath.Join(p.config.StagingDir, filepath.Base(src)))
	if err := p.uploadFile(ui, comm, dst, src); err != nil {
		return fmt.Errorf("Error uploading inventory file: %s", err)
	}

	// Generate and upload ansible.cfg if configured
	var ansibleCfgRemotePath string
	if p.config.AnsibleCfg != nil {
		content, err := generateAnsibleCfg(p.config.AnsibleCfg)
		if err != nil {
			return fmt.Errorf("Error generating ansible.cfg: %s", err)
		}

		// Create temporary local file
		tmpPath, err := createTempAnsibleCfg(content)
		if err != nil {
			return fmt.Errorf("Error creating temporary ansible.cfg: %s", err)
		}
		defer os.Remove(tmpPath)

		ui.Message("Uploading generated ansible.cfg...")
		// Upload to remote staging directory
		ansibleCfgRemotePath = filepath.ToSlash(filepath.Join(p.config.StagingDir, "ansible.cfg"))
		if err := p.uploadFile(ui, comm, ansibleCfgRemotePath, tmpPath); err != nil {
			return fmt.Errorf("Error uploading ansible.cfg: %s", err)
		}
	}

	if len(p.config.GroupVars) > 0 {
		ui.Message("Uploading group_vars directory...")
		src := p.config.GroupVars
		dst := filepath.ToSlash(filepath.Join(p.config.StagingDir, "group_vars"))
		if err := p.uploadDir(ui, comm, dst, src); err != nil {
			return fmt.Errorf("Error uploading group_vars directory: %s", err)
		}
	}

	if len(p.config.HostVars) > 0 {
		ui.Message("Uploading host_vars directory...")
		src := p.config.HostVars
		dst := filepath.ToSlash(filepath.Join(p.config.StagingDir, "host_vars"))
		if err := p.uploadDir(ui, comm, dst, src); err != nil {
			return fmt.Errorf("Error uploading host_vars directory: %s", err)
		}
	}

	if len(p.config.RolePaths) > 0 {
		ui.Message("Uploading role directories...")
		for _, src := range p.config.RolePaths {
			dst := filepath.ToSlash(filepath.Join(p.config.StagingDir, "roles", filepath.Base(src)))
			if err := p.uploadDir(ui, comm, dst, src); err != nil {
				return fmt.Errorf("Error uploading roles: %s", err)
			}
		}
	}

	if len(p.config.CollectionPaths) > 0 {
		ui.Message("Uploading collection directories...")
		for _, src := range p.config.CollectionPaths {
			dst := filepath.ToSlash(filepath.Join(p.config.StagingDir, "collections", filepath.Base(src)))
			if err := p.uploadDir(ui, comm, dst, src); err != nil {
				return fmt.Errorf("Error uploading collections: %s", err)
			}
		}
	}

	if len(p.config.PlaybookPaths) > 0 {
		ui.Message("Uploading additional Playbooks...")
		playbookDir := filepath.ToSlash(filepath.Join(p.config.StagingDir, "playbooks"))
		if err := p.createDir(ui, comm, playbookDir); err != nil {
			return fmt.Errorf("Error creating playbooks directory: %s", err)
		}
		for _, src := range p.config.PlaybookPaths {
			dst := filepath.ToSlash(filepath.Join(playbookDir, filepath.Base(src)))
			if err := p.uploadDir(ui, comm, dst, src); err != nil {
				return fmt.Errorf("Error uploading playbooks: %s", err)
			}
		}
	}

	// Install dependencies using GalaxyManager
	galaxyManager := NewGalaxyManager(&p.config, ui, comm)
	if err := galaxyManager.InstallRequirements(); err != nil {
		return fmt.Errorf("Error installing requirements: %s", err)
	}

	if err := p.executeAnsible(ui, comm, ansibleCfgRemotePath); err != nil {
		return fmt.Errorf("Error executing Ansible Navigator: %s", err)
	}

	if p.config.CleanStagingDir {
		ui.Message("Removing staging directory...")
		if err := p.removeDir(ui, comm, p.config.StagingDir); err != nil {
			return fmt.Errorf("Error removing staging directory: %s", err)
		}
	}
	return nil
}

func (p *Provisioner) provisionPlaybookFiles(ui packersdk.Ui, comm packersdk.Communicator) error {
	var playbookDir string
	if p.config.PlaybookDir != "" {
		var err error
		playbookDir, err = filepath.Abs(p.config.PlaybookDir)
		if err != nil {
			return err
		}
	}
	for index, playbookFile := range p.playbookFiles {
		if playbookDir != "" && strings.HasPrefix(playbookFile, playbookDir) {
			p.playbookFiles[index] = strings.TrimPrefix(playbookFile, playbookDir)
			continue
		}
		if err := p.provisionPlaybookFile(ui, comm, playbookFile); err != nil {
			return err
		}
	}
	return nil
}

func (p *Provisioner) provisionPlaybookFile(ui packersdk.Ui, comm packersdk.Communicator, playbookFile string) error {
	ui.Message(fmt.Sprintf("Uploading playbook file: %s", playbookFile))

	remoteDir := filepath.ToSlash(filepath.Join(p.config.StagingDir, filepath.Dir(playbookFile)))
	remotePlaybookFile := filepath.ToSlash(filepath.Join(p.config.StagingDir, playbookFile))

	if err := p.createDir(ui, comm, remoteDir); err != nil {
		return fmt.Errorf("Error uploading playbook file: %s [%s]", playbookFile, err)
	}

	if err := p.uploadFile(ui, comm, remotePlaybookFile, playbookFile); err != nil {
		return fmt.Errorf("Error uploading playbook: %s [%s]", playbookFile, err)
	}

	return nil
}

func (p *Provisioner) executeGalaxy(ui packersdk.Ui, comm packersdk.Communicator) error {
	galaxyFile := filepath.ToSlash(filepath.Join(p.config.StagingDir, filepath.Base(p.config.GalaxyFile)))

	// ansible-galaxy install -r requirements.yml
	roleArgs := []string{"install", "-r", galaxyFile, "-p", filepath.ToSlash(p.config.GalaxyRolesPath)}

	// Instead of modifying args depending on config values and removing or modifying values from
	// the slice between role and collection installs, just use 2 slices and simplify everything
	collectionArgs := []string{"collection", "install", "-r", galaxyFile, "-p", filepath.ToSlash(p.config.GalaxyCollectionsPath)}

	// Add force to arguments
	if p.config.GalaxyForceInstall {
		roleArgs = append(roleArgs, "-f")
		collectionArgs = append(collectionArgs, "-f")
	}

	// Search galaxy_file for roles and collections keywords
	f, err := os.ReadFile(p.config.GalaxyFile)
	if err != nil {
		return err
	}
	hasRoles, _ := regexp.Match(`(?m)^roles:`, f)
	hasCollections, _ := regexp.Match(`(?m)^collections:`, f)

	// If if roles keyword present (v2 format), or no collections keyword present (v1), install roles
	if hasRoles || !hasCollections {
		if roleInstallError := p.invokeGalaxyCommand(roleArgs, ui, comm); roleInstallError != nil {
			return roleInstallError
		}
	}

	// If collections keyword present (v2 format), install collections
	if hasCollections {
		if collectionInstallError := p.invokeGalaxyCommand(collectionArgs, ui, comm); collectionInstallError != nil {
			return collectionInstallError
		}
	}

	return nil
}

// Intended to be invoked from p.executeGalaxy depending on the Ansible Galaxy parameters passed to Packer
func (p *Provisioner) invokeGalaxyCommand(args []string, ui packersdk.Ui, comm packersdk.Communicator) error {
	ctx := context.TODO()
	command := fmt.Sprintf("cd %s && %s %s",
		p.config.StagingDir, p.config.GalaxyCommand, strings.Join(args, " "))
	ui.Message(fmt.Sprintf("Executing Ansible Galaxy: %s", command))

	cmd := &packersdk.RemoteCmd{
		Command: command,
	}
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return err
	}
	if cmd.ExitStatus() != 0 {
		// ansible-galaxy version 2.0.0.2 doesn't return exit codes on error..
		return fmt.Errorf("Non-zero exit status: %d", cmd.ExitStatus())
	}
	return nil
}

func (p *Provisioner) executeAnsible(ui packersdk.Ui, comm packersdk.Communicator, ansibleCfgRemotePath string) error {
	inventory := filepath.ToSlash(filepath.Join(p.config.StagingDir, filepath.Base(p.config.InventoryFile)))

	if len(p.config.Plays) > 0 {
		// Execute multiple plays
		return p.executePlays(ui, comm, inventory, ansibleCfgRemotePath)
	} else {
		// Execute traditional playbook(s) for backward compatibility
		return p.executeTraditionalPlaybooks(ui, comm, inventory, ansibleCfgRemotePath)
	}
}

// executePlays executes multiple Ansible plays in sequence
func (p *Provisioner) executePlays(ui packersdk.Ui, comm packersdk.Communicator, inventory string, ansibleCfgRemotePath string) error {
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
			ui.Message(fmt.Sprintf("Uploading playbook: %s", play.Target))
			remotePath := filepath.ToSlash(filepath.Join(p.config.StagingDir, filepath.Base(play.Target)))
			if err := p.uploadFile(ui, comm, remotePath, play.Target); err != nil {
				return fmt.Errorf("Play '%s': failed to upload playbook: %s", playName, err)
			}
			playbookPath = remotePath
		} else {
			// It's a role FQDN - generate temporary playbook locally then upload
			ui.Message(fmt.Sprintf("Generating temporary playbook for role: %s", play.Target))
			tmpPlaybook, err := p.createRolePlaybook(play.Target, play)
			if err != nil {
				return fmt.Errorf("play %q: failed to generate role playbook: %w", playName, err)
			}

			// Upload generated playbook to remote
			remotePath := filepath.ToSlash(filepath.Join(p.config.StagingDir, filepath.Base(tmpPlaybook)))
			if err := p.uploadFile(ui, comm, remotePath, tmpPlaybook); err != nil {
				os.Remove(tmpPlaybook)
				return fmt.Errorf("Play '%s': failed to upload generated playbook: %s", playName, err)
			}
			playbookPath = remotePath
			cleanupFunc = func() {
				os.Remove(tmpPlaybook)
			}
		}

		// Build extra arguments for this play
		extraArgs := p.buildExtraArgs(play)

		// Execute the play
		err := p.executeAnsiblePlaybook(ui, comm, playbookPath, extraArgs, inventory, ansibleCfgRemotePath)

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

// executeTraditionalPlaybooks executes traditional playbook files for backward compatibility
func (p *Provisioner) executeTraditionalPlaybooks(ui packersdk.Ui, comm packersdk.Communicator, inventory string, ansibleCfgRemotePath string) error {
	extraArgs := fmt.Sprintf(" --extra-vars \"packer_build_name=%s packer_builder_type=%s packer_http_addr=%s -o IdentitiesOnly=yes\" ",
		p.config.PackerBuildName, p.config.PackerBuilderType, p.generatedData["PackerHTTPAddr"])
	if len(p.config.ExtraArguments) > 0 {
		extraArgs = extraArgs + strings.Join(p.config.ExtraArguments, " ")
	}

	if p.config.PlaybookFile != "" {
		playbookFile := filepath.ToSlash(filepath.Join(p.config.StagingDir, filepath.Base(p.config.PlaybookFile)))
		if err := p.executeAnsiblePlaybook(ui, comm, playbookFile, extraArgs, inventory, ansibleCfgRemotePath); err != nil {
			return err
		}
	}

	for _, playbookFile := range p.playbookFiles {
		playbookFile = filepath.ToSlash(filepath.Join(p.config.StagingDir, playbookFile))
		if err := p.executeAnsiblePlaybook(ui, comm, playbookFile, extraArgs, inventory, ansibleCfgRemotePath); err != nil {
			return err
		}
	}
	return nil
}

// buildExtraArgs constructs extra arguments for a specific play
func (p *Provisioner) buildExtraArgs(play Play) string {
	extraArgs := fmt.Sprintf(" --extra-vars \"packer_build_name=%s packer_builder_type=%s packer_http_addr=%s -o IdentitiesOnly=yes\" ",
		p.config.PackerBuildName, p.config.PackerBuilderType, p.generatedData["PackerHTTPAddr"])

	// Add global extra arguments
	if len(p.config.ExtraArguments) > 0 {
		extraArgs = extraArgs + strings.Join(p.config.ExtraArguments, " ")
	}

	// Add per-play arguments
	if play.Become {
		extraArgs = extraArgs + " --become"
	}

	if len(play.Tags) > 0 {
		extraArgs = extraArgs + " --tags " + strings.Join(play.Tags, ",")
	}

	for k, v := range play.ExtraVars {
		extraArgs = extraArgs + fmt.Sprintf(" -e %s=%s", k, v)
	}

	for _, varsFile := range play.VarsFiles {
		extraArgs = extraArgs + fmt.Sprintf(" -e @%s", varsFile)
	}

	return extraArgs
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
	ui packersdk.Ui, comm packersdk.Communicator, playbookFile, extraArgs, inventory string, ansibleCfgRemotePath string,
) error {
	ctx := context.TODO()
	env_vars := ""

	// Build environment variables for collections and roles paths
	galaxyManager := NewGalaxyManager(&p.config, ui, comm)
	envPaths := galaxyManager.SetupEnvironmentPaths()
	if len(envPaths) > 0 {
		env_vars = strings.Join(envPaths, " ") + " "
	}

	// Add ANSIBLE_CONFIG if provided
	if ansibleCfgRemotePath != "" {
		env_vars += fmt.Sprintf("ANSIBLE_CONFIG=%s ", ansibleCfgRemotePath)
	}

	// Add standard Ansible environment variables
	env_vars += "ANSIBLE_FORCE_COLOR=1 PYTHONUNBUFFERED=1"

	// Build PATH override if ansible_navigator_path is set
	pathPrefix := ""
	if len(p.config.AnsibleNavigatorPath) > 0 {
		pathPrefix = buildPathPrefixForRemoteShell(p.config.AnsibleNavigatorPath) + " "
	}

	// Build navigator-specific flags
	navigatorFlags := ""
	if p.config.NavigatorMode != "" {
		navigatorFlags += fmt.Sprintf(" --mode %s", p.config.NavigatorMode)
	}
	// Add execution environment flags if set (ansible-navigator v3+ format)
	if p.config.ExecutionEnvironment != "" {
		navigatorFlags += fmt.Sprintf(" --ee true --eei %s", p.config.ExecutionEnvironment)
	}

	// Command now defaults to just "ansible-navigator", so we need to add "run" as first arg
	command := fmt.Sprintf("cd %s && %s%s %s run%s %s%s -c local -i %s",
		p.config.StagingDir, pathPrefix, env_vars, p.config.Command, navigatorFlags, playbookFile, extraArgs, inventory,
	)
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

// generateAnsibleCfg generates INI-formatted content from a map of sections
// Returns the formatted string and any validation errors
func generateAnsibleCfg(sections map[string]map[string]string) (string, error) {
	if len(sections) == 0 {
		return "", fmt.Errorf("ansible_cfg cannot be empty")
	}

	var buf strings.Builder
	sectionNames := make([]string, 0, len(sections))
	for section := range sections {
		sectionNames = append(sectionNames, section)
	}
	sort.Strings(sectionNames)

	for _, section := range sectionNames {
		settings := sections[section]
		// Write section header
		buf.WriteString(fmt.Sprintf("[%s]\n", section))

		// Write settings
		keys := make([]string, 0, len(settings))
		for key := range settings {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			buf.WriteString(fmt.Sprintf("%s = %s\n", key, settings[key]))
		}

		// Add blank line between sections
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

// createTempAnsibleCfg writes ansible.cfg content to a temporary file
// Returns the absolute path to the file
func createTempAnsibleCfg(content string) (string, error) {
	tmpFile, err := tmp.File("packer-ansible-cfg-*.ini")
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

	// Return absolute path
	absPath, err := filepath.Abs(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to get absolute path for ansible.cfg: %w", err)
	}

	return absPath, nil
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
