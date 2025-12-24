// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

//go:generate packer-sdc mapstructure-to-hcl2 -type Config,Play,PathEntry,NavigatorConfig,ExecutionEnvironment,EnvironmentVariablesConfig,VolumeMount,AnsibleConfig,AnsibleConfigDefaults,AnsibleConfigConnection,AnsibleConfigPrivilegeEscalation,AnsibleConfigPersistentConnection,AnsibleConfigInventory,AnsibleConfigParamikoConnection,AnsibleConfigColors,AnsibleConfigDiff,AnsibleConfigGalaxy,LoggingConfig,PlaybookArtifact,CollectionDocCache,ColorConfig,EditorConfig,ImagesConfig,BastionConfig
//go:generate packer-sdc struct-markdown

package ansiblenavigator

import (
	"bufio"
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"golang.org/x/crypto/ssh"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/adapter"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer-plugin-sdk/tmp"
)

// Compile-time interface checks
var (
	_ packersdk.Provisioner = &Provisioner{}
)

// NavigatorConfig represents the ansible-navigator.yml configuration structure.
// This is the root configuration container for ansible-navigator settings.
type NavigatorConfig struct {
	// Ansible-navigator execution mode
	Mode string `mapstructure:"mode"`
	// Stdout output format
	Format string `mapstructure:"format"`
	// Color scheme settings
	Color *ColorConfig `mapstructure:"color"`
	// Editor settings
	Editor *EditorConfig `mapstructure:"editor"`
	// Image display settings
	Images *ImagesConfig `mapstructure:"images"`
	// Time zone
	TimeZone string `mapstructure:"time_zone"`
	// Inventory display columns
	InventoryColumns []string `mapstructure:"inventory_columns"`
	// Collection documentation cache path
	CollectionDocCachePath string `mapstructure:"collection_doc_cache_path"`
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
	// Container engine to use (auto, podman, docker)
	ContainerEngine string `mapstructure:"container_engine"`
	// Additional container runtime options (e.g., --net=host, --security-opt)
	ContainerOptions []string `mapstructure:"container_options"`
	// Arguments passed to image pull command
	PullArguments []string `mapstructure:"pull_arguments"`
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
	// Privilege escalation (become) settings
	PrivilegeEscalation *AnsibleConfigPrivilegeEscalation `mapstructure:"privilege_escalation"`
	// Persistent connection settings
	PersistentConnection *AnsibleConfigPersistentConnection `mapstructure:"persistent_connection"`
	// Inventory behavior settings
	Inventory *AnsibleConfigInventory `mapstructure:"inventory"`
	// Paramiko connection settings
	ParamikoConnection *AnsibleConfigParamikoConnection `mapstructure:"paramiko_connection"`
	// Output color settings
	Colors *AnsibleConfigColors `mapstructure:"colors"`
	// Diff display settings
	Diff *AnsibleConfigDiff `mapstructure:"diff"`
	// Galaxy client settings (ansible.cfg [galaxy] section)
	Galaxy *AnsibleConfigGalaxy `mapstructure:"galaxy"`
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

// AnsibleConfigPrivilegeEscalation represents ansible.cfg [privilege_escalation]
// settings. These provide defaults for privilege escalation behavior.
type AnsibleConfigPrivilegeEscalation struct {
	// Enable privilege escalation (become)
	Become bool `mapstructure:"become"`
	// Privilege escalation method (e.g. sudo, su, pbrun)
	BecomeMethod string `mapstructure:"become_method"`
	// Privilege escalation user (e.g. root)
	BecomeUser string `mapstructure:"become_user"`
}

// AnsibleConfigPersistentConnection represents ansible.cfg [persistent_connection]
// settings that tune connection persistence behavior.
type AnsibleConfigPersistentConnection struct {
	// Timeout (seconds) for establishing a persistent connection
	ConnectTimeout int `mapstructure:"connect_timeout"`
	// Timeout (seconds) for retries when establishing a connection
	ConnectRetryTimeout int `mapstructure:"connect_retry_timeout"`
	// Timeout (seconds) for remote command execution over the persistent connection
	CommandTimeout int `mapstructure:"command_timeout"`
}

// AnsibleConfigInventory represents ansible.cfg [inventory] settings.
type AnsibleConfigInventory struct {
	// Enable inventory plugins (comma-separated list in INI)
	EnablePlugins []string `mapstructure:"enable_plugins"`
}

// AnsibleConfigParamikoConnection represents ansible.cfg [paramiko_connection]
// settings.
type AnsibleConfigParamikoConnection struct {
	// ProxyCommand for Paramiko connections
	ProxyCommand string `mapstructure:"proxy_command"`
}

// AnsibleConfigColors represents ansible.cfg [colors] settings.
type AnsibleConfigColors struct {
	// Force colored output
	ForceColor bool `mapstructure:"force_color"`
}

// AnsibleConfigDiff represents ansible.cfg [diff] settings.
type AnsibleConfigDiff struct {
	// Always show diffs
	Always bool `mapstructure:"always"`
	// Number of context lines to include in diffs
	Context int `mapstructure:"context"`
}

// AnsibleConfigGalaxy represents ansible.cfg [galaxy] settings.
// NOTE: This is Ansible runtime configuration only; it does not affect the plugin's
// dependency installation behavior.
type AnsibleConfigGalaxy struct {
	// List of Galaxy server names to use (comma-separated list in INI)
	ServerList []string `mapstructure:"server_list"`
	// Ignore TLS certificate validation
	IgnoreCerts bool `mapstructure:"ignore_certs"`
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
	// Path to a playbook artifact to replay.
	// (YAML key: playbook-artifact.replay)
	Replay string `mapstructure:"replay"`
	// Path to write the playbook artifact.
	// (YAML key: playbook-artifact.save-as)
	SaveAs string `mapstructure:"save_as"`
}

// CollectionDocCache represents collection documentation cache settings
type CollectionDocCache struct {
	// Path to collection doc cache
	Path string `mapstructure:"path"`
	// Timeout for collection doc cache
	Timeout int `mapstructure:"timeout"`
}

// ColorConfig represents top-level color configuration in ansible-navigator.yml.
type ColorConfig struct {
	Enable bool `mapstructure:"enable"`
	Osc4   bool `mapstructure:"osc4"`
}

// EditorConfig represents top-level editor configuration in ansible-navigator.yml.
type EditorConfig struct {
	Command string `mapstructure:"command"`
	Console bool   `mapstructure:"console"`
}

// ImagesConfig represents top-level images configuration in ansible-navigator.yml.
type ImagesConfig struct {
	Details []string `mapstructure:"details"`
}

// BastionConfig represents SSH bastion (jump host) configuration for establishing
// SSH tunnels to target machines that are not directly accessible.
// When connection_mode is set to "ssh_tunnel", the bastion host is used as a proxy
// to reach the target machine.
type BastionConfig struct {
	// Enable bastion functionality. When true and bastion.host is set,
	// the plugin establishes an SSH tunnel through the bastion host.
	// Auto-enabled when bastion.host is provided.
	Enabled bool `mapstructure:"enabled"`

	// Bastion (jump host) address for SSH tunneling.
	// Required when connection_mode is "ssh_tunnel".
	// Example: "bastion.example.com"
	Host string `mapstructure:"host"`

	// SSH port on the bastion host.
	// Defaults to 22 if not specified.
	Port int `mapstructure:"port"`

	// SSH username for authenticating to the bastion host.
	// Required when bastion is enabled.
	User string `mapstructure:"user"`

	// Path to the SSH private key file for bastion authentication.
	// Either this or password must be provided when bastion is enabled.
	// Supports HOME expansion (~ and ~/path).
	PrivateKeyFile string `mapstructure:"private_key_file"`

	// Password for bastion authentication.
	// Either this or private_key_file must be provided when bastion is enabled.
	Password string `mapstructure:"password"`
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
	// User to become when using privilege escalation
	BecomeUser string `mapstructure:"become_user"`
	// Tags to skip for this play
	SkipTags []string `mapstructure:"skip_tags"`
	// Extra arguments to pass verbatim to ansible-navigator for this play.
	// These are appended after `ansible-navigator run` (and enforced `--mode`),
	// and before plugin-generated inventory/extra-vars/etc.
	ExtraArgs []string `mapstructure:"extra_args"`
}

// Config holds the configuration for the Ansible Navigator provisioner.
// It supports both traditional playbook-based provisioning and modern
// collection-based workflows with execution environments.
type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	ctx                 interpolate.Context
	// The command to invoke ansible-navigator. Defaults to `ansible-navigator`.
	// This must be the executable name or path only, without additional arguments.
	// Use extra_arguments or play-level options for additional flags.
	// Examples: "ansible-navigator", "/usr/local/bin/ansible-navigator", "~/bin/ansible-navigator"
	Command string `mapstructure:"command"`
	// Additional directories to prepend to PATH when locating and running ansible-navigator.
	// Entries are HOME-expanded and prepended to PATH during version checks and execution.
	// Example: ["~/bin", "/opt/ansible/bin"]
	AnsibleNavigatorPath []string `mapstructure:"ansible_navigator_path"`
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

	// Repeated play blocks supporting both playbooks and role FQDNs
	Plays []Play `mapstructure:"play"`
	// Path to a unified requirements.yml file containing both roles and collections
	RequirementsFile string `mapstructure:"requirements_file"`
	// Destination directory for installed roles.
	// This value is passed to ansible-galaxy as the roles install path and exported to Ansible via ANSIBLE_ROLES_PATH.
	// Defaults to ~/.packer.d/ansible_roles_cache if not specified.
	RolesPath string `mapstructure:"roles_path"`
	// Destination directory for installed collections.
	// This value is passed to ansible-galaxy as the collections install path and exported to Ansible via ANSIBLE_COLLECTIONS_PATH.
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

	// The groups into which the Ansible host should
	//  be placed. When unspecified, the host is not associated with any groups.
	Groups []string `mapstructure:"groups"`
	// The groups which should be present in
	//  inventory file but remain empty.
	EmptyGroups []string `mapstructure:"empty_groups"`
	//  The alias by which the Ansible host should be
	// known. Defaults to `default`. This setting is ignored when using a custom
	// inventory file.
	HostAlias string `mapstructure:"host_alias"`
	// The `ansible_user` to use. Defaults to the user running
	//  packer, NOT the user set for your communicator. If you want to use the same
	//  user as the communicator, you will need to manually set it again in this
	//  field.
	User string `mapstructure:"user"`
	// The port on which to attempt to listen for SSH
	//  connections. This value is a starting point. The provisioner will attempt
	//  listen for SSH connections on the first available of ten ports, starting at
	//  `local_port`. A system-chosen port is used when `local_port` is missing or
	//  empty.
	LocalPort int `mapstructure:"local_port"`
	// The SSH key that will be used to run the SSH
	//  server on the host machine to forward commands to the target machine.
	//  Ansible connects to this server and will validate the identity of the
	//  server using the system known_hosts. The default behavior is to generate
	//  and use a onetime key. Host key checking is disabled via the
	//  `ANSIBLE_HOST_KEY_CHECKING` environment variable if the key is generated.
	SSHHostKeyFile string `mapstructure:"ssh_host_key_file"`
	// The SSH public key of the Ansible
	//  `ssh_user`. The default behavior is to generate and use a onetime key. If
	//  this key is generated, the corresponding private key is passed to
	//  `ansible-playbook` with the `-e ansible_ssh_private_key_file` option.
	SSHAuthorizedKeyFile string `mapstructure:"ssh_authorized_key_file"`
	// Change the key type used for the adapter.
	//
	// Supported values:
	//
	// * ECDSA (default)
	// * RSA
	//
	// NOTE: using RSA may cause problems if the key is used to authenticate with rsa-sha1
	// as modern OpenSSH versions reject this by default as it is unsafe.
	AdapterKeyType string `mapstructure:"ansible_proxy_key_type"`
	// The IP address the SSH proxy should bind to.
	// Defaults to "127.0.0.1". Set to "0.0.0.0" to allow external connections
	// (e.g. from a container).
	AnsibleProxyBindAddress string `mapstructure:"ansible_proxy_bind_address"`
	// The host address the container should use to connect back to the proxy.
	// Defaults to "127.0.0.1". For WSL2/containers, this might be "host.docker.internal"
	// or "host.containers.internal".
	AnsibleProxyHost string `mapstructure:"ansible_proxy_host"`
	// The command to run on the machine being
	//  provisioned by Packer to handle the SFTP protocol that Ansible will use to
	//  transfer files. The command should read and write on stdin and stdout,
	//  respectively. Defaults to `/usr/lib/sftp-server -e`.
	SFTPCmd string `mapstructure:"sftp_command"`
	// Check if ansible-navigator is installed prior to running.
	// Set this to `true`, for example, if you're going to install it during the packer run.
	SkipVersionCheck bool `mapstructure:"skip_version_check"`
	// Maximum time to wait for ansible-navigator version check to complete.
	// Defaults to "60s". This prevents indefinite hangs when ansible-navigator
	// is not properly configured or cannot be found.
	// Format: duration string (e.g., "30s", "1m", "90s").
	// Set to "0" to disable timeout (not recommended).
	VersionCheckTimeout *string `mapstructure:"version_check_timeout"`
	// Internal: tracks whether version_check_timeout was explicitly provided by the user.
	// This allows warning when skip_version_check makes it ineffective.
	versionCheckTimeoutWasSet bool `mapstructure:"-"`
	UseSFTP                   bool `mapstructure:"use_sftp"`
	// The directory in which to place the
	//  temporary generated Ansible inventory file. By default, this is the
	//  system-specific temporary file location. The fully-qualified name of this
	//  temporary file will be passed to the `-i` argument of the `ansible` command
	//  when this provisioner runs ansible. Specify this if you have an existing
	//  inventory directory with `host_vars` `group_vars` that you would like to
	//  use in the playbook that this provisioner will run.
	InventoryDirectory string `mapstructure:"inventory_directory"`
	// This template represents the format for the lines added to the temporary
	// inventory file that Packer will create to run Ansible against your image.
	// The default for recent versions of Ansible is:
	// "{{ .HostAlias }} ansible_host={{ .Host }} ansible_user={{ .User }} ansible_port={{ .Port }}\n"
	// Available template engines are: This option is a template engine;
	// variables available to you include the examples in the default (Host,
	// HostAlias, User, Port) as well as any variables available to you via the
	// "build" template engine.
	InventoryFileTemplate string `mapstructure:"inventory_file_template"`
	// The inventory file to use during provisioning.
	//  When unspecified, Packer will create a temporary inventory file and will
	//  use the `host_alias`.
	InventoryFile string `mapstructure:"inventory_file"`
	// Limit playbook execution to specific hosts or groups.
	// This corresponds to ansible-playbook's --limit flag.
	// Example: "webservers:&production" or "host1,host2"
	Limit string `mapstructure:"limit"`
	// If `true`, the Ansible provisioner will
	//  not delete the temporary inventory file it creates in order to connect to
	//  the instance. This is useful if you are trying to debug your ansible run
	//  and using "--on-error=ask" in order to leave your instance running while you
	//  test your playbook. this option is not used if you set an `inventory_file`.
	KeepInventoryFile bool `mapstructure:"keep_inventory_file"`

	// Force overwriting an existing role/collection and its dependencies.
	//  Adds `--force-with-deps` option to `ansible-galaxy` command. By default,
	//  this is `false`.
	GalaxyForceWithDeps bool `mapstructure:"galaxy_force_with_deps"`

	// ConnectionMode determines how Ansible connects to the target machine.
	//
	// Valid values:
	//   - "proxy" (default): Use Packer's SSH proxy adapter. Works for most builds including Docker.
	//   - "ssh_tunnel": Establish SSH tunnel through a bastion host. Required when targets are only
	//     accessible via jump host (common with WSL2 execution environments).
	//   - "direct": Connect directly to the target without proxy. Use when the target IP is directly
	//     accessible and proxy overhead is unnecessary.
	//
	// When using "ssh_tunnel", you must provide bastion_host, bastion_user, and either
	// bastion_private_key_file or bastion_password.
	ConnectionMode string `mapstructure:"connection_mode"`

	// Force WinRM to use HTTP instead of HTTPS.
	//
	// Set this to true to force Ansible to use HTTP instead of HTTPS to communicate
	// over WinRM to the destination host.
	//
	// Ansible uses the port as a heuristic to determine whether to use HTTP
	// or not. In the current state, Packer assigns a random port for connecting
	// to WinRM and Ansible's heuristic fails to determine that it should be
	// using HTTP, even when the communicator is setup to use it.
	//
	// Alternatively, you may also directly add the following arguments to the
	// `extra_arguments` section for ansible: `"-e", "ansible_winrm_scheme=http"`.
	//
	// Default: `false`
	WinRMUseHTTP bool `mapstructure:"ansible_winrm_use_http"`

	// Display extra vars JSON content in output for debugging.
	// When enabled, logs the extra vars JSON passed to ansible-navigator with sensitive values redacted.
	// Default: false
	ShowExtraVars bool `mapstructure:"show_extra_vars"`

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
	userWasEmpty    bool

	// Bastion configuration for SSH tunnel mode.
	// Use the nested block syntax for new configurations:
	//   bastion {
	//     host = "bastion.example.com"
	//     port = 22
	//     user = "ubuntu"
	//     private_key_file = "~/.ssh/id_rsa"
	//   }
	Bastion *BastionConfig `mapstructure:"bastion"`

	// DEPRECATED: Use bastion.host instead.
	// Bastion (jump host) address for SSH tunneling mode.
	// Required when connection_mode is "ssh_tunnel".
	// Example: "bastion.example.com"
	BastionHost string `mapstructure:"bastion_host"`

	// DEPRECATED: Use bastion.port instead.
	// SSH port on the bastion host. Defaults to 22.
	BastionPort int `mapstructure:"bastion_port"`

	// DEPRECATED: Use bastion.user instead.
	// SSH username for authenticating to the bastion host.
	// Required when ssh_tunnel_mode is true.
	BastionUser string `mapstructure:"bastion_user"`

	// DEPRECATED: Use bastion.private_key_file instead.
	// Path to the SSH private key file for bastion authentication.
	// Either this or bastion_password must be provided when ssh_tunnel_mode is true.
	// Supports HOME expansion (~ and ~/path).
	BastionPrivateKeyFile string `mapstructure:"bastion_private_key_file"`

	// DEPRECATED: Use bastion.password instead.
	// Password for bastion authentication.
	// Either this or bastion_private_key_file must be provided when ssh_tunnel_mode is true.
	BastionPassword string `mapstructure:"bastion_password"`
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

	// Validate connection_mode
	validModes := []string{"proxy", "ssh_tunnel", "direct"}
	if c.ConnectionMode == "" {
		c.ConnectionMode = "proxy" // Apply default
	}
	modeValid := false
	for _, mode := range validModes {
		if c.ConnectionMode == mode {
			modeValid = true
			break
		}
	}
	if !modeValid {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
			"connection_mode must be one of %v, got: %q", validModes, c.ConnectionMode))
	}

	// Validate bastion requirements when using ssh_tunnel mode
	if c.ConnectionMode == "ssh_tunnel" {
		// Check both new bastion block and legacy flat fields for validation
		bastionHost := ""
		bastionUser := ""
		bastionPrivateKeyFile := ""
		bastionPassword := ""
		bastionPort := 0

		if c.Bastion != nil {
			bastionHost = c.Bastion.Host
			bastionUser = c.Bastion.User
			bastionPrivateKeyFile = c.Bastion.PrivateKeyFile
			bastionPassword = c.Bastion.Password
			bastionPort = c.Bastion.Port
		} else {
			// Fall back to legacy flat fields for backward compatibility
			bastionHost = c.BastionHost
			bastionUser = c.BastionUser
			bastionPrivateKeyFile = c.BastionPrivateKeyFile
			bastionPassword = c.BastionPassword
			bastionPort = c.BastionPort
		}

		if bastionHost == "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
				"bastion.host is required when connection_mode='ssh_tunnel'"))
		}
		if bastionUser == "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
				"bastion.user is required when connection_mode='ssh_tunnel'"))
		}
		if bastionPrivateKeyFile == "" && bastionPassword == "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
				"either bastion.private_key_file or bastion.password must be provided when connection_mode='ssh_tunnel'"))
		}
		if bastionPort < 1 || bastionPort > 65535 {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
				"bastion.port must be between 1 and 65535, got %d", bastionPort))
		}
		if bastionPrivateKeyFile != "" {
			if err := validateFileConfig(bastionPrivateKeyFile, "bastion.private_key_file", true); err != nil {
				errs = packersdk.MultiErrorAppend(errs, err)
			}
		}
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

		// Validate playbook files
		if strings.HasSuffix(play.Target, ".yml") || strings.HasSuffix(play.Target, ".yaml") {
			if err := validateFileConfig(play.Target, fmt.Sprintf("play %d target", i), true); err != nil {
				errs = packersdk.MultiErrorAppend(errs, err)
			}
		}
	}

	// Validate files
	if c.RequirementsFile != "" {
		if err := validateFileConfig(c.RequirementsFile, "requirements_file", true); err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	if c.SSHAuthorizedKeyFile != "" {
		if err := validateFileConfig(c.SSHAuthorizedKeyFile, "ssh_authorized_key_file", true); err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	if c.SSHHostKeyFile != "" {
		if err := validateFileConfig(c.SSHHostKeyFile, "ssh_host_key_file", true); err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	// Validate inventory directory
	if c.InventoryDirectory != "" {
		if err := validateInventoryDirectoryConfig(c.InventoryDirectory); err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	// Validate port
	if c.LocalPort > 65535 {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
			"local_port: %d must be a valid port", c.LocalPort))
	}

	// Validate adapter key type
	if c.AdapterKeyType != "" && c.AdapterKeyType != "RSA" && c.AdapterKeyType != "ECDSA" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
			"invalid value for ansible_proxy_key_type: %q. Supported values are ECDSA or RSA",
			c.AdapterKeyType))
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
			c.NavigatorConfig.Format == "" &&
			c.NavigatorConfig.Color == nil &&
			c.NavigatorConfig.Editor == nil &&
			c.NavigatorConfig.Images == nil &&
			c.NavigatorConfig.TimeZone == "" &&
			len(c.NavigatorConfig.InventoryColumns) == 0 &&
			c.NavigatorConfig.CollectionDocCachePath == "" &&
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
		// the nested section blocks (which map to a generated ansible.cfg).
		if c.NavigatorConfig.AnsibleConfig != nil {
			ac := c.NavigatorConfig.AnsibleConfig
			if ac.Config != "" && (ac.Defaults != nil || ac.SSHConnection != nil || ac.PrivilegeEscalation != nil || ac.PersistentConnection != nil || ac.Inventory != nil || ac.ParamikoConnection != nil || ac.Colors != nil || ac.Diff != nil || ac.Galaxy != nil) {
				errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
					"navigator_config.ansible_config.config is mutually exclusive with nested ansible_config section blocks"))
			}
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}
	return nil
}

