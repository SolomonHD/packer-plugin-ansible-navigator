// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

//go:generate packer-sdc mapstructure-to-hcl2 -type Config,Play,PathEntry
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

	// Repeated play blocks supporting both playbooks and role FQDNs
	Plays []Play `mapstructure:"play"`
	// Path to a unified requirements.yml file containing both roles and collections
	RequirementsFile string `mapstructure:"requirements_file"`
	// Directory to cache downloaded roles. Similar to collections_cache_dir but for roles.
	// Defaults to ~/.packer.d/ansible_roles_cache if not specified.
	RolesCacheDir string `mapstructure:"roles_cache_dir"`
	// When true, skip network operations for both collections and roles.
	// Uses only locally cached dependencies.
	OfflineMode bool `mapstructure:"offline_mode"`
	// When true, always reinstall both collections and roles even if cached.
	ForceUpdate bool `mapstructure:"force_update"`

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
	VersionCheckTimeout string `mapstructure:"version_check_timeout"`
	UseSFTP             bool   `mapstructure:"use_sftp"`
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

	// Force overwriting an existing role.
	//  Adds `--force` option to `ansible-galaxy` command. By default, this is
	//  `false`.
	GalaxyForceInstall bool `mapstructure:"galaxy_force_install"`
	// Force overwriting an existing role and its dependencies.
	//  Adds `--force-with-deps` option to `ansible-galaxy` command. By default,
	//  this is `false`.
	GalaxyForceWithDeps bool `mapstructure:"galaxy_force_with_deps"`

	// Directory to cache downloaded collections.
	// Defaults to ~/.packer.d/ansible_collections_cache if not specified.
	CollectionsCacheDir string `mapstructure:"collections_cache_dir"`
	// When `true`, set up a localhost proxy adapter
	// so that Ansible has an IP address to connect to, even if your guest does not
	// have an IP address. For example, the adapter is necessary for Docker builds
	// to use the Ansible provisioner. If you set this option to `false`, but
	// Packer cannot find an IP address to connect Ansible to, it will
	// automatically set up the adapter anyway.
	//
	//  In order for Ansible to connect properly even when use_proxy is false, you
	// need to make sure that you are either providing a valid username and ssh key
	// to the ansible provisioner directly, or that the username and ssh key
	// being used by the ssh communicator will work for your needs. If you do not
	// provide a user to ansible, it will use the user associated with your
	// builder, not the user running Packer.
	//  use_proxy=false is currently only supported for SSH and WinRM.
	//
	// Currently, this defaults to `true` for all connection types. In the future,
	// this option will be changed to default to `false` for SSH and WinRM
	// connections where the provisioner has access to a host IP.
	UseProxy config.Trilean `mapstructure:"use_proxy"`
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

	// Modern declarative ansible-navigator configuration via YAML file generation.
	// Maps directly to ansible-navigator.yml schema structure.
	// Supports full ansible-navigator.yml structure including:
	//   - ansible section (config overrides, playbook settings)
	//   - execution-environment object (enabled, image, pull-policy, environment-variables)
	//   - mode (stdout, json, yaml, interactive)
	//   - All other ansible-navigator.yml options
	// When provided:
	//   - Plugin generates temporary ansible-navigator.yml file
	//   - Sets ANSIBLE_NAVIGATOR_CONFIG environment variable
	//   - Automatically sets EE temp dir defaults when execution-environment.enabled = true
	//   - Cleans up config file after execution
	// When both navigator_config and legacy options present, navigator_config takes precedence.
	NavigatorConfig map[string]interface{} `mapstructure:"navigator_config"`
	userWasEmpty    bool
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
	if c.VersionCheckTimeout != "" {
		if _, err := time.ParseDuration(c.VersionCheckTimeout); err != nil {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
				"invalid version_check_timeout: %q (must be a valid duration like '30s', '1m', '90s'): %w",
				c.VersionCheckTimeout, err))
		}
	}

	// Validate navigator_config
	if c.NavigatorConfig != nil && len(c.NavigatorConfig) == 0 {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
			"navigator_config cannot be an empty map. Either provide configuration or omit the field"))
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
	p.config.CollectionsCacheDir = expandUserPath(p.config.CollectionsCacheDir)
	p.config.RolesCacheDir = expandUserPath(p.config.RolesCacheDir)
	p.config.WorkDir = expandUserPath(p.config.WorkDir)

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

	// Set default version check timeout
	if p.config.VersionCheckTimeout == "" {
		p.config.VersionCheckTimeout = "60s"
	}

	// Set default cache directories if not specified
	if p.config.CollectionsCacheDir == "" {
		usr, err := user.Current()
		if err == nil {
			p.config.CollectionsCacheDir = filepath.Join(usr.HomeDir, ".packer.d", "ansible_collections_cache")
		}
	}

	if p.config.RolesCacheDir == "" {
		usr, err := user.Current()
		if err == nil {
			p.config.RolesCacheDir = filepath.Join(usr.HomeDir, ".packer.d", "ansible_roles_cache")
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

func (p *Provisioner) getVersion() error {
	// Use the configured command (defaults to "ansible-navigator")
	command := p.config.Command

	// Parse timeout duration
	timeout, err := time.ParseDuration(p.config.VersionCheckTimeout)
	if err != nil {
		// This should never happen as we validate in Prepare(), but handle gracefully
		timeout = 60 * time.Second
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
				"ansible-navigator version check timed out after %s. "+
					"This may indicate ansible-navigator is not properly installed or configured. "+
					"Solutions:\n"+
					"  1. Ensure ansible-navigator is installed and in PATH\n"+
					"  2. Use 'ansible_navigator_path' to specify additional directories\n"+
					"  3. Use 'skip_version_check = true' to bypass the check\n"+
					"  4. Increase 'version_check_timeout' (current: %s)",
				p.config.VersionCheckTimeout, p.config.VersionCheckTimeout)
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
		if p.config.UseProxy.False() && p.generatedData["ConnType"] == "winrm" {
			hostTemplate = DefaultWinRMInventoryFilev2
		}
	}

	// interpolate template to generate host with necessary vars.
	ctxData := p.generatedData
	ctxData["HostAlias"] = p.config.HostAlias
	ctxData["User"] = p.config.User
	if !p.config.UseProxy.False() {
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

	p.generatedData = generatedData
	p.config.ctx.Data = generatedData

	// Set up proxy if host IP is missing or communicator type is wrong.
	if p.config.UseProxy.False() {
		hostIP, ok := generatedData["Host"].(string)
		if !ok || hostIP == "" {
			ui.Error("Warning: use_proxy is false, but instance does" +
				" not have an IP address to give to Ansible. Falling back" +
				" to use localhost proxy.")
			p.config.UseProxy = config.TriTrue
		}
		connType := generatedData["ConnType"]
		if connType != "ssh" && connType != "winrm" {
			ui.Error("Warning: use_proxy is false, but communicator is " +
				"neither ssh nor winrm, so without the proxy ansible will not" +
				" function. Falling back to localhost proxy.")
			p.config.UseProxy = config.TriTrue
		}
	}

	privKeyFile := ""
	if !p.config.UseProxy.False() {
		// We set up the proxy if useProxy is either true or unset.
		pkf, err := p.setupAdapterFunc(ui, comm)
		if err != nil {
			return err
		}
		// This is necessary to avoid accidentally redeclaring
		// privKeyFile in the scope of this if statement.
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
	} else {
		connType := generatedData["ConnType"].(string)
		switch connType {
		case "ssh":
			ui.Message("Not using Proxy adapter for Ansible run:\n" +
				"\tUsing ssh keys from Packer communicator...")
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
			ui.Message("Not using Proxy adapter for Ansible run:\n" +
				"\tUsing WinRM Password from Packer communicator...")
		}
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

func (p *Provisioner) createCmdArgs(httpAddr, inventory, playbook, privKeyFile string) (args []string, envVars []string) {
	args = []string{}

	if p.config.PackerBuildName != "" {
		// HCL configs don't currently have the PakcerBuildName. Don't
		// cause weirdness with a half-set variable
		args = append(args, "-e", fmt.Sprintf("packer_build_name=%q", p.config.PackerBuildName))
	}

	args = append(args, "-e", fmt.Sprintf("packer_builder_type=%s", p.config.PackerBuilderType))

	// expose packer_http_addr extra variable
	if httpAddr != commonsteps.HttpAddrNotImplemented {
		args = append(args, "-e", fmt.Sprintf("packer_http_addr=%s", httpAddr))
	}

	if p.generatedData["ConnType"] == "ssh" && len(privKeyFile) > 0 {
		// Add ssh extra args to set IdentitiesOnly
		args = append(args, "--ssh-extra-args", "'-o IdentitiesOnly=yes'")
	}

	// Add limit if specified
	if p.config.Limit != "" {
		args = append(args, "--limit", p.config.Limit)
	}

	// Add password to ansible call.
	if !checkArg("ansible_password", args) && p.config.UseProxy.False() && p.generatedData["ConnType"] == "winrm" {
		args = append(args, "-e", fmt.Sprintf("ansible_password=%s", p.generatedData["Password"]))
	}

	if !checkArg("ansible_password", args) && len(privKeyFile) > 0 {
		// "-e ansible_ssh_private_key_file" is preferable to "--private-key"
		// because it is a higher priority variable and therefore won't get
		// overridden by dynamic variables. See #5852 for more details.
		args = append(args, "-e", fmt.Sprintf("ansible_ssh_private_key_file=%s", privKeyFile))
	}

	if checkArg("ansible_password", args) && p.generatedData["ConnType"] == "ssh" {
		if !checkArg("ansible_host_key_checking", args) {
			args = append(args, "-e", "ansible_host_key_checking=False")
		}
	}
	// This must be the last arg appended to args
	args = append(args, "-i", inventory, playbook)
	return args, envVars
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

	// Generate and setup ansible-navigator.yml if configured
	var navigatorConfigPath string
	if p.config.NavigatorConfig != nil && len(p.config.NavigatorConfig) > 0 {
		yamlContent, err := generateNavigatorConfigYAML(p.config.NavigatorConfig)
		if err != nil {
			return fmt.Errorf("failed to generate navigator_config YAML: %w", err)
		}

		navigatorConfigPath, err = createNavigatorConfigFile(yamlContent)
		if err != nil {
			return fmt.Errorf("failed to create temporary ansible-navigator.yml: %w", err)
		}

		ui.Message(fmt.Sprintf("Generated ansible-navigator.yml at %s", navigatorConfigPath))

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

// executePlays executes multiple Ansible plays in sequence.
// If a play fails and keep_going is false, execution stops immediately.
// Otherwise, errors are logged and execution continues to the next play.
func (p *Provisioner) executePlays(ui packersdk.Ui, comm packersdk.Communicator, privKeyFile string, httpAddr string, navigatorConfigPath string) error {
	inventory := p.config.InventoryFile

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
		} else {
			// It's a role - generate a temporary playbook
			ui.Message(fmt.Sprintf("Generating temporary playbook for role: %s", play.Target))
			tmpPlaybook, err := createRolePlaybook(play.Target, play)
			if err != nil {
				return fmt.Errorf("play %q: failed to generate role playbook: %w", playName, err)
			}
			playbookPath = tmpPlaybook
			cleanupFunc = func() {
				os.Remove(tmpPlaybook)
			}
		}

		// Build command arguments
		args, envvars := p.createCmdArgs(httpAddr, inventory, playbookPath, privKeyFile)

		// Add per-play arguments
		if play.Become && !checkArg("--become", args) {
			args = append([]string{"--become"}, args...)
		}

		if play.BecomeUser != "" && !checkArg("--become-user", args) {
			args = append([]string{"--become-user", play.BecomeUser}, args...)
		}

		if len(play.Tags) > 0 {
			for _, tag := range play.Tags {
				args = append([]string{"--tags", tag}, args...)
			}
		}

		if len(play.SkipTags) > 0 {
			for _, tag := range play.SkipTags {
				args = append([]string{"--skip-tags", tag}, args...)
			}
		}

		for k, v := range play.ExtraVars {
			args = append([]string{"-e", fmt.Sprintf("%s=%s", k, v)}, args...)
		}

		for _, varsFile := range play.VarsFiles {
			args = append([]string{"-e", fmt.Sprintf("@%s", varsFile)}, args...)
		}

		// Execute the play - prepend "run" as first argument
		cmdArgs := append([]string{"run"}, args...)
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
		}
		if len(envvars) > 0 {
			cmd.Env = append(cmd.Env, envvars...)
		}

		// Set working directory if specified
		if p.config.WorkDir != "" {
			cmd.Dir = p.config.WorkDir
		}

		err := p.executeAnsibleCommand(ui, cmd, playName)

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
	tmpFile, err := os.CreateTemp("", "packer-ansible-cfg-*.ini")
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
