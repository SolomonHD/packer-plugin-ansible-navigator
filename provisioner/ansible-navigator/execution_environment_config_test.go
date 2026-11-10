// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansiblenavigator

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/hcl/v2/hclparse"
	sdkconfig "github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Follow the HCL-to-cty-to-config path used by Packer, so the regression also
// detects generated-schema drift rather than only exercising mapstructure.
func decodeExecutionEnvironmentHCL(t *testing.T, source string) Config {
	t.Helper()
	file, diags := hclparse.NewParser().ParseHCL([]byte(source), "ee.pkr.hcl")
	require.False(t, diags.HasErrors(), diags.Error())
	value, diags := hcldec.Decode(file.Body, new(Provisioner).ConfigSpec(), nil)
	require.False(t, diags.HasErrors(), diags.Error())
	var cfg Config
	require.NoError(t, sdkconfig.Decode(&cfg, nil, value))
	return cfg
}

const executionEnvironmentHCL = `
navigator_config {
  mode = "stdout"
  execution_environment {
    enabled = true
    image = "example.invalid/ee:fixture"
    container_engine = %q
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
}
`

func TestExecutionEnvironmentHCLAdvancedFields(t *testing.T) {
	for _, engine := range []string{"docker", "podman", "auto"} {
		t.Run(engine, func(t *testing.T) {
			cfg := decodeExecutionEnvironmentHCL(t, fmt.Sprintf(executionEnvironmentHCL, engine))
			require.NotNil(t, cfg.NavigatorConfig)
			ee := cfg.NavigatorConfig.ExecutionEnvironment
			require.NotNil(t, ee)
			require.True(t, ee.Enabled)
			require.Equal(t, "example.invalid/ee:fixture", ee.Image)
			require.Equal(t, engine, ee.ContainerEngine)
			require.Equal(t, []string{"--security-opt", "label=disable", "--network=host"}, ee.ContainerOptions)
			require.Equal(t, "always", ee.PullPolicy)
			require.Equal(t, []string{"--quiet", "--platform=linux/amd64"}, ee.PullArguments)
			require.Equal(t, []VolumeMount{
				{Src: "/runtime/data", Dest: "/container/data", Options: "ro"},
				{Src: "/runtime/config", Dest: "/container/config", Options: "ro"},
			}, ee.VolumeMounts)
		})
	}
}

func TestExecutionEnvironmentHCLRejectsVolumeMountAssignment(t *testing.T) {
	file, diags := hclparse.NewParser().ParseHCL([]byte(`
navigator_config {
  execution_environment {
    volume_mounts = [{src = "/runtime/data", dest = "/container/data", options = "ro"}]
  }
}
`), "invalid-ee.pkr.hcl")
	require.False(t, diags.HasErrors(), diags.Error())
	_, diags = hcldec.Decode(file.Body, new(Provisioner).ConfigSpec(), nil)
	require.True(t, diags.HasErrors(), "mount list assignment must not silently bypass the block schema")
	require.Contains(t, diags.Error(), "Unsupported argument")
	require.Contains(t, diags.Error(), "volume_mounts")
}

func generatedExecutionEnvironment(t *testing.T, content string) map[string]interface{} {
	t.Helper()
	var document map[string]interface{}
	require.NoError(t, yaml.Unmarshal([]byte(content), &document))
	root, ok := document["ansible-navigator"].(map[string]interface{})
	require.True(t, ok, "missing ansible-navigator root: %s", content)
	ee, ok := root["execution-environment"].(map[string]interface{})
	require.True(t, ok, "missing execution-environment: %s", content)
	return ee
}