// Provisioner implements the Packer provisioner interface for Ansible Navigator.
// It manages the lifecycle of Ansible provisioning including SSH proxy setup,
// inventory management, and ansible-navigator command execution.
type Provisioner struct {
	config            Config
	adapter           *adapter.Adapter
	done              chan struct{}
	ansibleVersion    string
	ansibleMajVersion uint
	generatedData     map[string]interface{}

	setupAdapterFunc   func(ui packersdk.Ui, comm packersdk.Communicator) (string, error)
	executeAnsibleFunc func(ui packersdk.Ui, comm packersdk.Communicator, privKeyFile string) error
}

// ConfigSpec returns the HCL2 object spec for the provisioner configuration.
// This is required by the Packer plugin SDK.
func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

// Prepare validates and prepares the provisioner configuration.
// It sets defaults, validates the config, and checks for ansible-navigator availability.
func (p *Provisioner) Prepare(raws ...interface{}) error {
	p.done = make(chan struct{})

	err := config.Decode(&p.config, &config.DecodeOpts{
		PluginType:         "ansible",
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"inventory_file_template",
			},
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

	// Apply HOME expansion to path-like configuration fields
	p.config.InventoryFile = expandUserPath(p.config.InventoryFile)
	p.config.InventoryDirectory = expandUserPath(p.config.InventoryDirectory)
	p.config.RequirementsFile = expandUserPath(p.config.RequirementsFile)
	p.config.SSHHostKeyFile = expandUserPath(p.config.SSHHostKeyFile)
	p.config.SSHAuthorizedKeyFile = expandUserPath(p.config.SSHAuthorizedKeyFile)
	p.config.CollectionsPath = expandUserPath(p.config.CollectionsPath)
	p.config.RolesPath = expandUserPath(p.config.RolesPath)
	p.config.GalaxyCommand = expandUserPath(p.config.GalaxyCommand)
	p.config.BastionPrivateKeyFile = expandUserPath(p.config.BastionPrivateKeyFile)

	// Apply HOME expansion to plays
	for i := range p.config.Plays {
		p.config.Plays[i].Target = expandUserPath(p.config.Plays[i].Target)
		for j := range p.config.Plays[i].VarsFiles {
			p.config.Plays[i].VarsFiles[j] = expandUserPath(p.config.Plays[i].VarsFiles[j])
		}
	}

	if p.config.HostAlias == "" {
		p.config.HostAlias = "default"
	}

	// Migrate legacy flat bastion fields to nested bastion block
	hasLegacyBastionFields := p.config.BastionHost != "" ||
		p.config.BastionPort != 0 ||
		p.config.BastionUser != "" ||
		p.config.BastionPrivateKeyFile != "" ||
		p.config.BastionPassword != ""

	if hasLegacyBastionFields {
		// Display deprecation warning
		log.Printf("[WARN] Deprecated bastion configuration detected. The flat bastion_* fields are deprecated.")
		log.Printf("[WARN] Please migrate to the nested bastion block syntax:")
		log.Printf("[WARN]   bastion {")
		log.Printf("[WARN]     host = \"bastion.example.com\"")
		log.Printf("[WARN]     port = 22")
		log.Printf("[WARN]     user = \"ubuntu\"")
		log.Printf("[WARN]     private_key_file = \"~/.ssh/id_rsa\"")
		log.Printf("[WARN]   }")

		// Create bastion struct if it doesn't exist
		if p.config.Bastion == nil {
			p.config.Bastion = &BastionConfig{}
		}

		// Migrate flat fields to nested struct (new block values take precedence)
		if p.config.Bastion.Host == "" && p.config.BastionHost != "" {
			p.config.Bastion.Host = p.config.BastionHost
		}
		if p.config.Bastion.Port == 0 && p.config.BastionPort != 0 {
			p.config.Bastion.Port = p.config.BastionPort
		}
		if p.config.Bastion.User == "" && p.config.BastionUser != "" {
			p.config.Bastion.User = p.config.BastionUser
		}
		if p.config.Bastion.PrivateKeyFile == "" && p.config.BastionPrivateKeyFile != "" {
			p.config.Bastion.PrivateKeyFile = p.config.BastionPrivateKeyFile
		}
		if p.config.Bastion.Password == "" && p.config.BastionPassword != "" {
			p.config.Bastion.Password = p.config.BastionPassword
		}
	}

	// Set default bastion port if bastion block is defined
	if p.config.Bastion != nil {
		if p.config.Bastion.Port == 0 {
			p.config.Bastion.Port = 22
		}

		// Auto-enable bastion if host is set
		if p.config.Bastion.Host != "" {
			p.config.Bastion.Enabled = true
		}

		// Apply HOME expansion to bastion private key file
		if p.config.Bastion.PrivateKeyFile != "" {
			p.config.Bastion.PrivateKeyFile = expandUserPath(p.config.Bastion.PrivateKeyFile)
		}
	}

	// Also maintain backward compatibility for legacy flat fields
	// (in case validation references them before migration completes)
	if p.config.BastionPort == 0 {
		p.config.BastionPort = 22
	}

	// Set default connection_mode
	if p.config.ConnectionMode == "" {
		p.config.ConnectionMode = "proxy"
	}

	// Detect explicit timeout setting before defaulting
	p.config.versionCheckTimeoutWasSet = p.config.VersionCheckTimeout != nil

	// Set default version check timeout
	if p.config.VersionCheckTimeout == nil {
		p.config.VersionCheckTimeout = stringPtr("60s")
	}

	// Defaults for galaxy
	if p.config.GalaxyCommand == "" {
		p.config.GalaxyCommand = "ansible-galaxy"
	}

	// Set default install directories if not specified
	if p.config.CollectionsPath == "" {
		usr, err := user.Current()
		if err == nil {
			p.config.CollectionsPath = filepath.Join(usr.HomeDir, ".packer.d", "ansible_collections_cache")
		}
	}

	if p.config.RolesPath == "" {
		usr, err := user.Current()
		if err == nil {
			p.config.RolesPath = filepath.Join(usr.HomeDir, ".packer.d", "ansible_roles_cache")
		}
	}

	if !p.config.SkipVersionCheck {
		err = p.getVersion()
		if err != nil {
			return err
		}
	}

	if p.config.User == "" {
		p.config.userWasEmpty = true
		usr, err := user.Current()
		if err != nil {
			return fmt.Errorf("failed to get current user: %w", err)
		} else {
			p.config.User = usr.Username
		}
	}
	if p.config.User == "" {
		return fmt.Errorf("user: could not determine current user from environment")
	}

	// These fields exist so that we can replace the functions for testing
	// logic inside of the Provision func; in actual use, these don't ever
	// need to get set.
	if p.setupAdapterFunc == nil {
		p.setupAdapterFunc = p.setupAdapter
	}
	if p.executeAnsibleFunc == nil {
		p.executeAnsibleFunc = p.executeAnsible
	}

	if p.config.AdapterKeyType == "" {
		p.config.AdapterKeyType = "ECDSA"
	}
	p.config.AdapterKeyType = strings.ToUpper(p.config.AdapterKeyType)

	if p.config.AnsibleProxyBindAddress == "" {
		p.config.AnsibleProxyBindAddress = "127.0.0.1"
	}
	if p.config.AnsibleProxyHost == "" {
		p.config.AnsibleProxyHost = "127.0.0.1"
	}

	// Validate configuration
	if err := p.config.Validate(); err != nil {
		return err
	}

	return nil
}

// detectShim checks if the given file is a version manager shim by reading its header
func detectShim(path string) (string, bool) {
	file, err := os.Open(path)
	if err != nil {
		return "", false
	}
	defer file.Close()

	// Read first 512 bytes to detect shim patterns
	scanner := bufio.NewScanner(file)
	lines := 0
	for scanner.Scan() && lines < 10 {
		line := scanner.Text()
		lines++

		// Check for common shim patterns
		if strings.Contains(line, "asdf exec") || strings.Contains(line, "ASDF") {
			return "asdf", true
		}
		if strings.Contains(line, "rbenv exec") || strings.Contains(line, "RBENV") {
			return "rbenv", true
		}
		if strings.Contains(line, "pyenv exec") || strings.Contains(line, "PYENV") {
			return "pyenv", true
		}
	}

	return "", false
}

// resolveShim attempts to resolve a shim to its real binary using the version manager
func resolveShim(command string, manager string) (string, error) {
	var whichCmd *exec.Cmd

	switch manager {
	case "asdf":
		whichCmd = exec.Command("asdf", "which", command)
	case "rbenv":
		whichCmd = exec.Command("rbenv", "which", command)
	case "pyenv":
		whichCmd = exec.Command("pyenv", "which", command)
	default:
		return "", fmt.Errorf("unsupported version manager: %s", manager)
	}

	output, err := whichCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to resolve shim with %s: %w", manager, err)
	}

	resolvedPath := strings.TrimSpace(string(output))
	if resolvedPath == "" {
		return "", fmt.Errorf("%s returned empty path", manager)
	}

	return resolvedPath, nil
}

