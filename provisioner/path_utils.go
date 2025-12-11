// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package provisioner

import (
	"os"
	"path/filepath"
	"strings"
)

// expandUserPath expands HOME-relative paths on the local side.
// It handles:
// - "~" -> $HOME
// - "~/subdir" -> $HOME/subdir
// - "~user/..." -> unchanged (no multi-user home resolution)
// - Other paths -> unchanged
//
// This function is intentionally limited to tilde expansion only and
// does NOT support general environment variable expansion or shell interpolation.
func expandUserPath(path string) string {
	if path == "" {
		return path
	}

	// Only expand if it starts with ~
	if !strings.HasPrefix(path, "~") {
		return path
	}

	// Don't expand ~user/ patterns (multi-user home directories)
	if len(path) > 1 && path[1] != '/' && path[1] != filepath.Separator {
		return path
	}

	// Get HOME directory
	home, err := os.UserHomeDir()
	if err != nil {
		// If we can't get HOME, return the path unchanged
		return path
	}

	// Handle bare "~"
	if path == "~" {
		return home
	}

	// Handle "~/..." pattern
	if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~"+string(filepath.Separator)) {
		return filepath.Join(home, path[2:])
	}

	// Shouldn't reach here, but return unchanged if we do
	return path
}
