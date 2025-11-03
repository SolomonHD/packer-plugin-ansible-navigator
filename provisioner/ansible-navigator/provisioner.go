// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

//go:generate packer-sdc mapstructure-to-hcl2 -type Config
//go:generate packer-sdc struct-markdown

package ansible

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
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/adapter"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer-plugin-sdk/tmp"
)

// Play represents a single play execution with its configuration
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
	// The command to invoke ansible. Defaults to
	//  `ansible-navigator run`. If you would like to provide a more complex command,
	//  for example, something that sets up a virtual environment before calling
	//  ansible, take a look at the ansible wrapper guide [here](#using-a-wrapping-script-for-your-ansible-call) for inspiration.
	//  Please note that Packer expects Command to be a path to an executable.
	//  Arbitrary bash scripting will not work and needs to go inside an
	//  executable script.
	Command string `mapstructure:"command"`
	// Execution mode for ansible-navigator. Valid values: stdout, json, yaml, interactive.
	// Defaults to "stdout" for non-interactive environments (Packer-safe).
	// When set to "interactive" without a TTY, it automatically switches to "stdout".
	NavigatorMode string `mapstructure:"navigator_mode"`
	// Enable structured JSON parsing and detailed task-level reporting.
	// When true, parses JSON events from ansible-navigator and provides enhanced error reporting.
	// Only effective when navigator_mode is set to "json".
	StructuredLogging bool `mapstructure:"structured_logging"`
	// Optional path to write a structured summary JSON file containing task results and failures.
	// Only used when structured_logging is enabled.
	LogOutputPath string `mapstructure:"log_output_path"`
	// Extra arguments to pass to Ansible. These arguments _will not_ be passed
	// through a shell and arguments should not be quoted. Usage example:
	//
	// ```json
	//    "extra_arguments": [ "--extra-vars", "Region={{user `Region`}} Stage={{user `Stage`}}" ]
	// ```
	//
	// In certain scenarios where you want to pass ansible command line
	// arguments that include parameter and value (for example
	// `--vault-password-file pwfile`), from ansible documentation this is
	// correct format but that is NOT accepted here. Instead you need to do it
	// like `--vault-password-file=pwfile`.
	//
	// If you are running a Windows build on AWS, Azure, Google Compute, or
	// OpenStack and would like to access the auto-generated password that
	// Packer uses to connect to a Windows instance via WinRM, you can use the
	// template variable
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
	//
	// ```hcl
	// extra_arguments = [
	//    "--extra-vars", "winrm_password=${build.Password}"
	// ]
	// ```
	//
	// If the lefthand side of a value contains 'secret' or 'password' (case
	// insensitive) it will be hidden from output. For example, passing
	// "my_password=secr3t" will hide "secr3t" from output.
	ExtraArguments []string `mapstructure:"extra_arguments"`
	// Environment variables to set before
	//   running Ansible. Usage example:
	//
	//   ```json
	//     "ansible_env_vars": [ "ANSIBLE_HOST_KEY_CHECKING=False", "ANSIBLE_SSH_ARGS='-o ForwardAgent=yes -o ControlMaster=auto -o ControlPersist=60s'", "ANSIBLE_NOCOLOR=True" ]
	//   ```
	//
	//   This is a [template engine](/packer/docs/templates/legacy_json_templates/engine). Therefore, you
	//   may use user variables and template functions in this field.
	//
	//   For example, if you are running a Windows build on AWS, Azure,
	//   Google Compute, or OpenStack and would like to access the auto-generated
	//   password that Packer uses to connect to a Windows instance via WinRM, you
	//   can use the template variable `{{.WinRMPassword}}` in this option. Example:
	//
	//   ```json
	//   "ansible_env_vars": [ "WINRM_PASSWORD={{.WinRMPassword}}" ],
	//   ```
	AnsibleEnvVars []string `mapstructure:"ansible_env_vars"`
	// The playbook to be run by Ansible.
	// DEPRECATED: Use plays array instead. Maintained for backward compatibility.
	PlaybookFile string `mapstructure:"playbook_file"`
	// Array of play definitions supporting both playbooks and role FQDNs
	Plays []Play `mapstructure:"plays"`
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
	// Specifies --ssh-extra-args on command line defaults to -o IdentitiesOnly=yes
	AnsibleSSHExtraArgs []string `mapstructure:"ansible_ssh_extra_args"`
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
	// The command to run on the machine being
	//  provisioned by Packer to handle the SFTP protocol that Ansible will use to
	//  transfer files. The command should read and write on stdin and stdout,
	//  respectively. Defaults to `/usr/lib/sftp-server -e`.
	SFTPCmd string `mapstructure:"sftp_command"`
	// Check if ansible is installed prior to
	//  running. Set this to `true`, for example, if you're going to install
	//  ansible during the packer run.
	SkipVersionCheck bool `mapstructure:"skip_version_check"`
	UseSFTP          bool `mapstructure:"use_sftp"`
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
	// If `true`, the Ansible provisioner will
	//  not delete the temporary inventory file it creates in order to connect to
	//  the instance. This is useful if you are trying to debug your ansible run
	//  and using "--on-error=ask" in order to leave your instance running while you
	//  test your playbook. this option is not used if you set an `inventory_file`.
	KeepInventoryFile bool `mapstructure:"keep_inventory_file"`
	// A requirements file which provides a way to
	//  install roles or collections with the [ansible-galaxy
	//  cli](https://docs.ansible.com/ansible/latest/galaxy/user_guide.html#the-ansible-galaxy-command-line-tool)
	//  on the local machine before executing `ansible-playbook`. By default, this is empty.
	GalaxyFile string `mapstructure:"galaxy_file"`
	// The command to invoke ansible-galaxy. By default, this is
	// `ansible-galaxy`.
	GalaxyCommand string `mapstructure:"galaxy_command"`
	// Force overwriting an existing role.
	//  Adds `--force` option to `ansible-galaxy` command. By default, this is
	//  `false`.
	GalaxyForceInstall bool `mapstructure:"galaxy_force_install"`
	// Force overwriting an existing role and its dependencies.
	//  Adds `--force-with-deps` option to `ansible-galaxy` command. By default,
	//  this is `false`.
	GalaxyForceWithDeps bool `mapstructure:"galaxy_force_with_deps"`
	// The path to the directory on your local system in which to
	//   install the roles. Adds `--roles-path /path/to/your/roles` to
	//   `ansible-galaxy` command. By default, this is empty, and thus `--roles-path`
	//   option is not added to the command.
	RolesPath string `mapstructure:"roles_path"`
	// The path to the directory on your local system in which to
	//   install the collections. Adds `--collections-path /path/to/your/collections` to
	//   `ansible-galaxy` command. By default, this is empty, and thus `--collections-path`
	//   option is not added to the command.
	CollectionsPath string `mapstructure:"collections_path"`
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
	// Path to a requirements.yml file for installing collections.
	// This is an alternative to specifying collections inline.
	CollectionsRequirements string `mapstructure:"collections_requirements"`
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
	userWasEmpty bool
}

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

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

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
		p.config.Command = "ansible-navigator run"
	}

	if p.config.GalaxyCommand == "" {
		p.config.GalaxyCommand = "ansible-galaxy"
	}

	if p.config.HostAlias == "" {
		p.config.HostAlias = "default"
	}

	// Set default navigator_mode to stdout for non-interactive environments
	if p.config.NavigatorMode == "" {
		p.config.NavigatorMode = "stdout"
	}

	var errs *packersdk.MultiError

	// Validate navigator_mode
	validModes := map[string]bool{
		"stdout":      true,
		"json":        true,
		"yaml":        true,
		"interactive": true,
	}
	if !validModes[p.config.NavigatorMode] {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("invalid navigator_mode: %s (must be one of stdout, json, yaml, interactive)", p.config.NavigatorMode))
	}

	// Check if interactive mode is requested without TTY
	if p.config.NavigatorMode == "interactive" && !term.IsTerminal(int(os.Stdout.Fd())) {
		log.Printf("[Warning] No TTY detected — switching ansible-navigator mode to 'stdout'.")
		p.config.NavigatorMode = "stdout"
	}

	// Validate dual invocation mode - playbook_file and plays are mutually exclusive
	if p.config.PlaybookFile != "" && len(p.config.Plays) > 0 {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("you may specify only one of `playbook_file` or `plays`"))
	}

	// At least one must be specified
	if p.config.PlaybookFile == "" && len(p.config.Plays) == 0 {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("either `playbook_file` or `plays` must be defined"))
	}

	// Validate playbook_file if specified (deprecated, but supported for backward compatibility)
	if p.config.PlaybookFile != "" {
		ui := &packersdk.BasicUi{
			Reader: os.Stdin,
			Writer: os.Stdout,
		}
		ui.Say("Warning: 'playbook_file' is deprecated. Please use 'plays' array instead.")

		err = validateFileConfig(p.config.PlaybookFile, "playbook_file", true)
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	// Validate plays if specified
	if len(p.config.Plays) > 0 {
		for i, play := range p.config.Plays {
			if play.Target == "" {
				errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("play %d: target must be specified", i))
				continue
			}

			// Validate if target is a playbook file
			if strings.HasSuffix(play.Target, ".yml") || strings.HasSuffix(play.Target, ".yaml") {
				err = validateFileConfig(play.Target, fmt.Sprintf("play %d target", i), true)
				if err != nil {
					errs = packersdk.MultiErrorAppend(errs, err)
				}
			}
			// Role FQDNs don't need file validation
		}
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

	// Check that the galaxy file exists, if configured
	if len(p.config.GalaxyFile) > 0 {
		err = validateFileConfig(p.config.GalaxyFile, "galaxy_file", true)
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	// Check that the authorized key file exists
	if len(p.config.SSHAuthorizedKeyFile) > 0 {
		err = validateFileConfig(p.config.SSHAuthorizedKeyFile, "ssh_authorized_key_file", true)
		if err != nil {
			log.Println(p.config.SSHAuthorizedKeyFile, "does not exist")
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}
	if len(p.config.SSHHostKeyFile) > 0 {
		err = validateFileConfig(p.config.SSHHostKeyFile, "ssh_host_key_file", true)
		if err != nil {
			log.Println(p.config.SSHHostKeyFile, "does not exist")
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	} else {
		p.config.AnsibleEnvVars = append(p.config.AnsibleEnvVars, "ANSIBLE_HOST_KEY_CHECKING=False")
	}

	if !p.config.UseSFTP {
		p.config.AnsibleEnvVars = append(p.config.AnsibleEnvVars, "ANSIBLE_SCP_IF_SSH=True")
	}

	if p.config.LocalPort > 65535 {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("local_port: %d must be a valid port", p.config.LocalPort))
	}

	if len(p.config.InventoryDirectory) > 0 {
		err = validateInventoryDirectoryConfig(p.config.InventoryDirectory)
		if err != nil {
			log.Println(p.config.InventoryDirectory, "does not exist")
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	if !p.config.SkipVersionCheck {
		err = p.getVersion()
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	if p.config.User == "" {
		p.config.userWasEmpty = true
		usr, err := user.Current()
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		} else {
			p.config.User = usr.Username
		}
	}
	if p.config.User == "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("user: could not determine current user from environment."))
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

	switch p.config.AdapterKeyType {
	case "RSA", "ECDSA":
	default:
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
			"Invalid value for ansible_proxy_key_type: %q. Supported values are ECDSA or RSA.",
			p.config.AdapterKeyType))
	}

	if p.config.WinRMUseHTTP {
		addWinRMScheme := true
		for _, arg := range p.config.ExtraArguments {
			if strings.HasPrefix(arg, "ansible_winrm_scheme") {
				addWinRMScheme = false
				log.Printf("ansible_winrm_scheme already defined in arguments, will ignore")
				break
			}
		}
		if addWinRMScheme {
			log.Printf("setting http as winrm scheme")
			p.config.ExtraArguments = append(p.config.ExtraArguments, "-e", "ansible_winrm_scheme=http")
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}
	return nil
}

func (p *Provisioner) getVersion() error {
	// Check if ansible-navigator is available
	command := "ansible-navigator"
	if strings.Contains(p.config.Command, "ansible-navigator") {
		command = "ansible-navigator"
	} else {
		command = p.config.Command
	}

	out, err := exec.Command(command, "--version").Output()
	if err != nil {
		return fmt.Errorf(
			"Error: ansible-navigator not found in PATH. Please install it before running this provisioner: %s", err.Error())
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
		return fmt.Errorf("Could not parse major version from \"%s\".", version)
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
		return "", fmt.Errorf("error creating host signer: %s", err)
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
			l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
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
	tf, err := ioutil.TempFile(p.config.InventoryDirectory, "packer-provisioner-ansible")
	if err != nil {
		return fmt.Errorf("Error preparing inventory file: %s", err)
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
		ctxData["Host"] = "127.0.0.1"
		ctxData["Port"] = p.config.LocalPort
	}
	p.config.ctx.Data = ctxData

	host, err := interpolate.Render(hostTemplate, &p.config.ctx)
	if err != nil {
		return fmt.Errorf("Error generating inventory file from template: %s", err)
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
		return fmt.Errorf("Error preparing inventory file: %s", err)
	}
	tf.Close()
	p.config.InventoryFile = tf.Name()

	return nil
}

func (p *Provisioner) Provision(ctx context.Context, ui packersdk.Ui, comm packersdk.Communicator, generatedData map[string]interface{}) error {
	ui.Say("Provisioning with Ansible Navigator...")

	// Set ANSIBLE_NAVIGATOR_MODE environment variable
	if existingMode := os.Getenv("ANSIBLE_NAVIGATOR_MODE"); existingMode != "" {
		ui.Message(fmt.Sprintf("[Notice] Overriding ANSIBLE_NAVIGATOR_MODE environment variable (was: %s, now: %s)", existingMode, p.config.NavigatorMode))
	}
	os.Setenv("ANSIBLE_NAVIGATOR_MODE", p.config.NavigatorMode)

	// Interpolate env vars to check for generated values like password and port
	p.generatedData = generatedData
	p.config.ctx.Data = generatedData
	for i, envVar := range p.config.AnsibleEnvVars {
		envVar, err := interpolate.Render(envVar, &p.config.ctx)
		if err != nil {
			return fmt.Errorf("Could not interpolate ansible env vars: %s", err)
		}
		p.config.AnsibleEnvVars[i] = envVar
	}
	// Interpolate extra vars to check for generated values like password and port
	for i, arg := range p.config.ExtraArguments {
		arg, err := interpolate.Render(arg, &p.config.ctx)
		if err != nil {
			return fmt.Errorf("Could not interpolate ansible env vars: %s", err)
		}
		p.config.ExtraArguments[i] = arg
	}

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
		return fmt.Errorf("Error executing Ansible: %s", err)
	}

	return nil
}

func (p *Provisioner) executeGalaxy(ui packersdk.Ui, comm packersdk.Communicator) error {
	galaxyFile := filepath.ToSlash(p.config.GalaxyFile)

	// ansible-galaxy install -r requirements.yml
	roleArgs := []string{"install", "-r", galaxyFile}
	// Instead of modifying args depending on config values and removing or modifying values from
	// the slice between role and collection installs, just use 2 slices and simplify everything
	collectionArgs := []string{"collection", "install", "-r", galaxyFile}
	// Add force to arguments
	if p.config.GalaxyForceInstall {
		roleArgs = append(roleArgs, "-f")
		collectionArgs = append(collectionArgs, "-f")
	}
	// Add --force-with-deps to arguments
	if p.config.GalaxyForceWithDeps {
		roleArgs = append(roleArgs, "--force-with-deps")
		collectionArgs = append(collectionArgs, "--force-with-deps")
	}

	// Add roles_path argument if specified
	if p.config.RolesPath != "" {
		roleArgs = append(roleArgs, "-p", filepath.ToSlash(p.config.RolesPath))
	}
	// Add collections_path argument if specified
	if p.config.CollectionsPath != "" {
		collectionArgs = append(collectionArgs, "-p", filepath.ToSlash(p.config.CollectionsPath))
	}

	// Search galaxy_file for roles and collections keywords
	f, err := ioutil.ReadFile(galaxyFile)
	if err != nil {
		return err
	}
	hasRoles, _ := regexp.Match(`(?m)^roles:`, f)
	hasCollections, _ := regexp.Match(`(?m)^collections:`, f)

	// If roles keyword present (v2 format), or no collections keyword present (v1), install roles
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
	ui.Message("Executing Ansible Galaxy")
	cmd := exec.Command(p.config.GalaxyCommand, args...)

	//Setting up AnsibleEnvVars at begining so additional checks can take them into account
	cmd.Env = os.Environ()
	if len(p.config.AnsibleEnvVars) > 0 {
		cmd.Env = append(cmd.Env, p.config.AnsibleEnvVars...)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	wg := sync.WaitGroup{}
	repeat := func(r io.ReadCloser) {
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
		wg.Done()
	}
	wg.Add(2)
	go repeat(stdout)
	go repeat(stderr)

	if err := cmd.Start(); err != nil {
		return err
	}
	wg.Wait()
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("Non-zero exit status: %s", err)
	}
	return nil
}

func (p *Provisioner) createCmdArgs(httpAddr, inventory, playbook, privKeyFile string) (args []string, envVars []string) {
	args = []string{}

	//Setting up AnsibleEnvVars at begining so additional checks can take them into account
	if len(p.config.AnsibleEnvVars) > 0 {
		envVars = append(envVars, p.config.AnsibleEnvVars...)
	}

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
		if len(p.config.AnsibleSSHExtraArgs) > 0 {
			args = append(args, "--ssh-extra-args", fmt.Sprintf("'%s'", strings.Join(p.config.AnsibleSSHExtraArgs, "' '")))
		} else {
			args = append(args, "--ssh-extra-args", "'-o IdentitiesOnly=yes'")
		}
	}

	args = append(args, p.config.ExtraArguments...)

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
		if !checkArg("ansible_host_key_checking", args) && !checkArg("ANSIBLE_HOST_KEY_CHECKING", envVars) {
			args = append(args, "-e", "ansible_host_key_checking=False")
		}
	}
	// This must be the last arg appended to args
	args = append(args, "-i", inventory, playbook)
	return args, envVars
}

// setCollectionsPath sets the ANSIBLE_COLLECTIONS_PATHS environment variable
func (p *Provisioner) setCollectionsPath() error {
	if p.config.CollectionsCacheDir == "" {
		return nil
	}

	// Get existing ANSIBLE_COLLECTIONS_PATHS value
	existingPath := os.Getenv("ANSIBLE_COLLECTIONS_PATHS")

	// Prepend our cache directory to the path
	var newPath string
	if existingPath != "" {
		newPath = p.config.CollectionsCacheDir + ":" + existingPath
	} else {
		newPath = p.config.CollectionsCacheDir
	}

	// Set the environment variable
	if err := os.Setenv("ANSIBLE_COLLECTIONS_PATHS", newPath); err != nil {
		return fmt.Errorf("failed to set ANSIBLE_COLLECTIONS_PATHS: %s", err)
	}

	return nil
}

// setRolesPath sets the ANSIBLE_ROLES_PATH environment variable
func (p *Provisioner) setRolesPath() error {
	if p.config.RolesCacheDir == "" {
		return nil
	}

	// Get existing ANSIBLE_ROLES_PATH value
	existingPath := os.Getenv("ANSIBLE_ROLES_PATH")

	// Prepend our cache directory to the path
	var newPath string
	if existingPath != "" {
		newPath = p.config.RolesCacheDir + ":" + existingPath
	} else {
		newPath = p.config.RolesCacheDir
	}

	// Set the environment variable
	if err := os.Setenv("ANSIBLE_ROLES_PATH", newPath); err != nil {
		return fmt.Errorf("failed to set ANSIBLE_ROLES_PATH: %s", err)
	}

	return nil
}

// generateRolePlaybook creates a temporary playbook file for executing a role
func generateRolePlaybook(role string, play Play) (string, error) {
	// Create a temporary file for the playbook
	tmpFile, err := ioutil.TempFile("", "packer-role-playbook-*.yml")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary playbook file: %s", err)
	}

	// Build the playbook content
	playbookContent := "---\n- hosts: all\n"

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
		return "", fmt.Errorf("failed to write temporary playbook: %s", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to close temporary playbook: %s", err)
	}

	return tmpFile.Name(), nil
}