func (p *Provisioner) getVersion() error {
	// Use the configured command (defaults to "ansible-navigator")
	command := p.config.Command
	// Safe: defaulted in Prepare
	timeoutStr := "60s"
	if p.config.VersionCheckTimeout != nil {
		timeoutStr = *p.config.VersionCheckTimeout
	}

	// Parse timeout duration
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		// This should never happen as we validate in Prepare(), but handle gracefully
		timeout = 60 * time.Second
	}

	// Detect and resolve shims before version check
	cmdPath, err := exec.LookPath(command)
	if err == nil {
		// Check if the resolved path is a shim
		if manager, isShim := detectShim(cmdPath); isShim {
			log.Printf("Detected %s shim at %s", manager, cmdPath)

			// Attempt to resolve the shim
			resolvedPath, resolveErr := resolveShim(command, manager)
			if resolveErr != nil {
				// Resolution failed - provide actionable error
				return fmt.Errorf(
					"ansible-navigator is installed via %s but could not be resolved automatically.\n"+
						"This may cause hangs or timeouts during version checks.\n\n"+
						"Solutions:\n"+
						"  1. Use 'command' to specify the full path:\n"+
						"     command = \"/full/path/to/ansible-navigator\"\n\n"+
						"  2. Find the path with: %s which ansible-navigator\n\n"+
						"  3. Use 'ansible_navigator_path' to add directories to PATH\n\n"+
						"  4. Use 'skip_version_check = true' to bypass (not recommended)\n\n"+
						"Resolution error: %v",
					manager, manager, resolveErr)
			}

			// Resolution succeeded - use the real binary
			log.Printf("Resolved %s shim to: %s", manager, resolvedPath)
			command = resolvedPath
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create command with context and modified PATH if ansible_navigator_path is set
	cmd := exec.CommandContext(ctx, command, "--version")
	if len(p.config.AnsibleNavigatorPath) > 0 {
		cmd.Env = buildEnvWithPath(p.config.AnsibleNavigatorPath)
	}

	out, err := cmd.Output()
	if err != nil {
		// Check if it was a timeout
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf(
				"ansible-navigator version check timed out after %s.\n"+
					"This may indicate ansible-navigator is not properly installed or configured.\n\n"+
					"Common causes:\n"+
					"  - Version manager shims (asdf, rbenv, pyenv) causing recursion loops\n"+
					"  - Network issues when pulling container images\n"+
					"  - Missing dependencies\n\n"+
					"Solutions:\n"+
					"  1. Use 'command' with the full path to ansible-navigator:\n"+
					"     command = \"/full/path/to/ansible-navigator\"\n"+
					"     (Find with: asdf which ansible-navigator, rbenv which ansible-navigator, or pyenv which ansible-navigator)\n\n"+
					"  2. Use 'ansible_navigator_path' to specify additional directories\n\n"+
					"  3. Use 'skip_version_check = true' to bypass the check\n\n"+
					"  4. Increase 'version_check_timeout' (current: %s)",
				timeoutStr, timeoutStr)
		}
		// Not a timeout - likely not found or other error
		return fmt.Errorf(
			"ansible-navigator not found in PATH or failed to execute. "+
				"Please install it before running this provisioner. "+
				"You can use 'ansible_navigator_path' to specify additional directories, "+
				"or 'skip_version_check = true' to bypass this check. Error: %s", err.Error())
	}

	versionRe := regexp.MustCompile(`\w (\d+\.\d+[.\d+]*)`)
	matches := versionRe.FindStringSubmatch(string(out))
	if matches == nil {
		return fmt.Errorf(
			"Could not find %s version in output:\n%s", command, string(out))
	}

	version := matches[1]
	log.Printf("%s version: %s", command, version)
	p.ansibleVersion = version

	majVer, err := strconv.ParseUint(strings.Split(version, ".")[0], 10, 0)
	if err != nil {
		return fmt.Errorf("could not parse major version from %q: %w", version, err)
	}
	p.ansibleMajVersion = uint(majVer)

	return nil
}

