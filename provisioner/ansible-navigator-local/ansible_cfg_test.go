package ansiblenavigatorlocal

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
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
		"",
		"",
	}, "\n")

	require.Equal(t, expected, got)
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
		Plays:      []Play{{Target: playbookFile.Name()}},
		AnsibleCfg: map[string]map[string]string{},
	}

	err = cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "ansible_cfg cannot be an empty map")
}

func TestProvisionerPrepare_AppliesAnsibleCfgDefaultsWhenExecutionEnvironmentSet(t *testing.T) {
	var p Provisioner
	config := testConfig()

	playbookFile, err := os.CreateTemp("", "playbook")
	require.NoError(t, err)
	defer os.Remove(playbookFile.Name())

	config["play"] = []map[string]interface{}{{"target": playbookFile.Name()}}
	config["execution_environment"] = "quay.io/ansible/creator-ee:latest"

	err = p.Prepare(config)
	require.NoError(t, err)

	require.NotNil(t, p.config.AnsibleCfg)
	require.Equal(t, "/tmp/.ansible/tmp", p.config.AnsibleCfg["defaults"]["remote_tmp"])
	require.Equal(t, "/tmp/.ansible-local", p.config.AnsibleCfg["defaults"]["local_tmp"])
}

func TestProvisionerPrepare_DoesNotOverrideExplicitAnsibleCfgWhenExecutionEnvironmentSet(t *testing.T) {
	var p Provisioner
	config := testConfig()

	playbookFile, err := os.CreateTemp("", "playbook")
	require.NoError(t, err)
	defer os.Remove(playbookFile.Name())

	config["play"] = []map[string]interface{}{{"target": playbookFile.Name()}}
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

func TestProvisionerProvision_UploadsAnsibleCfgAndSetsANSIBLE_CONFIG(t *testing.T) {
	var p Provisioner
	config := testConfig()

	playbookFile, err := os.CreateTemp("", "playbook")
	require.NoError(t, err)
	defer os.Remove(playbookFile.Name())

	config["play"] = []map[string]interface{}{{"target": playbookFile.Name()}}
	config["ansible_cfg"] = map[string]interface{}{
		"defaults": map[string]interface{}{
			"remote_tmp": "/tmp/.ansible/tmp",
			"local_tmp":  "/tmp/.ansible-local",
		},
	}

	err = p.Prepare(config)
	require.NoError(t, err)

	comm := &communicatorMock{}
	ui := packersdk.TestUi(t)

	require.NoError(t, p.Provision(
		context.Background(),
		ui,
		comm,
		make(map[string]interface{}),
	))

	// Upload includes <staging_dir>/ansible.cfg
	foundUpload := false
	for _, dest := range comm.uploadDestination {
		if strings.HasSuffix(filepath.ToSlash(dest), filepath.ToSlash(filepath.Join(p.stagingDir, "ansible.cfg"))) {
			foundUpload = true
			break
		}
	}
	require.True(t, foundUpload, "expected ansible.cfg to be uploaded")

	// Remote command includes ANSIBLE_CONFIG=<staging_dir>/ansible.cfg
	foundEnv := false
	needle := "ANSIBLE_CONFIG=" + filepath.ToSlash(filepath.Join(p.stagingDir, "ansible.cfg"))
	for _, cmd := range comm.startCommand {
		if strings.Contains(cmd, needle) {
			foundEnv = true
			break
		}
	}
	require.True(t, foundEnv, "expected remote command to set ANSIBLE_CONFIG")
}