func (p *Provisioner) executeAnsible(ui packersdk.Ui, comm packersdk.Communicator, privKeyFile string) error {
	httpAddr := p.generatedData["PackerHTTPAddr"].(string)

	// Handle unified requirements file if specified
	if p.config.RequirementsFile != "" {
		ui.Message(fmt.Sprintf("Installing dependencies from unified requirements file: %s", p.config.RequirementsFile))
		if err := p.executeUnifiedRequirements(ui); err != nil {
			return fmt.Errorf("Error installing unified requirements: %s", err)
		}
	} else {
		// Fall back to old collection management if requirements_file not specified
		if err := ensureCollections(ui, &p.config); err != nil {
			return fmt.Errorf("Error managing Ansible collections: %s", err)
		}

		// Fetch external dependencies from galaxy_file if specified
		if len(p.config.GalaxyFile) > 0 {
			if err := p.executeGalaxy(ui, comm); err != nil {
				return fmt.Errorf("Error executing Ansible Galaxy: %s", err)
			}
		}
	}

	// Set environment variables for dependency paths
	if p.config.CollectionsCacheDir != "" {
		if err := p.setCollectionsPath(); err != nil {
			return fmt.Errorf("Error setting collections path: %s", err)
		}
	}

	if p.config.RolesCacheDir != "" {
		if err := p.setRolesPath(); err != nil {
			return fmt.Errorf("Error setting roles path: %s", err)
		}
	}

	if len(p.config.Plays) > 0 {
		// Execute multiple plays
		return p.executePlays(ui, comm, privKeyFile, httpAddr)
	} else {
		// Execute single playbook (backward compatibility)
		return p.executeSinglePlaybook(ui, privKeyFile, httpAddr)
	}
}