func stringPtr(s string) *string { return &s }

func (p *Provisioner) setupAdapter(ui packersdk.Ui, comm packersdk.Communicator) (string, error) {
	ui.Message("Setting up proxy adapter for Ansible....")

	k, err := newUserKey(p.config.SSHAuthorizedKeyFile, p.config.AdapterKeyType)
	if err != nil {
		return "", err
	}

	hostSigner, err := newSigner(p.config.SSHHostKeyFile, p.config.AdapterKeyType)
	if err != nil {
		return "", fmt.Errorf("error creating host signer: %w", err)
	}

	keyChecker := ssh.CertChecker{
		UserKeyFallback: func(conn ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			if user := conn.User(); user != p.config.User {
				return nil, fmt.Errorf("authentication failed: %s is not a valid user", user)
			}

			if !bytes.Equal(k.Marshal(), pubKey.Marshal()) {
				return nil, errors.New("authentication failed: unauthorized key")
			}

			return nil, nil
		},
		IsUserAuthority: func(k ssh.PublicKey) bool { return true },
	}

	config := &ssh.ServerConfig{
		AuthLogCallback: func(conn ssh.ConnMetadata, method string, err error) {
			log.Printf("authentication attempt from %s to %s as %s using %s", conn.RemoteAddr(), conn.LocalAddr(), conn.User(), method)
		},
		PublicKeyCallback: keyChecker.Authenticate,
		//NoClientAuth:      true,
	}

	config.AddHostKey(hostSigner)

	localListener, err := func() (net.Listener, error) {

		port := p.config.LocalPort
		tries := 1
		if port != 0 {
			tries = 10
		}
		for i := 0; i < tries; i++ {
			l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", p.config.AnsibleProxyBindAddress, port))
			port++
			if err != nil {
				ui.Say(err.Error())
				continue
			}
			_, portStr, err := net.SplitHostPort(l.Addr().String())
			if err != nil {
				ui.Say(err.Error())
				continue
			}
			p.config.LocalPort, err = strconv.Atoi(portStr)
			if err != nil {
				ui.Say(err.Error())
				continue
			}
			return l, nil
		}
		return nil, errors.New("Error setting up SSH proxy connection")
	}()

	if err != nil {
		return "", err
	}

	ui = &packersdk.SafeUi{
		Sem: make(chan int, 1),
		Ui:  ui,
	}
	p.adapter = adapter.NewAdapter(p.done, localListener, config, p.config.SFTPCmd, ui, comm)

	return k.privKeyFile, nil
}

