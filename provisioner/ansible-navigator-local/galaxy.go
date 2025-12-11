// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansiblenavigatorlocal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// GalaxyManager handles all Ansible Galaxy operations for roles and collections
// Adapted for local provisioner communicator-based execution
type GalaxyManager struct {
	config                *Config
	ui                    packersdk.Ui
	comm                  packersdk.Communicator
	stagingDir            string
	galaxyRolesPath       string
	galaxyCollectionsPath string
}

// NewGalaxyManager creates a new GalaxyManager instance
func NewGalaxyManager(config *Config, ui packersdk.Ui, comm packersdk.Communicator, stagingDir string, galaxyRolesPath string, galaxyCollectionsPath string) *GalaxyManager {
	return &GalaxyManager{
		config:                config,
		ui:                    ui,
		comm:                  comm,
		stagingDir:            stagingDir,
		galaxyRolesPath:       galaxyRolesPath,
		galaxyCollectionsPath: galaxyCollectionsPath,
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
	// Validate file exists locally
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

	// Construct remote path for requirements file
	remoteReqFile := filepath.ToSlash(filepath.Join(gm.stagingDir, filepath.Base(filePath)))

	// Install roles if present (or if it's v1 format without collections)
	if hasRoles || !hasCollections {
		if err := gm.installRolesFromFile(remoteReqFile); err != nil {
			return err
		}
	}

	// Install collections if present
	if hasCollections {
		if err := gm.installCollectionsFromFile(remoteReqFile); err != nil {
			return err
		}
	}

	if !hasRoles && !hasCollections {
		gm.ui.Message("Warning: requirements file does not contain 'roles:' or 'collections:' sections")
	}

	return nil
}

// installRolesFromFile installs roles from a requirements file
func (gm *GalaxyManager) installRolesFromFile(remoteFilePath string) error {
	gm.ui.Message("Installing roles from requirements file...")
	args := []string{"install", "-r", remoteFilePath}

	// Add roles path
	rolesPath := gm.galaxyRolesPath
	if gm.config.RolesCacheDir != "" {
		rolesPath = gm.config.RolesCacheDir
	}
	if rolesPath != "" {
		args = append(args, "-p", filepath.ToSlash(rolesPath))
	}

	// Add force options
	if gm.config.GalaxyForceInstall || gm.config.ForceUpdate {
		args = append(args, "--force")
	}

	return gm.executeGalaxyCommand(args, "roles")
}

// installCollectionsFromFile installs collections from a requirements file
func (gm *GalaxyManager) installCollectionsFromFile(remoteFilePath string) error {
	gm.ui.Message("Installing collections from requirements file...")
	args := []string{"collection", "install", "-r", remoteFilePath}

	// Add collections path
	collectionsPath := gm.galaxyCollectionsPath
	if gm.config.CollectionsCacheDir != "" {
		collectionsPath = gm.config.CollectionsCacheDir
	}
	if collectionsPath != "" {
		args = append(args, "-p", filepath.ToSlash(collectionsPath))
	}

	// Add force options
	if gm.config.GalaxyForceInstall || gm.config.ForceUpdate {
		args = append(args, "--force")
	}

	return gm.executeGalaxyCommand(args, "collections")
}

// NOTE: legacy inline collections and legacy galaxy_file paths are intentionally removed.

// executeGalaxyCommand executes an ansible-galaxy command via communicator
func (gm *GalaxyManager) executeGalaxyCommand(args []string, target string) error {
	ctx := context.TODO()
	command := fmt.Sprintf("cd %s && %s %s",
		gm.stagingDir, "ansible-galaxy", strings.Join(args, " "))
	gm.ui.Message(fmt.Sprintf("Executing Ansible Galaxy: %s", command))

	cmd := &packersdk.RemoteCmd{
		Command: command,
	}
	if err := cmd.RunWithUi(ctx, gm.comm, gm.ui); err != nil {
		return err
	}
	if cmd.ExitStatus() != 0 {
		return fmt.Errorf("ansible-galaxy failed for %s: exit code %d", target, cmd.ExitStatus())
	}

	return nil
}

// SetupEnvironmentPaths configures ANSIBLE_COLLECTIONS_PATHS and ANSIBLE_ROLES_PATH
// Returns environment variable strings to be prepended to commands
func (gm *GalaxyManager) SetupEnvironmentPaths() []string {
	envVars := []string{}

	// Set collections path
	if gm.config.CollectionsCacheDir != "" {
		envVars = append(envVars, fmt.Sprintf("ANSIBLE_COLLECTIONS_PATH=$ANSIBLE_COLLECTIONS_PATH:%s",
			gm.config.CollectionsCacheDir))
	}

	// Set roles path
	if gm.config.RolesCacheDir != "" {
		envVars = append(envVars, fmt.Sprintf("ANSIBLE_ROLES_PATH=$ANSIBLE_ROLES_PATH:%s",
			gm.config.RolesCacheDir))
	}

	return envVars
}
