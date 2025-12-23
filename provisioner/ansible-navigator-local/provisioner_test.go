// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ansiblenavigatorlocal

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/stretchr/testify/require"
)

func TestProvisioner_Impl(t *testing.T) {
	var raw interface{} = &Provisioner{}
	if _, ok := raw.(packersdk.Provisioner); !ok {
		t.Fatalf("must be a Provisioner")
	}
}

func TestProvisionerPrepare_Defaults(t *testing.T) {
	var p Provisioner
	config := testConfig()

	playbookFile, err := os.CreateTemp("", "playbook-*.yml")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbookFile.Name())

	config["play"] = []map[string]interface{}{
		{
			"target": playbookFile.Name(),
		},
	}

	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !strings.HasPrefix(filepath.ToSlash(p.stagingDir), DefaultStagingDir) {
		t.Fatalf("unexpected staging dir %s, expected prefix %s", p.stagingDir, DefaultStagingDir)
	}
}

func TestProvisionerPrepare_DecodesNavigatorConfigExecutionEnvironmentNewFields(t *testing.T) {
	var p Provisioner
	config := testConfig()

	playbookFile, err := os.CreateTemp("", "playbook-*.yml")
	require.NoError(t, err)
	defer os.Remove(playbookFile.Name())

	config["play"] = []map[string]interface{}{{"target": playbookFile.Name()}}
	config["navigator_config"] = map[string]interface{}{
		"execution_environment": map[string]interface{}{
			"enabled":           true,
			"container_engine":  "podman",
			"container_options": []string{"--net=host"},
			"pull_arguments":    []string{"--tls-verify=false"},
		},
	}

	require.NoError(t, p.Prepare(config))
	require.NotNil(t, p.config.NavigatorConfig)
	require.NotNil(t, p.config.NavigatorConfig.ExecutionEnvironment)
	require.Equal(t, "podman", p.config.NavigatorConfig.ExecutionEnvironment.ContainerEngine)
	require.Equal(t, []string{"--net=host"}, p.config.NavigatorConfig.ExecutionEnvironment.ContainerOptions)
	require.Equal(t, []string{"--tls-verify=false"}, p.config.NavigatorConfig.ExecutionEnvironment.PullArguments)
}