// setupSSHTunnel establishes an SSH tunnel through a bastion host to the target machine.
// It creates a local port forward that Ansible can use to reach the target.
// Returns the local port number, a cleanup closer, and an error if setup fails.
func (p *Provisioner) setupSSHTunnel(ui packersdk.Ui, targetHost string, targetPort int) (int, io.Closer, error) {
	ui.Message("Setting up SSH tunnel through bastion...")

	// Get bastion config from new structure or fall back to legacy flat fields
	var bastionHost, bastionUser, bastionPrivateKeyFile, bastionPassword string
	var bastionPort int

	if p.config.Bastion != nil {
		bastionHost = p.config.Bastion.Host
		bastionPort = p.config.Bastion.Port
		bastionUser = p.config.Bastion.User
		bastionPrivateKeyFile = p.config.Bastion.PrivateKeyFile
		bastionPassword = p.config.Bastion.Password
	} else {
		// Fall back to legacy flat fields
		bastionHost = p.config.BastionHost
		bastionPort = p.config.BastionPort
		bastionUser = p.config.BastionUser
		bastionPrivateKeyFile = p.config.BastionPrivateKeyFile
		bastionPassword = p.config.BastionPassword
	}

	// Parse bastion authentication methods
	var authMethods []ssh.AuthMethod

	// Try private key authentication first if specified
	if bastionPrivateKeyFile != "" {
		keyBytes, err := os.ReadFile(bastionPrivateKeyFile)
		if err != nil {
			return 0, nil, fmt.Errorf("failed to read bastion private key file: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(keyBytes)
		if err != nil {
			return 0, nil, fmt.Errorf("failed to parse bastion private key: %w", err)
		}

		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	// Add password authentication if specified
	if bastionPassword != "" {
		authMethods = append(authMethods, ssh.Password(bastionPassword))
	}

	// Configure SSH client for bastion connection
	bastionConfig := &ssh.ClientConfig{
		User:            bastionUser,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Accept any host key
		Timeout:         30 * time.Second,
	}

	// Connect to bastion host
	bastionAddr := fmt.Sprintf("%s:%d", bastionHost, bastionPort)
	ui.Message(fmt.Sprintf("Connecting to bastion host %s...", bastionAddr))

	bastionClient, err := ssh.Dial("tcp", bastionAddr, bastionConfig)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to connect to bastion host %s: %w", bastionAddr, err)
	}

	// Allocate local port for tunnel
	var localPort int
	var localListener net.Listener

	port := p.config.LocalPort
	tries := 1
	if port != 0 {
		tries = 10
	}

	// Try to allocate a local port
	for i := 0; i < tries; i++ {
		addr := fmt.Sprintf("127.0.0.1:%d", port)
		l, err := net.Listen("tcp", addr)
		if err != nil {
			ui.Say(fmt.Sprintf("Port %d unavailable: %v", port, err))
			port++
			continue
		}

		// Extract the actual port (may be system-assigned if port was 0)
		_, portStr, err := net.SplitHostPort(l.Addr().String())
		if err != nil {
			l.Close()
			ui.Say(fmt.Sprintf("Failed to parse local address: %v", err))
			port++
			continue
		}

		localPort, err = strconv.Atoi(portStr)
		if err != nil {
			l.Close()
			ui.Say(fmt.Sprintf("Failed to parse port number: %v", err))
			port++
			continue
		}

		localListener = l
		break
	}

	if localListener == nil {
		bastionClient.Close()
		return 0, nil, fmt.Errorf("failed to allocate local port for tunnel")
	}

	ui.Message(fmt.Sprintf("Tunnel listening on 127.0.0.1:%d", localPort))
	ui.Message(fmt.Sprintf("Forwarding to %s:%d through bastion", targetHost, targetPort))

	// Create cleanup closer
	closer := &tunnelCloser{
		listener:      localListener,
		bastionClient: bastionClient,
		ui:            ui,
	}

	// Start forwarding goroutine
	go func() {
		for {
			localConn, err := localListener.Accept()
			if err != nil {
				// Listener was closed
				return
			}

			// Handle this connection in a goroutine
			go func(local net.Conn) {
				defer local.Close()

				// Dial the target through the bastion
				targetAddr := fmt.Sprintf("%s:%d", targetHost, targetPort)
				remoteConn, err := bastionClient.Dial("tcp", targetAddr)
				if err != nil {
					ui.Error(fmt.Sprintf("Failed to establish tunnel to target %s: %v", targetAddr, err))
					return
				}
				defer remoteConn.Close()

				// Copy data bidirectionally
				done := make(chan struct{}, 2)
				go func() {
					io.Copy(remoteConn, local)
					done <- struct{}{}
				}()
				go func() {
					io.Copy(local, remoteConn)
					done <- struct{}{}
				}()
				<-done
			}(localConn)
		}
	}()

	return localPort, closer, nil
}

// tunnelCloser implements io.Closer for SSH tunnel cleanup
type tunnelCloser struct {
	listener      net.Listener
	bastionClient *ssh.Client
	ui            packersdk.Ui
}

func (tc *tunnelCloser) Close() error {
	tc.ui.Message("Closing SSH tunnel...")

	// Close the listener first to stop accepting new connections
	if tc.listener != nil {
		tc.listener.Close()
	}

	// Close the bastion SSH client
	if tc.bastionClient != nil {
		tc.bastionClient.Close()
	}

	tc.ui.Message("SSH tunnel closed")
	return nil
}

const DefaultSSHInventoryFilev2 = "{{ .HostAlias }} ansible_host={{ .Host }} ansible_user={{ .User }} ansible_port={{ .Port }}\n"
const DefaultSSHInventoryFilev1 = "{{ .HostAlias }} ansible_ssh_host={{ .Host }} ansible_ssh_user={{ .User }} ansible_ssh_port={{ .Port }}\n"
const DefaultWinRMInventoryFilev2 = "{{ .HostAlias}} ansible_host={{ .Host }} ansible_connection=winrm ansible_winrm_transport=basic ansible_shell_type=powershell ansible_user={{ .User}} ansible_port={{ .Port }}\n"

func (p *Provisioner) createInventoryFile() error {
	log.Printf("Creating inventory file for Ansible run...")
	tf, err := os.CreateTemp(p.config.InventoryDirectory, "packer-provisioner-ansible")
	if err != nil {
		return fmt.Errorf("error preparing inventory file: %w", err)
	}

	// If user has defiend their own inventory template, use it.
	hostTemplate := p.config.InventoryFileTemplate
	if hostTemplate == "" {
		// figure out which inventory line template to use
		hostTemplate = DefaultSSHInventoryFilev2
		if p.ansibleMajVersion < 2 {
			hostTemplate = DefaultSSHInventoryFilev1
		}
		if p.config.ConnectionMode == "direct" && p.generatedData["ConnType"] == "winrm" {
			hostTemplate = DefaultWinRMInventoryFilev2
		}
	}

	// interpolate template to generate host with necessary vars.
	ctxData := p.generatedData
	ctxData["HostAlias"] = p.config.HostAlias
	ctxData["User"] = p.config.User
	if p.config.ConnectionMode == "proxy" || p.config.ConnectionMode == "ssh_tunnel" {
		ctxData["Host"] = p.config.AnsibleProxyHost
		ctxData["Port"] = p.config.LocalPort
	}
	p.config.ctx.Data = ctxData

	host, err := interpolate.Render(hostTemplate, &p.config.ctx)
	if err != nil {
		return fmt.Errorf("error generating inventory file from template: %w", err)
	}

	w := bufio.NewWriter(tf)
	if _, err := w.WriteString(host); err != nil {
		log.Printf("[TRACE] error writing the generated inventory file: %s", err)
	}

	for _, group := range p.config.Groups {
		fmt.Fprintf(w, "[%s]\n%s", group, host)
	}

	for _, group := range p.config.EmptyGroups {
		fmt.Fprintf(w, "[%s]\n", group)
	}

	if err := w.Flush(); err != nil {
		tf.Close()
		os.Remove(tf.Name())
		return fmt.Errorf("error preparing inventory file: %w", err)
	}
	tf.Close()
	p.config.InventoryFile = tf.Name()

	return nil
}

// Provision executes the Ansible Navigator provisioning process.
// It sets up the SSH proxy (if needed), manages inventory, installs dependencies,
// and executes the configured playbooks or plays.
func (p *Provisioner) Provision(ctx context.Context, ui packersdk.Ui, comm packersdk.Communicator, generatedData map[string]interface{}) error {
	ui.Say("Provisioning with Ansible Navigator...")
	if p.config.SkipVersionCheck && p.config.versionCheckTimeoutWasSet {
		ui.Message("Warning: version_check_timeout is ignored when skip_version_check=true")
	}

	p.generatedData = generatedData
	p.config.ctx.Data = generatedData

	privKeyFile := ""

	// Handle connection based on connection_mode
	switch p.config.ConnectionMode {
	case "ssh_tunnel":
		// Extract target host and port from generatedData
		targetHost, ok := generatedData["Host"].(string)
		if !ok || targetHost == "" {
			return fmt.Errorf("SSH tunnel mode requires a valid target host")
		}

		// Extract target port with type-safe handling (int or string)
		var targetPort int
		switch v := generatedData["Port"].(type) {
		case int:
			targetPort = v
		case string:
			var err error
			targetPort, err = strconv.Atoi(v)
			if err != nil {
				return fmt.Errorf("SSH tunnel mode: invalid port value %q: %w", v, err)
			}
		default:
			return fmt.Errorf("SSH tunnel mode: Port must be int or string, got type %T with value %v", v, v)
		}

		// Validate port range
		if targetPort < 1 || targetPort > 65535 {
			return fmt.Errorf("SSH tunnel mode: port must be between 1-65535, got %d", targetPort)
		}

		// SSH tunnel mode - establish tunnel through bastion
		ui.Message("Using SSH tunnel mode - connecting through bastion host")

		// Establish SSH tunnel
		localPort, tunnel, err := p.setupSSHTunnel(ui, targetHost, targetPort)
		if err != nil {
			return fmt.Errorf("failed to setup SSH tunnel: %w", err)
		}

		// Store the local port for inventory generation
		p.config.LocalPort = localPort

		// Override generatedData to point to the tunnel
		p.generatedData["Host"] = "127.0.0.1"
		p.generatedData["Port"] = localPort

		bastionHost := p.config.BastionHost
		bastionPort := p.config.BastionPort
		if p.config.Bastion != nil {
			bastionHost = p.config.Bastion.Host
			bastionPort = p.config.Bastion.Port
		}

		ui.Message(fmt.Sprintf("SSH tunnel established: localhost:%d -> %s:%d (via bastion %s:%d)",
			localPort, targetHost, targetPort, bastionHost, bastionPort))

		// Debug logging for tunnel inventory integration
		debugEnabled := isPluginDebugEnabled(p.config.NavigatorConfig)
		debugf(ui, debugEnabled, "SSH tunnel mode active: inventory will use tunnel endpoint")
		debugf(ui, debugEnabled, "Tunnel connection: 127.0.0.1:%d", localPort)
		debugf(ui, debugEnabled, "Target user: %s", p.config.User)

		// Log target SSH key if available
		if sshKeyFile, ok := p.generatedData["SSHPrivateKeyFile"].(string); ok && sshKeyFile != "" {
			debugf(ui, debugEnabled, "Target SSH key: %s", sshKeyFile)
		}

		// Ensure tunnel cleanup
		defer func() {
			if tunnel != nil {
				tunnel.Close()
			}
		}()

		// Get SSH private key for Ansible to use when connecting through the tunnel
		// The tunnel provides network path, but Ansible still needs target credentials
		connType := generatedData["ConnType"].(string)
		if connType == "ssh" {
			SSHPrivateKeyFile := generatedData["SSHPrivateKeyFile"].(string)
			SSHAgentAuth := generatedData["SSHAgentAuth"].(bool)
			if SSHPrivateKeyFile != "" || SSHAgentAuth {
				privKeyFile = SSHPrivateKeyFile
			} else {
				// Get private key from generatedData and write to temp file
				SSHPrivateKey := generatedData["SSHPrivateKey"].(string)
				tmpSSHPrivateKey, err := tmp.File("ansible-key")
				if err != nil {
					return fmt.Errorf("error writing private key to temp file for ansible connection: %v", err)
				}
				_, err = tmpSSHPrivateKey.WriteString(SSHPrivateKey)
				if err != nil {
					return errors.New("failed to write private key to temp file")
				}
				err = tmpSSHPrivateKey.Close()
				if err != nil {
					return errors.New("failed to close private key temp file")
				}
				privKeyFile = tmpSSHPrivateKey.Name()
				defer os.Remove(privKeyFile)
			}

			// Match username to SSH keys if not explicitly set
			if p.config.userWasEmpty {
				p.config.User = generatedData["User"].(string)
			}
		}

	case "direct":
		// Direct connection mode - use SSH keys from communicator
		ui.Message("Using direct connection mode - connecting without proxy")

		connType := generatedData["ConnType"].(string)
		switch connType {
		case "ssh":
			// In this situation, we need to make sure we have the
			// private key we actually use to access the instance.
			SSHPrivateKeyFile := generatedData["SSHPrivateKeyFile"].(string)
			SSHAgentAuth := generatedData["SSHAgentAuth"].(bool)
			if SSHPrivateKeyFile != "" || SSHAgentAuth {
				privKeyFile = SSHPrivateKeyFile
			} else {
				// See if we can get a private key and write that to a tmpfile
				SSHPrivateKey := generatedData["SSHPrivateKey"].(string)
				tmpSSHPrivateKey, err := tmp.File("ansible-key")
				if err != nil {
					return fmt.Errorf("Error writing private key to temp file for"+
						"ansible connection: %v", err)
				}
				_, err = tmpSSHPrivateKey.WriteString(SSHPrivateKey)
				if err != nil {
					return errors.New("failed to write private key to temp file")
				}
				err = tmpSSHPrivateKey.Close()
				if err != nil {
					return errors.New("failed to close private key temp file")
				}
				privKeyFile = tmpSSHPrivateKey.Name()
			}

			// Also make sure that the username matches the SSH keys given.
			if p.config.userWasEmpty {
				p.config.User = generatedData["User"].(string)
			}
		case "winrm":
			ui.Message("Using WinRM password from Packer communicator")
		}

	case "proxy":
		// Proxy adapter mode (default)
		ui.Message("Using proxy adapter mode")

		pkf, err := p.setupAdapterFunc(ui, comm)
		if err != nil {
			return err
		}
		privKeyFile = pkf

		defer func() {
			log.Print("shutting down the SSH proxy")
			close(p.done)
			p.adapter.Shutdown()
		}()

		go p.adapter.Serve()

		// Remove the private key file
		if len(privKeyFile) > 0 {
			defer os.Remove(privKeyFile)
		}

	default:
		return fmt.Errorf("invalid connection_mode: %q (should have been caught by validation)", p.config.ConnectionMode)
	}

	if len(p.config.InventoryFile) == 0 {
		// Create the inventory file
		err := p.createInventoryFile()
		if err != nil {
			return err
		}
		if !p.config.KeepInventoryFile {
			// Delete the generated inventory file
			defer func() {
				os.Remove(p.config.InventoryFile)
				p.config.InventoryFile = ""
			}()
		}
	}

	if err := p.executeAnsibleFunc(ui, comm, privKeyFile); err != nil {
		return fmt.Errorf("error executing Ansible: %w", err)
	}

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
// and returns the file path. The caller is responsible for cleanup.
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

// createCmdArgs constructs ansible-navigator command arguments and returns
// the file path to the temporary extra vars file (which must be cleaned up by the caller).
func (p *Provisioner) createCmdArgs(ui packersdk.Ui, httpAddr, inventory, privKeyFile string) (args []string, envVars []string, extraVarsFilePath string, err error) {
	args = []string{}

	// Provisioner-generated extra vars MUST be conveyed via a single JSON object
	// passed through exactly one -e/--extra-vars argument pair to avoid malformed
	// argument construction and positional argument shifting.
	//
	// To prevent shell interpretation errors in execution environments, we write
	// the JSON to a temporary file and pass it via @filepath syntax.
	extraVars := make(map[string]interface{})

	if p.config.PackerBuildName != "" {
		// HCL configs don't currently have the PakcerBuildName. Don't
		// cause weirdness with a half-set variable
		extraVars["packer_build_name"] = p.config.PackerBuildName
	}
	extraVars["packer_builder_type"] = p.config.PackerBuilderType

	// expose packer_http_addr extra variable
	if httpAddr != commonsteps.HttpAddrNotImplemented {
		extraVars["packer_http_addr"] = httpAddr
	}

	// Add password to ansible call.
	ansiblePasswordSet := false
	if p.config.ConnectionMode == "direct" && p.generatedData["ConnType"] == "winrm" {
		if password, ok := p.generatedData["Password"]; ok {
			extraVars["ansible_password"] = fmt.Sprint(password)
			ansiblePasswordSet = true
		}
	}

	if !ansiblePasswordSet && len(privKeyFile) > 0 {
		// "-e ansible_ssh_private_key_file" is preferable to "--private-key"
		// because it is a higher priority variable and therefore won't get
		// overridden by dynamic variables. See #5852 for more details.
		extraVars["ansible_ssh_private_key_file"] = privKeyFile
	}

	// If using SSH password auth, disable host key checking unless user overrides.
	// (This mirrors the previous behavior, but uses JSON extra-vars.)
	if ansiblePasswordSet && p.generatedData["ConnType"] == "ssh" {
		if _, ok := extraVars["ansible_host_key_checking"]; !ok {
			extraVars["ansible_host_key_checking"] = false
		}
	}

	// Write extra vars to temporary file to prevent shell interpretation issues
	extraVarsFilePath, err = createExtraVarsFile(extraVars)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to create extra vars file: %w", err)
	}

	// Pass file path with @ prefix (Ansible's file-based extra-vars syntax)
	args = append(args, fmt.Sprintf("--extra-vars=@%s", extraVarsFilePath))

	// Log extra vars if ShowExtraVars is enabled
	if p.config.ShowExtraVars {
		logExtraVarsJSON(ui, extraVars)
	}

	if p.generatedData["ConnType"] == "ssh" && len(privKeyFile) > 0 {
		// Add ssh extra args to set IdentitiesOnly
		args = append(args, "--ssh-extra-args=-o IdentitiesOnly=yes")
	}

	// Add limit if specified
	if p.config.Limit != "" {
		args = append(args, fmt.Sprintf("--limit=%s", p.config.Limit))
	}

	// This must be the last arg appended to args (the play target is appended later).
	args = append(args, fmt.Sprintf("-i=%s", inventory))
	return args, envVars, extraVarsFilePath, nil
}

// createRolePlaybook generates a temporary playbook file for executing an Ansible role.
// It converts a role FQDN into a valid playbook with the specified configuration.
// The caller is responsible for cleaning up the temporary file.
func createRolePlaybook(role string, play Play) (string, error) {
	// Create a temporary file for the playbook
	tmpFile, err := os.CreateTemp("", "packer-role-playbook-*.yml")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary playbook file: %w", err)
	}

	// Build the playbook content
	playbookContent := "---\n- hosts: all\n"

	if play.Become {
		playbookContent += "  become: yes\n"
	}

	if play.BecomeUser != "" {
		playbookContent += fmt.Sprintf("  become_user: %s\n", play.BecomeUser)
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

func (p *Provisioner) executeAnsible(ui packersdk.Ui, comm packersdk.Communicator, privKeyFile string) error {
	httpAddr := p.generatedData["PackerHTTPAddr"].(string)

	debugEnabled := isPluginDebugEnabled(p.config.NavigatorConfig)
	debugf(ui, debugEnabled, "Plugin debug mode enabled (gated by navigator_config.logging.level=debug)")

	// Generate and setup ansible-navigator.yml if configured
	var navigatorConfigPath string
	if p.config.NavigatorConfig != nil {
		// Pass collections path directly to navigator config generation.
		// The path should point to the directory that contains the ansible_collections/ subdirectory.
		// ansible-galaxy installs to <collections_path>/ansible_collections/<namespace>/<collection>,
		// and Ansible expects ANSIBLE_COLLECTIONS_PATH to point to the parent directory.
		collectionsPath := p.config.CollectionsPath

		// If ansible_config.defaults / ansible_config.ssh_connection are provided
		// (explicitly or via EE defaults), generate an ansible.cfg file and reference
		// it from the navigator config YAML via ansible.config.path.
		if p.config.NavigatorConfig.AnsibleConfig != nil && needsGeneratedAnsibleCfg(p.config.NavigatorConfig.AnsibleConfig) {
			if p.config.NavigatorConfig.AnsibleConfig.Config != "" {
				return fmt.Errorf("invalid navigator_config.ansible_config: config is mutually exclusive with defaults/ssh_connection")
			}

			cfgContent, err := generateAnsibleCfgContent(p.config.NavigatorConfig.AnsibleConfig)
			if err != nil {
				return fmt.Errorf("failed to generate ansible.cfg content: %w", err)
			}
			if cfgContent != "" {
				cfgPath, err := createTempAnsibleCfgFile(cfgContent)
				if err != nil {
					return fmt.Errorf("failed to create temporary ansible.cfg: %w", err)
				}
				ui.Message(fmt.Sprintf("Generated ansible.cfg at %s", cfgPath))
				p.config.NavigatorConfig.AnsibleConfig.Config = cfgPath
				defer func() {
					_ = os.Remove(cfgPath)
				}()
			}
		}

		yamlContent, err := generateNavigatorConfigYAML(p.config.NavigatorConfig, collectionsPath)
		if err != nil {
			return fmt.Errorf("failed to generate navigator_config YAML: %w", err)
		}

		navigatorConfigPath, err = createNavigatorConfigFile(yamlContent)
		if err != nil {
			return fmt.Errorf("failed to create temporary ansible-navigator.yml: %w", err)
		}

		ui.Message(fmt.Sprintf("Generated ansible-navigator.yml at %s", navigatorConfigPath))
		debugf(ui, debugEnabled, "ANSIBLE_NAVIGATOR_CONFIG will be set to %s", navigatorConfigPath)

		// Ensure cleanup on exit (success or failure)
		defer func() {
			if err := os.Remove(navigatorConfigPath); err != nil {
				ui.Message(fmt.Sprintf("Warning: failed to remove temporary ansible-navigator.yml: %v", err))
			}
		}()
	}

	// Install dependencies using GalaxyManager
	galaxyManager := NewGalaxyManager(&p.config, ui)

	if err := galaxyManager.InstallRequirements(); err != nil {
		return fmt.Errorf("failed to install requirements: %w", err)
	}

	// Setup environment paths for collections and roles
	if err := galaxyManager.SetupEnvironmentPaths(); err != nil {
		return fmt.Errorf("failed to setup environment paths: %w", err)
	}

	// Execute plays (required by validation)
	return p.executePlays(ui, comm, privKeyFile, httpAddr, navigatorConfigPath)
}

// buildRunCommandArgsForPlay constructs the full command arguments for a play
// and returns the command args, environment variables, and path to the temp extra vars file.
// The caller is responsible for cleaning up the extra vars file.
func (p *Provisioner) buildRunCommandArgsForPlay(ui packersdk.Ui, play Play, httpAddr, inventory, playbookPath, privKeyFile string) (cmdArgs []string, envvars []string, extraVarsFilePath string, err error) {
	// Build command arguments (excluding play target; appended last for deterministic ordering)
	baseArgs, envvars, extraVarsFilePath, err := p.createCmdArgs(ui, httpAddr, inventory, privKeyFile)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to create command args: %w", err)
	}

	// Play-level flags (deterministic order)
	playArgs := make([]string, 0)
	if play.Become {
		playArgs = append(playArgs, "--become")
	}
	if play.BecomeUser != "" {
		playArgs = append(playArgs, fmt.Sprintf("--become-user=%s", play.BecomeUser))
	}
	for _, tag := range play.Tags {
		playArgs = append(playArgs, fmt.Sprintf("--tags=%s", tag))
	}
	for _, tag := range play.SkipTags {
		playArgs = append(playArgs, fmt.Sprintf("--skip-tags=%s", tag))
	}
	if len(play.ExtraVars) > 0 {
		keys := make([]string, 0, len(play.ExtraVars))
		for k := range play.ExtraVars {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			playArgs = append(playArgs, fmt.Sprintf("-e=%s=%s", k, play.ExtraVars[k]))
		}
	}
	for _, varsFile := range play.VarsFiles {
		playArgs = append(playArgs, fmt.Sprintf("-e=@%s", varsFile))
	}

	// Deterministic ordering:
	//   1) ansible-navigator run (+ enforced --mode)
	//   2) play.extra_args (verbatim)
	//   3) plugin-generated inventory/extra-vars/etc (including play-level flags)
	//   4) play target (playbook path)
	cmdArgs = []string{"run"}
	if p.config.NavigatorConfig != nil && p.config.NavigatorConfig.Mode != "" {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--mode=%s", p.config.NavigatorConfig.Mode))
	}
	cmdArgs = append(cmdArgs, play.ExtraArgs...)
	cmdArgs = append(cmdArgs, playArgs...)
	cmdArgs = append(cmdArgs, baseArgs...)
	cmdArgs = append(cmdArgs, playbookPath)

	return cmdArgs, envvars, extraVarsFilePath, nil
}

// executePlays executes multiple Ansible plays in sequence.
// If a play fails and keep_going is false, execution stops immediately.
// Otherwise, errors are logged and execution continues to the next play.
func (p *Provisioner) executePlays(ui packersdk.Ui, comm packersdk.Communicator, privKeyFile string, httpAddr string, navigatorConfigPath string) error {
	inventory := p.config.InventoryFile

	debugEnabled := isPluginDebugEnabled(p.config.NavigatorConfig)
	debugf(ui, debugEnabled, "ansible-navigator command=%q", p.config.Command)
	if len(p.config.AnsibleNavigatorPath) > 0 {
		debugf(ui, debugEnabled, "ansible_navigator_path prefixes=%v", p.config.AnsibleNavigatorPath)
	} else {
		debugf(ui, debugEnabled, "ansible_navigator_path not set; using existing PATH")
	}
	if navigatorConfigPath != "" {
		debugf(ui, debugEnabled, "ANSIBLE_NAVIGATOR_CONFIG=%s", navigatorConfigPath)
	}

	for i, play := range p.config.Plays {
		playName := play.Name
		if playName == "" {
			playName = fmt.Sprintf("Play %d", i+1)
		}

		ui.Say(fmt.Sprintf("Executing %s: %s", playName, play.Target))

		var playbookPath string
		var cleanupFunc func()

		// Determine if target is a playbook file or a role
		if strings.HasSuffix(play.Target, ".yml") || strings.HasSuffix(play.Target, ".yaml") {
			// It's a playbook file
			absPath, err := filepath.Abs(play.Target)
			if err != nil {
				return fmt.Errorf("Play '%s': failed to resolve playbook path: %s", playName, err)
			}
			playbookPath = absPath
			debugf(ui, debugEnabled, "Resolved playbook path: %s -> %s", play.Target, absPath)
		} else {
			// It's a role - generate a temporary playbook
			debugf(ui, debugEnabled, "Play target treated as role; generating temporary playbook for role=%s", play.Target)
			ui.Message(fmt.Sprintf("Generating temporary playbook for role: %s", play.Target))
			tmpPlaybook, err := createRolePlaybook(play.Target, play)
			if err != nil {
				return fmt.Errorf("play %q: failed to generate role playbook: %w", playName, err)
			}
			playbookPath = tmpPlaybook
			debugf(ui, debugEnabled, "Generated temporary playbook path=%s", tmpPlaybook)
			cleanupFunc = func() {
				os.Remove(tmpPlaybook)
			}
		}

		cmdArgs, envvars, extraVarsFilePath, err := p.buildRunCommandArgsForPlay(ui, play, httpAddr, inventory, playbookPath, privKeyFile)
		if err != nil {
			if cleanupFunc != nil {
				cleanupFunc()
			}
			return fmt.Errorf("play %q: failed to build command args: %w", playName, err)
		}

		// Ensure cleanup of temp extra vars file (even on error)
		defer func() {
			if extraVarsFilePath != "" {
				if err := os.Remove(extraVarsFilePath); err != nil {
					ui.Message(fmt.Sprintf("Warning: failed to remove temporary extra vars file %s: %v", extraVarsFilePath, err))
				} else {
					debugf(ui, debugEnabled, "Cleaned up temporary extra vars file: %s", extraVarsFilePath)
				}
			}
		}()

		cmd := exec.Command(p.config.Command, cmdArgs...)

		// Set environment with modified PATH if needed
		if len(p.config.AnsibleNavigatorPath) > 0 {
			cmd.Env = buildEnvWithPath(p.config.AnsibleNavigatorPath)
		} else {
			cmd.Env = os.Environ()
		}
		// Add ANSIBLE_NAVIGATOR_CONFIG if navigator_config was provided
		if navigatorConfigPath != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("ANSIBLE_NAVIGATOR_CONFIG=%s", navigatorConfigPath))
			debugf(ui, debugEnabled, "Setting ANSIBLE_NAVIGATOR_CONFIG for %s", playName)
		}
		if len(envvars) > 0 {
			cmd.Env = append(cmd.Env, envvars...)
		}

		// DEBUG-only EE/docker preflight diagnostics (no behavior changes)
		if debugEnabled && isExecutionEnvironmentEnabled(p.config.NavigatorConfig) {
			emitEEDockerPreflight(ui, debugEnabled, p.config.AnsibleNavigatorPath)
		}

		execErr := p.executeAnsibleCommand(ui, cmd, playName)

		// Cleanup temporary playbook if it was generated
		if cleanupFunc != nil {
			cleanupFunc()
		}

		if execErr != nil {
			ui.Error(fmt.Sprintf("Play '%s' failed: %v", playName, execErr))
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

func (p *Provisioner) executeAnsibleCommand(ui packersdk.Ui, cmd *exec.Cmd, target string) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}

	// Check if we should use structured JSON logging
	useStructuredLogging := p.config.StructuredLogging
	var summary *Summary

	if useStructuredLogging {
		summary = &Summary{
			PlaysRun:    0,
			TasksTotal:  0,
			TasksFailed: 0,
			FailedTasks: make([]NavigatorEvent, 0),
		}
	}

	// Handler for stdout - either JSON parsing or line-by-line
	stdoutHandler := func(r io.ReadCloser) {
		defer wg.Done()

		if useStructuredLogging {
			// Use streaming JSON decoder
			decoder := json.NewDecoder(r)
			for decoder.More() {
				var event NavigatorEvent
				if err := decoder.Decode(&event); err != nil {
					// Skip malformed JSON with a warning
					if p.config.VerboseTaskOutput {
						ui.Message(fmt.Sprintf("[Warning] Skipped malformed JSON event: %v", err))
					}
					continue
				}
				handleNavigatorEvent(ui, &event, summary, p.config.VerboseTaskOutput)
			}
		} else {
			// Regular line-by-line output
			reader := bufio.NewReader(r)
			for {
				line, err := reader.ReadString('\n')
				if line != "" {
					line = strings.TrimRightFunc(line, unicode.IsSpace)
					ui.Message(line)
				}
				if err != nil {
					if err == io.EOF {
						break
					} else {
						ui.Error(err.Error())
						break
					}
				}
			}
		}
	}

	// Handler for stderr - always line-by-line
	stderrHandler := func(r io.ReadCloser) {
		defer wg.Done()
		reader := bufio.NewReader(r)
		for {
			line, err := reader.ReadString('\n')
			if line != "" {
				line = strings.TrimRightFunc(line, unicode.IsSpace)
				ui.Error(line)
			}
			if err != nil {
				if err == io.EOF {
					break
				} else {
					ui.Error(err.Error())
					break
				}
			}
		}
	}

	wg.Add(2)
	go stdoutHandler(stdout)
	go stderrHandler(stderr)

	// remove sensitive data from command for logging
	flattenedCmd := strings.Join(cmd.Args, " ")
	sanitized := flattenedCmd

	for _, key := range []string{"WinRMPassword", "Password"} {
		secret, ok := p.generatedData[key]
		if ok && secret != "" {
			sanitized = strings.Replace(sanitized,
				secret.(string), "*****", -1)
		}
	}
	ui.Say(fmt.Sprintf("Executing Ansible Navigator for %s: %s", target, sanitized))

	if err := cmd.Start(); err != nil {
		return err
	}
	wg.Wait()

	// Report summary if structured logging was used
	if useStructuredLogging && summary != nil {
		if summary.TasksFailed > 0 {
			ui.Error(fmt.Sprintf("[Error] %d task(s) failed during play execution.", summary.TasksFailed))
			for _, failedTask := range summary.FailedTasks {
				ui.Error(fmt.Sprintf("  - Task '%s' on host '%s'", failedTask.Task, failedTask.Host))
			}
		}

		if summary.TasksTotal == 0 && summary.PlaysRun == 0 {
			ui.Message("[Warning] No valid events parsed from ansible-navigator output.")
		} else {
			ui.Message(fmt.Sprintf("Summary: %d play(s) executed, %d task(s) total, %d failed",
				summary.PlaysRun, summary.TasksTotal, summary.TasksFailed))
		}

		// Write summary JSON if path is specified
		if p.config.LogOutputPath != "" {
			if err := writeSummaryJSON(summary, p.config.LogOutputPath); err != nil {
				ui.Message(fmt.Sprintf("[Warning] Could not write structured log to %s: %v", p.config.LogOutputPath, err))
			} else {
				ui.Message(fmt.Sprintf("Structured log written to: %s", p.config.LogOutputPath))
			}
		}
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("non-zero exit status: %w", err)
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
		return fmt.Errorf("%s: %s is invalid: %w", config, name, err)
	} else if info.IsDir() {
		return fmt.Errorf("%s: %s must point to a file", config, name)
	}
	return nil
}

func validateInventoryDirectoryConfig(name string) error {
	info, err := os.Stat(name)
	if err != nil {
		return fmt.Errorf("inventory_directory: %s is invalid: %w", name, err)
	} else if !info.IsDir() {
		return fmt.Errorf("inventory_directory: %s must point to a directory", name)
	}
	return nil
}

type userKey struct {
	ssh.PublicKey
	privKeyFile string
}

func newUserKey(pubKeyFile string, keyType string) (*userKey, error) {
	userKey := new(userKey)
	if len(pubKeyFile) > 0 {
		pubKeyBytes, err := os.ReadFile(pubKeyFile)
		if err != nil {
			return nil, errors.New("Failed to read public key")
		}
		userKey.PublicKey, _, _, _, err = ssh.ParseAuthorizedKey(pubKeyBytes)
		if err != nil {
			return nil, errors.New("Failed to parse authorized key")
		}

		return userKey, nil
	}

	tf, err := tmp.File("ansible-key")
	if err != nil {
		return nil, errors.New("failed to create temp file for generated key")
	}

	switch keyType {
	case "RSA":
		err = generateRSAKeyToFile(userKey, tf)
	case "ECDSA":
		err = generateECDSAKeyToFile(userKey, tf)
	default:
		err = fmt.Errorf("unknown key type: %q", keyType)
	}
	if err != nil {
		return nil, err
	}

	err = tf.Close()
	if err != nil {
		return nil, errors.New("failed to close private key temp file")
	}
	userKey.privKeyFile = tf.Name()

	return userKey, nil
}

func generateECDSAKeyToFile(uk *userKey, target *os.File) error {
	privKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return errors.New("Failed to generate key pair")
	}
	uk.PublicKey, err = ssh.NewPublicKey(&privKey.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to extract public key from generated key pair: %w", err)
	}

	// To support Ansible calling back to us we need to write
	// this file down
	privateKeyDer, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return fmt.Errorf("failed to serialise private key for adapter: %w", err)
	}
	privateKeyBlock := &pem.Block{
		Type:    "PRIVATE KEY",
		Headers: nil,
		Bytes:   privateKeyDer,
	}

	err = pem.Encode(target, privateKeyBlock)
	if err != nil {
		return errors.New("failed to write private key to temp file")
	}

	return nil
}

