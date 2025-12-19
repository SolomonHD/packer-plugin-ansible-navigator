package ansiblenavigatorlocal

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/stretchr/testify/require"
)

func findRunCommand(cmds []string) string {
	for _, c := range cmds {
		if strings.Contains(c, "ansible-navigator") && strings.Contains(c, " run") {
			return c
		}
	}
	return ""
}

func TestEEDockerPreflight_IncludedOnlyWhenDebugAndEEEnabled(t *testing.T) {
	playbookFile, err := os.CreateTemp("", "playbook-*.yml")
	require.NoError(t, err)
	defer os.Remove(playbookFile.Name())

	t.Run("disabled when debug off", func(t *testing.T) {
		var p Provisioner
		cfg := testConfig()
		cfg["play"] = []map[string]interface{}{{"target": playbookFile.Name()}}
		cfg["navigator_config"] = map[string]interface{}{
			"logging":               map[string]interface{}{"level": "info"},
			"execution_environment": map[string]interface{}{"enabled": true},
		}

		require.NoError(t, p.Prepare(cfg))
		comm := &communicatorMock{}
		ui := packersdk.TestUi(t)
		require.NoError(t, p.Provision(context.Background(), ui, comm, map[string]interface{}{"PackerHTTPAddr": "127.0.0.1"}))

		runCmd := findRunCommand(comm.startCommand)
		require.NotEmpty(t, runCmd)
		require.NotContains(t, runCmd, "[EE preflight]")
	})

	t.Run("disabled when EE off", func(t *testing.T) {
		var p Provisioner
		cfg := testConfig()
		cfg["play"] = []map[string]interface{}{{"target": playbookFile.Name()}}
		cfg["navigator_config"] = map[string]interface{}{
			"logging":               map[string]interface{}{"level": "debug"},
			"execution_environment": map[string]interface{}{"enabled": false},
		}

		require.NoError(t, p.Prepare(cfg))
		comm := &communicatorMock{}
		ui := packersdk.TestUi(t)
		require.NoError(t, p.Provision(context.Background(), ui, comm, map[string]interface{}{"PackerHTTPAddr": "127.0.0.1"}))

		runCmd := findRunCommand(comm.startCommand)
		require.NotEmpty(t, runCmd)
		require.NotContains(t, runCmd, "[EE preflight]")
	})

	t.Run("enabled when debug on and EE on", func(t *testing.T) {
		var p Provisioner
		cfg := testConfig()
		cfg["play"] = []map[string]interface{}{{"target": playbookFile.Name()}}
		cfg["navigator_config"] = map[string]interface{}{
			"logging":               map[string]interface{}{"level": "debug"},
			"execution_environment": map[string]interface{}{"enabled": true},
		}

		require.NoError(t, p.Prepare(cfg))
		comm := &communicatorMock{}
		ui := packersdk.TestUi(t)
		require.NoError(t, p.Provision(context.Background(), ui, comm, map[string]interface{}{"PackerHTTPAddr": "127.0.0.1"}))

		runCmd := findRunCommand(comm.startCommand)
		require.NotEmpty(t, runCmd)
		require.Contains(t, runCmd, "[EE preflight] DOCKER_HOST")
		require.Contains(t, runCmd, "[EE preflight] docker client")
		require.Contains(t, runCmd, "[DEBUG][WARN] [EE preflight] detected dockerd")
	})
}

func TestEEDockerPreflightShell_DockerMissing(t *testing.T) {
	tmpDir := t.TempDir()

	// Provide ps/grep so the script's dockerd heuristic doesn't fail due to missing commands.
	psPath, err := exec.LookPath("ps")
	require.NoError(t, err)
	grepPath, err := exec.LookPath("grep")
	require.NoError(t, err)

	require.NoError(t, os.Symlink(psPath, filepath.Join(tmpDir, "ps")))
	require.NoError(t, os.Symlink(grepPath, filepath.Join(tmpDir, "grep")))

	snippet := buildEEDockerPreflightShell("")
	cmd := exec.Command("bash", "-lc", snippet)
	cmd.Env = append(os.Environ(), "PATH="+tmpDir, "DOCKER_HOST=")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err)

	require.Contains(t, string(out), "[DEBUG] [EE preflight] docker client: missing in PATH")
}

func TestEEDockerPreflightShell_DockerdDetected_EmitsWarningOnly(t *testing.T) {
	tmpDir := t.TempDir()

	// Fake ps that always reports dockerd.
	psStub := filepath.Join(tmpDir, "ps")
	// Use an absolute shebang so it still runs when PATH is constrained.
	require.NoError(t, os.WriteFile(psStub, []byte("#!/bin/sh\necho dockerd\n"), 0o755))

	grepPath, err := exec.LookPath("grep")
	require.NoError(t, err)
	require.NoError(t, os.Symlink(grepPath, filepath.Join(tmpDir, "grep")))

	snippet := buildEEDockerPreflightShell("")
	cmd := exec.Command("bash", "-lc", snippet)
	cmd.Env = append(os.Environ(), "PATH="+tmpDir)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err)

	require.Contains(t, string(out), "[DEBUG][WARN] [EE preflight] detected dockerd process")
}
