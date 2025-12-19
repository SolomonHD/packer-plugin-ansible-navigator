package ansiblenavigator

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/stretchr/testify/require"
)

func TestEEDockerPreflight_Gating(t *testing.T) {
	uiOut := new(bytes.Buffer)
	ui := &packersdk.BasicUi{Reader: new(bytes.Buffer), Writer: uiOut}

	deps := eeDockerPreflightDeps{
		stat: func(string) (os.FileInfo, error) { return nil, os.ErrNotExist },
		lookPath: func(string) (string, error) {
			return "", os.ErrNotExist
		},
		commandOutput: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return nil, os.ErrNotExist
		},
		dockerHostProvider: func() (string, bool) { return "", false },
	}

	// Disabled
	emitEEDockerPreflightWithDeps(ui, false, nil, deps)
	require.Empty(t, strings.TrimSpace(uiOut.String()))
}

func TestEEDockerPreflight_DockerMissing_AndDockerdWarning(t *testing.T) {
	uiOut := new(bytes.Buffer)
	ui := &packersdk.BasicUi{Reader: new(bytes.Buffer), Writer: uiOut}

	deps := eeDockerPreflightDeps{
		stat: func(string) (os.FileInfo, error) { return nil, os.ErrNotExist },
		lookPath: func(name string) (string, error) {
			if name == "ps" {
				return "/bin/ps", nil
			}
			return "", os.ErrNotExist
		},
		commandOutput: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			// Simulate `ps -eo comm` output containing dockerd
			if name == "ps" {
				return []byte("dockerd\n"), nil
			}
			return nil, os.ErrNotExist
		},
		dockerHostProvider: func() (string, bool) { return "", false },
	}

	emitEEDockerPreflightWithDeps(ui, true, []string{"/nonexistent"}, deps)
	out := uiOut.String()

	require.Contains(t, out, "[DEBUG] [EE preflight] DOCKER_HOST: unset")
	require.Contains(t, out, "[DEBUG] [EE preflight] /var/run/docker.sock: missing")
	require.Contains(t, out, "[DEBUG] [EE preflight] docker client: missing in PATH")
	require.Contains(t, out, "[DEBUG][WARN] [EE preflight] detected dockerd process")
}
