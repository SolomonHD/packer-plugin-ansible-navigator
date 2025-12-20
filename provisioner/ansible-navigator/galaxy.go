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
	config *Config
	ui     packersdk.Ui
}

// NewGalaxyManager creates a new GalaxyManager instance
func NewGalaxyManager(config *Config, ui packersdk.Ui) *GalaxyManager {
	return &GalaxyManager{
		config: config,
		ui:     ui,
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
	args := []string{"install", fmt.Sprintf("-r=%s", filepath.ToSlash(filePath))}

	// Add roles path if specified
	if gm.config.RolesPath != "" {
		args = append(args, fmt.Sprintf("-p=%s", gm.config.RolesPath))
	}

	// Add offline option
	if gm.config.OfflineMode {
		args = append(args, "--offline")
	}

	// Add force options
	if gm.config.GalaxyForceWithDeps {
		args = append(args, "--force-with-deps")
	} else if gm.config.GalaxyForce {
		args = append(args, "--force")
	}

	// Append user-provided args last
	args = append(args, gm.config.GalaxyArgs...)

	return gm.executeGalaxyCommand(args, "roles")
}

// installCollectionsFromFile installs collections from a requirements file
func (gm *GalaxyManager) installCollectionsFromFile(filePath string) error {
	gm.ui.Message("Installing collections from requirements file...")
	args := []string{"collection", "install", fmt.Sprintf("-r=%s", filepath.ToSlash(filePath))}

	// Add collections path if specified
	if gm.config.CollectionsPath != "" {
		args = append(args, fmt.Sprintf("-p=%s", gm.config.CollectionsPath))
	}

	// Add offline option
	if gm.config.OfflineMode {
		args = append(args, "--offline")
	}

	// Add force options
	if gm.config.GalaxyForceWithDeps {
		args = append(args, "--force-with-deps")
	} else if gm.config.GalaxyForce {
		args = append(args, "--force")
	}

	// Append user-provided args last
	args = append(args, gm.config.GalaxyArgs...)

	return gm.executeGalaxyCommand(args, "collections")
}

// NOTE: legacy inline collections and legacy galaxy_file paths are intentionally removed.

// executeGalaxyCommand executes an ansible-galaxy command with streaming output
func (gm *GalaxyManager) executeGalaxyCommand(args []string, target string) error {
	cmd := exec.Command(gm.config.GalaxyCommand, args...)
	cmd.Env = os.Environ()

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

// SetupEnvironmentPaths configures ANSIBLE_COLLECTIONS_PATH and ANSIBLE_ROLES_PATH
func (gm *GalaxyManager) SetupEnvironmentPaths() error {
	// Treat roles_path/collections_path as opaque strings.
	// They are used both for Galaxy install destinations (-p) and Ansible discovery (env vars).
	if gm.config.CollectionsPath != "" {
		if err := os.Setenv("ANSIBLE_COLLECTIONS_PATH", gm.config.CollectionsPath); err != nil {
			return fmt.Errorf("failed to set ANSIBLE_COLLECTIONS_PATH: %w", err)
		}
	}
	if gm.config.RolesPath != "" {
		if err := os.Setenv("ANSIBLE_ROLES_PATH", gm.config.RolesPath); err != nil {
			return fmt.Errorf("failed to set ANSIBLE_ROLES_PATH: %w", err)
		}
	}
	return nil
}
