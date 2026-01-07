// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansiblenavigator

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/stretchr/testify/require"
)

func TestProvisioner_buildRunCommandArgsForPlay_ExtraArgsAndOrdering(t *testing.T) {
	p := &Provisioner{}
	p.config.PackerBuilderType = "docker"
	p.config.NavigatorConfig = &NavigatorConfig{Mode: "stdout"}

	// createCmdArgs relies on generatedData for a few conditionals.
	p.generatedData = map[string]interface{}{
		"ConnType": "ssh",
	}

	play := Play{
		Target:     "site.yml",
		ExtraArgs:  []string{"--check", "--diff"},
		Become:     true,
		BecomeUser: "root",
		Tags:       []string{"tag-a", "tag-b"},
		SkipTags:   []string{"skip-a"},
		ExtraVars:  map[string]string{"b": "2", "a": "1"},
		VarsFiles:  []string{"vars/first.yml", "vars/second.yml"},
	}

	ui := &packersdk.BasicUi{}
	// Pass expected CLI flags (simulating what Provision() would do)
	// cliFlags := []string{"--mode=stdout"} // Removed in pure YAML refactor
	cmdArgs, _, extraVarsFilePath, err := p.buildRunCommandArgsForPlay(ui, play, "127.0.0.1:8080", "/tmp/inventory.ini", "/tmp/site.yml", "/tmp/key")
	require.NoError(t, err)

	// Clean up temp file created by test
	if extraVarsFilePath != "" {
		defer os.Remove(extraVarsFilePath)
	}

	require.GreaterOrEqual(t, len(cmdArgs), 1)
	require.Equal(t, "run", cmdArgs[0])
	require.Len(t, cmdArgs, len(cmdArgs))

	// Ensure play.extra_args come after run
	require.GreaterOrEqual(t, len(cmdArgs), 3)
	require.Equal(t, []string{"run", "--check", "--diff"}, cmdArgs[:3])

	// Ensure play target is last and inventory is immediately before it.
	require.GreaterOrEqual(t, len(cmdArgs), 2)
	require.Equal(t, "-i=/tmp/inventory.ini", cmdArgs[len(cmdArgs)-2])
	require.Equal(t, "/tmp/site.yml", cmdArgs[len(cmdArgs)-1])

	// Ensure sorted extra_vars ordering (a before b).
	aIdx := -1
	bIdx := -1
	for i, arg := range cmdArgs {
		if arg == "-e=a=1" {
			aIdx = i
		}
		if arg == "-e=b=2" {
			bIdx = i
		}
	}
	require.NotEqual(t, -1, aIdx)
	require.NotEqual(t, -1, bIdx)
	require.Less(t, aIdx, bIdx)
}

func TestProvisioner_buildRunCommandArgsForPlay_ProvisionerExtraVars_JSONSinglePair(t *testing.T) {
	p := &Provisioner{}
	p.config.PackerBuilderType = "docker"
	p.config.PackerBuildName = "example-build"
	p.config.NavigatorConfig = &NavigatorConfig{Mode: "stdout"}

	// createCmdArgs relies on generatedData for a few conditionals.
	p.generatedData = map[string]interface{}{
		"ConnType": "ssh",
	}

	ui := &packersdk.BasicUi{}
	play := Play{Target: "site.yml"}
	cmdArgs, _, extraVarsFilePath, err := p.buildRunCommandArgsForPlay(ui, play, "127.0.0.1:8080", "/tmp/inventory.ini", "/tmp/site.yml", "/tmp/key")
	require.NoError(t, err)

	// Clean up temp file created by test
	if extraVarsFilePath != "" {
		defer os.Remove(extraVarsFilePath)
	}

	// Exactly one --extra-vars=@file argument for provisioner-generated extra vars.
	extraVarsIdx := -1
	extraVarsCount := 0
	var extraVarsArg string
	for i, a := range cmdArgs {
		if strings.HasPrefix(a, "--extra-vars=@") {
			extraVarsCount++
			extraVarsIdx = i
			extraVarsArg = a
		}
	}
	require.Equal(t, 1, extraVarsCount)
	require.GreaterOrEqual(t, extraVarsIdx, 0)

	// Verify file-based approach with @ prefix
	require.True(t, strings.HasPrefix(extraVarsArg, "--extra-vars=@"), "extra-vars argument should be --extra-vars=@file")
	actualFilePath := strings.TrimPrefix(extraVarsArg, "--extra-vars=@")

	// Verify file exists and contains valid JSON
	require.FileExists(t, actualFilePath)
	fileContent, err2 := os.ReadFile(actualFilePath)
	require.NoError(t, err2)

	// Parse and verify JSON content
	var parsed map[string]interface{}
	require.NoError(t, json.Unmarshal(fileContent, &parsed))
	require.Equal(t, "docker", parsed["packer_builder_type"])
	require.Equal(t, "example-build", parsed["packer_build_name"])
	require.Equal(t, "127.0.0.1:8080", parsed["packer_http_addr"])
	require.Equal(t, "/tmp/key", parsed["ansible_ssh_private_key_file"])

	// All flags should use --flag=value or -f=value format (no standalone flags)
	for i, arg := range cmdArgs {
		// Check that flags are not using the old separate-elements format
		if arg == "-e" || arg == "--extra-vars" || arg == "-i" || arg == "--mode" || arg == "--tags" {
			require.Failf(t, "Found standalone flag", "argument %d: %q should use --flag=value format instead of separate elements", i, arg)
		}
	}

	// Playbook path remains last.
	require.Equal(t, "/tmp/site.yml", cmdArgs[len(cmdArgs)-1])
}

func indexOfSequence(haystack []string, needle []string) int {
	if len(needle) == 0 {
		return 0
	}
	if len(needle) > len(haystack) {
		return -1
	}
	for i := 0; i <= len(haystack)-len(needle); i++ {
		ok := true
		for j := range needle {
			if haystack[i+j] != needle[j] {
				ok = false
				break
			}
		}
		if ok {
			return i
		}
	}
	return -1
}
