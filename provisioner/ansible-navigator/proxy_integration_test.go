// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

//go:build !windows

package ansiblenavigator

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/stretchr/testify/require"
)

func TestProvisionerPrepareResolvesAndPinsProxyTopology(t *testing.T) {
	tests := []struct {
		name        string
		engine      string
		options     []string
		bind        string
		host        string
		available   map[string]bool
		wantEngine  string
		wantNetwork proxyNetworkMode
		wantBind    string
		wantHost    string
	}{
		{name: "automatic Podman bridge", available: map[string]bool{"podman": true, "docker": true}, wantEngine: "podman", wantNetwork: proxyNetworkBridge, wantBind: "0.0.0.0", wantHost: "host.containers.internal"},
		{name: "automatic Docker bridge", engine: "auto", available: map[string]bool{"docker": true}, wantEngine: "docker", wantNetwork: proxyNetworkBridge, wantBind: "0.0.0.0", wantHost: "gateway.docker.internal"},
		{name: "explicit host network", engine: "docker", options: []string{"--net", "host"}, wantEngine: "docker", wantNetwork: proxyNetworkHost, wantBind: "127.0.0.1", wantHost: "127.0.0.1"},
		{name: "explicit endpoints", engine: "podman", bind: "192.0.2.10", host: "host.example.test", wantEngine: "podman", wantNetwork: proxyNetworkBridge, wantBind: "192.0.2.10", wantHost: "host.example.test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provisioner{}
			p.containerEngineLookPath = func(name string) (string, error) {
				if tt.available[name] {
					return "/usr/bin/" + name, nil
				}
				return "", exec.ErrNotFound
			}
			cfg := Config{
				SkipVersionCheck:        true,
				AnsibleProxyBindAddress: tt.bind,
				AnsibleProxyHost:        tt.host,
				Plays:                   []Play{{Target: "example.collection.role"}},
				NavigatorConfig: &NavigatorConfig{ExecutionEnvironment: &ExecutionEnvironment{
					Enabled: true, ContainerEngine: tt.engine, ContainerOptions: tt.options,
				}},
			}
			require.NoError(t, p.Prepare(&cfg))
			require.Equal(t, tt.wantEngine, p.config.NavigatorConfig.ExecutionEnvironment.ContainerEngine)
			require.Equal(t, tt.wantNetwork, p.proxyTopology.NetworkMode)
			require.Equal(t, tt.wantBind, p.config.AnsibleProxyBindAddress)
			require.Equal(t, tt.wantHost, p.config.AnsibleProxyHost)
			require.Equal(t, tt.options, p.config.NavigatorConfig.ExecutionEnvironment.ContainerOptions)
		})
	}
}

func TestProvisionerPrepareRejectsUnreachableBridgeTopology(t *testing.T) {
	p := &Provisioner{containerEngineLookPath: func(string) (string, error) { return "/usr/bin/docker", nil }}
	cfg := Config{
		SkipVersionCheck: true,
		AnsibleProxyHost: "127.0.0.1",
		Plays:            []Play{{Target: "example.collection.role"}},
		NavigatorConfig: &NavigatorConfig{ExecutionEnvironment: &ExecutionEnvironment{
			Enabled: true, ContainerEngine: "docker",
		}},
	}
	err := p.Prepare(&cfg)
	require.ErrorContains(t, err, "container loopback cannot reach the host proxy")
}

func TestEmitProxyRouteDiagnostics(t *testing.T) {
	ui := newMockUi().(*mockUi)
	p := &Provisioner{
		config: Config{
			AnsibleProxyHost: "gateway.docker.internal",
			LocalPort:        43210,
			NavigatorConfig:  &NavigatorConfig{Logging: &LoggingConfig{Level: "debug"}},
		},
		proxyTopology: proxyTopology{
			ContainerEngine:      "docker",
			NetworkMode:          proxyNetworkBridge,
			BindAddressOrigin:    proxyValueInferred,
			AdvertisedHostOrigin: proxyValueInferred,
			AutomaticHostMapping: dockerGatewayHostMapping,
		},
	}
	p.emitProxyRouteDiagnostics(ui, "0.0.0.0:43210")
	require.Contains(t, strings.Join(ui.messageMessages, "\n"), "listener=0.0.0.0:43210 advertised=gateway.docker.internal:43210")
	require.Contains(t, strings.Join(ui.messageMessages, "\n"), "engine=docker network=bridge")
	require.Contains(t, strings.Join(ui.messageMessages, "\n"), "automatic_host_mapping="+dockerGatewayHostMapping)
}

func TestExecutePlaysReportsActualExitCode(t *testing.T) {
	tmpDir := t.TempDir()
	stubPath := filepath.Join(tmpDir, "ansible-navigator-stub.sh")
	invFile := filepath.Join(tmpDir, "inventory")
	playbook := filepath.Join(tmpDir, "site.yml")
	require.NoError(t, os.WriteFile(stubPath, []byte("#!/bin/sh\nexit 7\n"), 0o755))
	require.NoError(t, os.WriteFile(invFile, []byte("localhost\n"), 0o644))
	require.NoError(t, os.WriteFile(playbook, []byte("- hosts: all\n  tasks: []\n"), 0o644))

	p := &Provisioner{
		config: Config{
			PackerConfig: common.PackerConfig{PackerBuildName: "test", PackerBuilderType: "test"},
			Command:      stubPath, InventoryFile: invFile, Plays: []Play{{Name: "failure", Target: playbook}},
		},
		generatedData: basicGenData(map[string]interface{}{"ConnType": "docker"}),
	}
	ui := &packersdk.BasicUi{Reader: new(bytes.Buffer), Writer: new(bytes.Buffer), ErrorWriter: new(bytes.Buffer)}
	err := p.executePlays(ui, nil, "", commonsteps.HttpAddrNotImplemented, "", "")
	require.ErrorContains(t, err, "exit code 7")
	var exitErr *exec.ExitError
	require.True(t, errors.As(err, &exitErr))
}
