// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansiblenavigatorlocal

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/hcl/v2/hclparse"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	sdkconfig "github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestExecutionEnvironmentHCL(t *testing.T) {
	const body = `
navigator_config {
  mode = "stdout"
  execution_environment {
    enabled = true
    image = "example.test/ee:qualified"
    container_engine = "podman"
    container_options = ["--security-opt", "label=disable", "--network=host"]
    pull_policy = "always"
    pull_arguments = ["--quiet", "--platform=linux/amd64"]
    volume_mounts {
      src = "/runtime/data"
      dest = "/container/data"
      options = "ro"
    }
    volume_mounts {
      src = "/runtime/config"
      dest = "/container/config"
      options = "ro"
    }
  }
}`
	file, diags := hclparse.NewParser().ParseHCL([]byte(body), "ee.pkr.hcl")
	require.False(t, diags.HasErrors(), diags.Error())
	value, diags := hcldec.Decode(file.Body, (&Provisioner{}).ConfigSpec(), nil)
	require.False(t, diags.HasErrors(), diags.Error())
	var config Config
	require.NoError(t, sdkconfig.Decode(&config, nil, value))
	ee := config.NavigatorConfig.ExecutionEnvironment
	require.True(t, ee.Enabled)
	require.Equal(t, "example.test/ee:qualified", ee.Image)
	require.Equal(t, "podman", ee.ContainerEngine)
	require.Equal(t, []string{"--security-opt", "label=disable", "--network=host"}, ee.ContainerOptions)
	require.Equal(t, "always", ee.PullPolicy)
	require.Equal(t, []string{"--quiet", "--platform=linux/amd64"}, ee.PullArguments)
	require.Equal(t, []VolumeMount{
		{Src: "/runtime/data", Dest: "/container/data", Options: "ro"},
		{Src: "/runtime/config", Dest: "/container/config", Options: "ro"},
	}, ee.VolumeMounts)

	// Verify the same decoded values reach the full YAML representation and
	// existing CLI route, rather than only checking generated schema entries.
	content, err := generateNavigatorConfigYAML(config.NavigatorConfig, "")
	require.NoError(t, err)
	root := decodeEEYAML(t, content)
	generated := root["execution-environment"].(map[string]interface{})
	require.Equal(t, ee.ContainerEngine, generated["container-engine"])
	require.Equal(t, []interface{}{"--security-opt", "label=disable", "--network=host"}, generated["container-options"])
	require.Equal(t, []interface{}{
		map[string]interface{}{"src": "/runtime/data", "dest": "/container/data", "options": "ro"},
		map[string]interface{}{"src": "/runtime/config", "dest": "/container/config", "options": "ro"},
	}, generated["volume-mounts"])
	flags := buildNavigatorCLIFlags(config.NavigatorConfig)
	require.Contains(t, flags, "--execution-environment-container-engine=podman")
	require.Contains(t, flags, `--container-options=["--security-opt","label=disable","--network=host"]`)
	require.Contains(t, flags, "--evm=/runtime/data:/container/data:ro")
	require.Contains(t, flags, "--evm=/runtime/config:/container/config:ro")
}

func TestExecutionEnvironmentHCLRejectsMountAssignment(t *testing.T) {
	const body = `navigator_config {
  execution_environment {
    volume_mounts = [{src = "/runtime/data", dest = "/container/data", options = "ro"}]
  }
}`
	file, diags := hclparse.NewParser().ParseHCL([]byte(body), "ee.pkr.hcl")
	require.False(t, diags.HasErrors(), diags.Error())
	_, diags = hcldec.Decode(file.Body, (&Provisioner{}).ConfigSpec(), nil)
	require.True(t, diags.HasErrors())
	require.Contains(t, diags.Error(), "volume_mounts")
}

func decodeEEYAML(t *testing.T, content string) map[string]interface{} {
	t.Helper()
	var document map[string]interface{}
	require.NoError(t, yaml.Unmarshal([]byte(content), &document))
	root, ok := document["ansible-navigator"].(map[string]interface{})
	require.True(t, ok, content)
	return root
}

func TestExecutionEnvironmentPullYAML(t *testing.T) {
	for _, tc := range []struct {
		name   string
		policy string
		args   []string
	}{
		{name: "neither"},
		{name: "policy only", policy: "missing"},
		{name: "arguments only", args: []string{"--quiet", "--platform=linux/amd64"}},
		{name: "both", policy: "always", args: []string{"--quiet", "--platform=linux/amd64"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config := &NavigatorConfig{ExecutionEnvironment: &ExecutionEnvironment{Enabled: true, PullPolicy: tc.policy, PullArguments: tc.args}}
			content, err := generateNavigatorConfigYAML(config, "")
			require.NoError(t, err)
			ee := decodeEEYAML(t, content)["execution-environment"].(map[string]interface{})
			want := map[string]interface{}{}
			if tc.policy != "" {
				want["policy"] = tc.policy
			}
			if len(tc.args) > 0 {
				want["arguments"] = []interface{}{"--quiet", "--platform=linux/amd64"}
			}
			if len(want) == 0 {
				require.NotContains(t, ee, "pull")
			} else {
				require.Equal(t, want, ee["pull"])
			}

			minimal, err := generateMinimalYAML(config)
			require.NoError(t, err)
			require.Equal(t, len(tc.args) > 0, hasUnmappedSettings(config))
			if len(tc.args) == 0 {
				require.Empty(t, minimal)
			} else {
				require.Equal(t, map[string]interface{}{
					"execution-environment": map[string]interface{}{
						"pull": map[string]interface{}{"arguments": []interface{}{"--quiet", "--platform=linux/amd64"}},
					},
				}, decodeEEYAML(t, minimal))
			}
			flags := buildNavigatorCLIFlags(config)
			if tc.policy != "" {
				require.Contains(t, flags, "--pull-policy="+tc.policy)
			} else {
				for _, flag := range flags {
					require.NotContains(t, flag, "--pull-policy=")
				}
			}
		})
	}
}

