// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

//go:build !windows

package ansiblenavigator

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const defaultProxyConnectivityImage = "busybox:1.36.1"

func proxyConnectivityPrerequisites(t *testing.T) (string, string) {
	t.Helper()
	if os.Getenv("PACKER_ACC") != "1" {
		t.Skip("PACKER_ACC=1 is required for real-container proxy connectivity tests")
	}
	if runtime.GOOS != "linux" {
		t.Skipf("real-container proxy connectivity tests require Linux; running on %s", runtime.GOOS)
	}
	containerRuntime := os.Getenv("PACKER_ANSIBLE_NAVIGATOR_CONTAINER_RUNTIME")
	if containerRuntime == "" {
		containerRuntime = "docker"
	}
	runtimePath, err := exec.LookPath(containerRuntime)
	if err != nil {
		t.Skipf("configured container runtime %q is unavailable: %v", containerRuntime, err)
	}
	image := os.Getenv("PACKER_ANSIBLE_NAVIGATOR_PROXY_TEST_IMAGE")
	if image == "" {
		image = defaultProxyConnectivityImage
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if output, err := exec.CommandContext(ctx, runtimePath, "image", "inspect", image).CombinedOutput(); err != nil {
		t.Skipf("proxy test image %q is not available locally (tests never pull it): %v (%s)", image, err, strings.TrimSpace(string(output)))
	}
	return runtimePath, image
}

func TestAccProxyConnectivityDockerBridge(t *testing.T) {
	runtimePath, image := proxyConnectivityPrerequisites(t)
	if filepath.Base(runtimePath) != "docker" {
		t.Skipf("Docker bridge host-gateway coverage requires docker; configured runtime is %q", runtimePath)
	}

	topology, err := resolveProxyTopology(proxyTopologyInput{ExecutionEnvironmentEnabled: true, ContainerEngine: "docker"})
	if err != nil {
		t.Fatal(err)
	}
	runProxyConnectivityContainer(t, runtimePath, image, topology)
}

func TestAccProxyConnectivityHostNetwork(t *testing.T) {
	runtimePath, image := proxyConnectivityPrerequisites(t)
	engine := filepath.Base(runtimePath)
	if engine != "docker" && engine != "podman" {
		t.Skipf("host-network coverage supports docker or podman; configured runtime is %q", runtimePath)
	}
	topology, err := resolveProxyTopology(proxyTopologyInput{
		ExecutionEnvironmentEnabled: true,
		ContainerEngine:             engine,
		ContainerOptions:            []string{"--network=host"},
	})
	if err != nil {
		t.Fatal(err)
	}
	runProxyConnectivityContainer(t, runtimePath, image, topology)
}

func runProxyConnectivityContainer(t *testing.T, runtimePath, image string, topology proxyTopology) {
	t.Helper()
	listener, err := net.Listen("tcp", net.JoinHostPort(topology.BindAddress, "0"))
	if err != nil {
		t.Fatalf("listen on resolved bind address %s: %v", topology.BindAddress, err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	_, portText, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	containerName := fmt.Sprintf("packer-ansible-navigator-proxy-%d", time.Now().UnixNano())
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = exec.CommandContext(ctx, runtimePath, "rm", "-f", containerName).Run()
	})

	args := []string{"run", "--rm", "--name", containerName}
	if topology.NetworkMode == proxyNetworkHost {
		args = append(args, "--network=host")
	}
	if topology.AutomaticHostMapping != "" {
		args = append(args, topology.AutomaticHostMapping)
	}
	args = append(args, image, "sh", "-c", "nc -z -w 5 \"$1\" \"$2\"", "proxy-check", topology.AdvertisedHost, portText)

	connected := make(chan struct{}, 1)
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr == nil {
			_ = conn.Close()
			connected <- struct{}{}
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, runtimePath, args...).CombinedOutput()
	if err != nil {
		if topology.NetworkMode == proxyNetworkHost {
			t.Skipf("runtime does not provide usable host networking: %v (%s)", err, strings.TrimSpace(string(output)))
		}
		t.Fatalf("container could not reach resolved bridge proxy route: %v (%s)", err, strings.TrimSpace(string(output)))
	}
	select {
	case <-connected:
	case <-time.After(5 * time.Second):
		t.Fatalf("container command succeeded but listener did not observe a connection on port %s", portText)
	}
}
