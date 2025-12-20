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
	require.Contains(t, comm.startCommand[0], "cd "+stagingDir+" && ansible-galaxy install -r="+remoteReq)
	require.Contains(t, comm.startCommand[1], "cd "+stagingDir+" && ansible-galaxy collection install -r="+remoteReq)
}

func TestGalaxyManager_SetupEnvironmentPaths_ReturnsExpectedEnvVars(t *testing.T) {
	ui := packersdk.TestUi(t)
	comm := &communicatorMock{}

	cfg := &Config{
		GalaxyCommand:   "ansible-galaxy",
		RolesPath:       "/tmp/roles",
		CollectionsPath: "/tmp/collections",
	}

	gm := NewGalaxyManager(cfg, ui, comm, "/tmp/stage", "/tmp/stage/galaxy_roles", "/tmp/stage/galaxy_collections")
	got := gm.SetupEnvironmentPaths()

	require.Contains(t, got, "ANSIBLE_ROLES_PATH=/tmp/roles")
	require.Contains(t, got, "ANSIBLE_COLLECTIONS_PATHS=/tmp/collections")
}

func TestGalaxyManager_InstallRequirements_UsesConfiguredCommandArgsAndForcePrecedence(t *testing.T) {
	tmpDir := t.TempDir()
	reqFile := filepath.Join(tmpDir, "requirements.yml")

	// Both sections present -> both installers should run.
	require.NoError(t, os.WriteFile(reqFile, []byte("roles:\n  - src: test.role\ncollections:\n  - name: community.general\n"), 0o644))

	comm := &communicatorMock{}
	ui := packersdk.TestUi(t)
	stagingDir := "/tmp/packer-provisioner-ansible-local-test"

	cfg := &Config{
		RequirementsFile:    reqFile,
		GalaxyCommand:       "/custom/ansible-galaxy",
		GalaxyArgs:          []string{"--ignore-certs"},
		OfflineMode:         true,
		RolesPath:           "/tmp/roles",
		CollectionsPath:     "/tmp/collections",
		GalaxyForce:         true,
		GalaxyForceWithDeps: true,
	}

	gm := NewGalaxyManager(cfg, ui, comm, stagingDir, stagingDir+"/galaxy_roles", stagingDir+"/galaxy_collections")
	require.NoError(t, gm.InstallRequirements())
	require.Len(t, comm.startCommand, 2)

	joined := comm.startCommand[0] + "\n" + comm.startCommand[1]
	require.Contains(t, joined, "cd "+stagingDir+" && /custom/ansible-galaxy")
	require.Contains(t, joined, " --offline")
	require.Contains(t, joined, " --force-with-deps")
	// Precedence: do not include --force when --force-with-deps is set.
	require.NotContains(t, joined, " --force ")
	require.Contains(t, joined, " --ignore-certs")
	require.Contains(t, joined, " -p=/tmp/roles")
	require.Contains(t, joined, " -p=/tmp/collections")
}
