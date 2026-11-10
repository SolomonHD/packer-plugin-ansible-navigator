// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansiblenavigator

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"time"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

const dockerContextTimeout = 10 * time.Second

// dockerEnvironmentValue observes exec.Cmd's last-value-wins environment semantics.
func dockerEnvironmentValue(env []string, key string) string {
	for i := len(env) - 1; i >= 0; i-- {
		if value, ok := strings.CutPrefix(env[i], key+"="); ok {
			return value
		}
	}
	return ""
}

// resolveDockerHost performs optional discovery, never a daemon health check.
// Docker retains its own context/host/TLS precedence through the unchanged env.
func resolveDockerHost(ui packersdk.Ui, env []string) (string, error) {
	return resolveDockerHostWithTimeout(ui, env, dockerContextTimeout)
}

func resolveDockerHostWithTimeout(ui packersdk.Ui, env []string, timeout time.Duration) (string, error) {
	if dockerEnvironmentValue(env, "DOCKER_HOST") != "" {
		return "", nil
	}
	dockerPath, err := lookupExecutable("docker", env, "")
	if err != nil {
		// Absence is optional, not evidence of Docker availability.
		return "", nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, dockerPath, "context", "inspect", "--format", "{{.Endpoints.docker.Host}}")
	cmd.Env = env
	// A misbehaving wrapper's descendants must not keep the output pipe open.
	cmd.WaitDelay = 100 * time.Millisecond
	output, err := cmd.Output()
	if err != nil {
		// Neither subprocess output nor errors are safe to print: they may include
		// credentials embedded in endpoint URLs or Docker configuration paths.
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			ui.Message("Warning: Docker context discovery timed out; continuing without a discovered endpoint. Check Docker selection in the Navigator environment.")
		} else {
			ui.Message("Warning: Docker context discovery failed; continuing without a discovered endpoint. Check Docker selection in the Navigator environment.")
		}
		return "", nil
	}
	return strings.TrimSpace(string(output)), nil
}
