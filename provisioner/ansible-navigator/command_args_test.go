// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansiblenavigator

import (
	"testing"

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

	cmdArgs, _ := p.buildRunCommandArgsForPlay(play, "127.0.0.1:8080", "/tmp/inventory.ini", "/tmp/site.yml", "/tmp/key")

	require.GreaterOrEqual(t, len(cmdArgs), 1)
	require.Equal(t, "run", cmdArgs[0])
	require.Len(t, cmdArgs, len(cmdArgs))

	// Ensure enforced --mode then play.extra_args
	require.GreaterOrEqual(t, len(cmdArgs), 5)
	require.Equal(t, []string{"run", "--mode", "stdout", "--check", "--diff"}, cmdArgs[:5])

	// Ensure play target is last and inventory is immediately before it.
	require.GreaterOrEqual(t, len(cmdArgs), 3)
	require.Equal(t, "-i", cmdArgs[len(cmdArgs)-3])
	require.Equal(t, "/tmp/inventory.ini", cmdArgs[len(cmdArgs)-2])
	require.Equal(t, "/tmp/site.yml", cmdArgs[len(cmdArgs)-1])

	// Ensure sorted extra_vars ordering (a before b).
	aIdx := indexOfSequence(cmdArgs, []string{"-e", "a=1"})
	bIdx := indexOfSequence(cmdArgs, []string{"-e", "b=2"})
	require.NotEqual(t, -1, aIdx)
	require.NotEqual(t, -1, bIdx)
	require.Less(t, aIdx, bIdx)
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