// executeUnifiedRequirements installs both roles and collections from a single requirements file
func (p *Provisioner) executeUnifiedRequirements(ui packersdk.Ui) error {
	requirementsPath := p.config.RequirementsFile

	// Validate requirements file exists
	if _, err := os.Stat(requirementsPath); os.IsNotExist(err) {
		return fmt.Errorf("requirements_file not found: %s", requirementsPath)
	}

	// Check offline mode
	if p.config.OfflineMode {
		ui.Message("Offline mode enabled: skipping dependency installation")
		return nil
	}

	// Ensure cache directories exist
	if p.config.CollectionsCacheDir != "" {
		if err := os.MkdirAll(p.config.CollectionsCacheDir, 0755); err != nil {
			return fmt.Errorf("failed to create collections cache directory: %s", err)
		}
	}

	if p.config.RolesCacheDir != "" {
		if err := os.MkdirAll(p.config.RolesCacheDir, 0755); err != nil {
			return fmt.Errorf("failed to create roles cache directory: %s", err)
		}
	}

	// Read the requirements file to determine what's in it
	content, err := ioutil.ReadFile(requirementsPath)
	if err != nil {
		return fmt.Errorf("failed to read requirements file: %s", err)
	}

	hasRoles := regexp.MustCompile(`(?m)^roles:`).Match(content)
	hasCollections := regexp.MustCompile(`(?m)^collections:`).Match(content)

	// Install roles if present
	if hasRoles {
		ui.Message("Installing roles from requirements file...")
		args := []string{"install", "-r", requirementsPath}

		if p.config.RolesCacheDir != "" {
			args = append(args, "-p", p.config.RolesCacheDir)
		}

		if p.config.ForceUpdate {
			args = append(args, "--force")
		}

		if err := p.runGalaxyCommand(ui, args, "roles"); err != nil {
			return fmt.Errorf("failed to install roles: %s", err)
		}
	}

	// Install collections if present
	if hasCollections {
		ui.Message("Installing collections from requirements file...")
		args := []string{"collection", "install", "-r", requirementsPath}

		if p.config.CollectionsCacheDir != "" {
			args = append(args, "-p", p.config.CollectionsCacheDir)
		}

		if p.config.ForceUpdate {
			args = append(args, "--force")
		}

		if err := p.runGalaxyCommand(ui, args, "collections"); err != nil {
			return fmt.Errorf("failed to install collections: %s", err)
		}
	}

	if !hasRoles && !hasCollections {
		ui.Message("Warning: requirements file does not contain 'roles:' or 'collections:' sections")
	}

	return nil
}

