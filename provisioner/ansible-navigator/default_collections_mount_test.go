// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansiblenavigator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	"github.com/stretchr/testify/require"
)

// An unused plugin-default cache need not exist on a fresh host. Observe the
// actual generated YAML at launch; no Docker resources participate in this test.
func TestExecuteAnsibleDefaultCollectionsMount(t *testing.T) {
	for _, tc := range []struct {
		name                                                string
		explicitPath, existing, requirements, explicitMount bool
		wantAutomatic                                       bool
	}{
		{name: "absent unused default"},
		{name: "existing default", existing: true, wantAutomatic: true},
		{name: "absent explicit path", explicitPath: true, wantAutomatic: true},
		{name: "explicit mount retained", explicitMount: true},
		{name: "requirements create default cache after YAML generation", requirements: true, wantAutomatic: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			cache := filepath.Join(dir, "collections-cache")
			capture := filepath.Join(dir, "captured.yml")
			t.Setenv("CAPTURE_CONFIG", capture)
			t.Setenv("DOCKER_HOST", "unix:///not-used-by-fixture")
			t.Setenv("ANSIBLE_COLLECTIONS_PATH", os.Getenv("ANSIBLE_COLLECTIONS_PATH"))
			t.Setenv("ANSIBLE_ROLES_PATH", os.Getenv("ANSIBLE_ROLES_PATH"))
			navigator := navigatorFixture(t, dir, "navigator", `/bin/cat "$ANSIBLE_NAVIGATOR_CONFIG" > "$CAPTURE_CONFIG"`)
			cfg := map[string]interface{}{
				"command": navigator, "skip_version_check": true,
				"play":             []map[string]interface{}{{"target": filepath.Join(dir, "site.yml")}},
				"navigator_config": map[string]interface{}{"mode": "stdout", "execution_environment": map[string]interface{}{"enabled": true, "container_engine": "docker"}},
			}
			require.NoError(t, os.WriteFile(filepath.Join(dir, "site.yml"), []byte("- hosts: all\n"), 0600))
			if tc.explicitPath {
				cfg["collections_path"] = cache
			}
			if tc.existing {
				require.NoError(t, os.Mkdir(cache, 0700))
			}
			if tc.requirements {
				requirements := filepath.Join(dir, "requirements.yml")
				require.NoError(t, os.WriteFile(requirements, []byte("collections: []\n"), 0600))
				cfg["requirements_file"] = requirements
				cfg["galaxy_command"] = navigatorFixture(t, dir, "galaxy", "/bin/mkdir -p '"+cache+"'")
			}
			p := &Provisioner{}
			require.NoError(t, p.Prepare(cfg))
			// Re-preparing the same omitted configuration must retain its default
			// provenance rather than reclassifying a computed path as explicit input.
			require.NoError(t, p.Prepare(cfg))
			// Replace the default with an isolated fresh-host equivalent after Prepare;
			// retain the distinction between omitted and explicit user configuration.
			p.config.CollectionsPath = cache
			if tc.explicitMount {
				p.config.NavigatorConfig.ExecutionEnvironment.VolumeMounts = []VolumeMount{{Src: cache, Dest: "/manual-collections", Options: "ro"}}
			}
			p.generatedData = map[string]interface{}{"PackerHTTPAddr": commonsteps.HttpAddrNotImplemented}
			require.NoError(t, p.executeAnsible(newMockUi(), nil, ""))
			content, err := os.ReadFile(capture)
			require.NoError(t, err)
			require.Equal(t, tc.wantAutomatic, strings.Contains(string(content), "dest: /tmp/.packer_ansible/collections"), string(content))
			if tc.explicitMount {
				require.Contains(t, string(content), "dest: /manual-collections")
				require.Contains(t, string(content), cache)
			}
			if tc.requirements {
				info, err := os.Stat(cache)
				require.NoError(t, err)
				require.True(t, info.IsDir())
			}
			if !tc.existing && !tc.requirements {
				_, err = os.Stat(cache)
				require.True(t, os.IsNotExist(err), "cache must not be created merely to supply a mount")
			}
		})
	}
}
