// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansible

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"unicode"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// ensureCollections installs required Ansible collections if specified
func ensureCollections(ui packersdk.Ui, c *Config) error {
	// Return early if no collections are configured
	if len(c.Collections) == 0 && c.CollectionsRequirements == "" {
		return nil
	}

	ui.Say("Managing Ansible collections...")

	// Resolve cache directory
	cacheDir, err := resolveCollectionsCacheDir(c.CollectionsCacheDir)
	if err != nil {
		return fmt.Errorf("failed to resolve collections cache directory: %s", err)
	}
	c.CollectionsCacheDir = cacheDir

	// Ensure cache directory exists
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create collections cache directory: %s", err)
	}

	ui.Message(fmt.Sprintf("Collections cache directory: %s", cacheDir))

	// Check if ansible-galaxy is available
	if err := checkAnsibleGalaxy(); err != nil {
		return err
	}

	// Install from requirements file if specified
	if c.CollectionsRequirements != "" {
		if err := installFromRequirements(ui, c, cacheDir); err != nil {
			return err
		}
	}

	// Install individual collections if specified
	if len(c.Collections) > 0 {
		if err := installCollections(ui, c, cacheDir); err != nil {
			return err
		}
	}

	ui.Say("Ansible collections management completed successfully")
	return nil
}

// resolveCollectionsCacheDir resolves the cache directory path, expanding ~ and setting defaults
func resolveCollectionsCacheDir(cacheDir string) (string, error) {
	if cacheDir == "" {
		// Default to ~/.packer.d/ansible_collections_cache
		usr, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("could not determine current user: %s", err)
		}
		cacheDir = filepath.Join(usr.HomeDir, ".packer.d", "ansible_collections_cache")
	}

	// Expand ~ in path
	if strings.HasPrefix(cacheDir, "~") {
		usr, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("could not determine current user: %s", err)
		}
		cacheDir = filepath.Join(usr.HomeDir, cacheDir[1:])
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(cacheDir)
	if err != nil {
		return "", fmt.Errorf("failed to convert to absolute path: %s", err)
	}

	return absPath, nil
}

// checkAnsibleGalaxy verifies that ansible-galaxy is available
func checkAnsibleGalaxy() error {
	cmd := exec.Command("ansible-galaxy", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error: ansible-galaxy not found in PATH. Please install Ansible before using collections management")
	}
	return nil
}

// installFromRequirements installs collections from a requirements.yml file
func installFromRequirements(ui packersdk.Ui, c *Config, cacheDir string) error {
	requirementsPath := c.CollectionsRequirements

	// Validate requirements file exists
	if _, err := os.Stat(requirementsPath); os.IsNotExist(err) {
		return fmt.Errorf("collections_requirements file not found: %s", requirementsPath)
	}

	ui.Message(fmt.Sprintf("Installing collections from requirements file: %s", requirementsPath))

	args := []string{"collection", "install", "-r", requirementsPath, "-p", cacheDir}

	if c.CollectionsForceUpdate {
		args = append(args, "--force")
	}

	if c.CollectionsOffline {
		ui.Message("Offline mode: skipping requirements file installation")
		// In offline mode, we assume collections are already present
		return nil
	}

	return runAnsibleGalaxy(ui, c, args, "requirements file")
}

// installCollections installs individual collections
func installCollections(ui packersdk.Ui, c *Config, cacheDir string) error {
	for _, collection := range c.Collections {
		if err := installCollection(ui, c, collection, cacheDir); err != nil {
			return err
		}
	}
	return nil
}

// installCollection installs a single collection
func installCollection(ui packersdk.Ui, c *Config, collection string, cacheDir string) error {
	// Parse collection specification
	// Format: "namespace.name:version" or "namespace.name@/path/to/collection"
	var collectionSpec string
	var isLocalPath bool

	if strings.Contains(collection, "@") {
		// Local path installation
		parts := strings.SplitN(collection, "@", 2)
		collectionSpec = parts[1]
		isLocalPath = true
		ui.Message(fmt.Sprintf("Installing collection from local path: %s", collection))
	} else {
		collectionSpec = collection
		ui.Message(fmt.Sprintf("Installing collection: %s", collection))
	}

	// Check if collection is already installed
	if !c.CollectionsForceUpdate {
		if isCollectionInstalled(collection, cacheDir) {
			ui.Message(fmt.Sprintf("Collection '%s' already cached, skipping installation", collection))
			return nil
		}
	}

	// In offline mode, fail if collection is not cached
	if c.CollectionsOffline {
		if !isCollectionInstalled(collection, cacheDir) {
			return fmt.Errorf("Collection '%s' not found and offline mode is enabled", collection)
		}
		ui.Message(fmt.Sprintf("Collection '%s' found in cache (offline mode)", collection))
		return nil
	}

	// Build installation command
	args := []string{"collection", "install", collectionSpec, "-p", cacheDir}

	if c.CollectionsForceUpdate {
		args = append(args, "--force")
	}

	if err := runAnsibleGalaxy(ui, c, args, collection); err != nil {
		return fmt.Errorf("Failed to install collection '%s': %s", collection, err)
	}

	// Validate installation
	if !isLocalPath {
		if !isCollectionInstalled(collection, cacheDir) {
			ui.Error(fmt.Sprintf("Reinstalling collection '%s' due to missing MANIFEST.json", collection))
			// Force reinstall
			args = append(args, "--force")
			if err := runAnsibleGalaxy(ui, c, args, collection); err != nil {
				return fmt.Errorf("Failed to reinstall collection '%s': %s", collection, err)
			}
		}
	}

	return nil
}

// isCollectionInstalled checks if a collection is installed by looking for MANIFEST.json
func isCollectionInstalled(collection string, cacheDir string) bool {
	// Parse collection name (remove version if present)
	collectionName := collection
	if strings.Contains(collectionName, ":") {
		collectionName = strings.Split(collectionName, ":")[0]
	}
	if strings.Contains(collectionName, "@") {
		collectionName = strings.Split(collectionName, "@")[0]
	}

	// Parse namespace and name
	parts := strings.Split(collectionName, ".")
	if len(parts) != 2 {
		return false
	}

	namespace := parts[0]
	name := parts[1]

	// Check for MANIFEST.json
	manifestPath := filepath.Join(cacheDir, "ansible_collections", namespace, name, "MANIFEST.json")
	_, err := os.Stat(manifestPath)
	return err == nil
}

// runAnsibleGalaxy executes an ansible-galaxy command with streaming output
func runAnsibleGalaxy(ui packersdk.Ui, c *Config, args []string, target string) error {
	cmd := exec.Command("ansible-galaxy", args...)

	// Set environment variables
	cmd.Env = os.Environ()
	if len(c.AnsibleEnvVars) > 0 {
		cmd.Env = append(cmd.Env, c.AnsibleEnvVars...)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %s", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %s", err)
	}

	wg := sync.WaitGroup{}
	repeat := func(r io.ReadCloser) {
		reader := bufio.NewReader(r)
		for {
			line, err := reader.ReadString('\n')
			if line != "" {
				line = strings.TrimRightFunc(line, unicode.IsSpace)
				ui.Message(line)
			}
			if err != nil {
				if err == io.EOF {
					break
				} else {
					ui.Error(err.Error())
					break
				}
			}
		}
		wg.Done()
	}

	wg.Add(2)
	go repeat(stdout)
	go repeat(stderr)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ansible-galaxy: %s", err)
	}

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ansible-galaxy exited with error for %s: %s", target, err)
	}

	return nil
}
