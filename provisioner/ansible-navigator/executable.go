// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansiblenavigator

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

const executableResolutionTimeout = 10 * time.Second

// lookupExecutable searches the child's PATH before constructing exec.Cmd. A
// child environment cannot change the lookup already done by exec.Command.
func lookupExecutable(command string, env []string, preparationDir string) (string, error) {
	if strings.IndexFunc(command, unicode.IsSpace) >= 0 || command == "" {
		return "", fmt.Errorf("command must be only the executable name or path (no arguments)")
	}
	if preparationDir == "" {
		var err error
		preparationDir, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}
	command = expandUserPath(command)
	check := func(candidate string) (string, error) {
		if !filepath.IsAbs(candidate) {
			candidate = filepath.Join(preparationDir, candidate)
		}
		return exec.LookPath(candidate)
	}
	if strings.ContainsAny(command, `/\\`) {
		return check(command)
	}
	var path string
	for _, entry := range env {
		if strings.HasPrefix(entry, "PATH=") {
			path = strings.TrimPrefix(entry, "PATH=")
		}
	}
	for _, dir := range filepath.SplitList(path) {
		if found, err := check(filepath.Join(dir, command)); err == nil {
			return found, nil
		}
	}
	return "", fmt.Errorf("executable %q not found or not executable on effective PATH", command)
}

func resolveShimInEnvironment(ctx context.Context, command, manager string, env []string, preparationDir string) (string, error) {
	switch manager {
	case "asdf", "rbenv", "pyenv":
	default:
		return "", fmt.Errorf("unsupported version manager: %s", manager)
	}
	managerPath, err := lookupExecutable(manager, env, preparationDir)
	if err != nil {
		return "", fmt.Errorf("%s manager lookup failed", manager)
	}
	cmd := exec.CommandContext(ctx, managerPath, "which", filepath.Base(command))
	cmd.Env = env
	cmd.Dir = preparationDir
	// Bound inherited output pipes too if a manager leaves a child running.
	cmd.WaitDelay = 50 * time.Millisecond
	output, err := cmd.Output()
	if ctx.Err() != nil {
		return "", fmt.Errorf("%s shim resolution timed out: %w", manager, ctx.Err())
	}
	if err != nil {
		return "", fmt.Errorf("%s shim resolution subprocess failed: %w", manager, err)
	}
	resolved := strings.TrimSpace(string(output))
	if resolved == "" {
		return "", fmt.Errorf("%s returned empty path", manager)
	}
	if !filepath.IsAbs(resolved) {
		return "", fmt.Errorf("%s returned a non-absolute executable path", manager)
	}
	resolved, err = lookupExecutable(resolved, env, preparationDir)
	if err != nil {
		return "", fmt.Errorf("%s returned an invalid executable", manager)
	}
	if _, shim := detectShim(resolved); shim {
		return "", fmt.Errorf("%s returned an unresolved shim chain", manager)
	}
	return resolved, nil
}

func (p *Provisioner) resolveExecutable(ctx context.Context) error {
	if p.resolvedCommand != "" {
		return nil
	}
	env := buildEnvWithPath(p.config.AnsibleNavigatorPath)
	selected, err := lookupExecutable(p.config.Command, env, p.preparationDir)
	if err != nil {
		return fmt.Errorf("Navigator executable lookup failed for command %q using ansible_navigator_path then inherited PATH: %w; set command to the executable's full path", p.config.Command, err)
	}
	if manager, shim := detectShim(selected); shim {
		resolutionCtx, cancel := context.WithTimeout(ctx, executableResolutionTimeout)
		defer cancel()
		selected, err = resolveShimInEnvironment(resolutionCtx, selected, manager, env, p.preparationDir)
		if err != nil {
			return fmt.Errorf("Navigator shim resolution failed for command %q: %w; set command to the real executable's full path (find it with %s which %s)", p.config.Command, err, manager, filepath.Base(p.config.Command))
		}
	}
	p.resolvedCommand = selected
	return nil
}
