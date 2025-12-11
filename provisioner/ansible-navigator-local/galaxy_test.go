package ansiblenavigatorlocal

import (
	"os"
	"path/filepath"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/stretchr/testify/require"
)

func TestGalaxyManager_InstallRequirements_RunsRolesAndCollectionsInstalls(t *testing.T) {
	tmpDir := t.TempDir()
	reqFile := filepath.Join(tmpDir, "requirements.yml")

	// Both sections present -> both installers should run.
	require.NoError(t, os.WriteFile(reqFile, []byte("roles:\n  - src: test.role\ncollections:\n  - name: community.general\n"), 0o644))

	cfg := &Config{
		RequirementsFile: reqFile,
		GalaxyCommand:    "ansible-galaxy",
	}

	comm := &communicatorMock{}
	ui := packersdk.TestUi(t)
	stagingDir := "/tmp/packer-provisioner-ansible-local-test"
	gm := NewGalaxyManager(cfg, ui, comm, stagingDir, stagingDir+"/galaxy_roles", stagingDir+"/galaxy_collections")

	require.NoError(t, gm.InstallRequirements())

	require.Len(t, comm.startCommand, 2)

	remoteReq := filepath.ToSlash(filepath.Join(stagingDir, filepath.Base(reqFile)))
	require.Contains(t, comm.startCommand[0], "ansible-galaxy install -r "+remoteReq)
	require.Contains(t, comm.startCommand[1], "ansible-galaxy collection install -r "+remoteReq)
}
