// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !windows
// +build !windows

package ansiblenavigator

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	confighelper "github.com/hashicorp/packer-plugin-sdk/template/config"
)

// Be sure to remove the Ansible stub file in each test with:
//
//	defer os.Remove(config["command"].(string))
func testConfig(t *testing.T) map[string]interface{} {
	m := make(map[string]interface{})
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	ansible_stub := path.Join(wd, "packer-ansible-stub.sh")

	err = os.WriteFile(ansible_stub, []byte("#!/usr/bin/env bash\necho ansible 1.6.0"), 0777)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	m["command"] = ansible_stub

	return m
}

func TestProvisioner_Impl(t *testing.T) {
	var raw interface{} = &Provisioner{}
	if _, ok := raw.(packersdk.Provisioner); !ok {
		t.Fatalf("must be a Provisioner")
	}
}

func TestProvisionerPrepare_Defaults(t *testing.T) {
	var p Provisioner
	config := testConfig(t)
	defer os.Remove(config["command"].(string))

	err := p.Prepare(config)
	if err == nil {
		t.Fatalf("should have error")
	}

	hostkey_file, err := os.CreateTemp("", "hostkey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(hostkey_file.Name())

	publickey_file, err := os.CreateTemp("", "publickey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(publickey_file.Name())

	playbook_file, err := os.CreateTemp("", "playbook")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbook_file.Name())

	config["ssh_host_key_file"] = hostkey_file.Name()
	config["ssh_authorized_key_file"] = publickey_file.Name()
	config["play"] = []map[string]interface{}{{"target": playbook_file.Name()}}
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbook_file.Name())

	err = os.Unsetenv("USER")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvisionerPrepare_PlaybookFile(t *testing.T) {
	var p Provisioner
	config := testConfig(t)
	defer os.Remove(config["command"].(string))

	hostkey_file, err := os.CreateTemp("", "hostkey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(hostkey_file.Name())

	publickey_file, err := os.CreateTemp("", "publickey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(publickey_file.Name())

	config["ssh_host_key_file"] = hostkey_file.Name()
	config["ssh_authorized_key_file"] = publickey_file.Name()

	err = p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	playbook_file, err := os.CreateTemp("", "playbook")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbook_file.Name())

	config["play"] = []map[string]interface{}{{"target": playbook_file.Name()}}
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvisionerPrepare_DecodesNavigatorConfigExecutionEnvironmentNewFields(t *testing.T) {
	var p Provisioner
	config := testConfig(t)
	defer os.Remove(config["command"].(string))

	hostkeyFile, err := os.CreateTemp("", "hostkey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(hostkeyFile.Name())

	publickeyFile, err := os.CreateTemp("", "publickey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(publickeyFile.Name())

	playbookFile, err := os.CreateTemp("", "playbook")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbookFile.Name())

	config["ssh_host_key_file"] = hostkeyFile.Name()
	config["ssh_authorized_key_file"] = publickeyFile.Name()
	config["play"] = []map[string]interface{}{{"target": playbookFile.Name()}}
	config["navigator_config"] = map[string]interface{}{
		"execution_environment": map[string]interface{}{
			"enabled":           true,
			"container_engine":  "podman",
			"container_options": []string{"--net=host"},
			"pull_arguments":    []string{"--tls-verify=false"},
		},
	}

	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.NavigatorConfig == nil || p.config.NavigatorConfig.ExecutionEnvironment == nil {
		t.Fatalf("expected navigator_config.execution_environment to be decoded")
	}

	ee := p.config.NavigatorConfig.ExecutionEnvironment
	if ee.ContainerEngine != "podman" {
		t.Fatalf("expected ContainerEngine=podman, got: %q", ee.ContainerEngine)
	}
	if len(ee.ContainerOptions) != 1 || ee.ContainerOptions[0] != "--net=host" {
		t.Fatalf("unexpected ContainerOptions: %#v", ee.ContainerOptions)
	}
	if len(ee.PullArguments) != 1 || ee.PullArguments[0] != "--tls-verify=false" {
		t.Fatalf("unexpected PullArguments: %#v", ee.PullArguments)
	}
}

func TestProvisionerPrepare_DecodesNavigatorConfigAnsibleConfigNewSections(t *testing.T) {
	var p Provisioner
	config := testConfig(t)
	defer os.Remove(config["command"].(string))

	hostkeyFile, err := os.CreateTemp("", "hostkey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(hostkeyFile.Name())

	publickeyFile, err := os.CreateTemp("", "publickey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(publickeyFile.Name())

	playbookFile, err := os.CreateTemp("", "playbook")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbookFile.Name())

	config["ssh_host_key_file"] = hostkeyFile.Name()
	config["ssh_authorized_key_file"] = publickeyFile.Name()
	config["play"] = []map[string]interface{}{{"target": playbookFile.Name()}}
	config["navigator_config"] = map[string]interface{}{
		"ansible_config": map[string]interface{}{
			"privilege_escalation": map[string]interface{}{
				"become":        true,
				"become_method": "sudo",
				"become_user":   "root",
			},
			"persistent_connection": map[string]interface{}{
				"connect_timeout":       30,
				"connect_retry_timeout": 15,
				"command_timeout":       60,
			},
			"inventory": map[string]interface{}{
				"enable_plugins": []string{"ini", "yaml"},
			},
			"paramiko_connection": map[string]interface{}{
				"proxy_command": "ssh -W %h:%p jumphost",
			},
			"colors": map[string]interface{}{
				"force_color": true,
			},
			"diff": map[string]interface{}{
				"always":  true,
				"context": 3,
			},
			"galaxy": map[string]interface{}{
				"server_list":  []string{"automation_hub"},
				"ignore_certs": true,
			},
		},
	}

	if err := p.Prepare(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.NavigatorConfig == nil || p.config.NavigatorConfig.AnsibleConfig == nil {
		t.Fatalf("expected navigator_config.ansible_config to be decoded")
	}

	ac := p.config.NavigatorConfig.AnsibleConfig
	if ac.PrivilegeEscalation == nil || !ac.PrivilegeEscalation.Become || ac.PrivilegeEscalation.BecomeMethod != "sudo" {
		t.Fatalf("unexpected privilege_escalation decode: %#v", ac.PrivilegeEscalation)
	}
	if ac.PersistentConnection == nil || ac.PersistentConnection.ConnectTimeout != 30 {
		t.Fatalf("unexpected persistent_connection decode: %#v", ac.PersistentConnection)
	}
	if ac.Inventory == nil || len(ac.Inventory.EnablePlugins) != 2 {
		t.Fatalf("unexpected inventory decode: %#v", ac.Inventory)
	}
	if ac.ParamikoConnection == nil || ac.ParamikoConnection.ProxyCommand == "" {
		t.Fatalf("unexpected paramiko_connection decode: %#v", ac.ParamikoConnection)
	}
	if ac.Colors == nil || !ac.Colors.ForceColor {
		t.Fatalf("unexpected colors decode: %#v", ac.Colors)
	}
	if ac.Diff == nil || !ac.Diff.Always || ac.Diff.Context != 3 {
		t.Fatalf("unexpected diff decode: %#v", ac.Diff)
	}
	if ac.Galaxy == nil || len(ac.Galaxy.ServerList) != 1 || !ac.Galaxy.IgnoreCerts {
		t.Fatalf("unexpected galaxy decode: %#v", ac.Galaxy)
	}
}

func TestProvisionerPrepare_DecodesNavigatorConfigRemainingTopLevelSettings(t *testing.T) {
	var p Provisioner
	config := testConfig(t)
	defer os.Remove(config["command"].(string))

	hostkeyFile, err := os.CreateTemp("", "hostkey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(hostkeyFile.Name())

	publickeyFile, err := os.CreateTemp("", "publickey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(publickeyFile.Name())

	playbookFile, err := os.CreateTemp("", "playbook")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbookFile.Name())

	config["ssh_host_key_file"] = hostkeyFile.Name()
	config["ssh_authorized_key_file"] = publickeyFile.Name()
	config["play"] = []map[string]interface{}{{"target": playbookFile.Name()}}
	config["navigator_config"] = map[string]interface{}{
		"format":                    "yaml",
		"time_zone":                 "America/New_York",
		"inventory_columns":         []string{"name", "address"},
		"collection_doc_cache_path": "/tmp/collection-doc-cache",
		"color":                     map[string]interface{}{"enable": true, "osc4": true},
		"editor":                    map[string]interface{}{"command": "vim", "console": true},
		"images":                    map[string]interface{}{"details": []string{"everything"}},
	}

	if err := p.Prepare(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.NavigatorConfig == nil {
		t.Fatalf("expected navigator_config to be decoded")
	}
	if p.config.NavigatorConfig.Format != "yaml" {
		t.Fatalf("expected Format=yaml, got: %q", p.config.NavigatorConfig.Format)
	}
	if p.config.NavigatorConfig.TimeZone != "America/New_York" {
		t.Fatalf("expected TimeZone=America/New_York, got: %q", p.config.NavigatorConfig.TimeZone)
	}
	if len(p.config.NavigatorConfig.InventoryColumns) != 2 || p.config.NavigatorConfig.InventoryColumns[0] != "name" {
		t.Fatalf("unexpected InventoryColumns: %#v", p.config.NavigatorConfig.InventoryColumns)
	}
	if p.config.NavigatorConfig.CollectionDocCachePath != "/tmp/collection-doc-cache" {
		t.Fatalf("expected CollectionDocCachePath, got: %q", p.config.NavigatorConfig.CollectionDocCachePath)
	}
	if p.config.NavigatorConfig.Color == nil || !p.config.NavigatorConfig.Color.Enable || !p.config.NavigatorConfig.Color.Osc4 {
		t.Fatalf("unexpected Color decode: %#v", p.config.NavigatorConfig.Color)
	}
	if p.config.NavigatorConfig.Editor == nil || p.config.NavigatorConfig.Editor.Command != "vim" || !p.config.NavigatorConfig.Editor.Console {
		t.Fatalf("unexpected Editor decode: %#v", p.config.NavigatorConfig.Editor)
	}
	if p.config.NavigatorConfig.Images == nil || len(p.config.NavigatorConfig.Images.Details) != 1 || p.config.NavigatorConfig.Images.Details[0] != "everything" {
		t.Fatalf("unexpected Images decode: %#v", p.config.NavigatorConfig.Images)
	}
}

func TestProvisionerPrepare_AnsibleConfigPathMutuallyExclusiveWithNewSections(t *testing.T) {
	var p Provisioner
	config := testConfig(t)
	defer os.Remove(config["command"].(string))

	hostkeyFile, err := os.CreateTemp("", "hostkey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(hostkeyFile.Name())

	publickeyFile, err := os.CreateTemp("", "publickey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(publickeyFile.Name())

	playbookFile, err := os.CreateTemp("", "playbook")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbookFile.Name())

	config["ssh_host_key_file"] = hostkeyFile.Name()
	config["ssh_authorized_key_file"] = publickeyFile.Name()
	config["play"] = []map[string]interface{}{{"target": playbookFile.Name()}}
	config["navigator_config"] = map[string]interface{}{
		"ansible_config": map[string]interface{}{
			"config": "/etc/ansible/ansible.cfg",
			"galaxy": map[string]interface{}{
				"server_list": []string{"automation_hub"},
			},
		},
	}

	err = p.Prepare(config)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("expected mutually exclusive error, got: %v", err)
	}
}

func TestProvisionerPrepare_HostKeyFile(t *testing.T) {
	var p Provisioner
	config := testConfig(t)
	defer os.Remove(config["command"].(string))

	publickey_file, err := os.CreateTemp("", "publickey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(publickey_file.Name())

	playbook_file, err := os.CreateTemp("", "playbook")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbook_file.Name())

	filename := make([]byte, 10)
	n, err := io.ReadFull(rand.Reader, filename)
	if n != len(filename) || err != nil {
		t.Fatal("could not create random file name")
	}

	config["ssh_host_key_file"] = fmt.Sprintf("%x", filename)
	config["ssh_authorized_key_file"] = publickey_file.Name()
	config["play"] = []map[string]interface{}{{"target": playbook_file.Name()}}

	err = p.Prepare(config)
	if err == nil {
		t.Fatal("should error if ssh_host_key_file does not exist")
	}

	hostkey_file, err := os.CreateTemp("", "hostkey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(hostkey_file.Name())

	config["ssh_host_key_file"] = hostkey_file.Name()
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvisionerPrepare_AuthorizedKeyFile(t *testing.T) {
	var p Provisioner
	config := testConfig(t)
	defer os.Remove(config["command"].(string))

	hostkey_file, err := os.CreateTemp("", "hostkey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(hostkey_file.Name())

	playbook_file, err := os.CreateTemp("", "playbook")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbook_file.Name())

	filename := make([]byte, 10)
	n, err := io.ReadFull(rand.Reader, filename)
	if n != len(filename) || err != nil {
		t.Fatal("could not create random file name")
	}

	config["ssh_host_key_file"] = hostkey_file.Name()
	config["play"] = []map[string]interface{}{{"target": playbook_file.Name()}}
	config["ssh_authorized_key_file"] = fmt.Sprintf("%x", filename)

	err = p.Prepare(config)
	if err == nil {
		t.Errorf("should error if ssh_authorized_key_file does not exist")
	}

	publickey_file, err := os.CreateTemp("", "publickey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(publickey_file.Name())

	config["ssh_authorized_key_file"] = publickey_file.Name()
	err = p.Prepare(config)
	if err != nil {
		t.Errorf("err: %s", err)
	}
}

func TestProvisionerPrepare_LocalPort(t *testing.T) {
	var p Provisioner
	config := testConfig(t)
	defer os.Remove(config["command"].(string))

	hostkey_file, err := os.CreateTemp("", "hostkey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(hostkey_file.Name())

	publickey_file, err := os.CreateTemp("", "publickey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(publickey_file.Name())

	playbook_file, err := os.CreateTemp("", "playbook")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbook_file.Name())

	config["ssh_host_key_file"] = hostkey_file.Name()
	config["ssh_authorized_key_file"] = publickey_file.Name()
	config["play"] = []map[string]interface{}{{"target": playbook_file.Name()}}

	config["local_port"] = 65537
	err = p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	config["local_port"] = 22222
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvisionerPrepare_InventoryDirectory(t *testing.T) {
	var p Provisioner
	config := testConfig(t)
	defer os.Remove(config["command"].(string))

	hostkey_file, err := os.CreateTemp("", "hostkey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(hostkey_file.Name())

	publickey_file, err := os.CreateTemp("", "publickey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(publickey_file.Name())

	playbook_file, err := os.CreateTemp("", "playbook")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbook_file.Name())

	config["ssh_host_key_file"] = hostkey_file.Name()
	config["ssh_authorized_key_file"] = publickey_file.Name()
	config["play"] = []map[string]interface{}{{"target": playbook_file.Name()}}

	config["inventory_directory"] = "doesnotexist"
	err = p.Prepare(config)
	if err == nil {
		t.Errorf("should error if inventory_directory does not exist")
	}

	inventoryDirectory, err := os.MkdirTemp("", "some_inventory_dir")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(inventoryDirectory)

	config["inventory_directory"] = inventoryDirectory
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestCreateInventoryFile(t *testing.T) {
	type inventoryFileTestCases struct {
		AnsibleVersion uint
		User           string
		Groups         []string
		EmptyGroups    []string
		UseProxy       confighelper.Trilean
		GeneratedData  map[string]interface{}
		Expected       string
	}

	TestCases := []inventoryFileTestCases{
		{
			AnsibleVersion: 1,
			User:           "testuser",
			UseProxy:       confighelper.TriFalse,
			GeneratedData:  basicGenData(nil),
			Expected:       "default ansible_ssh_host=123.45.67.89 ansible_ssh_user=testuser ansible_ssh_port=1234\n",
		},
		{
			AnsibleVersion: 2,
			User:           "testuser",
			UseProxy:       confighelper.TriFalse,
			GeneratedData:  basicGenData(nil),
			Expected:       "default ansible_host=123.45.67.89 ansible_user=testuser ansible_port=1234\n",
		},
		{
			AnsibleVersion: 1,
			User:           "testuser",
			Groups:         []string{"Group1", "Group2"},
			UseProxy:       confighelper.TriFalse,
			GeneratedData:  basicGenData(nil),
			Expected: `default ansible_ssh_host=123.45.67.89 ansible_ssh_user=testuser ansible_ssh_port=1234
[Group1]
default ansible_ssh_host=123.45.67.89 ansible_ssh_user=testuser ansible_ssh_port=1234
[Group2]
default ansible_ssh_host=123.45.67.89 ansible_ssh_user=testuser ansible_ssh_port=1234
`,
		},
		{
			AnsibleVersion: 1,
			User:           "testuser",
			EmptyGroups:    []string{"Group1", "Group2"},
			UseProxy:       confighelper.TriFalse,
			GeneratedData:  basicGenData(nil),
			Expected: `default ansible_ssh_host=123.45.67.89 ansible_ssh_user=testuser ansible_ssh_port=1234
[Group1]
[Group2]
`,
		},
		{
			AnsibleVersion: 1,
			User:           "testuser",
			Groups:         []string{"Group1", "Group2"},
			EmptyGroups:    []string{"Group3"},
			UseProxy:       confighelper.TriFalse,
			GeneratedData:  basicGenData(nil),
			Expected: `default ansible_ssh_host=123.45.67.89 ansible_ssh_user=testuser ansible_ssh_port=1234
[Group1]
default ansible_ssh_host=123.45.67.89 ansible_ssh_user=testuser ansible_ssh_port=1234
[Group2]
default ansible_ssh_host=123.45.67.89 ansible_ssh_user=testuser ansible_ssh_port=1234
[Group3]
`,
		},
		{
			AnsibleVersion: 2,
			User:           "testuser",
			UseProxy:       confighelper.TriFalse,
			GeneratedData: basicGenData(map[string]interface{}{
				"ConnType": "winrm",
				"Password": "12345",
			}),
			Expected: "default ansible_host=123.45.67.89 ansible_connection=winrm ansible_winrm_transport=basic ansible_shell_type=powershell ansible_user=testuser ansible_port=1234\n",
		},
	}

	for _, tc := range TestCases {
		var p Provisioner
		err := p.Prepare(testConfig(t))
		if err == nil {
			t.Fatalf("should have error")
		}
		defer os.Remove(p.config.Command)
		p.ansibleMajVersion = tc.AnsibleVersion
		p.config.User = tc.User
		p.config.Groups = tc.Groups
		p.config.EmptyGroups = tc.EmptyGroups
		p.config.UseProxy = tc.UseProxy
		p.generatedData = tc.GeneratedData

		err = p.createInventoryFile()
		if err != nil {
			t.Fatalf("error creating config using localhost and local port proxy")
		}
		if p.config.InventoryFile == "" {
			t.Fatalf("No inventory file was created")
		}
		defer os.Remove(p.config.InventoryFile)
		f, err := os.ReadFile(p.config.InventoryFile)
		if err != nil {
			t.Fatalf("couldn't read created inventoryfile: %s", err)
		}

		expected := tc.Expected
		if string(f) != expected {
			t.Fatalf("File didn't match expected:\n\n expected: \n%s\n; recieved: \n%s\n", expected, f)
		}
	}
}

func basicGenData(input map[string]interface{}) map[string]interface{} {
	gd := map[string]interface{}{
		"Host":              "123.45.67.89",
		"Port":              int64(1234),
		"ConnType":          "ssh",
		"SSHPrivateKeyFile": "",
		"SSHPrivateKey":     "asdf",
		"SSHAgentAuth":      false,
		"User":              "PartyPacker",
		"PackerHTTPAddr":    commonsteps.HttpAddrNotImplemented,
		"PackerHTTPIP":      commonsteps.HttpIPNotImplemented,
		"PackerHTTPPort":    commonsteps.HttpPortNotImplemented,
	}
	if input == nil {
		return gd
	}
	for k, v := range input {
		gd[k] = v
	}
	return gd
}

func TestUseProxy(t *testing.T) {
	type testcase struct {
		UseProxy                   confighelper.Trilean
		generatedData              map[string]interface{}
		expectedSetupAdapterCalled bool
		explanation                string
	}

	tcs := []testcase{
		{
			explanation:                "use_proxy is true; we should set up adapter",
			UseProxy:                   confighelper.TriTrue,
			generatedData:              basicGenData(nil),
			expectedSetupAdapterCalled: true,
		},
		{
			explanation: "use_proxy is false but no IP addr is available; we should set up adapter anyway.",
			UseProxy:    confighelper.TriFalse,
			generatedData: basicGenData(map[string]interface{}{
				"Host": "",
				"Port": nil,
			}),
			expectedSetupAdapterCalled: true,
		},
		{
			explanation:                "use_proxy is false; we shouldn't set up adapter.",
			UseProxy:                   confighelper.TriFalse,
			generatedData:              basicGenData(nil),
			expectedSetupAdapterCalled: false,
		},
		{
			explanation: "use_proxy is false but connType isn't ssh or winrm.",
			UseProxy:    confighelper.TriFalse,
			generatedData: basicGenData(map[string]interface{}{
				"ConnType": "docker",
			}),
			expectedSetupAdapterCalled: true,
		},
		{
			explanation:                "use_proxy is unset; we should default to setting up the adapter (for now).",
			UseProxy:                   confighelper.TriUnset,
			generatedData:              basicGenData(nil),
			expectedSetupAdapterCalled: true,
		},
		{
			explanation: "use_proxy is false and connType is winRM. we should not set up the adapter.",
			UseProxy:    confighelper.TriFalse,
			generatedData: basicGenData(map[string]interface{}{
				"ConnType": "winrm",
			}),
			expectedSetupAdapterCalled: false,
		},
		{
			explanation: "use_proxy is unset and connType is winRM. we should set up the adapter.",
			UseProxy:    confighelper.TriUnset,
			generatedData: basicGenData(map[string]interface{}{
				"ConnType": "winrm",
			}),
			expectedSetupAdapterCalled: true,
		},
	}

	for _, tc := range tcs {
		var p Provisioner
		err := p.Prepare(testConfig(t))
		if err == nil {
			t.Fatalf("%s should have error", tc.explanation)
		}
		p.config.UseProxy = tc.UseProxy
		defer os.Remove(p.config.Command)
		p.ansibleMajVersion = 1

		var l provisionLogicTracker
		l.setupAdapterCalled = false
		p.setupAdapterFunc = l.setupAdapter
		p.executeAnsibleFunc = l.executeAnsible
		ctx := context.TODO()
		comm := new(packersdk.MockCommunicator)
		ui := &packersdk.BasicUi{
			Reader: new(bytes.Buffer),
			Writer: new(bytes.Buffer),
		}
		//nolint:errcheck
		p.Provision(ctx, ui, comm, tc.generatedData)

		if l.setupAdapterCalled != tc.expectedSetupAdapterCalled {
			t.Fatalf("%s", tc.explanation)
		}
		os.Remove(p.config.Command)
	}
}

func TestProvisioner_WarnsOnSkipVersionCheckWithExplicitTimeout(t *testing.T) {
	var p Provisioner
	config := testConfig(t)
	defer os.Remove(config["command"].(string))

	hostkeyFile, err := os.CreateTemp("", "hostkey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(hostkeyFile.Name())

	publickeyFile, err := os.CreateTemp("", "publickey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(publickeyFile.Name())

	playbookFile, err := os.CreateTemp("", "playbook")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbookFile.Name())

	config["ssh_host_key_file"] = hostkeyFile.Name()
	config["ssh_authorized_key_file"] = publickeyFile.Name()
	config["play"] = []map[string]interface{}{{"target": playbookFile.Name()}}
	config["skip_version_check"] = true
	config["version_check_timeout"] = "10s"

	if err := p.Prepare(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Avoid exercising real SSH proxy setup / ansible execution
	p.setupAdapterFunc = func(ui packersdk.Ui, comm packersdk.Communicator) (string, error) { return "", nil }
	p.executeAnsibleFunc = func(ui packersdk.Ui, comm packersdk.Communicator, privKeyFile string) error { return nil }
	p.config.UseProxy = confighelper.TriFalse

	out := new(bytes.Buffer)
	ui := &packersdk.BasicUi{Reader: new(bytes.Buffer), Writer: out}
	comm := new(packersdk.MockCommunicator)

	gd := basicGenData(map[string]interface{}{"SSHPrivateKeyFile": "/dev/null"})
	if err := p.Provision(context.Background(), ui, comm, gd); err != nil {
		t.Fatalf("err: %s", err)
	}

	if !strings.Contains(out.String(), "Warning: version_check_timeout is ignored when skip_version_check=true") {
		t.Fatalf("expected warning in UI output; got: %q", out.String())
	}
}

func TestProvisioner_DoesNotWarnOnSkipVersionCheckWithoutExplicitTimeout(t *testing.T) {
	var p Provisioner
	config := testConfig(t)
	defer os.Remove(config["command"].(string))

	hostkeyFile, err := os.CreateTemp("", "hostkey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(hostkeyFile.Name())

	publickeyFile, err := os.CreateTemp("", "publickey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(publickeyFile.Name())

	playbookFile, err := os.CreateTemp("", "playbook")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbookFile.Name())

	config["ssh_host_key_file"] = hostkeyFile.Name()
	config["ssh_authorized_key_file"] = publickeyFile.Name()
	config["play"] = []map[string]interface{}{{"target": playbookFile.Name()}}
	config["skip_version_check"] = true
	// Do not set version_check_timeout; Prepare defaults it.

	if err := p.Prepare(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	p.setupAdapterFunc = func(ui packersdk.Ui, comm packersdk.Communicator) (string, error) { return "", nil }
	p.executeAnsibleFunc = func(ui packersdk.Ui, comm packersdk.Communicator, privKeyFile string) error { return nil }
	p.config.UseProxy = confighelper.TriFalse

	out := new(bytes.Buffer)
	ui := &packersdk.BasicUi{Reader: new(bytes.Buffer), Writer: out}
	comm := new(packersdk.MockCommunicator)

	gd := basicGenData(map[string]interface{}{"SSHPrivateKeyFile": "/dev/null"})
	if err := p.Provision(context.Background(), ui, comm, gd); err != nil {
		t.Fatalf("err: %s", err)
	}

	if strings.Contains(out.String(), "version_check_timeout is ignored") {
		t.Fatalf("did not expect warning in UI output; got: %q", out.String())
	}
}

func TestProvisioner_DoesNotWarnWhenSkipVersionCheckIsFalse(t *testing.T) {
	var p Provisioner
	config := testConfig(t)
	defer os.Remove(config["command"].(string))

	hostkeyFile, err := os.CreateTemp("", "hostkey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(hostkeyFile.Name())

	publickeyFile, err := os.CreateTemp("", "publickey")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(publickeyFile.Name())

	playbookFile, err := os.CreateTemp("", "playbook")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbookFile.Name())

	config["ssh_host_key_file"] = hostkeyFile.Name()
	config["ssh_authorized_key_file"] = publickeyFile.Name()
	config["play"] = []map[string]interface{}{{"target": playbookFile.Name()}}
	config["skip_version_check"] = false
	config["version_check_timeout"] = "10s"

	if err := p.Prepare(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	p.setupAdapterFunc = func(ui packersdk.Ui, comm packersdk.Communicator) (string, error) { return "", nil }
	p.executeAnsibleFunc = func(ui packersdk.Ui, comm packersdk.Communicator, privKeyFile string) error { return nil }
	p.config.UseProxy = confighelper.TriFalse

	out := new(bytes.Buffer)
	ui := &packersdk.BasicUi{Reader: new(bytes.Buffer), Writer: out}
	comm := new(packersdk.MockCommunicator)

	gd := basicGenData(map[string]interface{}{"SSHPrivateKeyFile": "/dev/null"})
	if err := p.Provision(context.Background(), ui, comm, gd); err != nil {
		t.Fatalf("err: %s", err)
	}

	if strings.Contains(out.String(), "version_check_timeout is ignored") {
		t.Fatalf("did not expect warning in UI output; got: %q", out.String())
	}
}
