// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansiblenavigator

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"unicode"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// GalaxyManager handles all Ansible Galaxy operations for roles and collections
type GalaxyManager struct {
	config  *Config
	ui      packersdk.Ui
	envVars []string
}

// NewGalaxyManager creates a new GalaxyManager instance
func NewGalaxyManager(config *Config, ui packersdk.Ui) *GalaxyManager {
	return &GalaxyManager{
		config:  config,
		ui:      ui,
		envVars: []string{},
	}
}

// InstallRequirements installs all requirements (roles and collections) based on configuration
func (gm *GalaxyManager) InstallRequirements() error {
	// requirements_file is the only supported dependency installation mechanism.
	if gm.config.RequirementsFile == "" {
		return nil
	}

	gm.ui.Message(fmt.Sprintf("Installing dependencies from requirements file: %s", gm.config.RequirementsFile))
	if err := gm.installFromFile(gm.config.RequirementsFile); err != nil {
		return fmt.Errorf("failed to install requirements: %w", err)
	}
	return nil
}

// installFromFile installs roles and/or collections from a requirements file
func (gm *GalaxyManager) installFromFile(filePath string) error {
	// Validate file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("requirements file not found: %s", filePath)
	} else if err != nil {
		return fmt.Errorf("error checking requirements file: %w", err)
	}

	// Check offline mode
	if gm.config.OfflineMode {
		gm.ui.Message("Offline mode enabled: skipping network operations")
		return nil
	}

	// Read file to determine content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read requirements file: %w", err)
	}

	hasRoles := regexp.MustCompile(`(?m)^roles:`).Match(content)
	hasCollections := regexp.MustCompile(`(?m)^collections:`).Match(content)

	// Install roles if present (or if it's v1 format without collections)
	if hasRoles || !hasCollections {
		if err := gm.installRolesFromFile(filePath); err != nil {
			return err
		}
	}

	// Install collections if present
	if hasCollections {
		if err := gm.installCollectionsFromFile(filePath); err != nil {
			return err
		}
	}

	if !hasRoles && !hasCollections {
		gm.ui.Message("Warning: requirements file does not contain 'roles:' or 'collections:' sections")
	}

	return nil
}

// installRolesFromFile installs roles from a requirements file
func (gm *GalaxyManager) installRolesFromFile(filePath string) error {
	gm.ui.Message("Installing roles from requirements file...")
	args := []string{"install", "-r", filepath.ToSlash(filePath)}

	// Add roles path if specified
	if gm.config.RolesCacheDir != "" {
		args = append(args, "-p", gm.config.RolesCacheDir)
	}

	// Add force options
	if gm.config.GalaxyForceInstall || gm.config.ForceUpdate {
		args = append(args, "--force")
	}
	if gm.config.GalaxyForceWithDeps {
		args = append(args, "--force-with-deps")
	}

	return gm.executeGalaxyCommand(args, "roles")
}

// installCollectionsFromFile installs collections from a requirements file
func (gm *GalaxyManager) installCollectionsFromFile(filePath string) error {
	gm.ui.Message("Installing collections from requirements file...")
	args := []string{"collection", "install", "-r", filepath.ToSlash(filePath)}

	// Add collections path if specified
	if gm.config.CollectionsCacheDir != "" {
		args = append(args, "-p", gm.config.CollectionsCacheDir)
	}

	// Add force options
	if gm.config.GalaxyForceInstall || gm.config.ForceUpdate {
		args = append(args, "--force")
	}
	if gm.config.GalaxyForceWithDeps {
		args = append(args, "--force-with-deps")
	}

	return gm.executeGalaxyCommand(args, "collections")
}

// NOTE: legacy inline collections and legacy galaxy_file paths are intentionally removed.

// executeGalaxyCommand executes an ansible-galaxy command with streaming output
func (gm *GalaxyManager) executeGalaxyCommand(args []string, target string) error {
	cmd := exec.Command("ansible-galaxy", args...)

	// Set environment
	cmd.Env = os.Environ()
	if len(gm.envVars) > 0 {
		cmd.Env = append(cmd.Env, gm.envVars...)
	}

	// Setup pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Output handler
	var wg sync.WaitGroup
	outputHandler := func(r io.ReadCloser) {
		defer wg.Done()
		reader := bufio.NewReader(r)
		for {
			line, err := reader.ReadString('\n')
			if line != "" {
				line = strings.TrimRightFunc(line, unicode.IsSpace)
				gm.ui.Message(line)
			}
			if err != nil {
				if err != io.EOF {
					gm.ui.Error(err.Error())
				}
				break
			}
		}
	}

	wg.Add(2)
	go outputHandler(stdout)
	go outputHandler(stderr)

	// Execute command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ansible-galaxy: %w", err)
	}

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ansible-galaxy failed for %s: %w", target, err)
	}

	return nil
}

// SetupEnvironmentPaths configures ANSIBLE_COLLECTIONS_PATHS and ANSIBLE_ROLES_PATH
func (gm *GalaxyManager) SetupEnvironmentPaths() error {
	// Set collections path
	if gm.config.CollectionsCacheDir != "" {
		if err := gm.setEnvironmentPath("ANSIBLE_COLLECTIONS_PATHS", gm.config.CollectionsCacheDir); err != nil {
			return err
		}
	}

	// Set roles path
	if gm.config.RolesCacheDir != "" {
		if err := gm.setEnvironmentPath("ANSIBLE_ROLES_PATH", gm.config.RolesCacheDir); err != nil {
			return err
		}
	}

	return nil
}

// setEnvironmentPath sets an environment variable, prepending to existing value if present
func (gm *GalaxyManager) setEnvironmentPath(envVar, path string) error {
	existing := os.Getenv(envVar)

	var newPath string
	if existing != "" {
		newPath = path + ":" + existing
	} else {
		newPath = path
	}

	if err := os.Setenv(envVar, newPath); err != nil {
		return fmt.Errorf("failed to set %s: %w", envVar, err)
	}

	return nil
}
