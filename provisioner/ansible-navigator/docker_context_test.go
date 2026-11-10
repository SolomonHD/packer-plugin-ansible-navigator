// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansiblenavigator

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/stretchr/testify/require"
)

func dockerContextFixture(t *testing.T, script string) string {
	t.Helper()
	requirePOSIXShellFixture(t)
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docker"), []byte("#!/bin/sh\n"+script), 0700))
	return dir
}

func dockerContextUI() (*packersdk.BasicUi, *bytes.Buffer) {
	out := new(bytes.Buffer)
	return &packersdk.BasicUi{Reader: new(bytes.Buffer), Writer: out}, out
}

func TestDockerContextEffectiveEnvironment(t *testing.T) {
	// The parent installation must not override the Navigator search path.
	parentDir := dockerContextFixture(t, "printf '%s' 'unix:///wrong-parent.sock'\n")
	t.Setenv("PATH", parentDir)
	t.Setenv("DOCKER_HOST", "unix:///parent-only.sock")
	dir := dockerContextFixture(t, `
[ "$*" = 'context inspect --format {{.Endpoints.docker.Host}}' ] || exit 2
[ "$DOCKER_CONTEXT" = 'acceptance-rootless' ] || exit 3
[ "$DOCKER_CONFIG" = '/fixture/config' ] || exit 4
[ "$DOCKER_TLS_VERIFY" = '1' ] || exit 5
[ "$DOCKER_CERT_PATH" = '/fixture/certs' ] || exit 6
printf '%s' 'unix:///run/user/1000/docker.sock'
`)
	env := []string{"PATH=" + dir, "DOCKER_CONTEXT=acceptance-rootless", "DOCKER_CONFIG=/fixture/config", "DOCKER_TLS_VERIFY=1", "DOCKER_CERT_PATH=/fixture/certs"}
	original := append([]string(nil), env...)
	ui, out := dockerContextUI()
	host, err := resolveDockerHost(ui, env)
	require.NoError(t, err)
	require.Equal(t, "unix:///run/user/1000/docker.sock", host)
	require.Equal(t, original, env)
	require.Equal(t, parentDir, os.Getenv("PATH"))
	require.Equal(t, "unix:///parent-only.sock", os.Getenv("DOCKER_HOST"))
	require.Empty(t, out.String())
}

func TestDockerContextExplicitHostSkipsDiscovery(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "launched")
	dir := dockerContextFixture(t, "printf launched > \"$LAUNCH_MARKER\"\nexit 1\n")
	env := []string{"PATH=" + dir, "DOCKER_HOST=unix:///selected.sock", "DOCKER_CONTEXT=named", "LAUNCH_MARKER=" + marker}
	ui, out := dockerContextUI()
	host, err := resolveDockerHost(ui, env)
	require.NoError(t, err)
	require.Empty(t, host, "no synthesized endpoint may replace the explicit host")
	_, err = os.Stat(marker)
	require.True(t, os.IsNotExist(err))
	require.Empty(t, out.String())
}

func TestDockerContextLastEnvironmentValueWins(t *testing.T) {
	dir := dockerContextFixture(t, "printf '%s' 'unix:///discovered.sock'\n")
	ui, _ := dockerContextUI()
	host, err := resolveDockerHost(ui, []string{"PATH=" + dir, "DOCKER_HOST=unix:///stale.sock", "DOCKER_HOST="})
	require.NoError(t, err)
	require.Equal(t, "unix:///discovered.sock", host)
}

func TestDockerContextMissingAndEmptyFailOpen(t *testing.T) {
	for _, test := range []struct{ name, dir string }{
		{"missing", t.TempDir()},
		{"empty", dockerContextFixture(t, "exit 0\n")},
	} {
		t.Run(test.name, func(t *testing.T) {
			ui, out := dockerContextUI()
			host, err := resolveDockerHost(ui, []string{"PATH=" + test.dir})
			require.NoError(t, err)
			require.Empty(t, host)
			require.Empty(t, out.String())
		})
	}
}

func TestDockerContextFailureWarningIsSanitized(t *testing.T) {
	dir := dockerContextFixture(t, "printf 'tcp://fixture-user:fixture-secret@example.invalid'\nprintf 'fixture-secret' >&2\nexit 7\n")
	ui, out := dockerContextUI()
	host, err := resolveDockerHost(ui, []string{"PATH=" + dir, "DOCKER_CONTEXT=private-context-fixture"})
	require.NoError(t, err, "optional context discovery must remain fail open")
	require.Empty(t, host)
	require.Contains(t, out.String(), "Docker context discovery failed")
	require.NotContains(t, out.String(), "fixture-secret")
	require.NotContains(t, out.String(), "example.invalid")
	require.NotContains(t, out.String(), "private-context-fixture")
}

func TestDockerContextTimeoutIsBounded(t *testing.T) {
	dir := dockerContextFixture(t, "exec /bin/sleep 10\n")
	ui, out := dockerContextUI()
	start := time.Now()
	host, err := resolveDockerHostWithTimeout(ui, []string{"PATH=" + dir}, 30*time.Millisecond)
	require.NoError(t, err)
	require.Empty(t, host)
	require.Less(t, time.Since(start), time.Second)
	require.Contains(t, out.String(), "Docker context discovery timed out")
	require.Equal(t, 10*time.Second, dockerContextTimeout)
}

func TestDockerContextEndpointRemainsPrivateDuringPlay(t *testing.T) {
	dir := t.TempDir()
	endpoint := "tcp://fixture-user:fixture-secret@example.invalid:2376"
	t.Setenv("EXPECTED_DOCKER_HOST", endpoint)
	command := navigatorFixture(t, dir, "ansible-navigator", `[ "$DOCKER_HOST" = "$EXPECTED_DOCKER_HOST" ] || exit 9`)
	playbook := filepath.Join(dir, "site.yml")
	require.NoError(t, os.WriteFile(playbook, []byte("- hosts: all\n"), 0600))
	p := &Provisioner{config: Config{
		Command:         command,
		Plays:           []Play{{Target: playbook}},
		NavigatorConfig: &NavigatorConfig{Logging: &LoggingConfig{Level: "debug"}},
	}}
	ui, out := dockerContextUI()
	ui.ErrorWriter = out
	require.NoError(t, p.executePlays(ui, nil, "", "", "", endpoint))
	require.Contains(t, out.String(), "DOCKER_HOST")
	require.NotContains(t, out.String(), "fixture-secret")
	require.NotContains(t, out.String(), "example.invalid")
}

func TestDockerEngineSelectionUsesNavigatorPath(t *testing.T) {
	dir, inherited := t.TempDir(), t.TempDir()
	t.Setenv("PATH", inherited)
	// No installed engine is inherited; the configured child search path is
	// authoritative for engine selection as well as context discovery.
	navigatorFixture(t, dir, "docker", "exit 0")
	p := &Provisioner{}
	require.NoError(t, p.Prepare(&Config{
		SkipVersionCheck:     true,
		AnsibleNavigatorPath: []string{dir},
		Plays:                []Play{{Target: "example.collection.role"}},
		NavigatorConfig:      &NavigatorConfig{ExecutionEnvironment: &ExecutionEnvironment{Enabled: true, ContainerEngine: "auto"}},
	}))
	require.Equal(t, "docker", p.config.NavigatorConfig.ExecutionEnvironment.ContainerEngine)
	require.Equal(t, "gateway.docker.internal", p.config.AnsibleProxyHost)
	require.Equal(t, inherited, os.Getenv("PATH"))
}