// runGalaxyCommand executes an ansible-galaxy command
func (p *Provisioner) runGalaxyCommand(ui packersdk.Ui, args []string, target string) error {
	cmd := exec.Command(p.config.GalaxyCommand, args...)

	cmd.Env = os.Environ()
	if len(p.config.AnsibleEnvVars) > 0 {
		cmd.Env = append(cmd.Env, p.config.AnsibleEnvVars...)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	repeat := func(r io.ReadCloser) {
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
		wg.Done()
	}
	wg.Add(2)
	go repeat(stdout)
	go repeat(stderr)

	if err := cmd.Start(); err != nil {
		return err
	}
	wg.Wait()
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("ansible-galaxy command failed for %s: %s", target, err)
	}

	return nil
}

// executePlays runs multiple plays in sequence
func (p *Provisioner) executePlays(ui packersdk.Ui, comm packersdk.Communicator, privKeyFile string, httpAddr string) error {
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
			tmpPlaybook, err := generateRolePlaybook(play.Target, play)
			if err != nil {
				return fmt.Errorf("Play '%s': failed to generate role playbook: %s", playName, err)
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

		if len(play.Tags) > 0 {
			for _, tag := range play.Tags {
				args = append([]string{"--tags", tag}, args...)
			}
		}

		for k, v := range play.ExtraVars {
			args = append([]string{"-e", fmt.Sprintf("%s=%s", k, v)}, args...)
		}

		for _, varsFile := range play.VarsFiles {
			args = append([]string{"-e", fmt.Sprintf("@%s", varsFile)}, args...)
		}

		// Add --mode flag at the beginning of args
		args = append([]string{"--mode", p.config.NavigatorMode}, args...)

		// Execute the play
		cmd := exec.Command(p.config.Command, args...)
		cmd.Env = os.Environ()
		if len(envvars) > 0 {
			cmd.Env = append(cmd.Env, envvars...)
		}

		err := p.executeAnsibleCommand(ui, cmd, playName)

		// Cleanup temporary playbook if it was generated
		if cleanupFunc != nil {
			cleanupFunc()
		}

		if err != nil {
			ui.Error(fmt.Sprintf("Play '%s' failed: %v", playName, err))
			return fmt.Errorf("Play '%s' failed with exit code 2", playName)
		}

		if i < len(p.config.Plays)-1 {
			ui.Message(fmt.Sprintf("Completed %s", playName))
		}
	}

	ui.Say("All plays completed successfully!")
	return nil
}

