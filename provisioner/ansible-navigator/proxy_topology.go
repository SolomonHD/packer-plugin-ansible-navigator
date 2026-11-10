// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansiblenavigator

import (
	"fmt"
	"net"
	"strings"
)

type proxyNetworkMode string

const (
	proxyNetworkDisabled proxyNetworkMode = "disabled"
	proxyNetworkBridge   proxyNetworkMode = "bridge"
	proxyNetworkHost     proxyNetworkMode = "host"
)

type proxyValueOrigin string

const (
	proxyValueExplicit proxyValueOrigin = "explicit"
	proxyValueInferred proxyValueOrigin = "inferred"
)

const dockerGatewayHostMapping = "--add-host=gateway.docker.internal:host-gateway"

// navigatorContainerEnginePreference is the documented ansible-navigator
// automatic selection order: Podman first, then Docker.
func navigatorContainerEnginePreference() [2]string {
	return [2]string{"podman", "docker"}
}

type proxyTopologyInput struct {
	ExecutionEnvironmentEnabled bool
	ContainerEngine             string
	ContainerOptions            []string
	AnsibleProxyBindAddress     string
	AnsibleProxyHost            string
	AvailableEngines            map[string]bool
}

type proxyTopology struct {
	ExecutionEnvironmentEnabled bool
	ContainerEngine             string
	NetworkMode                 proxyNetworkMode
	BindAddress                 string
	AdvertisedHost              string
	BindAddressOrigin           proxyValueOrigin
	AdvertisedHostOrigin        proxyValueOrigin
	AutomaticHostMapping        string
}

// resolveProxyTopology is side-effect free. Callers provide the engines found
// on PATH so executable probing remains separate from topology policy.
func resolveProxyTopology(input proxyTopologyInput) (proxyTopology, error) {
	topology := proxyTopology{
		ExecutionEnvironmentEnabled: input.ExecutionEnvironmentEnabled,
		NetworkMode:                 proxyNetworkDisabled,
		BindAddress:                 input.AnsibleProxyBindAddress,
		AdvertisedHost:              input.AnsibleProxyHost,
		BindAddressOrigin:           proxyValueExplicit,
		AdvertisedHostOrigin:        proxyValueExplicit,
	}

	if input.AnsibleProxyBindAddress == "" {
		topology.BindAddressOrigin = proxyValueInferred
	}
	if input.AnsibleProxyHost == "" {
		topology.AdvertisedHostOrigin = proxyValueInferred
	}

	if !input.ExecutionEnvironmentEnabled {
		applyProxyDefaults(&topology, "127.0.0.1", "127.0.0.1")
		return topology, nil
	}

	engine, err := resolveContainerEngine(input.ContainerEngine, input.AvailableEngines)
	if err != nil {
		return proxyTopology{}, err
	}
	topology.ContainerEngine = engine
	topology.NetworkMode = proxyNetworkBridge
	if usesHostNetwork(input.ContainerOptions) {
		topology.NetworkMode = proxyNetworkHost
		applyProxyDefaults(&topology, "127.0.0.1", "127.0.0.1")
		return topology, nil
	}

	switch engine {
	case "docker":
		applyProxyDefaults(&topology, "0.0.0.0", "gateway.docker.internal")
		if topology.AdvertisedHost == "gateway.docker.internal" {
			topology.AutomaticHostMapping = dockerGatewayHostMapping
		}
	case "podman":
		applyProxyDefaults(&topology, "0.0.0.0", "host.containers.internal")
	}

	if isLoopbackHost(topology.AdvertisedHost) {
		return proxyTopology{}, fmt.Errorf(
			"ansible_proxy_host %q is unreachable from a bridged %s execution environment: container loopback cannot reach the host proxy; omit ansible_proxy_host, set a container-reachable host, or add --network=host to container_options",
			topology.AdvertisedHost, engine)
	}
	if isLoopbackHost(topology.BindAddress) {
		return proxyTopology{}, fmt.Errorf(
			"ansible_proxy_bind_address %q is unreachable through a bridged %s execution environment: bind to a non-loopback address, omit ansible_proxy_bind_address, or add --network=host to container_options",
			topology.BindAddress, engine)
	}

	return topology, nil
}

func resolveContainerEngine(requested string, available map[string]bool) (string, error) {
	engine := strings.ToLower(strings.TrimSpace(requested))
	if engine != "" && engine != "auto" {
		if engine != "podman" && engine != "docker" {
			return "", fmt.Errorf("unsupported execution-environment container_engine %q: expected auto, podman, or docker", requested)
		}
		return engine, nil
	}

	for _, candidate := range navigatorContainerEnginePreference() {
		if available[candidate] {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("execution environment container_engine is auto, but neither podman nor docker was found in PATH; install a supported engine or set navigator_config.execution_environment.container_engine explicitly")
}

func usesHostNetwork(options []string) bool {
	for i, option := range options {
		if option == "--network=host" || option == "--net=host" {
			return true
		}
		if (option == "--network" || option == "--net") && i+1 < len(options) && options[i+1] == "host" {
			return true
		}
	}
	return false
}

func applyProxyDefaults(topology *proxyTopology, bindAddress, advertisedHost string) {
	if topology.BindAddress == "" {
		topology.BindAddress = bindAddress
	}
	if topology.AdvertisedHost == "" {
		topology.AdvertisedHost = advertisedHost
	}
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(strings.TrimSpace(host), "localhost") {
		return true
	}
	ip := net.ParseIP(strings.TrimSpace(host))
	return ip != nil && ip.IsLoopback()
}