func TestExecutionEnvironmentPullMinimalYAMLCoexists(t *testing.T) {
	config := &NavigatorConfig{
		Mode:                 "stdout",
		ExecutionEnvironment: &ExecutionEnvironment{Enabled: true, ContainerEngine: "podman", PullArguments: []string{"--quiet"}},
		PlaybookArtifact:     &PlaybookArtifact{Enable: true, SaveAs: "/target/result.json"},
		CollectionDocCache:   &CollectionDocCache{Path: "/target/cache", Timeout: 60},
	}
	content, err := generateMinimalYAML(config)
	require.NoError(t, err)
	require.Equal(t, map[string]interface{}{
		"execution-environment": map[string]interface{}{"pull": map[string]interface{}{"arguments": []interface{}{"--quiet"}}},
		"playbook-artifact":     map[string]interface{}{"enable": true, "save-as": "/target/result.json"},
		"collection-doc-cache":  map[string]interface{}{"path": "/target/cache", "timeout": 60},
	}, decodeEEYAML(t, content))
}

type pullSettingsCommunicator struct {
	communicatorMock
	settings  []byte
	localPath string
	failRun   bool
}

func (c *pullSettingsCommunicator) Upload(dst string, reader io.Reader, info *os.FileInfo) error {
	if strings.HasSuffix(filepath.ToSlash(dst), "/ansible-navigator.yml") {
		file, ok := reader.(*os.File)
		if !ok {
			return os.ErrInvalid
		}
		c.localPath = file.Name()
		content, err := io.ReadAll(reader)
		if err != nil {
			return err
		}
		c.settings = content
	}
	return c.communicatorMock.Upload(dst, reader, info)
}

func (c *pullSettingsCommunicator) Start(ctx context.Context, cmd *packersdk.RemoteCmd) error {
	if c.failRun && strings.Contains(cmd.Command, " run ") {
		c.startCommand = append(c.startCommand, cmd.Command)
		cmd.SetExited(2)
		return nil
	}
	return c.communicatorMock.Start(ctx, cmd)
}

func TestProvisionPullArgumentsUploadAndInvocation(t *testing.T) {
	for _, failRun := range []bool{false, true} {
		name := "success"
		if failRun {
			name = "run failure"
		}
		t.Run(name, func(t *testing.T) {
			playbook := filepath.Join(t.TempDir(), "site.yml")
			require.NoError(t, os.WriteFile(playbook, []byte("- hosts: all\n"), 0600))
			var p Provisioner
			require.NoError(t, p.Prepare(map[string]interface{}{
				"play":               []map[string]interface{}{{"target": playbook}},
				"skip_version_check": true,
				"navigator_config": map[string]interface{}{
					"mode": "stdout",
					"execution_environment": map[string]interface{}{
						"enabled": true, "image": "example.test/ee:qualified", "container_engine": "podman",
						"container_options": []string{"--network=host"}, "pull_policy": "always",
						"pull_arguments": []string{"--quiet", "--platform=linux/amd64"},
						"volume_mounts":  []map[string]interface{}{{"src": "/target/data", "dest": "/container/data", "options": "ro"}},
					},
				},
			}))
			comm := &pullSettingsCommunicator{failRun: failRun}
			err := p.Provision(context.Background(), packersdk.TestUi(t), comm, map[string]interface{}{"PackerHTTPAddr": "127.0.0.1"})
			if failRun {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.NotEmpty(t, comm.settings)
			require.Equal(t, map[string]interface{}{
				"execution-environment": map[string]interface{}{"pull": map[string]interface{}{"arguments": []interface{}{"--quiet", "--platform=linux/amd64"}}},
			}, decodeEEYAML(t, string(comm.settings)))
			settingsPath := filepath.ToSlash(filepath.Join(p.stagingDir, "ansible-navigator.yml"))
			require.Contains(t, comm.uploadDestination, settingsPath)
			var command string
			for _, candidate := range comm.startCommand {
				if strings.Contains(candidate, " run ") {
					command = candidate
				}
			}
			require.NotEmpty(t, command)
			for _, fragment := range []string{
				"--mode=stdout", "--execution-environment-container-engine=podman",
				"--execution-environment-image=example.test/ee:qualified", "--pull-policy=always",
				`--container-options=["--network=host"]`, "--evm=/target/data:/container/data:ro",
				"--settings=" + settingsPath,
			} {
				require.Contains(t, command, fragment)
			}
			require.NotContains(t, command, "--quiet")
			require.NotEmpty(t, comm.localPath)
			_, err = os.Stat(comm.localPath)
			require.True(t, os.IsNotExist(err), "temporary local settings must be removed")
		})
	}
}
