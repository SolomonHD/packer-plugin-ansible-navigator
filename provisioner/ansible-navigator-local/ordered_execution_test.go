package ansiblenavigatorlocal

import (
	"os"
	"path/filepath"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/stretchr/testify/require"
)

func TestProvisioner_ExecutePlays_ExecutesInOrder(t *testing.T) {
	var p Provisioner
	config := testConfig()

	play1, err := os.CreateTemp("", "01-first-*.yml")
	require.NoError(t, err)
	defer os.Remove(play1.Name())
	_, _ = play1.WriteString("- hosts: all\n  gather_facts: false\n  tasks: []\n")
	_ = play1.Close()

	play2, err := os.CreateTemp("", "02-second-*.yml")
	require.NoError(t, err)
	defer os.Remove(play2.Name())
	_, _ = play2.WriteString("- hosts: all\n  gather_facts: false\n  tasks: []\n")
	_ = play2.Close()

	config["play"] = []map[string]interface{}{{"target": play1.Name()}, {"target": play2.Name()}}
	require.NoError(t, p.Prepare(config))

	comm := &communicatorMock{}
	ui := packersdk.TestUi(t)
	// Required by buildExtraArgs() (uses %s formatting)
	p.generatedData = map[string]interface{}{"PackerHTTPAddr": "127.0.0.1"}

	inventoryRemote := filepath.ToSlash(filepath.Join(p.stagingDir, "inventory.ini"))
	require.NoError(t, p.executePlays(ui, comm, inventoryRemote, ""))

	remote1 := filepath.ToSlash(filepath.Join(p.stagingDir, filepath.Base(play1.Name())))
	remote2 := filepath.ToSlash(filepath.Join(p.stagingDir, filepath.Base(play2.Name())))

	require.Len(t, comm.startCommand, 2)
	require.Contains(t, comm.startCommand[0], remote1)
	require.Contains(t, comm.startCommand[1], remote2)
}
