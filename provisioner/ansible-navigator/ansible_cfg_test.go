//go:build !windows
// +build !windows

package ansiblenavigator

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	confighelper "github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/stretchr/testify/require"
)

func TestGenerateAnsibleCfg_Simple(t *testing.T) {
	sections := map[string]map[string]string{
		"defaults": {
			"remote_tmp": "/tmp/.ansible/tmp",
		},
	}

	got, err := generateAnsibleCfg(sections)
	require.NoError(t, err)
	require.Equal(t, "[defaults]\nremote_tmp = /tmp/.ansible/tmp\n\n", got)
}

func TestGenerateAnsibleCfg_MultiSection(t *testing.T) {
	sections := map[string]map[string]string{
		"ssh_connection": {
			"pipelining": "True",
			"ssh_args":   "-o ControlMaster=auto -o ControlPersist=60s",
		},
		"defaults": {
			"timeout":    "30",
			"remote_tmp": "/tmp/.ansible/tmp",
		},
	}

	got, err := generateAnsibleCfg(sections)
	require.NoError(t, err)

	// Sections and keys are sorted for deterministic output.
	expected := strings.Join([]string{
		"[defaults]",
		"remote_tmp = /tmp/.ansible/tmp",
		"timeout = 30",
		"",
		"[ssh_connection]",
		"pipelining = True",
		"ssh_args = -o ControlMaster=auto -o ControlPersist=60s",
		"",
		"",
	}, "\n")

	require.Equal(t, expected, got)
}

func TestGenerateAnsibleCfg_SpecialCharactersPreserved(t *testing.T) {
	sections := map[string]map[string]string{
		"defaults": {
			"ssh_args": "-o ProxyCommand=ssh -W %h:%p jumphost",
			"foo":      "bar=baz",
		},
	}

	got, err := generateAnsibleCfg(sections)
	require.NoError(t, err)
	require.Contains(t, got, "ssh_args = -o ProxyCommand=ssh -W %h:%p jumphost\n")
	require.Contains(t, got, "foo = bar=baz\n")
}

func TestGenerateAnsibleCfg_EmptyMapErrors(t *testing.T) {
	_, err := generateAnsibleCfg(map[string]map[string]string{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "ansible_cfg cannot be empty")
}

func TestConfigValidate_AnsibleCfgEmptyMapErrors(t *testing.T) {
	playbookFile, err := os.CreateTemp("", "playbook")
	require.NoError(t, err)
	defer os.Remove(playbookFile.Name())

	cfg := &Config{
		PlaybookFile: playbookFile.Name(),
		AnsibleCfg:   map[string]map[string]string{},
	}

	err = cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "ansible_cfg cannot be an empty map")
}

func TestProvisionerPrepare_AppliesAnsibleCfgDefaultsWhenExecutionEnvironmentSet(t *testing.T) {
	var p Provisioner
	config := testConfig(t)
	defer os.Remove(config["command"].(string))

	playbookFile, err := os.CreateTemp("", "playbook")
	require.NoError(t, err)
	defer os.Remove(playbookFile.Name())

	config["playbook_file"] = playbookFile.Name()
	config["execution_environment"] = "quay.io/ansible/creator-ee:latest"

	err = p.Prepare(config)
	require.NoError(t, err)

	require.NotNil(t, p.config.AnsibleCfg)
	require.Equal(t, "/tmp/.ansible/tmp", p.config.AnsibleCfg["defaults"]["remote_tmp"])
	require.Equal(t, "/tmp/.ansible-local", p.config.AnsibleCfg["defaults"]["local_tmp"])
}

func TestProvisionerPrepare_DoesNotOverrideExplicitAnsibleCfgWhenExecutionEnvironmentSet(t *testing.T) {
	var p Provisioner
	config := testConfig(t)
	defer os.Remove(config["command"].(string))

	playbookFile, err := os.CreateTemp("", "playbook")
	require.NoError(t, err)
	defer os.Remove(playbookFile.Name())

	config["playbook_file"] = playbookFile.Name()
	config["execution_environment"] = "quay.io/ansible/creator-ee:latest"
	config["ansible_cfg"] = map[string]interface{}{
		"defaults": map[string]interface{}{
			"remote_tmp": "/custom/remote",
			"local_tmp":  "/custom/local",
		},
	}

	err = p.Prepare(config)
	require.NoError(t, err)

	require.Equal(t, "/custom/remote", p.config.AnsibleCfg["defaults"]["remote_tmp"])
	require.Equal(t, "/custom/local", p.config.AnsibleCfg["defaults"]["local_tmp"])
}

func TestExecutionEnvironmentDefaults_GenerateAndCreateTempFile(t *testing.T) {
	var p Provisioner
	config := testConfig(t)
	defer os.Remove(config["command"].(string))

	playbookFile, err := os.CreateTemp("", "playbook")
	require.NoError(t, err)
	defer os.Remove(playbookFile.Name())

	config["playbook_file"] = playbookFile.Name()
	config["execution_environment"] = "quay.io/ansible/creator-ee:latest"

	err = p.Prepare(config)
	require.NoError(t, err)

	content, err := generateAnsibleCfg(p.config.AnsibleCfg)
	require.NoError(t, err)

	ansibleCfgPath, err := createTempAnsibleCfg(content)
	require.NoError(t, err)
	defer os.Remove(ansibleCfgPath)

	data, err := os.ReadFile(ansibleCfgPath)
	require.NoError(t, err)
	require.Contains(t, string(data), "remote_tmp = /tmp/.ansible/tmp")
	require.Contains(t, string(data), "local_tmp = /tmp/.ansible-local")
}

func TestExecuteSinglePlaybook_SetsANSIBLE_CONFIGWhenProvided(t *testing.T) {
	// Build a stub "ansible-navigator" that records ANSIBLE_CONFIG into OUTPUT_FILE.
	stubDir := t.TempDir()
	stubPath := filepath.Join(stubDir, "ansible-navigator-stub.sh")

	outputFile := filepath.Join(stubDir, "ansible_cfg_env.txt")
	os.Setenv("OUTPUT_FILE", outputFile)
	t.Cleanup(func() { _ = os.Unsetenv("OUTPUT_FILE") })

	stub := `#!/usr/bin/env bash
set -euo pipefail

for arg in "$@"; do
  if [[ "$arg" == "--version" ]]; then
    echo "ansible 1.6.0"
    exit 0
  fi
done

echo -n "${ANSIBLE_CONFIG:-}" > "${OUTPUT_FILE}"
exit 0
`
	require.NoError(t, os.WriteFile(stubPath, []byte(stub), 0o755))

	playbookFile, err := os.CreateTemp("", "playbook")
	require.NoError(t, err)
	defer os.Remove(playbookFile.Name())

	inventoryFile, err := os.CreateTemp("", "inventory")
	require.NoError(t, err)
	defer os.Remove(inventoryFile.Name())

	var p Provisioner
	p.config.Command = stubPath
	p.config.NavigatorMode = "stdout"
	p.config.PlaybookFile = playbookFile.Name()
	p.config.InventoryFile = inventoryFile.Name()
	p.config.PackerBuilderType = "fakebuilder"
	p.config.UseProxy = confighelper.TriTrue
	p.generatedData = basicGenData(nil)

	ui := &packersdk.BasicUi{Reader: new(bytes.Buffer), Writer: new(bytes.Buffer)}

	const cfgPath = "/tmp/packer-ansible-cfg-test.ini"
	require.NoError(t, p.executeSinglePlaybook(ui, "", commonsteps.HttpAddrNotImplemented, cfgPath))

	got, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	require.Equal(t, cfgPath, string(got))
}
