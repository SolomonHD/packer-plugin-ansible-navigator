//go:build !windows
// +build !windows

package ansiblenavigator

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/stretchr/testify/require"
)

func TestProvisioner_ExecutePlays_ExecutesInOrder(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "navigator_calls.txt")
	stubPath := filepath.Join(tmpDir, "ansible-navigator-stub.sh")
	invFile := filepath.Join(tmpDir, "inventory")
	play1 := filepath.Join(tmpDir, "01-first.yml")
	play2 := filepath.Join(tmpDir, "02-second.yaml")

	stub := `#!/usr/bin/env bash
set -euo pipefail

if [[ "${1:-}" == "--version" ]]; then
  echo "ansible-navigator 9.9.9"
  exit 0
fi

# Log the playbook (last arg) so we can assert call ordering.
echo "${@: -1}" >> "${OUTPUT_FILE}"
exit 0
`
	require.NoError(t, os.WriteFile(stubPath, []byte(stub), 0o755))
	require.NoError(t, os.WriteFile(invFile, []byte("all:\n  hosts:\n    localhost:\n"), 0o644))
	require.NoError(t, os.WriteFile(play1, []byte("- hosts: all\n  gather_facts: false\n  tasks: []\n"), 0o644))
	require.NoError(t, os.WriteFile(play2, []byte("- hosts: all\n  gather_facts: false\n  tasks: []\n"), 0o644))

	os.Setenv("OUTPUT_FILE", outputFile)
	t.Cleanup(func() { _ = os.Unsetenv("OUTPUT_FILE") })

	abs1, err := filepath.Abs(play1)
	require.NoError(t, err)
	abs2, err := filepath.Abs(play2)
	require.NoError(t, err)

	ui := &packersdk.BasicUi{Reader: new(bytes.Buffer), Writer: new(bytes.Buffer), ErrorWriter: new(bytes.Buffer)}
	p := &Provisioner{
		config: Config{
			PackerConfig: common.PackerConfig{
				PackerBuildName:   "test-build",
				PackerBuilderType: "test-builder",
			},
			Command:       stubPath,
			NavigatorMode: "stdout",
			InventoryFile: invFile,
			Plays: []Play{
				{Target: play1},
				{Target: play2},
			},
		},
		generatedData: basicGenData(map[string]interface{}{"ConnType": "docker"}),
	}

	require.NoError(t, p.executePlays(ui, nil, "", commonsteps.HttpAddrNotImplemented, "", ""))

	data, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	require.Len(t, lines, 2)
	require.Equal(t, abs1, strings.TrimSpace(lines[0]))
	require.Equal(t, abs2, strings.TrimSpace(lines[1]))
}