func TestProvisionerPrepare_RequiresPlayBlock(t *testing.T) {
	var p Provisioner
	config := testConfig()

	err := p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}
	if !strings.Contains(err.Error(), "at least one `play` block must be defined") {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestProvisionerPrepare_PlayTargetRequired(t *testing.T) {
	var p Provisioner
	config := testConfig()

	config["play"] = []map[string]interface{}{{}}
	if err := p.Prepare(config); err == nil {
		t.Fatal("should have error")
	}
}

func TestProvisionerProvision_UploadsInventoryAndExecutesPlaybook(t *testing.T) {
	var p Provisioner
	config := testConfig()

	playbookFile, err := os.CreateTemp("", "playbook-*.yml")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbookFile.Name())

	config["play"] = []map[string]interface{}{
		{
			"target": playbookFile.Name(),
		},
	}

	if err := p.Prepare(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	comm := &communicatorMock{}
	ui := packersdk.TestUi(t)

	if err := p.Provision(context.Background(), ui, comm, map[string]interface{}{"PackerHTTPAddr": "127.0.0.1"}); err != nil {
		t.Fatalf("err: %s", err)
	}

	invRemote := filepath.ToSlash(filepath.Join(p.stagingDir, "inventory.ini"))
	playRemote := filepath.ToSlash(filepath.Join(p.stagingDir, filepath.Base(playbookFile.Name())))

	foundInvUpload := false
	foundPlayUpload := false
	for _, dest := range comm.uploadDestination {
		if filepath.ToSlash(dest) == invRemote {
			foundInvUpload = true
		}
		if filepath.ToSlash(dest) == playRemote {
			foundPlayUpload = true
		}
	}
	if !foundInvUpload {
		t.Fatalf("expected inventory upload to %s; got: %#v", invRemote, comm.uploadDestination)
	}
	if !foundPlayUpload {
		t.Fatalf("expected playbook upload to %s; got: %#v", playRemote, comm.uploadDestination)
	}

	foundRun := false
	for _, cmd := range comm.startCommand {
		if strings.Contains(cmd, "ansible-navigator") && strings.Contains(cmd, " run") && strings.Contains(cmd, playRemote) && strings.Contains(cmd, "-i="+invRemote) {
			foundRun = true
			break
		}
	}
	if !foundRun {
		t.Fatalf("expected ansible-navigator run command containing play=%s and inventory=%s; got: %#v", playRemote, invRemote, comm.startCommand)
	}
}

func testConfig() map[string]interface{} {
	m := make(map[string]interface{})
	return m
}

func TestProvisioner_WarnsOnSkipVersionCheckWithExplicitTimeout(t *testing.T) {
	var p Provisioner
	config := testConfig()

	playbookFile, err := os.CreateTemp("", "playbook-*.yml")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbookFile.Name())

	config["play"] = []map[string]interface{}{{"target": playbookFile.Name()}}
	config["skip_version_check"] = true
	config["version_check_timeout"] = "10s"

	if err := p.Prepare(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	comm := &communicatorMock{}
	out := new(bytes.Buffer)
	ui := &packersdk.BasicUi{Reader: new(bytes.Buffer), Writer: out}

	if err := p.Provision(context.Background(), ui, comm, map[string]interface{}{"PackerHTTPAddr": "127.0.0.1"}); err != nil {
		t.Fatalf("err: %s", err)
	}

	if !strings.Contains(out.String(), "Warning: version_check_timeout is ignored when skip_version_check=true") {
		t.Fatalf("expected warning in UI output; got: %q", out.String())
	}
}

func TestProvisioner_DoesNotWarnOnSkipVersionCheckWithoutExplicitTimeout(t *testing.T) {
	var p Provisioner
	config := testConfig()

	playbookFile, err := os.CreateTemp("", "playbook-*.yml")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbookFile.Name())

	config["play"] = []map[string]interface{}{{"target": playbookFile.Name()}}
	config["skip_version_check"] = true
	// Do not set version_check_timeout; Prepare defaults it.

	if err := p.Prepare(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	comm := &communicatorMock{}
	out := new(bytes.Buffer)
	ui := &packersdk.BasicUi{Reader: new(bytes.Buffer), Writer: out}

	if err := p.Provision(context.Background(), ui, comm, map[string]interface{}{"PackerHTTPAddr": "127.0.0.1"}); err != nil {
		t.Fatalf("err: %s", err)
	}

	if strings.Contains(out.String(), "version_check_timeout is ignored") {
		t.Fatalf("did not expect warning in UI output; got: %q", out.String())
	}
}

func TestProvisioner_DoesNotWarnWhenSkipVersionCheckIsFalse(t *testing.T) {
	var p Provisioner
	config := testConfig()

	playbookFile, err := os.CreateTemp("", "playbook-*.yml")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbookFile.Name())

	config["play"] = []map[string]interface{}{{"target": playbookFile.Name()}}
	config["skip_version_check"] = false
	config["version_check_timeout"] = "10s"

	if err := p.Prepare(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	comm := &communicatorMock{}
	out := new(bytes.Buffer)
	ui := &packersdk.BasicUi{Reader: new(bytes.Buffer), Writer: out}

	if err := p.Provision(context.Background(), ui, comm, map[string]interface{}{"PackerHTTPAddr": "127.0.0.1"}); err != nil {
		t.Fatalf("err: %s", err)
	}

	if strings.Contains(out.String(), "version_check_timeout is ignored") {
		t.Fatalf("did not expect warning in UI output; got: %q", out.String())
	}
}

func TestProvisionerProvision_PlayExtraArgs_AppliedBeforeGeneratedArgsAndTarget(t *testing.T) {
	var p Provisioner
	config := testConfig()

	playbookFile, err := os.CreateTemp("", "playbook-*.yml")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(playbookFile.Name())

	config["play"] = []map[string]interface{}{
		{
			"target":     playbookFile.Name(),
			"extra_args": []string{"--check", "--diff"},
		},
	}
	config["navigator_config"] = map[string]interface{}{"mode": "stdout"}

	if err := p.Prepare(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	comm := &communicatorMock{}
	ui := packersdk.TestUi(t)

	if err := p.Provision(context.Background(), ui, comm, map[string]interface{}{"PackerHTTPAddr": "127.0.0.1"}); err != nil {
		t.Fatalf("err: %s", err)
	}

	invRemote := filepath.ToSlash(filepath.Join(p.stagingDir, "inventory.ini"))
	playRemote := filepath.ToSlash(filepath.Join(p.stagingDir, filepath.Base(playbookFile.Name())))

	var cmd string
	for _, c := range comm.startCommand {
		if strings.Contains(c, "ansible-navigator") && strings.Contains(c, " run") {
			cmd = c
			break
		}
	}
	require.NotEmpty(t, cmd)

	idxRun := strings.Index(cmd, " run --mode=stdout")
	idxCheck := strings.Index(cmd, " --check")
	idxDiff := strings.Index(cmd, " --diff")
	idxInv := strings.Index(cmd, "-i="+invRemote)
	idxPlay := strings.Index(cmd, playRemote)

	require.NotEqual(t, -1, idxRun)
	require.NotEqual(t, -1, idxCheck)
	require.NotEqual(t, -1, idxDiff)
	require.NotEqual(t, -1, idxInv)
	require.NotEqual(t, -1, idxPlay)

	// Ordering expectations: extra_args before inventory, and inventory before play target.
	require.Less(t, idxCheck, idxInv)
	require.Less(t, idxDiff, idxInv)
	require.Less(t, idxInv, idxPlay)
}

func TestProvisionerProvision_ProvisionerExtraVars_JSONSinglePairAndTargetLast(t *testing.T) {
	var p Provisioner
	config := testConfig()

	playbookFile, err := os.CreateTemp("", "playbook-*.yml")
	require.NoError(t, err)
	defer os.Remove(playbookFile.Name())

	config["play"] = []map[string]interface{}{{"target": playbookFile.Name()}}
	if err := p.Prepare(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	// These are normally populated by Packer; set them in the unit test so the
	// generated JSON extra-vars are meaningful.
	p.config.PackerBuilderType = "docker"
	p.config.PackerBuildName = "example-build"

	comm := &communicatorMock{}
	ui := packersdk.TestUi(t)

	if err := p.Provision(context.Background(), ui, comm, map[string]interface{}{"PackerHTTPAddr": "127.0.0.1:8080"}); err != nil {
		t.Fatalf("err: %s", err)
	}

	invRemote := filepath.ToSlash(filepath.Join(p.stagingDir, "inventory.ini"))
	playRemote := filepath.ToSlash(filepath.Join(p.stagingDir, filepath.Base(playbookFile.Name())))

	var cmd string
	for _, c := range comm.startCommand {
		if strings.Contains(c, "ansible-navigator") && strings.Contains(c, " run") {
			cmd = c
			break
		}
	}
	require.NotEmpty(t, cmd)

	// Verify JSON-based extra-vars are used via file reference (--extra-vars=@file).
	require.Contains(t, cmd, "--extra-vars=@")

	// Legacy malformed extra-vars string should not appear.
	require.NotContains(t, cmd, "-o IdentitiesOnly=yes")
	require.NotContains(t, cmd, "--extra-vars packer_build_name=")

	// Target should be last (and inventory should still be present).
	require.Contains(t, cmd, "-i="+invRemote)
	require.True(t, strings.HasSuffix(cmd, playRemote), "expected command to end with play target %q; got: %q", playRemote, cmd)
}