func TestExecutionEnvironmentYAMLPullCombinations(t *testing.T) {
	for _, tc := range []struct {
		name      string
		policy    string
		arguments []string
		want      map[string]interface{}
	}{
		{name: "neither"},
		{name: "empty arguments", arguments: []string{}},
		{name: "policy only", policy: "missing", want: map[string]interface{}{"policy": "missing"}},
		{name: "arguments only", arguments: []string{"--quiet", "--platform=linux/amd64"}, want: map[string]interface{}{"arguments": []interface{}{"--quiet", "--platform=linux/amd64"}}},
		{name: "policy and arguments", policy: "always", arguments: []string{"--quiet"}, want: map[string]interface{}{"policy": "always", "arguments": []interface{}{"--quiet"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &NavigatorConfig{ExecutionEnvironment: &ExecutionEnvironment{Enabled: true, PullPolicy: tc.policy, PullArguments: tc.arguments}}
			content, err := GenerateNavigatorConfigYAML(cfg, "", nil)
			require.NoError(t, err)
			ee := generatedExecutionEnvironment(t, content)
			if tc.want == nil {
				require.NotContains(t, ee, "pull")
			} else {
				require.Equal(t, tc.want, ee["pull"])
			}
			require.NotContains(t, ee, "pull-policy")
			require.NotContains(t, ee, "pull-arguments")
			require.NotContains(t, ee, "pull_arguments")
		})
	}
}

func TestExecutionEnvironmentHCLToYAMLWithAutomaticDefaults(t *testing.T) {
	cfg := decodeExecutionEnvironmentHCL(t, fmt.Sprintf(executionEnvironmentHCL, "docker"))
	// Bridge connectivity and stdout TTY defaults must preserve ordered user
	// options. The fixture passes topology explicitly without contacting Docker.
	cfg.NavigatorConfig.ExecutionEnvironment.ContainerOptions = []string{"--security-opt", "label=disable"}
	topology := &proxyTopology{ContainerEngine: "docker", NetworkMode: proxyNetworkBridge, AutomaticHostMapping: dockerGatewayHostMapping}
	collections := filepath.Join(t.TempDir(), "collections")
	content, err := GenerateNavigatorConfigYAML(cfg.NavigatorConfig, collections, topology)
	require.NoError(t, err)
	ee := generatedExecutionEnvironment(t, content)
	require.Equal(t, "docker", ee["container-engine"])
	require.Equal(t, "example.invalid/ee:fixture", ee["image"])
	require.Equal(t, []interface{}{"--security-opt", "label=disable", "--tty=false", dockerGatewayHostMapping}, ee["container-options"])
	require.Equal(t, map[string]interface{}{"policy": "always", "arguments": []interface{}{"--quiet", "--platform=linux/amd64"}}, ee["pull"])
	require.Equal(t, []interface{}{
		map[string]interface{}{"src": "/runtime/data", "dest": "/container/data", "options": "ro"},
		map[string]interface{}{"src": "/runtime/config", "dest": "/container/config", "options": "ro"},
		map[string]interface{}{"src": collections, "dest": "/tmp/.packer_ansible/collections", "options": "ro"},
	}, ee["volume-mounts"])
	second, err := GenerateNavigatorConfigYAML(cfg.NavigatorConfig, collections, topology)
	require.NoError(t, err)
	require.Equal(t, content, second, "repeated generation must not duplicate automatic options or mounts")
}

func TestExecutionEnvironmentYAMLExplicitCollectionsMount(t *testing.T) {
	collections := filepath.Join(t.TempDir(), "collections")
	for _, tc := range []struct {
		name  string
		mount VolumeMount
	}{
		{name: "same destination", mount: VolumeMount{Src: "/custom/collections", Dest: "/tmp/.packer_ansible/collections", Options: "rw"}},
		{name: "same source", mount: VolumeMount{Src: collections, Dest: "/custom/collections", Options: "ro"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &NavigatorConfig{ExecutionEnvironment: &ExecutionEnvironment{Enabled: true, VolumeMounts: []VolumeMount{tc.mount}}}
			content, err := GenerateNavigatorConfigYAML(cfg, collections, nil)
			require.NoError(t, err)
			require.Equal(t, []VolumeMount{tc.mount}, cfg.ExecutionEnvironment.VolumeMounts)
			second, err := GenerateNavigatorConfigYAML(cfg, collections, nil)
			require.NoError(t, err)
			require.Equal(t, content, second)
		})
	}
}