func generateRSAKeyToFile(uk *userKey, target *os.File) error {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return errors.New("Failed to generate key pair")
	}
	uk.PublicKey, err = ssh.NewPublicKey(&privKey.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to extract public key from generated key pair: %w", err)
	}

	privateKeyBlock := &pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   x509.MarshalPKCS1PrivateKey(privKey),
	}

	err = pem.Encode(target, privateKeyBlock)
	if err != nil {
		return errors.New("failed to write private key to temp file")
	}

	return nil
}

type signer struct {
	ssh.Signer
}

func newSigner(privKeyFile string, keyType string) (*signer, error) {
	signer := new(signer)

	if len(privKeyFile) > 0 {
		privateBytes, err := os.ReadFile(privKeyFile)
		if err != nil {
			return nil, errors.New("Failed to load private host key")
		}

		signer.Signer, err = ssh.ParsePrivateKey(privateBytes)
		if err != nil {
			return nil, errors.New("Failed to parse private host key")
		}

		return signer, nil
	}

	var privKey interface{}
	var err error

	switch keyType {
	case "RSA":
		privKey, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, errors.New("Failed to generate server key pair")
		}
	case "ECDSA":
		privKey, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
		if err != nil {
			return nil, errors.New("Failed to generate server key pair")
		}
	default:
		return nil, fmt.Errorf("Unsupported key type: %q", keyType)
	}

	signer.Signer, err = ssh.NewSignerFromKey(privKey)
	if err != nil {
		return nil, errors.New("Failed to extract private key from generated key pair")
	}

	return signer, nil
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

