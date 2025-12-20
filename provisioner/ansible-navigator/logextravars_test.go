// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansiblenavigator

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestLogExtraVarsJSON_ProducesValidJSON verifies that logExtraVarsJSON produces valid formatted JSON
func TestLogExtraVarsJSON_ProducesValidJSON(t *testing.T) {
	ui := newMockUi().(*mockUi)
	extraVars := map[string]interface{}{
		"packer_build_name":            "example-build",
		"packer_builder_type":          "docker",
		"ansible_ssh_private_key_file": "/tmp/key",
		"packer_http_addr":             "127.0.0.1:8080",
	}

	logExtraVarsJSON(ui, extraVars)

	// Should have received exactly one message
	require.Len(t, ui.messageMessages, 1)
	message := ui.messageMessages[0]

	// Should start with the prefix
	require.True(t, strings.HasPrefix(message, "[Extra Vars]"))

	// Extract JSON content
	jsonStart := strings.Index(message, "\n")
	require.NotEqual(t, -1, jsonStart)
	jsonContent := message[jsonStart+1:]

	// Should be valid JSON
	var parsed map[string]interface{}
	err := json.Unmarshal([]byte(jsonContent), &parsed)
	require.NoError(t, err)

	// Should contain all fields
	require.Equal(t, "example-build", parsed["packer_build_name"])
	require.Equal(t, "docker", parsed["packer_builder_type"])
	require.Equal(t, "/tmp/key", parsed["ansible_ssh_private_key_file"])
	require.Equal(t, "127.0.0.1:8080", parsed["packer_http_addr"])
}

// TestLogExtraVarsJSON_RedactsPasswords verifies that sensitive values are redacted
func TestLogExtraVarsJSON_RedactsPasswords(t *testing.T) {
	ui := newMockUi().(*mockUi)
	extraVars := map[string]interface{}{
		"ansible_password":    "secret123",
		"winrm_password":      "winrmpass",
		"packer_builder_type": "docker",
	}

	logExtraVarsJSON(ui, extraVars)

	// Should have received exactly one message
	require.Len(t, ui.messageMessages, 1)
	message := ui.messageMessages[0]

	// Extract JSON content
	jsonStart := strings.Index(message, "\n")
	require.NotEqual(t, -1, jsonStart)
	jsonContent := message[jsonStart+1:]

	// Should be valid JSON
	var parsed map[string]interface{}
	err := json.Unmarshal([]byte(jsonContent), &parsed)
	require.NoError(t, err)

	// Passwords should be redacted
	require.Equal(t, "*****", parsed["ansible_password"])
	require.Equal(t, "*****", parsed["winrm_password"])

	// Non-sensitive fields should not be redacted
	require.Equal(t, "docker", parsed["packer_builder_type"])

	// Original secrets should NOT appear in the message
	require.NotContains(t, message, "secret123")
	require.NotContains(t, message, "winrmpass")
}

// TestLogExtraVarsJSON_PreservesPrivateKeyPath verifies that private key file path is shown
func TestLogExtraVarsJSON_PreservesPrivateKeyPath(t *testing.T) {
	ui := newMockUi().(*mockUi)
	extraVars := map[string]interface{}{
		"ansible_ssh_private_key_file": "/home/user/.ssh/id_rsa",
		"ansible_password":             "secret",
	}

	logExtraVarsJSON(ui, extraVars)

	// Should have received exactly one message
	require.Len(t, ui.messageMessages, 1)
	message := ui.messageMessages[0]

	// Extract JSON content
	jsonStart := strings.Index(message, "\n")
	require.NotEqual(t, -1, jsonStart)
	jsonContent := message[jsonStart+1:]

	// Should be valid JSON
	var parsed map[string]interface{}
	err := json.Unmarshal([]byte(jsonContent), &parsed)
	require.NoError(t, err)

	// Private key path should be preserved (path is not secret, content is)
	require.Equal(t, "/home/user/.ssh/id_rsa", parsed["ansible_ssh_private_key_file"])

	// Password should still be redacted
	require.Equal(t, "*****", parsed["ansible_password"])
}

// TestShowExtraVars_DisabledByDefault verifies that feature is disabled when ShowExtraVars is false
func TestShowExtraVars_DisabledByDefault(t *testing.T) {
	p := &Provisioner{}
	p.config.PackerBuilderType = "docker"
	p.config.ShowExtraVars = false // Explicitly disabled

	// createCmdArgs relies on generatedData for a few conditionals.
	p.generatedData = map[string]interface{}{
		"ConnType": "ssh",
	}

	ui := newMockUi().(*mockUi)
	_, _ = p.createCmdArgs(ui, "127.0.0.1:8080", "/tmp/inventory.ini", "/tmp/key")

	// Should not have received any extra vars messages
	for _, msg := range ui.messageMessages {
		require.NotContains(t, msg, "[Extra Vars]")
	}
}

// TestShowExtraVars_EnabledWhenConfigured verifies that feature works when enabled
func TestShowExtraVars_EnabledWhenConfigured(t *testing.T) {
	p := &Provisioner{}
	p.config.PackerBuilderType = "docker"
	p.config.PackerBuildName = "example-build"
	p.config.ShowExtraVars = true // Explicitly enabled

	// createCmdArgs relies on generatedData for a few conditionals.
	p.generatedData = map[string]interface{}{
		"ConnType": "ssh",
	}

	ui := newMockUi().(*mockUi)
	_, _ = p.createCmdArgs(ui, "127.0.0.1:8080", "/tmp/inventory.ini", "/tmp/key")

	// Should have received an extra vars message
	foundExtraVarsMsg := false
	for _, msg := range ui.messageMessages {
		if strings.Contains(msg, "[Extra Vars]") {
			foundExtraVarsMsg = true
			// Verify it contains expected fields
			require.Contains(t, msg, "packer_builder_type")
			require.Contains(t, msg, "packer_build_name")
		}
	}
	require.True(t, foundExtraVarsMsg, "Should have logged extra vars when ShowExtraVars is enabled")
}