// executeSinglePlaybook runs a single playbook (backward compatibility)
func (p *Provisioner) executeSinglePlaybook(ui packersdk.Ui, privKeyFile string, httpAddr string) error {
	playbook, _ := filepath.Abs(p.config.PlaybookFile)
	inventory := p.config.InventoryFile

	args, envvars := p.createCmdArgs(httpAddr, inventory, playbook, privKeyFile)

	// Add --mode flag at the beginning of args
	args = append([]string{"--mode", p.config.NavigatorMode}, args...)

	cmd := exec.Command(p.config.Command, args...)

	cmd.Env = os.Environ()
	if len(envvars) > 0 {
		cmd.Env = append(cmd.Env, envvars...)
	}

	err := p.executeAnsibleCommand(ui, cmd, "playbook execution")
	if err != nil {
		return fmt.Errorf("ansible-navigator run failed for %s: %w", playbook, err)
	}

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
	useStructuredLogging := p.config.StructuredLogging && p.config.NavigatorMode == "json"
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
					ui.Message(fmt.Sprintf("[Warning] Skipped malformed JSON event: %v", err))
					continue
				}
				handleNavigatorEvent(ui, &event, summary)
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

	for _, arg := range p.config.ExtraArguments {
		args := strings.SplitN(arg, "=", 2)
		if len(args) != 2 {
			continue
		}
		if strings.Contains(strings.ToLower(args[0]), "password") ||
			strings.Contains(strings.ToLower(args[0]), "secret") {
			sanitized = strings.Replace(sanitized,
				args[1], "*****", -1)
		}
	}

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
		return fmt.Errorf("Non-zero exit status: %s", err)
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

func validateInventoryDirectoryConfig(name string) error {
	info, err := os.Stat(name)
	if err != nil {
		return fmt.Errorf("inventory_directory: %s is invalid: %s", name, err)
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
		pubKeyBytes, err := ioutil.ReadFile(pubKeyFile)
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
		return fmt.Errorf("Failed to extract public key from generated key pair: %s", err)
	}

	// To support Ansible calling back to us we need to write
	// this file down
	privateKeyDer, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return fmt.Errorf("Failed to serialise private key for adapter: %s", err)
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
		return fmt.Errorf("Failed to extract public key from generated key pair: %s", err)
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
		privateBytes, err := ioutil.ReadFile(privKeyFile)
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