// buildEnvWithPath constructs an environment with modified PATH prepending ansible_navigator_path entries
func buildEnvWithPath(ansibleNavigatorPath []string) []string {
	env := os.Environ()
	if len(ansibleNavigatorPath) == 0 {
		return env
	}

	// Get current PATH
	currentPath := os.Getenv("PATH")

	// Build new PATH with ansible_navigator_path entries prepended
	pathEntries := make([]string, 0, len(ansibleNavigatorPath)+1)
	for _, entry := range ansibleNavigatorPath {
		// Each entry is already HOME-expanded in Prepare()
		pathEntries = append(pathEntries, entry)
	}
	if currentPath != "" {
		pathEntries = append(pathEntries, currentPath)
	}

	newPath := strings.Join(pathEntries, string(os.PathListSeparator))

	// Replace or add PATH in environment
	pathSet := false
	for i, envVar := range env {
		if strings.HasPrefix(envVar, "PATH=") {
			env[i] = "PATH=" + newPath
			pathSet = true
			break
		}
	}
	if !pathSet {
		env = append(env, "PATH="+newPath)
	}

	return env
}

// checkArg Evaluates if argname is in args
func checkArg(argname string, args []string) bool {
	for _, arg := range args {
		for _, ansibleArg := range strings.Split(arg, "=") {
			if ansibleArg == argname {
				return true
			}
		}
	}
	return false
}
