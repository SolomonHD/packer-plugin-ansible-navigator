// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansiblenavigator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNavigatorContainerEnginePreference(t *testing.T) {
	require.Equal(t, [2]string{"podman", "docker"}, navigatorContainerEnginePreference())
}

func TestResolveProxyTopologyMatrix(t *testing.T) {
	tests := []struct {
		name    string
		input   proxyTopologyInput
		engine  string
		network proxyNetworkMode
		bind    string
		host    string
		mapping string
	}{
		{name: "EE disabled", input: proxyTopologyInput{}, network: proxyNetworkDisabled, bind: "127.0.0.1", host: "127.0.0.1"},
		{name: "Docker bridge", input: proxyTopologyInput{ExecutionEnvironmentEnabled: true, ContainerEngine: "docker"}, engine: "docker", network: proxyNetworkBridge, bind: "0.0.0.0", host: "gateway.docker.internal", mapping: dockerGatewayHostMapping},
		{name: "Podman bridge", input: proxyTopologyInput{ExecutionEnvironmentEnabled: true, ContainerEngine: "podman"}, engine: "podman", network: proxyNetworkBridge, bind: "0.0.0.0", host: "host.containers.internal"},
		{name: "Docker host network", input: proxyTopologyInput{ExecutionEnvironmentEnabled: true, ContainerEngine: "docker", ContainerOptions: []string{"--network=host"}}, engine: "docker", network: proxyNetworkHost, bind: "127.0.0.1", host: "127.0.0.1"},
		{name: "automatic Podman preference", input: proxyTopologyInput{ExecutionEnvironmentEnabled: true, AvailableEngines: map[string]bool{"podman": true, "docker": true}}, engine: "podman", network: proxyNetworkBridge, bind: "0.0.0.0", host: "host.containers.internal"},
		{name: "automatic Docker fallback", input: proxyTopologyInput{ExecutionEnvironmentEnabled: true, ContainerEngine: "auto", AvailableEngines: map[string]bool{"docker": true}}, engine: "docker", network: proxyNetworkBridge, bind: "0.0.0.0", host: "gateway.docker.internal", mapping: dockerGatewayHostMapping},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topology, err := resolveProxyTopology(tt.input)
			require.NoError(t, err)
			require.Equal(t, tt.engine, topology.ContainerEngine)
			require.Equal(t, tt.network, topology.NetworkMode)
			require.Equal(t, tt.bind, topology.BindAddress)
			require.Equal(t, tt.host, topology.AdvertisedHost)
			require.Equal(t, tt.mapping, topology.AutomaticHostMapping)
		})
	}
}

func TestResolveProxyTopologyPreservesExplicitValues(t *testing.T) {
	topology, err := resolveProxyTopology(proxyTopologyInput{
		ExecutionEnvironmentEnabled: true,
		ContainerEngine:             "docker",
		AnsibleProxyBindAddress:     "192.0.2.10",
		AnsibleProxyHost:            "host.example.test",
	})
	require.NoError(t, err)
	require.Equal(t, "192.0.2.10", topology.BindAddress)
	require.Equal(t, "host.example.test", topology.AdvertisedHost)
	require.Equal(t, proxyValueExplicit, topology.BindAddressOrigin)
	require.Equal(t, proxyValueExplicit, topology.AdvertisedHostOrigin)
	require.Empty(t, topology.AutomaticHostMapping)

	topology, err = resolveProxyTopology(proxyTopologyInput{
		ExecutionEnvironmentEnabled: true,
		ContainerEngine:             "podman",
		AnsibleProxyHost:            "host.example.test",
	})
	require.NoError(t, err)
	require.Equal(t, "0.0.0.0", topology.BindAddress)
	require.Equal(t, proxyValueInferred, topology.BindAddressOrigin)
	require.Equal(t, proxyValueExplicit, topology.AdvertisedHostOrigin)
}

func TestUsesHostNetworkSpellingsWithoutMutation(t *testing.T) {
	tests := [][]string{
		{"--network=host"},
		{"--network", "host"},
		{"--net=host"},
		{"--net", "host"},
	}
	for _, options := range tests {
		original := append([]string(nil), options...)
		require.True(t, usesHostNetwork(options), options)
		require.Equal(t, original, options)
	}
	require.False(t, usesHostNetwork([]string{"--network", "bridge", "--net=custom"}))
}

func TestResolveProxyTopologyValidationErrors(t *testing.T) {
	tests := []struct {
		name  string
		input proxyTopologyInput
		want  string
	}{
		{name: "no automatic engine", input: proxyTopologyInput{ExecutionEnvironmentEnabled: true}, want: "neither podman nor docker"},
		{name: "unsupported engine", input: proxyTopologyInput{ExecutionEnvironmentEnabled: true, ContainerEngine: "containerd"}, want: "unsupported"},
		{name: "bridge advertises loopback", input: proxyTopologyInput{ExecutionEnvironmentEnabled: true, ContainerEngine: "docker", AnsibleProxyHost: "127.0.0.1"}, want: "container loopback"},
		{name: "bridge listener is loopback", input: proxyTopologyInput{ExecutionEnvironmentEnabled: true, ContainerEngine: "docker", AnsibleProxyBindAddress: "127.0.0.1"}, want: "bind to a non-loopback address"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := resolveProxyTopology(tt.input)
			require.Error(t, err)
			require.True(t, strings.Contains(err.Error(), tt.want), err.Error())
		})
	}
}
